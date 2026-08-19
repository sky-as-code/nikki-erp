package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	itExt "github.com/sky-as-code/nikki-erp/modules/purchase/interfaces/external"

	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/models"
)

// Vendor and currency rules for an order (D3, D5, D5a).
//
// Both references are plain ulid fields with no edge. Validation is by port call, not by database
// constraint: a foreign key across a module boundary would make Purchase's schema depend on another
// module's tables, and the tooling enforces this — `-createsql -module=purchase` emits only
// Purchase's tables, so the Atlas diff would fail on a constraint pointing outside them.

// OrderReferenceValidator checks an order's vendor and currency, and defaults what it can.
type OrderReferenceValidator struct {
	vendors    itExt.VendorExtService
	currencies itExt.CurrencyExtService
}

func NewOrderReferenceValidator(
	vendors itExt.VendorExtService, currencies itExt.CurrencyExtService,
) *OrderReferenceValidator {
	return &OrderReferenceValidator{vendors: vendors, currencies: currencies}
}

// PrepareOrder validates the vendor and currency on a create, defaulting the currency from the
// vendor when the caller did not name one.
//
// Defaulting is a convenience with a real consequence, so it happens only when the field is absent:
// a caller who named a currency gets the one they named, even if the vendor usually invoices in
// another. Overriding an explicit choice would silently re-denominate an order.
func (this *OrderReferenceValidator) PrepareOrder(
	ctx corectx.Context, params dmodel.DynamicFields, vErrs *ft.ClientErrors,
) error {
	vendorId := stringOf(params, models.PurchaseOrderFieldVendorId)
	orgId := stringOf(params, basemodel.FieldOrgId)

	vendor, err := this.assertOrderableVendor(
		ctx, vendorId, orgId, models.PurchaseOrderFieldVendorId, vErrs)
	if err != nil {
		return err
	}

	if stringOf(params, models.PurchaseOrderFieldCurrencyId) == "" && vendor != nil &&
		vendor.DefaultCurrencyId != nil {
		params[models.PurchaseOrderFieldCurrencyId] = string(*vendor.DefaultCurrencyId)
	}

	return this.assertUsableCurrency(
		ctx, stringOf(params, models.PurchaseOrderFieldCurrencyId), vErrs)
}

// assertOrderableVendor refuses a vendor a new order may not name (D3).
//
// The refusal comes from Contacts rather than from a status comparison here, so that "may be
// ordered from" has one definition. A caller comparing the status itself would have to be found and
// changed the day a fifth status is added.
func (this *OrderReferenceValidator) assertOrderableVendor(
	ctx corectx.Context, vendorId, orgId, field string, vErrs *ft.ClientErrors,
) (*itExt.GetVendorResultData, error) {
	if vendorId == "" {
		// vendor_id is required_for_create, so the base call reports the omission. Restating it
		// here would report the same problem twice in two different shapes.
		return nil, nil
	}

	assertion, err := this.vendors.AssertOrderable(ctx, itExt.AssertOrderableQuery{
		PartyId: model.Id(vendorId),
		OrgId:   model.Id(orgId),
		Field:   field,
	})
	if err != nil {
		return nil, errors.Wrap(err, "assertOrderableVendor")
	}
	if assertion != nil && assertion.ClientErrors.Count() > 0 {
		vErrs.ConcatPtr(&assertion.ClientErrors)
		return nil, nil
	}

	found, err := this.vendors.GetVendor(ctx, itExt.GetVendorQuery{
		PartyId: model.Id(vendorId),
		OrgId:   model.Id(orgId),
	})
	if err != nil {
		return nil, errors.Wrap(err, "assertOrderableVendor")
	}
	if found == nil || !found.HasData {
		return nil, nil
	}
	return &found.Data, nil
}

// assertUsableCurrency refuses a currency a new order may not be denominated in.
//
// An empty currency is allowed: currency_id is optional on the schema, and an order for a vendor
// with no stated terms genuinely has none yet. What that costs is rounding — totals fall back to
// two places until a currency is set, which is why the fallback exists rather than being an error.
func (this *OrderReferenceValidator) assertUsableCurrency(
	ctx corectx.Context, currencyId string, vErrs *ft.ClientErrors,
) error {
	if currencyId == "" {
		return nil
	}

	assertion, err := this.currencies.AssertUsable(ctx, itExt.AssertUsableQuery{
		Id:    model.Id(currencyId),
		Field: models.PurchaseOrderFieldCurrencyId,
	})
	if err != nil {
		return errors.Wrap(err, "assertUsableCurrency")
	}
	if assertion != nil && assertion.ClientErrors.Count() > 0 {
		vErrs.ConcatPtr(&assertion.ClientErrors)
	}
	return nil
}

// ScaleFor returns the number of decimal places an order's amounts are rounded to.
//
// This is what [PUR-014] left as a fixed 2 and what makes Purchase the first reader of
// decimal_places. It matters because a currency has a real number of fractional digits — 0 for VND,
// 2 for USD, 3 for KWD — and rounding to MORE places than a currency has produces a total nobody
// can pay.
//
// It falls back to two places when there is no currency or the lookup fails, rather than erroring.
// An order with no currency yet is an ordinary draft, and failing a totals recompute because a
// currency could not be read would make the money unreadable over a problem that is not about
// money.
func (this *OrderReferenceValidator) ScaleFor(ctx corectx.Context, currencyId string) int32 {
	if currencyId == "" {
		return defaultScale
	}

	found, err := this.currencies.GetCurrency(ctx, itExt.GetCurrencyQuery{Id: model.Id(currencyId)})
	if err != nil || found == nil || !found.HasData {
		return defaultScale
	}
	// An inactive or archived currency still rounds: amounts already recorded in it must keep
	// reconciling, and that is a read, not a new selection.
	return found.Data.DecimalPlaces
}

// PrepareAgreement validates an agreement's vendor and currency.
//
// It differs from PrepareOrder in one way that matters: an agreement's vendor is OPTIONAL. A
// purchase template is a reusable skeleton that may be drafted before anyone has chosen who to buy
// from, so an absent vendor is a legitimate agreement rather than an incomplete one. When a vendor
// IS named it must still be orderable, and its currency still defaults the agreement's.
func (this *OrderReferenceValidator) PrepareAgreement(
	ctx corectx.Context, params dmodel.DynamicFields, vErrs *ft.ClientErrors,
) error {
	vendorId := stringOf(params, models.AgreementFieldVendorId)
	if vendorId == "" {
		return this.assertUsableCurrency(
			ctx, stringOf(params, models.AgreementFieldCurrencyId), vErrs)
	}

	vendor, err := this.assertOrderableVendor(
		ctx, vendorId, stringOf(params, basemodel.FieldOrgId),
		models.AgreementFieldVendorId, vErrs)
	if err != nil {
		return err
	}

	if stringOf(params, models.AgreementFieldCurrencyId) == "" && vendor != nil &&
		vendor.DefaultCurrencyId != nil {
		params[models.AgreementFieldCurrencyId] = string(*vendor.DefaultCurrencyId)
	}

	return this.assertUsableCurrency(
		ctx, stringOf(params, models.AgreementFieldCurrencyId), vErrs)
}
