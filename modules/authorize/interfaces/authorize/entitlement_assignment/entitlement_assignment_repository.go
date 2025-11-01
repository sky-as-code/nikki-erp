package entitlement_assignment

import (
	domain "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

type EntitlementAssignmentRepository interface {
	Create(ctx crud.Context, assignment *domain.EntitlementAssignment) (*domain.EntitlementAssignment, error)
	FindByFilter(ctx crud.Context, param FindByFilterParam) (*domain.EntitlementAssignment, error)
	FindAllByEntitlementId(ctx crud.Context, param FindAllByEntitlementIdParam) ([]domain.EntitlementAssignment, error)
	FindAllBySubject(ctx crud.Context, param FindBySubjectParam) ([]domain.EntitlementAssignment, error)
	FindById(ctx crud.Context, param FindByIdParam) (*domain.EntitlementAssignment, error)
	FindViewsById(ctx crud.Context, param FindViewsByIdParam) ([]domain.EntitlementAssignment, error)
	DeleteHard(ctx crud.Context, param DeleteHardParam) (int, error)
	DeleteHardByEntitlementId(ctx crud.Context, param DeleteHardByEntitlementIdParam) (int, error)
}

type CreateParam = CreateEntitlementAssignmentCommand
type FindByFilterParam = GetEntitlementAssignmentByFilterQuery
type FindAllByEntitlementIdParam = GetAllEntitlementAssignmentByEntitlementIdQuery
type FindByIdParam = GetByIdQuery
type FindBySubjectParam = GetAllEntitlementAssignmentBySubjectQuery
type FindViewsByIdParam = GetViewsByIdQuery
type DeleteHardByEntitlementIdParam = DeleteEntitlementAssignmentByEntitlementIdCommand
type DeleteHardParam = DeleteEntitlementAssignmentByIdCommand
