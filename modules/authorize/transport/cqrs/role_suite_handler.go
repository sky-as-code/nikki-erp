package cqrs

import (
	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/role_suite"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

func NewRoleSuiteHandler(roleSuiteSvc it.RoleSuiteService, logger logging.LoggerService) *RoleSuiteHandler {
	return &RoleSuiteHandler{
		Logger:       logger,
		RoleSuiteSvc: roleSuiteSvc,
	}
}

type RoleSuiteHandler struct {
	Logger       logging.LoggerService
	RoleSuiteSvc it.RoleSuiteService
}

// func (this *ResourceHandler) CreateResource(ctx context.Context, packet *cqrs.RequestPacket[it.CreateResourceCommand]) (*cqrs.Reply[it.CreateResourceResult], error) {
// 	cmd := packet.Request()
// 	result, err := this.ResourceSvc.CreateResource(ctx, *cmd)
// 	ft.PanicOnErr(err)

// 	reply := &cqrs.Reply[it.CreateResourceResult]{
// 		Result: *result,
// 	}
// 	return reply, nil
// }

// func (this *ResourceHandler) UpdateResource(ctx context.Context, packet *cqrs.RequestPacket[it.UpdateResourceCommand]) (*cqrs.Reply[it.UpdateResourceResult], error) {
// 	cmd := packet.Request()
// 	result, err := this.ResourceSvc.UpdateResource(ctx, *cmd)
// 	ft.PanicOnErr(err)

// 	reply := &cqrs.Reply[it.UpdateResourceResult]{
// 		Result: *result,
// 	}
// 	return reply, nil
// }

// func (this *ResourceHandler) GetResourceByName(ctx context.Context, packet *cqrs.RequestPacket[it.GetResourceByNameCommand]) (*cqrs.Reply[it.GetResourceByNameResult], error) {
// 	cmd := packet.Request()
// 	result, err := this.ResourceSvc.GetResourceByName(ctx, *cmd)
// 	ft.PanicOnErr(err)

// 	reply := &cqrs.Reply[it.GetResourceByNameResult]{
// 		Result: *result,
// 	}
// 	return reply, nil
// }

// func (this *ResourceHandler) SearchResources(ctx context.Context, packet *cqrs.RequestPacket[it.SearchResourcesCommand]) (*cqrs.Reply[it.SearchResourcesResult], error) {
// 	cmd := packet.Request()
// 	result, err := this.ResourceSvc.SearchResources(ctx, *cmd)
// 	ft.PanicOnErr(err)

// 	reply := &cqrs.Reply[it.SearchResourcesResult]{
// 		Result: *result,
// 	}
// 	return reply, nil
// }
