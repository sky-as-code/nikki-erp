package ticket

import corectx "github.com/sky-as-code/nikki-erp/modules/core/context"

type TicketService interface {
	CreateTicket(ctx corectx.Context, cmd CreateTicketCommand) (*CreateTicketResult, error)
	DeleteTicket(ctx corectx.Context, cmd DeleteTicketCommand) (*DeleteTicketResult, error)
	GetTicket(ctx corectx.Context, query GetTicketQuery) (*GetTicketResult, error)
	TicketExists(ctx corectx.Context, query TicketExistsQuery) (*TicketExistsResult, error)
	SearchTickets(ctx corectx.Context, query SearchTicketsQuery) (*SearchTicketsResult, error)
	UpdateTicket(ctx corectx.Context, cmd UpdateTicketCommand) (*UpdateTicketResult, error)
	SetTicketIsArchived(ctx corectx.Context, cmd SetTicketIsArchivedCommand) (*SetTicketIsArchivedResult, error)
	ManageTicketCategories(ctx corectx.Context, cmd ManageTicketCategoriesCommand) (*ManageTicketCategoriesResult, error)
}
