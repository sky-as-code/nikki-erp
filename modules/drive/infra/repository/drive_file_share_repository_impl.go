package repository

import (
	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/drive/domain"
	"github.com/sky-as-code/nikki-erp/modules/drive/infra/ent"
	entDrivefileshare "github.com/sky-as-code/nikki-erp/modules/drive/infra/ent/drivefileshare"
	"github.com/sky-as-code/nikki-erp/modules/drive/interfaces/drive_file_share"
)

type driveFileShareRepository struct {
	client *ent.Client
}

func NewDriveFileShareRepository(client *ent.Client) drive_file_share.DriveFileShareRepository {
	return &driveFileShareRepository{client: client}
}

func (d *driveFileShareRepository) Create(ctx crud.Context, driveFile *domain.DriveFileShare) (*domain.DriveFileShare, error) {
	panic("unimplemented")
}

func (d *driveFileShareRepository) DeleteById(ctx crud.Context, id model.Id) (int, error) {
	panic("unimplemented")
}

func (d *driveFileShareRepository) FindById(ctx crud.Context, id model.Id) (*domain.DriveFileShare, error) {
	panic("unimplemented")
}

func (d *driveFileShareRepository) ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, fault.ValidationErrors) {
	panic("unimplemented")
}

func (d *driveFileShareRepository) Search(ctx crud.Context, param drive_file_share.SearchParam) (*crud.PagedResult[*domain.DriveFileShare], error) {
	panic("unimplemented")
}

func (d *driveFileShareRepository) Update(ctx crud.Context, driveFile *domain.DriveFileShare) (*domain.DriveFileShare, error) {
	panic("unimplemented")
}

func BuildDriveFileShareDescriptor() *orm.EntityDescriptor {
	entity := ent.DriveFileShare{}
	builder := orm.DescribeEntity(entDrivefileshare.Label).
		Aliases("drive_file_shares", "dri_file_shares").
		Field(entDrivefileshare.FieldID, entity.ID).
		Field(entDrivefileshare.FieldEtag, entity.Etag).
		Field(entDrivefileshare.FieldCreatedAt, entity.CreatedAt).
		Field(entDrivefileshare.FieldUpdatedAt, entity.UpdatedAt).
		Field(entDrivefileshare.FieldScopeType, entity.ScopeType).
		Field(entDrivefileshare.FieldScopeRef, entity.ScopeRef).
		Field(entDrivefileshare.FieldFileRef, entity.FileRef).
		Field(entDrivefileshare.FieldUserRef, entity.UserRef).
		Field(entDrivefileshare.FieldPermission, entity.Permission)

	return builder.Descriptor()
}
