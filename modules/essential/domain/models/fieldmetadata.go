package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	FieldMetadataSchemaName = "essential_field_metadata"

	FieldMetadataFieldId              = basemodel.FieldId
	FieldMetadataFieldModelMetadataId = "model_metadata_id"
	FieldMetadataFieldName            = "name"
	FieldMetadataFieldCode            = "code"
	FieldMetadataFieldDataType        = "data_type"
	FieldMetadataFieldIsRequired      = "is_required"
	FieldMetadataFieldDisplayOrder    = "display_order"

	FieldMetadataEdgeModelMetadata = "model_metadata"
)

//go:embed fieldmetadata.json
var fieldMetadataSchemaJson string

func FieldMetadataSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(fieldMetadataSchemaJson)
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
