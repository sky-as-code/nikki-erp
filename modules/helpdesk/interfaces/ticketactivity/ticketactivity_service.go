package ticketactivity

import corectx "github.com/sky-as-code/nikki-erp/modules/core/context"

type TicketActivityService interface {
	CreateTicketActivity(ctx corectx.Context, cmd CreateTicketActivityCommand) (*CreateTicketActivityResult, error)
	DeleteTicketActivity(ctx corectx.Context, cmd DeleteTicketActivityCommand) (*DeleteTicketActivityResult, error)
	GetTicketActivity(ctx corectx.Context, query GetTicketActivityQuery) (*GetTicketActivityResult, error)
	TicketActivityExists(ctx corectx.Context, query TicketActivityExistsQuery) (*TicketActivityExistsResult, error)
	SearchTicketActivities(ctx corectx.Context, query SearchTicketActivitiesQuery) (*SearchTicketActivitiesResult, error)
	UpdateTicketActivity(ctx corectx.Context, cmd UpdateTicketActivityCommand) (*UpdateTicketActivityResult, error)
}
