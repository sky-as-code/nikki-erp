package drive_file

import (
	"time"

	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/drive/domain"
	"github.com/sky-as-code/nikki-erp/modules/drive/infra/ent"
)

type DriveFileRepository interface {
	FindById(ctx crud.Context, id model.Id) (*domain.DriveFile, error)
	ExistsByOwnerParentNameFolder(ctx crud.Context, ownerRef model.Id, parentRef *model.Id, name string, isFolder bool) (bool, error)
	ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, fault.ValidationErrors)
	Search(ctx crud.Context, param SearchParam) (*crud.PagedResult[*domain.DriveFile], error)
	SearchByParent(ctx crud.Context, param SearchByParentParam) (*crud.PagedResult[*domain.DriveFile], error)
	GetDriveFileChildren(ctx crud.Context, parentId model.Id) ([]*domain.DriveFile, error)
	GetDriveFileParents(ctx crud.Context, driveFileId model.Id) ([]*domain.DriveFile, error)
	GetExpiredTrashedDriveFiles(ctx crud.Context, before time.Time) ([]*domain.DriveFile, error)

	Create(ctx crud.Context, driveFile *domain.DriveFile) (*domain.DriveFile, error)
	Update(ctx crud.Context, driveFile *domain.DriveFile, prevEtag model.Etag) (*domain.DriveFile, error)
	Overwrite(ctx crud.Context, driveFile *domain.DriveFile, prevEtag model.Etag) (*domain.DriveFile, error)
	DeleteById(ctx crud.Context, id model.Id) (int, error)
	DeleteByIds(ctx crud.Context, ids []model.Id) (int, error)
	BeginTransaction(ctx crud.Context) (*ent.Tx, error)
}

type SearchByParentParam struct {
	ParentFileId model.Id
	Predicate    *orm.Predicate
	Order        []orm.OrderOption
	Page         int
	Size         int
}

type FindByIdParam = GetDriveFileByIdQuery

type SearchParam struct {
	Predicate *orm.Predicate
	Order     []orm.OrderOption
	Page      int
	Size      int
}
