package app

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/contacts/domain"
	itParty "github.com/sky-as-code/nikki-erp/modules/contacts/interfaces/party"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
)

func NewPartyServiceImpl(
	partyRepo itParty.PartyRepository,
) itParty.PartyService {
	return &PartyServiceImpl{
		partyRepo: partyRepo,
	}
}

type PartyServiceImpl struct {
	partyRepo itParty.PartyRepository
}

func (this *PartyServiceImpl) CreateParty(ctx corectx.Context, cmd itParty.CreatePartyCommand) (*itParty.CreatePartyResult, error) {
	return corecrud.Create(ctx, corecrud.CreateParam[domain.Party, *domain.Party]{
		Action:         "create party",
		BaseRepoGetter: this.partyRepo,
		Data:           cmd,
	})
}

func (this *PartyServiceImpl) UpdateParty(ctx corectx.Context, cmd itParty.UpdatePartyCommand) (*itParty.UpdatePartyResult, error) {
	return corecrud.Update(ctx, corecrud.UpdateParam[domain.Party, *domain.Party]{
		Action:       "update party",
		DbRepoGetter: this.partyRepo,
		Data:         cmd,
	})
}

func (this *PartyServiceImpl) DeleteParty(ctx corectx.Context, cmd itParty.DeletePartyCommand) (*itParty.DeletePartyResult, error) {
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{
		Action:       "delete party",
		DbRepoGetter: this.partyRepo,
		Cmd:          dyn.DeleteOneCommand(cmd),
	})
}

func (this *PartyServiceImpl) GetParty(ctx corectx.Context, query itParty.GetPartyQuery) (*itParty.GetPartyResult, error) {
	keyNode := dmodel.NewSearchNode()
	if query.Id != nil {
		keyNode.NewCondition(domain.PartyFieldId, dmodel.Equals, *query.Id)
	}
	if query.DisplayName != nil {
		keyNode.NewCondition(domain.PartyFieldDisplayName, dmodel.Equals, *query.DisplayName)
	}

	graph := &dmodel.SearchGraph{}
	graph.And(*keyNode)

	searchRes, err := this.SearchParties(ctx, itParty.SearchPartiesQuery(dyn.SearchQuery{
		Fields: query.Columns,
		Graph:  graph,
		Page:   0,
		Size:   1,
	}))
	if err != nil {
		return nil, err
	}

	result := &itParty.GetPartyResult{
		ClientErrors: searchRes.ClientErrors,
	}
	if searchRes.HasData && len(searchRes.Data.Items) > 0 {
		result.Data = searchRes.Data.Items[0]
		result.HasData = true
	}
	return result, nil
}

func (this *PartyServiceImpl) SearchParties(ctx corectx.Context, query itParty.SearchPartiesQuery) (*itParty.SearchPartiesResult, error) {
	return corecrud.Search[domain.Party](ctx, corecrud.SearchParam{
		Action:       "search parties",
		DbRepoGetter: this.partyRepo,
		Query:        dyn.SearchQuery(query),
	})
}

func (this *PartyServiceImpl) PartyExists(ctx corectx.Context, query itParty.PartyExistsQuery) (*itParty.PartyExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{
		Action:       "check if parties exist",
		DbRepoGetter: this.partyRepo,
		Query:        dyn.ExistsQuery(query),
	})
}
