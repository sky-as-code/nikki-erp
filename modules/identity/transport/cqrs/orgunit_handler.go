package cqrs

import (
	// "context"

	// "github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	// c "github.com/sky-as-code/nikki-erp/modules/identity/constants"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/orgunit"
)

func NewOrgUnitHandler(orgunitSvc it.OrgUnitService, logger logging.LoggerService) *OrgUnitHandler {
	return &OrgUnitHandler{
		Logger:     logger,
		OrgUnitSvc: orgunitSvc,
	}
}

type OrgUnitHandler struct {
	Logger     logging.LoggerService
	OrgUnitSvc it.OrgUnitService
}

// func (this *OrgUnitHandler) CreateOrgUnit(ctx context.Context, packet *cqrs.RequestPacket[it.CreateOrgUnitCommand]) (
// 	*cqrs.Reply[it.CreateOrgUnitResult], error,
// ) {
// 	return cqrs.HandlePacket2(ctx, string(c.IdentityModuleName), packet, this.OrgUnitSvc.CreateOrgUnit)
// }

// func (this *OrgUnitHandler) UpdateOrgUnit(ctx context.Context, packet *cqrs.RequestPacket[it.UpdateOrgUnitCommand]) (
// 	*cqrs.Reply[it.UpdateOrgUnitResult], error,
// ) {
// 	return cqrs.HandlePacket2(ctx, string(c.IdentityModuleName), packet, this.OrgUnitSvc.UpdateOrgUnit)
// }

// func (this *OrgUnitHandler) DeleteOrgUnit(ctx context.Context, packet *cqrs.RequestPacket[it.DeleteOrgUnitCommand]) (
// 	*cqrs.Reply[it.DeleteOrgUnitResult], error,
// ) {
// 	return cqrs.HandlePacket2(ctx, string(c.IdentityModuleName), packet, this.OrgUnitSvc.DeleteOrgUnit)
// }

// func (this *OrgUnitHandler) GetOrgUnit(ctx context.Context, packet *cqrs.RequestPacket[it.GetOrgUnitQuery]) (
// 	*cqrs.Reply[it.GetOrgUnitResult], error,
// ) {
// 	return cqrs.HandlePacket2(ctx, string(c.IdentityModuleName), packet, this.OrgUnitSvc.GetOrgUnit)
// }

// func (this *OrgUnitHandler) ManageOrgUnitUsers(ctx context.Context, packet *cqrs.RequestPacket[it.ManageOrgUnitUsersCommand]) (
// 	*cqrs.Reply[it.ManageOrgUnitUsersResult], error,
// ) {
// 	return cqrs.HandlePacket2(ctx, string(c.IdentityModuleName), packet, this.OrgUnitSvc.ManageOrgUnitUsers)
// }

// func (this *OrgUnitHandler) SearchOrgUnits(ctx context.Context, packet *cqrs.RequestPacket[it.SearchOrgUnitsQuery]) (
// 	*cqrs.Reply[it.SearchOrgUnitsResult], error,
// ) {
// 	return cqrs.HandlePacket2(ctx, string(c.IdentityModuleName), packet, this.OrgUnitSvc.SearchOrgUnits)
// }

// func (this *OrgUnitHandler) OrgUnitExists(ctx context.Context, packet *cqrs.RequestPacket[it.OrgUnitExistsQuery]) (
// 	*cqrs.Reply[it.OrgUnitExistsResult], error,
// ) {
// 	return cqrs.HandlePacket2(ctx, string(c.IdentityModuleName), packet, this.OrgUnitSvc.OrgUnitExists)
// }
