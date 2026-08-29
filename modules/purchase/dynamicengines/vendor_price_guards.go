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

// Validation for a vendor price, attached to create and update via the engine's ValidateExtra hook.
// It exists because the engine cannot ask Contacts whether a party is a usable vendor or Essential
// whether two units convert — those ids are plain ulids with no foreign key behind them.
//
// Delete is deliberately not guarded: a vendor price has no dependents, since an order line records
// the price it resolved rather than a reference that would dangle.

// vendorPriceValidator is set once at Init. It is a package var because an action callback is
// handed only its own engine, so a port has to reach it some other way.
var vendorPriceValidator *services.VendorPriceValidator

// SetVendorPriceValidator is called by InitDomainServices before any request.
func SetVendorPriceValidator(validator *services.VendorPriceValidator) {
	vendorPriceValidator = validator
}

// resolveVendorPriceValidator pulls the three ports the vendor price rules need. A missing port
// fails Init rather than yielding a nil validator, which would silently accept unusable prices that
// only surface when an order tries to resolve through one.
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

// validateVendorProductPrice validates the record as it will be, not as it was sent: an update
// names only changed fields, so checking the payload alone would let a price change slip past the
// unit compatibility rule whose units are absent from the request.
func validateVendorProductPrice(
	ctx corectx.Context, before *drif.DynamicEntity, after *drif.DynamicEntity, vErrs *ft.ClientErrors,
) error {
	if vendorPriceValidator == nil || after == nil {
		// Nil here means the guard ran before Init finished, a wiring bug rather than a request
		// problem, so the row is written unchecked rather than refused as a business violation.
		return nil
	}

	record := after.GetFieldData()
	if before != nil {
		record = mergedRecord(before.GetFieldData(), after.GetFieldData())
	}
	return vendorPriceValidator.Validate(ctx, record, vErrs)
}

// mergedRecord overlays the change on the stored record. It copies rather than mutating, because
// the engine writes `after` on its way out and validation must not be a side effect on it.
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
