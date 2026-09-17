package app

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	reguard "github.com/sky-as-code/nikki-erp/modules/core/requestguard"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	c "github.com/sky-as-code/nikki-erp/modules/sales/constants"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/services"
	it "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/channel"
	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
)

func assertPointPaymentPermission(
	ctx corectx.Context, action string, point dmodel.DynamicFields,
) *ft.ClientErrors {
	orgId := point.GetModelId(models.SalesPointFieldOrgId)
	if orgId == nil || *orgId == "" {
		vErrs := ft.NewClientErrors()
		vErrs.Append(*ft.NewInsufficientPermissionsError([]string{
			reguard.BuildExpression(action, c.SalesPointResource, c.ResourceScopeOrg, nil),
		}))
		return vErrs
	}
	return assertPermissionInOrg(ctx, action, c.SalesPointResource, c.ResourceScopeOrg, *orgId)
}

type PointPaymentApplicationServiceImpl struct {
	methods         itExt.PaymentMethodExtService
	pointMappings   *services.PointPaymentDomainServiceImpl
	channelMappings *services.ChannelPaymentDomainServiceImpl
}

func NewPointPaymentApplicationServiceImpl(
	methods itExt.PaymentMethodExtService,
	pointMappings *services.PointPaymentDomainServiceImpl,
	channelMappings *services.ChannelPaymentDomainServiceImpl,
) it.PointPaymentAppService {
	return &PointPaymentApplicationServiceImpl{
		methods:         methods,
		pointMappings:   pointMappings,
		channelMappings: channelMappings,
	}
}

func (this *PointPaymentApplicationServiceImpl) ListPointPaymentMethods(
	ctx corectx.Context, query it.ListPointPaymentMethodsQuery,
) (*it.ListPointPaymentMethodsResult, error) {
	point, cErrs, err := this.resolvePoint(ctx, query.SalesPointId)
	if err != nil {
		return nil, err
	}
	if cErrs != nil {
		return &it.ListPointPaymentMethodsResult{ClientErrors: *cErrs}, nil
	}
	if cErrs := assertPointPaymentPermission(ctx, composable.PermissionRead, point); cErrs != nil {
		return &it.ListPointPaymentMethodsResult{ClientErrors: *cErrs}, nil
	}

	pointId := stringOf(point, models.SalesPointFieldId)
	channelId := stringOf(point, models.SalesPointFieldSalesChannelId)

	upstream, err := this.methods.ListPaymentMethods(ctx, itExt.ListPaymentMethodsQuery{})
	if err != nil {
		return nil, err
	}
	if upstream.ClientErrors.Count() > 0 {
		return &it.ListPointPaymentMethodsResult{ClientErrors: upstream.ClientErrors}, nil
	}

	channelMappings, err := this.channelMappings.ListMappings(ctx, channelId)
	if err != nil {
		return nil, err
	}
	permitted := make(map[string]bool, len(channelMappings))
	for _, mapping := range channelMappings {
		permitted[stringOf(mapping, models.SalesChannelPaymentRelFieldPaymentMethodId)] = true
	}

	pointMappings, err := this.pointMappings.ListMappings(ctx, pointId)
	if err != nil {
		return nil, err
	}
	profiles := make(map[string]string, len(pointMappings))
	enabled := make(map[string]bool, len(pointMappings))
	for _, mapping := range pointMappings {
		methodId := stringOf(mapping, models.SalesPointPaymentRelFieldPaymentMethodId)
		enabled[methodId] = true
		profiles[methodId] = stringOf(mapping, models.SalesPointPaymentRelFieldPaymentProfileId)
	}

	merged := make([]it.PointPaymentMethodData, 0, len(upstream.Data))
	seen := make(map[string]bool, len(upstream.Data))
	for _, method := range upstream.Data {
		seen[method.Id] = true
		row := it.PointPaymentMethodData{
			PaymentMethodId:     method.Id,
			Code:                method.Code,
			Name:                method.Name,
			PaymentProfileId:    profiles[method.Id],
			IsEnabled:           enabled[method.Id],
			IsEnabledForChannel: permitted[method.Id],
			IsUsable:            method.IsUsable,
			UnusableReason:      method.UnusableReason,
		}
		if query.EnabledOnly && !(row.IsEnabled && row.IsEnabledForChannel) {
			continue
		}
		merged = append(merged, row)
	}

	for methodId := range enabled {
		if seen[methodId] {
			continue
		}
		merged = append(merged, it.PointPaymentMethodData{
			PaymentMethodId:     methodId,
			PaymentProfileId:    profiles[methodId],
			IsEnabled:           true,
			IsEnabledForChannel: permitted[methodId],
			IsUsable:            false,
			IsStale:             true,
			UnusableReason:      "unknown_to_paymentinvoice",
		})
	}

	return &it.ListPointPaymentMethodsResult{HasData: true, Data: merged}, nil
}

