package entitlement_assignment

import (
	"context"

	domain "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
)

type EntitlementAssignmentRepository interface {
	Create(ctx context.Context, assignment domain.EntitlementAssignment) (*domain.EntitlementAssignment, error)
	CreateBulk(ctx context.Context, assignments []domain.EntitlementAssignment) error
	FindAllBySubject(ctx context.Context, param FindBySubjectParam) ([]*domain.EntitlementAssignment, error)
}

type FindBySubjectParam = GetAllEntitlementAssignmentBySubjectQuery
