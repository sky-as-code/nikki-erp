package dynamicengines

import (
	stdErr "errors"

	"go.bryk.io/pkg/errors"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	c "github.com/sky-as-code/nikki-erp/modules/sales/constants"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
)

// Payment-method configuration is served as actions on the channel resource, not a resource of its
// own: a mapping row has no identity a client would name. The engine therefore runs the permission
// check before MainProcess, so these callbacks authorize nothing themselves.

const (
	ActionPaymentMethods       = "payment_methods"
	ActionEnablePaymentMethod  = "enable_payment_method"
	ActionDisablePaymentMethod = "disable_payment_method"
)

const paramPaymentMethodId = "payment_method_id"

// paymentMethods is the port onto paymentinvoice, set through a setter rather than imported: this
// package may not import infra/, where the binding lives.
var paymentMethods itExt.PaymentMethodExtService

// SetPaymentMethodPort must be called by Init before any request is served.
func SetPaymentMethodPort(port itExt.PaymentMethodExtService) {
	paymentMethods = port
}

func paymentMethodPort() (itExt.PaymentMethodExtService, error) {
	if paymentMethods == nil {
		return nil, errors.New(
			"the sales payment method port is not installed; SalesModule.Init must call " +
				"dynamicengines.SetPaymentMethodPort")
	}
	return paymentMethods, nil
}

// Listing reuses the read permission, since it reads the channel's own configuration. Enabling and
// disabling take their own codes: changing what a channel can be paid through is a materially
// different power from editing its name.
func defineChannelPaymentActions(engine drif.DynamicResourceEngine) error {
	return stdErr.Join(
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionPaymentMethods,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/payment_methods",
			Permission:  drif.PermissionRead,
			MainProcess: processListPaymentMethods,
		}),
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionEnablePaymentMethod,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/enable_payment_method",
			Permission:  c.ActionEnablePaymentMethod,
			MainProcess: processEnablePaymentMethod,
		}),
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionDisablePaymentMethod,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/disable_payment_method",
			Permission:  c.ActionDisablePaymentMethod,
			MainProcess: processDisablePaymentMethod,
		}),
	)
}

// processListPaymentMethods merges both lists server-side: only a reader holding both can see that
// a method this channel enabled is one paymentinvoice no longer reports.
func processListPaymentMethods(
	ctx corectx.Context, input drif.ProcessInput,
) (*drif.ActionResult, error) {
	channelId := readStringParam(input.Params, paramId)
	if channelId == "" {
		return channelPaymentRefusal("sales_channel.id_required",
			"listing payment methods requires the channel id"), nil
	}

	port, err := paymentMethodPort()
	if err != nil {
		return nil, err
	}
	service, err := ChannelPaymentService()
	if err != nil {
		return nil, err
	}

	// An upstream failure fails the whole call: the local mappings are a filter, and answering from
	// them alone would present it as the master list.
	upstream, err := port.ListPaymentMethods(ctx, itExt.ListPaymentMethodsQuery{})
	if err != nil {
		return nil, err
	}
	if upstream.ClientErrors.Count() > 0 {
		return &drif.ActionResult{ClientErrors: upstream.ClientErrors}, nil
	}

	mappings, err := service.ListMappings(ctx, channelId)
	if err != nil {
		return nil, err
	}
	enabled := make(map[string]bool, len(mappings))
	for _, mapping := range mappings {
		enabled[stringOf(mapping, models.SalesChannelPaymentRelFieldPaymentMethodId)] = true
	}

	rows := make([]map[string]any, 0, len(upstream.Data)+len(enabled))
	seen := make(map[string]bool, len(upstream.Data))
	for _, method := range upstream.Data {
		seen[method.Id] = true
		rows = append(rows, map[string]any{
			"payment_method_id": method.Id,
			"code":              method.Code,
			"name":              method.Name,
			"is_enabled":        enabled[method.Id],
			"is_usable":         method.IsUsable,
			"unusable_reason":   method.UnusableReason,
			"is_stale":          false,
		})
	}
	// Mappings upstream no longer reports. Reported as stale, never dropped, so an administrator can
	// see and undo a channel configured for something that cannot happen.
	for methodId := range enabled {
		if seen[methodId] {
			continue
		}
		rows = append(rows, map[string]any{
			"payment_method_id": methodId,
			"is_enabled":        true,
			"is_usable":         false,
			"unusable_reason":   "unknown_to_paymentinvoice",
			"is_stale":          true,
		})
	}

	return &drif.ActionResult{HasData: true, Data: rows}, nil
}

