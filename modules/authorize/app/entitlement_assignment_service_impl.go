package app

// import (
// 	"fmt"

// 	"github.com/sky-as-code/nikki-erp/common/fault"
// 	"github.com/sky-as-code/nikki-erp/common/model"
// 	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
// 	"github.com/sky-as-code/nikki-erp/modules/core/crud"

// 	"github.com/sky-as-code/nikki-erp/modules/authorize/domain"
// 	itAction "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/action"
// 	itEntitlement "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/entitlement"
// 	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/entitlement_assignment"
// 	itPermissionHistory "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/permission_history"
// 	itResource "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/resource"
// 	itHierarchy "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/hierarchy"
// 	itOrg "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/organization"
// )

// func NewEntitlementAssignmentServiceImpl(
// 	cqrsBus cqrs.CqrsBus,
// 	entitlementService itEntitlement.EntitlementDomainService,
// 	entitlementAssignmentRepo it.EntitlementAssignmentRepository,
// 	permissionHistoryRepo itPermissionHistory.PermissionHistoryRepository,
// 	resourceService itResource.ResourceDomainService,
// ) it.EntitlementAssignmentService {
// 	return &EntitlementAssignmentServiceImpl{
// 		cqrsBus:                   cqrsBus,
// 		entitlementService:        entitlementService,
// 		entitlementAssignmentRepo: entitlementAssignmentRepo,
// 		permissionHistoryRepo:     permissionHistoryRepo,
// 		resourceService:           resourceService,
// 	}
// }

// type EntitlementAssignmentServiceImpl struct {
// 	cqrsBus                   cqrs.CqrsBus
// 	entitlementService        itEntitlement.EntitlementDomainService
// 	entitlementAssignmentRepo it.EntitlementAssignmentRepository
// 	permissionHistoryRepo     itPermissionHistory.PermissionHistoryRepository
// 	resourceService           itResource.ResourceDomainService
// }

// func (this *EntitlementAssignmentServiceImpl) CreateEntitlementAssignment(ctx crud.Context, cmd it.CreateEntitlementAssignmentCommand) (result *it.CreateEntitlementAssignmentResult, err error) {
// 	result, err = crud.Create(ctx, crud.CreateParam[*domain.EntitlementGrant, it.CreateEntitlementAssignmentCommand, it.CreateEntitlementAssignmentResult]{
// 		Action:              "create entitlement assignment",
// 		Command:             cmd,
// 		AssertBusinessRules: this.assertBusinessRuleCreateEntitlementAssignment,
// 		RepoCreate:          this.entitlementAssignmentRepo.Create,
// 		SetDefault:          this.setEntitlementAssignmentDefaults,
// 		Sanitize:            func(assignment *domain.EntitlementGrant) {},
// 		ToFailureResult: func(vErrs *fault.ValidationErrors) *it.CreateEntitlementAssignmentResult {
// 			return &it.CreateEntitlementAssignmentResult{
// 				ClientError: vErrs.ToClientError(),
// 			}
// 		},
// 		ToSuccessResult: func(model *domain.EntitlementGrant) *it.CreateEntitlementAssignmentResult {
// 			return &it.CreateEntitlementAssignmentResult{
// 				Data:    model,
// 				HasData: model != nil,
// 			}
// 		},
// 	})

// 	return result, err
// }

// func (this *EntitlementAssignmentServiceImpl) FindAllBySubject(ctx crud.Context, query it.GetAllEntitlementAssignmentBySubjectQuery) (result *it.GetAllEntitlementAssignmentBySubjectResult, err error) {
// 	defer func() {
// 		if e := fault.RecoverPanicFailedTo(recover(), "get entitlement assignment by subject"); e != nil {
// 			err = e
// 		}
// 	}()

// 	vErrs := query.Validate()
// 	if vErrs.Count() > 0 {
// 		return &it.GetAllEntitlementAssignmentBySubjectResult{
// 			ClientError: vErrs.ToClientError(),
// 		}, nil
// 	}

// 	entitlementAssignments, err := this.entitlementAssignmentRepo.FindAllBySubject(ctx, it.GetAllEntitlementAssignmentBySubjectQuery{
// 		SubjectType: query.SubjectType,
// 		SubjectRef:  query.SubjectRef,
// 	})
// 	fault.PanicOnErr(err)

// 	return &it.GetAllEntitlementAssignmentBySubjectResult{
// 		Data: entitlementAssignments,
// 	}, nil
// }

// func (this *EntitlementAssignmentServiceImpl) DeleteHardAssignment(ctx crud.Context, cmd it.DeleteEntitlementAssignmentByIdCommand) (*it.DeleteEntitlementAssignmentByIdResult, error) {
// 	// Not implement IncludeTransaction yet (wait new code base)
// 	//
// 	//

