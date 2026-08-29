package services

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// SalesPricelistDomainServiceImpl adds the two rules a pricelist carries that plain CRUD cannot
// express: at most one default per organization, and a currency that stops being editable once
// anything depends on it. It wraps rather than replaces the engine's default service.
type SalesPricelistDomainServiceImpl struct {
	drif.DynamicResourceService
}

func NewSalesPricelistDomainService(base drif.DynamicResourceService) *SalesPricelistDomainServiceImpl {
	return &SalesPricelistDomainServiceImpl{DynamicResourceService: base}
}

// SetDefault promotes one pricelist to be its organization's default, demoting whatever held that
// place before.
//
// A single transaction rather than two client calls, which would leave a window with no default (or
// two) and could leave it that way if the second call failed. Demotion loops over every current
// default, so an invariant already broken by an earlier race is repaired instead of tripping this.
func (this *SalesPricelistDomainServiceImpl) SetDefault(
	ctx corectx.Context, pricelistId string,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	var result *dyn.OpResult[dyn.MutateResultData]

	err := withTransaction(ctx, models.SalesPricelistSchemaName, func(tranxCtx corectx.Context) error {
		target, err := loadRecord(tranxCtx, models.SalesPricelistSchemaName,
			models.SalesPricelistFieldId, pricelistId)
		if err != nil {
			return err
		}
		if target == nil {
			result = notFoundResult(models.SalesPricelistSchemaName, pricelistId)
			return nil
		}
		// An archived list may not become the default: the default is what a new order falls back to,
		// and an archived list may not price new business. Unarchive it first.
		if boolOf(target, basemodel.FieldIsArchived) {
			result = violationResult(models.SalesPricelistSchemaName,
				"sales_pricelist.archived",
				"an archived pricelist cannot be made the default; unarchive it first")
			return nil
		}

		orgId := stringOf(target, models.SalesPricelistFieldOrgId)
		if orgId == "" {
			result = violationResult(models.SalesPricelistSchemaName,
				"sales_pricelist.org_required",
				"a pricelist must belong to an organization to be its default")
			return nil
		}

		engine, err := engineFor(models.SalesPricelistSchemaName)
		if err != nil {
			return err
		}
		current, err := models.FindDefaultPricelists(tranxCtx, engine.ResourceRepository(), orgId)
		if err != nil {
			return err
		}

		for _, existing := range current {
			if stringOf(existing, models.SalesPricelistFieldId) == pricelistId {
				continue // Already the default. A retry is not an error.
			}
			if err := writeChanges(tranxCtx, models.SalesPricelistSchemaName, existing,
				dmodel.DynamicFields{models.SalesPricelistFieldIsDefault: false}); err != nil {
				return err
			}
		}

		result = mutateOk()
		if boolOf(target, models.SalesPricelistFieldIsDefault) {
			return nil // Already set; the demotion loop above was the only work to do.
		}
		return writeChanges(tranxCtx, models.SalesPricelistSchemaName, target,
			dmodel.DynamicFields{models.SalesPricelistFieldIsDefault: true})
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

// Update refuses a currency change once the list has priced anything.
//
// Prices carry no currency of their own, so changing the list's does not convert them, it
// reinterprets them, and no FX service exists to convert them either. Create a new list instead.
// The guard is on having RULES rather than on having priced a transaction: a transaction snapshots
// its own price and currency, so the rules are what would be reinterpreted, and the answer does not
// change as order history grows.
func (this *SalesPricelistDomainServiceImpl) Update(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	if _, changing := params[models.SalesPricelistFieldCurrencyId]; !changing {
		return this.DynamicResourceService.Update(ctx, params)
	}

	pricelistId := stringOf(params, models.SalesPricelistFieldId)
	stored, err := loadRecord(ctx, models.SalesPricelistSchemaName,
		models.SalesPricelistFieldId, pricelistId)
	if err != nil {
		return nil, err
	}
	// A missing record is the engine's business to report, not this guard's. Falling through lets it
	// produce its own not-found rather than a second shape for the same failure.
	if stored == nil {
		return this.DynamicResourceService.Update(ctx, params)
	}

	if !isCurrencyChanging(params, stored) {
		return this.DynamicResourceService.Update(ctx, params)
	}

	engine, err := engineFor(models.SalesPricelistSchemaName)
	if err != nil {
		return nil, err
	}
	items, err := models.CountPricelistItems(ctx, engine.ResourceRepository(), pricelistId)
	if err != nil {
		return nil, err
	}
	if len(items) > 0 {
		return violationResult(models.SalesPricelistSchemaName,
			"sales_pricelist.currency_in_use",
			"this pricelist already has prices quoted in its current currency; "+
				"create a new pricelist for a different currency"), nil
	}

	return this.DynamicResourceService.Update(ctx, params)
}

// isCurrencyChanging reports whether an update actually moves the pricelist to a different
// currency, as opposed to naming the one it already has.
//
// Split out from Update because it needs no repository, which makes the rule testable without a
// registry. A client that PUTs the whole record sends currency_id every time, so treating present
// as changing would refuse every such update the moment a list had one rule.
//
// An absent or non-string value counts as no change and falls through to the engine. Never a bare
// type assertion: the value arrives from a JSON body, so a number where an id belongs would panic
// the request rather than be reported as invalid by the engine's own validation.
func isCurrencyChanging(params dmodel.DynamicFields, stored dmodel.DynamicFields) bool {
	requested, ok := params[models.SalesPricelistFieldCurrencyId].(string)
	if !ok || requested == "" {
		return false
	}
	return requested != stringOf(stored, models.SalesPricelistFieldCurrencyId)
}
