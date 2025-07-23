package entitlement_assignment

import (
	"context"
)

type EntitlementAssignmentService interface {
	FindAllBySubject(ctx context.Context, query GetAllEntitlementAssignmentBySubjectQuery) (*GetAllEntitlementAssignmentBySubjectResult, error)
}
