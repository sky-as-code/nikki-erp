package app

import (
	"fmt"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
	itAct "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/action"
	itEnt "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/entitlement"
	itOrg "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/organization"
	itOrgUnit "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/orgunit"
	itRes "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/resource"
)

func NewEntitlementServiceImpl(
	entitlementRepo itEnt.EntitlementRepository,
	actionRepo itAct.ActionRepository,
	resourceRepo itRes.ResourceRepository,
	orgRepo itOrg.OrganizationRepository,
	orgUnitRepo itOrgUnit.OrgUnitRepository,
	cqrsBus cqrs.CqrsBus,
) itEnt.EntitlementService {
	return &EntitlementServiceImpl{
		cqrsBus:         cqrsBus,
		entitlementRepo: entitlementRepo,
		actionRepo:      actionRepo,
		resourceRepo:    resourceRepo,
		orgUnitRepo:     orgUnitRepo,
	}
}

type EntitlementServiceImpl struct {
	cqrsBus         cqrs.CqrsBus
	entitlementRepo itEnt.EntitlementRepository
	actionRepo      itAct.ActionRepository
	resourceRepo    itRes.ResourceRepository
	orgRepo         itOrg.OrganizationRepository
	orgUnitRepo     itOrgUnit.OrgUnitRepository
}

func (this *EntitlementServiceImpl) CreateEntitlement(
	ctx corectx.Context, cmd itEnt.CreateEntitlementCommand,
) (*itEnt.CreateEntitlementResult, error) {
	return corecrud.Create(ctx, corecrud.CreateParam[domain.Entitlement, *domain.Entitlement]{
		Action:         "create entitlement",
		BaseRepoGetter: this.entitlementRepo,
		Data:           cmd,
		BeforeValidation: func(ctx corectx.Context, ent *domain.Entitlement, vErrs *ft.ClientErrors) (*domain.Entitlement, error) {
			ent.CalculateExpression()
			return ent, this.validateScope(ctx, ent, vErrs)
		},
	})
}

func (this *EntitlementServiceImpl) validateScope(
	ctx corectx.Context, ent *domain.Entitlement, cErrsTotal *ft.ClientErrors,
) error {
	resourceId := ent.GetResourceId()
	resource, err := this.fetchResourceForAction(ctx, resourceId, cErrsTotal)
	if err != nil {
		return err
	}
	if cErrsTotal.Count() > 0 {
		return nil
	}

	minScope := *resource.GetMinScope()
	maxScope := *resource.GetMaxScope()
	entScope := *ent.GetScope()
	orgId := ent.GetOrgId()
	orgUnitId := ent.GetOrgUnitId()

	cErrs, err := dyn.StartValidationFlowCopy(cErrsTotal).
		Step(func(cErrs *ft.ClientErrors) error {
			return this.checkOrgExistence(ctx, entScope, orgId, cErrs)
		}).
		Step(func(cErrs *ft.ClientErrors) error {
			return this.checkOrgUnitExistence(ctx, entScope, orgUnitId, cErrs)
		}).
		Step(func(cErrs *ft.ClientErrors) error {
			if !domain.IsResourceScopeInBounds(minScope, maxScope, entScope) {
				cErrs.Append(*ft.NewValidationError(
					domain.EntitlementFieldScope, "err_scope_out_of_bounds",
					"scope must be between the resource min_scope and max_scope (both inclusive)",
				))
			}
			return nil
		}).
		End()
	cErrsTotal.Concat(cErrs)

	return err
}

func (this *EntitlementServiceImpl) checkOrgExistence(ctx corectx.Context, entScope domain.ResourceScope, orgId *model.Id, vErrs *ft.ClientErrors) error {
	if entScope == domain.ResourceScopeOrg && orgId == nil {
		vErrs.Append(*dmodel.NewMissingFieldErr(domain.EntitlementFieldOrgId))
		return nil
	}
	if entScope == domain.ResourceScopeOrg && orgId != nil {
		org := domain.NewOrganization()
		org.SetId(orgId)
		existence, err := this.orgRepo.Exists(ctx, []domain.Organization{*org})
		if err != nil {
			return err
		}
		if len(existence.Data.Existing) == 0 {
			vErrs.Append(*ft.NewNotFoundError(domain.EntitlementFieldOrgId))
			return nil
		}
	}
	return nil
}

