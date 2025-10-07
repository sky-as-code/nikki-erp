package app

import (
	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"

	"github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/entitlement_assignment"
	itPermissionHistory "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/permission_history"
)

func NewEntitlementAssignmentServiceImpl(
	entitlementAssignmentRepo it.EntitlementAssignmentRepository,
	permissionHistoryRepo itPermissionHistory.PermissionHistoryRepository,
) it.EntitlementAssignmentService {
	return &EntitlementAssignmentServiceImpl{
		entitlementAssignmentRepo: entitlementAssignmentRepo,
		permissionHistoryRepo:     permissionHistoryRepo,
	}
}

type EntitlementAssignmentServiceImpl struct {
	entitlementAssignmentRepo it.EntitlementAssignmentRepository
	permissionHistoryRepo     itPermissionHistory.PermissionHistoryRepository
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
