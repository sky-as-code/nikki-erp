package schema

import (
	"fmt"
	"reflect"

	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/array"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
)

const SchemaFieldTag = "entity"

type DynamicEntity map[string]any

type EntitySchema struct {
	// Persistent fields
	name             string         `json:"name"`
	tableName        string         `json:"table_name"`
	label            model.LangJson `json:"label"`
	description      model.LangJson `json:"description"`
	fieldsOrder      []string       `json:"fields_order"`
	compositeUniques [][]string     `json:"unique_fields"`
	primaryKeys      []string
	tenantKey        *string

	// Computed fields
	allUniqueKeys [][]string

	relations []EntityRelation        `json:"relations"`
	fields    map[string]*EntityField `json:"fields"`
}

func (this EntitySchema) Name() string {
	return this.name
}

func (this EntitySchema) Label() model.LangJson {
	return this.label
}

func (this EntitySchema) Description() model.LangJson {
	return this.description
}

func (this EntitySchema) Fields() map[string]*EntityField {
	return this.fields
}

func (this EntitySchema) Relations() []EntityRelation {
	return this.relations
}

func (this EntitySchema) Field(name string) (*EntityField, bool) {
	field, ok := this.fields[name]
	return field, ok
}

// TableName returns the table name associated with this schema.
func (this EntitySchema) TableName() string {
	return this.tableName
}

// UniqueFields returns the list of composite unique constraints.
// Each inner slice is a group of field names.
func (this EntitySchema) CompositeUniques() [][]string {
	return this.compositeUniques
}

// TenantKey returns the tenant key column name, or empty if not tenant-scoped.
func (this EntitySchema) TenantKey() string {
	if this.tenantKey == nil {
		return ""
	}
	return *this.tenantKey
}

// Column returns the field by name (alias for Field for ORM compatibility).
func (this EntitySchema) Column(name string) (*EntityField, bool) {
	return this.Field(name)
}

// KeyColumns returns primary keys plus tenant key if present.
func (this EntitySchema) KeyColumns() []string {
	keys := append([]string{}, this.primaryKeys...)
	if tk := this.TenantKey(); tk != "" && !array.Contains(keys, tk) {
		keys = append(keys, tk)
	}
	return keys
}

// IsPrimaryKey returns true if the given field is a primary key.
func (this EntitySchema) IsPrimaryKey(name string) bool {
	return array.Contains(this.primaryKeys, name)
}

// IsTenantKey returns true if the given field is the tenant key.
func (this EntitySchema) IsTenantKey(name string) bool {
	return this.TenantKey() == name
}

// Columns returns fields in definition order for SQL operations.
func (this EntitySchema) Columns() []*EntityField {
	result := make([]*EntityField, 0, len(this.fieldsOrder))
	for _, name := range this.fieldsOrder {
		if f, ok := this.fields[name]; ok && f != nil {
			result = append(result, f)
		}
	}
	return result
}

// PrimaryKeys returns the list of primary key column names.
func (this EntitySchema) PrimaryKeys() []string {
	return this.primaryKeys
}

// UniqueKeys returns all unique constraints (field-level and schema-level).
func (this EntitySchema) AllUniques() [][]string {
	return this.allUniqueKeys
}

// ValidateMap validates each map key against the corresponding schema field by invoking EntityField.Validate.
// Returns a new map with validated and sanitized values, or (nil, ClientErrors) when invalid.
func (this *EntitySchema) Validate(input DynamicEntity, forEdit ...bool) (DynamicEntity, *ft.ClientErrors) {
	isForEdit := len(forEdit) > 0 && forEdit[0]
	var errs ft.ClientErrors
	result := make(map[string]any, len(input))

	for key, field := range this.fields {
		val, exists := input[key]
		validated, vErr := field.Validate(val, isForEdit)
		if vErr != nil {
			errs.Append(*vErr)
			continue
		}
		if exists {
			result[key] = validated
		}
	}

	if errs.Count() > 0 {
		return nil, &errs
	}
	return result, nil
}

// ValidateStruct validates a struct pointer by converting to map and validating.
// Uses "json" struct tag: missing tag uses field name, tag "-" skips the field.
func (this *EntitySchema) ValidateStruct(structPtr any, forEdit ...bool) *ft.ClientErrors {
	inputMap, err := StructToDynamicEntity(structPtr)
	if err != nil {
		panic(errors.Wrap(err, "struct to map conversion failed"))
	}
	_, clientErrs := this.Validate(inputMap, forEdit...)
	return clientErrs
}

type EntityRelation struct {
	Edge           string        `json:"edge"`
	SrcField       string        `json:"src_field"`
	RelationType   RelationType  `json:"relation_type"`
	DestEntityName string        `json:"dest_entity_name"`
	DestEntity     *EntitySchema `json:"dest_entity"`
	DestField      string        `json:"dest_field"`

	ThroughEntity    *EntitySchema `json:"through_entity,omitempty"`
	ThroughTableName string        `json:"through_table_name,omitempty"`
	ThroughSrcCol    string        `json:"through_foreign_col,omitempty"`
	ThroughDestCol   string        `json:"through_dest_col,omitempty"`
}

type RelationType string

const (
	RelationTypeOneToOne   = RelationType("one:one")
	RelationTypeManyToOne  = RelationType("many:one")
	RelationTypeManyToMany = RelationType("many:many")
)

type EntityField struct {
	name         string
	label        model.LangJson
	dataType     FieldDataType
	description  model.LangJson
	isArray      bool
	isRequired   bool
	isPrimaryKey bool
	isTenantKey  bool
	isUnique     bool
	rules        []*FieldRule
	defaultValue *any
	relation     *EntityRelation
}

