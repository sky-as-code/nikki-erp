package services

import (
	"go.bryk.io/pkg/errors"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/models"
)

// Where an order's gateway credentials come from.
//
// An order may name a payment profile, and then it is collected into that merchant account; an
// order that names none is collected with the credentials in this deployment's configuration,
// which is what every order taken before profiles existed did and what a single-account deployment
// still does. The choice is made once, when the payment is created, and recorded on the order —
// because every later step has to use the same credentials the money was taken with. A refund
// issued from a different account would be refused, and a callback verified against a different
// secret would look like a forgery.

// loadProfileForCreate resolves the profile a new payment names, and refuses one that cannot
// collect it.
//
// Every refusal is a client error: they all describe the caller's request rather than a failure
// here. Answering nil with no error means the caller named no profile, which is allowed.
func (this *OrderDomainService) loadProfileForCreate(
	ctx corectx.Context, profileId string, method models.PaymentMethod, vErrs *ft.ClientErrors,
) (*models.PaymentProfile, error) {
	if profileId == "" {
		return nil, nil
	}

	profile, err := this.profiles.FindById(ctx, profileId)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		appendFieldViolation(vErrs, models.OrderFieldPaymentProfileId,
			"paymentinvoice.payment_profile_not_found",
			"no payment profile with id '"+profileId+"'")
		return nil, nil
	}

	// An archived profile is an account withdrawn from use. Refusing it here is what makes
	// archiving a control rather than a label — the alternative is that money keeps landing in an
	// account somebody has already closed the books on.
	if archived := profile.IsArchived(); archived != nil && *archived {
		appendFieldViolation(vErrs, models.OrderFieldPaymentProfileId,
			"paymentinvoice.payment_profile_archived",
			"payment profile '"+profileId+"' has been withdrawn from use")
		return nil, nil
	}

	// The profile names the gateway its credentials belong to, and the method names the adapter
	// that will be called. If they disagree, the payment would be sent to one gateway signed with
	// another's secret — refused at best, and at worst accepted against the wrong account.
	adapterCode := derefString(method.GetAdapterCode())
	if profileMethod := profile.GetMethod(); profileMethod == nil || string(*profileMethod) != adapterCode {
		appendFieldViolation(vErrs, models.OrderFieldPaymentProfileId,
			"paymentinvoice.payment_profile_method_mismatch",
			"payment profile '"+profileId+"' holds credentials for a different gateway than "+
				"payment method '"+derefString(method.GetCode())+"' is served by")
		return nil, nil
	}

	return profile, nil
}

// profileConfigForOrder returns the credentials an order was collected with, or nil when it was
// collected with the deployment's own.
//
// A profile that has since been deleted answers nil rather than failing: the deployment's
// credentials are the only ones left to try, and refusing to refund or to ask after an order
// because the row describing its account was removed would strand real money.
func (this *OrderDomainService) profileConfigForOrder(
	ctx corectx.Context, order models.Order,
) (map[string]any, error) {
	return this.profileConfigById(ctx, derefString((*string)(order.GetPaymentProfileId())))
}

// profileConfigById is profileConfigForOrder for the callers that hold only the id — the sweeps,
// which read a projection of the order rather than the whole record.
func (this *OrderDomainService) profileConfigById(
	ctx corectx.Context, profileId string,
) (map[string]any, error) {
	if profileId == "" {
		return nil, nil
	}

	profile, err := this.profiles.FindById(ctx, profileId)
	if err != nil {
		return nil, errors.Wrap(err, "profileConfigById")
	}
	if profile == nil {
		return nil, nil
	}

	return profile.ConfigValues(), nil
}

// ProfileConfigForOrderCode returns the credentials the order behind a gateway callback was
// collected with.
//
// The callbacks reach an order by its order_code and need its credentials before they can act:
// MoMo signs its callback with the secret of the account that took the money, so verifying one
// against the deployment's own secret would reject every payment collected through a profile.
//
// An unknown order code answers nil with no error. The callback layer must not distinguish that
// case in its response — telling an unauthenticated caller whether an order exists is how order
// codes get enumerated.
func (this *OrderDomainService) ProfileConfigForOrderCode(
	ctx corectx.Context, orderCode string,
) (map[string]any, error) {
	if orderCode == "" {
		return nil, nil
	}

	order, err := findOrderByCode(ctx, orderCode)
	if err != nil || order == nil {
		return nil, err
	}

	return this.profileConfigForOrder(ctx, *order)
}

// ProfileConfigsByMethod returns the credentials of every profile that can collect through one
// gateway.
//
// It serves the callback that identifies itself by a merchant account rather than by an order: a
// card terminal posts its merchant id and an encrypted body, so the account whose secret decrypts
// that body has to be found before anything inside it can be read. Which of them matches is the
// caller's question, because only the adapter knows what a merchant id is called in its own
// credentials — which is why this hands back the configs rather than doing the matching.
func (this *OrderDomainService) ProfileConfigsByMethod(
	ctx corectx.Context, method models.PaymentProfileMethod,
) ([]map[string]any, error) {
	profiles, err := this.profiles.FindActiveByMethod(ctx, method)
	if err != nil {
		return nil, err
	}

	configs := make([]map[string]any, 0, len(profiles))
	for _, profile := range profiles {
		if values := profile.ConfigValues(); len(values) > 0 {
			configs = append(configs, values)
		}
	}
	return configs, nil
}

// profileConfigOf reads the credentials off a profile that may not be there.
//
// A nil profile is the ordinary case — most deployments have one merchant account and configure it
// — so this is a helper rather than a check at every call site.
func profileConfigOf(profile *models.PaymentProfile) map[string]any {
	if profile == nil {
		return nil
	}
	return profile.ConfigValues()
}
