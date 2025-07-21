package entitlement_assignment

import (
	domain "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
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
