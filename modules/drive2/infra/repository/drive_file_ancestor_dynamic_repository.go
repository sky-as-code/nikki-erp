package repository

import (
	"fmt"

	"go.uber.org/dig"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	dyorm "github.com/sky-as-code/nikki-erp/common/dynamicmodel/orm"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/database"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/baserepo"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"github.com/sky-as-code/nikki-erp/modules/drive2/domain"
	it "github.com/sky-as-code/nikki-erp/modules/drive2/interfaces/drive_file_ancestor"
)

const insertBulkChunkSize = 200

type DriveFileAncestorDynamicRepositoryParam struct {
	dig.In

	Client        dyorm.DbClient
	ConfigSvc     config.ConfigService
	QueryBuilder  dyorm.QueryBuilder
	Logger        logging.LoggerService
	NewBaseRepoFn dyn.NewBaseDynamicRepositoryFn
}

func NewDriveFileAncestorDynamicRepository(param DriveFileAncestorDynamicRepositoryParam) it.DriveFileAncestorRepository {
	dynamicRepo := param.NewBaseRepoFn(
		dyn.NewBaseRepoParam{
			Client:       param.Client,
			ConfigSvc:    param.ConfigSvc,
			QueryBuilder: param.QueryBuilder,
			Logger:       param.Logger,
			Schema:       dmodel.MustGetSchema(domain.DriveFileAncestorSchemaName),
		},
	)
	return &DriveFileAncestorDynamicRepository{dynamicRepo: dynamicRepo}
}

type DriveFileAncestorDynamicRepository struct {
	dynamicRepo dyn.BaseDynamicRepository
}

func (this *DriveFileAncestorDynamicRepository) GetBaseRepo() dyn.BaseDynamicRepository {
	return this.dynamicRepo
}

func (this *DriveFileAncestorDynamicRepository) BeginTransaction(ctx corectx.Context) (database.DbTransaction, error) {
	return this.dynamicRepo.BeginTransaction(ctx)
}

func (this *DriveFileAncestorDynamicRepository) Insert(
	ctx corectx.Context, row domain.DriveFileAncestor,
) (*dyn.OpResult[int], error) {
	return baserepo.Insert(ctx, this.dynamicRepo, row)
}

func (this *DriveFileAncestorDynamicRepository) InsertBulk(
	ctx corectx.Context, rows []domain.DriveFileAncestor,
) (*dyn.OpResult[int], error) {
	return baserepo.InsertBulk(ctx, this.dynamicRepo, rows)
}

func (this *DriveFileAncestorDynamicRepository) Update(
	ctx corectx.Context, row domain.DriveFileAncestor,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	return baserepo.Update(ctx, this.dynamicRepo, row.GetFieldData())
}

func (this *DriveFileAncestorDynamicRepository) GetOne(
	ctx corectx.Context, param dyn.RepoGetOneParam,
) (*dyn.OpResult[domain.DriveFileAncestor], error) {
	return baserepo.GetOne[domain.DriveFileAncestor](ctx, this.dynamicRepo, param)
}

func (this *DriveFileAncestorDynamicRepository) Search(
	ctx corectx.Context, param dyn.RepoSearchParam,
) (*dyn.OpResult[dyn.PagedResultData[domain.DriveFileAncestor]], error) {
	return baserepo.Search[domain.DriveFileAncestor](ctx, this.dynamicRepo, param)
}

func (this *DriveFileAncestorDynamicRepository) DeleteOne(
	ctx corectx.Context, keys domain.DriveFileAncestor,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	return baserepo.DeleteOne(ctx, this.dynamicRepo, keys.GetFieldData())
}

func (this *DriveFileAncestorDynamicRepository) DeleteByFileRefs(
	ctx corectx.Context, fileRefs []model.Id,
) (deleted int, err error) {
	if len(fileRefs) == 0 {
		return 0, nil
	}
	table := this.dynamicRepo.Schema().TableName()
	col := domain.DriveFileAncestorFieldFileRef
	q := fmt.Sprintf(`DELETE FROM %s WHERE %s = $1`, table, col)
	cli := this.dynamicRepo.ExtractClient(ctx)
	for _, fid := range fileRefs {
		res, execErr := cli.Exec(ctx, q, string(fid))
		if execErr != nil {
			return deleted, execErr
		}
		n, rowsAffErr := res.RowsAffected()
		if rowsAffErr != nil {
			return deleted, rowsAffErr
		}
		deleted += int(n)
	}
	return deleted, nil
}