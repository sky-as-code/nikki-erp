package schema

import (
	"go.bryk.io/pkg/errors"
)

func AddAdhocSchema(name string, schema *AdhocSchema) error {
	return adhocRegistry.Add(name, schema)
}

type AdhocSchema struct {
	fields map[string]*AdhocField
}

func (this *AdhocSchema) Fields() map[string]*AdhocField {
	return this.fields
}

func (this *AdhocSchema) Field(name string) (*AdhocField, bool) {
	field, ok := this.fields[name]
	return field, ok
}

type AdhocField struct {
	field        *EntityField
	isHolder     bool
	holderSchema *AdhocSchema
}

func (f *AdhocField) Name() string {
	return f.field.name
}

func (f *AdhocField) Field() *EntityField {
	return f.field
}

func (f *AdhocField) IsRequired() bool {
	return f.field.isRequired
}

func (f *AdhocField) IsHolder() bool {
	return f.isHolder
}

func (f *AdhocField) HolderSchema() *AdhocSchema {
	return f.holderSchema
}

type AdhocSchemaBuilder struct {
	schema *AdhocSchema
}

func DefineAdhoc() *AdhocSchemaBuilder {
	return &AdhocSchemaBuilder{
		schema: &AdhocSchema{
			fields: make(map[string]*AdhocField),
		},
	}
}

func (this *AdhocSchemaBuilder) Field(fieldBuilder *FieldBuilder) *AdhocSchemaBuilder {
	if fieldBuilder == nil {
		panic(errors.New("field builder cannot be nil"))
	}

	field := fieldBuilder.Build()
	if err := validateFieldName(field); err != nil {
		panic(err)
	}

	this.schema.fields[field.name] = &AdhocField{
		field: field,
	}
	return this
}

func (this *AdhocSchemaBuilder) FieldHolder(name string, isRequired bool, holderSchemaBuilder *AdhocSchemaBuilder) *AdhocSchemaBuilder {
	if holderSchemaBuilder == nil {
		panic("holder schema builder cannot be nil")
	}
	holderSchema := holderSchemaBuilder.Build()

	this.schema.fields[name] = &AdhocField{
		isHolder:     true,
		field:        DefineField().Name(name).IsRequired(isRequired).Build(),
		holderSchema: holderSchema,
	}

	return this
}

func (this *AdhocSchemaBuilder) Build() *AdhocSchema {
	return this.schema
}
