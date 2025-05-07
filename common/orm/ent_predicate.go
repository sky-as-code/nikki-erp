package orm

import (
	"fmt"
	"reflect"
	"strings"

	"entgo.io/ent/dialect/sql"
	"go.bryk.io/pkg/errors"
)

func NewConditionExpr(field string, operator Operator, values ...any) ConditionExpr {
	arr := make([]any, 0, 2+len(values))
	arr = append(arr, field, operator)
	arr = append(arr, values...)
	return ConditionExpr(arr)
}

// Condition expression is made of at least 3 parts:
// 1. Field: the field to search
// 2. Operator: the operator to use, it must be appropriate with the value type
// 3. Value: the value to search for
// 3,4,5,6: Values for collection operators
//
// Eg: ["name", "ilike", "nikki"], ["age", ">", "30"], ["age", "in", "30", "40", "50"]
type ConditionExpr []any

func (this ConditionExpr) Field() string {
	return this[0].(string)
}

func (this ConditionExpr) Operator() Operator {
	return Operator(this[1].(string))
}

func (this ConditionExpr) Value() any {
	return this[2]
}

func (this ConditionExpr) Values() []any {
	return this[2:]
}

func (this ConditionExpr) Validate() error {
	if len(this) < 3 {
		return errors.Errorf("condition '%s' must have at least 3 parts", this)
	}

	for _, part := range this {
		if len(fmt.Sprint(part)) == 0 {
			return errors.Errorf("condition '%s' must have all non-empty parts", this)
		}
	}

	return nil
}

// ToPredicate converts a condition expression to a Predicate instance.
// Examples:
//
//	["first_name", "contains", "nikki"] => sql.FieldContainsFold(entUser.FieldFirstName, "nikki")
//
//	["age", ">", "30"] => sql.FieldGT(entUser.FieldAge, "30")
func (this ConditionExpr) ToPredicate(entity *EntityDescriptor) (Predicate, error) {
	err := this.Validate()
	if err != nil {
		return nil, err
	}

	field := this.Field()
	operator := this.Operator()
	value := this.Value()
	values := this.Values()

	fieldType, err := entity.MatchFieldType(field, value)
	if err != nil {
		return nil, err
	}

	if anyOp, ok := AnyOperators[operator]; ok {
		convertedValue := reflect.ValueOf(value).Convert(fieldType).Interface()
		return anyOp(field, convertedValue), nil
	}

	if collectionOp, ok := CollectionOperators[operator]; ok {
		return collectionOp(field, values...), nil
	}

	if nullOp, ok := NullOperators[operator]; ok {
		return nullOp(field), nil
	}

	isString := fieldType.Kind() == reflect.String

	if stringOp, ok := StringOperators[operator]; ok && isString {
		return stringOp(field, fmt.Sprint(value)), nil
	}

	return nil, errors.Errorf("invalid operator '%s' in expression '%s'", operator, this)
}

func NewCondition(expr []any, rootEntity *EntityDescriptor) (*Condition, error) {
	condExpr := ConditionExpr(expr)
	if err := condExpr.Validate(); err != nil {
		return nil, err
	}

	return &Condition{
		Expression: condExpr,
		Entity:     rootEntity,
	}, nil
}

type Condition struct {
	Expression ConditionExpr `json:"expression"`
	Entity     *EntityDescriptor
}

// ToPredicate converts a declarative condition to Ent API Predicate.
//
//	Expr: ["first_name", "^", "admin"]
//	Result: sql.FieldHasPrefixFold(entUser.FieldFirstName, "admin")
//
//	Expr: ["groups.leader.age", ">=", "30"]
//	Result: HasGroupsWith(HasLeaderWith(sql.FieldGT(entUser.FieldAge, "30")))
func (this Condition) ToPredicate() (pred Predicate, err error) {
	fields := strings.Split(this.Expression.Field(), ".")
	noEdge := len(fields) == 0

	if noEdge {
		// Expr: ["first_name", "^", "admin"]
		// Result: sql.FieldHasPrefixFold(entUser.FieldFirstName, "admin")
		return this.Expression.ToPredicate(this.Entity)
	}

	edgeName := fields[0]
	if len(edgeName) == 0 {
		return nil, errors.Errorf("invalid edge name in expression '%s'", this.Expression)
	}

	edgeEntity, ok := GetEntity(edgeName)
	if !ok {
		return nil, errors.Errorf("unregistered edge entity '%s' in expression '%s'", edgeName, this.Expression)
	}

	// Can be "groups.name" or "groups.leader.age"
	edgeField := fields[1]

	cond := &Condition{
		Expression: NewConditionExpr(
			edgeField,
			this.Expression.Operator(),
			this.Expression.Values()...,
		),
		Entity: edgeEntity,
	}
	hasEdgeWithFn := this.Entity.Edges[edgeName]
	if hasEdgeWithFn == nil {
		return nil, errors.Errorf("no Has<EdgeName>With() found in Descriptor for edge entity '%s' in expression '%s'", edgeName, this.Expression)
	}

	// ToPredicate() will continue processing edge field names if there are more.
	edgePred, err := cond.ToPredicate()
	if err != nil {
		return nil, err
	}

	return hasEdgeWithFn(edgePred), nil
}

type SearchGraph struct {
	And []SearchNode `json:"and"`
	Or  []SearchNode `json:"or"`
}

func (this SearchGraph) ToPredicate() (Predicate, error) {
	if len(this.And) > 0 {
		preds := make([]Predicate, 0, len(this.And))
		for _, node := range this.And {
			pred, err := node.ToPredicate()
			if err != nil {
				return nil, err
			}
			preds = append(preds, pred)
		}
		return sql.AndPredicates(preds...), nil

	}

	if len(this.Or) > 0 {
		preds := make([]Predicate, 0, len(this.Or))
		for _, node := range this.Or {
			pred, err := node.ToPredicate()
			if err != nil {
				return nil, err
			}
			preds = append(preds, pred)
		}
		return sql.OrPredicates(preds...), nil
	}

	return NoopPredicate, nil
}

// SearchNode represents a complex search criteria, its fields are mutually exclusive,
// which means only one field can be set at a time, the precedence is:
// Condition > NotCondition > And > Or
type SearchNode struct {
	Condition    *Condition   `json:"if"`
	NotCondition *Condition   `json:"ifnot"`
	And          []SearchNode `json:"and"`
	Or           []SearchNode `json:"or"`
}

func (this SearchNode) ToPredicate() (Predicate, error) {
	if this.Condition != nil {
		return this.Condition.ToPredicate()
	}

	if this.NotCondition != nil {
		p, err := this.NotCondition.ToPredicate()
		if err != nil {
			return nil, err
		}
		return sql.NotPredicates(p), nil
	}

	if len(this.And) > 0 {
		preds := make([]Predicate, 0, len(this.And))
		for _, node := range this.And {
			pred, err := node.ToPredicate()
			if err != nil {
				return nil, err
			}
			preds = append(preds, pred)
		}
		return sql.AndPredicates(preds...), nil

	}

	if len(this.Or) > 0 {
		preds := make([]Predicate, 0, len(this.Or))
		for _, node := range this.Or {
			pred, err := node.ToPredicate()
			if err != nil {
				return nil, err
			}
			preds = append(preds, pred)
		}
		return sql.OrPredicates(preds...), nil
	}

	return NoopPredicate, nil
}
