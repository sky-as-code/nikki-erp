package cqrs

import (
	// "context"

	// "github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	// c "github.com/sky-as-code/nikki-erp/modules/identity/constants"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/group"
)

func NewGroupHandler(groupSvc it.GroupService, logger logging.LoggerService) *GroupHandler {
	return &GroupHandler{
		Logger:   logger,
		GroupSvc: groupSvc,
	}
}

type GroupHandler struct {
	Logger   logging.LoggerService
	GroupSvc it.GroupService
}

// func (this *GroupHandler) CreateGroup(ctx context.Context, packet *cqrs.RequestPacket[it.CreateGroupCommand]) (
// 	*cqrs.Reply[it.CreateGroupResult], error,
// ) {
// 	return cqrs.HandlePacket2(ctx, string(c.IdentityModuleName), packet, this.GroupSvc.CreateGroup)
// }

// func (this *GroupHandler) UpdateGroup(ctx context.Context, packet *cqrs.RequestPacket[it.UpdateGroupCommand]) (
// 	*cqrs.Reply[it.UpdateGroupResult], error,
// ) {
// 	return cqrs.HandlePacket2(ctx, string(c.IdentityModuleName), packet, this.GroupSvc.UpdateGroup)
// }

// func (this *GroupHandler) DeleteGroup(ctx context.Context, packet *cqrs.RequestPacket[it.DeleteGroupCommand]) (
// 	*cqrs.Reply[it.DeleteGroupResult], error,
// ) {
// 	return cqrs.HandlePacket2(ctx, string(c.IdentityModuleName), packet, this.GroupSvc.DeleteGroup)
// }

// func (this *GroupHandler) GetGroup(ctx context.Context, packet *cqrs.RequestPacket[it.GetGroupQuery]) (
// 	*cqrs.Reply[it.GetGroupResult], error,
// ) {
// 	return cqrs.HandlePacket2(ctx, string(c.IdentityModuleName), packet, this.GroupSvc.GetGroup)
// }

// func (this *GroupHandler) GroupExists(ctx context.Context, packet *cqrs.RequestPacket[it.GroupExistsQuery]) (
// 	*cqrs.Reply[it.GroupExistsResult], error,
// ) {
// 	return cqrs.HandlePacket2(ctx, string(c.IdentityModuleName), packet, this.GroupSvc.GroupExists)
// }

// func (this *GroupHandler) ManageGroupUsers(ctx context.Context, packet *cqrs.RequestPacket[it.ManageGroupUsersCommand]) (
// 	*cqrs.Reply[it.ManageGroupUsersResult], error,
// ) {
// 	return cqrs.HandlePacket2(ctx, string(c.IdentityModuleName), packet, this.GroupSvc.ManageGroupUsers)
// }

// func (this *GroupHandler) SearchGroups(ctx context.Context, packet *cqrs.RequestPacket[it.SearchGroupsQuery]) (
// 	*cqrs.Reply[it.SearchGroupsResult], error,
// ) {
// 	return cqrs.HandlePacket2(ctx, string(c.IdentityModuleName), packet, this.GroupSvc.SearchGroups)
// }
