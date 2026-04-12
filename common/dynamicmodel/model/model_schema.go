package model

import (
	"fmt"
	"reflect"

	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/array"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/json"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/modelmapper"
)

type DynamicFields map[string]any

// func (this DynamicFields) MarshalText() ([]byte, error) {
// 	// Temp type to avoid infinite recursion
// 	raw := map[string]any(this)
// 	return json.Marshal(raw)
// }

// Implements encoding.TextUnmarshaler interface
func (this *DynamicFields) UnmarshalText(text []byte) error {
	// Temp type to avoid infinite recursion
	raw := map[string]any{}
	if err := json.Unmarshal(text, &raw); err != nil {
		return err
	}
	*this = DynamicFields(raw)
	return nil
}

// Implements json.Unmarshaler interface
func (this *DynamicFields) UnmarshalJSON(data []byte) error {
	return this.UnmarshalText(data)
}

func (this *DynamicFields) Merge(data DynamicFields) error {
	for key, value := range data {
		(*this)[key] = value
	}
	return nil
}

func Value(val any) value {
	return value{val: &val}
}

// Value is used to pass `any` value through multiple function calls
// without causing nested pointer indirections. Eg: interface{}(interface{}(interface{}(bool)))
type value struct {
	val *any
}

func (this value) Get() *any {
	return this.val
}

func (this value) Same(another any) bool {
	if this.val == nil {
		return this.val == another
	}
	return *this.val == another
}

func (this value) IsEmpty() bool {
	return this.val == nil
}

type ModelSchema struct {
	// Persistent fields
	name                string
	tableName           string
	label               model.LangJson
	description         model.LangJson
	fieldsOrder         []string
	compositeUniques    [][]string
	partialUniques      [][]string
	partialUniqueGroups []PartialUniqueGroupParam
	searchIndexGroups   []SearchIndexGroupParam
	primaryKeys         []string
	tenantKey           *string

	// Computed fields
	allUniqueKeys [][]string

	// toRelations: EdgeTo and Field().Foreign — drive FK constraints on the owning table.
	toRelations []ModelRelation
	// fromRelations: EdgeFrom only — inverse edges; no extra FK DDL beyond the peer to-relation.
	fromRelations []ModelRelation
	fields        map[string]*ModelField

	// m2mPeerByDest maps peer (destination) schema name to resolved M2M metadata.
	// Populated by SchemaRegistry.FinalizeRelations. At most one edge per peer name.
	m2mPeerByDest map[string]*M2mPeerLink

	// m2mPeerByEdge maps relation edge name to the same resolved M2M metadata as m2mPeerByDest.
	m2mPeerByEdge map[string]*M2mPeerLink

	// exclusiveFieldGroups: each inner slice lists any number of field names (minimum two per group)
	// where exactly one must be non-empty in validated input. Zero or more than one non-empty
	// in that group yields client errors. Schemas may define multiple groups.
	exclusiveFieldGroups [][]string
}

// M2mPeerLink holds junction and FK-prefix metadata for a finalized many-to-many edge from the
// owning schema toward DestSchema (peer). Used by repositories to insert junction rows without a schema registry.
type M2mPeerLink struct {
	DestSchema      *ModelSchema
	ThroughSchema   *ModelSchema
	SrcFieldPrefix  string
	DestFieldPrefix string
	Edge            string
}

// M2mPeerLinkForDest returns the link for associating this schema with the given peer schema name.
func (this *ModelSchema) M2mPeerLinkForDest(destSchemaName string) (*M2mPeerLink, bool) {
	if this.m2mPeerByDest == nil {
		return nil, false
	}
	link, ok := this.m2mPeerByDest[destSchemaName]
	return link, ok
}

// M2mPeerLinkForEdge returns finalized many-to-many metadata for an outgoing M2M edge name on this schema.
func (this *ModelSchema) M2mPeerLinkForEdge(edge string) (*M2mPeerLink, bool) {
	if this.m2mPeerByEdge == nil {
		return nil, false
	}
	link, ok := this.m2mPeerByEdge[edge]
	return link, ok
}

