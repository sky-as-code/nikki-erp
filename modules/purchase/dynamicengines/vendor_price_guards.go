package dynamicengines

import (
	stdErr "errors"

	"go.bryk.io/pkg/errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/services"
	itExt "github.com/sky-as-code/nikki-erp/modules/purchase/interfaces/external"
)

// Validation for a vendor price, attached to create and update.
//
// It hangs off the engine's ValidateExtra hook rather than a derived service because there is no
// behaviour to add — only checks. The engine already does the CRUD, the permission check and the
// org scoping; what it cannot do is ask Contacts whether a party is a usable vendor, or Essential
// whether two units are convertible, because those are plain ulids with no foreign key behind them.
//
// Delete is deliberately NOT guarded. A vendor price is master data with no dependents: a purchase
// order line records the price it resolved, not a reference that would dangle. Archiving is the
// ordinary way to retire one while keeping it resolvable for history (PRICE-INV-024), but a row
// created by mistake should be removable outright.

// vendorPriceValidator is set once at Init.
//
// A package var for the same reason the sales pricing ports are: an action callback is handed only
// its own engine, so a port has to reach it some other way.
var vendorPriceValidator *services.VendorPriceValidator

// SetVendorPriceValidator installs it. InitDomainServices calls this before any request.
func SetVendorPriceValidator(validator *services.VendorPriceValidator) {
	vendorPriceValidator = validator
}

// resolveVendorPriceValidator pulls the three ports the vendor price rules need.
//
// A missing port fails Init rather than yielding a nil validator, matching the two resolvers beside
// it. The alternative is a module that boots and then silently accepts vendor prices for parties
// nobody may order from, on products nobody may buy, in units that convert to nothing — and none of
// those rows announces itself until an order tries to resolve through one.
func resolveVendorPriceValidator() (*services.VendorPriceValidator, error) {
	var vendors itExt.VendorExtService
	var products itExt.ProductExtService
	var uoms itExt.UomExtService

	if err := deps.Invoke(func(svc itExt.VendorExtService) { vendors = svc }); err != nil {
		return nil, stdErr.Join(
			errors.New("the vendor port is not registered; purchase/infra/external must bind it"), err)
	}
	if err := deps.Invoke(func(svc itExt.ProductExtService) { products = svc }); err != nil {
		return nil, stdErr.Join(
			errors.New("the product port is not registered; purchase/infra/external must bind it"), err)
	}
	if err := deps.Invoke(func(svc itExt.UomExtService) { uoms = svc }); err != nil {
		return nil, stdErr.Join(
			errors.New("the UoM port is not registered; purchase/infra/external must bind it"), err)
	}
	return services.NewVendorPriceValidator(vendors, products, uoms), nil
}

func defineVendorProductPriceGuards(engine drif.DynamicResourceEngine) error {
	return attachWriteGuards(engine, models.VendorProductPriceSchemaName,
		validateVendorProductPrice, drif.ActionCreate, drif.ActionUpdate)
}

// validateVendorProductPrice validates the record as it will BE, not as it was sent.
//
// An update names only the fields it changes, so checking the payload alone would let a change to
// the price slip past the unit compatibility rule — the units are unchanged and therefore absent
// from the request, and the check would find nothing to look at. Merging the stored entity with the
// change is what makes the rules mean the same thing on create and on update.
func validateVendorProductPrice(
	ctx corectx.Context, before *drif.DynamicEntity, after *drif.DynamicEntity, vErrs *ft.ClientErrors,
) error {
	if vendorPriceValidator == nil || after == nil {
		// Init binds this or fails, so nil here means the guard ran before Init finished — which is
		// a wiring bug rather than a request problem. Writing the row unchecked is the lesser harm:
		// refusing would answer a business violation for something no caller can act on.
		return nil
	}

	record := after.GetFieldData()
	if before != nil {
		record = mergedRecord(before.GetFieldData(), after.GetFieldData())
	}
	return vendorPriceValidator.Validate(ctx, record, vErrs)
}

// mergedRecord overlays the change on the stored record.
//
// A copy rather than a mutation of either: the engine holds both entities and writes `after` on its
// way out, so mutating it here would make validation a side effect on the record being saved.
func mergedRecord(stored dmodel.DynamicFields, change dmodel.DynamicFields) dmodel.DynamicFields {
	merged := make(dmodel.DynamicFields, len(stored)+len(change))
	for field, value := range stored {
		merged[field] = value
	}
	for field, value := range change {
		merged[field] = value
	}
	return merged
}
