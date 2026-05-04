package repository

import (
	"database/sql"
	"fmt"
	"time"

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
	"github.com/sky-as-code/nikki-erp/modules/drive2/domain"
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
			Schema:       dmodel.MustGetSchema(domain.DriveFileSchemaName),
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
	ctx corectx.Context, keys domain.DriveFile,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	return baserepo.DeleteOne(ctx, this.dynamicRepo, keys.GetFieldData())
}

func (this *DriveFileDynamicRepository) Exists(
	ctx corectx.Context, keys []domain.DriveFile,
) (*dyn.OpResult[dyn.RepoExistsResult], error) {
	dynamicKeys := array.Map(keys, func(key domain.DriveFile) dmodel.DynamicFields {
		return key.GetFieldData()
	})
	return baserepo.Exists(ctx, this.dynamicRepo, dynamicKeys)
}

func (this *DriveFileDynamicRepository) Insert(
	ctx corectx.Context, driveFile domain.DriveFile,
) (*dyn.OpResult[int], error) {
	return baserepo.Insert(ctx, this.dynamicRepo, driveFile)
}

func (this *DriveFileDynamicRepository) GetOne(
	ctx corectx.Context, param dyn.RepoGetOneParam,
) (*dyn.OpResult[domain.DriveFile], error) {
	return baserepo.GetOne[domain.DriveFile](ctx, this.dynamicRepo, param)
}

func (this *DriveFileDynamicRepository) Search(
	ctx corectx.Context, param dyn.RepoSearchParam,
) (*dyn.OpResult[dyn.PagedResultData[domain.DriveFile]], error) {
	return baserepo.Search[domain.DriveFile](ctx, this.dynamicRepo, param)
}

func (this *DriveFileDynamicRepository) Update(
	ctx corectx.Context, driveFile domain.DriveFile,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	return baserepo.Update(ctx, this.dynamicRepo, driveFile.GetFieldData())
}

func (this *DriveFileDynamicRepository) FindByIds(ctx corectx.Context, ids []model.Id) ([]domain.DriveFile, error) {
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
) (*dyn.OpResult[dyn.PagedResultData[domain.DriveFile]], error) {
	panic("drive_file_dynamic_repository: SearchAccessible unimplemented")
}

func (this *DriveFileDynamicRepository) GetRootFileByUser(
	ctx corectx.Context, userId model.Id, param dyn.RepoSearchParam,
) (*dyn.OpResult[dyn.PagedResultData[domain.DriveFile]], error) {
	panic("drive_file_dynamic_repository: GetRootFileByUser unimplemented")
}

func (this *DriveFileDynamicRepository) SearchByParent(
	ctx corectx.Context, param it.DriveFileSearchByParentParam,
) (*dyn.OpResult[dyn.PagedResultData[domain.DriveFile]], error) {
	panic("drive_file_dynamic_repository: SearchByParent unimplemented")
}

func (this *DriveFileDynamicRepository) GetDriveFilesSharedByUser(
	ctx corectx.Context, userId model.Id, param dyn.RepoSearchParam,
) (*dyn.OpResult[dyn.PagedResultData[domain.DriveFile]], error) {
	panic("drive_file_dynamic_repository: GetDriveFilesSharedByUser unimplemented")
}

func (this *DriveFileDynamicRepository) GetDriveFileChildren(
	ctx corectx.Context, parentId model.Id,
) ([]domain.DriveFile, error) {
	// query := (`
	// 	WITH RECURSIVE subtree AS (
	// 		SELECT
	// 			id, etag, created_at, updated_at, deleted_at, owner_ref,
	// 			name, mime, is_folder, size, storage_path, storage_key, storage, visibility, status, parent_file_ref
	// 		FROM dri_files WHERE id = $1
	// 		UNION ALL
	// 		SELECT
	// 			f.id, f.etag, f.created_at, f.updated_at, f.deleted_at, f.owner_ref,
	// 			f.name, f.mime, f.is_folder, f.size, f.storage_path, f.storage_key, f.storage, f.visibility, f.status, f.parent_file_ref
	// 		FROM dri_files f
	// 		JOIN subtree s ON f.parent_file_ref = s.id
	// 	)
	//
	// 	SELECT
	// 		id, etag, created_at, updated_at, deleted_at, owner_ref,
	// 		name, mime, is_folder, size, storage_path, storage_key, storage, visibility, status, parent_file_ref
	// 	FROM subtree
	// 	WHERE status != $2 and id != $1
	// 	`, parentId, enum.DriveFileStatusName[enum.DriveFileStatusPendingDelete])
	panic("drive_file_dynamic_repository: GetDriveFileChildren unimplemented")
}

func (this *DriveFileDynamicRepository) GetDriveFileParents(
	ctx corectx.Context, driveFileId model.Id,
) ([]*domain.DriveFile, error) {
	client := this.dynamicRepo.ExtractClient(ctx)
	queryCols := lo.Map(domain.DriveFileAllFields, func(col string, _ int) any {
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
	rows, err := client.Query(ctx, query, driveFileId, domain.DriveFileStatusPendingDelete)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var driveFiles []*domain.DriveFile
	for rows.Next() {
		driveFile := &domain.DriveFile{}
		this.scanDriveFileFromRow(rows, driveFile)
		driveFiles = append(driveFiles, driveFile)
	}

	return driveFiles, nil
}

func (this *DriveFileDynamicRepository) GetExpiredTrashedDriveFiles(
	ctx corectx.Context, before time.Time,
) ([]domain.DriveFile, error) {
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
	ctx corectx.Context, driveFile domain.DriveFile, prevEtag model.Etag,
) (*dyn.OpResult[domain.DriveFile], error) {
	panic("drive_file_dynamic_repository: Overwrite unimplemented")
}

func (this *DriveFileDynamicRepository) DeleteByIds(ctx corectx.Context, ids []model.Id) (int, error) {
	panic("drive_file_dynamic_repository: DeleteByIds unimplemented")
}

func (this *DriveFileDynamicRepository) scanDriveFileFromRow(rows *sql.Rows, driveFile *domain.DriveFile) error {
	var driveFileFieldData = dmodel.DynamicFields{}

	for _, field := range domain.DriveFileAllFields {
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
