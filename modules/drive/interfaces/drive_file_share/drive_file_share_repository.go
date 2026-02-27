package drive_file_share

import (
	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/drive/domain"
)

type DriveFileShareRepository interface {
	Create(ctx crud.Context, driveFile *domain.DriveFileShare) (*domain.DriveFileShare, error)
	Update(ctx crud.Context, driveFile *domain.DriveFileShare) (*domain.DriveFileShare, error)
	FindById(ctx crud.Context, id model.Id) (*domain.DriveFileShare, error)
	ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, fault.ValidationErrors)
	Search(ctx crud.Context, param SearchParam) (*crud.PagedResult[*domain.DriveFileShare], error)
	DeleteById(ctx crud.Context, id model.Id) (int, error)
}

type FindByIdParam = GetDriveFileShareByIdQuery
type SearchParam = SearchDriveFileShareQuery
