package drive_file

import (
	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/drive/domain"
)

type DriveFileRepository interface {
	Create(ctx crud.Context, driveFile *domain.DriveFile) (*domain.DriveFile, error)
	Update(ctx crud.Context, driveFile *domain.DriveFile) (*domain.DriveFile, error)
	FindById(ctx crud.Context, id model.Id) (*domain.DriveFile, error)
	ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, fault.ValidationErrors)
	Search(ctx crud.Context, param SearchParam) (*crud.PagedResult[*domain.DriveFile], error)
	DeleteById(ctx crud.Context, id model.Id) (int, error)
}

type FindByIdParam = GetDriveFileByIdQuery
type SearchParam = SearchDriveFileQuery
