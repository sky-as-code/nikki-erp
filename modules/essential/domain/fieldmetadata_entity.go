package domain

import (
	"regexp"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	FieldMetadataSchemaName = "essential.field_metadata"

	FieldMetadataFieldId              = basemodel.FieldId
	FieldMetadataFieldTenantId        = "tenant_id"
	FieldMetadataFieldModelMetadataId = "model_metadata_id"
	FieldMetadataFieldName            = "name"
	FieldMetadataFieldCode            = "code"
	FieldMetadataFieldDataType        = "data_type"
	FieldMetadataFieldIsRequired      = "is_required"
	FieldMetadataFieldDisplayOrder    = "display_order"

	FieldMetadataEdgeModelMetadata = "model_metadata"
)

func FieldMetadataSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(FieldMetadataSchemaName).
		Label(model.LangJson{model.LanguageCodeEnUs: "Field Metadata"}).
		TableName("essential_field_metadata").
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Field(
			basemodel.DefineFieldId(FieldMetadataFieldTenantId).
				RequiredForCreate().
				TenantKey(),
		).
		Field(
			basemodel.DefineFieldId(FieldMetadataFieldModelMetadataId).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(FieldMetadataFieldName).
				DataType(dmodel.FieldDataTypeLangJson(1, model.MODEL_RULE_LONG_NAME_LENGTH)).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(FieldMetadataFieldCode).
				DataType(dmodel.FieldDataTypeString(1, model.MODEL_RULE_TINY_NAME_LENGTH, dmodel.FieldDataTypeStringOpts{
					Regex: regexp.MustCompile(`^[a-zA-Z0-9_]+$`),
				})).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(FieldMetadataFieldDataType).
				DataType(dmodel.FieldDataTypeString(1, model.MODEL_RULE_TINY_NAME_LENGTH)).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(FieldMetadataFieldIsRequired).
				DataType(dmodel.FieldDataTypeBoolean()).
				Default(false).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(FieldMetadataFieldDisplayOrder).
				DataType(dmodel.FieldDataTypeInt32(0, 10000)).
				Default(int32(0)).
				RequiredForCreate(),
		).
		CompositeUnique(
			FieldMetadataFieldTenantId,
			FieldMetadataFieldModelMetadataId,
			FieldMetadataFieldCode,
		).
		Extend(basemodel.ArchivableModelSchemaBuilder()).
		Extend(basemodel.VersionedModelSchemaBuilder()).
		Extend(basemodel.AuditableModelSchemaBuilder()).
		EdgeTo(
			dmodel.Edge(FieldMetadataEdgeModelMetadata).
				ManyToOne(ModelMetadataSchemaName, dmodel.DynamicFields{
					FieldMetadataFieldModelMetadataId: ModelMetadataFieldId,
				}).
				OnDelete(dmodel.RelationCascadeCascade),
		)
}

type FieldMetadata struct {
	basemodel.DynamicModelBase
}

func NewFieldMetadata() *FieldMetadata {
	return &FieldMetadata{basemodel.NewDynamicModel()}
}

func NewFieldMetadataFrom(src dmodel.DynamicFields) *FieldMetadata {
	return &FieldMetadata{basemodel.NewDynamicModel(src)}
}
