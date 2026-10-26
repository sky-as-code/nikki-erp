package itstorage

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// FileExtService is the cross-module port other modules use for their `*_file_path` fields.
// It asserts no permission: the caller's own application service has already authorized the
// write of the resource that carries the path.
type FileExtService interface {
	// AssertFileUsable returns the file id when path names an existing file inside orgId's space
	// (the tenant's when orgId is nil), and a client error otherwise.
	AssertFileUsable(ctx corectx.Context, path string, orgId *model.Id) (*model.Id, error)
	// DeleteFileByPath removes the file row, its size from every ancestor space, and the stored
	// object (after commit). A path naming no file is not an error.
	DeleteFileByPath(ctx corectx.Context, path string) error
}
