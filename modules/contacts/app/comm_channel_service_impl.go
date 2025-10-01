package app

import (
	"github.com/sky-as-code/nikki-erp/common/defense"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/contacts/domain"
	itChannel "github.com/sky-as-code/nikki-erp/modules/contacts/interfaces/comm_channel"
	"github.com/sky-as-code/nikki-erp/modules/contacts/interfaces/contacts_enum"
	itParty "github.com/sky-as-code/nikki-erp/modules/contacts/interfaces/party"
	pt "github.com/sky-as-code/nikki-erp/modules/contacts/interfaces/party"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

func NewCommChannelServiceImpl(
	commChannelRepo itChannel.CommChannelRepository,
	contactsEnumSvc contacts_enum.ContactsEnumService,
	partySvc itParty.PartyService,
	cqrsBus cqrs.CqrsBus,
) itChannel.CommChannelService {
	return &CommChannelServiceImpl{
		commChannelRepo: commChannelRepo,
		contactsEnumSvc: contactsEnumSvc,
		partySvc:        partySvc,
		cqrsBus:         cqrsBus,
	}
}

type CommChannelServiceImpl struct {
	commChannelRepo itChannel.CommChannelRepository
	contactsEnumSvc contacts_enum.ContactsEnumService
	partySvc        itParty.PartyService
	cqrsBus         cqrs.CqrsBus
}

func (this *CommChannelServiceImpl) CreateCommChannel(ctx crud.Context, cmd itChannel.CreateCommChannelCommand) (*itChannel.CreateCommChannelResult, error) {
	result, err := crud.Create(ctx, crud.CreateParam[*domain.CommChannel, itChannel.CreateCommChannelCommand, itChannel.CreateCommChannelResult]{
		Action:              "create communication channel",
		Command:             cmd,
		AssertBusinessRules: this.assertCreateRules,
		RepoCreate:          this.commChannelRepo.Create,
		SetDefault:          this.setCommChannelDefaults,
		Sanitize:            this.sanitizeCommChannel,
		ToFailureResult: func(vErrs *ft.ValidationErrors) *itChannel.CreateCommChannelResult {
			return &itChannel.CreateCommChannelResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.CommChannel) *itChannel.CreateCommChannelResult {
			return &itChannel.CreateCommChannelResult{
				HasData: true,
				Data:    model,
			}
		},
	})

	return result, err
	// defer func() {
	// 	if e := ft.RecoverPanic(recover(), "failed to create communication channel"); e != nil {
	// 		err = e
	// 	}
	// }()

	// commChannel := cmd.ToCommChannel()
	// commChannel.SetDefaults()

	// flow := val.StartValidationFlow()
	// vErrs, err := flow.
	// 	Step(func(vErrs *ft.ValidationErrors) error {
	// 		*vErrs = commChannel.Validate(false)
	// 		return nil
	// 	}).
	// 	Step(func(vErrs *ft.ValidationErrors) error {
	// 		this.sanitizeCommChannel(commChannel)
	// 		return this.assertPartyExists(ctx, *commChannel.PartyId, vErrs)
	// 	}).
	// 	End()
	// ft.PanicOnErr(err)

	// if vErrs.Count() > 0 {
	// 	return &cc.CreateCommChannelResult{
	// 		ClientError: vErrs.ToClientError(),
	// 	}, nil
	// }

	// createdChannel, err := this.commChannelRepo.Create(ctx, *commChannel)
	// ft.PanicOnErr(err)

	// return &cc.CreateCommChannelResult{
	// 	HasData: true,
	// 	Data:    createdChannel,
	// }, err
}

func (this *CommChannelServiceImpl) UpdateCommChannel(ctx crud.Context, cmd itChannel.UpdateCommChannelCommand) (*itChannel.UpdateCommChannelResult, error) {
	result, err := crud.Update(ctx, crud.UpdateParam[*domain.CommChannel, itChannel.UpdateCommChannelCommand, itChannel.UpdateCommChannelResult]{
		Action:              "update communication channel",
		Command:             cmd,
		AssertBusinessRules: this.assertUpdateRules,
		AssertExists:        this.assertCommChannelId,
		RepoUpdate:          this.commChannelRepo.Update,
		Sanitize:            this.sanitizeCommChannel,
		ToFailureResult: func(vErrs *ft.ValidationErrors) *itChannel.UpdateCommChannelResult {
			return &itChannel.UpdateCommChannelResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.CommChannel) *itChannel.UpdateCommChannelResult {
			return &itChannel.UpdateCommChannelResult{
				Data:    model,
				HasData: model != nil,
			}
		},
	})

	return result, err
	// defer func() {
	// 	if e := ft.RecoverPanic(recover(), "failed to update communication channel"); e != nil {
	// 		err = e
	// 	}
	// }()

	// commChannel := cmd.ToCommChannel()

	// flow := val.StartValidationFlow()
	// vErrs, err := flow.
	// 	Step(func(vErrs *ft.ValidationErrors) error {
	// 		*vErrs = commChannel.Validate(true)
	// 		return nil
	// 	}).
	// 	Step(func(vErrs *ft.ValidationErrors) error {
	// 		return this.assertCorrectCommChannel(ctx, *commChannel.Id, *commChannel.Etag, vErrs)
	// 	}).
	// 	Step(func(vErrs *ft.ValidationErrors) error {
	// 		// this.sanitizeCommChannel(commChannel)
	// 		return this.assertPartyExists(ctx, *commChannel.PartyId, vErrs)
	// 	}).
	// 	End()
	// ft.PanicOnErr(err)

	// if vErrs.Count() > 0 {
	// 	return &cc.UpdateCommChannelResult{
	// 		ClientError: vErrs.ToClientError(),
	// 	}, nil
	// }

	// prevEtag := commChannel.Etag
	// commChannel.Etag = model.NewEtag()
	// updatedChannel, err := this.commChannelRepo.Update(ctx, *commChannel, *prevEtag)
	// ft.PanicOnErr(err)

	// return &cc.UpdateCommChannelResult{
	// 	Data:    updatedChannel,
	// 	HasData: true,
	// }, err
}

func (this *CommChannelServiceImpl) DeleteCommChannel(ctx crud.Context, cmd itChannel.DeleteCommChannelCommand) (*itChannel.DeleteCommChannelResult, error) {
	result, err := crud.DeleteHard(ctx, crud.DeleteHardParam[*domain.CommChannel, itChannel.DeleteCommChannelCommand, itChannel.DeleteCommChannelResult]{
		Action:       "delete communication channel",
		Command:      cmd,
		AssertExists: this.assertCommChannelId,
		RepoDelete: func(ctx crud.Context, model *domain.CommChannel) (int, error) {
			return this.commChannelRepo.DeleteHard(ctx, itChannel.DeleteParam{Id: *model.Id})
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *itChannel.DeleteCommChannelResult {
			return &itChannel.DeleteCommChannelResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.CommChannel, deletedCount int) *itChannel.DeleteCommChannelResult {
			return crud.NewSuccessDeletionResult(cmd.Id, &deletedCount)
		},
	})

	return result, err
	// defer func() {
	// 	if e := ft.RecoverPanic(recover(), "failed to delete communication channel"); e != nil {
	// 		err = e
	// 	}
	// }()

	// vErrs := cmd.Validate()
	// _, err = this.assertCommChannelIdExists(ctx, cmd.Id, &vErrs)
	// ft.PanicOnErr(err)

	// if vErrs.Count() > 0 {
	// 	return &cc.DeleteCommChannelResult{
	// 		ClientError: vErrs.ToClientError(),
	// 	}, nil
	// }

	// deleted, err := this.commChannelRepo.DeleteHard(ctx, cc.DeleteParam(cmd))
	// ft.PanicOnErr(err)

	// // build deletion result with count
	// deletedCopy := deleted
	// return crud.NewSuccessDeletionResult(cmd.Id, &deletedCopy), nil
}

func (this *CommChannelServiceImpl) GetCommChannelById(ctx crud.Context, query itChannel.GetCommChannelByIdQuery) (*itChannel.GetCommChannelByIdResult, error) {
	result, err := crud.GetOne(ctx, crud.GetOneParam[*domain.CommChannel, itChannel.GetCommChannelByIdQuery, itChannel.GetCommChannelByIdResult]{
		Action:      "get communication channel by Id",
		Query:       query,
		RepoFindOne: this.getCommChannelByIdFull,
		ToFailureResult: func(vErrs *ft.ValidationErrors) *itChannel.GetCommChannelByIdResult {
			return &itChannel.GetCommChannelByIdResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.CommChannel) *itChannel.GetCommChannelByIdResult {
			return &itChannel.GetCommChannelByIdResult{
				Data:    model,
				HasData: model != nil,
			}
		},
	})

	return result, err
}

func (this *CommChannelServiceImpl) SearchCommChannels(ctx crud.Context, query itChannel.SearchCommChannelsQuery) (*itChannel.SearchCommChannelsResult, error) {
	result, err := crud.Search(ctx, crud.SearchParam[domain.CommChannel, itChannel.SearchCommChannelsQuery, itChannel.SearchCommChannelsResult]{
		Action: "search communication channels",
		Query:  query,
		SetQueryDefaults: func(query *itChannel.SearchCommChannelsQuery) {
			query.SetDefaults()
		},
		ParseSearchGraph: this.commChannelRepo.ParseSearchGraph,
		RepoSearch: func(ctx crud.Context, query itChannel.SearchCommChannelsQuery, predicate *orm.Predicate, order []orm.OrderOption) (*crud.PagedResult[domain.CommChannel], error) {
			return this.commChannelRepo.Search(ctx, itChannel.SearchParam{
				Predicate: predicate,
				Order:     order,
				Page:      *query.Page,
				Size:      *query.Size,
			})
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *itChannel.SearchCommChannelsResult {
			return &itChannel.SearchCommChannelsResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(pagedResult *crud.PagedResult[domain.CommChannel]) *itChannel.SearchCommChannelsResult {
			return &itChannel.SearchCommChannelsResult{
				Data:    pagedResult,
				HasData: pagedResult.Items != nil,
			}
		},
	})

	return result, err
}

// assert methods
//---------------------------------------------------------------------------------------------------------------------------------------------//

func (this *CommChannelServiceImpl) assertCreateRules(ctx crud.Context, channel *domain.CommChannel, vErrs *ft.ValidationErrors) error {
	err := this.assertPartyExists(ctx, channel, vErrs) // check foreign key for Channel.PartyId
	if err != nil {
		return err
	}

	enum, err := this.contactsEnumSvc.GetEnum(ctx, "contacts_channel_type", *channel.Type, vErrs) // check enum exist for Channel.Type
	if err != nil {
		return err
	}

	if enum == nil {
		vErrs.Append("type", "invalid communication channel type")
	}

	return nil
}

func (this *CommChannelServiceImpl) assertUpdateRules(ctx crud.Context, channel *domain.CommChannel, _ *domain.CommChannel, vErrs *ft.ValidationErrors) error {
	err := this.assertCorrectCommChannel(ctx, channel, vErrs) // check exists and etag match for Channel
	if err != nil {
		return err
	}

	err = this.assertPartyExists(ctx, channel, vErrs) // check foreign key for Channel.PartyId
	if err != nil {
		return err
	}

	return nil
}

//---------------------------------------------------------------------------------------------------------------------------------------------//

func (this *CommChannelServiceImpl) sanitizeCommChannel(commChannel *domain.CommChannel) {
	if commChannel.Value != nil {
		cleanedValue := defense.SanitizePlainText(*commChannel.Value, true)
		commChannel.Value = &cleanedValue
	}
}

func (this *CommChannelServiceImpl) setCommChannelDefaults(commChannel *domain.CommChannel) {
	commChannel.SetDefaults()
}

func (this *CommChannelServiceImpl) assertCorrectCommChannel(ctx crud.Context, channel *domain.CommChannel, vErrs *ft.ValidationErrors) error {
	dbChannel, err := this.assertCommChannelId(ctx, channel, vErrs)
	if err != nil {
		return err
	}

	if dbChannel != nil && *dbChannel.Etag != *channel.Etag {
		vErrs.Append("etag", "communication channel wrong etag")
		return nil
	}

	return nil
}

func (this *CommChannelServiceImpl) assertCommChannelId(ctx crud.Context, channel *domain.CommChannel, vErrs *ft.ValidationErrors) (*domain.CommChannel, error) {
	dbChannel, err := this.commChannelRepo.FindById(ctx, itChannel.FindByIdParam{
		Id: *channel.Id,
	})
	if err != nil {
		return nil, err
	}

	if dbChannel == nil {
		vErrs.Append("id", "communication channel not found")
		return nil, nil
	}
	return dbChannel, nil
}

func (this *CommChannelServiceImpl) assertPartyExists(ctx crud.Context, channel *domain.CommChannel, vErrs *ft.ValidationErrors) error {
	dbParty, err := this.partySvc.GetPartyById(ctx, pt.FindByIdParam{
		Id: *channel.PartyId,
	})
	if err != nil {
		return err
	}

	if dbParty.Data == nil {
		vErrs.Append("partyId", "party not found")
	}
	return nil
}

func (this *CommChannelServiceImpl) getCommChannelByIdFull(ctx crud.Context, query itChannel.FindByIdParam, vErrs *ft.ValidationErrors) (dbChannel *domain.CommChannel, err error) {
	dbChannel, err = this.commChannelRepo.FindById(ctx, query)
	if dbChannel == nil {
		vErrs.AppendNotFound("id", "communication channel id")
	}
	return
}
