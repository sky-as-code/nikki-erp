package repository

import (
	"go.uber.org/dig"

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
	it "github.com/sky-as-code/nikki-erp/modules/drive2/interfaces/drive_file_share"
)

type DriveFileShareDynamicRepositoryParam struct {
	dig.In

	Client        dyorm.DbClient
	ConfigSvc     config.ConfigService
	QueryBuilder  dyorm.QueryBuilder
	Logger        logging.LoggerService
	NewBaseRepoFn dyn.NewBaseDynamicRepositoryFn
}

func NewDriveFileShareDynamicRepository(param DriveFileShareDynamicRepositoryParam) it.DriveFileShareRepository {
	dynamicRepo := param.NewBaseRepoFn(
		dyn.NewBaseRepoParam{
			Client:       param.Client,
			ConfigSvc:    param.ConfigSvc,
			QueryBuilder: param.QueryBuilder,
			Logger:       param.Logger,
			Schema:       dmodel.MustGetSchema(models.DriveFileShareSchemaName),
		},
	)
	return &DriveFileShareDynamicRepository{dynamicRepo: dynamicRepo}
}

type DriveFileShareDynamicRepository struct {
	dynamicRepo dyn.BaseDynamicRepository
}

func (this *DriveFileShareDynamicRepository) GetBaseRepo() dyn.BaseDynamicRepository {
	return this.dynamicRepo
}

func (this *DriveFileShareDynamicRepository) BeginTransaction(ctx corectx.Context) (database.DbTransaction, error) {
	return this.dynamicRepo.BeginTransaction(ctx)
}

func (this *DriveFileShareDynamicRepository) DeleteOne(
	ctx corectx.Context, keys models.DriveFileShare,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	return baserepo.DeleteOne(ctx, this.dynamicRepo, keys.GetFieldData())
}

func (this *DriveFileShareDynamicRepository) Exists(
	ctx corectx.Context, keys []models.DriveFileShare,
) (*dyn.OpResult[dyn.RepoExistsResult], error) {
	dynamicKeys := array.Map(keys, func(key models.DriveFileShare) dmodel.DynamicFields {
		return key.GetFieldData()
	})
	return baserepo.Exists(ctx, this.dynamicRepo, dynamicKeys)
}

func (this *DriveFileShareDynamicRepository) Insert(
	ctx corectx.Context, share models.DriveFileShare,
) (*dyn.OpResult[int], error) {
	return baserepo.Insert(ctx, this.dynamicRepo, share)
}

func (this *DriveFileShareDynamicRepository) GetOne(
	ctx corectx.Context, param dyn.RepoGetOneParam,
) (*dyn.OpResult[models.DriveFileShare], error) {
	return baserepo.GetOne[models.DriveFileShare](ctx, this.dynamicRepo, param)
}

func (this *DriveFileShareDynamicRepository) Search(
	ctx corectx.Context, param dyn.RepoSearchParam,
) (*dyn.OpResult[dyn.PagedResultData[models.DriveFileShare]], error) {
	return baserepo.Search[models.DriveFileShare](ctx, this.dynamicRepo, param)
}

func (this *DriveFileShareDynamicRepository) Update(
	ctx corectx.Context, share models.DriveFileShare,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	return baserepo.Update(ctx, this.dynamicRepo, share.GetFieldData())
}

func (this *DriveFileShareDynamicRepository) ListByFileRef(
	ctx corectx.Context, param it.ListDriveFileShareByFileRefParam,
) (*dyn.OpResult[dyn.PagedResultData[models.DriveFileShare]], error) {
	panic("drive_file_share_dynamic_repository: ListByFileRef unimplemented")
}

func (this *DriveFileShareDynamicRepository) ListResolvedByFileRefs(
	ctx corectx.Context, fileRef model.Id, refs []model.Id, excludedUserRefs []model.Id, page int, size int,
) (*dyn.OpResult[dyn.PagedResultData[models.DriveFileShare]], error) {
	panic("drive_file_share_dynamic_repository: ListResolvedByFileRefs unimplemented")
}

func (this *DriveFileShareDynamicRepository) ListByFileRefsAndUserRef(
	ctx corectx.Context, driveFileIds []model.Id, userId model.Id,
) ([]models.DriveFileShare, error) {
	panic("drive_file_share_dynamic_repository: ListByFileRefsAndUserRef unimplemented")
}

func (this *DriveFileShareDynamicRepository) ListByUserRef(
	ctx corectx.Context, userRef model.Id,
) ([]models.DriveFileShare, error) {
	panic("drive_file_share_dynamic_repository: ListByUserRef unimplemented")
}

func (this *DriveFileShareDynamicRepository) ParseSearchGraph(criteria *string) (*dmodel.SearchGraph, ft.ValidationErrors) {
	panic("drive_file_share_dynamic_repository: ParseSearchGraph unimplemented")
}
