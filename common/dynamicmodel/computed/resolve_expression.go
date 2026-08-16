package computed

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

// Expression-kind resolution: validate the tree shape, resolve every field reference through the
// schema registry, infer the result type, and collect operands and dependencies.

func (this *resolver) buildExpressionPlan(
	schema *dmodel.ModelSchema, field *dmodel.ModelField, plan *FieldPlan,
) error {
	expr := plan.Def.Expression
	if depth := exprDepth(expr); depth > this.limits.MaxExpressionNestingDepth {
		return errors.Errorf(
			"computed field %s.%s nests %d levels deep, exceeding the maximum of %d",
			schema.Name(), field.Name(), depth, this.limits.MaxExpressionNestingDepth)
	}
	if err := validateExtractLiterals(expr); err != nil {
		return errors.Wrapf(err, "computed field %s.%s", schema.Name(), field.Name())
	}

	operands := map[string]bool{}
	inferred, err := InferType(expr, this.fieldTypeResolver(schema, plan, operands))
	if err != nil {
		return errors.Wrapf(err, "computed field %s.%s", schema.Name(), field.Name())
	}
	if inferred == TypeNull {
		inferred = Type(field.DataType().String())
	}
	plan.Type = inferred
	return nil
}

// fieldTypeResolver types a same-schema field reference. A reference to another computed field
// recursively resolves it — this is where dependency edges (and therefore cycles) appear.
func (this *resolver) fieldTypeResolver(
	schema *dmodel.ModelSchema, plan *FieldPlan, operands map[string]bool,
) FieldTypeResolver {
	return func(name string) (Type, error) {
		referenced, ok := schema.Field(name)
		if !ok {
			return TypeUnknown, errors.Errorf("Unknown field %q", schema.Name()+"."+name)
		}
		if referenced.IsEdgeModel() {
			return TypeUnknown, errors.Errorf(
				"field %q is an edge, not a scalar; reference a leaf through a related computed field instead", name)
		}
		plan.Dependencies = append(plan.Dependencies, FieldRef{Schema: schema.Name(), Field: name})
		if referenced.IsComputed() {
			return this.resolveComputedOperand(schema, referenced, plan, operands)
		}
		if referenced.IsVirtual() {
			return TypeUnknown, errors.Errorf(
				"field %q is filled by service code after the read; a computed field cannot depend on it", name)
		}
		if !operands[name] {
			operands[name] = true
			plan.PhysicalOperands = append(plan.PhysicalOperands, name)
		}
		return Type(referenced.DataType().String()), nil
	}
}

// resolveComputedOperand resolves a computed-on-computed reference and folds the dependency's
// own physical operands into this plan, so the projection carries everything evaluation needs.
func (this *resolver) resolveComputedOperand(
	schema *dmodel.ModelSchema, referenced *dmodel.ModelField, plan *FieldPlan, operands map[string]bool,
) (Type, error) {
	depPlan, err := this.resolveField(schema, referenced)
	if err != nil {
		return TypeUnknown, err
	}
	plan.ComputedDeps = append(plan.ComputedDeps, referenced.Name())
	for _, operand := range depPlan.PhysicalOperands {
		if !operands[operand] {
			operands[operand] = true
			plan.PhysicalOperands = append(plan.PhysicalOperands, operand)
		}
	}
	return depPlan.Type, nil
}

func exprDepth(expr Expr) int {
	switch node := expr.(type) {
	case BinaryExpr:
		return 1 + maxInt(exprDepth(node.Left), exprDepth(node.Right))
	case UnaryExpr:
		return 1 + exprDepth(node.Operand)
	case FunctionExpr:
		return 1 + maxDepth(node.Args)
	case CaseExpr:
		depth := 0
		for _, branch := range node.Whens {
			depth = maxInt(depth, maxInt(exprDepth(branch.When), exprDepth(branch.Then)))
		}
		if node.Else != nil {
			depth = maxInt(depth, exprDepth(node.Else))
		}
		return 1 + depth
	}
	return 1
}

func maxDepth(exprs []Expr) int {
	depth := 0
	for _, expr := range exprs {
		depth = maxInt(depth, exprDepth(expr))
	}
	return depth
}

func maxInt(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

// validateExtractLiterals enforces that extract()'s first argument is a compile-time literal
// from the fixed part whitelist, everywhere in the tree.
func validateExtractLiterals(expr Expr) error {
	switch node := expr.(type) {
	case BinaryExpr:
		if err := validateExtractLiterals(node.Left); err != nil {
			return err
		}
		return validateExtractLiterals(node.Right)
	case UnaryExpr:
		return validateExtractLiterals(node.Operand)
	case FunctionExpr:
		return validateExtractCall(node)
	case CaseExpr:
		for _, branch := range node.Whens {
			if err := validateExtractLiterals(branch.When); err != nil {
				return err
			}
			if err := validateExtractLiterals(branch.Then); err != nil {
				return err
			}
		}
		if node.Else != nil {
			return validateExtractLiterals(node.Else)
		}
	}
	return nil
}

func validateExtractCall(node FunctionExpr) error {
	for _, arg := range node.Args {
		if err := validateExtractLiterals(arg); err != nil {
			return err
		}
	}
	if node.Name != "extract" || len(node.Args) == 0 {
		return nil
	}
	literal, ok := node.Args[0].(LiteralExpr)
	if !ok {
		return errors.New("function \"extract\" requires a literal part name as its first argument")
	}
	part, ok := literal.Value.(string)
	if !ok || !IsValidExtractPart(part) {
		return errors.Errorf("function \"extract\" does not support part %v", literal.Value)
	}
	return nil
}