// 	return crud.DeleteHard(ctx, crud.DeleteHardParam[*domain.EntitlementGrant, it.DeleteEntitlementAssignmentByIdCommand, it.DeleteEntitlementAssignmentByIdResult]{
// 		Action:              "delete entitlement assignment",
// 		Command:             cmd,
// 		AssertExists:        this.assertEntitlementAssignmentExistsById,
// 		AssertBusinessRules: this.assertBusinessRuleDeleteEntitlementAssignment,
// 		RepoDelete: func(ctx crud.Context, row *domain.EntitlementGrant) (int, error) {
// 			return this.entitlementAssignmentRepo.DeleteHard(ctx, it.DeleteEntitlementAssignmentByIdCommand{Id: *row.GetId()})
// 		},
// 		ToFailureResult: func(vErrs *fault.ValidationErrors) *it.DeleteEntitlementAssignmentByIdResult {
// 			return &it.DeleteEntitlementAssignmentByIdResult{
// 				ClientError: vErrs.ToClientError(),
// 			}
// 		},
// 		ToSuccessResult: func(row *domain.EntitlementGrant, deletedCount int) *it.DeleteEntitlementAssignmentByIdResult {
// 			return crud.NewSuccessDeletionResult(*row.GetId(), &deletedCount)
// 		},
// 	})
// }

// func (this *EntitlementAssignmentServiceImpl) DeleteByEntitlementId(ctx crud.Context, cmd it.DeleteEntitlementAssignmentByEntitlementIdCommand) (*it.DeleteEntitlementAssignmentByEntitlementIdResult, error) {
// 	// Not implement IncludeTransaction yet (wait new code base)
// 	//
// 	//

// 	return crud.DeleteHard(ctx, crud.DeleteHardParam[*domain.EntitlementGrant, it.DeleteEntitlementAssignmentByEntitlementIdCommand, it.DeleteEntitlementAssignmentByEntitlementIdResult]{
// 		Action:              "delete entitlement assignment",
// 		Command:             cmd,
// 		AssertExists:        nil,
// 		AssertBusinessRules: this.assertBusinessRuleDeleteEntitlementAssignmentByEntitlementId,
// 		RepoDelete: func(ctx crud.Context, model *domain.EntitlementGrant) (int, error) {
// 			return this.entitlementAssignmentRepo.DeleteHardByEntitlementId(
// 				ctx,
// 				it.DeleteEntitlementAssignmentByEntitlementIdCommand{EntitlementId: cmd.EntitlementId},
// 			)
// 		},
// 		ToFailureResult: func(vErrs *fault.ValidationErrors) *it.DeleteEntitlementAssignmentByEntitlementIdResult {
// 			return &it.DeleteEntitlementAssignmentByEntitlementIdResult{
// 				ClientError: vErrs.ToClientError(),
// 			}
// 		},
// 		ToSuccessResult: func(model *domain.EntitlementGrant, deletedCount int) *it.DeleteEntitlementAssignmentByEntitlementIdResult {
// 			return crud.NewSuccessDeletionResult("", &deletedCount)
// 		},
// 	})
// }

// func (this *EntitlementAssignmentServiceImpl) assertEntitlementAssignmentExistsById(ctx crud.Context, entitlementAssignment *domain.EntitlementGrant, vErrs *fault.ValidationErrors) (dbEntitlementAssignment *domain.EntitlementGrant, err error) {
// 	dbEntitlementAssignment, err = this.entitlementAssignmentRepo.FindById(ctx, it.FindByIdParam{Id: *entitlementAssignment.GetId()})
// 	fault.PanicOnErr(err)

// 	if dbEntitlementAssignment == nil {
// 		vErrs.AppendNotFound("id", "entitlement assignment")
// 	}
// 	return dbEntitlementAssignment, err
// }

// func (this *EntitlementAssignmentServiceImpl) enableFieldsInPermissionHistory(ctx crud.Context, assignment *domain.EntitlementGrant) error {
// 	return this.permissionHistoryRepo.EnableField(
// 		ctx,
// 		itPermissionHistory.EnableFieldCommand{
// 			AssignmentId: assignment.GetId(),
// 			ResolvedExpr: *assignment.GetResolvedExpr(),
// 		},
// 	)
// }

// func (this *EntitlementAssignmentServiceImpl) assertBusinessRuleDeleteEntitlementAssignment(ctx crud.Context, cmd it.DeleteEntitlementAssignmentByIdCommand, assignment *domain.EntitlementGrant, vErrs *fault.ValidationErrors) error {
// 	err := this.enableFieldsInPermissionHistory(ctx, assignment)
// 	fault.PanicOnErr(err)

