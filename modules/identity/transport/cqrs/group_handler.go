package cqrs

import (
	"context"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
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

func (this *GroupHandler) CreateGroup(ctx context.Context, packet *cqrs.RequestPacket[it.CreateGroupCommand]) (*cqrs.Reply[it.CreateGroupResult], error) {
	cmd := packet.Request()
	result, err := this.GroupSvc.CreateGroup(ctx, *cmd)
	ft.PanicOnErr(err)

	reply := &cqrs.Reply[it.CreateGroupResult]{
		Result: *result,
	}
	return reply, nil
}

func (this *GroupHandler) UpdateGroup(ctx context.Context, packet *cqrs.RequestPacket[it.UpdateGroupCommand]) (*cqrs.Reply[it.UpdateGroupResult], error) {
	cmd := packet.Request()
	result, err := this.GroupSvc.UpdateGroup(ctx, *cmd)
	ft.PanicOnErr(err)

	reply := &cqrs.Reply[it.UpdateGroupResult]{
		Result: *result,
	}
	return reply, nil
}

func (this *GroupHandler) DeleteGroup(ctx context.Context, packet *cqrs.RequestPacket[it.DeleteGroupCommand]) (*cqrs.Reply[it.DeleteGroupResult], error) {
	cmd := packet.Request()
	result, err := this.GroupSvc.DeleteGroup(ctx, *cmd)
	ft.PanicOnErr(err)

	return &cqrs.Reply[it.DeleteGroupResult]{
		Result: *result,
	}, nil
}

func (this *GroupHandler) GetGroupById(ctx context.Context, packet *cqrs.RequestPacket[it.GetGroupByIdQuery]) (*cqrs.Reply[it.GetGroupByIdResult], error) {
	cmd := packet.Request()
	result, err := this.GroupSvc.GetGroupById(ctx, *cmd)
	ft.PanicOnErr(err)

	reply := &cqrs.Reply[it.GetGroupByIdResult]{
		Result: *result,
	}
	return reply, nil
}

func (this *GroupHandler) SearchGroups(ctx context.Context, packet *cqrs.RequestPacket[it.SearchGroupsQuery]) (*cqrs.Reply[it.SearchGroupsResult], error) {
	cmd := packet.Request()
	result, err := this.GroupSvc.SearchGroups(ctx, *cmd)
	if err != nil {
		return nil, err
	}

	reply := &cqrs.Reply[it.SearchGroupsResult]{
		Result: *result,
	}
	return reply, nil
}

func (this *GroupHandler) AddRemoveUsers(ctx context.Context, packet *cqrs.RequestPacket[it.AddRemoveUsersCommand]) (*cqrs.Reply[it.AddRemoveUsersResult], error) {
	cmd := packet.Request()
	result, err := this.GroupSvc.AddRemoveUsers(ctx, *cmd)
	ft.PanicOnErr(err)

	return &cqrs.Reply[it.AddRemoveUsersResult]{
		Result: *result,
	}, nil
}
