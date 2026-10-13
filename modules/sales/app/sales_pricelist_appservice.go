package app

import (
	"github.com/samber/lo"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itCatalog "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/catalog"
	itEvent "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/event"
)

func NewSalesPricelistApplicationService(
	base composable.CrudApplicationService, pricingChangedPub itEvent.PricingChangedEventPublisher,
) itCatalog.SalesPricelistApplicationService {
	return &SalesPricelistApplicationServiceImpl{
		CrudApplicationService: base,
		pricelistSvc:           base.DomainService().(itCatalog.SalesPricelistDomainService),
		pricingChangedPub:      pricingChangedPub,
	}
}

type SalesPricelistApplicationServiceImpl struct {
	composable.CrudApplicationService
	pricelistSvc      itCatalog.SalesPricelistDomainService
	pricingChangedPub itEvent.PricingChangedEventPublisher
}

// SetDefault reuses the update permission: naming the fallback list is an ordinary edit of the
// list's own flag, and a separate grant would have to be discovered and assigned to restore a
// power the caller already has over the record.
func (this *SalesPricelistApplicationServiceImpl) SetDefault(
	ctx corectx.Context, cmd itCatalog.SetDefaultPricelistCommand,
) (*itCatalog.SetDefaultPricelistResult, error) {
	if cErrs, err := assertRecordAction(this, ctx, PermissionSetDefaultPricelist, cmd); cErrs != nil || err != nil {
		return mutateFailure(cErrs, err)
	}
	res, err := this.pricelistSvc.SetDefault(ctx, readStringParam(cmd, paramRecordId))
	if err == nil && res.ClientErrors.Count() == 0 {
		this.pricingChangedPub.PublishAsync(ctx, itEvent.PricingChangedPricelist,
			readStringParam(cmd, basemodel.FieldOrgId), readStringParam(cmd, paramRecordId))
	}

	return res, err
}

func NewSalesPricelistItemApplicationService(
	base composable.CrudApplicationService, pricingChangedPub itEvent.PricingChangedEventPublisher,
) itCatalog.SalesPricelistItemApplicationService {
	return &SalesPricelistItemApplicationServiceImpl{CrudApplicationService: base, pricingChangedPub: pricingChangedPub}
}

type SalesPricelistItemApplicationServiceImpl struct {
	composable.CrudApplicationService

	pricingChangedPub itEvent.PricingChangedEventPublisher
}

// The plain-CRUD catalogue resources: no custom action, but each still gets its own application
// service so a rule has a home later and the layer is published by its own type.

func NewSalesComboApplicationService(
	base composable.CrudApplicationService, pricingChangedPub itEvent.PricingChangedEventPublisher,
) itCatalog.SalesComboApplicationService {
	return &SalesComboApplicationServiceImpl{CrudApplicationService: base, pricingChangedPub: pricingChangedPub}
}

type SalesComboApplicationServiceImpl struct {
	composable.CrudApplicationService

	pricingChangedPub itEvent.PricingChangedEventPublisher
}

func NewSalesComboComponentApplicationService(
	base composable.CrudApplicationService, pricingChangedPub itEvent.PricingChangedEventPublisher,
) itCatalog.SalesComboComponentApplicationService {
	return &SalesComboComponentApplicationServiceImpl{CrudApplicationService: base, pricingChangedPub: pricingChangedPub}
}

type SalesComboComponentApplicationServiceImpl struct {
	composable.CrudApplicationService

	pricingChangedPub itEvent.PricingChangedEventPublisher
}

func (this *SalesPricelistApplicationServiceImpl) Create(
	ctx corectx.Context, cmd composable.CreateCommand,
) (*composable.CreateResult, error) {
	res, err := this.CrudApplicationService.Create(ctx, cmd)
	if err == nil && res.HasData && res.ClientErrors.Count() == 0 {
		this.pricingChangedPub.PublishAsync(ctx, itEvent.PricingChangedPricelist, lo.FromPtr(res.Data.GetString(basemodel.FieldOrgId)), lo.FromPtr(res.Data.GetString(basemodel.FieldId)))
	}

	return res, err
}