// 	return nil
// }

// func (this *EntitlementAssignmentServiceImpl) assertBusinessRuleDeleteEntitlementAssignmentByEntitlementId(ctx crud.Context, cmd it.DeleteEntitlementAssignmentByEntitlementIdCommand, assignment *domain.EntitlementGrant, vErrs *fault.ValidationErrors) error {
// 	assignments, err := this.entitlementAssignmentRepo.FindAllByEntitlementId(
// 		ctx,
// 		it.FindAllByEntitlementIdParam{EntitlementId: cmd.EntitlementId},
// 	)
// 	fault.PanicOnErr(err)

// 	for _, assignment := range assignments {
// 		err = this.enableFieldsInPermissionHistory(ctx, &assignment)
// 		fault.PanicOnErr(err)
// 	}

// 	return nil
// }

// func (this *EntitlementAssignmentServiceImpl) setEntitlementAssignmentDefaults(assignment *domain.EntitlementGrant) {
// 	if assignment.GetId() != nil {
// 		return
// 	}
// 	idPtr, err := model.NewId()
// 	fault.PanicOnErr(err)
// 	assignment.SetId(idPtr)
// }

// func (this *EntitlementAssignmentServiceImpl) assertBusinessRuleCreateEntitlementAssignment(ctx crud.Context, assignment *domain.EntitlementGrant, vErrs *fault.ValidationErrors) error {
// 	entRes, err := this.entitlementService.GetEntitlementById(ctx, itEntitlement.GetEntitlementByIdQuery{
// 		Id: *assignment.GetEntitlementId(),
// 	})
// 	fault.PanicOnErr(err)

// 	if entRes.ClientError != nil {
// 		return entRes.ClientError
// 	}

// 	ent := entRes.Data
// 	if ent == nil {
// 		vErrs.AppendNotFound("entitlement_id", "entitlement id")
// 		return nil
// 	}

// 	if ent.GetResourceId() == nil {
// 		if assignment.GetScopeRef() != nil {
// 			vErrs.AppendNotAllowed("scope_ref", "global entitlement must not have scopeRef")
// 			return nil
// 		}

// 		this.assertEntitlementAssignmentUnique(ctx, assignment, vErrs)
// 		return nil
// 	}

// 	err = this.assertScopeRefByScopeType(ctx, ent, assignment, vErrs)
// 	fault.PanicOnErr(err)

// 	this.validateFields(ctx, ent, assignment, vErrs)
// 	this.assertEntitlementAssignmentUnique(ctx, assignment, vErrs)

// 	return nil
// }

// func (this *EntitlementAssignmentServiceImpl) assertScopeRefByScopeType(ctx crud.Context, ent *domain.Entitlement, assignment *domain.EntitlementGrant, vErrs *fault.ValidationErrors) error {
// 	rid := ent.GetResourceId()
// 	if rid == nil {
// 		return nil
// 	}
// 	resRes, err := this.resourceService.GetResourceById(ctx, itResource.GetResourceByIdQuery{Id: *rid})
// 	fault.PanicOnErr(err)
// 	if resRes.ClientError != nil {
// 		vErrs.MergeClientError(resRes.ClientError)
// 		return nil
// 	}
// 	if resRes.Data == nil {
// 		vErrs.AppendNotFound("resource_id", "resource")
// 		return nil
// 	}
// 	res := resRes.Data
// 	st := res.GetScopeType()
// 	if st == nil {
// 		vErrs.AppendNotAllowed("resource.scope_type", "scope type")
// 		return nil
// 	}
// 	switch *st {
// 	case domain.ResourceMinScopeDomain:
// 		if assignment.GetScopeRef() != nil {
// 			vErrs.AppendNotAllowed("scope_ref", "scopeRef of domain-level resource")
// 			return nil
// 		}

// 	case domain.ResourceMinScopeOrg:
// 		if assignment.GetScopeRef() == nil {
// 			vErrs.AppendNotAllowed("scope_ref", "scopeRef of org-level resource")
// 			return nil
// 		}

// 		existCmd := itOrg.OrgExistsQuery{
// 			Ids: []model.Id{*assignment.GetScopeRef()},
// 		}
// 		existRes := itOrg.OrgExistsResult{}
// 		err := this.cqrsBus.Request(ctx, existCmd, &existRes)
// 		fault.PanicOnErr(err)

// 		if existRes.ClientErrors.Count() > 0 {
// 			for i := range existRes.ClientErrors {
// 				e := existRes.ClientErrors[i]
// 				vErrs.Append(e.Field, e.Message)
// 			}
// 			return nil
// 		}

