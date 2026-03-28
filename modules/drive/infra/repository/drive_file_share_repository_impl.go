package repository

import (
	"fmt"
	"strings"

	entsql "entgo.io/ent/dialect/sql"
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
		combined = orm.Predicate(entsql.AndPredicates(combined, *param.Predicate))
	}

	return d.Search(ctx, drive_file_share.SearchParam{
		Predicate: &combined,
		Order:     param.Order,
		Page:      param.Page,
		Size:      param.Size,
	})
}

func (d *driveFileShareRepository) ListResolvedByFileRefs(
	ctx crud.Context,
	fileRef model.Id,
	refs []model.Id,
	excludedUserRefs []model.Id,
	page int,
	size int,
) (*crud.PagedResult[*domain.DriveFileShare], error) {
	if len(refs) == 0 {
		return &crud.PagedResult[*domain.DriveFileShare]{Items: []*domain.DriveFileShare{}, Total: 0}, nil
	}
	if size <= 0 {
		size = 20
	}
	if page < 0 {
		page = 0
	}

	placeholders := make([]string, 0, len(refs))
	args := make([]any, 0, len(refs)+len(excludedUserRefs)+2)
	for i, id := range refs {
		placeholders = append(placeholders, fmt.Sprintf("$%d", i+1))
		args = append(args, string(id))
	}
	inClause := strings.Join(placeholders, ",")

	excludeClause := ""
	if len(excludedUserRefs) > 0 {
		ex := make([]string, 0, len(excludedUserRefs))
		for i := range excludedUserRefs {
			p := len(args) + i + 1
			ex = append(ex, fmt.Sprintf("$%d", p))
		}
		for _, uid := range excludedUserRefs {
			args = append(args, string(uid))
		}
		excludeClause = fmt.Sprintf(" AND user_ref NOT IN (%s) ", strings.Join(ex, ","))
	}

	countSQL := fmt.Sprintf(`
		SELECT COUNT(1)
		FROM (
			SELECT user_ref
			FROM dri_file_shares
			WHERE file_ref IN (%s) %s
			GROUP BY user_ref
		) t
	`, inClause, excludeClause)

	dbClient := d.client.DB()
	var total int
	if err := dbClient.QueryRowContext(ctx.InnerContext(), countSQL, args...).Scan(&total); err != nil {
		return nil, err
	}
	if total == 0 {
		return &crud.PagedResult[*domain.DriveFileShare]{Items: []*domain.DriveFileShare{}, Total: 0}, nil
	}

	offset := page * size
	// Append target file ref (to distinguish direct vs inherited),
	// then limit/offset for paging.
	dataArgs := append(args, string(fileRef), size, offset)
	targetPos := len(args) + 1
	limitPos := len(args) + 2
	offsetPos := len(args) + 3
	dataSQL := fmt.Sprintf(`
		SELECT
			user_ref,
			MAX(
				CASE permission
					-- direct share on the target file
					WHEN 'edit-trash' THEN
						CASE WHEN file_ref = $%d THEN 6 ELSE 3 END
					WHEN 'edit' THEN
						CASE WHEN file_ref = $%d THEN 5 ELSE 2 END
					WHEN 'view' THEN
						CASE WHEN file_ref = $%d THEN 4 ELSE 1 END
					ELSE 0
				END
			) AS perm_rank
		FROM dri_file_shares
		WHERE file_ref IN (%s) %s
		GROUP BY user_ref
		ORDER BY perm_rank DESC, user_ref ASC
		LIMIT $%d OFFSET $%d
	`, targetPos, targetPos, targetPos, inClause, excludeClause, limitPos, offsetPos)

	rows, err := dbClient.QueryContext(ctx.InnerContext(), dataSQL, dataArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]*domain.DriveFileShare, 0, size)
	for rows.Next() {
		var userRef string
		var rank int
		if err := rows.Scan(&userRef, &rank); err != nil {
			return nil, err
		}
		perm := enum.DriveFilePermNone
		switch rank {
		// direct share ranks
		case 6:
			perm = enum.DriveFilePermEditTrash
		case 5:
			perm = enum.DriveFilePermEdit
		case 4:
			perm = enum.DriveFilePermView
		// inherited share ranks (came from ancestor file_ref)
		case 3:
			perm = enum.DriveFilePermInheritedEditTrash
		case 2:
			perm = enum.DriveFilePermInheritedEdit
		case 1:
			perm = enum.DriveFilePermInheritedView
		}
		if perm == enum.DriveFilePermNone {
			continue
		}
		items = append(items, &domain.DriveFileShare{
			FileRef:    fileRef,
			UserRef:    model.Id(userRef),
			Permission: perm,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &crud.PagedResult[*domain.DriveFileShare]{
		Items: items,
		Total: total,
		Page:  page,
		Size:  size,
	}, nil
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
