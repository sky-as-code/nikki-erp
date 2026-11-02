package services

import (
	"go.bryk.io/pkg/errors"
	"go.uber.org/dig"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	domain "github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	itEnt "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/entitlement"
	itOrg "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/organization"
	itOrgUnit "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/orgunit"
	itPerm "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/permission"
	itRes "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/resource"
	itRole "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/role"
	itUser "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/user"
)

// RoleCrudDeps are what the role rules read and write beside the role itself: the entitlements
// an update's delta changes, the resources and scopes a grant is planned against, and the
// permission cache every change to a role keeps current.
type RoleCrudDeps struct {
	dig.In

	EntitlementRepo itEnt.EntitlementCrudRepository
	EntitlementSvc  itEnt.EntitlementCrudDomainService
	ResourceRepo    itRes.ResourceCrudRepository
	OrgRepo         itOrg.OrgCrudRepository
	OrgUnitRepo     itOrgUnit.OrgUnitCrudRepository
	PermRepo        itPerm.PermissionRepository
	HistoryRepo     itPerm.PermissionHistoryRepository
	// UserRepo tells a user from a service: both live in iam_users (see principalTypeScope).
	UserRepo itUser.UserCrudRepository
}

// NewRoleCrudDomainService is handed the composable default by the iam_role onion.
func NewRoleCrudDomainService(base dynamicresource.CrudDomainService, deps RoleCrudDeps) itRole.RoleCrudDomainService {
	return &RoleCrudDomainServiceImpl{
		CrudDomainService: base,
		roleRepo:          base.Repository(),
		entitlementRepo:   deps.EntitlementRepo,
		entitlementSvc:    deps.EntitlementSvc,
		resourceRepo:      deps.ResourceRepo,
		orgRepo:           deps.OrgRepo,
		orgUnitRepo:       deps.OrgUnitRepo,
		permRepo:          deps.PermRepo,
		auditor:           permissionAuditor{historyRepo: deps.HistoryRepo},
		userRepo:          deps.UserRepo,
	}
}

type RoleCrudDomainServiceImpl struct {
	dynamicresource.CrudDomainService
	roleRepo        dynamicresource.CrudRepository
	entitlementRepo itEnt.EntitlementCrudRepository
	// entitlementSvc and resourceRepo serve the entitlement delta of Update.
	entitlementSvc itEnt.EntitlementCrudDomainService
	resourceRepo   itRes.ResourceCrudRepository
	// orgRepo and orgUnitRepo resolve entitlement scope display names in DescribeRoles and check
	// that an org-unit entitlement belongs to the role's org.
	orgRepo     itOrg.OrgCrudRepository
	orgUnitRepo itOrgUnit.OrgUnitCrudRepository
	permRepo    itPerm.PermissionRepository
	auditor     permissionAuditor
	userRepo    dynamicresource.CrudRepository
}

// Create refuses a role claiming to be system-managed: those come from seeds only.
func (this *RoleCrudDomainServiceImpl) Create(
	ctx corectx.Context, cmd dynamicresource.CreateCommand, options ...dynamicresource.CreateOptions,
) (*dynamicresource.CreateResult, error) {
	opts := withRoleManagedCreateGuard(options)
	opts.ValidateExtra = dynamicresource.ChainValidateExtra(
		refuseServiceIn(this.userRepo, domain.RoleFieldOwnerUserId), opts.ValidateExtra)
	return this.CrudDomainService.Create(ctx, cmd, opts)
}

