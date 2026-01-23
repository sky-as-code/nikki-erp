package entitlement_assignment

import "github.com/sky-as-code/nikki-erp/modules/core/crud"

type EntitlementAssignmentService interface {
	CreateEntitlementAssignment(ctx crud.Context, cmd CreateEntitlementAssignmentCommand) (*CreateEntitlementAssignmentResult, error)
	FindAllBySubject(ctx crud.Context, query GetAllEntitlementAssignmentBySubjectQuery) (*GetAllEntitlementAssignmentBySubjectResult, error)
	DeleteHardAssignment(ctx crud.Context, cmd DeleteEntitlementAssignmentByIdCommand) (*DeleteEntitlementAssignmentByIdResult, error)
	DeleteByEntitlementId(ctx crud.Context, cmd DeleteEntitlementAssignmentByEntitlementIdCommand) (*DeleteEntitlementAssignmentByEntitlementIdResult, error)
}
