package schema

import (
	"fmt"
	"strings"

	"github.com/sky-as-code/nikki-erp/common/model"
)

type EntitySchemaBuilder struct {
	schema EntitySchema
}

func DefineEntity() *EntitySchemaBuilder {
	return &EntitySchemaBuilder{
		schema: EntitySchema{
			fields: make(map[string]*EntityField),
		},
	}
}

func (b *EntitySchemaBuilder) Label(label model.LangJson) *EntitySchemaBuilder {
	b.schema.label = label
	return b
}

func (b *EntitySchemaBuilder) LabelRef(key string) *EntitySchemaBuilder {
	return b.Label(model.LangJson{"$s": key})
}

func (b *EntitySchemaBuilder) Description(description model.LangJson) *EntitySchemaBuilder {
	b.schema.description = description
	return b
}

func (b *EntitySchemaBuilder) Name(name string) *EntitySchemaBuilder {
	b.schema.name = name
	return b
}

func (b *EntitySchemaBuilder) Field(fieldBuilder *FieldBuilder) *EntitySchemaBuilder {
	if fieldBuilder == nil {
		return b
	}

	field := fieldBuilder.Build()
	if err := validateFieldName(field); err != nil {
		panic(err)
	}
	if b.schema.fields == nil {
		b.schema.fields = make(map[string]*EntityField)
	}
	b.schema.fields[field.name] = field

	if field.relation != nil {
		rel := field.relation
		rel.SrcField = field.name
		b.schema.relations = append(b.schema.relations, *rel)
		field.relation = nil
	}

	return b
}

func validateFieldName(field *EntityField) error {
	if strings.TrimSpace(field.name) == "" {
		return fmt.Errorf("field name is required")
	}
	return nil
}

func (b *EntitySchemaBuilder) Rule(name any, args ...any) *EntitySchemaBuilder {
	rule := EntityRule{name}
	rule = append(rule, args...)
	b.schema.rules = append(b.schema.rules, rule)
	return b
}

func (b *EntitySchemaBuilder) TableName(tableName string) *EntitySchemaBuilder {
	b.schema.tableName = tableName
	return b
}

func (b *EntitySchemaBuilder) Build() *EntitySchema {
	return &b.schema
}

type FieldBuilder struct {
	field *EntityField
}

func DefineField() *FieldBuilder {
	return &FieldBuilder{
		field: &EntityField{},
	}
}

func (b *FieldBuilder) Name(name string) *FieldBuilder {
	b.field.name = name
	return b
}

func (b *FieldBuilder) Label(label model.LangJson) *FieldBuilder {
	b.field.label = label
	return b
}

func (b *FieldBuilder) IsRequired(isRequired bool) *FieldBuilder {
	b.field.isRequired = isRequired
	return b
}

func (b *FieldBuilder) Required() *FieldBuilder {
	b.field.isRequired = true
	return b
}

func (b *FieldBuilder) LabelRef(key string) *FieldBuilder {
	return b.Label(model.LangJson{model.LabelRefLanguageCode: key})
}

func (b *FieldBuilder) Description(description model.LangJson) *FieldBuilder {
	b.field.description = description
	return b
}

func (b *FieldBuilder) DataType(dataType FieldDataType, options ...FieldDataTypeOptions) *FieldBuilder {
	b.field.dataType = dataType

	if len(options) == 0 {
		return b
	}

	opts := b.field.dataTypeOptions
	if opts == nil {
		opts = FieldDataTypeOptions{}
		b.field.dataTypeOptions = opts
	}

	for _, option := range options {
		for key, value := range option {
			opts[key] = value
		}
	}

	return b
}

func (b *FieldBuilder) Rule(rule FieldRule) *FieldBuilder {
	rules := b.field.rules
	rules = append(rules, &rule)
	b.field.rules = rules
	return b
}

func (b *FieldBuilder) Default(value any) *FieldBuilder {
	b.field.defaultValue = value
	return b
}

func (b *FieldBuilder) Foreign(relationBuilder *RelationBuilder) *FieldBuilder {
	if relationBuilder == nil {
		return b
	}
	b.field.relation = relationBuilder.Build()
	return b
}

func (b *FieldBuilder) Build() *EntityField {
	if b.field.name == "" {
		panic("field name is required")
	}
	return b.field
}

type RelationBuilder struct {
	relation *EntityRelation
}

func Edge(edgeName string) *RelationBuilder {
	return &RelationBuilder{
		relation: &EntityRelation{
			Edge: edgeName,
		},
	}
}

func (b *RelationBuilder) OneToOne(entityName string, destField string) *RelationBuilder {
	b.relation.RelationType = RelationTypeOneToOne
	b.relation.DestEntityName = entityName
	b.relation.DestField = destField
	return b
}

func (b *RelationBuilder) ManyToOne(entityName string, targetField string) *RelationBuilder {
	b.relation.RelationType = RelationTypeManyToOne
	b.relation.DestEntityName = entityName
	b.relation.DestField = targetField
	return b
}

func (b *RelationBuilder) ManyToMany(throughTableName string, throughSrcCol string, throughDestCol string) *RelationBuilder {
	b.relation.RelationType = RelationTypeManyToMany
	b.relation.ThroughTableName = throughTableName
	b.relation.ThroughSrcCol = throughSrcCol
	b.relation.ThroughDestCol = throughDestCol
	return b
}

func (b *RelationBuilder) Build() *EntityRelation {
	return b.relation
}