// 	case domain.ResourceMinScopeHierarchy:
// 		if assignment.GetScopeRef() == nil {
// 			vErrs.AppendNotAllowed("scope_ref", "scopeRef of hierarchy-level resource")
// 			return nil
// 		}

// 		existCmd := &itHierarchy.HierarchyLevelExistsQuery{
// 			Ids: []model.Id{*assignment.GetScopeRef()},
// 		}
// 		existRes := itHierarchy.HierarchyLevelExistsResult{}
// 		err := this.cqrsBus.Request(ctx, *existCmd, &existRes)
// 		fault.PanicOnErr(err)

// 		if len(existRes.ClientErrors) > 0 {
// 			for i := range existRes.ClientErrors {
// 				e := existRes.ClientErrors[i]
// 				vErrs.Append(e.Field, e.Message)
// 			}
// 			return nil
// 		}

// 		if !existRes.Data.Exists(*assignment.GetScopeRef()) {
// 			vErrs.AppendNotFound("scope_ref", "scope ref")
// 			return nil
// 		}
// 		return nil

// 	case domain.ResourceMinTypePrivate:
// 		if assignment.GetScopeRef() == nil {
// 			vErrs.AppendNotAllowed("scope_ref", "scopeRef of private resource")
// 			return nil
// 		}

// 	default:
// 		vErrs.AppendNotAllowed("resource.scope_type", "scope type")
// 		return nil
// 	}

// 	return nil
// }

// func (this *EntitlementAssignmentServiceImpl) assertEntitlementAssignmentUnique(ctx crud.Context, assignment *domain.EntitlementGrant, vErrs *fault.ValidationErrors) {
// 	exists, err := this.entitlementAssignmentRepo.FindByFilter(ctx, it.FindByFilterParam{
// 		SubjectType:   *assignment.GetSubjectType(),
// 		SubjectRef:    *assignment.GetSubjectRef(),
// 		EntitlementId: *assignment.GetEntitlementId(),
// 		ScopeRef:      assignment.GetScopeRef(),
// 	})
// 	fault.PanicOnErr(err)

// 	if exists != nil {
// 		vErrs.AppendAlreadyExists("subject_type", "subject type")
// 	}
// }

// func (this *EntitlementAssignmentServiceImpl) validateFields(ctx crud.Context, entitlement *domain.Entitlement, assignment *domain.EntitlementGrant, vErrs *fault.ValidationErrors) {
// 	var expectedResourceName string
// 	if rid := entitlement.GetResourceId(); rid != nil {
// 		rr, err := this.resourceService.GetResourceById(ctx, itResource.GetResourceByIdQuery{Id: *rid})
// 		fault.PanicOnErr(err)
// 		if rr.Data != nil && rr.Data.GetName() != nil {
// 			expectedResourceName = *rr.Data.GetName()
// 		}
// 	}

// 	var expectedActionName *string
// 	if aid := entitlement.GetActionId(); aid != nil {
// 		ar := itAction.GetActionByIdResult{}
// 		err := this.cqrsBus.Request(ctx, itAction.GetActionByIdQuery{Id: *aid}, &ar)
// 		fault.PanicOnErr(err)
// 		if ar.Data != nil {
// 			expectedActionName = ar.Data.GetName()
// 		}
// 	}

// 	if expectedActionName == nil {
// 		if assignment.GetActionName() != nil {
// 			vErrs.AppendNotAllowed("action_name", "action name")
// 		}
// 	} else if assignment.GetActionName() == nil || *assignment.GetActionName() != *expectedActionName {
// 		vErrs.AppendNotAllowed("action_name", "action name")
// 	}

// 	if assignment.GetResourceName() == nil || *assignment.GetResourceName() != expectedResourceName {
// 		vErrs.AppendNotAllowed("resource_name", "resource name")
// 	}

// 	this.validateResolvedExpr(entitlement, assignment, vErrs)
// }

// func (this *EntitlementAssignmentServiceImpl) validateResolvedExpr(entitlement *domain.Entitlement, assignment *domain.EntitlementGrant, vErrs *fault.ValidationErrors) {
// 	scopePart := "*"
// 	if assignment.GetScopeRef() != nil {
// 		scopePart = *assignment.GetScopeRef()
// 	}
// 	ae := entitlement.GetActionExpr()
// 	if ae == nil || assignment.GetSubjectRef() == nil || assignment.GetResolvedExpr() == nil {
// 		return
// 	}
// 	expectedResolvedExpr := fmt.Sprintf("%s:%s:%s", *assignment.GetSubjectRef(), scopePart, *ae)
// 	if *assignment.GetResolvedExpr() != expectedResolvedExpr {
// 		vErrs.AppendNotAllowed("resolved_expr", "resolvedExpr")
// 	}
// }
