package modelmetadata

import corectx "github.com/sky-as-code/nikki-erp/modules/core/context"

type ModelMetadataService interface {
	CreateModelMetadata(ctx corectx.Context, cmd CreateModelMetadataCommand) (*CreateModelMetadataResult, error)
	DeleteModelMetadata(ctx corectx.Context, cmd DeleteModelMetadataCommand) (*DeleteModelMetadataResult, error)
	ModelMetadataExists(ctx corectx.Context, query ModelMetadataExistsQuery) (*ModelMetadataExistsResult, error)
	GetModelMetadata(ctx corectx.Context, query GetModelMetadataQuery) (*GetModelMetadataResult, error)
	SearchModelMetadata(ctx corectx.Context, query SearchModelMetadataQuery) (*SearchModelMetadataResult, error)
	UpdateModelMetadata(ctx corectx.Context, cmd UpdateModelMetadataCommand) (*UpdateModelMetadataResult, error)
}
