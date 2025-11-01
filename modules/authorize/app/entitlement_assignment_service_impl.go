package app

import (
	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"

	"github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	itEntitlement "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/entitlement"
	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/entitlement_assignment"
	itPermissionHistory "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/permission_history"
	itResource "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/resource"
	itHierarchy "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/hierarchy"
	itOrg "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/organization"
)

func NewEntitlementAssignmentServiceImpl(
	cqrsBus cqrs.CqrsBus,
	entitlementService itEntitlement.EntitlementService,
	entitlementAssignmentRepo it.EntitlementAssignmentRepository,
	permissionHistoryRepo itPermissionHistory.PermissionHistoryRepository,
	resourceService itResource.ResourceService,
) it.EntitlementAssignmentService {
	return &EntitlementAssignmentServiceImpl{
		cqrsBus:                   cqrsBus,
		entitlementService:        entitlementService,
		entitlementAssignmentRepo: entitlementAssignmentRepo,
		permissionHistoryRepo:     permissionHistoryRepo,
		resourceService:           resourceService,
	}
}

type EntitlementAssignmentServiceImpl struct {
	cqrsBus                   cqrs.CqrsBus
	entitlementService        itEntitlement.EntitlementService
	entitlementAssignmentRepo it.EntitlementAssignmentRepository
	permissionHistoryRepo     itPermissionHistory.PermissionHistoryRepository
	resourceService           itResource.ResourceService
}

func (this *EntitlementAssignmentServiceImpl) CreateEntitlementAssignment(ctx crud.Context, cmd it.CreateEntitlementAssignmentCommand) (result *it.CreateEntitlementAssignmentResult, err error) {
	result, err = crud.Create(ctx, crud.CreateParam[*domain.EntitlementAssignment, it.CreateEntitlementAssignmentCommand, it.CreateEntitlementAssignmentResult]{
		Action:              "create entitlement assignment",
		Command:             cmd,
		AssertBusinessRules: this.assertBusinessRuleCreateEntitlementAssignment,
		RepoCreate:          this.entitlementAssignmentRepo.Create,
		SetDefault:          this.setEntitlementAssignmentDefaults,
		Sanitize:            func(assignment *domain.EntitlementAssignment) {},
		ToFailureResult: func(vErrs *fault.ValidationErrors) *it.CreateEntitlementAssignmentResult {
			return &it.CreateEntitlementAssignmentResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.EntitlementAssignment) *it.CreateEntitlementAssignmentResult {
			return &it.CreateEntitlementAssignmentResult{
				Data:    model,
				HasData: model != nil,
			}
		},
	})

	return result, err
}

func (this *EntitlementAssignmentServiceImpl) FindAllBySubject(ctx crud.Context, query it.GetAllEntitlementAssignmentBySubjectQuery) (result *it.GetAllEntitlementAssignmentBySubjectResult, err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "get entitlement assignment by subject"); e != nil {
			err = e
		}
	}()

	vErrs := query.Validate()
	if vErrs.Count() > 0 {
		return &it.GetAllEntitlementAssignmentBySubjectResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	entitlementAssignments, err := this.entitlementAssignmentRepo.FindAllBySubject(ctx, it.GetAllEntitlementAssignmentBySubjectQuery{
		SubjectType: query.SubjectType,
		SubjectRef:  query.SubjectRef,
	})
	fault.PanicOnErr(err)

	return &it.GetAllEntitlementAssignmentBySubjectResult{
		Data: entitlementAssignments,
	}, nil
}

func (this *EntitlementAssignmentServiceImpl) DeleteHardAssignment(ctx crud.Context, cmd it.DeleteEntitlementAssignmentByIdCommand) (*it.DeleteEntitlementAssignmentByIdResult, error) {
	// Not implement IncludeTransaction yet (wait new code base)
	//
	//

	return crud.DeleteHard(ctx, crud.DeleteHardParam[*domain.EntitlementAssignment, it.DeleteEntitlementAssignmentByIdCommand, it.DeleteEntitlementAssignmentByIdResult]{
		Action:              "delete entitlement assignment",
		Command:             cmd,
		AssertExists:        this.assertEntitlementAssignmentExistsById,
		AssertBusinessRules: this.assertBusinessRuleDeleteEntitlementAssignment,
		RepoDelete: func(ctx crud.Context, model *domain.EntitlementAssignment) (int, error) {
			return this.entitlementAssignmentRepo.DeleteHard(ctx, it.DeleteEntitlementAssignmentByIdCommand{Id: *model.Id})
		},
		ToFailureResult: func(vErrs *fault.ValidationErrors) *it.DeleteEntitlementAssignmentByIdResult {
			return &it.DeleteEntitlementAssignmentByIdResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.EntitlementAssignment, deletedCount int) *it.DeleteEntitlementAssignmentByIdResult {
			return crud.NewSuccessDeletionResult(*model.Id, &deletedCount)
		},
	})
}