func (this *EntitlementServiceImpl) checkOrgUnitExistence(ctx corectx.Context, entScope domain.ResourceScope, orgUnitId *model.Id, vErrs *ft.ClientErrors) error {
	if entScope == domain.ResourceScopeOrgUnit && orgUnitId == nil {
		vErrs.Append(*dmodel.NewMissingFieldErr(domain.EntitlementFieldOrgUnitId))
		return nil
	}
	if entScope == domain.ResourceScopeOrgUnit && orgUnitId != nil {
		orgUnit := domain.NewOrganizationalUnit()
		orgUnit.SetId(orgUnitId)
		existence, err := this.orgUnitRepo.Exists(ctx, []domain.OrganizationalUnit{*orgUnit})
		if err != nil {
			return err
		}
		if len(existence.Data.Existing) == 0 {
			vErrs.Append(*ft.NewNotFoundError(domain.EntitlementFieldOrgUnitId))
			return nil
		}
	}
	return nil
}

func (this *EntitlementServiceImpl) fetchResourceForAction(
	ctx corectx.Context, resourceId *model.Id, vErrs *ft.ClientErrors,
) (*domain.Resource, error) {

	resRes, err := this.resourceRepo.GetOne(ctx, dyn.RepoGetOneParam{
		Filter: dmodel.DynamicFields{
			domain.ResourceFieldId: *resourceId,
		},
		Columns: []string{fmt.Sprintf("%s.%s", domain.ResourceEdgeActions, domain.ActionFieldId)},
	})
	if err != nil {
		return nil, err
	}
	if !resRes.HasData {
		vErrs.Append(*ft.NewNotFoundError(domain.EntitlementFieldResourceId))
		return nil, nil
	}
	if resRes.Data.GetActions() == nil {
		vErrs.Append(*ft.NewNotFoundError(domain.EntitlementFieldActionId))
		return nil, nil
	}
	r := resRes.Data
	return &r, nil
}

func (this *EntitlementServiceImpl) DeleteEntitlement(
	ctx corectx.Context, cmd itEnt.DeleteEntitlementCommand,
) (*itEnt.DeleteEntitlementResult, error) {
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{
		Action:       "delete entitlement",
		DbRepoGetter: this.entitlementRepo,
		Cmd:          dyn.DeleteOneCommand(cmd),
	})
}

func (this *EntitlementServiceImpl) EntitlementExists(
	ctx corectx.Context, query itEnt.EntitlementExistsQuery,
) (*itEnt.EntitlementExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{
		Action:       "check if entitlement exists",
		DbRepoGetter: this.entitlementRepo,
		Query:        dyn.ExistsQuery(query),
	})
}

func (this *EntitlementServiceImpl) GetEntitlement(
	ctx corectx.Context, query itEnt.GetEntitlementQuery,
) (*itEnt.GetEntitlementResult, error) {
	return corecrud.GetOne[domain.Entitlement](ctx, corecrud.GetOneParam{
		Action:       "get entitlement",
		DbRepoGetter: this.entitlementRepo,
		Query:        dyn.GetOneQuery(query),
	})
}

func (this *EntitlementServiceImpl) ManageEntitlementRoles(
	ctx corectx.Context, cmd itEnt.ManageEntitlementRolesCommand,
) (*itEnt.ManageEntitlementRolesResult, error) {
	return corecrud.ManageM2m(ctx, corecrud.ManageM2mParam{
		Action:             "manage entitlement roles",
		DbRepoGetter:       this.entitlementRepo,
		DestSchemaName:     domain.RoleSchemaName,
		SrcId:              cmd.EntitlementId,
		SrcIdFieldForError: "entitlement_id",
		AssociatedIds:      cmd.Add,
		DisassociatedIds:   cmd.Remove,
	})
}

func (this *EntitlementServiceImpl) SearchEntitlements(
	ctx corectx.Context, query itEnt.SearchEntitlementsQuery,
) (*itEnt.SearchEntitlementsResult, error) {
	return corecrud.Search[domain.Entitlement](ctx, corecrud.SearchParam{
		Action:       "search entitlements",
		DbRepoGetter: this.entitlementRepo,
		Query:        dyn.SearchQuery(query),
	})
}

func (this *EntitlementServiceImpl) SetEntitlementIsArchived(
	ctx corectx.Context, cmd itEnt.SetEntitlementIsArchivedCommand,
) (*itEnt.SetEntitlementIsArchivedResult, error) {
	return corecrud.SetIsArchived(ctx, this.entitlementRepo, dyn.SetIsArchivedCommand(cmd))
}

func (this *EntitlementServiceImpl) UpdateEntitlement(
	ctx corectx.Context, cmd itEnt.UpdateEntitlementCommand,
) (*itEnt.UpdateEntitlementResult, error) {
	return corecrud.Update(ctx, corecrud.UpdateParam[domain.Entitlement, *domain.Entitlement]{
		Action:       "update entitlement",
		DbRepoGetter: this.entitlementRepo,
		Data:         cmd,
	})
}
