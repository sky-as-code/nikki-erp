package ticketfeedback

import corectx "github.com/sky-as-code/nikki-erp/modules/core/context"

type TicketFeedbackDomainService interface {
	CreateTicketFeedback(ctx corectx.Context, cmd CreateTicketFeedbackCommand) (*CreateTicketFeedbackResult, error)
	DeleteTicketFeedback(ctx corectx.Context, cmd DeleteTicketFeedbackCommand) (*DeleteTicketFeedbackResult, error)
	GetTicketFeedback(ctx corectx.Context, query GetTicketFeedbackQuery) (*GetTicketFeedbackResult, error)
	TicketFeedbackExists(ctx corectx.Context, query TicketFeedbackExistsQuery) (*TicketFeedbackExistsResult, error)
	SearchTicketFeedbacks(ctx corectx.Context, query SearchTicketFeedbacksQuery) (*SearchTicketFeedbacksResult, error)
	UpdateTicketFeedback(ctx corectx.Context, cmd UpdateTicketFeedbackCommand) (*UpdateTicketFeedbackResult, error)
}

type TicketFeedbackAppService interface {
	CreateTicketFeedback(ctx corectx.Context, cmd CreateTicketFeedbackCommand) (*CreateTicketFeedbackResult, error)
	DeleteTicketFeedback(ctx corectx.Context, cmd DeleteTicketFeedbackCommand) (*DeleteTicketFeedbackResult, error)
	GetTicketFeedback(ctx corectx.Context, query GetTicketFeedbackQuery) (*GetTicketFeedbackResult, error)
	TicketFeedbackExists(ctx corectx.Context, query TicketFeedbackExistsQuery) (*TicketFeedbackExistsResult, error)
	SearchTicketFeedbacks(ctx corectx.Context, query SearchTicketFeedbacksQuery) (*SearchTicketFeedbacksResult, error)
	UpdateTicketFeedback(ctx corectx.Context, cmd UpdateTicketFeedbackCommand) (*UpdateTicketFeedbackResult, error)
}
