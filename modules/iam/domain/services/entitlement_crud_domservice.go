package services

import (
	"go.bryk.io/pkg/errors"
	"go.uber.org/dig"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	domain "github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	itAct "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/action"
	itEnt "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/entitlement"
	itOrg "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/organization"
	itOrgUnit "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/orgunit"
	itPerm "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/permission"
	itRes "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/resource"
)

// EntitlementCrudDeps are the repositories the entitlement rules read and the permission cache
// they keep current.
type EntitlementCrudDeps struct {
	dig.In

	ActionRepo   itAct.ActionCrudRepository
	ResourceRepo itRes.ResourceCrudRepository
	OrgRepo      itOrg.OrgCrudRepository
	OrgUnitRepo  itOrgUnit.OrgUnitCrudRepository
	PermRepo     itPerm.PermissionRepository
	HistoryRepo  itPerm.PermissionHistoryRepository
}

// NewEntitlementCrudDomainService is handed the composable default by the iam_entitlement onion.
func NewEntitlementCrudDomainService(
	base composable.CrudDomainService, deps EntitlementCrudDeps,
) itEnt.EntitlementCrudDomainService {
	return &EntitlementCrudDomainServiceImpl{
		CrudDomainService: base,
		actionRepo:        deps.ActionRepo,
		resourceRepo:      deps.ResourceRepo,
		orgRepo:           deps.OrgRepo,
		orgUnitRepo:       deps.OrgUnitRepo,
		permRepo:          deps.PermRepo,
		auditor:           permissionAuditor{historyRepo: deps.HistoryRepo},
	}
}

type EntitlementCrudDomainServiceImpl struct {
	composable.CrudDomainService
	actionRepo   itAct.ActionCrudRepository
	resourceRepo itRes.ResourceCrudRepository
	orgRepo      itOrg.OrgCrudRepository
	orgUnitRepo  itOrgUnit.OrgUnitCrudRepository
	permRepo     itPerm.PermissionRepository
	auditor      permissionAuditor
}

func (this *EntitlementCrudDomainServiceImpl) Create(
	ctx corectx.Context, cmd itEnt.CreateEntitlementCommand, options ...composable.CreateOptions,
) (*itEnt.CreateEntitlementResult, error) {
	opts := composable.CreateOptions{}
	if len(options) > 0 {
		opts = options[0]
	}
	return this.create(ctx, cmd, opts, true)
}

func (this *EntitlementCrudDomainServiceImpl) CreateDeferRebuild(
	ctx corectx.Context, cmd itEnt.CreateEntitlementCommand,
) (*itEnt.CreateEntitlementResult, error) {
	return this.create(ctx, cmd, composable.CreateOptions{}, false)
}

// create validates the scope before schema validation and computes the expression after it:
// `expression` is auto-generated, and validation blanks every auto-generated field, so a value
// computed earlier would be discarded and the column left NULL.
func (this *EntitlementCrudDomainServiceImpl) create(
	ctx corectx.Context, cmd itEnt.CreateEntitlementCommand, opts composable.CreateOptions, rebuild bool,
) (*itEnt.CreateEntitlementResult, error) {
	opts.BeforeValidation = this.validateScopeHook(opts.BeforeValidation)
	opts.AfterValidationSuccess = this.expressionHook(opts.AfterValidationSuccess)

	var result *itEnt.CreateEntitlementResult
	err := composable.InTransaction(ctx, this.Repository(), func(tranxCtx corectx.Context) error {
		var err error
		result, err = this.CrudDomainService.Create(tranxCtx, cmd, opts)
		if err != nil || result.ClientErrors.Count() > 0 || !result.HasData {
			return err
		}
		return this.afterCreated(tranxCtx, result.Data, rebuild)
	})
	if err != nil {
		return nil, errors.Wrap(err, "create entitlement")
	}
	return result, nil
}

// afterCreated makes a new entitlement reach the role's existing holders, not just the next person
// assigned to it, and records the grant.
func (this *EntitlementCrudDomainServiceImpl) afterCreated(
	ctx corectx.Context, created dmodel.DynamicFields, rebuild bool,
) error {
	roleId := created.GetModelId(domain.EntitlementFieldRoleId)
	if roleId == nil {
		return nil
	}
	if err := this.rebuildRoleIf(ctx, rebuild, *roleId); err != nil {
		return err
	}
	return this.auditor.recordEntitlementTransition(
		ctx, domain.PermissionHistoryEffectGrant, domain.PermissionHistoryReasonEntAdded,
		roleId, created.GetModelId(domain.EntitlementFieldId), created.GetString(domain.EntitlementFieldExpression),
	)
}

func (this *EntitlementCrudDomainServiceImpl) validateScopeHook(
	next composable.BeforeValidationFn,
) composable.BeforeValidationFn {
	return func(
		ctx corectx.Context, entity *composable.DynamicEntity, vErrs *ft.ClientErrors,
	) (*composable.DynamicEntity, error) {
		if err := this.validateScope(ctx, domain.NewEntitlementFrom(entity.GetFieldData()), vErrs); err != nil {
			return entity, err
		}
		if next == nil {
			return entity, nil
		}
		return next(ctx, entity, vErrs)
	}
}

