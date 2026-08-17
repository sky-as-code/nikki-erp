package model

import (
	"encoding/json"
	"fmt"
	"strings"

	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/array"
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
	// Linked / NotLinked: only for graph conditions on a many edge (one:many, many:many).
	// Field is the edge name (no dot). Value is the peer / child row primary key to test linkage.
	Linked    Operator = "linked"
	NotLinked Operator = "not_linked"
)

type Condition []any

func NewCondition(field string, operator Operator, values ...any) Condition {
	arr := make([]any, 0, 2+len(values))
	arr = append(arr, field, operator)
	arr = append(arr, values...)
	return Condition(arr)
}

func (c Condition) Field() string {
	if len(c) == 0 {
		return ""
	}
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
	if len(c) < 3 {
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
	condition Condition
	and       []SearchNode
	or        []SearchNode
	order     SearchOrder
}

func (this *SearchGraph) NewCondition(field string, operator Operator, values ...any) *SearchGraph {
	return this.Condition(NewCondition(field, operator, values...))
}

func (this *SearchGraph) Condition(c Condition) *SearchGraph {
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

func (this *SearchGraph) ToSearchNode() *SearchNode {
	return &SearchNode{
		condition: this.condition,
		and:       this.and,
		or:        this.or,
	}
}

func (this *SearchGraph) GetCondition() Condition {
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

// MarshalJSON implements json.Marshaler.
func (this *SearchGraph) MarshalJSON() ([]byte, error) {
	payload := struct {
		Condition Condition    `json:"if,omitempty"`
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
		Condition Condition    `json:"if,omitempty"`
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

func (this *SearchGraph) UnmarshalText(text []byte) error {
	return this.UnmarshalJSON(text)
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
		return errors.New(
			"SearchGraph.validate: condition, and, or are mutually exclusive; at most one may be set")
	}
	return nil
}

func NewSearchNode() *SearchNode {
	return &SearchNode{}
}

type SearchNode struct {
	condition Condition
	and       []SearchNode
	or        []SearchNode
}

func (this *SearchNode) NewCondition(field string, operator Operator, values ...any) *SearchNode {
	return this.Condition(NewCondition(field, operator, values...))
}

func (this *SearchNode) Condition(c Condition) *SearchNode {
	this.condition = c
	this.and = nil
	this.or = nil
	_ = this.validate()
	return this
}

func (this *SearchNode) And(nodes ...SearchNode) *SearchNode {
	this.condition = nil
	this.and = nodes
	this.or = nil
	_ = this.validate()
	return this
}

func (this *SearchNode) Or(nodes ...SearchNode) *SearchNode {
	this.condition = nil
	this.and = nil
	this.or = nodes
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
		return errors.New(
			"SearchNode.validate: condition, and, or are mutually exclusive; at most one may be set")
	}
	return nil
}

// MarshalJSON implements json.Marshaler.
func (this *SearchNode) MarshalJSON() ([]byte, error) {
	payload := struct {
		Condition Condition    `json:"if,omitempty"`
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
		Condition Condition    `json:"if,omitempty"`
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

func (this *SearchNode) GetCondition() Condition {
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
	schemaUnique, err := validateCompositeUniquesForDb(schema, columnSet)
	if err != nil {
		return err
	}
	validatedPartials, err := validatePartialUniquesForDb(schema, columnSet)
	if err != nil {
		return err
	}
	validatedSearchIndexGroups, err := validateSearchIndexGroupsForDb(schema, columnSet)
	if err != nil {
		return err
	}
	if len(primaryKeys) == 0 {
		return errors.Errorf("populateDbMetadata: model '%s' must define at least one primary key column", name)
	}
	schema.primaryKeys = append([]string{}, primaryKeys...)
	schema.partialUniques = validatedPartials
	schema.searchIndexGroups = validatedSearchIndexGroups
	schema.allUniqueKeys = append(fieldUnique, schemaUnique...)
	schema.allUniqueColumns = array.Map(schema.allUniqueKeys, func(group CompositeUniqueParam) []string {
		return group.Fields
	})
	if tenantKey != "" {
		schema.tenantKey = &tenantKey
	} else {
		schema.tenantKey = nil
	}
	return nil
}

func requireName(raw string) (string, error) {
	name := strings.TrimSpace(raw)
	if name == "" {
		return "", errors.New("requireName: model schema name is required")
	}
	return name, nil
}

func buildDbMetadata(
	fields map[string]*ModelField,
	fieldsOrder []string,
	schemaName string,
) (columnSet map[string]struct{}, primary []string, tenant string, uniques []CompositeUniqueParam, err error) {
	columnSet = make(map[string]struct{}, len(fields))
	uniques = make([]CompositeUniqueParam, 0)

	for _, fieldName := range fieldsOrder {
		field, ok := fields[fieldName]
		if !ok || field == nil {
			continue
		}
		if err := validateFieldName(field); err != nil {
			return nil, nil, "", nil, errors.Wrapf(err, "buildDbMetadata: model '%s'", schemaName)
		}
		if field.IsVirtual() {
			continue
		}
		columnName := field.Name()
		columnSet[columnName] = struct{}{}

		if field.IsUnique() {
			// Field-level uniques carry no index name; the query builder derives one.
			uniques = append(uniques, CompositeUniqueParam{Fields: []string{columnName}})
		}
		if field.IsPrimaryKey() {
			primary = append(primary, columnName)
		}
		if field.IsTenantKey() {
			if tenant != "" && tenant != columnName {
				return nil, nil, "", nil, errors.Errorf(
					"buildDbMetadata: model '%s': field '%s' must not be a tenant key because '%s' is already one",
					schemaName, columnName, tenant)
			}
			tenant = columnName
		}
	}
	return columnSet, primary, tenant, uniques, nil
}

func validateCompositeUniquesForDb(
	schema *ModelSchema,
	columnSet map[string]struct{},
) ([]CompositeUniqueParam, error) {
	composites := schema.CompositeUniques()
	uniqueKeys := make([]CompositeUniqueParam, 0, len(composites))
	for _, composite := range composites {
		validated, err := validateCompositeUniqueKey(schema, columnSet, composite.Fields)
		if err != nil {
			return nil, err
		}
		if len(validated) > 0 {
			uniqueKeys = append(uniqueKeys, CompositeUniqueParam{
				IndexName: strings.TrimSpace(composite.IndexName),
				Fields:    validated,
			})
		}
	}
	return uniqueKeys, nil
}

func validateCompositeUniqueKey(
	schema *ModelSchema,
	columnSet map[string]struct{},
	compositeKey []string,
) ([]string, error) {
	if len(compositeKey) == 0 {
		return nil, nil
	}
	validated := make([]string, 0, len(compositeKey))
	for _, name := range compositeKey {
		trimmed := strings.TrimSpace(name)
		if trimmed == "" {
			continue
		}
		if _, ok := columnSet[trimmed]; !ok {
			return nil, errors.Errorf(
				"validateCompositeUniquesForDb: model '%s': unknown column reference '%s' in composite unique",
				schema.Name(), trimmed)
		}
		field := schema.fields[trimmed]
		if field != nil && !field.IsRequiredForCreate() {
			return nil, errors.Errorf(
				"validateCompositeUniquesForDb: model '%s': composite unique includes field '%s' which is not "+
					"requiredForCreate; use PartialUnique() instead",
				schema.Name(), trimmed)
		}
		validated = append(validated, trimmed)
	}
	return validated, nil
}

func ensurePartialUniqueColumnExists(
	schema *ModelSchema, columnSet map[string]struct{}, col string,
) error {
	if columnSet == nil {
		return nil
	}
	if _, ok := columnSet[col]; !ok {
		return errors.Errorf(
			"validatePartialUniquesForDb: model '%s': unknown column reference '%s' in partial unique",
			schema.Name(), col)
	}
	return nil
}

func validatePartialUniquesForDb(
	schema *ModelSchema,
	columnSet map[string]struct{},
) ([]PartialUniqueParam, error) {
	raw := schema.partialUniques
	out := make([]PartialUniqueParam, 0, len(raw))
	for _, param := range raw {
		validated, err := validatePartialUniqueParam(schema, columnSet, param)
		if err != nil {
			return nil, err
		}
		if len(validated.NotNullFields) > 0 {
			out = append(out, validated)
		}
	}
	return out, nil
}

func validatePartialUniqueParam(
	schema *ModelSchema,
	columnSet map[string]struct{},
	param PartialUniqueParam,
) (PartialUniqueParam, error) {
	nullableField := strings.TrimSpace(param.NullableField)
	if nullableField == "" {
		return PartialUniqueParam{}, nil
	}
	nullable, err := resolvePartialUniqueNullable(schema, columnSet, nullableField)
	if err != nil {
		return PartialUniqueParam{}, err
	}
	notNullFields, err := resolvePartialUniqueNotNulls(schema, columnSet, param.NotNullFields, nullableField)
	if err != nil {
		return PartialUniqueParam{}, err
	}
	if len(notNullFields) == 0 {
		return PartialUniqueParam{}, errors.Errorf(
			"validatePartialUniquesForDb: model '%s': partial unique on nullable field '%s' requires at least "+
				"one not-null field",
			schema.Name(), nullableField)
	}
	if nullable.IsRequiredForCreate() {
		return PartialUniqueParam{}, errors.Errorf(
			"validatePartialUniquesForDb: model '%s': nullable field '%s' must not be requiredForCreate",
			schema.Name(), nullableField)
	}
	return PartialUniqueParam{
		IndexName:     strings.TrimSpace(param.IndexName),
		NotNullFields: notNullFields,
		NullableField: nullableField,
	}, nil
}

func resolvePartialUniqueNullable(
	schema *ModelSchema, columnSet map[string]struct{}, nullableField string,
) (*ModelField, error) {
	if err := ensurePartialUniqueColumnExists(schema, columnSet, nullableField); err != nil {
		return nil, err
	}
	nullable, ok := schema.fields[nullableField]
	if !ok || nullable == nil {
		return nil, errors.Errorf(
			"validatePartialUniquesForDb: model '%s': unknown column '%s' in partial unique",
			schema.Name(), nullableField)
	}
	return nullable, nil
}

func resolvePartialUniqueNotNulls(
	schema *ModelSchema, columnSet map[string]struct{}, rawFields []string, nullableField string,
) ([]string, error) {
	notNullFields := make([]string, 0, len(rawFields))
	for _, name := range rawFields {
		trimmed := strings.TrimSpace(name)
		if trimmed == "" {
			continue
		}
		if err := ensurePartialUniqueColumnExists(schema, columnSet, trimmed); err != nil {
			return nil, err
		}
		field := schema.fields[trimmed]
		if field == nil {
			return nil, errors.Errorf(
				"validatePartialUniquesForDb: model '%s': unknown column '%s' in partial unique",
				schema.Name(), trimmed)
		}
		if trimmed == nullableField {
			return nil, errors.Errorf(
				"validatePartialUniquesForDb: model '%s': field '%s' cannot be both nullable and not-null "+
					"in the same partial unique",
				schema.Name(), trimmed)
		}
		if !field.IsRequiredForCreate() {
			return nil, errors.Errorf(
				"validatePartialUniquesForDb: model '%s': not-null field '%s' must be requiredForCreate",
				schema.Name(), trimmed)
		}
		notNullFields = append(notNullFields, trimmed)
	}
	return notNullFields, nil
}

func validateSearchIndexGroupsForDb(
	schema *ModelSchema,
	columnSet map[string]struct{},
) ([]SearchIndexGroupParam, error) {
	raw := schema.searchIndexGroups
	out := make([]SearchIndexGroupParam, 0, len(raw))
	for _, group := range raw {
		validated, err := validateSearchIndexGroup(schema, columnSet, group)
		if err != nil {
			return nil, err
		}
		if len(validated.Fields) > 0 {
			out = append(out, validated)
		}
	}
	return out, nil
}

func validateSearchIndexGroup(
	schema *ModelSchema,
	columnSet map[string]struct{},
	group SearchIndexGroupParam,
) (SearchIndexGroupParam, error) {
	fields := make([]string, 0, len(group.Fields))
	for _, name := range group.Fields {
		trimmed := strings.TrimSpace(name)
		if trimmed == "" {
			continue
		}
		if _, ok := columnSet[trimmed]; !ok {
			return SearchIndexGroupParam{}, errors.Errorf(
				"validateSearchIndexGroupsForDb: model '%s': unknown column '%s' in search index group",
				schema.Name(), trimmed)
		}
		fields = append(fields, trimmed)
	}
	if len(fields) == 0 {
		return SearchIndexGroupParam{}, nil
	}
	return SearchIndexGroupParam{
		IndexName: strings.TrimSpace(group.IndexName),
		Fields:    fields,
	}, nil
}
