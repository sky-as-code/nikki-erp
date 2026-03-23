package repository

import (
	"entgo.io/ent/dialect/sql"
	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	db "github.com/sky-as-code/nikki-erp/modules/core/database"
	"github.com/sky-as-code/nikki-erp/modules/drive/domain"
	"github.com/sky-as-code/nikki-erp/modules/drive/enum"
	"github.com/sky-as-code/nikki-erp/modules/drive/infra/ent"
	entDrivefileshare "github.com/sky-as-code/nikki-erp/modules/drive/infra/ent/drivefileshare"
	"github.com/sky-as-code/nikki-erp/modules/drive/infra/ent/predicate"
	"github.com/sky-as-code/nikki-erp/modules/drive/interfaces/drive_file_share"
)

type driveFileShareRepository struct {
	client *ent.Client
}

func NewDriveFileShareRepository(client *ent.Client) drive_file_share.DriveFileShareRepository {
	return &driveFileShareRepository{client: client}
}

func (d *driveFileShareRepository) driveFileShareClient(ctx crud.Context) *ent.DriveFileShareClient {
	tx, isOk := ctx.GetDbTranx().(*ent.Tx)
	if isOk {
		return tx.DriveFileShare
	}
	return d.client.DriveFileShare
}

func (d *driveFileShareRepository) Create(ctx crud.Context, share *domain.DriveFileShare) (*domain.DriveFileShare, error) {
	creation := d.driveFileShareClient(ctx).Create().
		SetID(*share.Id).
		SetEtag(*share.Etag).
		SetFileRef(string(share.FileRef)).
		SetUserRef(string(share.UserRef)).
		SetPermission(enum.DriveFileSharePermName[share.Permission])

	return db.Mutate(ctx.InnerContext(), creation, ent.IsNotFound, entToDriveFileShare)
}

func (d *driveFileShareRepository) Update(ctx crud.Context, share *domain.DriveFileShare, prevEtag model.Etag) (*domain.DriveFileShare, error) {
	update := d.driveFileShareClient(ctx).UpdateOneID(*share.Id).
		Where(entDrivefileshare.Etag(string(prevEtag)))

	if share.Etag != nil {
		update.SetEtag(*share.Etag)
	}
	if share.Permission != 0 {
		update.SetPermission(enum.DriveFileSharePermName[share.Permission])
	}

	return db.Mutate(ctx.InnerContext(), update, ent.IsNotFound, entToDriveFileShare)
}

func (d *driveFileShareRepository) FindById(ctx crud.Context, id model.Id) (*domain.DriveFileShare, error) {
	dbQuery := d.driveFileShareClient(ctx).Query().
		Where(entDrivefileshare.ID(id))
	return db.FindOne(ctx.InnerContext(), dbQuery, ent.IsNotFound, entToDriveFileShare)
}

func (d *driveFileShareRepository) ListByFileRefsAndUserRef(ctx crud.Context, driveFileIds []model.Id, userId model.Id) ([]*domain.DriveFileShare, error) {
	if len(driveFileIds) == 0 {
		return []*domain.DriveFileShare{}, nil
	}
	refs := make([]string, 0, len(driveFileIds))
	for _, id := range driveFileIds {
		refs = append(refs, string(id))
	}
	query := d.driveFileShareClient(ctx).Query().
		Where(
			entDrivefileshare.FileRefIn(refs...),
			entDrivefileshare.UserRef(string(userId)),
		)
	return db.List(ctx.InnerContext(), query, entToDriveFileShares)
}

func (d *driveFileShareRepository) ListByFileRef(ctx crud.Context, param drive_file_share.ListByFileRefParam) (*crud.PagedResult[*domain.DriveFileShare], error) {
	var parentPred predicate.DriveFileShare
	parentPred = entDrivefileshare.FileRef(string(param.FileRef))
	combined := orm.Predicate(parentPred)
	if param.Predicate != nil {
		combined = orm.Predicate(sql.AndPredicates(combined, *param.Predicate))
	}

	return d.Search(ctx, drive_file_share.SearchParam{
		Predicate: &combined,
		Order:     param.Order,
		Page:      param.Page,
		Size:      param.Size,
	})
}

func (d *driveFileShareRepository) ListByUserRef(ctx crud.Context, userRef model.Id) ([]*domain.DriveFileShare, error) {
	query := d.driveFileShareClient(ctx).Query().
		Where(entDrivefileshare.UserRef(string(userRef)))
	return db.List(ctx.InnerContext(), query, entToDriveFileShares)
}

func (d *driveFileShareRepository) DeleteById(ctx crud.Context, id model.Id) (int, error) {
	return d.driveFileShareClient(ctx).Delete().
		Where(entDrivefileshare.ID(id)).
		Exec(ctx.InnerContext())
}

func (d *driveFileShareRepository) ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, fault.ValidationErrors) {
	return db.ParseSearchGraphStr[ent.DriveFileShare, domain.DriveFileShare](criteria, entDrivefileshare.Label)
}

func (d *driveFileShareRepository) Search(ctx crud.Context, param drive_file_share.SearchParam) (*crud.PagedResult[*domain.DriveFileShare], error) {
	query := d.driveFileShareClient(ctx).Query()
	return db.Search(
		ctx.InnerContext(),
		param.Predicate,
		param.Order,
		db.PagingOptions{
			Page: param.Page,
			Size: param.Size,
		},
		query,
		entToDriveFileShares,
	)
}

func BuildDriveFileShareDescriptor() *orm.EntityDescriptor {
	entity := ent.DriveFileShare{}
	builder := orm.DescribeEntity(entDrivefileshare.Label).
		Aliases("drive_file_shares", "dri_file_shares").
		Field(entDrivefileshare.FieldID, entity.ID).
		Field(entDrivefileshare.FieldEtag, entity.Etag).
		Field(entDrivefileshare.FieldCreatedAt, entity.CreatedAt).
		Field(entDrivefileshare.FieldUpdatedAt, entity.UpdatedAt).
		Field(entDrivefileshare.FieldFileRef, entity.FileRef).
		Field(entDrivefileshare.FieldUserRef, entity.UserRef).
		Field(entDrivefileshare.FieldPermission, entity.Permission)

	return builder.Descriptor()
}
