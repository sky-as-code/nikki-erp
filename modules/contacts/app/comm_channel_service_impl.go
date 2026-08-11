package app

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/contacts/domain"
	itChannel "github.com/sky-as-code/nikki-erp/modules/contacts/interfaces/commchannel"
	itParty "github.com/sky-as-code/nikki-erp/modules/contacts/interfaces/party"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
)

func NewCommChannelServiceImpl(
	commChannelRepo itChannel.CommChannelRepository,
	partySvc itParty.PartyService,
) itChannel.CommChannelService {
	return &CommChannelServiceImpl{
		commChannelRepo: commChannelRepo,
		partySvc:        partySvc,
	}
}

type CommChannelServiceImpl struct {
	commChannelRepo itChannel.CommChannelRepository
	partySvc        itParty.PartyService
}

func (this *CommChannelServiceImpl) CreateCommChannel(ctx corectx.Context, cmd itChannel.CreateCommChannelCommand) (*itChannel.CreateCommChannelResult, error) {
	partyId := cmd.GetPartyId()
	if partyId != nil {
		existsResult, err := this.partySvc.PartyExists(ctx, itParty.PartyExistsQuery{Ids: []model.Id{*partyId}})
		if err != nil {
			return nil, err
		}
		if existsResult.ClientErrors.Count() > 0 {
			return &itChannel.CreateCommChannelResult{ClientErrors: existsResult.ClientErrors}, nil
		}
		if !existsResult.Data.Exists(*partyId) {
			var vErrs ft.ClientErrors
			vErrs.Append(*ft.NewBusinessViolation(
				domain.CommChannelFieldPartyId,
				"comm_channel.party_not_found",
				"party does not exist",
			))
			return &itChannel.CreateCommChannelResult{ClientErrors: vErrs}, nil
		}
	}
	return corecrud.Create(ctx, corecrud.CreateParam[domain.CommChannel, *domain.CommChannel]{
		Action:         "create communication channel",
		BaseRepoGetter: this.commChannelRepo,
		Data:           cmd,
	})
}

func (this *CommChannelServiceImpl) UpdateCommChannel(ctx corectx.Context, cmd itChannel.UpdateCommChannelCommand) (*itChannel.UpdateCommChannelResult, error) {
	return corecrud.Update(ctx, corecrud.UpdateParam[domain.CommChannel, *domain.CommChannel]{
		Action:       "update communication channel",
		DbRepoGetter: this.commChannelRepo,
		Data:         cmd,
	})
}

func (this *CommChannelServiceImpl) DeleteCommChannel(ctx corectx.Context, cmd itChannel.DeleteCommChannelCommand) (*itChannel.DeleteCommChannelResult, error) {
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{
		Action:       "delete communication channel",
		DbRepoGetter: this.commChannelRepo,
		Cmd:          dyn.DeleteOneCommand(cmd),
	})
}

func (this *CommChannelServiceImpl) GetCommChannel(ctx corectx.Context, query itChannel.GetCommChannelQuery) (*itChannel.GetCommChannelResult, error) {
	var id string
	if query.Id != nil {
		id = *query.Id
	}
	return corecrud.GetOne[domain.CommChannel](ctx, corecrud.GetOneParam{
		Action:       "get communication channel",
		DbRepoGetter: this.commChannelRepo,
		Query: dyn.GetOneQuery{
			Id:     id,
			Fields: query.Columns,
		},
	})
}

func (this *CommChannelServiceImpl) SearchCommChannels(ctx corectx.Context, query itChannel.SearchCommChannelsQuery) (*itChannel.SearchCommChannelsResult, error) {
	sanitized, cErrs := query.GetSchema().ValidateStruct(query)
	if cErrs.Count() > 0 {
		return &itChannel.SearchCommChannelsResult{ClientErrors: cErrs}, nil
	}
	query = *(sanitized.(*itChannel.SearchCommChannelsQuery))

	cond := dmodel.NewCondition(domain.CommChannelFieldPartyId, dmodel.Equals, query.PartyId)
	graph := dmodel.NewSearchGraph()
	if query.Graph != nil {
		node := query.Graph.ToSearchNode()
		graph.And(
			*dmodel.NewSearchNode().Condition(cond),
			*node,
		)
	} else {
		graph.Condition(cond)
	}
	return corecrud.Search[domain.CommChannel](ctx, corecrud.SearchParam{
		Action:       "search communication channels",
		DbRepoGetter: this.commChannelRepo,
		Query: dyn.SearchQuery{
			Fields:          query.Fields,
			Graph:           graph,
			Page:            query.Page,
			Size:            query.Size,
			IncludeArchived: query.IncludeArchived,
		},
	})
}

func (this *CommChannelServiceImpl) CommChannelExists(ctx corectx.Context, query itChannel.CommChannelExistsQuery) (*itChannel.CommChannelExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{
		Action:       "check if communication channels exist",
		DbRepoGetter: this.commChannelRepo,
		Query:        dyn.ExistsQuery(query),
	})
}
