package media

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

type InventoryMediaService interface {
	Upload(ctx corectx.Context, cmd UploadMediaCommand) (*UploadMediaResult, error)
	Delete(ctx corectx.Context, cmd DeleteMediaCommand) (*DeleteMediaResult, error)

	GetMedia(ctx corectx.Context, query GetMediaQuery) (*GetMediaResult, error)
	SearchMedia(ctx corectx.Context, query SearchMediaQuery) (*SearchMediaResult, error)
}
