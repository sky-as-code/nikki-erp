package domain

import (
	"regexp"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	ModelMetadataSchemaName = "essential.model_metadata"

	ModelMetadataFieldId          = basemodel.FieldId
	ModelMetadataFieldTenantId    = "tenant_id"
	ModelMetadataFieldName        = "name"
	ModelMetadataFieldCode        = "code"
	ModelMetadataFieldCodePrefix  = "code_prefix"
	ModelMetadataFieldCodeLastSeq = "code_last_seq"
	ModelMetadataFieldPadding     = "padding"

	ModelMetadataEdgeFieldMetadataList = "field_metadata_list"
)

func ModelMetadataSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(ModelMetadataSchemaName).
		Label(model.LangJson{model.LanguageCodeEnUs: "Model Metadata"}).
		TableName("essential_model_metadata").
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Field(
			basemodel.DefineFieldId(ModelMetadataFieldTenantId).
				RequiredForCreate().
				TenantKey(),
		).
		Field(
			dmodel.DefineField().
				Name(ModelMetadataFieldName).
				DataType(dmodel.FieldDataTypeLangJson(1, model.MODEL_RULE_LONG_NAME_LENGTH)).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(ModelMetadataFieldCode).
				DataType(dmodel.FieldDataTypeString(1, model.MODEL_RULE_TINY_NAME_LENGTH, dmodel.FieldDataTypeStringOpts{
					Regex: regexp.MustCompile(`^[a-zA-Z0-9_]+$`),
				})).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(ModelMetadataFieldCodePrefix).
				DataType(dmodel.FieldDataTypeString(0, model.MODEL_RULE_TINY_NAME_LENGTH)).
				Default(""),
		).
		Field(
			dmodel.DefineField().
				Name(ModelMetadataFieldCodeLastSeq).
				DataType(dmodel.FieldDataTypeInt32(0, 1000000000)).
				Default(int32(0)).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(ModelMetadataFieldPadding).
				DataType(dmodel.FieldDataTypeInt32(1, 20)).
				Default(int32(8)).
				RequiredForCreate(),
		).
		CompositeUnique(ModelMetadataFieldTenantId, ModelMetadataFieldCode).
		Extend(basemodel.ArchivableModelSchemaBuilder()).
		Extend(basemodel.VersionedModelSchemaBuilder()).
		Extend(basemodel.AuditableModelSchemaBuilder()).
		EdgeFrom(
			dmodel.Edge(ModelMetadataEdgeFieldMetadataList).
				Label(model.LangJson{model.LanguageCodeEnUs: "Field metadata list"}).
				Existing(FieldMetadataSchemaName, FieldMetadataEdgeModelMetadata),
		)
}

type ModelMetadata struct {
	basemodel.DynamicModelBase
}

func NewModelMetadata() *ModelMetadata {
	return &ModelMetadata{basemodel.NewDynamicModel()}
}

func NewModelMetadataFrom(src dmodel.DynamicFields) *ModelMetadata {
	return &ModelMetadata{basemodel.NewDynamicModel(src)}
}
