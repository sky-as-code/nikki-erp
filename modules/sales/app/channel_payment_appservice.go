package app

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	c "github.com/sky-as-code/nikki-erp/modules/sales/constants"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/services"
	"github.com/sky-as-code/nikki-erp/modules/sales/dynamicengines"
	it "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/channel"
	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
)

// ChannelPaymentApplicationServiceImpl configures which payment methods a sales channel accepts.
//
// It holds the paymentinvoice port because every operation here needs the other module's view:
// listing merges against it, enabling validates against it, and only disabling deliberately does
// not.
type ChannelPaymentApplicationServiceImpl struct {
	methods itExt.PaymentMethodExtService
}

func NewChannelPaymentApplicationServiceImpl(
	methods itExt.PaymentMethodExtService,
) it.ChannelPaymentAppService {
	return &ChannelPaymentApplicationServiceImpl{methods: methods}
}

// ListChannelPaymentMethods answers the merged view of CR §29.
//
// The merge is the requirement: the frontend must never join across two modules. Three states come
// out of it and all three are visible — a method offered upstream and enabled here, one offered and
// not enabled, and one enabled here that upstream no longer reports at all. The third is the stale
// case (CR §34) and it is the reason the mapping list is walked separately rather than being used
// only to set a flag on the upstream rows.
func (this *ChannelPaymentApplicationServiceImpl) ListChannelPaymentMethods(
	ctx corectx.Context, query it.ListChannelPaymentMethodsQuery,
) (*it.ListChannelPaymentMethodsResult, error) {
	if cErrs := assertPermission(ctx, drif.PermissionRead,
		c.SalesChannelResource, c.ResourceScopeOrg); cErrs != nil {
		return &it.ListChannelPaymentMethodsResult{ClientErrors: *cErrs}, nil
	}

	channelId, cErrs, err := this.resolveChannelId(ctx, query.SalesChannelId, query.SalesChannelCode)
	if err != nil {
		return nil, err
	}
	if cErrs != nil {
		return &it.ListChannelPaymentMethodsResult{ClientErrors: *cErrs}, nil
	}

	// Upstream first, and its failure is fatal to the whole operation (CR §35). Falling back to the
	// local mappings alone would present them as the master list, which they are not: they are a
	// filter over it. A screen rendered from the filter alone would show a plausible, wrong answer
	// — every enabled method present, every disabled one silently absent — and an administrator
	// would have no way to tell it apart from a correct one.
	upstream, err := this.methods.ListPaymentMethods(ctx, itExt.ListPaymentMethodsQuery{})
	if err != nil {
		return nil, err
	}
	if upstream.ClientErrors.Count() > 0 {
		return &it.ListChannelPaymentMethodsResult{ClientErrors: upstream.ClientErrors}, nil
	}

	service, err := dynamicengines.ChannelPaymentService()
	if err != nil {
		return nil, err
	}
	mappings, err := service.ListMappings(ctx, channelId)
	if err != nil {
		return nil, err
	}

	enabled := make(map[string]bool, len(mappings))
	for _, mapping := range mappings {
		enabled[stringOf(mapping, models.SalesChannelPaymentRelFieldPaymentMethodId)] = true
	}

	merged := make([]it.ChannelPaymentMethodData, 0, len(upstream.Data))
	seen := make(map[string]bool, len(upstream.Data))
	for _, method := range upstream.Data {
		seen[method.Id] = true
		row := it.ChannelPaymentMethodData{
			PaymentMethodId: method.Id,
			Code:            method.Code,
			Name:            method.Name,
			IsEnabled:       enabled[method.Id],
			IsUsable:        method.IsUsable,
			UnusableReason:  method.UnusableReason,
		}
		if query.EnabledOnly && !row.IsEnabled {
			continue
		}
		merged = append(merged, row)
	}

	// Now the mappings upstream did not account for. A row here names a method paymentinvoice has
	// stopped reporting — deleted, or belonging to a feature this build no longer ships. It is
	// reported rather than dropped, because dropping it would hide the only evidence that the
	// channel is configured for something that cannot happen, and would leave the administrator
	// with nothing to click to fix it.
	for methodId := range enabled {
		if seen[methodId] {
			continue
		}
		merged = append(merged, it.ChannelPaymentMethodData{
			PaymentMethodId: methodId,
			IsEnabled:       true,
			IsUsable:        false,
			IsStale:         true,
			UnusableReason:  "unknown_to_paymentinvoice",
		})
	}

	return &it.ListChannelPaymentMethodsResult{HasData: true, Data: merged}, nil
}

