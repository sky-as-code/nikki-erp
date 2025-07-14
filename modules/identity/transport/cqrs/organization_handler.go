package cqrs

import (
	"context"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	itOrg "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/organization"
	itUser "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/user"
)

func NewOrganizationHandler(orgSvc itOrg.OrganizationService, logger logging.LoggerService) *OrganizationHandler {
	return &OrganizationHandler{
		Logger: logger,
		OrgSvc: orgSvc,
	}
}

type OrganizationHandler struct {
	Logger logging.LoggerService
	OrgSvc itOrg.OrganizationService
}

func (this *OrganizationHandler) CreateOrganization(ctx context.Context, packet *cqrs.RequestPacket[itOrg.CreateOrganizationCommand]) (*cqrs.Reply[itOrg.CreateOrganizationResult], error) {
	cmd := packet.Request()
	result, err := this.OrgSvc.CreateOrganization(ctx, *cmd)
	ft.PanicOnErr(err)

	reply := &cqrs.Reply[itOrg.CreateOrganizationResult]{
		Result: *result,
	}
	return reply, nil
}

func (this *OrganizationHandler) UpdateOrganization(ctx context.Context, packet *cqrs.RequestPacket[itOrg.UpdateOrganizationCommand]) (*cqrs.Reply[itOrg.UpdateOrganizationResult], error) {
	cmd := packet.Request()
	result, err := this.OrgSvc.UpdateOrganization(ctx, *cmd)
	ft.PanicOnErr(err)

	reply := &cqrs.Reply[itOrg.UpdateOrganizationResult]{
		Result: *result,
	}
	return reply, nil
}

func (this *OrganizationHandler) DeleteOrganization(ctx context.Context, packet *cqrs.RequestPacket[itOrg.DeleteOrganizationCommand]) (*cqrs.Reply[itOrg.DeleteOrganizationResult], error) {
	cmd := packet.Request()
	result, err := this.OrgSvc.DeleteOrganization(ctx, *cmd)
	ft.PanicOnErr(err)

	return &cqrs.Reply[itOrg.DeleteOrganizationResult]{
		Result: *result,
	}, nil
}

func (this *OrganizationHandler) GetOrganizationBySlug(ctx context.Context, packet *cqrs.RequestPacket[itOrg.GetOrganizationBySlugQuery]) (*cqrs.Reply[itOrg.GetOrganizationBySlugResult], error) {
	cmd := packet.Request()
	result, err := this.OrgSvc.GetOrganizationBySlug(ctx, *cmd)
	ft.PanicOnErr(err)

	reply := &cqrs.Reply[itOrg.GetOrganizationBySlugResult]{
		Result: *result,
	}
	return reply, nil
}

func (this *OrganizationHandler) SearchOrganizations(ctx context.Context, packet *cqrs.RequestPacket[itOrg.SearchOrganizationsQuery]) (*cqrs.Reply[itOrg.SearchOrganizationsResult], error) {
	cmd := packet.Request()
	result, err := this.OrgSvc.SearchOrganizations(ctx, *cmd)
	if err != nil {
		return nil, err
	}

	reply := &cqrs.Reply[itOrg.SearchOrganizationsResult]{
		Result: *result,
	}
	return reply, nil
}

func (this *OrganizationHandler) ListOrgStatuses(ctx context.Context, packet *cqrs.RequestPacket[itOrg.ListOrgStatusesQuery]) (*cqrs.Reply[itUser.ListIdentStatusesResult], error) {
	cmd := packet.Request()
	result, err := this.OrgSvc.ListOrgStatuses(ctx, *cmd)
	if err != nil {
		return nil, err
	}

	reply := &cqrs.Reply[itUser.ListIdentStatusesResult]{
		Result: *result,
	}
	return reply, nil
}
