package repository

import (
	"database/sql"
	"strings"
	"time"

	entSql "entgo.io/ent/dialect/sql"
	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	db "github.com/sky-as-code/nikki-erp/modules/core/database"
	"github.com/sky-as-code/nikki-erp/modules/drive/domain"
	"github.com/sky-as-code/nikki-erp/modules/drive/enum"
	"github.com/sky-as-code/nikki-erp/modules/drive/infra/ent"
	entDrivefile "github.com/sky-as-code/nikki-erp/modules/drive/infra/ent/drivefile"
	entDrivefileshare "github.com/sky-as-code/nikki-erp/modules/drive/infra/ent/drivefileshare"
	"github.com/sky-as-code/nikki-erp/modules/drive/infra/ent/predicate"
	"github.com/sky-as-code/nikki-erp/modules/drive/interfaces/drive_file"
)

var deletedAtNotDeleted = time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)

func ptrOrEmpty(p *model.Id) model.Id {
	if p == nil {
		return ""
	}
	return *p
}

type driveFileRepository struct {
	client *ent.Client
}

func NewDriveFileRepository(client *ent.Client) drive_file.DriveFileRepository {
	return &driveFileRepository{client: client}
}

var notPendingDelete = entDrivefile.StatusNEQ(enum.DriveFileStatusName[enum.DriveFileStatusPendingDelete])

func (d *driveFileRepository) driveFileClient(ctx crud.Context) *ent.DriveFileClient {
	tx, isOk := ctx.GetDbTranx().(*ent.Tx)
	if isOk {
		return tx.DriveFile
	}

	return d.client.DriveFile
}

func (d *driveFileRepository) ExistsByOwnerParentNameFolder(ctx crud.Context, ownerRef model.Id, parentRef *model.Id, name string, isFolder bool) (bool, error) {
	ownerPred := entDrivefile.OwnerRef(string(ownerRef))
	namePred := entDrivefile.Name(name)
	folderPred := entDrivefile.IsFolder(isFolder)
	deletedPred := entDrivefile.DeletedAt(deletedAtNotDeleted)

	var parentPred predicate.DriveFile
	if parentRef != nil && *parentRef != "" {
		parentPred = entDrivefile.ParentFileRef(string(*parentRef))
	} else {
		parentPred = entDrivefile.ParentFileRefIsNil()
	}

	query := d.driveFileClient(ctx).Query().
		Where(ownerPred, parentPred, namePred, folderPred, deletedPred).
		Limit(1)

	exists, err := query.Exist(ctx.InnerContext())
	return exists, err
}

func (d *driveFileRepository) Create(ctx crud.Context, driveFile *domain.DriveFile) (
	*domain.DriveFile, error) {
	creation := d.driveFileClient(ctx).Create().
		SetID(*driveFile.Id).
		SetEtag(*driveFile.Etag).
		SetOwnerRef(string(ptrOrEmpty(driveFile.OwnerRef))).
		SetName(driveFile.Name).
		SetMime(driveFile.MINE).
		SetIsFolder(driveFile.IsFolder).
		SetSize(int64(driveFile.Size)).
		SetStoragePath(driveFile.StoragePath).
		SetStorageKey(driveFile.StorageKey).
		SetStorage(enum.DriveFileStorageName[driveFile.Storage]).
		SetVisibility(enum.DriveFileVisibilityName[driveFile.Visibility]).
		SetStatus(enum.DriveFileStatusName[driveFile.Status]).
		SetDeletedAt(deletedAtNotDeleted)

	if driveFile.ParentDriveFileRef != nil && *driveFile.ParentDriveFileRef != "" {
		creation.SetParentFileRef(string(*driveFile.ParentDriveFileRef))
	}

	creation.SetNillableMaterializedPath(driveFile.MaterializedPath)

	result, err := db.Mutate(ctx.InnerContext(), creation, ent.IsNotFound, entToDriveFile)
	if err != nil {
		if ent.IsConstraintError(err) && isDriveFileUniqueNameConstraint(err) {
			return nil, &fault.ClientError{
				Code:    "duplicate_name",
				Details: fault.ValidationErrors{"name": "a file or folder with this name already exists in this location"},
			}
		}
		return nil, err
	}
	return result, nil
}

