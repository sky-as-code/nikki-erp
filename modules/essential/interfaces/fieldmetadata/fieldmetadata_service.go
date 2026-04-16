package fieldmetadata

import corectx "github.com/sky-as-code/nikki-erp/modules/core/context"

type FieldMetadataService interface {
	CreateFieldMetadata(ctx corectx.Context, cmd CreateFieldMetadataCommand) (*CreateFieldMetadataResult, error)
	DeleteFieldMetadata(ctx corectx.Context, cmd DeleteFieldMetadataCommand) (*DeleteFieldMetadataResult, error)
	FieldMetadataExists(ctx corectx.Context, query FieldMetadataExistsQuery) (*FieldMetadataExistsResult, error)
	GetFieldMetadata(ctx corectx.Context, query GetFieldMetadataQuery) (*GetFieldMetadataResult, error)
	SearchFieldMetadata(ctx corectx.Context, query SearchFieldMetadataQuery) (*SearchFieldMetadataResult, error)
	UpdateFieldMetadata(ctx corectx.Context, cmd UpdateFieldMetadataCommand) (*UpdateFieldMetadataResult, error)
}