func (this *EntitlementCrudDomainServiceImpl) expressionHook(
	next composable.AfterValidationFn,
) composable.AfterValidationFn {
	return func(ctx corectx.Context, entity *composable.DynamicEntity) (*composable.DynamicEntity, error) {
		ent := domain.NewEntitlementFrom(entity.GetFieldData())
		if err := this.calculateExpression(ctx, ent, ft.NewClientErrors()); err != nil {
			return entity, err
		}
		if next == nil {
			return entity, nil
		}
		return next(ctx, entity)
	}
}

func (this *EntitlementCrudDomainServiceImpl) Update(
	ctx corectx.Context, cmd itEnt.UpdateEntitlementCommand, options ...composable.UpdateOptions,
) (*itEnt.UpdateEntitlementResult, error) {
	return this.mutate(ctx, func(tranxCtx corectx.Context) (*composable.MutateResult, error) {
		return this.CrudDomainService.Update(tranxCtx, cmd, options...)
	}, this.rebuildForEntitlementFn(composable.ParamId(cmd)))
}

func (this *EntitlementCrudDomainServiceImpl) UpdateDeferRebuild(
	ctx corectx.Context, cmd itEnt.UpdateEntitlementCommand,
) (*itEnt.UpdateEntitlementResult, error) {
	return this.CrudDomainService.Update(ctx, cmd)
}

// SetArchived rebuilds the role's holders: archiving is a revocation and unarchiving a grant, and
// neither reaches the cache on its own, because archived rows are filtered inside the rebuild.
func (this *EntitlementCrudDomainServiceImpl) SetArchived(
	ctx corectx.Context, cmd composable.SetArchivedCommand,
) (*composable.MutateResult, error) {
	return this.mutate(ctx, func(tranxCtx corectx.Context) (*composable.MutateResult, error) {
		return this.CrudDomainService.SetArchived(tranxCtx, cmd)
	}, this.rebuildForEntitlementFn(composable.ParamId(cmd)))
}

func (this *EntitlementCrudDomainServiceImpl) Delete(
	ctx corectx.Context, cmd itEnt.DeleteEntitlementCommand, options ...composable.DeleteOptions,
) (*itEnt.DeleteEntitlementResult, error) {
	return this.delete(ctx, cmd, options, true)
}

func (this *EntitlementCrudDomainServiceImpl) DeleteDeferRebuild(
	ctx corectx.Context, cmd itEnt.DeleteEntitlementCommand,
) (*itEnt.DeleteEntitlementResult, error) {
	return this.delete(ctx, cmd, nil, false)
}

// delete reads the owning role before the row goes away: afterwards there is nothing left to
// resolve it from. The cascade removes this entitlement's cached rows, but the role's holders
// still need the rest of their rows recomputed.
func (this *EntitlementCrudDomainServiceImpl) delete(
	ctx corectx.Context, cmd itEnt.DeleteEntitlementCommand, options []composable.DeleteOptions, rebuild bool,
) (*itEnt.DeleteEntitlementResult, error) {
	entId := composable.ParamId(cmd)
	var roleId *model.Id
	return this.mutate(ctx, func(tranxCtx corectx.Context) (*composable.MutateResult, error) {
		var err error
		if roleId, err = this.roleIdOfEntitlement(tranxCtx, entId); err != nil {
			return nil, err
		}
		return this.CrudDomainService.Delete(tranxCtx, cmd, options...)
	}, func(tranxCtx corectx.Context) error {
		if roleId == nil {
			return nil
		}
		if err := this.rebuildRoleIf(tranxCtx, rebuild, *roleId); err != nil {
			return err
		}
		return this.auditor.recordEntitlementTransition(
			tranxCtx, domain.PermissionHistoryEffectRevoke, domain.PermissionHistoryReasonEntDeleted, roleId, &entId, nil)
	})
}

// mutate runs a mutation and its side effect in one transaction. The side effect runs only when
// the mutation succeeded.
func (this *EntitlementCrudDomainServiceImpl) mutate(
	ctx corectx.Context,
	op func(tranxCtx corectx.Context) (*composable.MutateResult, error),
	after func(tranxCtx corectx.Context) error,
) (*composable.MutateResult, error) {
	var result *composable.MutateResult
	err := composable.InTransaction(ctx, this.Repository(), func(tranxCtx corectx.Context) error {
		var err error
		result, err = op(tranxCtx)
		if err != nil || result.ClientErrors.Count() > 0 || !result.HasData {
			return err
		}
		return after(tranxCtx)
	})
	if err != nil {
		return nil, errors.Wrap(err, "mutate entitlement")
	}
	return result, nil
}

// rebuildForEntitlementFn refreshes the cached permissions of everyone holding the role that owns
// the entitlement. Without it the cache fails in the unsafe direction: deletions cascade away
// immediately, but an addition never reaches the people who already hold the role.
func (this *EntitlementCrudDomainServiceImpl) rebuildForEntitlementFn(entId model.Id) func(corectx.Context) error {
	return func(ctx corectx.Context) error {
		roleId, err := this.roleIdOfEntitlement(ctx, entId)
		if err != nil || roleId == nil {
			return err
		}
		return this.permRepo.RebuildUserPermissionsForRole(ctx, *roleId)
	}
}

// rebuildRoleIf lets a caller that changes several entitlements of one role pay for a single
// permission rebuild at the end instead of one per entitlement.
func (this *EntitlementCrudDomainServiceImpl) rebuildRoleIf(ctx corectx.Context, rebuild bool, roleId model.Id) error {
	if !rebuild {
		return nil
	}
	return this.permRepo.RebuildUserPermissionsForRole(ctx, roleId)
}
