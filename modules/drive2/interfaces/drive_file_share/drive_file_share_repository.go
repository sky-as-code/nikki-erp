package drive_file_share

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/drive2/domain/models"
)

type DriveFileShareRepository interface {
	dyn.DynamicModelRepository

	DeleteOne(ctx corectx.Context, keys models.DriveFileShare) (*dyn.OpResult[dyn.MutateResultData], error)
	Exists(ctx corectx.Context, keys []models.DriveFileShare) (*dyn.OpResult[dyn.RepoExistsResult], error)
	Insert(ctx corectx.Context, share models.DriveFileShare) (*dyn.OpResult[int], error)
	GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (*dyn.OpResult[models.DriveFileShare], error)
	Search(ctx corectx.Context, param dyn.RepoSearchParam) (
		*dyn.OpResult[dyn.PagedResultData[models.DriveFileShare]], error)
	Update(ctx corectx.Context, share models.DriveFileShare) (*dyn.OpResult[dyn.MutateResultData], error)

	ListByFileRef(ctx corectx.Context, param ListDriveFileShareByFileRefParam) (
		*dyn.OpResult[dyn.PagedResultData[models.DriveFileShare]], error)
	ListResolvedByFileRefs(ctx corectx.Context, fileRef model.Id, refs []model.Id, excludedUserRefs []model.Id, page int, size int) (
		*dyn.OpResult[dyn.PagedResultData[models.DriveFileShare]], error)
	ListByFileRefsAndUserRef(ctx corectx.Context, driveFileIds []model.Id, userId model.Id) ([]models.DriveFileShare, error)
	ListByUserRef(ctx corectx.Context, userRef model.Id) ([]models.DriveFileShare, error)
	ParseSearchGraph(criteria *string) (*dmodel.SearchGraph, ft.ValidationErrors)
}

type ListDriveFileShareByFileRefParam struct {
	FileRef model.Id
	dyn.RepoSearchParam
}
