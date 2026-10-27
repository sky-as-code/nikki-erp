// Package file declares the layers of the itstorage_file resource: one stored object and its
// metadata.
//
// Over REST the built-in surface is read and delete. A file is created only by the streamed
// upload, which writes the object and the row together.
package file

import (
	"io"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/storage"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/itstorage/domain/models"
)

// FileRepository adds the set-based statements the org cleanup needs. Both join the transaction
// the context carries and confine themselves to the acting tenant.
type FileRepository interface {
	composable.CrudRepository

	ListObjectKeysInSpaces(ctx corectx.Context, spaceIds []model.Id) ([]string, error)
	DeleteInSpaces(ctx corectx.Context, spaceIds []model.Id) (int, error)
	// SumSizesByUploader totals the files of the spaces per uploader, skipping files without one.
	SumSizesByUploader(ctx corectx.Context, spaceIds []model.Id) (map[model.Id]int64, error)

	// SetStatus moves the files among ids whose status is fromStatus to toStatus and returns the
	// object keys of those it moved, by id. Moving to in_use stamps in_use_at, and skips a file
	// that has it already: a released file never comes back into use.
	SetStatus(ctx corectx.Context, ids []model.Id, fromStatus string, toStatus string) (map[model.Id]string, error)
	// ListObjectKeysWithStatus returns, by id, the object keys of the files among ids that have
	// the status, as the context's transaction (or, without one, the committed data) sees them.
	ListObjectKeysWithStatus(ctx corectx.Context, ids []model.Id, status string) (map[model.Id]string, error)
}

type FileDomainService interface {
	composable.CrudDomainService

	// Upload validates the declared metadata and the content, checks the tenant quota, streams
	// the body to storage and records the file. It authorizes nothing.
	Upload(ctx corectx.Context, cmd UploadFileCommand) (*UploadFileResult, error)

	// FindByPath returns nil when the path names no file of the acting tenant, or names one by a
	// path that is no longer its current one.
	FindByPath(ctx corectx.Context, filePath string) (*ResolvedFile, error)
	// FindFile returns nil when no file of the acting tenant has the id.
	FindFile(ctx corectx.Context, fileId model.Id) (*models.File, error)

	// The file lifecycle behind the core/storage.StorageConsumer port; see its documentation.
	AssertFilesUsable(ctx corectx.Context, query storage.AssertFilesUsableQuery) (ft.ClientErrors, error)
	MarkFilesInUse(ctx corectx.Context, cmd storage.MarkFilesInUseCommand) error
	ReleaseFiles(ctx corectx.Context, cmd storage.ReleaseFilesCommand) error
	DuplicateFile(ctx corectx.Context, cmd storage.DuplicateFileCommand) (*model.Id, error)
	GetFileMeta(ctx corectx.Context, query storage.GetFileMetaQuery) (*storage.FileMeta, ft.ClientErrors, error)

	// OpenObject streams the stored object, honouring an HTTP Range header when given.
	OpenObject(ctx corectx.Context, file *models.File, rangeHeader string) (*DownloadFileResultData, error)

	// DeleteSpaceTree deletes the space, everything below it and their files' rows in the
	// transaction ctx carries, takes its size off its ancestors, and returns the object keys to
	// remove once the transaction commits.
	DeleteSpaceTree(ctx corectx.Context, space *models.Space) ([]string, error)

	// RemoveObjects removes stored objects best-effort, logging each failure as an orphan.
	RemoveObjects(ctx corectx.Context, objectKeys []string)
}

type FileApplicationService interface {
	composable.CrudApplicationService

	UploadFile(ctx corectx.Context, cmd UploadFileCommand) (*UploadFileResult, error)
	GetFileMeta(ctx corectx.Context, query GetFileMetaQuery) (*GetFileMetaResult, error)
	DownloadFile(ctx corectx.Context, query DownloadFileQuery) (*DownloadFileResult, error)
	DownloadFileById(ctx corectx.Context, query DownloadFileByIdQuery) (*DownloadFileResult, error)
	GetFileMetaById(ctx corectx.Context, query GetFileMetaByIdQuery) (*GetFileMetaResult, error)
}

// ResolvedFile is a file together with the space it currently sits in.
type ResolvedFile struct {
	File  *models.File
	Space *models.Space
	// FilePath is the path under the space's current name, which may differ from the one the
	// caller asked with when the space was renamed since.
	FilePath string
}

// UploadFileCommand carries the text parts of the multipart request and the file part's stream.
// Body is read once and is not closed by the service.
type UploadFileCommand struct {
	OrgId     *model.Id `json:"org_id"`
	SpaceId   *model.Id `json:"space_id"`
	Name      string    `json:"name"`
	MimeType  string    `json:"mime_type"`
	SizeBytes int64     `json:"size_bytes"`
	Body      io.Reader `json:"-"`
}

// UploadFileResultData is the created record, as a built-in create answers it, plus its path.
type UploadFileResultData struct {
	Created  dmodel.DynamicFields `json:"created"`
	FilePath string               `json:"file_path"`
}

type UploadFileResult = dyn.OpResult[UploadFileResultData]

type GetFileMetaQuery struct {
	Path string `json:"path" query:"path"`
}

// GetFileMetaResult carries the file record plus its file_path.
type GetFileMetaResult = dyn.OpResult[dmodel.DynamicFields]

type DownloadFileQuery struct {
	Path        string `json:"path"`
	RangeHeader string `json:"range_header"`
}

type GetFileMetaByIdQuery struct {
	Id model.Id `json:"id"`
}

type DownloadFileByIdQuery struct {
	Id          model.Id `json:"id"`
	RangeHeader string   `json:"range_header"`
}

type DownloadFileResultData struct {
	Name          string        `json:"name"`
	MimeType      string        `json:"mime_type"`
	Body          io.ReadCloser `json:"-"`
	ContentLength *int64        `json:"content_length"`
	ContentRange  *string       `json:"content_range"`
}

type DownloadFileResult = dyn.OpResult[DownloadFileResultData]

type (
	GetFileByIdQuery  = composable.GetByIdQuery
	SearchFilesQuery  = composable.SearchQuery
	DeleteFileCommand = composable.DeleteCommand
)
