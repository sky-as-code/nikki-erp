package services

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// The iam_user resource sees only rows of type user (see principalTypeScope); services are the
// iam_service resource's. GetEnabledUser and the permission repository read unfiltered.

func (this *UserCrudDomainServiceImpl) Search(
	ctx corectx.Context, params dynamicresource.SearchQuery, options ...corecrud.ServiceSearchOptions,
) (*dynamicresource.SearchResult, error) {
	if err := userTypeScope.narrowSearch(params); err != nil {
		return nil, err
	}
	return this.CrudDomainService.Search(ctx, params, options...)
}

func (this *UserCrudDomainServiceImpl) GetById(
	ctx corectx.Context, query dynamicresource.GetByIdQuery,
) (*dynamicresource.GetOneResult, error) {
	ok, err := userTypeScope.isOfType(ctx, this.Repository(), query)
	if err != nil || !ok {
		return &dynamicresource.GetOneResult{}, err
	}
	return this.CrudDomainService.GetById(ctx, query)
}

// GetOne may address the record by any unique key, so the type is checked on what it found.
func (this *UserCrudDomainServiceImpl) GetOne(
	ctx corectx.Context, query dynamicresource.GetOneQuery,
) (*dynamicresource.GetOneResult, error) {
	result, err := this.CrudDomainService.GetOne(ctx, query)
	if err != nil || result == nil || !result.HasData {
		return result, err
	}
	ok, err := userTypeScope.isOfType(ctx, this.Repository(), result.Data.Item)
	if err != nil || !ok {
		return &dynamicresource.GetOneResult{}, err
	}
	return result, nil
}

func (this *UserCrudDomainServiceImpl) Update(
	ctx corectx.Context, cmd dynamicresource.UpdateCommand, options ...dynamicresource.UpdateOptions,
) (*dynamicresource.MutateResult, error) {
	if refused, err := userTypeScope.assertOfType(ctx, this.Repository(), cmd); err != nil || refused != nil {
		return refused, err
	}
	return this.CrudDomainService.Update(ctx, cmd, options...)
}

func (this *UserCrudDomainServiceImpl) Delete(
	ctx corectx.Context, cmd dynamicresource.DeleteCommand, options ...dynamicresource.DeleteOptions,
) (*dynamicresource.MutateResult, error) {
	if refused, err := userTypeScope.assertOfType(ctx, this.Repository(), cmd); err != nil || refused != nil {
		return refused, err
	}
	return this.CrudDomainService.Delete(ctx, cmd, options...)
}

func (this *UserCrudDomainServiceImpl) SetArchived(
	ctx corectx.Context, cmd dynamicresource.SetArchivedCommand,
) (*dynamicresource.MutateResult, error) {
	if refused, err := userTypeScope.assertOfType(ctx, this.Repository(), cmd); err != nil || refused != nil {
		return refused, err
	}
	return this.CrudDomainService.SetArchived(ctx, cmd)
}
