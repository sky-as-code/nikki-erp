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
}

func (s EntitySchema) Name() string {
	return s.name
}

func (s EntitySchema) Label() model.LangJson {
	return s.label
}

func (s EntitySchema) Description() model.LangJson {
	return s.description
}

func (s EntitySchema) Fields() map[string]*EntityField {
	return s.fields
}

func (s EntitySchema) Rules() []EntityRule {
	return s.rules
}

func (s EntitySchema) Relations() []EntityRelation {
	return s.relations
}

func (s EntitySchema) Field(name string) (*EntityField, bool) {
	field, ok := s.fields[name]
	return field, ok
}

type EntityRelation struct {
	SourceField   string        `json:"source_field"`
	RelationType  RelationType  `json:"relation_type"`
	ThroughEntity *EntitySchema `json:"through_entity"`
	ForeignEntity EntitySchema  `json:"foreign_entity"`
	ForeignField  string        `json:"foreign_field"`
}

type RelationType string

const (
	RelationTypeBelongsToOne  = RelationType("belongsToOne")
	RelationTypeHasOne        = RelationType("hasOne")
	RelationTypeHasMany       = RelationType("hasMany")
	RelationTypeBelongsToMany = RelationType("belongsToMany")
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
	isRequired      bool
	rules           []*FieldRule
	defaultValue    any
}

// Getter methods
func (f *EntityField) Name() string {
	return f.name
}

func (f *EntityField) Label() model.LangJson {
	return f.label
}

func (f *EntityField) DataType() FieldDataType {
	return f.dataType
}

func (f *EntityField) DataTypeOptions() FieldDataTypeOptions {
	return f.dataTypeOptions
}

func (f *EntityField) Description() model.LangJson {
	return f.description
}

func (f *EntityField) IsRequired() bool {
	return f.isRequired
}

func (f *EntityField) Rules() []*FieldRule {
	return f.rules
}

func (f *EntityField) Default() any {
	return f.defaultValue
}

// Setter methods
// func (f *EntityField) SetName(name string) {
// 	f.name = name
// }

// func (f *EntityField) SetLabel(label model.LangJson) {
// 	f.label = label
// }

// func (f *EntityField) SetDataType(dataType FieldDataType) {
// 	f.dataType = dataType
// }

// func (f *EntityField) SetDataTypeOptions(options FieldDataTypeOptions) {
// 	f.dataTypeOptions = options
// }

// func (f *EntityField) SetDescription(description model.LangJson) {
// 	f.description = description
// }

// func (f *EntityField) SetIsRequired(isRequired bool) {
// 	f.isRequired = isRequired
// }

// func (f *EntityField) SetRules(rules []FieldRule) {
// 	f.rules = rules
// }

// func (f *EntityField) SetDefault(value any) {
// 	f.defaultValue = value
// }

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
	FieldDataTypeOptPrecision = FieldDataTypeOptName("precision")
)

type FieldRule []any

func (r FieldRule) RuleName() FieldRuleName {
	if len(r) == 0 {
		return ""
	}
	if name, ok := r[0].(FieldRuleName); ok {
		return name
	}
	return FieldRuleName(fmt.Sprint(r[0]))
}

func (r FieldRule) RuleOptions() any {
	if len(r) < 2 {
		return nil
	}
	return r[1]
}

type FieldRuleName string

const (
	FieldRuleMaxType     = FieldRuleName("max")
	FieldRuleMinType     = FieldRuleName("min")
	FieldRuleLengthType  = FieldRuleName("length")
	FieldRuleOneOfType   = FieldRuleName("oneOf")
	FieldRulePrimaryType = FieldRuleName("primary")
	FieldRuleTenantType  = FieldRuleName("tenant")
	FieldRuleUniqueType  = FieldRuleName("unique")
)

func FieldRuleMax(value any) FieldRule {
	return FieldRule{FieldRuleMaxType, value}
}

func FieldRuleMin(value any) FieldRule {
	return FieldRule{FieldRuleMinType, value}
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

func (f *EntityField) Clone() *EntityField {
	cloned := &EntityField{
		name:            f.name,
		label:           f.label,
		dataType:        f.dataType,
		dataTypeOptions: make(FieldDataTypeOptions),
		description:     f.description,
		isRequired:      f.isRequired,
		rules:           make([]*FieldRule, len(f.rules)),
		defaultValue:    f.defaultValue,
	}

	// Deep copy DataTypeOptions
	if f.dataTypeOptions != nil {
		for k, v := range f.dataTypeOptions {
			cloned.dataTypeOptions[k] = v
		}
	}

	// Deep copy Rules
	copy(cloned.rules, f.rules)

	return cloned
}
