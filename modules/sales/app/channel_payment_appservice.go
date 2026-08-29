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
// It holds the paymentinvoice port because listing merges against it and enabling validates against
// it; only disabling deliberately does not.
type ChannelPaymentApplicationServiceImpl struct {
	methods itExt.PaymentMethodExtService
}

func NewChannelPaymentApplicationServiceImpl(
	methods itExt.PaymentMethodExtService,
) it.ChannelPaymentAppService {
	return &ChannelPaymentApplicationServiceImpl{methods: methods}
}

// ListChannelPaymentMethods merges upstream methods with this channel's mappings so the frontend
// never joins across two modules. The mapping list is walked separately, not just used as a flag on
// upstream rows, so that mappings upstream no longer reports still surface as stale.
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

	// An upstream failure is fatal: the local mappings are a filter over the master list, not the
	// list itself, so falling back to them would render a plausible but wrong screen.
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

	// Mappings upstream no longer reports. Reported as stale rather than dropped, so the
	// administrator can see and undo a channel configured for something that cannot happen.
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

// EnableChannelPaymentMethod validates the channel and the method upstream before writing the
// mapping, so a refusal leaves no trace and a retry cannot look like a duplicate.
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

	// A retired channel is not configured: the change has no effect now and becomes a surprise if
	// the channel is ever brought back.
	if !models.NewSalesChannelFrom(channel).IsActive() {
		return paymentRejection("sales_channel.not_usable",
			"a suspended or archived sales channel cannot have its payment methods changed"), nil
	}

	// Without this check the failure would surface at checkout rather than at the configuration
	// screen. No amount is passed: this asks whether the method may ever be offered, not whether one
	// payment is within its bounds.
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

// DisableChannelPaymentMethod removes the mapping. It deliberately skips the upstream check and
// does not require an active channel: removal is always safe, and the cases where it is most needed
// (a stale mapping, a channel suspended because of it) are exactly what validation would refuse.
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

// IsPaymentMethodEnabledForChannel answers the mapping alone; usability is paymentinvoice's
// separate judgement, so the caller can tell a per-channel configuration problem from a method
// broken everywhere. An unresolvable channel answers false rather than erroring: default-deny.
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

// resolveChannel accepts either identifier, since REST callers hold the id and integrating modules
// store only the code. The id wins when both are given.
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
