package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/helpdesk/domain"
	it "github.com/sky-as-code/nikki-erp/modules/helpdesk/interfaces/ticketcategory"
)

func NewTicketCategoryServiceImpl(repo it.TicketCategoryRepository, cqrsBus cqrs.CqrsBus) it.TicketCategoryService {
	return &TicketCategoryServiceImpl{cqrsBus: cqrsBus, repo: repo}
}

type TicketCategoryServiceImpl struct {
	cqrsBus cqrs.CqrsBus
	repo    it.TicketCategoryRepository
}

func (this *TicketCategoryServiceImpl) CreateTicketCategory(
	ctx corectx.Context, cmd it.CreateTicketCategoryCommand,
) (*it.CreateTicketCategoryResult, error) {
	return corecrud.Create(ctx, corecrud.CreateParam[domain.TicketCategory, *domain.TicketCategory]{Action: "create ticketCategory", BaseRepoGetter: this.repo, Data: cmd})
}

func (this *TicketCategoryServiceImpl) DeleteTicketCategory(
	ctx corectx.Context, cmd it.DeleteTicketCategoryCommand,
) (*it.DeleteTicketCategoryResult, error) {
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{Action: "delete ticketCategory", DbRepoGetter: this.repo, Cmd: dyn.DeleteOneCommand(cmd)})
}

func (this *TicketCategoryServiceImpl) GetTicketCategory(
	ctx corectx.Context, query it.GetTicketCategoryQuery,
) (*it.GetTicketCategoryResult, error) {
	return corecrud.GetOne[domain.TicketCategory](ctx, corecrud.GetOneParam{Action: "get ticketCategory", DbRepoGetter: this.repo, Query: dyn.GetOneQuery(query)})
}

func (this *TicketCategoryServiceImpl) TicketCategoryExists(
	ctx corectx.Context, query it.TicketCategoryExistsQuery,
) (*it.TicketCategoryExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{Action: "check if ticketCategory exists", DbRepoGetter: this.repo, Query: dyn.ExistsQuery(query)})
}

func (this *TicketCategoryServiceImpl) SearchTicketCategories(
	ctx corectx.Context, query it.SearchTicketCategoriesQuery,
) (*it.SearchTicketCategoriesResult, error) {
	return corecrud.Search[domain.TicketCategory](ctx, corecrud.SearchParam{Action: "search ticketCategorys", DbRepoGetter: this.repo, Query: dyn.SearchQuery(query)})
}

func (this *TicketCategoryServiceImpl) UpdateTicketCategory(
	ctx corectx.Context, cmd it.UpdateTicketCategoryCommand,
) (*it.UpdateTicketCategoryResult, error) {
	return corecrud.Update(ctx, corecrud.UpdateParam[domain.TicketCategory, *domain.TicketCategory]{Action: "update ticketCategory", DbRepoGetter: this.repo, Data: cmd})
}

func (this *TicketCategoryServiceImpl) SetTicketCategoryIsArchived(
	ctx corectx.Context, cmd it.SetTicketCategoryIsArchivedCommand,
) (*it.SetTicketCategoryIsArchivedResult, error) {
	return corecrud.SetIsArchived(ctx, this.repo, dyn.SetIsArchivedCommand(cmd))
}
