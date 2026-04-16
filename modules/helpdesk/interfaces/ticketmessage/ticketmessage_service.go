package ticketmessage

import corectx "github.com/sky-as-code/nikki-erp/modules/core/context"

type TicketMessageService interface {
	CreateTicketMessage(ctx corectx.Context, cmd CreateTicketMessageCommand) (*CreateTicketMessageResult, error)
	DeleteTicketMessage(ctx corectx.Context, cmd DeleteTicketMessageCommand) (*DeleteTicketMessageResult, error)
	GetTicketMessage(ctx corectx.Context, query GetTicketMessageQuery) (*GetTicketMessageResult, error)
	TicketMessageExists(ctx corectx.Context, query TicketMessageExistsQuery) (*TicketMessageExistsResult, error)
	SearchTicketMessages(ctx corectx.Context, query SearchTicketMessagesQuery) (*SearchTicketMessagesResult, error)
	UpdateTicketMessage(ctx corectx.Context, cmd UpdateTicketMessageCommand) (*UpdateTicketMessageResult, error)
}
