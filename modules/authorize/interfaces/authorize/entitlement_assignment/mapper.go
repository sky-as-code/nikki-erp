package entitlement_assignment

import (
	// "github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/authorize/domain"
)

func (this CreateEntitlementAssignmentCommand) ToEntitlementAssignment() *domain.EntitlementAssignment {
	return &domain.EntitlementAssignment{
		SubjectType:   domain.WrapEntitlementAssignmentSubjectType(*this.SubjectType),
		SubjectRef:    this.SubjectRef,
		ActionName:    this.ActionName,
		ResourceName:  this.ResourceName,
		ResolvedExpr:  this.ResolvedExpr,
		EntitlementId: this.EntitlementId,
	}
}

// func (this UpdateActionCommand) ToAction() *domain.Action {
// 	return &domain.Action{
// 		ModelBase: model.ModelBase{
// 			Id:   &this.Id,
// 			Etag: &this.Etag,
// 		},
// 		Description: this.Description,
// 	}
// }
