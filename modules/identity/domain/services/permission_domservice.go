package services

import (
	"fmt"
	"math"

	"go.uber.org/dig"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/baserepo"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain/models"
	domain "github.com/sky-as-code/nikki-erp/modules/identity/domain/models"
	itOrg "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/organization"
	itOrgUnit "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/orgunit"
	itPerm "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/permission"
	itUser "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/user"
)

type NewPermissionServiceParam struct {
	dig.In

	OrgRepo            itOrg.OrganizationRepository
	OrgUnitRepo        itOrgUnit.OrgUnitRepository
	UserPermissionRepo itPerm.PermissionRepository
	UserRepo           itUser.UserRepository
}

func NewPermissionDomainServiceImpl(param NewPermissionServiceParam) itPerm.PermissionDomainService {
	return &PermissionDomainServiceImpl{
		orgRepo:        param.OrgRepo,
		orgUnitRepo:    param.OrgUnitRepo,
		permissionRepo: param.UserPermissionRepo,
		userRepo:       param.UserRepo,
	}
}

type PermissionDomainServiceImpl struct {
	orgRepo        itOrg.OrganizationRepository
	orgUnitRepo    itOrgUnit.OrgUnitRepository
	permissionRepo itPerm.PermissionRepository
	userRepo       itUser.UserRepository
}

// Implements PermissionDomainService interface
func (this *PermissionDomainServiceImpl) IsAuthorized(
	ctx corectx.Context, query itPerm.IsAuthorizedQuery,
) (*itPerm.IsAuthorizedResult, error) {
	sanitized, cErrs := query.GetSchema().ValidateStruct(query)
	if cErrs.Count() > 0 {
		return nil, errors.Wrap(cErrs.ToError(), "IsAuthorized")
	}
	query = *sanitized.(*itPerm.IsAuthorizedQuery)
	// No need to check action_code and resource_code existence.

	filter := dmodel.DynamicFields{}
	if query.UserId != nil {
		filter[domain.UserFieldId] = *query.UserId
	}
	if query.UserEmail != nil {
		filter[domain.UserFieldEmail] = *query.UserEmail
	}

	resUser, err := this.userRepo.GetOne(ctx, dyn.RepoGetOneParam{
		Filter: filter,
		Fields: []string{domain.UserFieldId, domain.UserFieldIsOwner, domain.UserFieldStatus, basemodel.FieldIsArchived},
	})
	if err != nil {
		return nil, err
	}
	if resUser.ClientErrors.Count() > 0 {
		return nil, errors.Wrap(resUser.ClientErrors.ToError(), "IsAuthorized")
	}
	if !resUser.HasData || !resUser.Data.IsActive() {
		return &itPerm.IsAuthorizedResult{
			Data:    false,
			HasData: true,
		}, nil
	}

	foundUser := resUser.Data

	if foundUser.IsOwner() {
		return &itPerm.IsAuthorizedResult{
			Data:    true,
			HasData: true,
		}, nil
	}

	resMat, err := this.permissionRepo.MatchPermisions(ctx, itPerm.RepoMatchUserPermParam{
		UserId:       *foundUser.GetId(),
		ResourceCode: query.ResourceCode,
		ActionCode:   query.ActionCode,
		Scope:        query.Scope,
		ScopeId:      query.ScopeId,
	})
	if err != nil {
		return nil, err
	}
	if resMat.ClientErrors.Count() > 0 {
		return nil, errors.Wrap(resMat.ClientErrors.ToError(), "IsAuthorized")
	}
	// TODO: Return `foundUser`
	return &itPerm.IsAuthorizedResult{
		Data:    resMat.HasData,
		HasData: true,
	}, nil
}

func (this *PermissionDomainServiceImpl) ListAllUserPermissions(
	ctx corectx.Context, query itPerm.ListAllUserPermissionsQuery,
) (*itPerm.ListAllUserPermissionsResult, error) {
	sanitized, cErrs := query.GetSchema().ValidateStruct(query)
	if cErrs.Count() > 0 {
		return &itPerm.ListAllUserPermissionsResult{
			ClientErrors: cErrs,
		}, nil
	}
	query = *sanitized.(*itPerm.ListAllUserPermissionsQuery)

	graph := &dmodel.SearchGraph{}
	if query.UserId != nil {
		graph.NewCondition(models.UserPermFieldUserId, dmodel.Equals, *query.UserId)
	}
	if query.UserEmail != nil {
		graph.NewCondition(fmt.Sprintf("%s.%s", models.UserPermEdgeUser, models.UserFieldEmail), dmodel.Equals, *query.UserEmail)
	}

	result, err := baserepo.Search[models.UserPermission](ctx, this.permissionRepo.GetBaseRepo(), dyn.RepoSearchParam{
		Graph:  graph,
		Fields: []string{domain.UserPermFieldEntExpression},
		Page:   0,
		Size:   math.MaxInt32,
	})
	if err != nil {
		return nil, err
	}
	if result.ClientErrors.Count() > 0 {
		return &itPerm.ListAllUserPermissionsResult{
			ClientErrors: result.ClientErrors,
		}, nil
	}
	return &itPerm.ListAllUserPermissionsResult{
		Data:    result.Data.Items,
		HasData: true,
	}, nil
}