func isDriveFileUniqueNameConstraint(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "duplicate") ||
		strings.Contains(msg, "unique") ||
		strings.Contains(msg, "drivefile_owner_ref")
}

func (d *driveFileRepository) Update(ctx crud.Context, driveFile *domain.DriveFile, prevEtag model.Etag) (*domain.DriveFile, error) {
	update := d.driveFileClient(ctx).UpdateOneID(*driveFile.Id).
		Where(entDrivefile.Etag(prevEtag))

	if driveFile.Name != "" {
		update.SetName(driveFile.Name)
	}

	if driveFile.MINE != "" {
		update.SetMime(driveFile.MINE)
	}

	if driveFile.Etag != nil {
		update.SetEtag(*driveFile.Etag)
	}

	if driveFile.Visibility != 0 {
		update.SetVisibility(enum.DriveFileVisibilityName[driveFile.Visibility])
	}

	if driveFile.Status != 0 {
		update.SetStatus(enum.DriveFileStatusName[driveFile.Status])
	}

	if driveFile.Size > 0 {
		update.SetSize(int64(driveFile.Size))
	}

	if driveFile.DeletedAt != nil {
		update.SetDeletedAt(*driveFile.DeletedAt)
	}

	if driveFile.MaterializedPath != nil {
		update.SetNillableMaterializedPath(driveFile.MaterializedPath)
	}

	return db.Mutate(ctx.InnerContext(), update, ent.IsNotFound, entToDriveFile)
}

func (d *driveFileRepository) Overwrite(
	ctx crud.Context,
	driveFile *domain.DriveFile,
	prevEtag model.Etag) (
	*domain.DriveFile, error) {
	update := d.driveFileClient(ctx).UpdateOneID(*driveFile.Id).
		Where(entDrivefile.Etag(prevEtag))

	update.SetName(driveFile.Name)
	update.SetMime(driveFile.MINE)
	update.SetEtag(*driveFile.Etag)
	update.SetVisibility(enum.DriveFileVisibilityName[driveFile.Visibility])
	update.SetStatus(enum.DriveFileStatusName[driveFile.Status])
	update.SetSize(int64(driveFile.Size))
	update.SetDeletedAt(*driveFile.DeletedAt)
	update.SetNillableParentFileRef(driveFile.ParentDriveFileRef)
	update.SetNillableMaterializedPath(driveFile.MaterializedPath)

	return db.Mutate(ctx.InnerContext(), update, ent.IsNotFound, entToDriveFile)
}

func (d *driveFileRepository) FindById(ctx crud.Context, id model.Id) (*domain.DriveFile, error) {
	dbQuery := d.driveFileClient(ctx).Query().
		Where(entDrivefile.ID(id), notPendingDelete)
	return db.FindOne(ctx.InnerContext(), dbQuery, ent.IsNotFound, entToDriveFile)
}

func (d *driveFileRepository) DeleteById(ctx crud.Context, id model.Id) (int, error) {
	return d.driveFileClient(ctx).Delete().
		Where(entDrivefile.ID(id)).
		Exec(ctx.InnerContext())
}

func (d *driveFileRepository) DeleteByIds(ctx crud.Context, ids []model.Id) (int, error) {
	return d.driveFileClient(ctx).Delete().
		Where(entDrivefile.IDIn(ids...)).
		Exec(ctx.InnerContext())
}

func (d *driveFileRepository) ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, fault.ValidationErrors) {
	return db.ParseSearchGraphStr[ent.DriveFile, domain.DriveFile](criteria, entDrivefile.Label)
}

