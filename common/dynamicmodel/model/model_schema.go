package model

import (
	"fmt"
	"reflect"

	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/array"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/modelmapper"
)

type DynamicFields map[string]any

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

type ModelSchema struct {
	// Persistent fields
	name             string
	tableName        string
	label            model.LangJson
	description      model.LangJson
	fieldsOrder      []string
	compositeUniques [][]string
	primaryKeys      []string
	tenantKey        *string

	// Computed fields
	allUniqueKeys [][]string

	relations []ModelRelation
	fields    map[string]*ModelField
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

func (this ModelSchema) Relations() []ModelRelation {
	return this.relations
}

func (this ModelSchema) Field(name string) (*ModelField, bool) {
	field, ok := this.fields[name]
	return field, ok
}

func (this ModelSchema) MustField(name string) *ModelField {
	field := this.fields[name]
	return field
}

// TableName returns the table name associated with this schema.
func (this ModelSchema) TableName() string {
	return this.tableName
}

// UniqueFields returns the list of composite unique constraints.
// Each inner slice is a group of field names.
func (this ModelSchema) CompositeUniques() [][]string {
	return this.compositeUniques
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

// ValidateMap validates each map key against the corresponding schema field by invoking ModelField.Validate.
// Returns a new map with validated and sanitized values, or (nil, ClientErrors) when invalid.
func (this *ModelSchema) Validate(input DynamicFields, forEdit ...bool) (DynamicFields, ft.ClientErrors) {
	isForEdit := len(forEdit) > 0 && forEdit[0]

	var errs ft.ClientErrors
	result := make(map[string]any, len(input))

	for _, name := range this.fieldsOrder {
		field := this.fields[name]
		if field.isReadOnly && !isForEdit {
			input[name] = nil
		} else if field.isReadOnly && !this.IsPrimaryKey(name) && !this.IsVersioningKey(name) {
			continue
		}

		val, _ := input[name]
		result[name] = val
		validated, vErr := field.Validate(val, isForEdit)
		if vErr != nil {
			errs.Append(*vErr)
			continue
		}
		if !validated.Same(val) {
			result[name] = *validated.Get()
		}
	}

	if errs.Count() > 0 {
		return nil, errs
	}
	return result, nil
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

type ModelRelation struct {
	Edge           string          `json:"edge"`
	SrcField       string          `json:"src_field"`
	RelationType   RelationType    `json:"relation_type"`
	label          model.LangJson  `json:"label"`
	DestSchemaName string          `json:"dest_schema_name"`
	DestField      string          `json:"dest_field"`
	OnDelete       RelationCascade `json:"on_delete"`
	OnUpdate       RelationCascade `json:"on_update"`

	ThroughModel     *ModelSchema `json:"through_model,omitempty"`
	ThroughTableName string       `json:"through_table_name,omitempty"`
	ThroughSrcCol    string       `json:"through_foreign_col,omitempty"`
	ThroughDestCol   string       `json:"through_dest_col,omitempty"`
}

type RelationType string

const (
	RelationTypeOneToOne   = RelationType("one:one")
	RelationTypeOneToMany  = RelationType("one:many")
	RelationTypeManyToOne  = RelationType("many:one")
	RelationTypeManyToMany = RelationType("many:many")
)

type ModelField struct {
	name             string
	label            model.LangJson
	dataType         FieldDataType
	description      model.LangJson
	isReadOnly       bool
	isRequiredCreate bool
	isRequiredUpdate bool
	isVersioningKey  bool
	isPrimaryKey     bool
	isTenantKey      bool
	isUnique         bool
	rules            []*FieldRule
	defaultValue     *value
	defaultFn        func() any
	relation         *ModelRelation
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
	return IsModelDataType(this.dataType)
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

func (this *ModelField) IsReadOnly() bool {
	return this.isReadOnly
}

func (this *ModelField) IsRequiredForCreate() bool {
	return this.isRequiredCreate
}

func (this *ModelField) IsRequiredForUpdate() bool {
	return this.isRequiredUpdate
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

// ColumnType returns the SQL column type string (from DataType).
func (this *ModelField) ColumnType() string {
	return this.dataType.String()
}

// ColumnNullable returns "NOT NULL" if required, else "NULL".
func (this *ModelField) ColumnNullable() string {
	if this.isRequiredCreate {
		return "NOT NULL"
	}
	return "NULL"
}

// IsNullable returns true if the column allows NULL.
func (this *ModelField) IsNullable() bool {
	return !this.isRequiredCreate
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

	if isNil(val) {
		if isForEdit {
			if this.isRequiredUpdate {
				return Value(nil), ft.NewValidationError(this.name, "common.err_missing_required_field", "field is required")
			}
			return Value(val), nil
		}

		if this.defaultValue != nil {
			return *this.defaultValue, nil
		}
		if this.defaultFn != nil {
			return Value(this.defaultFn()), nil
		}
		if this.isRequiredCreate && !this.isReadOnly {
			return Value(nil), ft.NewValidationError(this.name, "common.err_missing_required_field", "field is required")
		}
		return Value(val), nil
	}

	if vErr := this.applyFieldRulesForValue(val); vErr != nil {
		vErr.Field = this.name
		return Value(nil), vErr
	}

	validated, vErr := this.dataType.Validate(Value(val))
	if vErr != nil {
		vErr.Field = this.name
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
	FieldDataTypeOptPrecision         = FieldDataTypeOptName("precision")
	FieldDataTypeOptSanitizeType      = FieldDataTypeOptName("sanitizeType")
	FieldDataTypeOptLangJsonWhitelist = FieldDataTypeOptName("langJsonWhitelist")
	FieldDataTypeOptLength            = FieldDataTypeOptName("length")
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

func (this *ModelField) Clone() *ModelField {
	cloned := &ModelField{
		name:             this.name,
		label:            this.label,
		dataType:         this.dataType,
		description:      this.description,
		isRequiredCreate: this.isRequiredCreate,
		isRequiredUpdate: this.isRequiredUpdate,
		isPrimaryKey:     this.isPrimaryKey,
		isTenantKey:      this.isTenantKey,
		isUnique:         this.isUnique,
		rules:            make([]*FieldRule, len(this.rules)),
		defaultValue:     this.defaultValue,
		defaultFn:        this.defaultFn,
		relation:         this.relation,
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
