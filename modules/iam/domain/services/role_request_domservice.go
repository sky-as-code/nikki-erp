package services

import (
	"github.com/sky-as-code/nikki-erp/common/safe"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	domain "github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	itRr "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/rolerequest"
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
	ctx corectx.Context, cmd itRr.CreateRoleRequestCommand, options ...corecrud.ServiceCreateOptions[*domain.RoleRequest],
) (*itRr.CreateRoleRequestResult, error) {
	opts := safe.GetOptional(options, corecrud.ServiceCreateOptions[*domain.RoleRequest]{})
	return corecrud.Create(ctx, corecrud.CreateParam[domain.RoleRequest, *domain.RoleRequest]{
		Action:                 "create grant request",
		BaseRepoGetter:         this.roleRequestRepo,
		Data:                   cmd,
		AfterValidationSuccess: opts.AfterValidationSuccess,
	})
}

func (this *RoleRequestDomainServiceImpl) DeleteRoleRequest(
	ctx corectx.Context, cmd itRr.DeleteRoleRequestCommand, options ...corecrud.ServiceDeleteOptions,
) (*itRr.DeleteRoleRequestResult, error) {
	opts := safe.GetOptional(options, corecrud.ServiceDeleteOptions{})
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{
		Action:                 "delete grant request",
		DbRepoGetter:           this.roleRequestRepo,
		Cmd:                    dyn.DeleteOneCommand(cmd),
		AfterValidationSuccess: opts.AfterValidationSuccess,
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
	ctx corectx.Context, query itRr.SearchRoleRequestsQuery, options ...corecrud.ServiceSearchOptions,
) (*itRr.SearchRoleRequestsResult, error) {
	opts := safe.GetOptional(options, corecrud.ServiceSearchOptions{})
	return corecrud.Search[domain.RoleRequest](ctx, corecrud.SearchParam{
		Action:                 "search grant requests",
		DbRepoGetter:           this.roleRequestRepo,
		Query:                  dyn.SearchQuery(query),
		AfterValidationSuccess: opts.AfterValidationSuccess,
	})
}

func (this *RoleRequestDomainServiceImpl) UpdateRoleRequest(
	ctx corectx.Context, cmd itRr.UpdateRoleRequestCommand, options ...corecrud.ServiceUpdateOptions[*domain.RoleRequest],
) (*itRr.UpdateRoleRequestResult, error) {
	opts := safe.GetOptional(options, corecrud.ServiceUpdateOptions[*domain.RoleRequest]{})
	return corecrud.Update(ctx, corecrud.UpdateParam[domain.RoleRequest, *domain.RoleRequest]{
		Action:                 "update grant request",
		DbRepoGetter:           this.roleRequestRepo,
		Data:                   cmd,
		AfterValidationSuccess: opts.AfterValidationSuccess,
	})
}
