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

// Vendor and currency rules for an order. Both references are plain ulid fields with no edge:
// validation is by port call, not by foreign key, because the SQL tooling emits only Purchase's
// own tables and an Atlas diff would fail on a constraint pointing outside them.

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
// vendor only when the caller named none; overriding an explicit choice would silently
// re-denominate the order.
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

// assertOrderableVendor refuses a vendor a new order may not name. The refusal comes from Contacts
// rather than a status comparison here, so "may be ordered from" has one definition.
func (this *OrderReferenceValidator) assertOrderableVendor(
	ctx corectx.Context, vendorId, orgId, field string, vErrs *ft.ClientErrors,
) (*itExt.GetVendorResultData, error) {
	if vendorId == "" {
		// vendor_id is required_for_create, so the base call already reports the omission.
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

// assertUsableCurrency refuses a currency a new order may not be denominated in. An empty currency
// is allowed; totals then round to two places until one is set.
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

// ScaleFor returns the number of decimal places an order's amounts are rounded to, taken from the
// currency: 0 for VND, 2 for USD, 3 for KWD. Rounding to more places than a currency has produces a
// total nobody can pay. With no currency, or a failed lookup, it falls back to two places rather
// than erroring, since an order with no currency yet is an ordinary draft.
func (this *OrderReferenceValidator) ScaleFor(ctx corectx.Context, currencyId string) int32 {
	if currencyId == "" {
		return defaultScale
	}

	found, err := this.currencies.GetCurrency(ctx, itExt.GetCurrencyQuery{Id: model.Id(currencyId)})
	if err != nil || found == nil || !found.HasData {
		return defaultScale
	}
	// An inactive or archived currency still rounds, so amounts already recorded in it keep
	// reconciling.
	return found.Data.DecimalPlaces
}

// PrepareAgreement validates an agreement's vendor and currency. Unlike an order, an agreement's
// vendor is optional, since a template may be drafted before anyone has chosen a supplier; a named
// vendor must still be orderable and still defaults the currency.
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
