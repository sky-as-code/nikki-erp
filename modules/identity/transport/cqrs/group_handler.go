package cqrs

import (
	"context"

	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
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

func (this *GroupHandler) AddRemoveUsers(ctx context.Context, packet *cqrs.RequestPacket[it.AddRemoveUsersCommand]) (*cqrs.Reply[it.AddRemoveUsersResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.GroupSvc.AddRemoveUsers)
}

func (this *GroupHandler) CreateGroup(ctx context.Context, packet *cqrs.RequestPacket[it.CreateGroupCommand]) (*cqrs.Reply[it.CreateGroupResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.GroupSvc.CreateGroup)
}

func (this *GroupHandler) UpdateGroup(ctx context.Context, packet *cqrs.RequestPacket[it.UpdateGroupCommand]) (*cqrs.Reply[it.UpdateGroupResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.GroupSvc.UpdateGroup)
}

func (this *GroupHandler) DeleteGroup(ctx context.Context, packet *cqrs.RequestPacket[it.DeleteGroupCommand]) (*cqrs.Reply[it.DeleteGroupResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.GroupSvc.DeleteGroup)
}

func (this *GroupHandler) GetGroupById(ctx context.Context, packet *cqrs.RequestPacket[it.GetGroupByIdQuery]) (*cqrs.Reply[it.GetGroupByIdResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.GroupSvc.GetGroupById)
}

func (this *GroupHandler) SearchGroups(ctx context.Context, packet *cqrs.RequestPacket[it.SearchGroupsQuery]) (*cqrs.Reply[it.SearchGroupsResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.GroupSvc.SearchGroups)
}

func (this *GroupHandler) GroupExists(ctx context.Context, packet *cqrs.RequestPacket[it.GroupExistsCommand]) (*cqrs.Reply[it.GroupExistsResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.GroupSvc.Exist)
}
