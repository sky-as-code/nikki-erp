package ticketassignment

import corectx "github.com/sky-as-code/nikki-erp/modules/core/context"

type TicketAssignmentService interface {
	CreateTicketAssignment(ctx corectx.Context, cmd CreateTicketAssignmentCommand) (*CreateTicketAssignmentResult, error)
	DeleteTicketAssignment(ctx corectx.Context, cmd DeleteTicketAssignmentCommand) (*DeleteTicketAssignmentResult, error)
	GetTicketAssignment(ctx corectx.Context, query GetTicketAssignmentQuery) (*GetTicketAssignmentResult, error)
	TicketAssignmentExists(ctx corectx.Context, query TicketAssignmentExistsQuery) (*TicketAssignmentExistsResult, error)
	SearchTicketAssignments(ctx corectx.Context, query SearchTicketAssignmentsQuery) (*SearchTicketAssignmentsResult, error)
	UpdateTicketAssignment(ctx corectx.Context, cmd UpdateTicketAssignmentCommand) (*UpdateTicketAssignmentResult, error)
}
