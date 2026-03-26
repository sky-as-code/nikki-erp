package drive_file_signed_url

import (
	"context"

	"github.com/sky-as-code/nikki-erp/common/model"
)

type DriveFileSignedUrlService interface {
	Create(ctx context.Context, fileId model.Id) (string, error)
	Get(ctx context.Context, fileId model.Id) (string, error)
	GetAndDelete(ctx context.Context, fileId model.Id) (string, error)
	GetOrCreate(ctx context.Context, fileId model.Id) (string, error)
	Verify(ctx context.Context, fileId model.Id, token string) (bool, error)
}
