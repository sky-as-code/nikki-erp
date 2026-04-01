package basemodel

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
)

const (
	FieldAssociations = "add"
	FieldDesociations = "remove"
	FieldId           = "id"
	FieldCreatedAt    = "created_at"
	FieldColumns      = "columns"
	FieldIsArchived   = "is_archived"
	FieldGraph        = "graph"
	FieldPage         = "page"
	FieldSize         = "size"
	FieldUpdatedAt    = "updated_at"
	FieldEtag         = "etag"
)

var baseBuilder *dmodel.ModelSchemaBuilder

func BaseModelSchemaBuilder() *dmodel.ModelSchemaBuilder {
	if baseBuilder == nil {
		baseBuilder = dmodel.DefineModel("core.basemodel.base_model").
			Field(
				dmodel.DefineField().
					Name(FieldId).
					Label(model.LangJson{"en-US": "ID"}).
					DataType(dmodel.FieldDataTypeUlid()).
					UseTypeDefault().
					PrimaryKey(),
			)
	}
	return baseBuilder
}

func ArchivableModelSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel("core.basemodel.archivable_model").
		Field(
			dmodel.DefineField().
				Name(FieldIsArchived).
				DataType(dmodel.FieldDataTypeBoolean()).
				RequiredForCreate().
				Default(false).
				ReadOnly(),
		)
}

func AuditableModelSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel("core.basemodel.auditable_model").
		Field(
			dmodel.DefineField().
				Name(FieldCreatedAt).
				DataType(dmodel.FieldDataTypeDateTime()).
				RequiredForCreate().
				UseTypeDefault().
				ReadOnly(),
		).
		Field(
			dmodel.DefineField().
				Name(FieldUpdatedAt).
				DataType(dmodel.FieldDataTypeDateTime()).
				ReadOnly(),
		)
}

func VersionedModelSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel("core.basemodel.versioned_model").
		Field(
			dmodel.DefineField().
				Name(FieldEtag).
				DataType(dmodel.FieldDataTypeEtag()).
				VersioningKey().
				UseTypeDefault(),
		)
}

func SetBaseModelSchemaBuilder(builder *dmodel.ModelSchemaBuilder) {
	baseBuilder = builder
}
