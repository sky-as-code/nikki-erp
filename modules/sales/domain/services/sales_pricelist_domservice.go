package services

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// SalesPricelistDomainServiceImpl adds to the engine's default service the two rules a pricelist
// carries that plain CRUD cannot express: at most one default per organization, and a currency
// that stops being editable once anything depends on it.
//
// It wraps rather than replaces, like every other derived service here: ordinary create, read and
// delete keep running through the engine's implementation underneath.
type SalesPricelistDomainServiceImpl struct {
	drif.DynamicResourceService
}

func NewSalesPricelistDomainService(base drif.DynamicResourceService) *SalesPricelistDomainServiceImpl {
	return &SalesPricelistDomainServiceImpl{DynamicResourceService: base}
}

// SetDefault promotes one pricelist to be its organization's default, demoting whatever held that
// place before (BR-PRICE-SAL-003, section 9).
//
// It is a single operation rather than two updates, and that is the whole reason it exists. Doing
// it as "clear the old, set the new" from a client leaves a window in which the organization has
// no default — and if the second call fails, it stays that way. Doing it as "set the new, clear the
// old" leaves a window with two. Inside one transaction there is no window at all.
//
// Demotion is a loop rather than a single update because FindDefaultPricelists returns every
// current default: if the invariant has already been broken by an earlier race, this repairs it
// instead of tripping over it.
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
		// An archived list may not become the default: the default is what a new order falls back
		// to, and an archived list may not price new business (PRICE-INV-023). Unarchive it first.
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

// Update refuses a currency change once the list has priced anything (section 8).
//
// Every price on a list is quoted in that list's currency and carries no currency of its own, so
// changing it does not convert those prices — it reinterprets them. A list of VND prices relabelled
// USD is off by a factor of twenty-five thousand, silently, and there is no FX service that could
// have converted them even if this wanted to (BR-PRICE-CUR-004). Create a new list instead.
//
// The guard is on having RULES rather than on having priced a transaction. A transaction already
// snapshots its own price and currency, so it is not retroactively harmed; the rules are what would
// be reinterpreted. Checking the cheaper and stricter condition also means the answer does not
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
	// A missing record is the engine's business to report, not this guard's. Falling through lets
	// it produce its own not-found rather than inventing a second shape for the same failure.
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
// Split out from Update because it is the whole decision and it needs no repository: everything it
// reads is already in hand. That makes the rule testable without a registry, which is what the
// rest of this package does with its pure parts too.
//
// A client that PUTs the whole record sends currency_id every time. Treating "present" as
// "changing" would refuse every such update the moment a list had one rule, which is not what
// section 8 asks for — it forbids REinterpreting the prices, not resubmitting the same currency.
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
