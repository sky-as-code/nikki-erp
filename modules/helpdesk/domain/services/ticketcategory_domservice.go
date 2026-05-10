package services

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/helpdesk/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/helpdesk/interfaces/ticketcategory"
)

func NewTicketCategoryDomainServiceImpl(repo it.TicketCategoryRepository, cqrsBus cqrs.CqrsBus) it.TicketCategoryDomainService {
	return &TicketCategoryDomainServiceImpl{cqrsBus: cqrsBus, repo: repo}
}

type TicketCategoryDomainServiceImpl struct {
	cqrsBus cqrs.CqrsBus
	repo    it.TicketCategoryRepository
}

func (this *TicketCategoryDomainServiceImpl) CreateTicketCategory(
	ctx corectx.Context, cmd it.CreateTicketCategoryCommand,
) (*it.CreateTicketCategoryResult, error) {
	return corecrud.Create(ctx, corecrud.CreateParam[models.TicketCategory, *models.TicketCategory]{Action: "create ticketCategory", BaseRepoGetter: this.repo, Data: cmd})
}

func (this *TicketCategoryDomainServiceImpl) DeleteTicketCategory(
	ctx corectx.Context, cmd it.DeleteTicketCategoryCommand,
) (*it.DeleteTicketCategoryResult, error) {
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{Action: "delete ticketCategory", DbRepoGetter: this.repo, Cmd: dyn.DeleteOneCommand(cmd)})
}

func (this *TicketCategoryDomainServiceImpl) GetTicketCategory(
	ctx corectx.Context, query it.GetTicketCategoryQuery,
) (*it.GetTicketCategoryResult, error) {
	return corecrud.GetOne[models.TicketCategory](ctx, corecrud.GetOneParam{Action: "get ticketCategory", DbRepoGetter: this.repo, Query: dyn.GetOneQuery(query)})
}

func (this *TicketCategoryDomainServiceImpl) TicketCategoryExists(
	ctx corectx.Context, query it.TicketCategoryExistsQuery,
) (*it.TicketCategoryExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{Action: "check if ticketCategory exists", DbRepoGetter: this.repo, Query: dyn.ExistsQuery(query)})
}

func (this *TicketCategoryDomainServiceImpl) SearchTicketCategories(
	ctx corectx.Context, query it.SearchTicketCategoriesQuery,
) (*it.SearchTicketCategoriesResult, error) {
	return corecrud.Search[models.TicketCategory](ctx, corecrud.SearchParam{Action: "search ticketCategorys", DbRepoGetter: this.repo, Query: dyn.SearchQuery(query)})
}

func (this *TicketCategoryDomainServiceImpl) UpdateTicketCategory(
	ctx corectx.Context, cmd it.UpdateTicketCategoryCommand,
) (*it.UpdateTicketCategoryResult, error) {
	return corecrud.Update(ctx, corecrud.UpdateParam[models.TicketCategory, *models.TicketCategory]{Action: "update ticketCategory", DbRepoGetter: this.repo, Data: cmd})
}

func (this *TicketCategoryDomainServiceImpl) SetTicketCategoryIsArchived(
	ctx corectx.Context, cmd it.SetTicketCategoryIsArchivedCommand,
) (*it.SetTicketCategoryIsArchivedResult, error) {
	return corecrud.SetIsArchived(ctx, this.repo, dyn.SetIsArchivedCommand(cmd))
}
