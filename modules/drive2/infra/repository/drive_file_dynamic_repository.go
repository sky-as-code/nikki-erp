package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"

	"go.uber.org/dig"

	"github.com/samber/lo"
	"github.com/sky-as-code/nikki-erp/common/array"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	dyorm "github.com/sky-as-code/nikki-erp/common/dynamicmodel/orm"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/database"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/baserepo"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"github.com/sky-as-code/nikki-erp/modules/drive2/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/drive2/interfaces/drive_file"
)

type DriveFileDynamicRepositoryParam struct {
	dig.In

	Client        dyorm.DbClient
	ConfigSvc     config.ConfigService
	QueryBuilder  dyorm.QueryBuilder
	Logger        logging.LoggerService
	NewBaseRepoFn dyn.NewBaseDynamicRepositoryFn
}

func NewDriveFileDynamicRepository(param DriveFileDynamicRepositoryParam) it.DriveFileRepository {
	dynamicRepo := param.NewBaseRepoFn(
		dyn.NewBaseRepoParam{
			Client:       param.Client,
			ConfigSvc:    param.ConfigSvc,
			QueryBuilder: param.QueryBuilder,
			Logger:       param.Logger,
			Schema:       dmodel.MustGetSchema(models.DriveFileSchemaName),
		},
	)
	return &DriveFileDynamicRepository{dynamicRepo: dynamicRepo}
}

type DriveFileDynamicRepository struct {
	dynamicRepo dyn.BaseDynamicRepository
}

func (this *DriveFileDynamicRepository) GetBaseRepo() dyn.BaseDynamicRepository {
	return this.dynamicRepo
}

func (this *DriveFileDynamicRepository) BeginTransaction(ctx corectx.Context) (database.DbTransaction, error) {
	return this.dynamicRepo.BeginTransaction(ctx)
}

func (this *DriveFileDynamicRepository) DeleteOne(
	ctx corectx.Context, keys models.DriveFile,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	return baserepo.DeleteOne(ctx, this.dynamicRepo, keys.GetFieldData())
}

func (this *DriveFileDynamicRepository) Exists(
	ctx corectx.Context, keys []models.DriveFile,
) (*dyn.OpResult[dyn.RepoExistsResult], error) {
	dynamicKeys := array.Map(keys, func(key models.DriveFile) dmodel.DynamicFields {
		return key.GetFieldData()
	})
	return baserepo.Exists(ctx, this.dynamicRepo, dynamicKeys)
}

func (this *DriveFileDynamicRepository) Insert(
	ctx corectx.Context, driveFile models.DriveFile,
) (*dyn.OpResult[int], error) {
	return baserepo.Insert(ctx, this.dynamicRepo, driveFile)
}

func (this *DriveFileDynamicRepository) GetOne(
	ctx corectx.Context, param dyn.RepoGetOneParam,
) (*dyn.OpResult[models.DriveFile], error) {
	return baserepo.GetOne[models.DriveFile](ctx, this.dynamicRepo, param)
}

func (this *DriveFileDynamicRepository) Search(
	ctx corectx.Context, param dyn.RepoSearchParam,
) (*dyn.OpResult[dyn.PagedResultData[models.DriveFile]], error) {
	return baserepo.Search[models.DriveFile](ctx, this.dynamicRepo, param)
}

func (this *DriveFileDynamicRepository) Update(
	ctx corectx.Context, driveFile models.DriveFile,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	return baserepo.Update(ctx, this.dynamicRepo, driveFile.GetFieldData())
}

func (this *DriveFileDynamicRepository) FindByIds(ctx corectx.Context, ids []model.Id) ([]models.DriveFile, error) {
	panic("drive_file_dynamic_repository: FindByIds unimplemented")
}

func (this *DriveFileDynamicRepository) ExistsByOwnerParentNameFolder(
	ctx corectx.Context, ownerRef model.Id, parentRef *model.Id, name string, isFolder bool,
) (bool, error) {
	panic("drive_file_dynamic_repository: ExistsByOwnerParentNameFolder unimplemented")
}

func (this *DriveFileDynamicRepository) ParseSearchGraph(criteria *string) (*dmodel.SearchGraph, ft.ValidationErrors) {
	panic("drive_file_dynamic_repository: ParseSearchGraph unimplemented")
}