func (this *PointPaymentApplicationServiceImpl) EnablePointPaymentMethod(
	ctx corectx.Context, command it.PointPaymentMethodCommand,
) (*it.PointPaymentMutationResult, error) {
	if command.PaymentMethodId == "" {
		return pointPaymentRejection("sales_point.payment_method_required",
			"enabling a payment method requires its id"), nil
	}

	point, cErrs, err := this.resolvePoint(ctx, command.SalesPointId)
	if err != nil {
		return nil, err
	}
	if cErrs != nil {
		return &it.PointPaymentMutationResult{ClientErrors: *cErrs}, nil
	}
	if cErrs := assertPointPaymentPermission(ctx, c.ActionEnablePaymentMethod, point); cErrs != nil {
		return &it.PointPaymentMutationResult{ClientErrors: *cErrs}, nil
	}

	if !models.NewSalesPointFrom(point).IsActive() {
		return pointPaymentRejection("sales_point.not_usable",
			"a suspended or archived sales point cannot have its payment methods changed"), nil
	}

	channelId := stringOf(point, models.SalesPointFieldSalesChannelId)
	permitted, err := this.channelMappings.IsEnabled(ctx, channelId, command.PaymentMethodId)
	if err != nil {
		return nil, err
	}
	if !permitted {
		return pointPaymentRejection("sales_point.payment_method_not_on_channel",
			"the sales channel of this point does not accept that payment method"), nil
	}

	usable, err := this.methods.AssertUsable(ctx, itExt.AssertUsableQuery{
		PaymentMethodId: command.PaymentMethodId,
	})
	if err != nil {
		return nil, err
	}
	if usable.ClientErrors.Count() > 0 {
		return &it.PointPaymentMutationResult{ClientErrors: usable.ClientErrors}, nil
	}

	result, err := this.pointMappings.Enable(ctx,
		stringOf(point, models.SalesPointFieldId),
		command.PaymentMethodId, command.PaymentProfileId)
	if err != nil {
		return nil, err
	}
	if result.ClientErrors.Count() > 0 {
		return &it.PointPaymentMutationResult{ClientErrors: result.ClientErrors}, nil
	}
	return &it.PointPaymentMutationResult{HasData: true}, nil
}

func (this *PointPaymentApplicationServiceImpl) DisablePointPaymentMethod(
	ctx corectx.Context, command it.PointPaymentMethodCommand,
) (*it.PointPaymentMutationResult, error) {
	if command.PaymentMethodId == "" {
		return pointPaymentRejection("sales_point.payment_method_required",
			"disabling a payment method requires its id"), nil
	}

	point, cErrs, err := this.resolvePoint(ctx, command.SalesPointId)
	if err != nil {
		return nil, err
	}
	if cErrs != nil {
		return &it.PointPaymentMutationResult{ClientErrors: *cErrs}, nil
	}
	if cErrs := assertPointPaymentPermission(ctx, c.ActionDisablePaymentMethod, point); cErrs != nil {
		return &it.PointPaymentMutationResult{ClientErrors: *cErrs}, nil
	}

	result, err := this.pointMappings.Disable(ctx,
		stringOf(point, models.SalesPointFieldId), command.PaymentMethodId)
	if err != nil {
		return nil, err
	}
	if result.ClientErrors.Count() > 0 {
		return &it.PointPaymentMutationResult{ClientErrors: result.ClientErrors}, nil
	}
	return &it.PointPaymentMutationResult{HasData: true}, nil
}

func (this *PointPaymentApplicationServiceImpl) resolvePoint(
	ctx corectx.Context, salesPointId string,
) (dmodel.DynamicFields, *ft.ClientErrors, error) {
	if salesPointId == "" {
		return nil, pointNotFound("a sales point id is required"), nil
	}

	engineRepo, err := services.RepositoryFor(models.SalesPointSchemaName)
	if err != nil {
		return nil, nil, err
	}

	found, err := engineRepo.FindByKeys(ctx, dmodel.DynamicFields{
		models.SalesPointFieldId: salesPointId,
	})
	if err != nil {
		return nil, nil, err
	}
	if found == nil || !found.HasData {
		return nil, pointNotFound("no sales point with id '" + salesPointId + "'"), nil
	}
	return found.Data, nil, nil
}

func pointNotFound(message string) *ft.ClientErrors {
	vErrs := ft.NewClientErrors()
	vErrs.Append(*ft.NewBusinessViolation(
		models.SalesPointSchemaName, "sales_point.not_found", message))
	return vErrs
}

func pointPaymentRejection(key, message string) *it.PointPaymentMutationResult {
	vErrs := ft.NewClientErrors()
	vErrs.Append(*ft.NewBusinessViolation(models.SalesPointSchemaName, key, message))
	return &it.PointPaymentMutationResult{ClientErrors: *vErrs}
}
