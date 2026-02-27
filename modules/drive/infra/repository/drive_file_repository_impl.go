package repository

import (
	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/drive/domain"
	"github.com/sky-as-code/nikki-erp/modules/drive/infra/ent"
	entDrivefile "github.com/sky-as-code/nikki-erp/modules/drive/infra/ent/drivefile"
	"github.com/sky-as-code/nikki-erp/modules/drive/interfaces/drive_file"
)

type driveFileRepository struct {
	client *ent.Client
}

func NewDriveFileRepository(client *ent.Client) drive_file.DriveFileRepository {
	return &driveFileRepository{client: client}
}

func BuildDriveFileDescriptor() *orm.EntityDescriptor {
	entity := ent.DriveFile{}
	builder := orm.DescribeEntity(entDrivefile.Label).
		Aliases("drive_files", "dri_files").
		Field(entDrivefile.FieldID, entity.ID).
		Field(entDrivefile.FieldEtag, entity.Etag).
		Field(entDrivefile.FieldCreatedAt, entity.CreatedAt).
		Field(entDrivefile.FieldUpdatedAt, entity.UpdatedAt).
		Field(entDrivefile.FieldDeletedAt, entity.DeletedAt).
		Field(entDrivefile.FieldScopeType, entity.ScopeType).
		Field(entDrivefile.FieldScopeRef, entity.ScopeRef).
		Field(entDrivefile.FieldOwnerRef, entity.OwnerRef).
		Field(entDrivefile.FieldParentFileRef, entity.ParentFileRef).
		Field(entDrivefile.FieldName, entity.Name).
		Field(entDrivefile.FieldMime, entity.Mime).
		Field(entDrivefile.FieldIsFolder, entity.IsFolder).
		Field(entDrivefile.FieldSize, entity.Size).
		Field(entDrivefile.FieldPath, entity.Path).
		Field(entDrivefile.FieldStorage, entity.Storage).
		Field(entDrivefile.FieldVisiblity, entity.Visiblity)

	return builder.Descriptor()
}

// Create implements drive_file.DriveFileRepository.
func (d *driveFileRepository) Create(ctx crud.Context, driveFile *domain.DriveFile) (*domain.DriveFile, error) {
	panic("unimplemented")
}

// DeleteById implements drive_file.DriveFileRepository.
func (d *driveFileRepository) DeleteById(ctx crud.Context, id model.Id) (int, error) {
	panic("unimplemented")
}

// FindById implements drive_file.DriveFileRepository.
func (d *driveFileRepository) FindById(ctx crud.Context, id model.Id) (*domain.DriveFile, error) {
	panic("unimplemented")
}

// ParseSearchGraph implements drive_file.DriveFileRepository.
func (d *driveFileRepository) ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, fault.ValidationErrors) {
	panic("unimplemented")
}

// Search implements drive_file.DriveFileRepository.
func (d *driveFileRepository) Search(ctx crud.Context, param drive_file.SearchParam) (*crud.PagedResult[*domain.DriveFile], error) {
	panic("unimplemented")
}

// Update implements drive_file.DriveFileRepository.
func (d *driveFileRepository) Update(ctx crud.Context, driveFile *domain.DriveFile) (*domain.DriveFile, error) {
	panic("unimplemented")
}