func processEnablePaymentMethod(
	ctx corectx.Context, input drif.ProcessInput,
) (*drif.ActionResult, error) {
	channelId := readStringParam(input.Params, paramId)
	methodId := readStringParam(input.Params, paramPaymentMethodId)
	if channelId == "" || methodId == "" {
		return channelPaymentRefusal("sales_channel.payment_method_required",
			"enabling a payment method requires the channel id and the payment method id"), nil
	}

	channelService, err := channelServiceOf(input)
	if err != nil {
		return nil, err
	}
	// A retired channel is not reconfigured: the change has no effect now and becomes a surprise if
	// the channel comes back.
	mutable, err := channelService.AssertMutable(ctx, channelId)
	if err != nil {
		return nil, err
	}
	if mutable.ClientErrors.Count() > 0 {
		return &drif.ActionResult{ClientErrors: mutable.ClientErrors}, nil
	}

	port, err := paymentMethodPort()
	if err != nil {
		return nil, err
	}
	// No amount is passed: this asks whether the method may ever be offered on this channel, not
	// whether one payment is within its bounds.
	usable, err := port.AssertUsable(ctx, itExt.AssertUsableQuery{PaymentMethodId: methodId})
	if err != nil {
		return nil, err
	}
	if usable.ClientErrors.Count() > 0 {
		return &drif.ActionResult{ClientErrors: usable.ClientErrors}, nil
	}

	service, err := ChannelPaymentService()
	if err != nil {
		return nil, err
	}
	result, err := service.Enable(ctx, channelId, methodId)
	return toMutateActionResult(result, err)
}

// processDisablePaymentMethod deliberately checks neither the channel's state nor the method's
// usability: removal is always safe, and the cases where it is most needed (a stale mapping, a
// channel suspended because of it) are exactly what validation would refuse.
func processDisablePaymentMethod(
	ctx corectx.Context, input drif.ProcessInput,
) (*drif.ActionResult, error) {
	channelId := readStringParam(input.Params, paramId)
	methodId := readStringParam(input.Params, paramPaymentMethodId)
	if channelId == "" || methodId == "" {
		return channelPaymentRefusal("sales_channel.payment_method_required",
			"disabling a payment method requires the channel id and the payment method id"), nil
	}

	service, err := ChannelPaymentService()
	if err != nil {
		return nil, err
	}
	result, err := service.Disable(ctx, channelId, methodId)
	return toMutateActionResult(result, err)
}

// channelPaymentRefusal answers a malformed request with a violation rather than an error, so the
// REST layer replies 400.
func channelPaymentRefusal(key, message string) *drif.ActionResult {
	vErrs := ft.NewClientErrors()
	vErrs.Append(*ft.NewBusinessViolation(models.SalesChannelSchemaName, key, message))
	return &drif.ActionResult{ClientErrors: *vErrs}
}

// stringOf avoids a bare type assertion: a repository round-trip can hand back a different concrete
// type, and a bare assertion panics the request.
func stringOf(record map[string]any, field string) string {
	if record == nil {
		return ""
	}
	value, ok := record[field]
	if !ok || value == nil {
		return ""
	}
	if typed, ok := value.(string); ok {
		return typed
	}
	if typed, ok := value.(*string); ok && typed != nil {
		return *typed
	}
	return ""
}