// EnableChannelPaymentMethod validates the method upstream and then writes the mapping (CR §31).
//
// The channel must be usable and the method must be usable. Both are checked before anything is
// written, so a refusal leaves no trace — which matters because the caller will retry, and a
// half-applied enable would make the retry look like a duplicate.
func (this *ChannelPaymentApplicationServiceImpl) EnableChannelPaymentMethod(
	ctx corectx.Context, command it.ChannelPaymentMethodCommand,
) (*it.ChannelPaymentMutationResult, error) {
	if cErrs := assertPermission(ctx, c.ActionEnablePaymentMethod,
		c.SalesChannelResource, c.ResourceScopeOrg); cErrs != nil {
		return &it.ChannelPaymentMutationResult{ClientErrors: *cErrs}, nil
	}
	if command.PaymentMethodId == "" {
		return paymentRejection("sales_channel.payment_method_required",
			"enabling a payment method requires its id"), nil
	}

	channel, cErrs, err := this.resolveChannel(ctx, command.SalesChannelId, command.SalesChannelCode)
	if err != nil {
		return nil, err
	}
	if cErrs != nil {
		return &it.ChannelPaymentMutationResult{ClientErrors: *cErrs}, nil
	}

	// A suspended or archived channel is not configured, for the same reason it takes no orders:
	// changing what a retired channel accepts is a change with no effect that later becomes a
	// surprise if it is ever brought back.
	if !models.NewSalesChannelFrom(channel).IsActive() {
		return paymentRejection("sales_channel.not_usable",
			"a suspended or archived sales channel cannot have its payment methods changed"), nil
	}

	// The upstream check. Without it a channel could be configured for a method this deployment
	// cannot serve, and the failure would surface at the checkout rather than at the configuration
	// screen where somebody could act on it. No amount is passed: this decides whether the method
	// may ever be offered, not whether one particular payment would pass its bounds.
	usable, err := this.methods.AssertUsable(ctx, itExt.AssertUsableQuery{
		PaymentMethodId: command.PaymentMethodId,
	})
	if err != nil {
		return nil, err
	}
	if usable.ClientErrors.Count() > 0 {
		return &it.ChannelPaymentMutationResult{ClientErrors: usable.ClientErrors}, nil
	}

	service, err := dynamicengines.ChannelPaymentService()
	if err != nil {
		return nil, err
	}
	result, err := service.Enable(ctx,
		stringOf(channel, models.SalesChannelFieldId), command.PaymentMethodId)
	if err != nil {
		return nil, err
	}
	if result.ClientErrors.Count() > 0 {
		return &it.ChannelPaymentMutationResult{ClientErrors: result.ClientErrors}, nil
	}
	return &it.ChannelPaymentMutationResult{HasData: true}, nil
}

// DisableChannelPaymentMethod removes the mapping (CR §32).
//
// It deliberately does not validate the method upstream, and does not require the channel to be
// active. Both omissions are the same rule: taking a payment method away is always safe, and the
// states in which somebody most needs to is exactly the set where a validating version would refuse
// — a stale mapping, or a channel suspended because of it.
func (this *ChannelPaymentApplicationServiceImpl) DisableChannelPaymentMethod(
	ctx corectx.Context, command it.ChannelPaymentMethodCommand,
) (*it.ChannelPaymentMutationResult, error) {
	if cErrs := assertPermission(ctx, c.ActionDisablePaymentMethod,
		c.SalesChannelResource, c.ResourceScopeOrg); cErrs != nil {
		return &it.ChannelPaymentMutationResult{ClientErrors: *cErrs}, nil
	}
	if command.PaymentMethodId == "" {
		return paymentRejection("sales_channel.payment_method_required",
			"disabling a payment method requires its id"), nil
	}

	channelId, cErrs, err := this.resolveChannelId(ctx,
		command.SalesChannelId, command.SalesChannelCode)
	if err != nil {
		return nil, err
	}
	if cErrs != nil {
		return &it.ChannelPaymentMutationResult{ClientErrors: *cErrs}, nil
	}

	service, err := dynamicengines.ChannelPaymentService()
	if err != nil {
		return nil, err
	}
	result, err := service.Disable(ctx, channelId, command.PaymentMethodId)
	if err != nil {
		return nil, err
	}
	if result.ClientErrors.Count() > 0 {
		return &it.ChannelPaymentMutationResult{ClientErrors: result.ClientErrors}, nil
	}
	return &it.ChannelPaymentMutationResult{HasData: true}, nil
}

