package model

import (
	"encoding/json"
	"fmt"
	"strings"

	"go.bryk.io/pkg/errors"
)

type Operator string

const (
	Equals        Operator = "="
	NotEquals     Operator = "!="
	GreaterThan   Operator = ">"
	GreaterEqual  Operator = ">="
	LessThan      Operator = "<"
	LessEqual     Operator = "<="
	Contains      Operator = "*"
	NotContains   Operator = "!*"
	StartsWith    Operator = "^"
	NotStartsWith Operator = "!^"
	EndsWith      Operator = "$"
	NotEndsWith   Operator = "!$"
	In            Operator = "in"
	NotIn         Operator = "not_in"
	IsSet         Operator = "is_set"
	IsNotSet      Operator = "not_set"
)

type Condition []any

func NewCondition(field string, operator Operator, values ...any) Condition {
	arr := make([]any, 0, 2+len(values))
	arr = append(arr, field, operator)
	arr = append(arr, values...)
	return Condition(arr)
}

func (c Condition) Field() string {
	return fmt.Sprint(c[0])
}

func (c Condition) Operator() Operator {
	if len(c) < 2 {
		return ""
	}
	if op, ok := c[1].(Operator); ok {
		return op
	}
	return Operator(fmt.Sprint(c[1]))
}

func (c Condition) Value() any {
	if len(c) < 3 {
		return nil
	}
	return c[2]
}

func (c Condition) Values() []any {
	if len(c) <= 2 {
		return nil
	}
	return c[2:]
}

type OrderDirection string

const (
	Asc  OrderDirection = "asc"
	Desc OrderDirection = "desc"
)

func NewSearchOrderItem(field string, direction ...OrderDirection) SearchOrderItem {
	if len(direction) > 0 {
		return SearchOrderItem{field, string(direction[0])}
	}
	return SearchOrderItem{field}
}

type SearchOrderItem []string

func (item SearchOrderItem) Field() string {
	return item[0]
}

func (item SearchOrderItem) Direction() OrderDirection {
	if len(item) == 2 {
		return OrderDirection(strings.ToLower(item[1]))
	}
	return Asc
}

func NewSearchOrder(field string, direction ...OrderDirection) SearchOrder {
	return SearchOrder{NewSearchOrderItem(field, direction...)}
}

func NewSearchOrderMulti(items map[string]OrderDirection) SearchOrder {
	orderItems := make([]SearchOrderItem, 0, len(items))
	for field, direction := range items {
		orderItems = append(orderItems, NewSearchOrderItem(field, direction))
	}
	return SearchOrder(orderItems)
}

type SearchOrder []SearchOrderItem

func NewSearchGraph() *SearchGraph {
	return &SearchGraph{
		condition: nil,
		and:       nil,
		or:        nil,
		order:     nil,
	}
}

type SearchGraph struct {
	condition *Condition
	and       []SearchNode
	or        []SearchNode
	order     SearchOrder
}

func (this *SearchGraph) Condition(c *Condition) *SearchGraph {
	this.condition = c
	this.and = nil
	this.or = nil
	_ = this.validate()
	return this
}

func (this *SearchGraph) And(nodes ...SearchNode) *SearchGraph {
	this.condition = nil
	this.and = nodes
	this.or = nil
	_ = this.validate()
	return this
}

func (this *SearchGraph) Or(nodes ...SearchNode) *SearchGraph {
	this.condition = nil
	this.and = nil
	this.or = nodes
	_ = this.validate()
	return this
}

func (this *SearchGraph) OrderBy(field string, direction ...OrderDirection) *SearchGraph {
	this.order = NewSearchOrder(field, direction...)
	return this
}

func (this *SearchGraph) Order(o SearchOrder) *SearchGraph {
	this.order = o
	return this
}

func (this *SearchGraph) validate() error {
	setCount := 0
	if this.condition != nil {
		setCount++
	}
	if len(this.and) > 0 {
		setCount++
	}
	if len(this.or) > 0 {
		setCount++
	}
	if setCount > 1 {
		return errors.New("SearchGraph: condition, and, or are mutually exclusive; at most one may be set")
	}
	return nil
}

// MarshalJSON implements json.Marshaler.
func (this *SearchGraph) MarshalJSON() ([]byte, error) {
	payload := struct {
		Condition *Condition   `json:"if,omitempty"`
		And       []SearchNode `json:"and,omitempty"`
		Or        []SearchNode `json:"or,omitempty"`
		Order     SearchOrder  `json:"order,omitempty"`
	}{
		Condition: this.condition,
		And:       this.and,
		Or:        this.or,
		Order:     this.order,
	}
	return json.Marshal(payload)
}

