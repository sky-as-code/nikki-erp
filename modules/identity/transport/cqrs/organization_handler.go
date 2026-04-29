package cqrs

import (
	// "context"

	// "github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	// c "github.com/sky-as-code/nikki-erp/modules/identity/constants"
	itOrg "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/organization"
)

func NewOrganizationHandler(orgSvc itOrg.OrganizationDomainService, logger logging.LoggerService) *OrganizationHandler {
	return &OrganizationHandler{
		Logger: logger,
		OrgSvc: orgSvc,
	}
}

type OrganizationHandler struct {
	Logger logging.LoggerService
	OrgSvc itOrg.OrganizationDomainService
}

// func (this *OrganizationHandler) CreateOrg(ctx context.Context, packet *cqrs.RequestPacket[itOrg.CreateOrgCommand]) (
// 	*cqrs.Reply[itOrg.CreateOrgResult], error,
// ) {
// 	return cqrs.HandlePacket2(ctx, string(c.IdentityModuleName), packet, this.OrgSvc.CreateOrg)
// }

// func (this *OrganizationHandler) DeleteOrg(ctx context.Context, packet *cqrs.RequestPacket[itOrg.DeleteOrgCommand]) (
// 	*cqrs.Reply[itOrg.DeleteOrgResult], error,
// ) {
// 	return cqrs.HandlePacket2(ctx, string(c.IdentityModuleName), packet, this.OrgSvc.DeleteOrg)
// }

// func (this *OrganizationHandler) GetOrg(ctx context.Context, packet *cqrs.RequestPacket[itOrg.GetOrgQuery]) (
// 	*cqrs.Reply[itOrg.GetOrgResult], error,
// ) {
// 	return cqrs.HandlePacket2(ctx, string(c.IdentityModuleName), packet, this.OrgSvc.GetOrg)
// }

// func (this *OrganizationHandler) SearchOrgs(ctx context.Context, packet *cqrs.RequestPacket[itOrg.SearchOrgsQuery]) (
// 	*cqrs.Reply[itOrg.SearchOrgsResult], error,
// ) {
// 	return cqrs.HandlePacket2(ctx, string(c.IdentityModuleName), packet, this.OrgSvc.SearchOrgs)
// }

// func (this *OrganizationHandler) OrgExists(ctx context.Context, packet *cqrs.RequestPacket[itOrg.OrgExistsQuery]) (
// 	*cqrs.Reply[itOrg.OrgExistsResult], error,
// ) {
// 	return cqrs.HandlePacket2(ctx, string(c.IdentityModuleName), packet, this.OrgSvc.OrgExists)
// }

// func (this *OrganizationHandler) ManageOrgUsers(ctx context.Context, packet *cqrs.RequestPacket[itOrg.ManageOrgUsersCommand]) (
// 	*cqrs.Reply[itOrg.ManageOrgUsersResult], error,
// ) {
// 	return cqrs.HandlePacket2(ctx, string(c.IdentityModuleName), packet, this.OrgSvc.ManageOrgUsers)
// }

// func (this *OrganizationHandler) UpdateOrg(ctx context.Context, packet *cqrs.RequestPacket[itOrg.UpdateOrgCommand]) (
// 	*cqrs.Reply[itOrg.UpdateOrgResult], error,
// ) {
// 	return cqrs.HandlePacket2(ctx, string(c.IdentityModuleName), packet, this.OrgSvc.UpdateOrg)
// }
