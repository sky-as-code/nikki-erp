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
	FieldGraph           = "graph"
	FieldPage            = "page"
	FieldSize            = "size"
	FieldUpdatedAt       = "updated_at"
	FieldEtag            = "etag"
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

func ArchivableModelSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel("core.basemodel.archivable_model").
		Field(
			dmodel.DefineField().
				Name(FieldArchivedAt).
				DataType(dmodel.FieldDataTypeDateTime()).
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
				RequiredForCreate().
				RequiredForUpdate().
				ReadOnly(),
		)
}

func SetBaseModelSchemaBuilder(builder *dmodel.ModelSchemaBuilder) {
	baseBuilder = builder
}