func (this *SalesPricelistApplicationServiceImpl) Update(
	ctx corectx.Context, cmd composable.UpdateCommand,
) (*composable.MutateResult, error) {
	res, err := this.CrudApplicationService.Update(ctx, cmd)
	if err == nil && res.ClientErrors.Count() == 0 {
		this.pricingChangedPub.PublishAsync(ctx, itEvent.PricingChangedPricelist, lo.FromPtr(cmd.GetString(basemodel.FieldOrgId)), lo.FromPtr(cmd.GetString(basemodel.FieldId)))
	}

	return res, err
}

func (this *SalesPricelistApplicationServiceImpl) Delete(
	ctx corectx.Context, cmd composable.DeleteCommand,
) (*composable.MutateResult, error) {
	res, err := this.CrudApplicationService.Delete(ctx, cmd)
	if err == nil && res.ClientErrors.Count() == 0 {
		this.pricingChangedPub.PublishAsync(ctx, itEvent.PricingChangedPricelist, lo.FromPtr(cmd.GetString(basemodel.FieldOrgId)), lo.FromPtr(cmd.GetString(basemodel.FieldId)))
	}

	return res, err
}

func (this *SalesPricelistApplicationServiceImpl) SetArchived(
	ctx corectx.Context, cmd composable.SetArchivedCommand,
) (*composable.MutateResult, error) {
	res, err := this.CrudApplicationService.SetArchived(ctx, cmd)
	if err == nil && res.ClientErrors.Count() == 0 {
		this.pricingChangedPub.PublishAsync(ctx, itEvent.PricingChangedPricelist, lo.FromPtr(cmd.GetString(basemodel.FieldOrgId)), lo.FromPtr(cmd.GetString(basemodel.FieldId)))
	}

	return res, err
}

func (this *SalesPricelistItemApplicationServiceImpl) Create(
	ctx corectx.Context, cmd composable.CreateCommand,
) (*composable.CreateResult, error) {
	res, err := this.CrudApplicationService.Create(ctx, cmd)
	if err == nil && res.HasData && res.ClientErrors.Count() == 0 {
		this.pricingChangedPub.PublishAsync(ctx, itEvent.PricingChangedPricelistItem, lo.FromPtr(res.Data.GetString(basemodel.FieldOrgId)), lo.FromPtr(res.Data.GetString(basemodel.FieldId)))
	}

	return res, err
}

func (this *SalesPricelistItemApplicationServiceImpl) Update(
	ctx corectx.Context, cmd composable.UpdateCommand,
) (*composable.MutateResult, error) {
	res, err := this.CrudApplicationService.Update(ctx, cmd)
	if err == nil && res.ClientErrors.Count() == 0 {
		this.pricingChangedPub.PublishAsync(ctx, itEvent.PricingChangedPricelistItem, lo.FromPtr(cmd.GetString(basemodel.FieldOrgId)), lo.FromPtr(cmd.GetString(basemodel.FieldId)))
	}

	return res, err
}

func (this *SalesPricelistItemApplicationServiceImpl) Delete(
	ctx corectx.Context, cmd composable.DeleteCommand,
) (*composable.MutateResult, error) {
	res, err := this.CrudApplicationService.Delete(ctx, cmd)
	if err == nil && res.ClientErrors.Count() == 0 {
		this.pricingChangedPub.PublishAsync(ctx, itEvent.PricingChangedPricelistItem, lo.FromPtr(cmd.GetString(basemodel.FieldOrgId)), lo.FromPtr(cmd.GetString(basemodel.FieldId)))
	}

	return res, err
}

func (this *SalesPricelistItemApplicationServiceImpl) SetArchived(
	ctx corectx.Context, cmd composable.SetArchivedCommand,
) (*composable.MutateResult, error) {
	res, err := this.CrudApplicationService.SetArchived(ctx, cmd)
	if err == nil && res.ClientErrors.Count() == 0 {
		this.pricingChangedPub.PublishAsync(ctx, itEvent.PricingChangedPricelistItem, lo.FromPtr(cmd.GetString(basemodel.FieldOrgId)), lo.FromPtr(cmd.GetString(basemodel.FieldId)))
	}

	return res, err
}

func (this *SalesComboApplicationServiceImpl) Create(
	ctx corectx.Context, cmd composable.CreateCommand,
) (*composable.CreateResult, error) {
	res, err := this.CrudApplicationService.Create(ctx, cmd)
	if err == nil && res.HasData && res.ClientErrors.Count() == 0 {
		this.pricingChangedPub.PublishAsync(ctx, itEvent.PricingChangedCombo, lo.FromPtr(res.Data.GetString(basemodel.FieldOrgId)), lo.FromPtr(res.Data.GetString(basemodel.FieldId)))
	}

	return res, err
}

