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
	validatedPartialGroups, err := validatePartialUniqueGroupsForDb(schema, columnSet)
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
	schema.partialUniqueGroups = validatedPartialGroups
	schema.searchIndexGroups = validatedSearchIndexGroups
	schema.allUniqueKeys = append(fieldUnique, schemaUnique...)
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
) (columnSet map[string]struct{}, primary []string, tenant string, uniques [][]string, err error) {
	columnSet = make(map[string]struct{}, len(fields))
	uniques = make([][]string, 0)

	for _, fieldName := range fieldsOrder {
		field, ok := fields[fieldName]
		if !ok || field == nil {
			continue
		}
		if err := validateFieldName(field); err != nil {
			return nil, nil, "", nil, errors.Wrapf(err, "buildDbMetadata: model '%s'", schemaName)
		}
		if field.IsVirtualModelField() {
			continue
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
) ([][]string, error) {
	uniqueFields := schema.CompositeUniques()
	uniqueKeys := make([][]string, 0, len(uniqueFields))
	for _, compositeKey := range uniqueFields {
		validated, err := validateCompositeUniqueKey(schema, columnSet, compositeKey)
		if err != nil {
			return nil, err
		}
		if len(validated) > 0 {
			uniqueKeys = append(uniqueKeys, validated)
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

func validatePartialUniquesForDb(
	schema *ModelSchema,
	columnSet map[string]struct{},
) ([][]string, error) {
	raw := schema.partialUniques
	out := make([][]string, 0, len(raw))
	for _, pair := range raw {
		validated, err := validatePartialUniquePair(schema, columnSet, pair)
		if err != nil {
			return nil, err
		}
		if len(validated) == 2 {
			out = append(out, validated)
		}
	}
	return out, nil
}

func validatePartialUniquePair(
	schema *ModelSchema,
	columnSet map[string]struct{},
	pair []string,
) ([]string, error) {
	if len(pair) == 0 {
		return nil, nil
	}
	a, b, err := parsePartialUniquePairNames(schema, pair)
	if err != nil {
		return nil, err
	}
	if err := ensurePartialUniqueColumnsExist(schema, columnSet, a, b); err != nil {
		return nil, err
	}
	return finalizePartialUniquePair(schema, a, b)
}

func parsePartialUniquePairNames(schema *ModelSchema, pair []string) (string, string, error) {
	if len(pair) != 2 {
		return "", "", errors.Errorf(
			"validatePartialUniquesForDb: model '%s': PartialUnique supports exactly two fields per key",
			schema.Name())
	}
	a := strings.TrimSpace(pair[0])
	b := strings.TrimSpace(pair[1])
	if a == "" || b == "" {
		return "", "", errors.Errorf(
			"validatePartialUniquesForDb: model '%s': PartialUnique requires two non-empty field names",
			schema.Name())
	}
	return a, b, nil
}

func ensurePartialUniqueColumnsExist(
	schema *ModelSchema, columnSet map[string]struct{}, a, b string,
) error {
	for _, col := range []string{a, b} {
		if _, ok := columnSet[col]; !ok {
			return errors.Errorf(
				"validatePartialUniquesForDb: model '%s': unknown column reference '%s' in partial unique",
				schema.Name(), col)
		}
	}
	return nil
}

func finalizePartialUniquePair(schema *ModelSchema, a, b string) ([]string, error) {
	fa := schema.fields[a]
	fb := schema.fields[b]
	if fa == nil || fb == nil {
		return nil, errors.Errorf(
			"validatePartialUniquesForDb: model '%s': missing field for partial unique", schema.Name())
	}
	ra := fa.IsRequiredForCreate()
	rb := fb.IsRequiredForCreate()
	if ra && rb {
		return nil, errors.Errorf(
			"validatePartialUniquesForDb: model '%s': partial unique on '%s' and '%s': both are requiredForCreate; "+
				"use CompositeUnique() instead",
			schema.Name(), a, b)
	}
	if !ra && !rb {
		return nil, errors.Errorf(
			"validatePartialUniquesForDb: model '%s': partial unique on '%s' and '%s': one field must be "+
				"requiredForCreate, the other must not",
			schema.Name(), a, b)
	}
	return []string{a, b}, nil
}

func validatePartialUniqueGroupsForDb(
	schema *ModelSchema,
	columnSet map[string]struct{},
) ([]PartialUniqueGroupParam, error) {
	raw := schema.partialUniqueGroups
	legacyPartial := schema.partialUniques
	out := make([]PartialUniqueGroupParam, 0, len(raw))
	for _, pair := range legacyPartial {
		group, err := partialUniquePairToGroup(schema, pair)
		if err != nil {
			return nil, err
		}
		if len(group.NotNullFields) > 0 {
			raw = append(raw, group)
		}
	}
	for _, group := range raw {
		validated, err := validatePartialUniqueGroup(schema, columnSet, group)
		if err != nil {
			return nil, err
		}
		if len(validated.NotNullFields) > 0 {
			out = append(out, validated)
		}
	}
	return out, nil
}

func partialUniquePairToGroup(schema *ModelSchema, pair []string) (PartialUniqueGroupParam, error) {
	a, b, err := parsePartialUniquePairNames(schema, pair)
	if err != nil {
		return PartialUniqueGroupParam{}, err
	}
	validatedPair, err := finalizePartialUniquePair(schema, a, b)
	if err != nil {
		return PartialUniqueGroupParam{}, err
	}
	if len(validatedPair) != 2 {
		return PartialUniqueGroupParam{}, nil
	}
	return PartialUniqueGroupParam{
		NotNullFields: []string{validatedPair[0]},
		NullableField: validatedPair[1],
	}, nil
}

func validatePartialUniqueGroup(
	schema *ModelSchema,
	columnSet map[string]struct{},
	group PartialUniqueGroupParam,
) (PartialUniqueGroupParam, error) {
	indexName := strings.TrimSpace(group.IndexName)
	nullableField := strings.TrimSpace(group.NullableField)
	if nullableField == "" {
		return PartialUniqueGroupParam{}, nil
	}
	if columnSet != nil {
		if err := ensurePartialUniqueColumnsExist(schema, columnSet, nullableField, nullableField); err != nil {
			return PartialUniqueGroupParam{}, err
		}
	}
	nullable, ok := schema.fields[nullableField]
	if !ok || nullable == nil {
		return PartialUniqueGroupParam{}, errors.Errorf(
			"validatePartialUniqueGroupsForDb: model '%s': unknown column '%s' in partial unique group",
			schema.Name(), nullableField)
	}
	notNullFields := make([]string, 0, len(group.NotNullFields))
	for _, name := range group.NotNullFields {
		trimmed := strings.TrimSpace(name)
		if trimmed == "" {
			continue
		}
		if columnSet != nil {
			if err := ensurePartialUniqueColumnsExist(schema, columnSet, trimmed, trimmed); err != nil {
				return PartialUniqueGroupParam{}, err
			}
		}
		field := schema.fields[trimmed]
		if field == nil {
			return PartialUniqueGroupParam{}, errors.Errorf(
				"validatePartialUniqueGroupsForDb: model '%s': unknown column '%s' in partial unique group",
				schema.Name(), trimmed)
		}
		if trimmed == nullableField {
			return PartialUniqueGroupParam{}, errors.Errorf(
				"validatePartialUniqueGroupsForDb: model '%s': field '%s' cannot be both nullable and not-null in the same partial unique group",
				schema.Name(), trimmed)
		}
		notNullFields = append(notNullFields, trimmed)
	}
	if len(notNullFields) == 1 {
		notNullField := schema.fields[notNullFields[0]]
		// Preserve legacy PartialUnique(a,b) behavior: order is auto-normalized.
		if notNullField != nil && !notNullField.IsRequiredForCreate() && nullable.IsRequiredForCreate() {
			nullableField, notNullFields[0] = notNullFields[0], nullableField
			nullable = schema.fields[nullableField]
			notNullField = schema.fields[notNullFields[0]]
		}
		if notNullField == nil || !notNullField.IsRequiredForCreate() {
			return PartialUniqueGroupParam{}, errors.Errorf(
				"validatePartialUniqueGroupsForDb: model '%s': not-null field '%s' must be requiredForCreate",
				schema.Name(), notNullFields[0])
		}
		if nullable == nil || nullable.IsRequiredForCreate() {
			return PartialUniqueGroupParam{}, errors.Errorf(
				"validatePartialUniqueGroupsForDb: model '%s': nullable field '%s' must not be requiredForCreate",
				schema.Name(), nullableField)
		}
		return PartialUniqueGroupParam{
			IndexName:     indexName,
			NotNullFields: notNullFields,
			NullableField: nullableField,
		}, nil
	}
	if nullable.IsRequiredForCreate() {
		return PartialUniqueGroupParam{}, errors.Errorf(
			"validatePartialUniqueGroupsForDb: model '%s': nullable field '%s' must not be requiredForCreate",
			schema.Name(), nullableField)
	}
	for _, trimmed := range notNullFields {
		field := schema.fields[trimmed]
		if field == nil || !field.IsRequiredForCreate() {
			return PartialUniqueGroupParam{}, errors.Errorf(
				"validatePartialUniqueGroupsForDb: model '%s': not-null field '%s' must be requiredForCreate",
				schema.Name(), trimmed)
		}
	}
	if len(notNullFields) == 0 {
		return PartialUniqueGroupParam{}, errors.Errorf(
			"validatePartialUniqueGroupsForDb: model '%s': partial unique group '%s' requires at least one not-null field",
			schema.Name(), indexName)
	}
	return PartialUniqueGroupParam{
		IndexName:     indexName,
		NotNullFields: notNullFields,
		NullableField: nullableField,
	}, nil
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
