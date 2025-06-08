package orm

import (
	"fmt"
	"reflect"
	"strings"

	"entgo.io/ent/dialect/sql"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
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

func (this Condition) Validate() ft.ValidationErrors {
	vErr := ft.NewValidationErrors()
	if len(this) < 3 {
		vErr.Appendf("query.condition", "condition '%s' must have at least 3 parts", this)
		return vErr
	}

	for _, part := range this {
		if len(fmt.Sprint(part)) == 0 {
			vErr.Appendf("query.condition", "condition '%s' must have all non-empty parts", this)
			return vErr
		}
	}

	return vErr
}

// ToPredicate converts a condition expression to a Predicate instance.
// Examples:
//
//	["first_name", "contains", "nikki"] => sql.FieldContainsFold(entUser.FieldFirstName, "nikki")
//
//	["age", ">", "30"] => sql.FieldGT(entUser.FieldAge, "30")
func (this Condition) ToPredicate(entityName string) (Predicate, ft.ValidationErrors) {
	vErr := this.Validate()
	if vErr.Count() > 0 {
		return nil, vErr
	}

	entity, ok := GetEntity(entityName)
	if !ok {
		vErr.Appendf("query.condition", "unrecognized entity '%s'", entityName)
		return nil, vErr
	}

	rawField := this.Field()
	fields := strings.Split(rawField, ".")
	noEdge := len(fields) == 1

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
		vErr.Appendf("query.condition", "unrecognized relationship '%s' in condition '%s'", edgeName, this)
		return nil, vErr
	}

	// ToPredicate() will continue processing more nested edges.
	edgePred, err := subCond.ToPredicate(edgeName)
	if err != nil {
		return nil, err
	}

	return hasEdgeWithFn(edgePred), nil
}

func (this Condition) toSimplePredicate(entity *EntityDescriptor) (Predicate, ft.ValidationErrors) {
	field := this.Field()
	operator := this.Operator()
	value := this.Value()
	values := this.Values()

	vErr := ft.NewValidationErrors()
	fieldType, err := entity.MatchFieldType(field, value)
	if err != nil {
		vErr.Appendf("query.condition", "%s in condition '%s'", err.Error(), this)
		return nil, vErr
	}

	if anyOp, ok := AnyOperators[operator]; ok {
		converted, err := util.ConvertType(value, fieldType)
		if err != nil {
			vErr.Appendf("query.condition", "%s in condition '%s'", err.Error(), this)
			return nil, vErr
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

	vErr.Appendf("query.condition", "invalid operator '%s' in condition '%s'", operator, this)
	return nil, vErr
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

func (this SearchOrder) Validate(entityName string) ft.ValidationErrors {
	vErr := ft.NewValidationErrors()
	if this.Direction != Asc && this.Direction != Desc {
		vErr.Appendf("query.order", "invalid order direction '%s'", this.Direction)
		return vErr
	}

	entity, ok := GetEntity(entityName)
	if !ok {
		vErr.Appendf("query.condition", "unrecognized entity '%s'", entityName)
		return vErr
	}

	_, err := entity.FieldType(this.Field)
	if err != nil {
		vErr.Appendf("query.condition", "%s in condition '%s'", err.Error(), this)
		return vErr
	}

	return vErr
}

func (this SearchOrder) ToOrderOption(entityName string) OrderOption {
	term := orderTermMap[this.Direction]
	return sql.OrderByField(this.Field, term).ToFunc()
}

func ValidateSearchOrders(orders []SearchOrder, entityName string) ft.ValidationErrors {
	vErr := ft.NewValidationErrors()
	for _, order := range orders {
		if err := order.Validate(entityName); err != nil {
			vErr.Appendf("query.order", "%s in order '%s'", err.Error(), order)
		}
	}
	return vErr
}

type SearchGraph struct {
	Condition *Condition    `json:"if,omitempty"`
	And       []SearchNode  `json:"and,omitempty"`
	Or        []SearchNode  `json:"or,omitempty"`
	Order     []SearchOrder `json:"order,omitempty"`
}

func (this SearchGraph) ToPredicate(entityName string) (Predicate, ft.ValidationErrors) {
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

func ToOrder(entityName string, graph SearchGraph) ([]OrderOption, ft.ValidationErrors) {
	if vErr := ValidateSearchOrders(graph.Order, entityName); vErr != nil {
		return nil, vErr
	}

	orders := make([]OrderOption, 0, len(graph.Order))
	for _, order := range graph.Order {
		orderOpt := order.ToOrderOption(entityName)
		orders = append(orders, orderOpt)
	}

	return orders, nil
}

// SearchNode represents a complex search criteria, its fields are mutually exclusive,
// which means only one field can be set at a time, the precedence is:
// Condition > NotCondition > And > Or
type SearchNode struct {
	Condition *Condition   `json:"if,omitempty"`
	And       []SearchNode `json:"and,omitempty"`
	Or        []SearchNode `json:"or,omitempty"`
}

func (this SearchNode) ToPredicate(entityName string) (Predicate, ft.ValidationErrors) {
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
