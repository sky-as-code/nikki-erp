package cqrs

import (
	"context"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/organization"
)

func NewOrganizationHandler(orgSvc it.OrganizationService, logger logging.LoggerService) *OrganizationHandler {
	return &OrganizationHandler{
		Logger: logger,
		OrgSvc: orgSvc,
	}
}

type OrganizationHandler struct {
	Logger logging.LoggerService
	OrgSvc it.OrganizationService
}

func (this *OrganizationHandler) CreateOrganization(ctx context.Context, packet *cqrs.RequestPacket[it.CreateOrganizationCommand]) (*cqrs.Reply[it.CreateOrganizationResult], error) {
	cmd := packet.Request()
	result, err := this.OrgSvc.CreateOrganization(ctx, *cmd)
	ft.PanicOnErr(err)

	reply := &cqrs.Reply[it.CreateOrganizationResult]{
		Result: *result,
	}
	return reply, nil
}

func (this *OrganizationHandler) UpdateOrganization(ctx context.Context, packet *cqrs.RequestPacket[it.UpdateOrganizationCommand]) (*cqrs.Reply[it.UpdateOrganizationResult], error) {
	cmd := packet.Request()
	result, err := this.OrgSvc.UpdateOrganization(ctx, *cmd)
	ft.PanicOnErr(err)

	reply := &cqrs.Reply[it.UpdateOrganizationResult]{
		Result: *result,
	}
	return reply, nil
}

func (this *OrganizationHandler) DeleteOrganization(ctx context.Context, packet *cqrs.RequestPacket[it.DeleteOrganizationCommand]) (*cqrs.Reply[it.DeleteOrganizationResult], error) {
	cmd := packet.Request()
	result, err := this.OrgSvc.DeleteOrganization(ctx, *cmd)
	ft.PanicOnErr(err)

	return &cqrs.Reply[it.DeleteOrganizationResult]{
		Result: *result,
	}, nil
}

func (this *OrganizationHandler) GetOrganizationBySlug(ctx context.Context, packet *cqrs.RequestPacket[it.GetOrganizationBySlugQuery]) (*cqrs.Reply[it.GetOrganizationBySlugResult], error) {
	cmd := packet.Request()
	result, err := this.OrgSvc.GetOrganizationBySlug(ctx, *cmd)
	ft.PanicOnErr(err)

	reply := &cqrs.Reply[it.GetOrganizationBySlugResult]{
		Result: *result,
	}
	return reply, nil
}

func (this *OrganizationHandler) SearchOrganizations(ctx context.Context, packet *cqrs.RequestPacket[it.SearchOrganizationsQuery]) (*cqrs.Reply[it.SearchOrganizationsResult], error) {
	cmd := packet.Request()
	result, err := this.OrgSvc.SearchOrganizations(ctx, *cmd)
	if err != nil {
		return nil, err
	}

	reply := &cqrs.Reply[it.SearchOrganizationsResult]{
		Result: *result,
	}
	return reply, nil
}
