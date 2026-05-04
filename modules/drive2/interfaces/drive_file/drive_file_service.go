package drive_file

import (
	"io"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/drive2/domain"
)

type DriveFileService interface {
	CreateDriveFile(ctx corectx.Context, cmd CreateDriveFileCommand) (*CreateDriveFileResult, error)
	UpdateDriveFileMetadata(ctx corectx.Context, cmd UpdateDriveFileMetadataCommand) (
		*UpdateDriveFileResult, error)
	UpdateBulkDriveFileMetadata(ctx corectx.Context, cmd UpdateBulkDriveFileMetadataCommand) (
		*UpdateBulkDriveFileMetadataResult, error)
	UpdateDriveFileContent(ctx corectx.Context, cmd UpdateDriveFileContentCommand) (
		*UpdateBulkDriveFileMetadataResult, error)
	DeleteDriveFile(ctx corectx.Context, cmd DeleteDriveFileCommand) (*DeleteDriveFileResult, error)
	MoveDriveFileToTrash(ctx corectx.Context, cmd MoveDriveFileToTrashCommand) (
		*MoveDriveFileToTrashResult, error)
	RestoreDriveFile(ctx corectx.Context, cmd RestoreDriveFileCommand) (*RestoreDriveFileResult, error)
	MoveDriveFile(ctx corectx.Context, cmd MoveDriveFileCommand) (*MoveDriveFileResult, error)
	DeleteTrashedDriveFile(ctx corectx.Context) error

	GetDriveFileById(ctx corectx.Context, query GetDriveFileByIdQuery) (*GetDriveFileByIdResult, error)
	DownloadDriveFile(ctx corectx.Context, query GetDriveFileByIdQuery) (*domain.DriveFile, io.ReadCloser, error)
	GetDriveFileByParent(ctx corectx.Context, query GetDriveFileByParentQuery) (*GetDriveFileByParentResult, error)
	SearchDriveFile(ctx corectx.Context, query SearchDriveFileQuery) (*SearchDriveFileResult, error)
	SearchDriveFilesShared(ctx corectx.Context, query SearchDriveFilesSharedQuery) (
		*SearchDriveFilesSharedResult, error)
	GetDriveFileAncestors(ctx corectx.Context, query GetDriveFileAncestorsQuery) (
		*GetDriveFileAncestorsResult, error)
}
