package cqrs

import (
	"context"

	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	itOrg "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/organization"
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
	return cqrs.HandlePacket(ctx, packet, this.OrgSvc.CreateOrganization)
}

func (this *OrganizationHandler) UpdateOrganization(ctx context.Context, packet *cqrs.RequestPacket[itOrg.UpdateOrganizationCommand]) (*cqrs.Reply[itOrg.UpdateOrganizationResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.OrgSvc.UpdateOrganization)
}

func (this *OrganizationHandler) DeleteOrganization(ctx context.Context, packet *cqrs.RequestPacket[itOrg.DeleteOrganizationCommand]) (*cqrs.Reply[itOrg.DeleteOrganizationResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.OrgSvc.DeleteOrganization)
}

func (this *OrganizationHandler) GetOrganizationBySlug(ctx context.Context, packet *cqrs.RequestPacket[itOrg.GetOrganizationBySlugQuery]) (*cqrs.Reply[itOrg.GetOrganizationBySlugResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.OrgSvc.GetOrganizationBySlug)
}

func (this *OrganizationHandler) SearchOrganizations(ctx context.Context, packet *cqrs.RequestPacket[itOrg.SearchOrganizationsQuery]) (*cqrs.Reply[itOrg.SearchOrganizationsResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.OrgSvc.SearchOrganizations)
}

func (this *OrganizationHandler) ExistsOrgById(ctx context.Context, packet *cqrs.RequestPacket[itOrg.ExistsOrgByIdCommand]) (*cqrs.Reply[itOrg.ExistsOrgByIdResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.OrgSvc.ExistsOrgById)
}
