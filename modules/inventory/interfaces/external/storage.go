package external

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// FileStorageExtService is Inventory's port onto IT Storage, for the template's and variant's
// `*_image_file_path`: it checks that a path names a file of the record's org and removes the file
// a write replaced or cleared. It asserts no permission; the resource's own application service
// has authorized the write.
type FileStorageExtService interface {
	AssertFileUsable(ctx corectx.Context, path string, orgId *model.Id) (*model.Id, error)
	DeleteFileByPath(ctx corectx.Context, path string) error
}
