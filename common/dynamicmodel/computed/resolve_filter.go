package computed

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

// Filter validation for the SQL-compiled kinds. The filter reuses the platform's search-node
// shape and is validated against the SOURCE schema: every field must be a physical scalar there,
// every operator must be one a subquery predicate supports, and every value must be a scalar or
// a whole-string "${ctx.key}" reference to a declared context key.

func (this *resolver) validateSqlFilter(
	sourceSchema *dmodel.ModelSchema, filter *dmodel.SearchNode, contextKeys []string, plan *FieldPlan,
) error {
	usedKeys := map[string]bool{}
	if filter != nil {
		if err := this.walkFilterNode(sourceSchema, filter, plan, usedKeys, 1); err != nil {
			return err
		}
	}
	return checkContextKeys(contextKeys, usedKeys)
}

func (this *resolver) walkFilterNode(
	sourceSchema *dmodel.ModelSchema, node *dmodel.SearchNode, plan *FieldPlan,
	usedKeys map[string]bool, depth int,
) error {
	if depth > this.limits.MaxFilterNestingDepth {
		return errors.Errorf(
			"filter nests %d levels deep, exceeding the maximum of %d", depth, this.limits.MaxFilterNestingDepth)
	}
	if condition := node.GetCondition(); condition != nil {
		return this.validateFilterCondition(sourceSchema, condition, plan, usedKeys)
	}
	children := node.GetAnd()
	if len(children) == 0 {
		children = node.GetOr()
	}
	if len(children) == 0 {
		return errors.New("filter node must set one of: if, and, or")
	}
	for i := range children {
		if err := this.walkFilterNode(sourceSchema, &children[i], plan, usedKeys, depth+1); err != nil {
			return err
		}
	}
	return nil
}

func (this *resolver) validateFilterCondition(
	sourceSchema *dmodel.ModelSchema, condition dmodel.Condition, plan *FieldPlan,
	usedKeys map[string]bool,
) error {
	if len(condition) < 2 {
		return errors.New("filter condition needs at least a field and an operator")
	}
	if _, err := this.resolveSourceScalar(sourceSchema, condition.Field(), plan); err != nil {
		return errors.Wrap(err, "filter")
	}
	operator := condition.Operator()
	if !filterOperatorAllowed(operator) {
		return errors.Errorf("operator %q is not supported in a computed filter", operator)
	}
	if operatorNeedsValue(operator) && len(condition.Values()) == 0 {
		return errors.Errorf("operator %q requires a value", operator)
	}
	for _, value := range condition.Values() {
		if err := validateFilterValue(value, usedKeys); err != nil {
			return err
		}
	}
	return nil
}

// filterOperatorAllowed excludes the graph-only linked/not_linked operators: they reference an
// edge of the queried schema and compile to their own subquery, which must never nest inside a
// computed field's subquery.
func filterOperatorAllowed(operator dmodel.Operator) bool {
	switch operator {
	case dmodel.Equals, dmodel.NotEquals,
		dmodel.GreaterThan, dmodel.GreaterEqual, dmodel.LessThan, dmodel.LessEqual,
		dmodel.Contains, dmodel.NotContains,
		dmodel.StartsWith, dmodel.NotStartsWith, dmodel.EndsWith, dmodel.NotEndsWith,
		dmodel.In, dmodel.NotIn,
		dmodel.IsSet, dmodel.IsNotSet:
		return true
	}
	return false
}

func operatorNeedsValue(operator dmodel.Operator) bool {
	return operator != dmodel.IsSet && operator != dmodel.IsNotSet
}

// validateFilterValue accepts scalars only. A whole-string "${ctx.key}" value is a context
// reference and is recorded; any other "${...}" placeholder is rejected outright so a filter can
// never smuggle an unresolved substitution into SQL.
func validateFilterValue(value any, usedKeys map[string]bool) error {
	if key, ok := CtxKeyOf(value); ok {
		usedKeys[key] = true
		return nil
	}
	if str, ok := value.(string); ok && looksLikePlaceholder(str) {
		return errors.Errorf("filter value %q is not a valid context reference; use \"${ctx.key}\"", str)
	}
	if value == nil {
		return nil
	}
	if _, err := literalType(value); err != nil {
		return errors.Wrap(err, "filter value")
	}
	return nil
}

func checkContextKeys(declared []string, used map[string]bool) error {
	declaredSet := map[string]bool{}
	for _, key := range declared {
		declaredSet[key] = true
	}
	for key := range used {
		if !declaredSet[key] {
			return errors.Errorf(
				"filter references context key %q which is not declared in the definition's context list", key)
		}
	}
	for _, key := range declared {
		if !used[key] {
			return errors.Errorf("context key %q is declared but never referenced by the filter", key)
		}
	}
	return nil
}
