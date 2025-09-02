package app

import (
	"github.com/sky-as-code/nikki-erp/common/defense"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/contacts/domain"
	pt "github.com/sky-as-code/nikki-erp/modules/contacts/interfaces/party"
	rel "github.com/sky-as-code/nikki-erp/modules/contacts/interfaces/relationship"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

func NewRelationshipServiceImpl(
	relationshipRepo rel.RelationshipRepository,
	partyRepo pt.PartyRepository,
	cqrsBus cqrs.CqrsBus,
) rel.RelationshipService {
	return &RelationshipServiceImpl{
		relationshipRepo: relationshipRepo,
		partyRepo:        partyRepo,
		cqrsBus:          cqrsBus,
	}
}

type RelationshipServiceImpl struct {
	relationshipRepo rel.RelationshipRepository
	partyRepo        pt.PartyRepository
	cqrsBus          cqrs.CqrsBus
}

func (this *RelationshipServiceImpl) CreateRelationship(ctx crud.Context, cmd rel.CreateRelationshipCommand) (result *rel.CreateRelationshipResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to create relationship"); e != nil {
			err = e
		}
	}()

	relationship := cmd.ToEntity()
	relationship.SetDefaults()

	flow := val.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *ft.ValidationErrors) error {
			*vErrs = relationship.Validate(false)
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			this.sanitizeRelationship(relationship)
			return this.assertPartiesExist(ctx, *relationship.TargetPartyId, *relationship.TargetPartyId, vErrs)
		}).
		End()
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &rel.CreateRelationshipResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	createdRelationship, err := this.relationshipRepo.Create(ctx, *relationship)
	ft.PanicOnErr(err)

	return &rel.CreateRelationshipResult{Data: createdRelationship}, err
}

func (this *RelationshipServiceImpl) sanitizeRelationship(relationship *domain.Relationship) {
	if relationship.Note != nil {
		cleaned := defense.SanitizePlainText(*relationship.Note)
		relationship.Note = &cleaned
	}
}

func (this *RelationshipServiceImpl) UpdateRelationship(ctx crud.Context, cmd rel.UpdateRelationshipCommand) (result *rel.UpdateRelationshipResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to update relationship"); e != nil {
			err = e
		}
	}()

	relationship := cmd.ToEntity()

	flow := val.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *ft.ValidationErrors) error {
			*vErrs = relationship.Validate(true)
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			return this.assertCorrectRelationship(ctx, *relationship.Id, *relationship.Etag, vErrs)
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			this.sanitizeRelationship(relationship)
			if relationship.TargetPartyId != nil {
				return this.assertPartyExists(ctx, *relationship.TargetPartyId, vErrs)
			}
			return nil
		}).
		End()
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &rel.UpdateRelationshipResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	prevEtag := relationship.Etag
	relationship.Etag = model.NewEtag()
	updatedRelationship, err := this.relationshipRepo.Update(ctx, *relationship, *prevEtag)
	ft.PanicOnErr(err)

	return &rel.UpdateRelationshipResult{
		Data:    updatedRelationship,
		HasData: true,
	}, err
}

func (this *RelationshipServiceImpl) DeleteRelationship(ctx crud.Context, cmd rel.DeleteRelationshipCommand) (result *rel.DeleteRelationshipResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to delete relationship"); e != nil {
			err = e
		}
	}()

	vErrs := cmd.Validate()
	_, err = this.assertRelationshipIdExists(ctx, cmd.Id, &vErrs)
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &rel.DeleteRelationshipResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	deleted, err := this.relationshipRepo.DeleteHard(ctx, rel.DeleteParam(cmd))
	ft.PanicOnErr(err)
	deletedCopy := deleted
	return crud.NewSuccessDeletionResult(cmd.Id, &deletedCopy), nil
}

func (this *RelationshipServiceImpl) GetRelationshipById(ctx crud.Context, query rel.GetRelationshipByIdQuery) (result *rel.GetRelationshipByIdResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to get relationship"); e != nil {
			err = e
		}
	}()

	vErrs := query.Validate()
	dbRelationship, err := this.assertRelationshipIdExists(ctx, query.Id, &vErrs)
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &rel.GetRelationshipByIdResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	return &rel.GetRelationshipByIdResult{
		Data: dbRelationship,
	}, nil
}