func (d *driveFileRepository) Search(ctx crud.Context, param drive_file.SearchParam) (*crud.PagedResult[*domain.DriveFile], error) {
	query := d.driveFileClient(ctx).Query()
	return db.Search(
		ctx.InnerContext(),
		param.Predicate,
		param.Order,
		db.PagingOptions{
			Page: param.Page,
			Size: param.Size,
		},
		query,
		entToDriveFiles,
	)
}

func (d *driveFileRepository) SearchAccessible(
	ctx crud.Context,
	userId model.Id,
	param drive_file.SearchParam,
) (*crud.PagedResult[*domain.DriveFile], error) {
	accessiblePred := orm.Predicate(entSql.AndPredicates(
		notPendingDelete,
		permissionPredicate(string(userId)),
	))

	combinedPred := accessiblePred
	if param.Predicate != nil {
		combinedPred = orm.Predicate(entSql.AndPredicates(accessiblePred, *param.Predicate))
	}

	return d.Search(ctx, drive_file.SearchParam{
		Predicate: &combinedPred,
		Order:     param.Order,
		Page:      param.Page,
		Size:      param.Size,
	})
}

func (d *driveFileRepository) GetRootFileByUser(
	ctx crud.Context,
	userId model.Id,
	param drive_file.SearchParam,
) (*crud.PagedResult[*domain.DriveFile], error) {
	rootPred := entDrivefile.ParentFileRefIsNil()
	// Root listing is only for the user's own roots (owner_ref = userId).
	// Shares do not apply here.
	accessiblePred := orm.Predicate(entSql.AndPredicates(
		notPendingDelete,
		rootPred,
		entDrivefile.OwnerRef(string(userId)),
	))

	combinedPred := accessiblePred
	if param.Predicate != nil {
		combinedPred = orm.Predicate(entSql.AndPredicates(accessiblePred, *param.Predicate))
	}

	return d.Search(ctx, drive_file.SearchParam{
		Predicate: &combinedPred,
		Order:     param.Order,
		Page:      param.Page,
		Size:      param.Size,
	})
}


func permissionPredicate(userId string) predicate.DriveFile {
	return func(s *entSql.Selector) {
		fileIDCol := s.C(entDrivefile.FieldID)
		filePathCol := s.C(entDrivefile.FieldMaterializedPath)
		shareTable := entSql.Table(entDrivefileshare.Table).As("dfs")
		shareUserRefCol := shareTable.C(entDrivefileshare.FieldUserRef)

		directShareSub := entSql.
			Select("id").
			From(shareTable).
			Where(
				entSql.And(
					entSql.EQ(shareUserRefCol, userId),
					entSql.ExprP("dfs."+entDrivefileshare.FieldFileRef+" = "+fileIDCol),
				),
			)

		// Path format: /id1/id2/... (no trailing slash). Match ancestor id as /id/ mid-path or path suffix .../id.
		ancestorShareSub := entSql.
			Select("id").
			From(shareTable).
			Where(
				entSql.And(
					entSql.EQ(shareUserRefCol, userId),
					entSql.Or(
						entSql.ExprP(filePathCol+" LIKE ('%/' || dfs."+entDrivefileshare.FieldFileRef+" || '/%')"),
						entSql.ExprP(filePathCol+" LIKE ('%/' || dfs."+entDrivefileshare.FieldFileRef+")"),
					),
				),
			)

		s.Where(
			entSql.Or(
				entSql.EQ(s.C(entDrivefile.FieldOwnerRef), userId),
				entSql.Exists(directShareSub),
				entSql.Exists(ancestorShareSub),
			),
		)
	}
}

func (d *driveFileRepository) SearchByParent(ctx crud.Context, param drive_file.SearchByParentParam) (*crud.PagedResult[*domain.DriveFile], error) {
	var parentPred predicate.DriveFile

	if param.ParentFileId == "" {
		parentPred = entDrivefile.ParentFileRefIsNil()
	} else {
		parentPred = entDrivefile.ParentFileRef(string(param.ParentFileId))
	}

	combined := orm.Predicate(entSql.AndPredicates(parentPred, notPendingDelete))
	if param.Predicate != nil {
		combined = orm.Predicate(entSql.AndPredicates(combined, *param.Predicate))
	}

	return d.Search(ctx, drive_file.SearchParam{
		Predicate: &combined,
		Order:     param.Order,
		Page:      param.Page,
		Size:      param.Size,
	})
}