func (this *EntitlementAssignmentServiceImpl) DeleteByEntitlementId(ctx crud.Context, cmd it.DeleteEntitlementAssignmentByEntitlementIdCommand) (*it.DeleteEntitlementAssignmentByEntitlementIdResult, error) {
	// Not implement IncludeTransaction yet (wait new code base)
	//
	//

	return crud.DeleteHard(ctx, crud.DeleteHardParam[*domain.EntitlementAssignment, it.DeleteEntitlementAssignmentByEntitlementIdCommand, it.DeleteEntitlementAssignmentByEntitlementIdResult]{
		Action:              "delete entitlement assignment",
		Command:             cmd,
		AssertExists:        nil,
		AssertBusinessRules: this.assertBusinessRuleDeleteEntitlementAssignmentByEntitlementId,
		RepoDelete: func(ctx crud.Context, model *domain.EntitlementAssignment) (int, error) {
			return this.entitlementAssignmentRepo.DeleteHardByEntitlementId(
				ctx,
				it.DeleteEntitlementAssignmentByEntitlementIdCommand{EntitlementId: cmd.EntitlementId},
			)
		},
		ToFailureResult: func(vErrs *fault.ValidationErrors) *it.DeleteEntitlementAssignmentByEntitlementIdResult {
			return &it.DeleteEntitlementAssignmentByEntitlementIdResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.EntitlementAssignment, deletedCount int) *it.DeleteEntitlementAssignmentByEntitlementIdResult {
			return crud.NewSuccessDeletionResult("", &deletedCount)
		},
	})
}

func (this *EntitlementAssignmentServiceImpl) assertEntitlementAssignmentExistsById(ctx crud.Context, entitlementAssignment *domain.EntitlementAssignment, vErrs *fault.ValidationErrors) (dbEntitlementAssignment *domain.EntitlementAssignment, err error) {
	dbEntitlementAssignment, err = this.entitlementAssignmentRepo.FindById(ctx, it.FindByIdParam{Id: *entitlementAssignment.Id})
	fault.PanicOnErr(err)

	if dbEntitlementAssignment == nil {
		vErrs.AppendNotFound("id", "entitlement assignment")
	}
	return dbEntitlementAssignment, err
}

func (this *EntitlementAssignmentServiceImpl) enableFieldsInPermissionHistory(ctx crud.Context, assignment *domain.EntitlementAssignment) error {
	return this.permissionHistoryRepo.EnableField(
		ctx,
		itPermissionHistory.EnableFieldCommand{
			AssignmentId: assignment.Id,
			ResolvedExpr: *assignment.ResolvedExpr,
		},
	)
}

func (this *EntitlementAssignmentServiceImpl) assertBusinessRuleDeleteEntitlementAssignment(ctx crud.Context, cmd it.DeleteEntitlementAssignmentByIdCommand, assignment *domain.EntitlementAssignment, vErrs *fault.ValidationErrors) error {
	err := this.enableFieldsInPermissionHistory(ctx, assignment)
	fault.PanicOnErr(err)

	return nil
}

func (this *EntitlementAssignmentServiceImpl) assertBusinessRuleDeleteEntitlementAssignmentByEntitlementId(ctx crud.Context, cmd it.DeleteEntitlementAssignmentByEntitlementIdCommand, assignment *domain.EntitlementAssignment, vErrs *fault.ValidationErrors) error {
	assignments, err := this.entitlementAssignmentRepo.FindAllByEntitlementId(
		ctx,
		it.FindAllByEntitlementIdParam{EntitlementId: cmd.EntitlementId},
	)
	fault.PanicOnErr(err)

	for _, assignment := range assignments {
		err = this.enableFieldsInPermissionHistory(ctx, &assignment)
		fault.PanicOnErr(err)
	}

	return nil
}

func (this *EntitlementAssignmentServiceImpl) setEntitlementAssignmentDefaults(assignment *domain.EntitlementAssignment) {
	assignment.SetDefaults()
}

