package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	ModelMetadataSchemaName = "essential_model_metadata"

	ModelMetadataFieldId          = basemodel.FieldId
	ModelMetadataFieldName        = "name"
	ModelMetadataFieldCode        = "code"
	ModelMetadataFieldCodePrefix  = "code_prefix"
	ModelMetadataFieldCodeLastSeq = "code_last_seq"
	ModelMetadataFieldPadding     = "padding"

	ModelMetadataEdgeFieldMetadataList = "field_metadata_list"
)

//go:embed modelmetadata.json
var modelMetadataSchemaJson string

func ModelMetadataSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(modelMetadataSchemaJson)
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
