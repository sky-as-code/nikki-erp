package external

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// FileStorageExtService is iam's port onto IT Storage, for the user's avatar_file_path: it checks
// that a path names a usable file and removes the file a write replaced or cleared. It asserts no
// permission; the user's own application service has authorized the write.
type FileStorageExtService interface {
	AssertFileUsable(ctx corectx.Context, path string, orgId *model.Id) (*model.Id, error)
	DeleteFileByPath(ctx corectx.Context, path string) error
}
