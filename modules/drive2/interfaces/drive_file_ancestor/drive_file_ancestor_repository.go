package drive_file_ancestor

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/drive2/domain/models"
)

// DriveFileAncestorRepository persists drive file closure rows (dri_file_ancestors).
type DriveFileAncestorRepository interface {
	dyn.DynamicModelRepository

	Insert(ctx corectx.Context, row models.DriveFileAncestor) (*dyn.OpResult[int], error)
	InsertBulk(ctx corectx.Context, rows []models.DriveFileAncestor) (*dyn.OpResult[int], error)
	Update(ctx corectx.Context, row models.DriveFileAncestor) (*dyn.OpResult[dyn.MutateResultData], error)
	GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (*dyn.OpResult[models.DriveFileAncestor], error)
	Search(ctx corectx.Context, param dyn.RepoSearchParam) (
		*dyn.OpResult[dyn.PagedResultData[models.DriveFileAncestor]], error)
	DeleteOne(ctx corectx.Context, keys models.DriveFileAncestor) (*dyn.OpResult[dyn.MutateResultData], error)

	DeleteByFileRefs(ctx corectx.Context, fileRefs []model.Id) (deleted int, err error)
}