// UnmarshalJSON implements json.Unmarshaler.
func (this *SearchGraph) UnmarshalJSON(data []byte) error {
	var raw struct {
		Condition *Condition   `json:"if,omitempty"`
		And       []SearchNode `json:"and,omitempty"`
		Or        []SearchNode `json:"or,omitempty"`
		Order     SearchOrder  `json:"order,omitempty"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	this.condition = raw.Condition
	this.and = raw.And
	this.or = raw.Or
	this.order = raw.Order
	return this.validate()
}

func (this *SearchGraph) GetCondition() *Condition {
	return this.condition
}

func (this *SearchGraph) GetAnd() []SearchNode {
	return this.and
}

func (this *SearchGraph) GetOr() []SearchNode {
	return this.or
}

func (this *SearchGraph) GetOrder() SearchOrder {
	return this.order
}

type SearchNode struct {
	condition *Condition
	and       []SearchNode
	or        []SearchNode
}

func (this *SearchNode) Condition(c *Condition) *SearchNode {
	this.condition = c
	this.and = nil
	this.or = nil
	_ = this.validate()
	return this
}

func (this *SearchNode) And(node SearchNode) *SearchNode {
	this.condition = nil
	this.and = []SearchNode{node}
	this.or = nil
	_ = this.validate()
	return this
}

func (this *SearchNode) Or(node SearchNode) *SearchNode {
	this.condition = nil
	this.and = nil
	this.or = []SearchNode{node}
	_ = this.validate()
	return this
}

func (this *SearchNode) validate() error {
	setCount := 0
	if this.condition != nil {
		setCount++
	}
	if len(this.and) > 0 {
		setCount++
	}
	if len(this.or) > 0 {
		setCount++
	}
	if setCount > 1 {
		return errors.New("SearchNode: condition, and, or are mutually exclusive; at most one may be set")
	}
	return nil
}

// MarshalJSON implements json.Marshaler.
func (this *SearchNode) MarshalJSON() ([]byte, error) {
	payload := struct {
		Condition *Condition   `json:"if,omitempty"`
		And       []SearchNode `json:"and,omitempty"`
		Or        []SearchNode `json:"or,omitempty"`
	}{
		Condition: this.condition,
		And:       this.and,
		Or:        this.or,
	}
	return json.Marshal(payload)
}

// UnmarshalJSON implements json.Unmarshaler.
func (this *SearchNode) UnmarshalJSON(data []byte) error {
	var raw struct {
		Condition *Condition   `json:"if,omitempty"`
		And       []SearchNode `json:"and,omitempty"`
		Or        []SearchNode `json:"or,omitempty"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	this.condition = raw.Condition
	this.and = raw.And
	this.or = raw.Or
	return this.validate()
}

func (this *SearchNode) GetCondition() *Condition {
	return this.condition
}

func (this *SearchNode) GetAnd() []SearchNode {
	return this.and
}

func (this *SearchNode) GetOr() []SearchNode {
	return this.or
}

func populateDbMetadata(schema *ModelSchema) error {
	name, err := requireName(schema.Name())
	if err != nil {
		return err
	}
	columnSet, primaryKeys, tenantKey, fieldUnique, err := buildDbMetadata(schema.Fields(), schema.fieldsOrder, name)
	if err != nil {
		return err
	}
	schemaUnique, err := validateAndCollectEntityUnique(schema, columnSet)
	if err != nil {
		return err
	}
	if len(primaryKeys) == 0 {
		return errors.Errorf("entity '%s' must define at least one primary key column", name)
	}
	schema.primaryKeys = primaryKeys
	schema.allUniqueKeys = append(fieldUnique, schemaUnique...)
	if tenantKey != "" {
		schema.tenantKey = &tenantKey
	}
	return nil
}

func requireName(raw string) (string, error) {
	name := strings.TrimSpace(raw)
	if name == "" {
		return "", errors.New("entity schema name is required")
	}
	return name, nil
}

func buildDbMetadata(
	fields map[string]*ModelField,
	fieldsOrder []string,
	entityName string,
) (columnSet map[string]struct{}, primary []string, tenant string, uniques [][]string, err error) {
	columnSet = make(map[string]struct{}, len(fields))
	uniques = make([][]string, 0)

	for _, fieldName := range fieldsOrder {
		field, ok := fields[fieldName]
		if !ok || field == nil {
			continue
		}
		if err := validateFieldName(field); err != nil {
			return nil, nil, "", nil, errors.Wrapf(err, "entity '%s'", entityName)
		}
		columnName := field.Name()
		columnSet[columnName] = struct{}{}

		if field.IsUnique() {
			uniques = append(uniques, []string{columnName})
		}
		if field.IsPrimaryKey() {
			primary = append(primary, columnName)
		}
		if field.IsTenantKey() {
			if tenant != "" && tenant != columnName {
				return nil, nil, "", nil, errors.Errorf(
					"entity '%s': field '%s' must not be a tenant key because '%s' is already one",
					entityName, columnName, tenant)
			}
			tenant = columnName
		}
	}
	return columnSet, primary, tenant, uniques, nil
}

func validateAndCollectEntityUnique(
	schema *ModelSchema,
	columnSet map[string]struct{},
) ([][]string, error) {
	uniqueFields := schema.CompositeUniques()
	uniqueKeys := make([][]string, 0, len(uniqueFields))
	for _, compositeKey := range uniqueFields {
		if len(compositeKey) == 0 {
			continue
		}
		validated := make([]string, 0, len(compositeKey))
		for _, name := range compositeKey {
			trimmed := strings.TrimSpace(name)
			if trimmed == "" {
				continue
			}
			if _, ok := columnSet[trimmed]; !ok {
				return nil, errors.Errorf(
					"entity '%s': unknown column reference '%s' in unique constraint", schema.Name(), trimmed)
			}
			validated = append(validated, trimmed)
		}
		if len(validated) > 0 {
			uniqueKeys = append(uniqueKeys, validated)
		}
	}
	return uniqueKeys, nil
}
