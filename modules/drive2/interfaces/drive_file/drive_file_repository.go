package drive_file

import (
	"time"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/drive2/domain"
)

type DriveFileRepository interface {
	dyn.DynamicModelRepository

	DeleteOne(ctx corectx.Context, keys domain.DriveFile) (*dyn.OpResult[dyn.MutateResultData], error)
	Exists(ctx corectx.Context, keys []domain.DriveFile) (*dyn.OpResult[dyn.RepoExistsResult], error)
	Insert(ctx corectx.Context, driveFile domain.DriveFile) (*dyn.OpResult[int], error)
	GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (*dyn.OpResult[domain.DriveFile], error)
	Search(ctx corectx.Context, param dyn.RepoSearchParam) (*dyn.OpResult[dyn.PagedResultData[domain.DriveFile]], error)
	Update(ctx corectx.Context, driveFile domain.DriveFile) (*dyn.OpResult[dyn.MutateResultData], error)

	FindByIds(ctx corectx.Context, ids []model.Id) ([]domain.DriveFile, error)
	ExistsByOwnerParentNameFolder(ctx corectx.Context, ownerRef model.Id, parentRef *model.Id, name string, isFolder bool) (bool, error)
	ParseSearchGraph(criteria *string) (*dmodel.SearchGraph, ft.ValidationErrors)
	SearchAccessible(ctx corectx.Context, userId model.Id, param dyn.RepoSearchParam) (
		*dyn.OpResult[dyn.PagedResultData[domain.DriveFile]], error)
	GetRootFileByUser(ctx corectx.Context, userId model.Id, param dyn.RepoSearchParam) (
		*dyn.OpResult[dyn.PagedResultData[domain.DriveFile]], error)
	SearchByParent(ctx corectx.Context, param DriveFileSearchByParentParam) (
		*dyn.OpResult[dyn.PagedResultData[domain.DriveFile]], error)
	GetDriveFilesSharedByUser(ctx corectx.Context, userId model.Id, param dyn.RepoSearchParam) (
		*dyn.OpResult[dyn.PagedResultData[domain.DriveFile]], error)
	GetDriveFileChildren(ctx corectx.Context, parentId model.Id) ([]domain.DriveFile, error)
	GetDriveFileParents(ctx corectx.Context, driveFileId model.Id) ([]*domain.DriveFile, error)
	GetExpiredTrashedDriveFiles(ctx corectx.Context, before time.Time) ([]domain.DriveFile, error)

	InsertAncestors(ctx corectx.Context, fileId model.Id, ancestorIds []model.Id) error
	DeleteAncestorsByFileIds(ctx corectx.Context, fileIds []model.Id) error
	GetAncestorIds(ctx corectx.Context, fileId model.Id) ([]model.Id, error)
	GetAncestorIdsForFiles(ctx corectx.Context, fileIds []model.Id) (map[model.Id][]model.Id, error)

	Overwrite(ctx corectx.Context, driveFile domain.DriveFile, prevEtag model.Etag) (
		*dyn.OpResult[domain.DriveFile], error)
	DeleteByIds(ctx corectx.Context, ids []model.Id) (int, error)
}

type DriveFileSearchByParentParam struct {
	ParentFileId model.Id
	dyn.RepoSearchParam
}