func (d *driveFileRepository) GetDriveFilesSharedByUser(ctx crud.Context, userId model.Id, param drive_file.SearchParam) (*crud.PagedResult[*domain.DriveFile], error) {
	driveFilesQuery := d.client.DriveFileShare.
		Query().
		Where(entDrivefileshare.UserRef(string(userId))).
		QueryDriveFiles().
		Where(notPendingDelete)

	return db.Search(
		ctx.InnerContext(),
		param.Predicate,
		param.Order,
		db.PagingOptions{
			Page: param.Page,
			Size: param.Size,
		},
		driveFilesQuery,
		entToDriveFiles,
	)
}

func (d *driveFileRepository) GetDriveFileChildren(ctx crud.Context, parentId model.Id) ([]*domain.DriveFile, error) {
	dbClient := d.client.DB()

	rows, err := dbClient.QueryContext(ctx,
		`
		WITH RECURSIVE subtree AS (
			SELECT 
				id, etag, created_at, updated_at, deleted_at, owner_ref,
				materialized_path,
				name, mime, is_folder, size, storage_path, storage_key, storage, visibility, status, parent_file_ref
			FROM dri_files WHERE id = $1
			UNION ALL
			SELECT 
				f.id, f.etag, f.created_at, f.updated_at, f.deleted_at, f.owner_ref,
				f.materialized_path,
				f.name, f.mime, f.is_folder, f.size, f.storage_path, f.storage_key, f.storage, f.visibility, f.status, f.parent_file_ref
			FROM dri_files f 
			JOIN subtree s ON f.parent_file_ref = s.id
		)

		SELECT 
			id, etag, created_at, updated_at, deleted_at, owner_ref,
			materialized_path,
			name, mime, is_folder, size, storage_path, storage_key, storage, visibility, status, parent_file_ref
		FROM subtree
		WHERE status != $2 and id != $1
		`, parentId, enum.DriveFileStatusName[enum.DriveFileStatusPendingDelete])
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var driveFiles []*domain.DriveFile
	for rows.Next() {
		driveFile := &domain.DriveFile{}
		d.scanDriveFileFromRow(rows, driveFile)

		driveFiles = append(driveFiles, driveFile)
	}

	return driveFiles, nil
}

func (d *driveFileRepository) GetDriveFileParents(ctx crud.Context, parentId model.Id) ([]*domain.DriveFile, error) {
	dbClient := d.client.DB()
	rows, err := dbClient.QueryContext(ctx,
		`
		WITH RECURSIVE ancestors AS (
			SELECT 
				id, etag, created_at, updated_at, deleted_at, owner_ref,
				materialized_path,
				name, mime, is_folder, size, storage_path, storage_key, storage, visibility, status, parent_file_ref
			FROM dri_files WHERE id = $1
			UNION ALL
			SELECT 
				f.id, f.etag, f.created_at, f.updated_at, f.deleted_at, f.owner_ref,
				f.materialized_path,
				f.name, f.mime, f.is_folder, f.size, f.storage_path, f.storage_key, f.storage, f.visibility, f.status, f.parent_file_ref
			FROM dri_files f
			JOIN ancestors a ON f.id = a.parent_file_ref
		)
		SELECT 
			id, etag, created_at, updated_at, deleted_at, owner_ref,
			materialized_path,
			name, mime, is_folder, size, storage_path, storage_key, storage, visibility, status, parent_file_ref
		FROM ancestors
		WHERE status != $2
		`, parentId, enum.DriveFileStatusName[enum.DriveFileStatusPendingDelete])
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var driveFiles []*domain.DriveFile
	for rows.Next() {
		driveFile := &domain.DriveFile{}
		d.scanDriveFileFromRow(rows, driveFile)

		driveFiles = append(driveFiles, driveFile)
	}

	return driveFiles, nil
}

