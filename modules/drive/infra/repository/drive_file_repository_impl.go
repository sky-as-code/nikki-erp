package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	entSql "entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	db "github.com/sky-as-code/nikki-erp/modules/core/database"
	"github.com/sky-as-code/nikki-erp/modules/drive/domain"
	"github.com/sky-as-code/nikki-erp/modules/drive/enum"
	"github.com/sky-as-code/nikki-erp/modules/drive/infra/ent"
	entDrivefile "github.com/sky-as-code/nikki-erp/modules/drive/infra/ent/drivefile"
	entDrivefileancestor "github.com/sky-as-code/nikki-erp/modules/drive/infra/ent/drivefileancestor"
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

func (d *driveFileRepository) driveFileAncestorClient(ctx crud.Context) *ent.DriveFileAncestorClient {
	tx, isOk := ctx.GetDbTranx().(*ent.Tx)
	if isOk {
		return tx.DriveFileAncestor
	}
	return d.client.DriveFileAncestor
}

func (d *driveFileRepository) InsertAncestors(ctx crud.Context, fileId model.Id, ancestorIds []model.Id) error {
	if len(ancestorIds) == 0 {
		return nil
	}
	client := d.driveFileAncestorClient(ctx)
	builders := make([]*ent.DriveFileAncestorCreate, 0, len(ancestorIds))
	for depth, ancestorId := range ancestorIds {
		builders = append(builders,
			client.Create().
				SetID(uuid.NewString()).
				SetFileRef(string(fileId)).
				SetAncestorRef(string(ancestorId)).
				SetDepth(len(ancestorIds) - depth),
		)
	}
	return client.CreateBulk(builders...).Exec(ctx.InnerContext())
}

func (d *driveFileRepository) DeleteAncestorsByFileIds(ctx crud.Context, fileIds []model.Id) error {
	if len(fileIds) == 0 {
		return nil
	}
	strIds := make([]string, len(fileIds))
	for i, id := range fileIds {
		strIds[i] = string(id)
	}
	_, err := d.driveFileAncestorClient(ctx).Delete().
		Where(entDrivefileancestor.FileRefIn(strIds...)).
		Exec(ctx.InnerContext())
	return err
}

func (d *driveFileRepository) GetAncestorIds(ctx crud.Context, fileId model.Id) ([]model.Id, error) {
	ancestors, err := d.driveFileAncestorClient(ctx).Query().
		Where(entDrivefileancestor.FileRef(string(fileId))).
		Order(entDrivefileancestor.ByDepth(entSql.OrderDesc())).
		All(ctx.InnerContext())
	if err != nil {
		return nil, err
	}
	ids := make([]model.Id, 0, len(ancestors))
	for _, a := range ancestors {
		ids = append(ids, model.Id(a.AncestorRef))
	}
	return ids, nil
}

