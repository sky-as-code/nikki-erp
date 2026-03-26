package drive_file_share

import (
	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/drive/domain"
)

type DriveFileShareRepository interface {
	Create(ctx crud.Context, share *domain.DriveFileShare) (*domain.DriveFileShare, error)
	Update(ctx crud.Context, share *domain.DriveFileShare, prevEtag model.Etag) (*domain.DriveFileShare, error)
	FindById(ctx crud.Context, id model.Id) (*domain.DriveFileShare, error)
	ListByFileRef(ctx crud.Context, param ListByFileRefParam) (*crud.PagedResult[*domain.DriveFileShare], error)
	ListResolvedByFileRefs(ctx crud.Context, fileRef model.Id, refs []model.Id, excludedUserRefs []model.Id, page int, size int) (*crud.PagedResult[*domain.DriveFileShare], error)
	// ListByFileRefsAndUserRef returns shares for the given user on any of the listed drive files (file_ref IN driveFileIds AND user_ref = userId).
	ListByFileRefsAndUserRef(ctx crud.Context, driveFileIds []model.Id, userId model.Id) ([]*domain.DriveFileShare, error)
	ListByUserRef(ctx crud.Context, userRef model.Id) ([]*domain.DriveFileShare, error)
	ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, fault.ValidationErrors)
	Search(ctx crud.Context, param SearchParam) (*crud.PagedResult[*domain.DriveFileShare], error)
	DeleteById(ctx crud.Context, id model.Id) (int, error)
}

type ListByFileRefParam struct {
	SearchParam
	FileRef model.Id
}

type SearchParam struct {
	Predicate *orm.Predicate
	Order     []orm.OrderOption
	Page      int
	Size      int
}
