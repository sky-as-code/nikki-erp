package entitlement_assignment

import (
	"github.com/sky-as-code/nikki-erp/common/model"
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

func (this DeleteEntitlementAssignmentByIdCommand) ToDomainModel() *domain.EntitlementAssignment {
	assignment := &domain.EntitlementAssignment{}
	model.MustCopy(this, assignment)

	return assignment
}

func (this DeleteEntitlementAssignmentByEntitlementIdCommand) ToDomainModel() *domain.EntitlementAssignment {
	assignment := &domain.EntitlementAssignment{}
	model.MustCopy(this, assignment)

	return assignment
}
