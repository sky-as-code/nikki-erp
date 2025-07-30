package entitlement_assignment

import (
	"context"

	domain "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
)

type EntitlementAssignmentRepository interface {
	Create(ctx context.Context, assignment domain.EntitlementAssignment) (*domain.EntitlementAssignment, error)
	CreateBulk(ctx context.Context, assignments []domain.EntitlementAssignment) error
	FindAllBySubject(ctx context.Context, param FindBySubjectParam) ([]*domain.EntitlementAssignment, error)
	FindViewsById(ctx context.Context, param FindViewsByIdParam) ([]*domain.EntitlementAssignment, error)
	FindAllByEntitlementId(ctx context.Context, param FindAllByEntitlementIdParam) ([]*domain.EntitlementAssignment, error)
	DeleteHard(ctx context.Context, param DeleteHard) (int, error)
}

type FindBySubjectParam = GetAllEntitlementAssignmentBySubjectQuery
type FindViewsByIdParam = GetViewsByIdQuery
type FindAllByEntitlementIdParam = GetAllEntitlementAssignmentByEntitlementIdQuery
type DeleteHard = DeleteEntitlementAssignmentByIdQuery
