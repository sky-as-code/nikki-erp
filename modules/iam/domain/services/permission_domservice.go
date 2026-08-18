package services

import (
	"fmt"
	"math"
	"time"

	"go.uber.org/dig"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/baserepo"
	reguard "github.com/sky-as-code/nikki-erp/modules/core/requestguard"
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	domain "github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	itOrg "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/organization"
	itOrgUnit "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/orgunit"
	itPerm "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/permission"
	itUser "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/user"
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
		Fields: []string{
			domain.UserFieldId, domain.UserFieldIsOwner, domain.UserFieldStatus, basemodel.FieldIsArchived,
			domain.UserFieldOrgUnitId,
			fmt.Sprintf("%s.%s", domain.UserEdgeOrgUnit, domain.OrgUnitFieldOrgId),
			fmt.Sprintf("%s.%s", domain.UserEdgeOrgs, domain.OrgFieldId),
		},
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
		// The subject's own memberships, not the calling module's: this query
		// answers "may THIS user do it", so the bare org/unit grants must be
		// judged against the memberships of the user being asked about.
		EvalContext: reguard.EvalContext{
			UserOrgIds:   foundUser.GetOrgIds(),
			OrgUnitId:    foundUser.GetOrgUnitId(),
			OrgUnitOrgId: foundUser.GetOrgUnitOrgId(),
		},
		IsRecordOwnedByCaller: query.IsRecordOwnedByCaller,
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

	userNode := dmodel.NewSearchNode()
	if query.UserId != nil {
		userNode.NewCondition(models.UserPermFieldUserId, dmodel.Equals, *query.UserId)
	}
	if query.UserEmail != nil {
		userNode.NewCondition(fmt.Sprintf("%s.%s", models.UserPermEdgeUser, models.UserFieldEmail), dmodel.Equals, *query.UserEmail)
	}

	// Expired grants must not reach the request context. They are filtered here
	// rather than swept on a schedule, so expiry takes effect on the next request
	// instead of on the next rebuild.
	graph := dmodel.NewSearchGraph()
	graph.And(
		*userNode,
		*dmodel.NewSearchNode().Or(
			*dmodel.NewSearchNode().NewCondition(models.UserPermFieldExpiresAt, dmodel.IsNotSet),
			*dmodel.NewSearchNode().NewCondition(models.UserPermFieldExpiresAt, dmodel.GreaterThan, time.Now()),
		),
	)

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
