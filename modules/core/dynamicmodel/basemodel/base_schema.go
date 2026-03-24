package basemodel

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
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

var baseBuilder *dmodel.EntitySchemaBuilder

func BaseModelSchemaBuilder() *dmodel.EntitySchemaBuilder {
	if baseBuilder == nil {
		baseBuilder = dmodel.DefineEntity("core.base_model").
			Field(
				dmodel.DefineField().
					Name(FieldId).
					Label(model.LangJson{"en-US": "ID"}).
					DataType(dmodel.FieldDataTypeModelId()).
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

func ArchivableModelSchemaBuilder() *dmodel.EntitySchemaBuilder {
	return dmodel.DefineEntity("core.basemodel.archivable_model").
		Field(
			dmodel.DefineField().
				Name(FieldArchivedAt).
				DataType(dmodel.FieldDataTypeDateTime()).
				ReadOnly(),
		)
}

func GetOneQuerySchemaBuilder() *dmodel.EntitySchemaBuilder {
	return dmodel.DefineEntity("core.basemodel.get_one_query").
		Field(dmodel.DefineField().
			Name(FieldIncludeArchived).
			DataType(dmodel.FieldDataTypeBoolean()).
			Default(false)).
		Field(dmodel.DefineField().
			Name(FieldColumns).
			DataType(dmodel.FieldDataTypeString(model.MODEL_RULE_COL_LENGTH_MIN, model.MODEL_RULE_COL_LENGTH_MAX).ArrayType()))
}

func SearchQuerySchemaBuilder() *dmodel.EntitySchemaBuilder {
	return dmodel.DefineEntity("core.basemodel.get_one_query").
		Field(dmodel.DefineField().
			Name(FieldIncludeArchived).
			DataType(dmodel.FieldDataTypeBoolean()).
			Default(false)).
		Field(dmodel.DefineField().
			Name(FieldColumns).
			DataType(dmodel.FieldDataTypeString(model.MODEL_RULE_COL_LENGTH_MIN, model.MODEL_RULE_COL_LENGTH_MAX).ArrayType())).
		Field(dmodel.DefineField().
			Name(FieldPage).
			DataType(dmodel.FieldDataTypeInteger()).
			Default(model.MODEL_RULE_PAGE_INDEX_START)).
		Field(dmodel.DefineField().
			Name(FieldSize).
			DataType(dmodel.FieldDataTypeInteger()).
			Default(model.MODEL_RULE_PAGE_DEFAULT_SIZE))
}

func AuditableModelSchemaBuilder() *dmodel.EntitySchemaBuilder {
	return dmodel.DefineEntity("core.basemodel.auditable_model").
		Field(
			dmodel.DefineField().
				Name(FieldCreatedAt).
				DataType(dmodel.FieldDataTypeDateTime()).
				RequiredForCreate().
				ReadOnly(),
		).
		Field(
			dmodel.DefineField().
				Name(FieldUpdatedAt).
				DataType(dmodel.FieldDataTypeDateTime()).
				ReadOnly(),
		)
}

func VersionedModelSchemaBuilder() *dmodel.EntitySchemaBuilder {
	return dmodel.DefineEntity("core.basemodel.versioned_model").
		Field(
			dmodel.DefineField().
				Name(FieldEtag).
				DataType(dmodel.FieldDataTypeEtag()).
				RequiredForCreate().
				RequiredForUpdate().
				ReadOnly(),
		)
}

func SetBaseModelSchemaBuilder(builder *dmodel.EntitySchemaBuilder) {
	baseBuilder = builder
}
