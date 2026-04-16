package ticketcategory

import corectx "github.com/sky-as-code/nikki-erp/modules/core/context"

type TicketCategoryService interface {
	CreateTicketCategory(ctx corectx.Context, cmd CreateTicketCategoryCommand) (*CreateTicketCategoryResult, error)
	DeleteTicketCategory(ctx corectx.Context, cmd DeleteTicketCategoryCommand) (*DeleteTicketCategoryResult, error)
	GetTicketCategory(ctx corectx.Context, query GetTicketCategoryQuery) (*GetTicketCategoryResult, error)
	TicketCategoryExists(ctx corectx.Context, query TicketCategoryExistsQuery) (*TicketCategoryExistsResult, error)
	SearchTicketCategories(ctx corectx.Context, query SearchTicketCategoriesQuery) (*SearchTicketCategoriesResult, error)
	UpdateTicketCategory(ctx corectx.Context, cmd UpdateTicketCategoryCommand) (*UpdateTicketCategoryResult, error)
	SetTicketCategoryIsArchived(ctx corectx.Context, cmd SetTicketCategoryIsArchivedCommand) (*SetTicketCategoryIsArchivedResult, error)
}
