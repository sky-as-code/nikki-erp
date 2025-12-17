package schema

import (
	"fmt"

	"github.com/sky-as-code/nikki-erp/common/model"
)

const SchemaStructTag = "json"

type EntitySchema struct {
	name        string                  `json:"name"`
	label       model.LangJson          `json:"label"`
	description model.LangJson          `json:"description"`
	fields      map[string]*EntityField `json:"fields"`
	rules       []EntityRule            `json:"rules"`
	relations   []EntityRelation        `json:"relations"`
	tableName   string                  `json:"table_name"`
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

func (this EntitySchema) Rules() []EntityRule {
	return this.rules
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

type EntityRule []any
type EntityRuleName string

const (
	EntityRuleNameUnique = EntityRuleName("unique")
)

type EntityField struct {
	name            string
	label           model.LangJson
	dataType        FieldDataType
	dataTypeOptions FieldDataTypeOptions
	description     model.LangJson
	isArray         bool
	isRequired      bool
	rules           []*FieldRule
	defaultValue    any
	relation        *EntityRelation
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

func (this *EntityField) DataTypeOptions() FieldDataTypeOptions {
	return this.dataTypeOptions
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

func (this *EntityField) Rules() []*FieldRule {
	return this.rules
}

func (this *EntityField) Default() any {
	return this.defaultValue
}

type FieldDataType string

const (
	FieldDataTypeEmail  = FieldDataType("email")
	FieldDataTypePhone  = FieldDataType("phone")
	FieldDataTypeString = FieldDataType("string")
	FieldDataTypeSecret = FieldDataType("secret")
	FieldDataTypeUrl    = FieldDataType("url")
	FieldDataTypeUlid   = FieldDataType("ulid")
	FieldDataTypeUuid   = FieldDataType("uuid")

	FieldDataTypeInteger = FieldDataType("integer")
	FieldDataTypeFloat   = FieldDataType("float")
	FieldDataTypeBoolean = FieldDataType("boolean")

	FieldDataTypeDate     = FieldDataType("date")
	FieldDataTypeTime     = FieldDataType("time")
	FieldDataTypeDateTime = FieldDataType("dateTime")

	FieldDataTypeEnumString = FieldDataType("enumString")
	FieldDataTypeEnumNumber = FieldDataType("enumNumber")

	FieldDataTypeEtag     = FieldDataType("nikkiEtag")
	FieldDataTypeLangJson = FieldDataType("nikkiLangJson")
	FieldDataTypeLangCode = FieldDataType("nikkiLangCode")
	FieldDataTypeModelId  = FieldDataType("nikkiModelId")
	FieldDataTypeSlug     = FieldDataType("nikkiSlug")
)

// These types has their own validation rules, so the dynamic validator
// will skip processing some FieldRule for them (i.e: FieldRuleLengthType).
var CustomDataTypes = []FieldDataType{
	FieldDataTypeEtag,
	FieldDataTypeLangJson,
	FieldDataTypeLangCode,
	FieldDataTypeModelId,
	FieldDataTypeSlug,
}

type FieldDataTypeOptName string
type FieldDataTypeOptions map[FieldDataTypeOptName]any

const (
	FieldDataTypeOptEnumValues = FieldDataTypeOptName("enumValues")
	FieldDataTypeOptPrecision  = FieldDataTypeOptName("precision")
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
	FieldRulePrimaryType     = FieldRuleName("primary")
	FieldRuleTenantType      = FieldRuleName("tenant")
	FieldRuleUniqueType      = FieldRuleName("unique")
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

func FieldRulePrimary() FieldRule {
	return FieldRule{FieldRulePrimaryType}
}

func FieldRuleTenant() FieldRule {
	return FieldRule{FieldRuleTenantType}
}

func FieldRuleUnique() FieldRule {
	return FieldRule{FieldRuleUniqueType}
}

func (this *EntityField) Clone() *EntityField {
	cloned := &EntityField{
		name:            this.name,
		label:           this.label,
		dataType:        this.dataType,
		dataTypeOptions: make(FieldDataTypeOptions),
		description:     this.description,
		isArray:         this.isArray,
		isRequired:      this.isRequired,
		rules:           make([]*FieldRule, len(this.rules)),
		defaultValue:    this.defaultValue,
		relation:        this.relation,
	}

	// Deep copy DataTypeOptions
	if this.dataTypeOptions != nil {
		for k, v := range this.dataTypeOptions {
			cloned.dataTypeOptions[k] = v
		}
	}

	// Deep copy Rules
	copy(cloned.rules, this.rules)

	return cloned
}
