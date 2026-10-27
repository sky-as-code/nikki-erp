package services

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// The iam_user resource sees only rows of type user (see principalTypeScope); services are the
// iam_service resource's. GetEnabledUser and the permission repository read unfiltered.

func (this *UserCrudDomainServiceImpl) Search(
	ctx corectx.Context, params composable.SearchQuery, options ...corecrud.ServiceSearchOptions,
) (*composable.SearchResult, error) {
	if err := userTypeScope.narrowSearch(params); err != nil {
		return nil, err
	}
	return this.CrudDomainService.Search(ctx, params, options...)
}

func (this *UserCrudDomainServiceImpl) GetById(
	ctx corectx.Context, query composable.GetByIdQuery,
) (*composable.GetOneResult, error) {
	ok, err := userTypeScope.isOfType(ctx, this.Repository(), query)
	if err != nil || !ok {
		return &composable.GetOneResult{}, err
	}
	return this.CrudDomainService.GetById(ctx, query)
}

// GetOne may address the record by any unique key, so the type is checked on what it found.
func (this *UserCrudDomainServiceImpl) GetOne(
	ctx corectx.Context, query composable.GetOneQuery,
) (*composable.GetOneResult, error) {
	result, err := this.CrudDomainService.GetOne(ctx, query)
	if err != nil || result == nil || !result.HasData {
		return result, err
	}
	ok, err := userTypeScope.isOfType(ctx, this.Repository(), result.Data.Item)
	if err != nil || !ok {
		return &composable.GetOneResult{}, err
	}
	return result, nil
}

func (this *UserCrudDomainServiceImpl) Update(
	ctx corectx.Context, cmd composable.UpdateCommand, options ...composable.UpdateOptions,
) (*composable.MutateResult, error) {
	if refused, err := userTypeScope.assertOfType(ctx, this.Repository(), cmd); err != nil || refused != nil {
		return refused, err
	}
	return this.CrudDomainService.Update(ctx, cmd, options...)
}

func (this *UserCrudDomainServiceImpl) Delete(
	ctx corectx.Context, cmd composable.DeleteCommand, options ...composable.DeleteOptions,
) (*composable.MutateResult, error) {
	if refused, err := userTypeScope.assertOfType(ctx, this.Repository(), cmd); err != nil || refused != nil {
		return refused, err
	}
	return this.CrudDomainService.Delete(ctx, cmd, options...)
}

func (this *UserCrudDomainServiceImpl) SetArchived(
	ctx corectx.Context, cmd composable.SetArchivedCommand,
) (*composable.MutateResult, error) {
	if refused, err := userTypeScope.assertOfType(ctx, this.Repository(), cmd); err != nil || refused != nil {
		return refused, err
	}
	return this.CrudDomainService.SetArchived(ctx, cmd)
}