func (this ModelSchema) Name() string {
	return this.name
}

func (this ModelSchema) Label() model.LangJson {
	return this.label
}

func (this ModelSchema) Description() model.LangJson {
	return this.description
}

func (this ModelSchema) Fields() map[string]*ModelField {
	return this.fields
}

func (this ModelSchema) ToRelations() []ModelRelation {
	return this.toRelations
}

func (this ModelSchema) FromRelations() []ModelRelation {
	return this.fromRelations
}

// Relations returns to-relations first, then from-relations (navigation, graph, legacy callers).
func (this ModelSchema) Relations() []ModelRelation {
	n := len(this.toRelations) + len(this.fromRelations)
	if n == 0 {
		return nil
	}
	out := make([]ModelRelation, 0, n)
	out = append(out, this.toRelations...)
	out = append(out, this.fromRelations...)
	return out
}

func (this ModelSchema) Field(name string) (*ModelField, bool) {
	field, ok := this.fields[name]
	return field, ok
}

func (this ModelSchema) MustField(name string) *ModelField {
	field, ok := this.Field(name)
	if !ok {
		panic(errors.Errorf("MustField: field '%s' not found in schema '%s'", name, this.name))
	}
	return field
}

// TableName returns the table name associated with this schema.
func (this ModelSchema) TableName() string {
	return this.tableName
}

// CompositeUniques returns schema-level composite UNIQUE constraints (all columns NOT NULL).
// Each inner slice is a group of field names.
func (this ModelSchema) CompositeUniques() [][]string {
	return this.compositeUniques
}

// PartialUniques returns pairs of field names for partial unique indexes: UNIQUE (required column)
// WHERE (nullable column) IS NULL. Only populated after ShouldBuildDb / populateDbMetadata validation.
func (this ModelSchema) PartialUniques() [][]string {
	return this.partialUniques
}

// PartialUniqueGroups returns grouped partial unique index definitions.
func (this ModelSchema) PartialUniqueGroups() []PartialUniqueGroupParam {
	return this.partialUniqueGroups
}

// SearchIndexGroups returns grouped CREATE INDEX definitions.
func (this ModelSchema) SearchIndexGroups() []SearchIndexGroupParam {
	return this.searchIndexGroups
}

// TenantKey returns the tenant key column name, or empty if not tenant-scoped.
func (this ModelSchema) TenantKey() string {
	if this.tenantKey == nil {
		return ""
	}
	return *this.tenantKey
}

// Column returns the field by name (alias for Field for ORM compatibility).
func (this ModelSchema) Column(name string) (*ModelField, bool) {
	return this.Field(name)
}

// KeyColumns returns primary keys plus tenant key if present.
func (this ModelSchema) KeyColumns() []string {
	keys := append([]string{}, this.primaryKeys...)
	if tk := this.TenantKey(); tk != "" && !array.Contains(keys, tk) {
		keys = append(keys, tk)
	}
	return keys
}

// IsVersioningKey returns true if the given field is an audit key.
func (this ModelSchema) IsVersioningKey(name string) bool {
	field, ok := this.fields[name]
	if !ok {
		return false
	}
	return field.IsVersioningKey()
}

// IsPrimaryKey returns true if the given field is a primary key.
func (this ModelSchema) IsPrimaryKey(name string) bool {
	return array.Contains(this.primaryKeys, name)
}

// IsTenantKey returns true if the given field is the tenant key.
func (this ModelSchema) IsTenantKey(name string) bool {
	return this.TenantKey() == name
}

// Columns returns fields in definition order for SQL operations.
// Model-typed fields (virtual edge fields) are excluded as they have no DB column.
func (this ModelSchema) Columns() []*ModelField {
	result := make([]*ModelField, 0, len(this.fieldsOrder))
	for _, name := range this.fieldsOrder {
		if f, ok := this.fields[name]; ok && f != nil && !f.IsVirtualModelField() {
			result = append(result, f)
		}
	}
	return result
}

