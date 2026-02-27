package drive_file

import (
	"io"

	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

type DriveFileService interface {
	CreateDriveFile(ctx crud.Context, cmd CreateDriveFileCommand) (*CreateDriveFileResult, error)
	UpdateDriveFile(ctx crud.Context, cmd UpdateDriveFileCommand) (*UpdateDriveFileResult, error)
	DeleteDriveFile(ctx crud.Context, cmd DeleteDriveFileCommand) (*DeleteDriveFileResult, error)
	MoveDriveFileToTrash(ctx crud.Context, cmd MoveDriveFileToTrashCommand) (*MoveDriveFileToTrashResult, error)

	GetDriveFileById(ctx crud.Context, query GetDriveFileByIdQuery) (*GetDriveFileByIdResult, error)
	DownloadDriveFile(ctx crud.Context, query GetDriveFileByIdQuery) (io.ReadCloser, error)
	GetDriveFileByParent(ctx crud.Context, query GetDriveFileByParentQuery) (*GetDriveFileByParentResult, error)
	SearchDriveFile(ctx crud.Context, query SearchDriveFileQuery) (*SearchDriveFileResult, error)
}
