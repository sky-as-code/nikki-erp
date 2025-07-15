package app

import (
	"context"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	// "github.com/sky-as-code/nikki-erp/common/model"
	// "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/entitlement_assignment"
	"github.com/sky-as-code/nikki-erp/modules/core/event"
)

func NewEntitlementAssignmentServiceImpl(entitlementAssignmentRepo it.EntitlementAssignmentRepository, eventBus event.EventBus) it.EntitlementAssignmentService {
	return &EntitlementAssignmentServiceImpl{
		entitlementAssignmentRepo: entitlementAssignmentRepo,
		eventBus:                  eventBus,
	}
}

type EntitlementAssignmentServiceImpl struct {
	entitlementAssignmentRepo it.EntitlementAssignmentRepository
	eventBus                  event.EventBus
}

func (this *EntitlementAssignmentServiceImpl) FindAllBySubject(ctx context.Context, query it.GetAllEntitlementAssignmentBySubjectQuery) (result *it.GetAllEntitlementAssignmentBySubjectResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to get entitlement assignment by subject"); e != nil {
			err = e
		}
	}()

	vErrs := query.Validate()
	if vErrs.Count() > 0 {
		return &it.GetAllEntitlementAssignmentBySubjectResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	entitlementAssignments, err := this.entitlementAssignmentRepo.FindAllBySubject(ctx, it.FindBySubjectParam{
		SubjectType: query.SubjectType,
		SubjectRef:  query.SubjectRef,
	})
	ft.PanicOnErr(err)

	if len(entitlementAssignments) == 0 {
		vErrs.Append("subjectType", "entitlement assignment not found")
		vErrs.Append("subjectRef", "entitlement assignment not found")
		return &it.GetAllEntitlementAssignmentBySubjectResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	return &it.GetAllEntitlementAssignmentBySubjectResult{
		Data: entitlementAssignments,
	}, nil
}
