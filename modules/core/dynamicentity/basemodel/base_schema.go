package basemodel

import (
	"github.com/sky-as-code/nikki-erp/common/dynamicentity/schema"
	"github.com/sky-as-code/nikki-erp/common/model"
)

const (
	FieldId              = "id"
	FieldArchivedAt      = "archived_at"
	FieldCreatedAt       = "created_at"
	FieldColumns         = "columns"
	FieldIncludeArchived = "include_archived"
	FieldPage            = "page"
	FieldSize            = "size"
	FieldUpdatedAt       = "updated_at"
	FieldEtag            = "etag"
)

var baseBuilder *schema.EntitySchemaBuilder

func BaseModelSchemaBuilder() *schema.EntitySchemaBuilder {
	if baseBuilder == nil {
		baseBuilder = schema.DefineEntity("core.base_model").
			Field(
				schema.DefineField().
					Name(FieldId).
					Label(model.LangJson{"en-US": "ID"}).
					DataType(schema.FieldDataTypeModelId()).
					PrimaryKey().
					DefaultFn(func() any {
						id, err := model.NewId()
						if err != nil {
							panic(err)
						}
						return *id
					}),
			)
	}
	return baseBuilder
}

func ArchivableModelSchemaBuilder() *schema.EntitySchemaBuilder {
	return schema.DefineEntity("core.basemodel.archivable_model").
		Field(
			schema.DefineField().
				Name(FieldArchivedAt).
				DataType(schema.FieldDataTypeDateTime()).
				ReadOnly(),
		)
}

func GetOneQuerySchemaBuilder() *schema.EntitySchemaBuilder {
	return schema.DefineEntity("core.basemodel.get_one_query").
		Field(schema.DefineField().
			Name(FieldIncludeArchived).
			DataType(schema.FieldDataTypeBoolean()).
			Default(false)).
		Field(schema.DefineField().
			Name(FieldColumns).
			DataType(schema.FieldDataTypeString(model.MODEL_RULE_COL_LENGTH_MIN, model.MODEL_RULE_COL_LENGTH_MAX).ArrayType()))
}

func SearchQuerySchemaBuilder() *schema.EntitySchemaBuilder {
	return schema.DefineEntity("core.basemodel.get_one_query").
		Field(schema.DefineField().
			Name(FieldIncludeArchived).
			DataType(schema.FieldDataTypeBoolean()).
			Default(false)).
		Field(schema.DefineField().
			Name(FieldColumns).
			DataType(schema.FieldDataTypeString(model.MODEL_RULE_COL_LENGTH_MIN, model.MODEL_RULE_COL_LENGTH_MAX).ArrayType())).
		Field(schema.DefineField().
			Name(FieldPage).
			DataType(schema.FieldDataTypeInteger()).
			Default(model.MODEL_RULE_PAGE_INDEX_START)).
		Field(schema.DefineField().
			Name(FieldSize).
			DataType(schema.FieldDataTypeInteger()).
			Default(model.MODEL_RULE_PAGE_DEFAULT_SIZE))
}

func AuditableModelSchemaBuilder() *schema.EntitySchemaBuilder {
	return schema.DefineEntity("core.basemodel.auditable_model").
		Field(
			schema.DefineField().
				Name(FieldCreatedAt).
				DataType(schema.FieldDataTypeDateTime()).
				RequiredForCreate().
				ReadOnly(),
		).
		Field(
			schema.DefineField().
				Name(FieldUpdatedAt).
				DataType(schema.FieldDataTypeDateTime()).
				ReadOnly(),
		)
}

func VersionedModelSchemaBuilder() *schema.EntitySchemaBuilder {
	return schema.DefineEntity("core.basemodel.versioned_model").
		Field(
			schema.DefineField().
				Name(FieldEtag).
				DataType(schema.FieldDataTypeEtag()).
				RequiredForCreate().
				RequiredForUpdate().
				ReadOnly(),
		)
}

func SetBaseModelSchemaBuilder(builder *schema.EntitySchemaBuilder) {
	baseBuilder = builder
}
