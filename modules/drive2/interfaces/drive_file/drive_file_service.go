package drive_file

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

type DriveFileAppService interface {
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
	DownloadDriveFile(ctx corectx.Context, query DownloadDriveFileQuery) (*DownloadDriveFileResult, error)
	GetDriveFileByParent(ctx corectx.Context, query GetDriveFileByParentQuery) (*GetDriveFileByParentResult, error)
	SearchDriveFile(ctx corectx.Context, query SearchDriveFileQuery) (*SearchDriveFileResult, error)
	SearchDriveFilesShared(ctx corectx.Context, query SearchDriveFilesSharedQuery) (
		*SearchDriveFilesSharedResult, error)
	GetDriveFileAncestors(ctx corectx.Context, query GetDriveFileAncestorsQuery) (
		*GetDriveFileAncestorsResult, error)
}

type DriveFileDomainService interface {
	CreateDriveFile(ctx corectx.Context, cmd CreateDriveFileCommand) (*CreateDriveFileResult, error)
	UpdateDriveFileMetadata(ctx corectx.Context, cmd UpdateDriveFileMetadataCommand) (*UpdateDriveFileResult, error)
	UpdateBulkDriveFileMetadata(ctx corectx.Context, cmd UpdateBulkDriveFileMetadataCommand) (
		*UpdateBulkDriveFileMetadataResult, error)
	DeleteDriveFile(ctx corectx.Context, cmd DeleteDriveFileCommand) (*DeleteDriveFileResult, error)
	DeleteDriveFiles(ctx corectx.Context, cmd DeleteDriveFilesCommand) (*DeleteDriveFilesResult, error)
	MoveDriveFileToTrash(ctx corectx.Context, cmd MoveDriveFileToTrashCommand) (
		*MoveDriveFileToTrashResult, error)
	RestoreDriveFile(ctx corectx.Context, cmd RestoreDriveFileCommand) (*RestoreDriveFileResult, error)
	MoveDriveFile(ctx corectx.Context, cmd MoveDriveFileCommand) (*MoveDriveFileResult, error)
	GetDriveFileById(ctx corectx.Context, query GetDriveFileByIdQuery) (*GetDriveFileByIdResult, error)
	GetDriveFileByParent(ctx corectx.Context, query GetDriveFileByParentQuery) (*GetDriveFileByParentResult, error)
	GetDriveFileChildren(ctx corectx.Context, query GetDriveFileChildrenQuery) (*GetDriveFileChildrenResult, error)
	GetDriveFileAncestors(ctx corectx.Context, query GetDriveFileAncestorsQuery) (*GetDriveFileAncestorsResult, error)
	SearchDriveFile(ctx corectx.Context, query SearchDriveFileQuery) (*SearchDriveFileResult, error)
}