func (this *SalesComboApplicationServiceImpl) Update(
	ctx corectx.Context, cmd composable.UpdateCommand,
) (*composable.MutateResult, error) {
	res, err := this.CrudApplicationService.Update(ctx, cmd)
	if err == nil && res.ClientErrors.Count() == 0 {
		this.pricingChangedPub.PublishAsync(ctx, itEvent.PricingChangedCombo, lo.FromPtr(cmd.GetString(basemodel.FieldOrgId)), lo.FromPtr(cmd.GetString(basemodel.FieldId)))
	}

	return res, err
}

func (this *SalesComboApplicationServiceImpl) Delete(
	ctx corectx.Context, cmd composable.DeleteCommand,
) (*composable.MutateResult, error) {
	res, err := this.CrudApplicationService.Delete(ctx, cmd)
	if err == nil && res.ClientErrors.Count() == 0 {
		this.pricingChangedPub.PublishAsync(ctx, itEvent.PricingChangedCombo, lo.FromPtr(cmd.GetString(basemodel.FieldOrgId)), lo.FromPtr(cmd.GetString(basemodel.FieldId)))
	}

	return res, err
}

func (this *SalesComboApplicationServiceImpl) SetArchived(
	ctx corectx.Context, cmd composable.SetArchivedCommand,
) (*composable.MutateResult, error) {
	res, err := this.CrudApplicationService.SetArchived(ctx, cmd)
	if err == nil && res.ClientErrors.Count() == 0 {
		this.pricingChangedPub.PublishAsync(ctx, itEvent.PricingChangedCombo, lo.FromPtr(cmd.GetString(basemodel.FieldOrgId)), lo.FromPtr(cmd.GetString(basemodel.FieldId)))
	}

	return res, err
}

func (this *SalesComboComponentApplicationServiceImpl) Create(
	ctx corectx.Context, cmd composable.CreateCommand,
) (*composable.CreateResult, error) {
	res, err := this.CrudApplicationService.Create(ctx, cmd)
	if err == nil && res.HasData && res.ClientErrors.Count() == 0 {
		this.pricingChangedPub.PublishAsync(ctx, itEvent.PricingChangedComboComponent, lo.FromPtr(res.Data.GetString(basemodel.FieldOrgId)), lo.FromPtr(res.Data.GetString(basemodel.FieldId)))
	}

	return res, err
}

func (this *SalesComboComponentApplicationServiceImpl) Update(
	ctx corectx.Context, cmd composable.UpdateCommand,
) (*composable.MutateResult, error) {
	res, err := this.CrudApplicationService.Update(ctx, cmd)
	if err == nil && res.ClientErrors.Count() == 0 {
		this.pricingChangedPub.PublishAsync(ctx, itEvent.PricingChangedComboComponent, lo.FromPtr(cmd.GetString(basemodel.FieldOrgId)), lo.FromPtr(cmd.GetString(basemodel.FieldId)))
	}

	return res, err
}

func (this *SalesComboComponentApplicationServiceImpl) Delete(
	ctx corectx.Context, cmd composable.DeleteCommand,
) (*composable.MutateResult, error) {
	res, err := this.CrudApplicationService.Delete(ctx, cmd)
	if err == nil && res.ClientErrors.Count() == 0 {
		this.pricingChangedPub.PublishAsync(ctx, itEvent.PricingChangedComboComponent, lo.FromPtr(cmd.GetString(basemodel.FieldOrgId)), lo.FromPtr(cmd.GetString(basemodel.FieldId)))
	}

	return res, err
}

func (this *SalesComboComponentApplicationServiceImpl) SetArchived(
	ctx corectx.Context, cmd composable.SetArchivedCommand,
) (*composable.MutateResult, error) {
	res, err := this.CrudApplicationService.SetArchived(ctx, cmd)
	if err == nil && res.ClientErrors.Count() == 0 {
		this.pricingChangedPub.PublishAsync(ctx, itEvent.PricingChangedComboComponent, lo.FromPtr(cmd.GetString(basemodel.FieldOrgId)), lo.FromPtr(cmd.GetString(basemodel.FieldId)))
	}

	return res, err
}