// Picks creates a new instance of ModelSchema with only the specified fields.
// Other information such as table name, labels, descriptions, etc. are not copied.
func (this ModelSchema) Pick(fieldNames []string) *ModelSchema {
	newSchema := &ModelSchema{
		name:   this.name,
		fields: make(map[string]*ModelField),
	}
	for _, name := range fieldNames {
		field, ok := this.fields[name]
		if ok {
			newSchema.fields[name] = field
		}
	}
	return newSchema
}

// PrimaryKeys returns the list of primary key column names.
func (this ModelSchema) PrimaryKeys() []string {
	return this.primaryKeys
}

// UniqueKeys returns all unique constraints (field-level and schema-level).
func (this ModelSchema) AllUniques() [][]string {
	return this.allUniqueKeys
}

type ModelSchemaValidateOpts struct {
	// Whether to validate for edit or create (default).
	ForEdit bool
	// Whether to strip read-only fields from the sanitized result. Default: true.
	// StripReadOnly bool
	// Whether to set default values for auto-generated fields. Default: false.
	// AutoGenerateValues bool
}

func (this *ModelSchema) isSystemField(field *ModelField) bool {
	return field.IsAutoGenerated() && !this.IsPrimaryKey(field.name) && !this.IsVersioningKey(field.name)
}

func (this *ModelSchema) isNoUpdate(field *ModelField, forEdit bool) bool {
	isNoUpdate := field.IsNoUpdate()
	return isNoUpdate && forEdit
}

// ValidateMap validates each map key against the corresponding schema field by invoking ModelField.Validate.
// Returns a new map with validated and sanitized values, or (nil, ClientErrors) when invalid.
func (this *ModelSchema) Validate(input DynamicFields, forEdit ...bool) (DynamicFields, ft.ClientErrors) {
	isForEdit := len(forEdit) > 0 && forEdit[0]

	var errs ft.ClientErrors
	result := make(map[string]any, len(input))

	for _, name := range this.fieldsOrder {
		field := this.fields[name]
		if field.IsAutoGenerated() && !isForEdit {
			input[name] = nil // field.Validate() will populate value
		} else if this.isSystemField(field) || this.isNoUpdate(field, isForEdit) {
			continue
		}

		val, exists := input[name]
		if !exists && isForEdit && !this.IsVersioningKey(name) {
			continue
		}

		result[name] = val
		validated, vErr := field.Validate(val, isForEdit)
		if vErr != nil {
			errs.Append(*vErr)
			continue
		}
		if isNilOrEmpty(val) && field.requiredWithFieldName != "" {
			otherField := this.MustField(field.requiredWithFieldName)
			otherVal, ok := input[otherField.name]
			if !ok || isNilOrEmpty(otherVal) {
				errs.Append(*NewMissingFieldErr(field.name))
				continue
			}
		}
		if !validated.IsEmpty() {
			result[name] = *validated.Get()
		}
	}

	this.appendExclusiveFieldErrors(&errs, result)

	if errs.Count() > 0 {
		return nil, errs
	}
	return result, nil
}

func (this *ModelSchema) appendExclusiveFieldErrors(errs *ft.ClientErrors, result DynamicFields) {
	for _, group := range this.exclusiveFieldGroups {
		if len(group) < 2 {
			continue
		}
		presentCount := 0
		for _, name := range group {
			val, ok := result[name]
			if !ok {
				continue
			}
			if !isNilOrEmpty(val) {
				presentCount++
			}
		}
		if presentCount > 1 {
			errs.Append(*ft.NewExclusiveFieldsError(group))
		} else if presentCount == 0 {
			errs.Append(*ft.NewExclusiveFieldsMissingError(group))
		}
	}
}

