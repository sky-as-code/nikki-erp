package cqrs

import (
	"context"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/resource"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

func NewResourceHandler(resourceSvc it.ResourceService, logger logging.LoggerService) *ResourceHandler {
	return &ResourceHandler{
		Logger:      logger,
		ResourceSvc: resourceSvc,
	}
}

type ResourceHandler struct {
	Logger      logging.LoggerService
	ResourceSvc it.ResourceService
}

func (this *ResourceHandler) CreateResource(ctx context.Context, packet *cqrs.RequestPacket[it.CreateResourceCommand]) (*cqrs.Reply[it.CreateResourceResult], error) {
	cmd := packet.Request()
	result, err := this.ResourceSvc.CreateResource(ctx, *cmd)
	ft.PanicOnErr(err)

	reply := &cqrs.Reply[it.CreateResourceResult]{
		Result: *result,
	}
	return reply, nil
}

func (this *ResourceHandler) UpdateResource(ctx context.Context, packet *cqrs.RequestPacket[it.UpdateResourceCommand]) (*cqrs.Reply[it.UpdateResourceResult], error) {
	cmd := packet.Request()
	result, err := this.ResourceSvc.UpdateResource(ctx, *cmd)
	ft.PanicOnErr(err)

	reply := &cqrs.Reply[it.UpdateResourceResult]{
		Result: *result,
	}
	return reply, nil
}

func (this *ResourceHandler) GetResourceByName(ctx context.Context, packet *cqrs.RequestPacket[it.GetResourceByNameQuery]) (*cqrs.Reply[it.GetResourceByNameResult], error) {
	cmd := packet.Request()
	result, err := this.ResourceSvc.GetResourceByName(ctx, *cmd)
	ft.PanicOnErr(err)

	reply := &cqrs.Reply[it.GetResourceByNameResult]{
		Result: *result,
	}
	return reply, nil
}

func (this *ResourceHandler) SearchResources(ctx context.Context, packet *cqrs.RequestPacket[it.SearchResourcesQuery]) (*cqrs.Reply[it.SearchResourcesResult], error) {
	cmd := packet.Request()
	result, err := this.ResourceSvc.SearchResources(ctx, *cmd)
	ft.PanicOnErr(err)

	reply := &cqrs.Reply[it.SearchResourcesResult]{
		Result: *result,
	}
	return reply, nil
}
