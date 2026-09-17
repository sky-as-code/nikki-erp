package services

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
	itChannel "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/channel"
	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
)

func NewPointPaymentMethodExtService(
	points *PointPaymentDomainServiceImpl,
	channels *ChannelPaymentDomainServiceImpl,
	methods itExt.PaymentMethodExtService,
) itChannel.PointPaymentMethodExtService {
	return &pointPaymentMethodExtServiceImpl{points: points, channels: channels, methods: methods}
}

type pointPaymentMethodExtServiceImpl struct {
	points   *PointPaymentDomainServiceImpl
	channels *ChannelPaymentDomainServiceImpl
	methods  itExt.PaymentMethodExtService
}

var _ itChannel.PointPaymentMethodExtService = (*pointPaymentMethodExtServiceImpl)(nil)

func (this *pointPaymentMethodExtServiceImpl) PayableMethodsOfPoint(
	ctx corectx.Context, query itChannel.PayableMethodsQuery,
) (*itChannel.PayableMethodsResult, error) {
	empty := &itChannel.PayableMethodsResult{Methods: []itChannel.PayableMethod{}, HasData: true}
	if query.SalesPointId == "" {
		return empty, nil
	}

	point, err := loadRecord(ctx, models.SalesPointSchemaName,
		models.SalesPointFieldId, query.SalesPointId)
	if err != nil {
		return nil, err
	}
	if point == nil {
		return empty, nil
	}
	if !models.NewSalesPointFrom(point).IsActive() {
		return empty, nil
	}

	channelId := stringOf(point, models.SalesPointFieldSalesChannelId)
	if channelId == "" {
		return empty, nil
	}

	channel, err := loadRecord(ctx, models.SalesChannelSchemaName,
		models.SalesChannelFieldId, channelId)
	if err != nil {
		return nil, err
	}
	if channel == nil || !models.NewSalesChannelFrom(channel).IsActive() {
		return empty, nil
	}

	channelMappings, err := this.channels.ListMappings(ctx, channelId)
	if err != nil {
		return nil, err
	}
	permitted := make(map[string]bool, len(channelMappings))
	for _, mapping := range channelMappings {
		methodId := stringOf(mapping, models.SalesChannelPaymentRelFieldPaymentMethodId)
		if methodId != "" {
			permitted[methodId] = true
		}
	}
	if len(permitted) == 0 {
		return empty, nil
	}

	pointMappings, err := this.points.ListMappings(ctx, query.SalesPointId)
	if err != nil {
		return nil, err
	}

	accepted := make(map[string]bool, len(pointMappings))
	order := make([]string, 0, len(pointMappings))
	for _, mapping := range pointMappings {
		if stringOf(mapping, basemodel.FieldOrgId) != stringOf(point, basemodel.FieldOrgId) {
			continue
		}
		methodId := stringOf(mapping, models.SalesPointPaymentRelFieldPaymentMethodId)
		if methodId == "" || accepted[methodId] || !permitted[methodId] {
			continue
		}
		accepted[methodId] = true
		order = append(order, methodId)
	}
	if len(order) == 0 {
		return empty, nil
	}

	if this.methods == nil {
		return empty, nil
	}
	upstream, err := this.methods.ListPaymentMethods(ctx, itExt.ListPaymentMethodsQuery{})
	if err != nil {
		return nil, err
	}
	if upstream == nil || upstream.ClientErrors.Count() > 0 {
		return empty, nil
	}

	named := make(map[string]itExt.PaymentMethodData, len(upstream.Data))
	for _, method := range upstream.Data {
		named[method.Id] = method
	}

	payable := make([]itChannel.PayableMethod, 0, len(order))
	for _, methodId := range order {
		method, known := named[methodId]
		if !known {
			continue
		}
		payable = append(payable, itChannel.PayableMethod{
			Id:   method.Id,
			Code: method.Code,
			Name: method.Name,
		})
	}

	return &itChannel.PayableMethodsResult{Methods: payable, HasData: true}, nil
}