func (d *driveFileRepository) GetExpiredTrashedDriveFiles(ctx crud.Context, before time.Time) ([]*domain.DriveFile, error) {
	entities, err := d.driveFileClient(ctx).
		Query().
		Where(
			entDrivefile.DeletedAtNEQ(deletedAtNotDeleted),
			entDrivefile.DeletedAtLTE(before),
		).
		All(ctx.InnerContext())
	if err != nil {
		return nil, err
	}
	return entToDriveFiles(entities), nil
}

func (d *driveFileRepository) BeginTransaction(ctx crud.Context) (*ent.Tx, error) {
	return d.client.Tx(ctx)
}

func (d *driveFileRepository) scanDriveFileFromRow(rows *sql.Rows, driveFile *domain.DriveFile) error {
	var (
		id                 model.Id
		etag               model.Etag
		createdAt          time.Time
		updatedAt          time.Time
		deletedAt          time.Time
		ownerRef           model.Id
		materializedPathNs sql.NullString
		name               string
		mime               string
		isFolder           bool
		size               int64
		storagePath        string
		storageKey         string
		storage            string
		visibility         string
		status             string
		parentRef          *string
	)

	err := rows.Scan(
		&id,
		&etag,
		&createdAt,
		&updatedAt,
		&deletedAt,
		&ownerRef,
		&materializedPathNs,
		&name,
		&mime,
		&isFolder,
		&size,
		&storagePath,
		&storageKey,
		&storage,
		&visibility,
		&status,
		&parentRef,
	)
	if err != nil {
		return err
	}

	driveFile.Id = &id
	driveFile.Etag = &etag
	driveFile.CreatedAt = &createdAt
	driveFile.UpdatedAt = &updatedAt
	driveFile.OwnerRef = &ownerRef
	if materializedPathNs.Valid {
		s := materializedPathNs.String
		driveFile.MaterializedPath = &s
	}
	driveFile.Name = name
	driveFile.MINE = mime
	driveFile.IsFolder = isFolder
	driveFile.Size = uint64(size)
	driveFile.StoragePath = storagePath
	driveFile.StorageKey = storageKey
	driveFile.Storage = enum.DriveFileStorageValue[storage]
	driveFile.Visibility = enum.DriveFileVisibilityValue[visibility]
	driveFile.Status = enum.DriveFileStatusValue[status]

	if driveFile.Storage == 0 {
		driveFile.Storage = enum.DriveFileStorageDefault
	}
	if driveFile.Visibility == 0 {
		driveFile.Visibility = enum.DriveFileVisibilityDefault
	}
	if driveFile.Status == 0 {
		driveFile.Status = enum.DriveFileStatusDefault
	}

	if !deletedAt.Equal(deletedAtNotDeleted) {
		driveFile.DeletedAt = &deletedAt
	}

	if parentRef != nil {
		parentId := model.Id(*parentRef)
		driveFile.ParentDriveFileRef = &parentId
	}

	return nil
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
		Field(entDrivefile.FieldOwnerRef, entity.OwnerRef).
		Field(entDrivefile.FieldParentFileRef, entity.ParentFileRef).
		Field(entDrivefile.FieldMaterializedPath, entity.MaterializedPath).
		Field(entDrivefile.FieldName, entity.Name).
		Field(entDrivefile.FieldMime, entity.Mime).
		Field(entDrivefile.FieldIsFolder, entity.IsFolder).
		Field(entDrivefile.FieldSize, entity.Size).
		Field(entDrivefile.FieldStoragePath, entity.StoragePath).
		Field(entDrivefile.FieldStorageKey, entity.StorageKey).
		Field(entDrivefile.FieldStorage, entity.Storage).
		Field(entDrivefile.FieldVisibility, entity.Visibility).
		Field(entDrivefile.FieldStatus, entity.Status)

	return builder.Descriptor()
}
