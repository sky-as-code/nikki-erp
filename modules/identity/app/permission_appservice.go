package app

import (
	"fmt"

	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain/models"
	itPerm "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/permission"
	itUser "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/user"
)

func NewPermissionApplicationServiceImpl(
	permissionDomSvc itPerm.PermissionDomainService,
	userDomSvc itUser.UserDomainService,
) itPerm.PermissionAppService {
	return &PermissionApplicationServiceImpl{
		permissionDomSvc: permissionDomSvc,
		userDomSvc:       userDomSvc,
	}
}

type PermissionApplicationServiceImpl struct {
	permissionDomSvc itPerm.PermissionDomainService
	userDomSvc       itUser.UserDomainService
}

func (this *PermissionApplicationServiceImpl) IsAuthorized(ctx corectx.Context, query itPerm.IsAuthorizedQuery) (*itPerm.IsAuthorizedResult, error) {
	return this.permissionDomSvc.IsAuthorized(ctx, query)
}

func (this *PermissionApplicationServiceImpl) GetUserEntitlements(
	ctx corectx.Context, query itPerm.GetUserEntitlementsQuery,
) (*itPerm.GetUserEntitlementsResult, error) {
	sanitized, cErrs := query.GetSchema().ValidateStruct(query)
	if cErrs.Count() > 0 {
		return &itPerm.GetUserEntitlementsResult{
			ClientErrors: cErrs,
		}, nil
	}
	query = *sanitized.(*itPerm.GetUserEntitlementsQuery)

	resUser, err := this.getEnabledUser(ctx, query.UserEmail, query.UserId)
	if err != nil {
		return nil, err
	}

	if resUser == nil {
		return &itPerm.GetUserEntitlementsResult{
			HasData: false,
		}, nil
	}

	resEnt, err := this.permissionDomSvc.ListAllUserPermissions(ctx, itPerm.ListAllUserPermissionsQuery(query))
	if err != nil {
		return nil, err
	}
	if resEnt.ClientErrors.Count() > 0 {
		return &itPerm.GetUserEntitlementsResult{
			ClientErrors: resEnt.ClientErrors,
		}, nil
	}

	resUser.Entitlements = array.Map(resEnt.Data, func(item models.UserPermission) string {
		return item.MustGetEntExpression()
	})
	return &itPerm.GetUserEntitlementsResult{
		Data:    *resUser,
		HasData: true,
	}, nil
}

func (this *PermissionApplicationServiceImpl) getEnabledUser(ctx corectx.Context, userEmail *string, userId *model.Id) (*itPerm.GetUserEntitlementsResultData, error) {
	result, err := this.userDomSvc.GetEnabledUser(ctx, itUser.GetUserQuery{
		Email: userEmail,
		Id:    userId,
		Fields: []string{
			models.UserFieldId,
			models.UserFieldAvatarUrl,
			models.UserFieldDisplayName,
			models.UserFieldEmail,
			models.UserFieldIsOwner,
			models.UserFieldOrgUnitId,
			fmt.Sprintf("%s.%s", models.UserEdgeOrgUnit, models.OrgUnitFieldOrgId),
			fmt.Sprintf("%s.%s", models.UserEdgeOrgs, models.OrgFieldId),
		},
	})
	if err != nil {
		return nil, err
	}
	if result.ClientErrors.Count() > 0 || !result.HasData {
		return nil, errors.Wrap(result.ClientErrors.ToError(), "getEnabledUser")
	}

	return &itPerm.GetUserEntitlementsResultData{
		IsOwner:    result.Data.IsOwner(),
		UserId:     result.Data.MustGetId(),
		UserOrgIds: result.Data.GetOrgIds(),
		OrgUnitId:  result.Data.GetOrgUnitId(),
		User:       result.Data.GetFieldData(),
		// OrgUnitOrgId: result.Data.GetOrgUnitOrgId(),
	}, nil
}
