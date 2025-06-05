package orm

import (
	"fmt"
	"reflect"
	"strings"

	"entgo.io/ent/dialect/sql"
	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/util"
)

func NewCondition(field string, operator Operator, values ...any) Condition {
	arr := make([]any, 0, 2+len(values))
	arr = append(arr, field, operator)
	arr = append(arr, values...)
	return Condition(arr)
}

// Condition expression is made of at least 3 parts:
// 1. Field: the field to search
// 2. Operator: the operator to use, it must be appropriate with the value type
// 3. Value: the value to search for
// 3,4,5,6: Values for collection operators
//
// Eg: ["name", "*", "nikki"], ["age", ">", "30"], ["age", "in", "30", "40", "50"]
type Condition []any

func (this Condition) Field() string {
	return this[0].(string)
}

func (this Condition) Operator() Operator {
	return Operator(this[1].(string))
}

func (this Condition) Value() any {
	return this[2]
}

func (this Condition) Values() []any {
	return this[2:]
}

func (this Condition) Validate() error {
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
func (this Condition) ToPredicate(entityName string) (Predicate, error) {
	err := this.Validate()
	if err != nil {
		return nil, err
	}

	entity, ok := GetEntity(entityName)
	if !ok {
		return nil, errors.Errorf("unregistered entity '%s' in expression '%s'", entityName, this)
	}

	rawField := this.Field()
	fields := strings.Split(rawField, ".")
	noEdge := len(fields) == 0

	if noEdge {
		// Expr: ["first_name", "^", "admin"]
		// Result: sql.FieldHasPrefixFold(entUser.FieldFirstName, "admin")
		return this.toSimplePredicate(entity)
	}

	edgeName := fields[0]

	// Can be "name", but can also "parent.name" or "parent.leader.name"
	edgeField := fields[1]

	subCond := NewCondition(
		edgeField,
		this.Operator(),
		this.Values()...,
	)

	hasEdgeWithFn := entity.Edges[edgeName]
	if hasEdgeWithFn == nil {
		return nil, errors.Errorf("no Has<EdgeName>With() found in Descriptor for edge entity '%s' in expression '%s'", edgeName, this)
	}

	// ToPredicate() will continue processing more nested edges.
	edgePred, err := subCond.ToPredicate(edgeName)
	if err != nil {
		return nil, err
	}

	return hasEdgeWithFn(edgePred), nil
}

func (this Condition) toSimplePredicate(entity *EntityDescriptor) (Predicate, error) {
	field := this.Field()
	operator := this.Operator()
	value := this.Value()
	values := this.Values()

	fieldType, err := entity.MatchFieldType(field, value)
	if err != nil {
		return nil, err
	}

	if anyOp, ok := AnyOperators[operator]; ok {
		converted, err := util.ConvertType(value, fieldType)
		if err != nil {
			return nil, err
		}
		return anyOp(field, converted), nil
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

type OrderDirection string

const (
	Asc  OrderDirection = "asc"
	Desc OrderDirection = "desc"
)

var orderTermMap = map[OrderDirection]sql.OrderTermOption{
	Asc:  sql.OrderAsc(),
	Desc: sql.OrderDesc(),
}

type SearchOrder struct {
	Field     string         `json:"field"`
	Direction OrderDirection `json:"dir"`
}

func (this SearchOrder) Validate(entityName string) error {
	if this.Direction != Asc && this.Direction != Desc {
		return errors.Errorf("invalid order direction '%s'", this.Direction)
	}

	entity, ok := GetEntity(entityName)
	if !ok {
		return errors.Errorf("unregistered entity '%s'", entityName)
	}

	_, err := entity.FieldType(this.Field)
	if err != nil {
		return err
	}

	return nil
}

func (this SearchOrder) ToOrderOption(entityName string) (OrderOption, error) {
	term := orderTermMap[this.Direction]
	return sql.OrderByField(this.Field, term).ToFunc(), nil
}

func ValidateSearchOrders(orders []SearchOrder, entityName string) error {
	for _, order := range orders {
		if err := order.Validate(entityName); err != nil {
			return err
		}
	}
	return nil
}

type SearchGraph struct {
	Condition *Condition    `json:"if"`
	And       []SearchNode  `json:"and"`
	Or        []SearchNode  `json:"or"`
	Order     []SearchOrder `json:"order"`
}

func (this SearchGraph) ToPredicate(entityName string) (Predicate, error) {
	if this.Condition != nil {
		return this.Condition.ToPredicate(entityName)
	}

	if len(this.And) > 0 {
		preds := make([]Predicate, 0, len(this.And))
		for _, node := range this.And {
			pred, err := node.ToPredicate(entityName)
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
			pred, err := node.ToPredicate(entityName)
			if err != nil {
				return nil, err
			}
			preds = append(preds, pred)
		}
		return sql.OrPredicates(preds...), nil
	}

	return NoopPredicate, nil
}

func ToOrder[TOpt ~OrderOption](entityName string, graph *SearchGraph) ([]TOpt, error) {
	if err := ValidateSearchOrders(graph.Order, entityName); err != nil {
		return nil, err
	}

	orders := make([]TOpt, 0, len(graph.Order))
	for _, order := range graph.Order {
		orderOpt, err := order.ToOrderOption(entityName)
		if err != nil {
			return nil, err
		}
		orders = append(orders, orderOpt)
	}

	return orders, nil
}

// SearchNode represents a complex search criteria, its fields are mutually exclusive,
// which means only one field can be set at a time, the precedence is:
// Condition > NotCondition > And > Or
type SearchNode struct {
	Condition *Condition   `json:"if"`
	And       []SearchNode `json:"and"`
	Or        []SearchNode `json:"or"`
}

func (this SearchNode) ToPredicate(entityName string) (Predicate, error) {
	if this.Condition != nil {
		return this.Condition.ToPredicate(entityName)
	}

	if len(this.And) > 0 {
		preds := make([]Predicate, 0, len(this.And))
		for _, node := range this.And {
			pred, err := node.ToPredicate(entityName)
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
			pred, err := node.ToPredicate(entityName)
			if err != nil {
				return nil, err
			}
			preds = append(preds, pred)
		}
		return sql.OrPredicates(preds...), nil
	}

	return NoopPredicate, nil
}