// assertBusinessRuleCreateEntitlementAssignment validates assignment creation rules
// This ensures proper scope isolation and prevents cross-org data leaks
//
// Responsibility: Validate scopeRef against resource's ScopeType
// - For domain-level resources: scopeRef must be nil
// - For org-level resources: scopeRef must be a valid orgId
// - For hierarchy-level resources: scopeRef must be a valid hierarchyLevelId
// - For private resources: scopeRef must be provided (not fully implemented yet)
//
//
//	When a global entitlement (ent.OrgId=nil) is added to an org-level role,
//	the scopeRef validation here ensures the scope stays within the org boundary
func (this *EntitlementAssignmentServiceImpl) assertBusinessRuleCreateEntitlementAssignment(ctx crud.Context, assignment *domain.EntitlementAssignment, vErrs *fault.ValidationErrors) error {
	entRes, err := this.entitlementService.GetEntitlementById(ctx, itEntitlement.GetEntitlementByIdQuery{
		Id: *assignment.EntitlementId,
	})
	fault.PanicOnErr(err)

	if entRes.ClientError != nil {
		return entRes.ClientError
	}

	ent := entRes.Data
	if ent == nil {
		vErrs.AppendNotFound("entitlement_id", "entitlement id")
		return nil
	}

	// Global entitlement (no resource): scopeRef must be nil
	if ent.ResourceId == nil {
		if assignment.ScopeRef != nil {
			vErrs.AppendNotAllow("scope_ref", "scopeRef of global entitlement")
			return nil
		}

		this.assertEntitlementAssignmentUnique(ctx, assignment, vErrs)
		return nil
	}

	resourceRes, err := this.resourceService.GetResourceById(ctx, itResource.GetResourceByIdQuery{Id: *ent.ResourceId})
	fault.PanicOnErr(err)

	if resourceRes.ClientError != nil {
		return resourceRes.ClientError
	}

	resource := resourceRes.Data
	if resource == nil {
		vErrs.AppendNotFound("resource_id", *ent.ResourceId)
		return nil
	}

	switch *resource.ScopeType {
	case domain.ResourceScopeTypeDomain:
		if assignment.ScopeRef != nil {
			vErrs.AppendNotAllow("scope_ref", "scopeRef of domain-level resource")
			return nil
		}

	case domain.ResourceScopeTypeOrg:
		if assignment.ScopeRef == nil {
			vErrs.AppendNotAllow("scope_ref", "scopeRef of org-level resource")
			return nil
		}

		existCmd := &itOrg.ExistsOrgByIdCommand{
			Id: *assignment.ScopeRef,
		}
		existRes := itOrg.ExistsOrgByIdResult{}
		err = this.cqrsBus.Request(ctx, *existCmd, &existRes)
		fault.PanicOnErr(err)

		if existRes.ClientError != nil {
			vErrs.MergeClientError(existRes.ClientError)
			return nil
		}

		if !existRes.Data {
			vErrs.AppendNotFound("scope_ref", *assignment.ScopeRef)
			return nil
		}

	case domain.ResourceScopeTypeHierarchy:
		if assignment.ScopeRef == nil {
			vErrs.AppendNotAllow("scope_ref", "scopeRef of hierarchy-level resource")
			return nil
		}

		existCmd := &itHierarchy.ExistsHierarchyLevelByIdQuery{
			Id: *assignment.ScopeRef,
		}
		existRes := itHierarchy.ExistsHierarchyLevelByIdResult{}
		err := this.cqrsBus.Request(ctx, *existCmd, &existRes)
		fault.PanicOnErr(err)

		if existRes.ClientError != nil {
			vErrs.MergeClientError(existRes.ClientError)
			return nil
		}

		if !existRes.Data {
			vErrs.AppendNotFound("scope_ref", "scope ref")
			return nil
		}
		return nil

	case domain.ResourceScopeTypePrivate:
		if assignment.ScopeRef == nil {
			vErrs.AppendNotAllow("scope_ref", "scopeRef of private resource")
			return nil
		}
		// Temporary not implement yet

	default:
		vErrs.AppendNotAllow("resource.scope_type", "scope type")
		return nil
	}

	this.assertEntitlementAssignmentUnique(ctx, assignment, vErrs)

	return nil
}

func (this *EntitlementAssignmentServiceImpl) assertEntitlementAssignmentUnique(ctx crud.Context, assignment *domain.EntitlementAssignment, vErrs *fault.ValidationErrors) {
	exists, err := this.entitlementAssignmentRepo.FindByFilter(ctx, it.FindByFilterParam{
		SubjectType:   *assignment.SubjectType,
		SubjectRef:    *assignment.SubjectRef,
		EntitlementId: *assignment.EntitlementId,
		ScopeRef:      assignment.ScopeRef,
	})
	fault.PanicOnErr(err)

	if exists != nil {
		vErrs.AppendAlreadyExists("subject_type", "subject type")
	}
}
