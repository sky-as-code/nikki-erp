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

// The payment-method configuration of a sales channel, served as actions on the channel resource
// rather than as a resource of its own.
//
// That is the shape of the thing: a mapping row has no identity a client would name, so there is
// nothing to CRUD. Serving them as channel actions also means the engine performs the permission
// check before MainProcess runs, against the same codes 1007002_sales_iam.sql seeds — so these
// callbacks authorize nothing themselves, exactly like the lifecycle ones beside them.

const (
	ActionPaymentMethods       = "payment_methods"
	ActionEnablePaymentMethod  = "enable_payment_method"
	ActionDisablePaymentMethod = "disable_payment_method"
)

const paramPaymentMethodId = "payment_method_id"

// paymentMethods is the port onto paymentinvoice, installed by Init.
//
// It is a package variable set through a setter rather than an import, for the layering rule: this
// package may not import infra/, and infra/external is where the binding to the other module lives.
// The interface it is typed as belongs to Sales, so naming it here crosses no boundary.
var paymentMethods itExt.PaymentMethodExtService

// SetPaymentMethodPort installs the port the payment actions read. Init calls it before any request
// is served.
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

// defineChannelPaymentActions adds the three payment-configuration actions to the channel engine.
//
// Listing reuses the read permission: it is a read of the channel's configuration, and a role able
// to see a channel but not what it accepts would be looking at half a record. Enabling and
// disabling each take their own code, because changing what a channel can be paid through is a
// materially different power from editing its name.
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

// processListPaymentMethods answers the merged view (CR §29).
//
// The merge happens here rather than in the browser because it needs both lists at once: a method
// this channel has enabled that paymentinvoice no longer reports is stale, and only a reader
// holding both can see that. A frontend calling two module APIs would have to reimplement the rule
// and would get it wrong the first time either call failed.
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

	// Upstream first, and a failure here fails the whole call (CR §35). Answering from the local
	// mappings alone would present a filter as the master list: every enabled method present, every
	// disabled one silently absent, and nothing to tell that apart from a correct answer.
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
	// The mappings upstream did not account for. Reported, never dropped (CR §34): dropping one
	// would hide the only evidence that a channel is configured for something that cannot happen,
	// and leave nothing for an administrator to act on.
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

// processEnablePaymentMethod validates upstream, then writes the mapping (CR §31).
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
	// A suspended or archived channel is not reconfigured, for the same reason it takes no orders:
	// the change would have no effect now and would be a surprise if the channel came back.
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
	// No amount is passed: this decides whether the method may ever be offered on this channel, not
	// whether one particular payment would pass its bounds.
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

// processDisablePaymentMethod removes the mapping (CR §32).
//
// Neither the channel's state nor the method's usability is checked, and both omissions are the
// same rule: taking a payment method away is always safe, and the states where somebody most needs
// to are exactly the ones a validating version would refuse — a stale mapping, or a channel
// suspended because of it.
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
// REST layer replies 400: the caller left something out and can put it back.
func channelPaymentRefusal(key, message string) *drif.ActionResult {
	vErrs := ft.NewClientErrors()
	vErrs.Append(*ft.NewBusinessViolation(models.SalesChannelSchemaName, key, message))
	return &drif.ActionResult{ClientErrors: *vErrs}
}

// stringOf reads one string out of a record without a bare type assertion, for the same reason
// readStringParam does: a repository round-trip can hand back a different concrete type, and a bare
// assertion panics the request.
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
