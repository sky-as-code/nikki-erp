package services

import (
	"github.com/sky-as-code/nikki-erp/common/safe"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/drive2/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/drive2/interfaces/drive_file_ancestor"
)

func NewDriveFileAncestorDomainService(
	ancestorRepo it.DriveFileAncestorRepository,
) it.DriveFileAncestorDomainService {
	return &DriveFileAncestorDomainServiceImpl{ancestorRepo: ancestorRepo}
}

type DriveFileAncestorDomainServiceImpl struct {
	ancestorRepo it.DriveFileAncestorRepository
}

func (this *DriveFileAncestorDomainServiceImpl) CreateDriveFileAncestor(
	ctx corectx.Context, cmd it.CreateDriveFileAncestorCommand,
	options ...corecrud.ServiceCreateOptions[*models.DriveFileAncestor],
) (*it.CreateDriveFileAncestorResult, error) {
	opts := safe.GetOptional(options, corecrud.ServiceCreateOptions[*models.DriveFileAncestor]{})
	return corecrud.Create(ctx, corecrud.CreateParam[models.DriveFileAncestor, *models.DriveFileAncestor]{
		Action:                 "create drive file ancestor",
		BaseRepoGetter:         this.ancestorRepo,
		Data:                   cmd,
		AfterValidationSuccess: opts.AfterValidationSuccess,
	})
}

func (this *DriveFileAncestorDomainServiceImpl) CreateBulkDriveFileAncestors(
	ctx corectx.Context, cmd it.CreateBulkDriveFileAncestorsCommand,
) (*it.CreateBulkDriveFileAncestorsResult, error) {
	if len(cmd.Items) == 0 {
		return &it.CreateBulkDriveFileAncestorsResult{}, nil
	}
	data := make([]*models.DriveFileAncestor, len(cmd.Items))
	for idx := range cmd.Items {
		data[idx] = &cmd.Items[idx]
	}
	return corecrud.CreateBulk(
		ctx,
		corecrud.CreateBulkParam[models.DriveFileAncestor, *models.DriveFileAncestor, *models.DriveFileAncestor]{
			Action:         "create bulk drive file ancestors",
			BaseRepoGetter: this.ancestorRepo,
			Data:           data,
		},
	)
}

func (this *DriveFileAncestorDomainServiceImpl) DeleteDriveFileAncestor(
	ctx corectx.Context, cmd it.DeleteDriveFileAncestorCommand, options ...corecrud.ServiceDeleteOptions,
) (*it.DeleteDriveFileAncestorResult, error) {
	opts := safe.GetOptional(options, corecrud.ServiceDeleteOptions{})
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{
		Action:                 "delete drive file ancestor",
		DbRepoGetter:           this.ancestorRepo,
		Cmd:                    dyn.DeleteOneCommand(cmd),
		AfterValidationSuccess: opts.AfterValidationSuccess,
	})
}

func (this *DriveFileAncestorDomainServiceImpl) DriveFileAncestorExists(
	ctx corectx.Context, query it.DriveFileAncestorExistsQuery,
) (*it.DriveFileAncestorExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{
		Action:       "check if drive file ancestors exist",
		DbRepoGetter: this.ancestorRepo,
		Query:        dyn.ExistsQuery(query),
	})
}

func (this *DriveFileAncestorDomainServiceImpl) GetDriveFileAncestor(
	ctx corectx.Context, query it.GetDriveFileAncestorQuery,
) (*dyn.OpResult[models.DriveFileAncestor], error) {
	return corecrud.GetOne[models.DriveFileAncestor](ctx, corecrud.GetOneParam{
		Action:       "get drive file ancestor",
		DbRepoGetter: this.ancestorRepo,
		Query:        dyn.GetOneQuery(query),
	})
}

func (this *DriveFileAncestorDomainServiceImpl) SearchDriveFileAncestors(
	ctx corectx.Context, query it.SearchDriveFileAncestorsQuery, options ...corecrud.ServiceSearchOptions,
) (*it.SearchDriveFileAncestorsResult, error) {
	opts := safe.GetOptional(options, corecrud.ServiceSearchOptions{})
	return corecrud.Search[models.DriveFileAncestor](ctx, corecrud.SearchParam{
		Action:                 "search drive file ancestors",
		DbRepoGetter:           this.ancestorRepo,
		Query:                  dyn.SearchQuery(query),
		AfterValidationSuccess: opts.AfterValidationSuccess,
	})
}

func (this *DriveFileAncestorDomainServiceImpl) UpdateDriveFileAncestor(
	ctx corectx.Context, cmd it.UpdateDriveFileAncestorCommand,
	options ...corecrud.ServiceUpdateOptions[*models.DriveFileAncestor],
) (*it.UpdateDriveFileAncestorResult, error) {
	opts := safe.GetOptional(options, corecrud.ServiceUpdateOptions[*models.DriveFileAncestor]{})
	return corecrud.Update(ctx, corecrud.UpdateParam[models.DriveFileAncestor, *models.DriveFileAncestor]{
		Action:                 "update drive file ancestor",
		DbRepoGetter:           this.ancestorRepo,
		Data:                   cmd,
		AfterValidationSuccess: opts.AfterValidationSuccess,
	})
}