// ValidateStruct validates a struct pointer by converting to map and validating.
// Uses "json" struct tag: missing tag uses field name, tag "-" skips the field.
func (this *ModelSchema) ValidateStruct(target any, forEdit ...bool) (any, ft.ClientErrors) {
	inputMap, err := modelmapper.StructToMap(&target)
	if err != nil {
		panic(errors.Wrap(err, "struct to map conversion failed"))
	}
	sanitizedFields, clientErrs := this.Validate(inputMap, forEdit...)
	err = modelmapper.MapToStruct(sanitizedFields, &target)
	if err != nil {
		panic(errors.Wrap(err, "map to struct reconstruction failed"))
	}
	return target, clientErrs
}

type RelationCascade string

const (
	RelationCascadeNoAction   = RelationCascade("NO ACTION")
	RelationCascadeSetNull    = RelationCascade("SET NULL")
	RelationCascadeSetDefault = RelationCascade("SET DEFAULT")
	RelationCascadeCascade    = RelationCascade("CASCADE")
)

// Sql returns the SQL keyword for this cascade action, defaulting to NO ACTION for the zero value.
func (this RelationCascade) Sql() string {
	if this == "" {
		return string(RelationCascadeNoAction)
	}
	return string(this)
}

// ForeignKeyColumnPair describes one column of a (possibly composite) foreign key.
// FkColumn is always on the table that owns the FK constraint; ReferencedColumn is on the referenced table.
type ForeignKeyColumnPair struct {
	FkColumn         string `json:"fk_column"`
	ReferencedColumn string `json:"referenced_column"`
}

type ModelRelation struct {
	Edge           string         `json:"edge"`
	SrcField       string         `json:"src_field"`
	RelationType   RelationType   `json:"relation_type"`
	label          model.LangJson `json:"label"`
	DestSchemaName string         `json:"dest_schema_name"`
	DestField      string         `json:"dest_field"`
	// ForeignKeys is the canonical multi-column FK. When empty, SrcField/DestField represent a single pair.
	ForeignKeys []ForeignKeyColumnPair `json:"foreign_keys,omitempty"`
	// UnvalidatedFkMap is consumed by SchemaRegistry.FinalizeRelations (src field name -> dest field name).
	UnvalidatedFkMap DynamicFields `json:"-"`
	// InversePeerSchemaName and InversePeerEdgeName are set by EdgeFrom / Existing() and cleared after finalize.
	InversePeerSchemaName string          `json:"inverse_peer_schema_name,omitempty"`
	InversePeerEdgeName   string          `json:"inverse_peer_edge_name,omitempty"`
	OnDelete              RelationCascade `json:"on_delete"`
	OnUpdate              RelationCascade `json:"on_update"`

	M2mThroughModel      *ModelSchema `json:"through_model,omitempty"`
	M2mThroughSchemaName string       `json:"through_table_name,omitempty"`
	// M2mSrcFieldPrefix is this (src) schema's junction-table FK prefix, not a single physical column name.
	// JOINs use PrefixedThroughColumn(M2mSrcFieldPrefix, pk) for each entry in this schema's PrimaryKeys(),
	// and PrefixedThroughColumn(M2mSrcFieldPrefix, tenantKey) when a tenant key exists (e.g. user -> user_id,
	// user_tenant_id). The peer (dest) side uses DestFieldPrefix the same way.
	M2mSrcFieldPrefix string `json:"src_field_prefix,omitempty"`
	// M2mDestFieldPrefix is the peer (dest) schema's junction FK prefix; set by FinalizeRelations.
	M2mDestFieldPrefix string `json:"dest_field_prefix,omitempty"`
}

type RelationType string

const (
	RelationTypeOneToOne   = RelationType("one:one")
	RelationTypeOneToMany  = RelationType("one:many")
	RelationTypeManyToOne  = RelationType("many:one")
	RelationTypeManyToMany = RelationType("many:many")
)

