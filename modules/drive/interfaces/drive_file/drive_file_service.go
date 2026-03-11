package drive_file

import (
	"io"

	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/drive/domain"
)

type DriveFileService interface {
	CreateDriveFile(ctx crud.Context, cmd CreateDriveFileCommand) (*CreateDriveFileResult, error)
	UpdateDriveFileMetadata(ctx crud.Context, cmd UpdateDriveFileMetadataCommand) (*UpdateDriveFileMetadataResult, error)
	UpdateBulkDriveFileMetadata(ctx crud.Context, cmd UpdateBulkDriveFileMetadataCommand) (*UpdateBulkDriveFileMetadataResult, error)
	UpdateDriveFileContent(ctx crud.Context, cmd UpdateDriveFileContentCommand) (*UpdateDriveFileContentResult, error)
	DeleteDriveFile(ctx crud.Context, cmd DeleteDriveFileCommand) (*DeleteDriveFileResult, error)
	MoveDriveFileToTrash(ctx crud.Context, cmd MoveDriveFileToTrashCommand) (*MoveDriveFileToTrashResult, error)
	RestoreDriveFile(ctx crud.Context, cmd RestoreDriveFileCommand) (*RestoreDriveFileResult, error)
	MoveDriveFile(ctx crud.Context, cmd MoveDriveFileCommand) (*MoveDriveFileResult, error)
	DeleteTrashedDriveFile(ctx crud.Context) error

	GetDriveFileById(ctx crud.Context, query GetDriveFileByIdQuery) (*GetDriveFileByIdResult, error)
	DownloadDriveFile(ctx crud.Context, query GetDriveFileByIdQuery) (*domain.DriveFile, io.ReadCloser, error)
	GetDriveFileByParent(ctx crud.Context, query GetDriveFileByParentQuery) (*GetDriveFileByParentResult, error)
	SearchDriveFile(ctx crud.Context, query SearchDriveFileQuery) (*SearchDriveFileResult, error)
	GetDriveFileAncestors(ctx crud.Context, query GetDriveFileAncestorsQuery) (*GetDriveFileAncestorsResult, error)
}
