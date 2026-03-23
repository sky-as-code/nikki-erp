package schema

import (
	"fmt"
	"reflect"

	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/array"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
)

type DynamicFields map[string]any

type EntitySchema struct {
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

	relations []EntityRelation
	fields    map[string]*EntityField
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
// Entity-typed fields (virtual edge fields) are excluded as they have no DB column.
func (this EntitySchema) Columns() []*EntityField {
	result := make([]*EntityField, 0, len(this.fieldsOrder))
	for _, name := range this.fieldsOrder {
		if f, ok := this.fields[name]; ok && f != nil && !isEntityField(f) {
			result = append(result, f)
		}
	}
	return result
}

func isEntityField(f *EntityField) bool {
	return f.dataType != nil && f.dataType.String() == "entity"
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
func (this *EntitySchema) Validate(input DynamicFields, forEdit ...bool) (DynamicFields, *ft.ClientErrors) {
	isForEdit := len(forEdit) > 0 && forEdit[0]
	var errs ft.ClientErrors
	result := make(map[string]any, len(input))

	for _, name := range this.fieldsOrder {
		field := this.fields[name]
		// for key, field := range this.fields {
		val, exists := input[name]
		validated, vErr := field.Validate(val, isForEdit)
		if vErr != nil {
			errs.Append(*vErr)
			continue
		}
		if exists || validated != val {
			result[name] = validated
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

type EntityRelation struct {
	Edge           string          `json:"edge"`
	SrcField       string          `json:"src_field"`
	RelationType   RelationType    `json:"relation_type"`
	label          model.LangJson  `json:"label"`
	DestEntityName string          `json:"dest_entity_name"`
	DestField      string          `json:"dest_field"`
	OnDelete       RelationCascade `json:"on_delete"`
	OnUpdate       RelationCascade `json:"on_update"`

	ThroughEntity    *EntitySchema `json:"through_entity,omitempty"`
	ThroughTableName string        `json:"through_table_name,omitempty"`
	ThroughSrcCol    string        `json:"through_foreign_col,omitempty"`
	ThroughDestCol   string        `json:"through_dest_col,omitempty"`
}

type RelationType string

const (
	RelationTypeOneToOne   = RelationType("one:one")
	RelationTypeOneToMany  = RelationType("one:many")
	RelationTypeManyToOne  = RelationType("many:one")
	RelationTypeManyToMany = RelationType("many:many")
)

type EntityField struct {
	name             string
	label            model.LangJson
	dataType         FieldDataType
	description      model.LangJson
	isArray          bool
	isAuto           bool
	isRequiredCreate bool
	isRequiredUpdate bool
	isPrimaryKey     bool
	isTenantKey      bool
	isUnique         bool
	rules            []*FieldRule
	defaultValue     *any
	defaultFn        func() any
	relation         *EntityRelation
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

func (this *EntityField) IsAutoGenerated() bool {
	return this.isAuto
}

func (this *EntityField) IsRequiredForCreate() bool {
	return this.isRequiredCreate
}

func (this *EntityField) IsRequiredForUpdate() bool {
	return this.isRequiredUpdate
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
	if this.isRequiredCreate {
		return "NOT NULL"
	}
	return "NULL"
}

// IsNullable returns true if the column allows NULL.
func (this *EntityField) IsNullable() bool {
	return !this.isRequiredCreate
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

func (this *EntityField) DefaultFn() func() any {
	return this.defaultFn
}

// Validate invokes the field's data type Validate (which validates and may sanitize),
// then applies field rules. Returns the validated value and technical error if any.
// When value is empty: uses default if available; otherwise errors only when required with no fallback.
func (this *EntityField) Validate(value any, forEdit ...bool) (any, *ft.ClientErrorItem) {
	isForEdit := len(forEdit) > 0 && forEdit[0]

	if isNil(value) {
		if isForEdit {
			if this.isRequiredUpdate {
				return nil, ft.NewValidationError(this.name, "common.err_missing_required_field", "field is required")
			}
			return value, nil
		}

		if this.defaultValue != nil {
			return *this.defaultValue, nil
		}
		if this.defaultFn != nil {
			return this.defaultFn(), nil
		}
		if this.isRequiredCreate && !this.isAuto {
			return nil, ft.NewValidationError(this.name, "common.err_missing_required_field", "field is required")
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

func (this *EntityField) Clone() *EntityField {
	cloned := &EntityField{
		name:             this.name,
		label:            this.label,
		dataType:         this.dataType,
		description:      this.description,
		isArray:          this.isArray,
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
