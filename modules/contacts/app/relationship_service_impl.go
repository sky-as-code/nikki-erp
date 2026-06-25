package app

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/contacts/domain"
	itParty "github.com/sky-as-code/nikki-erp/modules/contacts/interfaces/party"
	itRelationship "github.com/sky-as-code/nikki-erp/modules/contacts/interfaces/relationship"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
)

func NewRelationshipServiceImpl(
	relationshipRepo itRelationship.RelationshipRepository,
	partySvc itParty.PartyService,
) itRelationship.RelationshipService {
	return &RelationshipServiceImpl{
		relationshipRepo: relationshipRepo,
		partySvc:         partySvc,
	}
}

type RelationshipServiceImpl struct {
	relationshipRepo itRelationship.RelationshipRepository
	partySvc         itParty.PartyService
}

func (this *RelationshipServiceImpl) CreateRelationship(ctx corectx.Context, cmd itRelationship.CreateRelationshipCommand) (*itRelationship.CreateRelationshipResult, error) {
	ids := make([]model.Id, 0, 2)
	if partyId := cmd.GetPartyId(); partyId != nil {
		ids = append(ids, *partyId)
	}
	if targetPartyId := cmd.GetTargetPartyId(); targetPartyId != nil {
		ids = append(ids, *targetPartyId)
	}
	if len(ids) > 0 {
		existsResult, err := this.partySvc.PartyExists(ctx, itParty.PartyExistsQuery{Ids: ids})
		if err != nil {
			return nil, err
		}
		if existsResult.ClientErrors.Count() > 0 {
			return &itRelationship.CreateRelationshipResult{ClientErrors: existsResult.ClientErrors}, nil
		}
		var vErrs ft.ClientErrors
		if partyId := cmd.GetPartyId(); partyId != nil && !existsResult.Data.Exists(*partyId) {
			vErrs.Append(*ft.NewBusinessViolation(
				domain.RelationshipFieldPartyId,
				"relationship.party_not_found",
				"party does not exist",
			))
		}
		if targetPartyId := cmd.GetTargetPartyId(); targetPartyId != nil && !existsResult.Data.Exists(*targetPartyId) {
			vErrs.Append(*ft.NewBusinessViolation(
				domain.RelationshipFieldTargetPartyId,
				"relationship.target_party_not_found",
				"target party does not exist",
			))
		}
		if vErrs.Count() > 0 {
			return &itRelationship.CreateRelationshipResult{ClientErrors: vErrs}, nil
		}
	}
	return corecrud.Create(ctx, corecrud.CreateParam[domain.Relationship, *domain.Relationship]{
		Action:         "create relationship",
		BaseRepoGetter: this.relationshipRepo,
		Data:           cmd,
	})
}

func (this *RelationshipServiceImpl) UpdateRelationship(ctx corectx.Context, cmd itRelationship.UpdateRelationshipCommand) (*itRelationship.UpdateRelationshipResult, error) {
	return corecrud.Update(ctx, corecrud.UpdateParam[domain.Relationship, *domain.Relationship]{
		Action:       "update relationship",
		DbRepoGetter: this.relationshipRepo,
		Data:         cmd,
	})
}

func (this *RelationshipServiceImpl) DeleteRelationship(ctx corectx.Context, cmd itRelationship.DeleteRelationshipCommand) (*itRelationship.DeleteRelationshipResult, error) {
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{
		Action:       "delete relationship",
		DbRepoGetter: this.relationshipRepo,
		Cmd:          dyn.DeleteOneCommand(cmd),
	})
}

func (this *RelationshipServiceImpl) GetRelationship(ctx corectx.Context, query itRelationship.GetRelationshipQuery) (*itRelationship.GetRelationshipResult, error) {
	var id string
	if query.Id != nil {
		id = *query.Id
	}
	return corecrud.GetOne[domain.Relationship](ctx, corecrud.GetOneParam{
		Action:       "get relationship",
		DbRepoGetter: this.relationshipRepo,
		Query: dyn.GetOneQuery{
			Id:     id,
			Fields: query.Columns,
		},
	})
}

func (this *RelationshipServiceImpl) SearchRelationships(ctx corectx.Context, query itRelationship.SearchRelationshipsQuery) (*itRelationship.SearchRelationshipsResult, error) {
	sanitized, cErrs := query.GetSchema().ValidateStruct(query)
	if cErrs.Count() > 0 {
		return &itRelationship.SearchRelationshipsResult{ClientErrors: cErrs}, nil
	}
	query = *(sanitized.(*itRelationship.SearchRelationshipsQuery))

	cond := dmodel.NewCondition(domain.RelationshipFieldPartyId, dmodel.Equals, query.PartyId)
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
	return corecrud.Search[domain.Relationship](ctx, corecrud.SearchParam{
		Action:       "search relationships",
		DbRepoGetter: this.relationshipRepo,
		Query: dyn.SearchQuery{
			Fields: query.Fields,
			Graph:  graph,
			Page:   query.Page,
			Size:   query.Size,
		},
	})
}

func (this *RelationshipServiceImpl) RelationshipExists(ctx corectx.Context, query itRelationship.RelationshipExistsQuery) (*itRelationship.RelationshipExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{
		Action:       "check if relationships exist",
		DbRepoGetter: this.relationshipRepo,
		Query:        dyn.ExistsQuery(query),
	})
}