type ModelField struct {
	name            string
	label           model.LangJson
	dataType        FieldDataType
	description     model.LangJson
	isAutoGenerated bool
	// Determines the "NOT NULL" constraint for the database column,
	// and causes the field to be required for create operations.
	isRequiredForCreate bool
	// Causes the field to be required for update operations,
	// but doesn't affect the generated CREATE SQL query.
	isRequiredForUpdate bool
	// Causes the field to be required when the other field is present.
	requiredWithFieldName string
	isVersioningKey       bool
	isPrimaryKey          bool
	isTenantKey           bool
	isUnique              bool
	// Allows setting value on create but not on update.
	noUpdate       bool
	rules          []*FieldRule
	defaultValue   *value
	defaultFn      func() any
	useTypeDefault bool
}

// Getter methods
func (this *ModelField) Name() string {
	return this.name
}

func (this *ModelField) Label() model.LangJson {
	return this.label
}

func (this *ModelField) DataType() FieldDataType {
	return this.dataType
}

// IsVirtualModelField is true for FieldDataTypeModel (in-app only; no database column).
func (this *ModelField) IsVirtualModelField() bool {
	if this == nil || this.dataType == nil {
		return false
	}
	return IsFieldDataTypeModel(this.dataType)
}

func (this *ModelField) Description() model.LangJson {
	return this.description
}

func (this *ModelField) IsArray() bool {
	return this.dataType.IsArray()
}

func (this *ModelField) IsVersioningKey() bool {
	return this.isVersioningKey
}

func (this *ModelField) IsAutoGenerated() bool {
	return this.isAutoGenerated
}

func (this *ModelField) IsRequiredForCreate() bool {
	return this.isRequiredForCreate
}

func (this *ModelField) IsRequiredForUpdate() bool {
	return this.isRequiredForUpdate
}

func (this *ModelField) IsPrimaryKey() bool {
	return this.isPrimaryKey
}

func (this *ModelField) IsTenantKey() bool {
	return this.isTenantKey
}

func (this *ModelField) IsUnique() bool {
	return this.isUnique
}

func (this *ModelField) IsNoUpdate() bool {
	return this.noUpdate
}

// ColumnType returns the SQL column type string (from DataType).
func (this *ModelField) ColumnType() string {
	return this.dataType.String()
}

// ColumnNullable returns "NOT NULL" if required, else "NULL".
func (this *ModelField) ColumnNullable() string {
	if this.isRequiredForCreate {
		return "NOT NULL"
	}
	return "NULL"
}

// IsNullable returns true if the column allows NULL.
func (this *ModelField) IsNullable() bool {
	return !this.isRequiredForCreate
}

func (this *ModelField) Rules() []*FieldRule {
	return this.rules
}

func (this *ModelField) Default() *value {
	return this.defaultValue
}

func (this *ModelField) DefaultFn() func() any {
	return this.defaultFn
}

// Validate invokes the field's data type Validate (which validates and may sanitize),
// then applies field rules. Returns the validated value and technical error if any.
// When value is empty: uses default if available; otherwise errors only when required with no fallback.
func (this *ModelField) Validate(val any, forEdit ...bool) (value, *ft.ClientErrorItem) {
	isForEdit := len(forEdit) > 0 && forEdit[0]

	var wrappedVal value
	if isNilOrEmpty(val) {
		if isForEdit {
			if this.IsRequiredForUpdate() {
				return Value(nil), NewMissingFieldErr(this.name)
			}
			return Value(val), nil
		}

		if this.defaultValue != nil && !this.defaultValue.IsEmpty() {
			wrappedVal = *this.defaultValue
		} else if this.defaultFn != nil {
			wrappedVal = Value(this.defaultFn())
		} else if this.useTypeDefault {
			wrappedVal = this.dataType.DefaultValue()
		} else if this.IsRequiredForCreate() && !this.IsAutoGenerated() {
			return Value(nil), NewMissingFieldErr(this.name)
		}

		if wrappedVal.IsEmpty() {
			return Value(val), nil
		}
	} else {
		wrappedVal = Value(val)
	}

	if vErr := this.applyFieldRulesForValue(*wrappedVal.Get()); vErr != nil {
		vErr.Field = this.name
		return Value(nil), vErr
	}

	validated, vErr := this.dataType.Validate(wrappedVal)
	if vErr != nil {
		vErr.Field = this.name + vErr.Field
		return Value(nil), vErr
	}
	return validated, nil
}

