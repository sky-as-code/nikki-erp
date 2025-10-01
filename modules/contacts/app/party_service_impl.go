package app

import (
	"github.com/sky-as-code/nikki-erp/common/defense"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/contacts/domain"
	"github.com/sky-as-code/nikki-erp/modules/contacts/interfaces/contacts_enum"
	itParty "github.com/sky-as-code/nikki-erp/modules/contacts/interfaces/party"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

func NewPartyServiceImpl(
	partyRepo itParty.PartyRepository,
	cqrsBus cqrs.CqrsBus,
	contactsEnumSvc contacts_enum.ContactsEnumService,
) itParty.PartyService {
	return &PartyServiceImpl{
		partyRepo:       partyRepo,
		cqrsBus:         cqrsBus,
		contactsEnumSvc: contactsEnumSvc,
	}
}

type PartyServiceImpl struct {
	partyRepo       itParty.PartyRepository
	cqrsBus         cqrs.CqrsBus
	contactsEnumSvc contacts_enum.ContactsEnumService
}

func (this *PartyServiceImpl) CreateParty(ctx crud.Context, cmd itParty.CreatePartyCommand) (*itParty.CreatePartyResult, error) {
	result, err := crud.Create(ctx, crud.CreateParam[*domain.Party, itParty.CreatePartyCommand, itParty.CreatePartyResult]{
		Action:              "create party",
		Command:             cmd,
		AssertBusinessRules: this.assertCreateRules,
		RepoCreate:          this.partyRepo.Create,
		SetDefault:          this.setPartyDefaults,
		Sanitize:            this.sanitizeParty,
		ToFailureResult: func(vErrs *ft.ValidationErrors) *itParty.CreatePartyResult {
			return &itParty.CreatePartyResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.Party) *itParty.CreatePartyResult {
			return &itParty.CreatePartyResult{
				HasData: true,
				Data:    model,
			}
		},
	})

	return result, err
}

func (this *PartyServiceImpl) UpdateParty(ctx crud.Context, cmd itParty.UpdatePartyCommand) (*itParty.UpdatePartyResult, error) {
	result, err := crud.Update(ctx, crud.UpdateParam[*domain.Party, itParty.UpdatePartyCommand, itParty.UpdatePartyResult]{
		Action:              "update party",
		Command:             cmd,
		AssertBusinessRules: this.assertUpdateRules,
		AssertExists:        this.assertPartyIdExists,
		RepoUpdate:          this.partyRepo.Update,
		Sanitize:            this.sanitizeParty,
		ToFailureResult: func(vErrs *ft.ValidationErrors) *itParty.UpdatePartyResult {
			return &itParty.UpdatePartyResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.Party) *itParty.UpdatePartyResult {
			return &itParty.UpdatePartyResult{
				HasData: true,
				Data:    model,
			}
		},
	})

	return result, err
}

func (this *PartyServiceImpl) DeleteParty(ctx crud.Context, cmd itParty.DeletePartyCommand) (*itParty.DeletePartyResult, error) {
	result, err := crud.DeleteHard(ctx, crud.DeleteHardParam[*domain.Party, itParty.DeletePartyCommand, itParty.DeletePartyResult]{
		Action:       "delete party",
		Command:      cmd,
		AssertExists: this.assertPartyIdExists,
		RepoDelete: func(ctx crud.Context, model *domain.Party) (int, error) {
			return this.partyRepo.DeleteHard(ctx, itParty.DeleteParam{Id: *model.Id})
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *itParty.DeletePartyResult {
			return &itParty.DeletePartyResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.Party, deletedCount int) *itParty.DeletePartyResult {
			return crud.NewSuccessDeletionResult(cmd.Id, &deletedCount)
		},
	})

	return result, err
}

func (this *PartyServiceImpl) GetPartyById(ctx crud.Context, query itParty.GetPartyByIdQuery) (*itParty.GetPartyByIdResult, error) {
	result, err := crud.GetOne(ctx, crud.GetOneParam[*domain.Party, itParty.GetPartyByIdQuery, itParty.GetPartyByIdResult]{
		Action:      "get party by id",
		Query:       query,
		RepoFindOne: this.getPartyByIdFull,
		ToFailureResult: func(vErrs *ft.ValidationErrors) *itParty.GetPartyByIdResult {
			return &itParty.GetPartyByIdResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.Party) *itParty.GetPartyByIdResult {
			return &itParty.GetPartyByIdResult{
				HasData: true,
				Data:    model,
			}
		},
	})

	return result, err
}

func (this *PartyServiceImpl) SearchParties(ctx crud.Context, query itParty.SearchPartiesQuery) (*itParty.SearchPartiesResult, error) {
	result, err := crud.Search(ctx, crud.SearchParam[domain.Party, itParty.SearchPartiesQuery, itParty.SearchPartiesResult]{
		Action: "search parties",
		Query:  query,
		SetQueryDefaults: func(query *itParty.SearchPartiesQuery) {
			query.SetDefaults()
		},
		ParseSearchGraph: this.partyRepo.ParseSearchGraph,
		RepoSearch: func(ctx crud.Context, query itParty.SearchPartiesQuery, predicate *orm.Predicate, order []orm.OrderOption) (*crud.PagedResult[domain.Party], error) {
			return this.partyRepo.Search(ctx, itParty.SearchParam{
				Predicate: predicate,
				Order:     order,
				Page:      *query.Page,
				Size:      *query.Size,
			})
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *itParty.SearchPartiesResult {
			return &itParty.SearchPartiesResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(pagedResult *crud.PagedResult[domain.Party]) *itParty.SearchPartiesResult {
			return &itParty.SearchPartiesResult{
				Data:    pagedResult,
				HasData: pagedResult.Items != nil,
			}
		},
	})

	return result, err
}

// assert methods
//---------------------------------------------------------------------------------------------------------------------------------------------//

func (this *PartyServiceImpl) assertCreateRules(ctx crud.Context, party *domain.Party, vErrs *ft.ValidationErrors) error {
	err := this.assertUniquePartyDisplayName(ctx, party, vErrs) // display name must be unique
	if err != nil {
		return err
	}

	if party.Title != nil {
		err = this.assertEnumParty(ctx, "contacts_party_title", *party.Title, vErrs) // title must exist
		if err != nil {
			return err
		}
	}

	return nil
}

func (this *PartyServiceImpl) assertUpdateRules(ctx crud.Context, party *domain.Party, _ *domain.Party, vErrs *ft.ValidationErrors) error {

	err := this.assertUniquePartyDisplayName(ctx, party, vErrs) // display name must be unique
	if err != nil {
		return err
	}

	if party.Title != nil {
		err = this.assertEnumParty(ctx, "contacts_party_title", *party.Title, vErrs) // title must exist
		if err != nil {
			return err
		}
	}

	return nil
}

// ---------------------------------------------------------------------------------------------------------------------------------------------//

func (this *PartyServiceImpl) setPartyDefaults(party *domain.Party) {
	party.SetDefaults()
}

func (this *PartyServiceImpl) sanitizeParty(party *domain.Party) {
	if party.DisplayName != nil {
		cleanedName := defense.SanitizePlainText(*party.DisplayName, true)
		party.DisplayName = &cleanedName
	}
	if party.LegalName != nil {
		cleanedName := defense.SanitizePlainText(*party.LegalName, true)
		party.LegalName = &cleanedName
	}
	if party.Note != nil {
		cleanedNote := defense.SanitizePlainText(*party.Note, true)
		party.Note = &cleanedNote
	}
}

func (this *PartyServiceImpl) assertPartyIdExists(ctx crud.Context, party *domain.Party, vErrs *ft.ValidationErrors) (*domain.Party, error) {
	dbParty, err := this.partyRepo.FindById(ctx, itParty.FindByIdParam{
		Id: *party.Id,
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

func (this *PartyServiceImpl) assertUniquePartyDisplayName(ctx crud.Context, party *domain.Party, vErrs *ft.ValidationErrors) error {
	if party.DisplayName == nil {
		return nil
	}

	dbParty, err := this.partyRepo.FindByDisplayName(ctx, itParty.FindByDisplayNameParam{
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

func (this *PartyServiceImpl) assertEnumParty(ctx crud.Context, typeEnum, valueEnum string, vErrs *ft.ValidationErrors) error {
	enum, err := this.contactsEnumSvc.GetEnum(ctx, typeEnum, valueEnum, vErrs)
	if err != nil {
		return err
	}

	if enum.Data == nil {
		vErrs.Append("title", "title does not exist")
		return nil
	}
	return nil
}

func (this *PartyServiceImpl) getPartyByIdFull(ctx crud.Context, query itParty.FindByIdParam, vErrs *ft.ValidationErrors) (dbParty *domain.Party, err error) {
	dbParty, err = this.partyRepo.FindById(ctx, query)
	if dbParty == nil {
		vErrs.AppendNotFound("id", "party id")
	}
	return
}

// func (this *PartyServiceImpl) assertNationalityExists(ctx crud.Context, id model.Id, valErrs *ft.ValidationErrors) error {
// 	if id == "" {
// 		return nil
// 	}

// 	existCmd := &eif.EnumExistsQuery{
// 		Id:         id,
// 		EntityName: "nationality",
// 	}

// 	existRes := eif.EnumExistsResult{}

// 	err := this.cqrsBus.Request(ctx, existCmd, &existRes)
// 	ft.PanicOnErr(err)

// 	if existRes.ClientError != nil {
// 		valErrs.MergeClientError(existRes.ClientError)
// 		return nil
// 	}

// 	if existRes.Data == true {
// 		valErrs.Append("nationality", "nationality already exists")
// 	}

// 	return nil
// }

// func (this *PartyServiceImpl) assertLanguageExists(ctx crud.Context, id model.Id, valErrs *ft.ValidationErrors) error {
// 	if id == "" {
// 		return nil
// 	}

// 	existCmd := &eif.EnumExistsQuery{
// 		Id:         id,
// 		EntityName: "language",
// 	}

// 	existRes := eif.EnumExistsResult{}

// 	err := this.cqrsBus.Request(ctx, existCmd, &existRes)
// 	ft.PanicOnErr(err)

// 	if existRes.ClientError != nil {
// 		valErrs.MergeClientError(existRes.ClientError)
// 		return nil
// 	}

// 	if existRes.Data == true {
// 		valErrs.Append("language", "language already exists")
// 	}

// 	return nil
// }
