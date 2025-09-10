package app

import (
	"strings"

	"github.com/sky-as-code/nikki-erp/common/defense"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/contacts/domain"
	cc "github.com/sky-as-code/nikki-erp/modules/contacts/interfaces/comm_channel"
	pt "github.com/sky-as-code/nikki-erp/modules/contacts/interfaces/party"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

func NewCommChannelServiceImpl(
	commChannelRepo cc.CommChannelRepository,
	partyRepo pt.PartyRepository,
	cqrsBus cqrs.CqrsBus,
) cc.CommChannelService {
	return &CommChannelServiceImpl{
		commChannelRepo: commChannelRepo,
		partyRepo:       partyRepo,
		cqrsBus:         cqrsBus,
	}
}

type CommChannelServiceImpl struct {
	commChannelRepo cc.CommChannelRepository
	partyRepo       pt.PartyRepository
	cqrsBus         cqrs.CqrsBus
}

func (this *CommChannelServiceImpl) CreateCommChannel(ctx crud.Context, cmd cc.CreateCommChannelCommand) (result *cc.CreateCommChannelResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to create communication channel"); e != nil {
			err = e
		}
	}()

	commChannel := cmd.ToEntity()
	commChannel.SetDefaults()

	flow := val.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *ft.ValidationErrors) error {
			*vErrs = commChannel.Validate(false)
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			this.sanitizeCommChannel(commChannel)
			return this.assertPartyExists(ctx, *commChannel.PartyId, vErrs)
		}).
		End()
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &cc.CreateCommChannelResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	createdChannel, err := this.commChannelRepo.Create(ctx, *commChannel)
	ft.PanicOnErr(err)

	return &cc.CreateCommChannelResult{Data: createdChannel}, err
}

func (this *CommChannelServiceImpl) sanitizeCommChannel(commChannel *domain.CommChannel) {
	if commChannel.Value != nil {
		cleanedValue := strings.TrimSpace(*commChannel.Value)
		cleanedValue = defense.SanitizePlainText(cleanedValue)
		commChannel.Value = &cleanedValue
	}
}

func (this *CommChannelServiceImpl) UpdateCommChannel(ctx crud.Context, cmd cc.UpdateCommChannelCommand) (result *cc.UpdateCommChannelResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to update communication channel"); e != nil {
			err = e
		}
	}()

	commChannel := cmd.ToEntity()

	flow := val.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *ft.ValidationErrors) error {
			*vErrs = commChannel.Validate(true)
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			return this.assertCorrectCommChannel(ctx, *commChannel.Id, *commChannel.Etag, vErrs)
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			this.sanitizeCommChannel(commChannel)
			return this.assertPartyExists(ctx, *commChannel.PartyId, vErrs)
		}).
		End()
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &cc.UpdateCommChannelResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	prevEtag := commChannel.Etag
	commChannel.Etag = model.NewEtag()
	updatedChannel, err := this.commChannelRepo.Update(ctx, *commChannel, *prevEtag)
	ft.PanicOnErr(err)

	return &cc.UpdateCommChannelResult{
		Data:    updatedChannel,
		HasData: true,
	}, err
}

func (this *CommChannelServiceImpl) DeleteCommChannel(ctx crud.Context, cmd cc.DeleteCommChannelCommand) (result *cc.DeleteCommChannelResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to delete communication channel"); e != nil {
			err = e
		}
	}()

	vErrs := cmd.Validate()
	_, err = this.assertCommChannelIdExists(ctx, cmd.Id, &vErrs)
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &cc.DeleteCommChannelResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	deleted, err := this.commChannelRepo.DeleteHard(ctx, cc.DeleteParam(cmd))
	ft.PanicOnErr(err)

	// build deletion result with count
	deletedCopy := deleted
	return crud.NewSuccessDeletionResult(cmd.Id, &deletedCopy), nil
}

func (this *CommChannelServiceImpl) GetCommChannelById(ctx crud.Context, query cc.GetCommChannelByIdQuery) (result *cc.GetCommChannelByIdResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to get communication channel"); e != nil {
			err = e
		}
	}()

	vErrs := query.Validate()
	dbChannel, err := this.assertCommChannelIdExists(ctx, query.Id, &vErrs)
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &cc.GetCommChannelByIdResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	return &cc.GetCommChannelByIdResult{
		Data: dbChannel,
	}, nil
}

func (this *CommChannelServiceImpl) SearchCommChannels(ctx crud.Context, query cc.SearchCommChannelsQuery) (result *cc.SearchCommChannelsResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to search communication channels"); e != nil {
			err = e
		}
	}()

	vErrsModel := query.Validate()
	predicate, order, vErrsGraph := this.commChannelRepo.ParseSearchGraph(query.Graph)

	vErrsModel.Merge(vErrsGraph)

	if vErrsModel.Count() > 0 {
		return &cc.SearchCommChannelsResult{
			ClientError: vErrsModel.ToClientError(),
		}, nil
	}
	query.SetDefaults()

	channels, err := this.commChannelRepo.Search(ctx, cc.SearchParam{
		Predicate: predicate,
		Order:     order,
		Page:      *query.Page,
		Size:      *query.Size,
		WithParty: query.WithParty,
	})
	ft.PanicOnErr(err)

	return &cc.SearchCommChannelsResult{
		Data: channels,
	}, nil
}

func (this *CommChannelServiceImpl) GetCommChannelsByParty(ctx crud.Context, query cc.GetCommChannelsByPartyQuery) (result *cc.GetCommChannelsByPartyResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to get communication channels by party"); e != nil {
			err = e
		}
	}()

	vErrs := query.Validate()
	err = this.assertPartyExists(ctx, query.PartyId, &vErrs)
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &cc.GetCommChannelsByPartyResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	channels, err := this.commChannelRepo.FindByParty(ctx, cc.FindByPartyParam{
		PartyId:   query.PartyId,
		Type:      query.Type,
		WithParty: query.WithParty,
	})
	ft.PanicOnErr(err)

	return &cc.GetCommChannelsByPartyResult{
		Data: channels,
	}, nil
}

func (this *CommChannelServiceImpl) assertCorrectCommChannel(ctx crud.Context, id model.Id, etag model.Etag, vErrs *ft.ValidationErrors) error {
	dbChannel, err := this.assertCommChannelIdExists(ctx, id, vErrs)
	if err != nil {
		return err
	}

	if dbChannel != nil && *dbChannel.Etag != etag {
		vErrs.Append("etag", "communication channel has been modified by another user")
		return nil
	}

	return nil
}

func (this *CommChannelServiceImpl) assertCommChannelIdExists(ctx crud.Context, id model.Id, vErrs *ft.ValidationErrors) (*domain.CommChannel, error) {
	dbChannel, err := this.commChannelRepo.FindById(ctx, cc.FindByIdParam{
		Id: id,
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

func (this *CommChannelServiceImpl) assertPartyExists(ctx crud.Context, partyId model.Id, vErrs *ft.ValidationErrors) error {
	dbParty, err := this.partyRepo.FindById(ctx, pt.FindByIdParam{
		Id: partyId,
	})
	if err != nil {
		return err
	}

	if dbParty == nil {
		vErrs.Append("partyId", "party not found")
	}
	return nil
}