func (this *ModelField) applyFieldRulesForValue(value any) *ft.ClientErrorItem {
	for _, rule := range this.rules {
		if rule == nil || len(*rule) == 0 {
			continue
		}

		if vErr := applyFieldRuleForValue(value, rule); vErr != nil {
			return vErr
		}
	}
	return nil
}

func isNil(val any) bool {
	// Check for literal `nil`.
	if val == nil {
		return true
	}
	v := reflect.ValueOf(val)
	switch v.Kind() {
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	}
	return false
}

func applyFieldRuleForValue(value any, rule *FieldRule) *ft.ClientErrorItem {
	ruleName := rule.RuleName()
	opts := rule.RuleOptions()
	var vErr *ft.ClientErrorItem
	switch ruleName {
	case FieldRuleArrayLengthType:
		vErr = ValidateArrayLength(value, opts)
	default:
		return nil
	}
	return vErr
}

type FieldDataTypeOptName string
type FieldDataTypeOptions map[FieldDataTypeOptName]any

const (
	FieldDataTypeOptEnumValues        = FieldDataTypeOptName("enumValues")
	FieldDataTypeOptLangJsonWhitelist = FieldDataTypeOptName("langJsonWhitelist")
	FieldDataTypeOptLength            = FieldDataTypeOptName("length")
	FieldDataTypeOptPattern           = FieldDataTypeOptName("pattern")
	FieldDataTypeOptRange             = FieldDataTypeOptName("range")
	FieldDataTypeOptSanitizeType      = FieldDataTypeOptName("sanitizeType")
	FieldDataTypeOptScale             = FieldDataTypeOptName("scale")
)

type SanitizeType string

const (
	SanitizeTypeNone      = SanitizeType("none")
	SanitizeTypeHtml      = SanitizeType("html")
	SanitizeTypePlainText = SanitizeType("plaintext")
)

type FieldRule []any

func (this FieldRule) RuleName() FieldRuleName {
	if len(this) == 0 {
		return ""
	}
	if name, ok := this[0].(FieldRuleName); ok {
		return name
	}
	return FieldRuleName(fmt.Sprint(this[0]))
}

func (this FieldRule) RuleOptions() any {
	if len(this) < 2 {
		return nil
	}
	return this[1]
}

type FieldRuleName string

const (
	FieldRuleArrayLengthType = FieldRuleName("arrlength")
)

func FieldRuleArrayLength(min, max int) FieldRule {
	return FieldRule{FieldRuleArrayLengthType, []int{min, max}}
}

// PrefixedThroughColumn returns junction column name prefix_fieldName (e.g. user + id -> user_id).
func PrefixedThroughColumn(prefix, fieldName string) string {
	return prefix + "_" + fieldName
}

func (this *ModelField) Clone() *ModelField {
	cloned := &ModelField{
		name:                this.name,
		label:               this.label,
		dataType:            this.dataType,
		description:         this.description,
		isRequiredForCreate: this.isRequiredForCreate,
		isRequiredForUpdate: this.isRequiredForUpdate,
		isAutoGenerated:     this.isAutoGenerated,
		isPrimaryKey:        this.isPrimaryKey,
		isTenantKey:         this.isTenantKey,
		isUnique:            this.isUnique,
		rules:               make([]*FieldRule, len(this.rules)),
		defaultValue:        this.defaultValue,
		defaultFn:           this.defaultFn,
		useTypeDefault:      this.useTypeDefault,
	}
	copy(cloned.rules, this.rules)
	return cloned
}

// Copy creates a new instance of ModelField with the same name, data type, and rules;
// other properties are not copied.
func (this *ModelField) Copy() *ModelField {
	copied := &ModelField{
		name:     this.name,
		dataType: this.dataType,
		rules:    make([]*FieldRule, len(this.rules)),
	}
	copy(copied.rules, this.rules)
	return copied
}
