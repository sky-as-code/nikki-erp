package app

import (
	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/core/event"

	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/entitlement_assignment"
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
