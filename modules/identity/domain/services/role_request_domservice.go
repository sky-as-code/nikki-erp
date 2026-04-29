package services

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	domain "github.com/sky-as-code/nikki-erp/modules/identity/domain/models"
	itRr "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/rolerequest"
)

func NewRoleRequestDomainServiceImpl(
	roleRequestRepo itRr.RoleRequestRepository,
	cqrsBus cqrs.CqrsBus,
) itRr.RoleRequestDomainService {
	return &RoleRequestDomainServiceImpl{cqrsBus: cqrsBus, roleRequestRepo: roleRequestRepo}
}

type RoleRequestDomainServiceImpl struct {
	cqrsBus         cqrs.CqrsBus
	roleRequestRepo itRr.RoleRequestRepository
}

func (this *RoleRequestDomainServiceImpl) CreateRoleRequest(
	ctx corectx.Context, cmd itRr.CreateRoleRequestCommand,
) (*itRr.CreateRoleRequestResult, error) {
	return corecrud.Create(ctx, corecrud.CreateParam[domain.RoleRequest, *domain.RoleRequest]{
		Action:         "create grant request",
		BaseRepoGetter: this.roleRequestRepo,
		Data:           cmd,
	})
}

func (this *RoleRequestDomainServiceImpl) DeleteRoleRequest(
	ctx corectx.Context, cmd itRr.DeleteRoleRequestCommand,
) (*itRr.DeleteRoleRequestResult, error) {
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{
		Action:       "delete grant request",
		DbRepoGetter: this.roleRequestRepo,
		Cmd:          dyn.DeleteOneCommand(cmd),
	})
}

func (this *RoleRequestDomainServiceImpl) GetRoleRequest(
	ctx corectx.Context, query itRr.GetRoleRequestQuery,
) (*dyn.OpResult[domain.RoleRequest], error) {
	return corecrud.GetOne[domain.RoleRequest](ctx, corecrud.GetOneParam{
		Action:       "get grant request",
		DbRepoGetter: this.roleRequestRepo,
		Query:        dyn.GetOneQuery(query),
	})
}

func (this *RoleRequestDomainServiceImpl) RoleRequestExists(
	ctx corectx.Context, query itRr.RoleRequestExistsQuery,
) (*itRr.RoleRequestExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{
		Action:       "check if grant request exists",
		DbRepoGetter: this.roleRequestRepo,
		Query:        dyn.ExistsQuery(query),
	})
}

func (this *RoleRequestDomainServiceImpl) SearchRoleRequests(
	ctx corectx.Context, query itRr.SearchRoleRequestsQuery,
) (*itRr.SearchRoleRequestsResult, error) {
	return corecrud.Search[domain.RoleRequest](ctx, corecrud.SearchParam{
		Action:       "search grant requests",
		DbRepoGetter: this.roleRequestRepo,
		Query:        dyn.SearchQuery(query),
	})
}

func (this *RoleRequestDomainServiceImpl) UpdateRoleRequest(
	ctx corectx.Context, cmd itRr.UpdateRoleRequestCommand,
) (*itRr.UpdateRoleRequestResult, error) {
	return corecrud.Update(ctx, corecrud.UpdateParam[domain.RoleRequest, *domain.RoleRequest]{
		Action:       "update grant request",
		DbRepoGetter: this.roleRequestRepo,
		Data:         cmd,
	})
}
