package orm

import (
	"fmt"
	"reflect"
	"strings"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/dialect/sql/sqljson"

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
	return this[1].(Operator)
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

	hasEdgeWithFn, err := entity.EdgePredicate(edgeName)
	if err != nil {
		vErr.Appendf("graph.condition", "%s in condition '%s'", err.Error(), this)
		return nil, vErr
	}

	// Recursive ToPredicate() will continue processing more nested edges.
	edgePred, vErrs := subCond.ToPredicate(edgeName)
	if vErrs.Count() > 0 {
		return nil, vErrs
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

	baseType := fieldType
	if fieldType.Kind() == reflect.Ptr {
		baseType = fieldType.Elem()
	}
	isString := baseType.Kind() == reflect.String

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

var orderDirMap = map[OrderDirection]sql.OrderTermOption{
	Asc:  sql.OrderAsc(),
	Desc: sql.OrderDesc(),
}

type SearchOrderItem []string

func (this SearchOrderItem) Fields() (columnField string, subFields []string) {
	fieldParts := strings.Split(this[0], ".")
	partCount := len(fieldParts)
	if partCount >= 2 {
		return fieldParts[0], fieldParts[1:]
	}
	return fieldParts[0], nil
}

func (this SearchOrderItem) Direction() OrderDirection {
	if len(this) == 2 {
		return OrderDirection(strings.ToLower(this[1]))
	}
	return Asc
}

func (this SearchOrderItem) Validate(entityName string) *ft.ValidationErrorItem {
	var direction *OrderDirection

	switch len(this) {
	case 2:
		direction = util.ToPtr(OrderDirection(strings.ToLower(this[1])))
	case 1:
		direction = nil
	default:
		return &ft.ValidationErrorItem{
			Field: "graph.order",
			Error: fmt.Sprintf("invalid order '%s'", this),
		}
	}

	if direction != nil && *direction != Asc && *direction != Desc {
		return &ft.ValidationErrorItem{
			Field: "graph.order",
			Error: fmt.Sprintf("invalid order direction '%s' in order '%s'", *direction, this),
		}
	}

	_, ok := GetEntity(entityName)
	if !ok {
		return &ft.ValidationErrorItem{
			Field: "graph.order",
			Error: fmt.Sprintf("unrecognized entity '%s' in order '%s'", entityName, this),
		}
	}

	return nil
}

func (this SearchOrderItem) ToOrderOption(entityName string, vErr *ft.ValidationErrors) OrderOption {
	columnField, subFields := this.Fields()

	entity, _ := GetEntity(entityName)                      // No error check, because we already validated.
	orderByEdgeFn, _ := entity.OrderByEdgeStep(columnField) // The error check is covered by entity.EdgePredicate
	fieldType, errField := entity.FieldType(columnField)
	_, errEdge := entity.EdgePredicate(columnField)
	isOrderByField := errField == nil
	isOrderByEdge := errEdge == nil

	if isOrderByField {
		opt, edgeErr := this.toOrderByFieldOption(fieldType, columnField, subFields)
		if edgeErr.Count() > 0 {
			vErr.Merge(edgeErr)
		}
		return opt
	} else if isOrderByEdge {
		entityName = columnField
		columnField = subFields[0]
		subFields = subFields[1:]
		opt, edgeErr := this.toOrderByEdgeOption(entityName, columnField, subFields, orderByEdgeFn)
		if edgeErr.Count() > 0 {
			vErr.Merge(edgeErr)
		}
		return opt
	}
	vErr.Appendf("graph.order", "%s in order '%s'", errField.Error(), this)
	return nil
}

func (this SearchOrderItem) toOrderByFieldOption(fieldType reflect.Type, columnField string, subFields []string) (OrderOption, ft.ValidationErrors) {
	vErr := ft.NewValidationErrors()
	dir := this.Direction()
	dirTerm := orderDirMap[dir]
	isOrderBySubField := len(subFields) > 0
	isJsonField := fieldType == reflect.TypeOf(map[string]string{})
	canOrderBySubField := isJsonField && isOrderBySubField
	if canOrderBySubField { // For JSON column
		// TODO: Validate whether subFields really exists in the JSON field
		return sqljson.OrderValue(columnField, sqljson.Path(subFields...)), vErr
	} else {
		return sql.OrderByField(columnField, dirTerm).ToFunc(), vErr
	}
}

func (this SearchOrderItem) toOrderByEdgeOption(
	entityName string, columnField string, subFields []string, orderByEdgeFn OrderByEdgeFn,
) (orderOption OrderOption, vErr ft.ValidationErrors) {
	vErr = ft.NewValidationErrors()
	dir := this.Direction()
	dirTerm := orderDirMap[dir]

	edgeEntity, ok := GetEntity(entityName)
	if !ok {
		vErr.Appendf("graph.order", "unrecognized entity '%s' in order '%s'", entityName, this)
		return nil, vErr
	}

	edgeFieldType, err := edgeEntity.FieldType(columnField)
	if err != nil {
		vErr.Appendf("graph.order", "%s in order '%s'", err.Error(), this)
		return nil, vErr
	}

	isOrderBySubField := len(subFields) > 0
	isJsonField := edgeFieldType == reflect.TypeOf(map[string]string{})
	canOrderBySubField := isJsonField && isOrderBySubField

	var term sql.OrderTerm
	if canOrderBySubField {
		this.orderByEdgeJsonField(columnField, subFields, dir, term)
	} else if !isOrderBySubField {
		term = sql.OrderByField(columnField, dirTerm)
	}

	orderOption = func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(
			s,
			orderByEdgeFn(),
			term,
		)
	}
	return
}

func (SearchOrderItem) orderByEdgeJsonField(columnField string, subFields []string, dir OrderDirection, term sql.OrderTerm) {
	opts := []sql.OrderTermOption{
		sql.OrderAs(fmt.Sprintf("%s_%s", columnField, strings.Join(subFields, "_"))),
	}
	if dir == Desc {
		opts = append(opts, sql.OrderDesc())
	}
	term = &sql.OrderExprTerm{
		Expr: func(_ *sql.Selector) sql.Querier {
			querier := sqljson.ValuePath(columnField, sqljson.Path(subFields...))
			return querier
		},
		OrderTermOptions: *sql.NewOrderTermOptions(opts...),
	}
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
	var vErr ft.ValidationErrors
	if vErr = this.Validate(entityName); vErr.Count() > 0 {
		return nil, vErr
	}

	opts := make([]OrderOption, 0, len(this))
	for _, item := range this {
		opt := item.ToOrderOption(entityName, &vErr)
		if opt == nil {
			return nil, vErr
		}
		opts = append(opts, opt)
	}
	return opts, vErr
}

type SearchGraph struct {
	Condition *Condition   `json:"if,omitempty"`
	And       []SearchNode `json:"and,omitempty"`
	Or        []SearchNode `json:"or,omitempty"`
	Order     SearchOrder  `json:"order,omitempty"`
}

func (this SearchGraph) ToPredicate(entityName string) (Predicate, ft.ValidationErrors) {
	return buildPredicate(entityName, this.Condition, this.And, this.Or)
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
	return buildPredicate(entityName, this.Condition, this.And, this.Or)
}

func buildPredicate(entityName string, condition *Condition, and []SearchNode, or []SearchNode) (Predicate, ft.ValidationErrors) {
	vErrs := ft.NewValidationErrors()

	if condition != nil {
		return condition.ToPredicate(entityName)
	}

	if len(and) > 0 {
		pred := buildLogicalPredicates(and, entityName, &vErrs, sql.AndPredicates)
		return pred, vErrs
		// preds := make([]Predicate, 0, len(and))
		// for _, node := range and {
		// 	pred, err := node.ToPredicate(entityName)
		// 	if err.Count() > 0 {
		// 		vErrs.Merge(err)
		// 		return nil, vErrs
		// 	}
		// 	preds = append(preds, pred)
		// }
		// return sql.AndPredicates(preds...), nil

	}

	if len(or) > 0 {
		pred := buildLogicalPredicates(or, entityName, &vErrs, sql.OrPredicates)
		return pred, vErrs
		// preds := make([]Predicate, 0, len(or))
		// for _, node := range or {
		// 	pred, err := node.ToPredicate(entityName)
		// 	if err.Count() > 0 {
		// 		vErrs.Merge(err)
		// 		return nil, vErrs
		// 	}
		// 	preds = append(preds, pred)
		// }
		// return sql.OrPredicates(preds...), nil
	}

	return NoopPredicate, vErrs
}

func buildLogicalPredicates(
	logicalOp []SearchNode, entityName string, vErrs *ft.ValidationErrors, logicalPred LogicalPredicateFn,
) Predicate {
	preds := make([]Predicate, 0, len(logicalOp))
	for _, node := range logicalOp {
		pred, err := node.ToPredicate(entityName)
		if err.Count() > 0 {
			vErrs.Merge(err)
			return nil
		}
		preds = append(preds, pred)
	}
	return logicalPred(preds...)
}

type LogicalPredicateFn func(predicates ...Predicate) Predicate
