package entitlement_assignment

import (
	domain "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

type EntitlementAssignmentRepository interface {
	FindAllBySubject(ctx crud.Context, param FindBySubjectParam) ([]*domain.EntitlementAssignment, error)
	FindViewsById(ctx crud.Context, param FindViewsByIdParam) ([]*domain.EntitlementAssignment, error)
	FindAllByEntitlementId(ctx crud.Context, param FindAllByEntitlementIdParam) ([]*domain.EntitlementAssignment, error)
	DeleteHard(ctx crud.Context, param DeleteHard) (int, error)
	DeleteHardTx(ctx crud.Context, param DeleteHard) (int, error)
}

type FindBySubjectParam = GetAllEntitlementAssignmentBySubjectQuery
type FindViewsByIdParam = GetViewsByIdQuery
type FindAllByEntitlementIdParam = GetAllEntitlementAssignmentByEntitlementIdQuery
type DeleteHard = DeleteEntitlementAssignmentByIdQuery