func (this *DriveFileDynamicRepository) SearchAccessible(
	ctx corectx.Context, userId model.Id, param dyn.RepoSearchParam,
) (*dyn.OpResult[dyn.PagedResultData[models.DriveFile]], error) {
	panic("drive_file_dynamic_repository: SearchAccessible unimplemented")
}

func (this *DriveFileDynamicRepository) GetRootFileByUser(
	ctx corectx.Context, userId model.Id, param dyn.RepoSearchParam,
) (*dyn.OpResult[dyn.PagedResultData[models.DriveFile]], error) {
	panic("drive_file_dynamic_repository: GetRootFileByUser unimplemented")
}

func (this *DriveFileDynamicRepository) SearchByParent(
	ctx corectx.Context, param it.DriveFileSearchByParentParam,
) (*dyn.OpResult[dyn.PagedResultData[models.DriveFile]], error) {
	panic("drive_file_dynamic_repository: SearchByParent unimplemented")
}

func (this *DriveFileDynamicRepository) GetDriveFilesSharedByUser(
	ctx corectx.Context, userId model.Id, param dyn.RepoSearchParam,
) (*dyn.OpResult[dyn.PagedResultData[models.DriveFile]], error) {
	panic("drive_file_dynamic_repository: GetDriveFilesSharedByUser unimplemented")
}

