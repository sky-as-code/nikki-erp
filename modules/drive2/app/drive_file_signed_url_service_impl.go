package app

import (
	"context"

	"github.com/sky-as-code/nikki-erp/common/model"
	tokenIt "github.com/sky-as-code/nikki-erp/modules/drive2/interfaces/drive_file_token"
)

type DriveFileSignedUrlServiceImpl struct{}

func NewDriveFileSignedUrlService() tokenIt.DriveFileSignedUrlService {
	return &DriveFileSignedUrlServiceImpl{}
}

func (this *DriveFileSignedUrlServiceImpl) Create(ctx context.Context, fileId model.Id) (string, error) {
	panic("drive_file_signed_url_service: Create unimplemented")
}

func (this *DriveFileSignedUrlServiceImpl) Get(ctx context.Context, fileId model.Id) (string, error) {
	panic("drive_file_signed_url_service: Get unimplemented")
}

func (this *DriveFileSignedUrlServiceImpl) GetAndDelete(ctx context.Context, fileId model.Id) (string, error) {
	panic("drive_file_signed_url_service: GetAndDelete unimplemented")
}

func (this *DriveFileSignedUrlServiceImpl) GetOrCreate(ctx context.Context, fileId model.Id) (string, error) {
	panic("drive_file_signed_url_service: GetOrCreate unimplemented")
}

func (this *DriveFileSignedUrlServiceImpl) Verify(ctx context.Context, fileId model.Id, token string) (bool, error) {
	panic("drive_file_signed_url_service: Verify unimplemented")
}
