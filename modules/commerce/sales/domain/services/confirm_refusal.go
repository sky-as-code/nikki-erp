package services

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"

	itExt "github.com/sky-as-code/nikki-erp/modules/commerce/sales/interfaces/external"
)

// confirmDeps are the ports a refused confirm needs to clean up and requote.
type confirmDeps struct {
	taxSvc       itExt.TaxCalculationExtService
	policy       SalesPolicy
	basisSvc     itExt.ProductPricingBasisExtService
	reservations itExt.FulfillmentReservationExtService
}

// refusedConfirm answers a confirm whose freeze transaction rolled back. A kiosk sale already held
// its stock at step 3b, before the freeze; a refusal leaves the order a draft, so that hold is given
// back rather than left against a sale that did not happen. A lost race for the last use answers
// with the recomputed quote (BR-T04), which no longer offers the exhausted promotion.
func refusedConfirm(
	ctx corectx.Context, orderId string, err error, heldKioskStock bool, deps confirmDeps,
) (*ConfirmOrderResult, *ft.ClientErrors, error) {
	cErrs := numberingRefusalOf(err)
	usageRefusal := false
	if cErrs == nil {
		cErrs = commitRefusalOf(err)
		usageRefusal = cErrs != nil
	}
	if cErrs == nil {
		return nil, nil, err
	}
	if heldKioskStock {
		if _, releaseErr := ReleaseOrderFulfillments(ctx, orderId, deps.reservations); releaseErr != nil {
			return nil, nil, releaseErr
		}
	}
	if usageRefusal {
		attachFreshQuote(ctx, orderId, cErrs, deps)
	}
	return nil, cErrs, nil
}

// attachFreshQuote reprices the draft and puts the new quote under vars.quote of each refusal. A
// reprice that fails leaves the refusal as it was: the caller can still preview the order.
func attachFreshQuote(ctx corectx.Context, orderId string, cErrs *ft.ClientErrors, deps confirmDeps) {
	priced, vErrs, err := RepriceOrder(ctx, orderId, deps.taxSvc, deps.policy, deps.basisSvc)
	if err != nil || vErrs != nil || priced == nil || priced.Promotion == nil {
		return
	}
	quote := priced.Promotion.View()
	for index := range *cErrs {
		if (*cErrs)[index].Vars == nil {
			(*cErrs)[index].Vars = map[string]any{}
		}
		(*cErrs)[index].Vars["quote"] = quote
	}
}