func (this *RelationshipServiceImpl) SearchRelationships(ctx crud.Context, query rel.SearchRelationshipsQuery) (result *rel.SearchRelationshipsResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to search relationships"); e != nil {
			err = e
		}
	}()

	vErrsModel := query.Validate()
	predicate, order, vErrsGraph := this.relationshipRepo.ParseSearchGraph(query.Graph)

	vErrsModel.Merge(vErrsGraph)

	if vErrsModel.Count() > 0 {
		return &rel.SearchRelationshipsResult{
			ClientError: vErrsModel.ToClientError(),
		}, nil
	}
	query.SetDefaults()

	relationships, err := this.relationshipRepo.Search(ctx, rel.SearchParam{
		Predicate: predicate,
		Order:     order,
		Page:      *query.Page,
		Size:      *query.Size,
	})
	ft.PanicOnErr(err)

	return &rel.SearchRelationshipsResult{
		Data: relationships,
	}, nil
}

func (this *RelationshipServiceImpl) GetRelationshipsByParty(ctx crud.Context, query rel.GetRelationshipsByPartyQuery) (result *rel.GetRelationshipsByPartyResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to get relationships by party"); e != nil {
			err = e
		}
	}()

	vErrs := query.Validate()
	err = this.assertPartyExists(ctx, query.PartyId, &vErrs)
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &rel.GetRelationshipsByPartyResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	relationships, err := this.relationshipRepo.FindByParty(ctx, rel.FindByPartyParam{
		PartyId: query.PartyId,
		Type:    query.Type,
	})
	ft.PanicOnErr(err)

	return &rel.GetRelationshipsByPartyResult{
		Data: relationships,
	}, nil
}

func (this *RelationshipServiceImpl) assertCorrectRelationship(ctx crud.Context, id model.Id, etag model.Etag, vErrs *ft.ValidationErrors) error {
	dbRelationship, err := this.assertRelationshipIdExists(ctx, id, vErrs)
	if err != nil {
		return err
	}

	if dbRelationship != nil && *dbRelationship.Etag != etag {
		vErrs.Append("etag", "relationship has been modified by another user")
		return nil
	}

	return nil
}

func (this *RelationshipServiceImpl) assertRelationshipIdExists(ctx crud.Context, id model.Id, vErrs *ft.ValidationErrors) (*domain.Relationship, error) {
	dbRelationship, err := this.relationshipRepo.FindById(ctx, rel.FindByIdParam{
		Id: id,
	})
	if err != nil {
		return nil, err
	}

	if dbRelationship == nil {
		vErrs.Append("id", "relationship not found")
		return nil, nil
	}
	return dbRelationship, nil
}

func (this *RelationshipServiceImpl) assertPartiesExist(ctx crud.Context, partyFromId, partyToId model.Id, vErrs *ft.ValidationErrors) error {
	err := this.assertPartyExists(ctx, partyFromId, vErrs)
	if err != nil {
		return err
	}

	return this.assertPartyToExists(ctx, partyToId, vErrs)
}

func (this *RelationshipServiceImpl) assertPartyExists(ctx crud.Context, partyId model.Id, vErrs *ft.ValidationErrors) error {
	dbParty, err := this.partyRepo.FindById(ctx, pt.FindByIdParam{
		Id: partyId,
	})
	if err != nil {
		return err
	}

	if dbParty == nil {
		vErrs.Append("partyFromId", "party not found")
	}
	return nil
}

func (this *RelationshipServiceImpl) assertPartyToExists(ctx crud.Context, partyId model.Id, vErrs *ft.ValidationErrors) error {
	dbParty, err := this.partyRepo.FindById(ctx, pt.FindByIdParam{
		Id: partyId,
	})
	if err != nil {
		return err
	}

	if dbParty == nil {
		vErrs.Append("partyToId", "party not found")
	}
	return nil
}
