package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	it "github.com/sky-as-code/nikki-erp/modules/helpdesk/interfaces/ticketcategory"
)

func NewTicketCategoryApplicationServiceImpl(ticketCategorySvc it.TicketCategoryDomainService) it.TicketCategoryAppService {
	return &TicketCategoryApplicationServiceImpl{ticketCategorySvc: ticketCategorySvc}
}

type TicketCategoryApplicationServiceImpl struct {
	ticketCategorySvc it.TicketCategoryDomainService
}

func (this *TicketCategoryApplicationServiceImpl) CreateTicketCategory(ctx corectx.Context, cmd it.CreateTicketCategoryCommand) (*it.CreateTicketCategoryResult, error) {
	return this.ticketCategorySvc.CreateTicketCategory(ctx, cmd)
}

func (this *TicketCategoryApplicationServiceImpl) DeleteTicketCategory(ctx corectx.Context, cmd it.DeleteTicketCategoryCommand) (*it.DeleteTicketCategoryResult, error) {
	return this.ticketCategorySvc.DeleteTicketCategory(ctx, cmd)
}

func (this *TicketCategoryApplicationServiceImpl) GetTicketCategory(ctx corectx.Context, query it.GetTicketCategoryQuery) (*it.GetTicketCategoryResult, error) {
	return this.ticketCategorySvc.GetTicketCategory(ctx, query)
}

func (this *TicketCategoryApplicationServiceImpl) TicketCategoryExists(ctx corectx.Context, query it.TicketCategoryExistsQuery) (*it.TicketCategoryExistsResult, error) {
	return this.ticketCategorySvc.TicketCategoryExists(ctx, query)
}

func (this *TicketCategoryApplicationServiceImpl) SearchTicketCategories(ctx corectx.Context, query it.SearchTicketCategoriesQuery) (*it.SearchTicketCategoriesResult, error) {
	return this.ticketCategorySvc.SearchTicketCategories(ctx, query)
}

func (this *TicketCategoryApplicationServiceImpl) UpdateTicketCategory(ctx corectx.Context, cmd it.UpdateTicketCategoryCommand) (*it.UpdateTicketCategoryResult, error) {
	return this.ticketCategorySvc.UpdateTicketCategory(ctx, cmd)
}

func (this *TicketCategoryApplicationServiceImpl) SetTicketCategoryIsArchived(ctx corectx.Context, cmd it.SetTicketCategoryIsArchivedCommand) (*it.SetTicketCategoryIsArchivedResult, error) {
	return this.ticketCategorySvc.SetTicketCategoryIsArchived(ctx, cmd)
}
