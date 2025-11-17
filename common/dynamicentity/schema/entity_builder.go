package schema

import (
	"fmt"
	"strings"

	"github.com/sky-as-code/nikki-erp/common/model"
)

type EntitySchemaBuilder struct {
	schema EntitySchema
}

func DefineEntity(name string) *EntitySchemaBuilder {
	return &EntitySchemaBuilder{
		schema: EntitySchema{
			name:   name,
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

func (b *FieldBuilder) Build() *EntityField {
	if b.field.name == "" {
		panic("field name is required")
	}
	return b.field
}
