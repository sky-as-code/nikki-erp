package app

import (
	"go.uber.org/dig"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
	itOrg "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/organization"
	itOrgUnit "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/orgunit"
	itPerm "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/permission"
	itUser "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/user"
)

type NewPermissionServiceParam struct {
	dig.In

	OrgRepo            itOrg.OrganizationRepository
	OrgUnitRepo        itOrgUnit.OrgUnitRepository
	UserPermissionRepo itPerm.UserPermissionRepository
	UserRepo           itUser.UserRepository
}

func NewPermissionServiceImpl(param NewPermissionServiceParam) itPerm.PermissionService {
	return &PermissionServiceImpl{
		orgRepo:            param.OrgRepo,
		orgUnitRepo:        param.OrgUnitRepo,
		userPermissionRepo: param.UserPermissionRepo,
		userRepo:           param.UserRepo,
	}
}

type PermissionServiceImpl struct {
	orgRepo            itOrg.OrganizationRepository
	orgUnitRepo        itOrgUnit.OrgUnitRepository
	userPermissionRepo itPerm.UserPermissionRepository
	userRepo           itUser.UserRepository
}

// Implements PermissionService interface
func (this *PermissionServiceImpl) IsAuthorized(
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

	resActor, err := this.userRepo.GetOne(ctx, dyn.RepoGetOneParam{Filter: filter})
	if err != nil {
		return nil, err
	}
	if resActor.ClientErrors.Count() > 0 {
		return nil, errors.Wrap(resActor.ClientErrors.ToError(), "IsAuthorized")
	}
	if !resActor.HasData || !resActor.Data.IsActive() {
		return &itPerm.IsAuthorizedResult{
			Data:    false,
			HasData: true,
		}, nil
	}

	resMat, err := this.userPermissionRepo.MatchPermisions(ctx, itPerm.RepoMatchUserPermParam{
		UserId:       *resActor.Data.GetId(),
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
	return &itPerm.IsAuthorizedResult{
		Data:    resMat.HasData,
		HasData: true,
	}, nil
}

// Implements PermissionService interface
func (this *PermissionServiceImpl) CheckPermissions(
	ctx corectx.Context, query itPerm.CheckPermissionsQuery,
) (*itPerm.CheckPermissionsResult, error) {
	sanitized, cErrs := query.GetSchema().ValidateStruct(query)
	if cErrs.Count() > 0 {
		return &itPerm.CheckPermissionsResult{ClientErrors: cErrs}, nil
	}
	query = *sanitized.(*itPerm.CheckPermissionsQuery)
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
	})
	if err != nil {
		return nil, err
	}
	if resUser.ClientErrors.Count() > 0 {
		return nil, errors.Wrap(resUser.ClientErrors.ToError(), "CheckPermissions")
	}

	if !resUser.HasData {
		return &itPerm.CheckPermissionsResult{
			Data: itPerm.CheckPermissionsResultData{
				IsAuthorized: false,
				RejectReason: "User not found",
			},
			HasData: false,
		}, nil
	}

	if resUser.Data.MustGetStatus() == domain.UserStatusActive {
		return &itPerm.CheckPermissionsResult{
			Data: itPerm.CheckPermissionsResultData{
				IsAuthorized: false,
				RejectReason: "User is not in active status",
			},
			HasData: true,
		}, nil
	}

	resMat, err := this.userPermissionRepo.MatchPermisions(ctx, itPerm.RepoMatchUserPermParam{
		UserId:       *resUser.Data.GetId(),
		ResourceCode: query.ResourceCode,
		ActionCode:   query.ActionCode,
		Scope:        query.Scope,
		ScopeId:      query.ScopeId,
	})
	if err != nil {
		return nil, err
	}
	if resMat.ClientErrors.Count() > 0 {
		return nil, errors.Wrap(resMat.ClientErrors.ToError(), "CheckPermissions")
	}
	return &itPerm.CheckPermissionsResult{
		Data: itPerm.CheckPermissionsResultData{
			IsAuthorized: resMat.HasData,
			RejectReason: "",
			Permissions:  resMat.Data,
		},
		HasData: true,
	}, nil
}
