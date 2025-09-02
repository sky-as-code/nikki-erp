package app

import (
	"strings"

	"github.com/sky-as-code/nikki-erp/common/defense"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/contacts/domain"
	it "github.com/sky-as-code/nikki-erp/modules/contacts/interfaces/party"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

func NewPartyServiceImpl(
	partyTagSvc it.PartyTagService,
	partyRepo it.PartyRepository,
	cqrsBus cqrs.CqrsBus,
) it.PartyService {
	return &PartyServiceImpl{
		PartyTagService: partyTagSvc,
		partyRepo:       partyRepo,
		cqrsBus:         cqrsBus,
	}
}

type PartyServiceImpl struct {
	it.PartyTagService
	partyRepo it.PartyRepository
	cqrsBus   cqrs.CqrsBus
}

func (ps *PartyServiceImpl) CreateParty(ctx crud.Context, cmd it.CreatePartyCommand) (result *it.CreatePartyResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to create party"); e != nil {
			err = e
		}
	}()

	party := cmd.ToEntity()
	party.SetDefaults()

	flow := val.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *ft.ValidationErrors) error {
			*vErrs = party.Validate(false)
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			ps.sanitizeParty(party)
			return ps.assertUniquePartyDisplayName(ctx, party, vErrs)
		}).
		End()
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &it.CreatePartyResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	createdParty, err := ps.partyRepo.Create(ctx, *party)
	ft.PanicOnErr(err)

	return &it.CreatePartyResult{Data: createdParty}, err
}

func (ps *PartyServiceImpl) sanitizeParty(party *domain.Party) {
	if party.DisplayName != nil {
		cleanedName := strings.TrimSpace(*party.DisplayName)
		cleanedName = defense.SanitizePlainText(cleanedName)
		party.DisplayName = &cleanedName
	}
	if party.LegalName != nil {
		cleanedName := strings.TrimSpace(*party.LegalName)
		cleanedName = defense.SanitizePlainText(cleanedName)
		party.LegalName = &cleanedName
	}
	if party.Note != nil {
		cleanedNote := strings.TrimSpace(*party.Note)
		cleanedNote = defense.SanitizePlainText(cleanedNote)
		party.Note = &cleanedNote
	}
}

func (ps *PartyServiceImpl) UpdateParty(ctx crud.Context, cmd it.UpdatePartyCommand) (result *it.UpdatePartyResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to update party"); e != nil {
			err = e
		}
	}()

	party := cmd.ToEntity()

	flow := val.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *ft.ValidationErrors) error {
			*vErrs = party.Validate(true)
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			return ps.assertCorrectParty(ctx, *party.Id, *party.Etag, vErrs)
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			ps.sanitizeParty(party)
			return ps.assertUniquePartyDisplayName(ctx, party, vErrs)
		}).
		End()
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &it.UpdatePartyResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	prevEtag := party.Etag
	party.Etag = model.NewEtag()
	updatedParty, err := ps.partyRepo.Update(ctx, *party, *prevEtag)
	ft.PanicOnErr(err)

	return &it.UpdatePartyResult{
		Data:    updatedParty,
		HasData: true,
	}, err
}

func (ps *PartyServiceImpl) DeleteParty(ctx crud.Context, cmd it.DeletePartyCommand) (result *it.DeletePartyResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to delete party"); e != nil {
			err = e
		}
	}()

	vErrs := cmd.Validate()
	_, err = ps.assertPartyIdExists(ctx, cmd.Id, &vErrs)
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &it.DeletePartyResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	deleted, err := ps.partyRepo.DeleteHard(ctx, it.DeleteParam(cmd))
	ft.PanicOnErr(err)
	deletedCopy := deleted
	return crud.NewSuccessDeletionResult(cmd.Id, &deletedCopy), nil
}

func (ps *PartyServiceImpl) GetPartyById(ctx crud.Context, query it.GetPartyByIdQuery) (result *it.GetPartyByIdResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to get party"); e != nil {
			err = e
		}
	}()

	vErrs := query.Validate()
	dbParty, err := ps.assertPartyIdExists(ctx, query.Id, &vErrs)
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &it.GetPartyByIdResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	return &it.GetPartyByIdResult{
		Data: dbParty,
	}, nil
}

func (ps *PartyServiceImpl) SearchParties(ctx crud.Context, query it.SearchPartiesQuery) (result *it.SearchPartiesResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to search parties"); e != nil {
			err = e
		}
	}()

	vErrsModel := query.Validate()
	predicate, order, vErrsGraph := ps.partyRepo.ParseSearchGraph(query.Graph)

	vErrsModel.Merge(vErrsGraph)

	if vErrsModel.Count() > 0 {
		return &it.SearchPartiesResult{
			ClientError: vErrsModel.ToClientError(),
		}, nil
	}
	query.SetDefaults()

	parties, err := ps.partyRepo.Search(ctx, it.SearchParam{
		Predicate:         predicate,
		Order:             order,
		Page:              *query.Page,
		Size:              *query.Size,
		WithCommChannels:  query.WithCommChannels,
		WithRelationships: query.WithRelationships,
	})
	ft.PanicOnErr(err)

	return &it.SearchPartiesResult{
		Data: parties,
	}, nil
}

func (ps *PartyServiceImpl) assertCorrectParty(ctx crud.Context, id model.Id, etag model.Etag, vErrs *ft.ValidationErrors) error {
	dbParty, err := ps.assertPartyIdExists(ctx, id, vErrs)
	if err != nil {
		return err
	}

	if dbParty != nil && *dbParty.Etag != etag {
		vErrs.Append("etag", "party has been modified by another user")
		return nil
	}

	return nil
}

func (ps *PartyServiceImpl) assertPartyIdExists(ctx crud.Context, id model.Id, vErrs *ft.ValidationErrors) (*domain.Party, error) {
	dbParty, err := ps.partyRepo.FindById(ctx, it.FindByIdParam{
		Id: id,
	})
	if err != nil {
		return nil, err
	}

	if dbParty == nil {
		vErrs.Append("id", "party not found")
		return nil, nil
	}
	return dbParty, nil
}

func (ps *PartyServiceImpl) assertUniquePartyDisplayName(ctx crud.Context, party *domain.Party, vErrs *ft.ValidationErrors) error {
	if party.DisplayName == nil {
		return nil
	}

	dbParty, err := ps.partyRepo.FindByDisplayName(ctx, it.FindByDisplayNameParam{
		DisplayName: *party.DisplayName,
	})
	if err != nil {
		return err
	}

	if dbParty != nil && (party.Id == nil || *dbParty.Id != *party.Id) {
		vErrs.Append("displayName", "party display name already exists")
	}
	return nil
}
