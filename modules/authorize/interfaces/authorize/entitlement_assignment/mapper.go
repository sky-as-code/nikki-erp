package entitlement_assignment

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/authorize/domain"
)

func (this CreateEntitlementAssignmentCommand) ToDomainModel() *domain.EntitlementAssignment {
	assignment := &domain.EntitlementAssignment{}
	model.MustCopy(this, assignment)

	return assignment
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
