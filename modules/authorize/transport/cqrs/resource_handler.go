package cqrs

import (
	"context"

	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"

	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/resource"
)

func NewResourceHandler(resourceSvc it.ResourceService, logger logging.LoggerService) *ResourceHandler {
	return &ResourceHandler{
		ResourceSvc: resourceSvc,
	}
}

type ResourceHandler struct {
	ResourceSvc it.ResourceService
}

func (this *ResourceHandler) CreateResource(ctx context.Context, packet *cqrs.RequestPacket[it.CreateResourceCommand]) (*cqrs.Reply[it.CreateResourceResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.ResourceSvc.CreateResource)
}

func (this *ResourceHandler) UpdateResource(ctx context.Context, packet *cqrs.RequestPacket[it.UpdateResourceCommand]) (*cqrs.Reply[it.UpdateResourceResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.ResourceSvc.UpdateResource)
}

func (this *ResourceHandler) GetResourceByName(ctx context.Context, packet *cqrs.RequestPacket[it.GetResourceByNameQuery]) (*cqrs.Reply[it.GetResourceByNameResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.ResourceSvc.GetResourceByName)
}

func (this *ResourceHandler) SearchResources(ctx context.Context, packet *cqrs.RequestPacket[it.SearchResourcesQuery]) (*cqrs.Reply[it.SearchResourcesResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.ResourceSvc.SearchResources)
}