// IsPaymentMethodEnabledForChannel is the enforcement query SALES-027 will call.
//
// It answers the mapping alone. Usability is paymentinvoice's judgement and is asked separately,
// because the two can disagree and the caller needs to know which one refused: a method not mapped
// is a configuration problem for this channel, a method mapped but unusable is a problem with the
// method everywhere.
//
// A channel that cannot be resolved answers false rather than erroring. Default-deny holds all the
// way down (CR §76): the safe answer to "may this take money" is no.
func (this *ChannelPaymentApplicationServiceImpl) IsPaymentMethodEnabledForChannel(
	ctx corectx.Context, query it.IsPaymentMethodEnabledQuery,
) (*it.IsPaymentMethodEnabledResult, error) {
	if cErrs := assertPermission(ctx, drif.PermissionRead,
		c.SalesChannelResource, c.ResourceScopeOrg); cErrs != nil {
		return &it.IsPaymentMethodEnabledResult{ClientErrors: *cErrs}, nil
	}

	channelId, cErrs, err := this.resolveChannelId(ctx,
		query.SalesChannelId, query.SalesChannelCode)
	if err != nil {
		return nil, err
	}
	if cErrs != nil {
		return &it.IsPaymentMethodEnabledResult{HasData: true, Data: false}, nil
	}

	service, err := dynamicengines.ChannelPaymentService()
	if err != nil {
		return nil, err
	}
	isEnabled, err := service.IsEnabled(ctx, channelId, query.PaymentMethodId)
	if err != nil {
		return nil, err
	}
	return &it.IsPaymentMethodEnabledResult{HasData: true, Data: isEnabled}, nil
}

// resolveChannel accepts either identifier and answers the row.
//
// Both are offered because the two kinds of caller hold different things: a REST client came from a
// listing and has the id, an integrating module stores only the code. The id wins when both are
// given, since it is the more specific of the two.
func (this *ChannelPaymentApplicationServiceImpl) resolveChannel(
	ctx corectx.Context, channelId string, channelCode string,
) (dmodel.DynamicFields, *ft.ClientErrors, error) {
	engine, err := services.EngineFor(models.SalesChannelSchemaName)
	if err != nil {
		return nil, nil, err
	}

	if channelId != "" {
		found, err := engine.ResourceRepository().FindByKeys(ctx, dmodel.DynamicFields{
			models.SalesChannelFieldId: channelId,
		})
		if err != nil {
			return nil, nil, err
		}
		if found == nil || !found.HasData {
			return nil, channelNotFound("no sales channel with id '" + channelId + "'"), nil
		}
		return found.Data, nil, nil
	}

	code := services.NormalizeChannelCode(channelCode)
	if code == "" {
		return nil, channelNotFound("a sales channel id or code is required"), nil
	}
	found, err := models.FindSalesChannelByCode(ctx, engine.ResourceRepository(), code)
	if err != nil {
		return nil, nil, err
	}
	if len(found) == 0 {
		return nil, channelNotFound("no sales channel with code '" + code + "'"), nil
	}
	return found[0], nil, nil
}

func (this *ChannelPaymentApplicationServiceImpl) resolveChannelId(
	ctx corectx.Context, channelId string, channelCode string,
) (string, *ft.ClientErrors, error) {
	channel, cErrs, err := this.resolveChannel(ctx, channelId, channelCode)
	if err != nil || cErrs != nil {
		return "", cErrs, err
	}
	return stringOf(channel, models.SalesChannelFieldId), nil, nil
}

func channelNotFound(message string) *ft.ClientErrors {
	vErrs := ft.NewClientErrors()
	vErrs.Append(*ft.NewBusinessViolation(
		models.SalesChannelSchemaName, "sales_channel.not_found", message))
	return vErrs
}

func paymentRejection(key, message string) *it.ChannelPaymentMutationResult {
	vErrs := ft.NewClientErrors()
	vErrs.Append(*ft.NewBusinessViolation(models.SalesChannelSchemaName, key, message))
	return &it.ChannelPaymentMutationResult{ClientErrors: *vErrs}
}