func (this *DriveFileDynamicRepository) GetDriveFileChildren(
	ctx corectx.Context, query it.GetDriveFileChildrenQuery,
) ([]*models.DriveFile, error) {
	client := this.dynamicRepo.ExtractClient(ctx)
	sqlText := `
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
		WHERE status != $2 AND id != $1`
	args := []any{query.DriveFileId, string(models.DriveFileStatusPendingDelete)}
	if query.Size > 0 {
		sqlText += fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
		args = append(args, query.Size, query.Page*query.Size)
	}

	rows, err := client.Query(ctx, sqlText, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	driveFiles := make([]*models.DriveFile, 0)
	for rows.Next() {
		var (
			id          model.Id
			etag        model.Etag
			createdAt   time.Time
			updatedAt   sql.NullTime
			deletedAt   sql.NullTime
			ownerRef    model.Id
			name        string
			mime        sql.NullString
			isFolder    bool
			size        int64
			storagePath sql.NullString
			storageKey  sql.NullString
			storage     sql.NullString
			visibility  string
			status      string
			parentRef   sql.NullString
		)
		if err := rows.Scan(
			&id, &etag, &createdAt, &updatedAt, &deletedAt, &ownerRef,
			&name, &mime, &isFolder, &size, &storagePath, &storageKey, &storage, &visibility, &status, &parentRef,
		); err != nil {
			return nil, err
		}

		fieldData := dmodel.DynamicFields{
			basemodel.FieldId:             id,
			basemodel.FieldEtag:           etag,
			basemodel.FieldCreatedAt:      model.WrapModelDateTime(createdAt),
			models.DriveFileFieldOwnerRef: ownerRef,
			models.DriveFileFieldName:     name,
			models.DriveFileFieldIsFolder: isFolder,
			models.DriveFileFieldSize:     size,
			models.DriveFileFieldVisibility: visibility,
			models.DriveFileFieldStatus:   status,
		}
		if mime.Valid {
			fieldData[models.DriveFileFieldMime] = mime.String
		}
		if storagePath.Valid {
			fieldData[models.DriveFileFieldStoragePath] = storagePath.String
		}
		if storageKey.Valid {
			fieldData[models.DriveFileFieldStorageKey] = storageKey.String
		}
		if storage.Valid {
			fieldData[models.DriveFileFieldStorage] = storage.String
		}
		if updatedAt.Valid {
			fieldData[basemodel.FieldUpdatedAt] = model.WrapModelDateTime(updatedAt.Time)
		}
		if deletedAt.Valid {
			fieldData[models.DriveFileFieldDeletedAt] = model.WrapModelDateTime(deletedAt.Time)
		}
		if parentRef.Valid {
			parentId := model.Id(parentRef.String)
			fieldData[models.DriveFileFieldParentFileRef] = parentId
		}

		driveFile := models.NewDriveFileFrom(fieldData)
		driveFiles = append(driveFiles, driveFile)
	}

	return driveFiles, rows.Err()
}

func (this *DriveFileDynamicRepository) CountDriveFileChildren(
	ctx corectx.Context, query it.GetDriveFileChildrenQuery,
) (int, error) {
	client := this.dynamicRepo.ExtractClient(ctx)
	rows, err := client.Query(ctx, `
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
		SELECT COUNT(*)
		FROM subtree
		WHERE status != $2 AND id != $1`,
		query.DriveFileId, string(models.DriveFileStatusPendingDelete),
	)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var total int
	if rows.Next() {
		if err := rows.Scan(&total); err != nil {
			return 0, err
		}
	}

	return total, rows.Err()
}

func (this *DriveFileDynamicRepository) GetDriveFileParents(
	ctx corectx.Context, driveFileId model.Id,
) ([]*models.DriveFile, error) {
	client := this.dynamicRepo.ExtractClient(ctx)
	queryCols := lo.Map(models.DriveFileAllFields, func(col string, _ int) any {
		return "f." + col
	})

	queryCols = append(queryCols, queryCols...)

	query := fmt.Sprintf(`
	SELECT %s, %s, %s, f.%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s 
		FROM dri_files f
		WHERE f.id = $1 AND f.status != $2

		UNION ALL

	SELECT %s, %s, %s, f.%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s 
		FROM dri_file_ancestors a
		JOIN dri_files f ON f.id = a.ancestor_ref
		WHERE a.file_ref = $1 AND f.status != $2
		ORDER BY parent_file_ref NULLS LAST`, queryCols...)
	rows, err := client.Query(ctx, query, driveFileId, models.DriveFileStatusPendingDelete)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var driveFiles []*models.DriveFile
	for rows.Next() {
		driveFile := &models.DriveFile{}
		this.scanDriveFileFromRow(rows, driveFile)
		driveFiles = append(driveFiles, driveFile)
	}

	return driveFiles, nil
}

func (this *DriveFileDynamicRepository) GetExpiredTrashedDriveFiles(
	ctx corectx.Context, before time.Time,
) ([]models.DriveFile, error) {
	panic("drive_file_dynamic_repository: GetExpiredTrashedDriveFiles unimplemented")
}

func (this *DriveFileDynamicRepository) InsertAncestors(
	ctx corectx.Context, fileId model.Id, ancestorIds []model.Id,
) error {
	panic("drive_file_dynamic_repository: InsertAncestors unimplemented")
}

func (this *DriveFileDynamicRepository) DeleteAncestorsByFileIds(ctx corectx.Context, fileIds []model.Id) error {
	panic("drive_file_dynamic_repository: DeleteAncestorsByFileIds unimplemented")
}

func (this *DriveFileDynamicRepository) GetAncestorIds(ctx corectx.Context, fileId model.Id) ([]model.Id, error) {
	panic("drive_file_dynamic_repository: GetAncestorIds unimplemented")
}

func (this *DriveFileDynamicRepository) GetAncestorIdsForFiles(
	ctx corectx.Context, fileIds []model.Id,
) (map[model.Id][]model.Id, error) {
	panic("drive_file_dynamic_repository: GetAncestorIdsForFiles unimplemented")
}

func (this *DriveFileDynamicRepository) Overwrite(
	ctx corectx.Context, driveFile models.DriveFile, prevEtag model.Etag,
) (*dyn.OpResult[models.DriveFile], error) {
	panic("drive_file_dynamic_repository: Overwrite unimplemented")
}

func (this *DriveFileDynamicRepository) DeleteByIds(ctx corectx.Context, ids []model.Id) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}

	table := this.dynamicRepo.Schema().TableName()
	placeholders := make([]string, len(ids))
	args := make([]any, len(ids))
	for idx, id := range ids {
		placeholders[idx] = fmt.Sprintf("$%d", idx+1)
		args[idx] = string(id)
	}

	q := fmt.Sprintf(
		`DELETE FROM %s WHERE %s IN (%s)`,
		table,
		basemodel.FieldId,
		strings.Join(placeholders, ", "),
	)

	cli := this.dynamicRepo.ExtractClient(ctx)
	res, err := cli.Exec(ctx, q, args...)
	if err != nil {
		return 0, err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	return int(affected), nil
}

func (this *DriveFileDynamicRepository) scanDriveFileFromRow(rows *sql.Rows, driveFile *models.DriveFile) error {
	var driveFileFieldData = dmodel.DynamicFields{}

	for _, field := range models.DriveFileAllFields {
		var val any
		err := rows.Scan(&val)
		if err != nil {
			return err
		}

		driveFileFieldData[field] = val
	}

	driveFile.SetFieldData(driveFileFieldData)

	return nil
}
