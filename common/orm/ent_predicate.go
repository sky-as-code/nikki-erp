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
		vErr.Appendf("graph.condition", "condition '%s' must have at least 3 parts", this)
		return vErr
	}

	for _, part := range this {
		if len(fmt.Sprint(part)) == 0 {
			vErr.Appendf("graph.condition", "condition '%s' must have all non-empty parts", this)
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
		vErr.Appendf("graph.condition", "unrecognized entity '%s'", entityName)
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
		vErr.Appendf("graph.condition", "unrecognized relationship '%s' in condition '%s'", edgeName, this)
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
		vErr.Appendf("graph.condition", "%s in condition '%s'", err.Error(), this)
		return nil, vErr
	}

	if anyOp, ok := AnyOperators[operator]; ok {
		converted, err := util.ConvertType(value, fieldType)
		if err != nil {
			vErr.Appendf("graph.condition", "%s in condition '%s'", err.Error(), this)
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

	vErr.Appendf("graph.condition", "invalid operator '%s' in condition '%s'", operator, this)
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

// type SearchOrder struct {
// 	Field     string         `json:"field"`
// 	Direction OrderDirection `json:"dir"`
// }

type SearchOrderItem []string

func (this SearchOrderItem) Field() string {
	return this[0]
}

func (this SearchOrderItem) Direction() OrderDirection {
	if len(this) == 2 {
		return OrderDirection(strings.ToLower(this[1]))
	}
	return Asc
}

func (this SearchOrderItem) Validate(entityName string) *ft.ValidationErrorItem {
	var field string
	var direction *OrderDirection
	length := len(this)
	if length == 2 {
		direction = util.ToPtr(OrderDirection(strings.ToLower(this[1])))
	} else if length == 1 {
		direction = nil
	} else {
		return &ft.ValidationErrorItem{
			Field: "graph.order",
			Error: fmt.Sprintf("invalid order '%s'", this),
		}
	}

	field = this[0]

	if direction != nil && *direction != Asc && *direction != Desc {
		return &ft.ValidationErrorItem{
			Field: "graph.order",
			Error: fmt.Sprintf("invalid order direction '%s'", *direction),
		}
	}

	entity, ok := GetEntity(entityName)
	if !ok {
		return &ft.ValidationErrorItem{
			Field: "graph.order",
			Error: fmt.Sprintf("unrecognized entity '%s'", entityName),
		}
	}

	_, err := entity.FieldType(field)
	if err != nil {
		return &ft.ValidationErrorItem{
			Field: "graph.order",
			Error: fmt.Sprintf("%s in order '%s'", err.Error(), this),
		}
	}

	return nil
}

type SearchOrder []SearchOrderItem

func (this SearchOrder) Validate(entityName string) ft.ValidationErrors {
	vErr := ft.NewValidationErrors()

	for i, item := range this {
		if err := item.Validate(entityName); err != nil {
			vErr.Appendf(
				fmt.Sprintf("graph.order.%d", i),
				"%s in order '%s'", err.Error, item,
			)
		}
	}

	return vErr
}

func (this SearchOrder) ToOrderOptions(entityName string) ([]OrderOption, ft.ValidationErrors) {
	if vErr := this.Validate(entityName); vErr.Count() > 0 {
		return nil, vErr
	}

	opts := make([]OrderOption, 0, len(this))
	for _, item := range this {
		term := orderTermMap[item.Direction()]
		opts = append(opts, sql.OrderByField(item.Field(), term).ToFunc())
	}
	return opts, nil
}

// func ValidateSearchOrders(orders []SearchOrder, entityName string) ft.ValidationErrors {
// 	vErr := ft.NewValidationErrors()
// 	for _, order := range orders {
// 		if err := order.Validate(entityName); err != nil {
// 			vErr.Appendf("graph.order", "%s in order '%s'", err.Error(), order)
// 		}
// 	}
// 	return vErr
// }

type SearchGraph struct {
	Condition *Condition   `json:"if,omitempty"`
	And       []SearchNode `json:"and,omitempty"`
	Or        []SearchNode `json:"or,omitempty"`
	Order     SearchOrder  `json:"order,omitempty"`
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

// func ToOrder(entityName string, graph SearchGraph) ([]OrderOption, ft.ValidationErrors) {
// 	if vErr := graph.Order.Validate(entityName); vErr != nil {
// 		return nil, vErr
// 	}

// 	orders := graph.Order.ToOrderOptions(entityName)
// 	return orders, nil
// }

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