// Update applies the role's own columns and, when the body carries `entitlements`, the
// entitlement delta, in one transaction with one permission rebuild. The role row is updated
// first, so the etag check guards the delta too and an entitlement-only call still moves the etag
// and the audit columns. A refused delta rolls the role row back.
func (this *RoleCrudDomainServiceImpl) Update(
	ctx corectx.Context, cmd dynamicresource.UpdateCommand, options ...dynamicresource.UpdateOptions,
) (*dynamicresource.MutateResult, error) {
	opts := withRoleManagedUpdateGuard(options)
	opts.ValidateExtra = dynamicresource.ChainValidateExtra(
		opts.ValidateExtra, refuseServiceIn(this.userRepo, domain.RoleFieldOwnerUserId))
	options = []dynamicresource.UpdateOptions{opts}
	delta, malformed := itRole.LiftEntitlementsDelta(cmd)
	if malformed {
		return &dynamicresource.MutateResult{ClientErrors: ft.ClientErrors{*ft.NewValidationError(
			itRole.UpdateRoleFieldEntitlements, ft.ErrorKey("err_malformed_request"),
			"entitlements must be {add: [{resource_id, action_id, description}], remove: [ids], "+
				"update: [{id, description}]}",
		)}}, nil
	}
	if delta == nil || delta.IsEmpty() {
		return this.CrudDomainService.Update(ctx, cmd, options...)
	}

	var result *dynamicresource.MutateResult
	refused, err := dynamicresource.InRefusableTransaction(ctx, this.Repository(),
		func(tranxCtx corectx.Context) (*ft.ClientErrors, error) {
			var err error
			if result, err = this.CrudDomainService.Update(tranxCtx, cmd, options...); err != nil {
				return nil, err
			}
			if result.ClientErrors.Count() > 0 || !result.HasData {
				return &result.ClientErrors, nil
			}
			cErrs, err := this.applyEntitlementsDelta(tranxCtx, dynamicresource.ParamId(cmd), *delta)
			return &cErrs, err
		})
	if err != nil {
		return nil, errors.Wrap(err, "update role")
	}
	if refused != nil {
		return &dynamicresource.MutateResult{ClientErrors: *refused}, nil
	}
	return result, nil
}

// SetArchived revokes or restores the role for everyone holding it. Nothing cascades - the
// rebuild function filters archived roles out - so without the rebuild the role keeps answering
// for everyone who already holds it.
func (this *RoleCrudDomainServiceImpl) SetArchived(
	ctx corectx.Context, cmd dynamicresource.SetArchivedCommand,
) (*dynamicresource.MutateResult, error) {
	var result *dynamicresource.MutateResult
	err := dynamicresource.InTransaction(ctx, this.Repository(), func(tranxCtx corectx.Context) error {
		var err error
		result, err = this.CrudDomainService.SetArchived(tranxCtx, cmd)
		if err != nil || result.ClientErrors.Count() > 0 || !result.HasData {
			return err
		}
		return this.permRepo.RebuildUserPermissionsForRole(tranxCtx, dynamicresource.ParamId(cmd))
	})
	if err != nil {
		return nil, errors.Wrap(err, "set role archived")
	}
	return result, nil
}

// Delete refuses a private or system-managed role.
func (this *RoleCrudDomainServiceImpl) Delete(
	ctx corectx.Context, cmd dynamicresource.DeleteCommand, options ...dynamicresource.DeleteOptions,
) (*dynamicresource.MutateResult, error) {
	opts := dynamicresource.DeleteOptions{}
	if len(options) > 0 {
		opts = options[0]
	}
	opts.ValidateExtra = dynamicresource.ChainValidateExtra(
		assertRoleNotSystemManaged, assertRoleNotPrivate, opts.ValidateExtra)
	return this.CrudDomainService.Delete(ctx, cmd, opts)
}

func assertRoleNotPrivate(
	_ corectx.Context, _ *dynamicresource.DynamicEntity, found *dynamicresource.DynamicEntity, vErrs *ft.ClientErrors,
) error {
	isPrivate := found.GetFieldData().GetBool(domain.RoleFieldIsPrivate)
	if isPrivate != nil && *isPrivate {
		vErrs.Append(*ft.NewBusinessViolation(domain.RoleFieldIsPrivate,
			ft.ErrorKey("err_private_role_deletion_not_allowed", "iam"), "private role deletion is not allowed."))
	}
	return nil
}

func (this *RoleCrudDomainServiceImpl) SearchPrincipalRoles(
	ctx corectx.Context, query dynamicresource.SearchQuery, principalEdge string, principalId model.Id,
) (*dynamicresource.SearchResult, error) {
	return dynamicresource.SearchLinked(ctx, this, query, principalEdge, principalId,
		[]string{basemodel.FieldId, domain.RoleFieldName})
}