// Getter methods
func (this *EntityField) Name() string {
	return this.name
}

func (this *EntityField) Label() model.LangJson {
	return this.label
}

func (this *EntityField) DataType() FieldDataType {
	return this.dataType
}

func (this *EntityField) Description() model.LangJson {
	return this.description
}

func (this *EntityField) IsArray() bool {
	return this.isArray
}

func (this *EntityField) IsRequired() bool {
	return this.isRequired
}

func (this *EntityField) IsPrimaryKey() bool {
	return this.isPrimaryKey
}

func (this *EntityField) IsTenantKey() bool {
	return this.isTenantKey
}

func (this *EntityField) IsUnique() bool {
	return this.isUnique
}

// ColumnType returns the SQL column type string (from DataType).
func (this *EntityField) ColumnType() string {
	return this.dataType.String()
}

// ColumnNullable returns "NOT NULL" if required, else "NULL".
func (this *EntityField) ColumnNullable() string {
	if this.isRequired {
		return "NOT NULL"
	}
	return "NULL"
}

// IsNullable returns true if the column allows NULL.
func (this *EntityField) IsNullable() bool {
	return !this.isRequired
}

func (this *EntityField) Rules() []*FieldRule {
	return this.rules
}

func (this *EntityField) Default() any {
	if this.defaultValue == nil {
		return nil
	}
	return *this.defaultValue
}

// Validate invokes the field's data type Validate (which validates and may sanitize),
// then applies field rules. Returns the validated value and technical error if any.
// When value is empty: uses default if available; otherwise errors only when required with no fallback.
func (this *EntityField) Validate(value any, forceOptional ...bool) (any, *ft.ClientErrorItem) {
	isForcedOptional := len(forceOptional) > 0 && forceOptional[0]

	if isEmptyValue(value) {
		if this.defaultValue != nil {
			return *this.defaultValue, nil
		}
		if this.isRequired && !isForcedOptional {
			return nil, &ft.ClientErrorItem{
				Field: this.name, Key: "common.err_missing_required_field", Message: "field is required", Vars: nil,
			}
		}
		return value, nil
	}
	validated, vErr := this.dataType.Validate(value)
	if vErr != nil {
		vErr.Field = this.name
		return nil, vErr
	}
	if vErr := this.applyFieldRulesForValue(validated); vErr != nil {
		vErr.Field = this.name
		return nil, vErr
	}
	return validated, nil
}

func (this *EntityField) applyFieldRulesForValue(value any) *ft.ClientErrorItem {
	for _, rule := range this.rules {
		if rule == nil || len(*rule) == 0 {
			continue
		}
		ruleName := rule.RuleName()
		if isCustomDataType(this.dataType) && ruleName == FieldRuleLengthType {
			continue
		}
		if vErr := applyFieldRuleForValue(value, rule); vErr != nil {
			return vErr
		}
	}
	return nil
}

func isCustomDataType(dt FieldDataType) bool {
	customNames := []string{"nikkiEtag", "nikkiLangJson", "nikkiLangCode", "nikkiModelId", "nikkiSlug"}
	for _, name := range customNames {
		if dt.String() == name {
			return true
		}
	}
	return false
}

func isEmptyValue(val any) bool {
	// Check for literal `nil`.
	if val == nil {
		return true
	}
	v := reflect.ValueOf(val)
	switch v.Kind() {
	case reflect.String:
		return v.Len() == 0
	case reflect.Slice, reflect.Map:
		return v.Len() == 0
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
	case FieldRuleMaxType:
		vErr = ValidateMax(value, opts)
	case FieldRuleMinType:
		vErr = ValidateMin(value, opts)
	case FieldRuleLengthType:
		vErr = ValidateLength(value, opts)
	case FieldRuleOneOfType:
		vErr = ValidateOneOf(value, opts)
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
)

type SanitizeType string

const (
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
	FieldRuleMaxType         = FieldRuleName("max")
	FieldRuleMinType         = FieldRuleName("min")
	FieldRuleArrayLengthType = FieldRuleName("arrlength")
	FieldRuleLengthType      = FieldRuleName("length")
	FieldRuleOneOfType       = FieldRuleName("oneOf")
)

func FieldRuleMax(value any) FieldRule {
	return FieldRule{FieldRuleMaxType, value}
}

func FieldRuleMin(value any) FieldRule {
	return FieldRule{FieldRuleMinType, value}
}

func FieldRuleArrayLength(min, max int) FieldRule {
	return FieldRule{FieldRuleArrayLengthType, []int{min, max}}
}

func FieldRuleLength(min, max int) FieldRule {
	return FieldRule{FieldRuleLengthType, []int{min, max}}
}

func FieldRuleOneOf(values ...any) FieldRule {
	return FieldRule{FieldRuleOneOfType, values}
}

func (this *EntityField) Clone() *EntityField {
	cloned := &EntityField{
		name:         this.name,
		label:        this.label,
		dataType:     this.dataType,
		description:  this.description,
		isArray:      this.isArray,
		isRequired:   this.isRequired,
		isPrimaryKey: this.isPrimaryKey,
		isTenantKey:  this.isTenantKey,
		isUnique:     this.isUnique,
		rules:        make([]*FieldRule, len(this.rules)),
		defaultValue: this.defaultValue,
		relation:     this.relation,
	}
	copy(cloned.rules, this.rules)
	return cloned
}
