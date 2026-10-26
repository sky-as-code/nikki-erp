package itstorage

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/storage/filestorage"
)

// FileExtService is the cross-module port for modules that serve stored files to their own
// principals, such as kiosks on the mTLS desktop API. It asserts no permission: the caller has
// already authenticated its principal and names the org the file must belong to. The file
// lifecycle (in use, release, duplicate) goes through core/storage.StorageConsumer instead.
type FileExtService interface {
	// OpenFile streams a file of orgId (or a tenant-wide one) for the given Range header. A file
	// outside that org answers not found, never forbidden, so its existence is not disclosed.
	OpenFile(ctx corectx.Context, query OpenFileQuery) (*OpenFileResult, error)
}

type OpenFileQuery struct {
	FileId      model.Id
	OrgId       *model.Id
	RangeHeader string
}

type OpenFileResult struct {
	// Found is false when the file does not exist or belongs to another org.
	Found    bool
	Name     string
	MimeType string
	Stream   *filestorage.StreamObjectResult
}