func (d *driveFileRepository) GetAncestorIdsForFiles(ctx crud.Context, fileIds []model.Id) (map[model.Id][]model.Id, error) {
	if len(fileIds) == 0 {
		return nil, nil
	}
	strIds := make([]string, len(fileIds))
	for i, id := range fileIds {
		strIds[i] = string(id)
	}
	ancestors, err := d.driveFileAncestorClient(ctx).Query().
		Where(entDrivefileancestor.FileRefIn(strIds...)).
		Order(entDrivefileancestor.ByDepth(entSql.OrderDesc())).
		All(ctx.InnerContext())
	if err != nil {
		return nil, err
	}
	result := make(map[model.Id][]model.Id, len(fileIds))
	for _, a := range ancestors {
		fid := model.Id(a.FileRef)
		result[fid] = append(result[fid], model.Id(a.AncestorRef))
	}
	return result, nil
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

		shareTable := entSql.Table(entDrivefileshare.Table).As("dfs")
		ancestorTable := entSql.Table(entDrivefileancestor.Table).As("anc")

		// 1. Direct share: user has a share row pointing at this file
		directShareSub := entSql.
			Select("1").
			From(shareTable).
			Where(
				entSql.And(
					entSql.EQ(shareTable.C(entDrivefileshare.FieldUserRef), userId),
					entSql.ExprP(fmt.Sprintf("dfs.%s = %s", entDrivefileshare.FieldFileRef, fileIDCol)),
				),
			)

		// 2. Ancestor share: user has a share on one of the file's ancestors
		ancestorShareSub := entSql.
			Select("1").
			From(ancestorTable).
			Join(shareTable).
			On(
				ancestorTable.C(entDrivefileancestor.FieldAncestorRef),
				shareTable.C(entDrivefileshare.FieldFileRef),
			).
			Where(
				entSql.And(
					entSql.ExprP(fmt.Sprintf("anc.%s = %s", entDrivefileancestor.FieldFileRef, fileIDCol)),
					entSql.EQ(shareTable.C(entDrivefileshare.FieldUserRef), userId),
				),
			)

		// 3. Ancestor owner: user owns one of the file's ancestors
		ancestorFileTable := entSql.Table(entDrivefile.Table).As("af")
		ancestorOwnerSub := entSql.
			Select("1").
			From(ancestorTable).
			Join(ancestorFileTable).
			On(
				ancestorTable.C(entDrivefileancestor.FieldAncestorRef),
				ancestorFileTable.C(entDrivefile.FieldID),
			).
			Where(
				entSql.And(
					entSql.ExprP(fmt.Sprintf("anc.%s = %s", entDrivefileancestor.FieldFileRef, fileIDCol)),
					entSql.EQ(ancestorFileTable.C(entDrivefile.FieldOwnerRef), userId),
				),
			)

		s.Where(
			entSql.Or(
				entSql.EQ(s.C(entDrivefile.FieldOwnerRef), userId),
				entSql.Exists(directShareSub),
				entSql.Exists(ancestorShareSub),
				entSql.Exists(ancestorOwnerSub),
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
				name, mime, is_folder, size, storage_path, storage_key, storage, visibility, status, parent_file_ref
			FROM dri_files WHERE id = $1
			UNION ALL
			SELECT 
				f.id, f.etag, f.created_at, f.updated_at, f.deleted_at, f.owner_ref,
				f.name, f.mime, f.is_folder, f.size, f.storage_path, f.storage_key, f.storage, f.visibility, f.status, f.parent_file_ref
			FROM dri_files f 
			JOIN subtree s ON f.parent_file_ref = s.id
		)

		SELECT 
			id, etag, created_at, updated_at, deleted_at, owner_ref,
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

func (d *driveFileRepository) GetDriveFileParents(ctx crud.Context, fileId model.Id) ([]*domain.DriveFile, error) {
	dbClient := d.client.DB()
	rows, err := dbClient.QueryContext(ctx,
		`
		SELECT
			f.id, f.etag, f.created_at, f.updated_at, f.deleted_at, f.owner_ref,
			f.name, f.mime, f.is_folder, f.size, f.storage_path, f.storage_key, f.storage, f.visibility, f.status, f.parent_file_ref
		FROM dri_files f
		WHERE f.id = $1 AND f.status != $2

		UNION ALL

		SELECT
			f.id, f.etag, f.created_at, f.updated_at, f.deleted_at, f.owner_ref,
			f.name, f.mime, f.is_folder, f.size, f.storage_path, f.storage_key, f.storage, f.visibility, f.status, f.parent_file_ref
		FROM dri_file_ancestors a
		JOIN dri_files f ON f.id = a.ancestor_ref
		WHERE a.file_ref = $1 AND f.status != $2
		ORDER BY parent_file_ref NULLS LAST
		`, fileId, enum.DriveFileStatusName[enum.DriveFileStatusPendingDelete])
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
		id          model.Id
		etag        model.Etag
		createdAt   time.Time
		updatedAt   time.Time
		deletedAt   time.Time
		ownerRef    model.Id
		name        string
		mime        string
		isFolder    bool
		size        int64
		storagePath string
		storageKey  string
		storage     string
		visibility  string
		status      string
		parentRef   *string
	)

	err := rows.Scan(
		&id,
		&etag,
		&createdAt,
		&updatedAt,
		&deletedAt,
		&ownerRef,
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
