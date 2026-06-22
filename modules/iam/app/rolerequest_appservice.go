package app

import (
	"fmt"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	c "github.com/sky-as-code/nikki-erp/modules/iam/constants"
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	itExt "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/external"
	it "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/rolerequest"
)

func NewRoleRequestApplicationServiceImpl(
	roleRequestSvc it.RoleRequestDomainService,
	roleRequestRepo it.RoleRequestRepository,
	userPrefSvc itExt.UserPreferenceUiDomainService,
) it.RoleRequestAppService {
	return &RoleRequestApplicationServiceImpl{
		roleRequestSvc:  roleRequestSvc,
		roleRequestRepo: roleRequestRepo,
		userPrefSvc:     userPrefSvc,
	}
}

type RoleRequestApplicationServiceImpl struct {
	roleRequestSvc  it.RoleRequestDomainService
	roleRequestRepo it.RoleRequestRepository
	userPrefSvc     itExt.UserPreferenceUiDomainService
}

func (this *RoleRequestApplicationServiceImpl) CreateRoleRequest(ctx corectx.Context, cmd it.CreateRoleRequestCommand) (*it.CreateRoleRequestResult, error) {
	if cErr := assertPermission(ctx, "create", c.ResourceIamGrantRequest, c.ResourceScopeDomain); cErr != nil {
		return &it.CreateRoleRequestResult{ClientErrors: *cErr}, nil
	}
	return this.roleRequestSvc.CreateRoleRequest(ctx, cmd)
}

func (this *RoleRequestApplicationServiceImpl) DeleteRoleRequest(ctx corectx.Context, cmd it.DeleteRoleRequestCommand) (*it.DeleteRoleRequestResult, error) {
	if cErr := assertPermission(ctx, "delete", c.ResourceIamGrantRequest, c.ResourceScopeDomain); cErr != nil {
		return &it.DeleteRoleRequestResult{ClientErrors: *cErr}, nil
	}
	return this.roleRequestSvc.DeleteRoleRequest(ctx, cmd)
}

func (this *RoleRequestApplicationServiceImpl) GetRoleRequest(ctx corectx.Context, query it.GetRoleRequestQuery) (*it.GetRoleRequestResult, error) {
	if cErr := assertPermission(ctx, "read", c.ResourceIamGrantRequest, c.ResourceScopeDomain); cErr != nil {
		return &it.GetRoleRequestResult{ClientErrors: *cErr}, nil
	}
	return corecrud.UiGetOne(ctx, corecrud.UiGetOneParam[models.RoleRequest, *models.RoleRequest]{
		Action: "get role request",
		Schema: this.roleRequestRepo.GetBaseRepo().Schema(),
		GetOneFn: func() (*dyn.OpResult[models.RoleRequest], error) {
			return this.roleRequestSvc.GetRoleRequest(ctx, query)
		},
	})
}

func (this *RoleRequestApplicationServiceImpl) RoleRequestExists(ctx corectx.Context, query it.RoleRequestExistsQuery) (*it.RoleRequestExistsResult, error) {
	if cErr := assertPermission(ctx, "read", c.ResourceIamGrantRequest, c.ResourceScopeDomain); cErr != nil {
		return &it.RoleRequestExistsResult{ClientErrors: *cErr}, nil
	}
	return this.roleRequestSvc.RoleRequestExists(ctx, query)
}

func (this *RoleRequestApplicationServiceImpl) SearchRoleRequests(ctx corectx.Context, query it.SearchRoleRequestsQuery) (*it.SearchRoleRequestsResult, error) {
	if cErr := assertPermission(ctx, "read", c.ResourceIamGrantRequest, c.ResourceScopeDomain); cErr != nil {
		return &it.SearchRoleRequestsResult{ClientErrors: *cErr}, nil
	}
	return corecrud.UiSearch(ctx, corecrud.UiSearchParam[models.RoleRequest, *models.RoleRequest]{
		Action:        "search role requests",
		FieldResolver: this.userPrefSvc.(corecrud.FieldsResolver),
		Schema:        this.roleRequestRepo.GetBaseRepo().Schema(),
		DefaultFields: []string{
			fmt.Sprintf("%s.%s", models.RoleReqEdgeRole, models.RoleFieldName),
			fmt.Sprintf("%s.%s", models.RoleReqEdgeRequestor, models.UserFieldDisplayName),
			fmt.Sprintf("%s.%s", models.RoleReqEdgeReceiverUser, models.UserFieldDisplayName),
			fmt.Sprintf("%s.%s", models.RoleReqEdgeReceiverGroup, models.GroupFieldName),
			models.RoleReqFieldStatus,
			models.RoleReqFieldType,
		},
		SearchFn: func(fn corecrud.AfterValidationSuccessFn[dyn.SearchQuery]) (*dyn.OpResult[dyn.PagedResultData[models.RoleRequest]], error) {
			return this.roleRequestSvc.SearchRoleRequests(ctx, query, corecrud.ServiceSearchOptions{
				AfterValidationSuccess: fn,
			})
		},
	})
}

func (this *RoleRequestApplicationServiceImpl) UpdateRoleRequest(ctx corectx.Context, cmd it.UpdateRoleRequestCommand) (*it.UpdateRoleRequestResult, error) {
	if cErr := assertPermission(ctx, "update", c.ResourceIamGrantRequest, c.ResourceScopeDomain); cErr != nil {
		return &it.UpdateRoleRequestResult{ClientErrors: *cErr}, nil
	}
	return this.roleRequestSvc.UpdateRoleRequest(ctx, cmd)
}
