package app

import (
	"context"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/contacts/domain"
	it "github.com/sky-as-code/nikki-erp/modules/contacts/interfaces/party"
	"github.com/sky-as-code/nikki-erp/modules/core/event"
	tag "github.com/sky-as-code/nikki-erp/modules/core/tag/interfaces"
)

func NewPartyTagServiceImpl(
	createTagSvc tag.TagServiceFactory,
	eventBus event.EventBus,
) (it.PartyTagService, error) {
	tagSvc, err := createTagSvc(domain.PartyTagType)
	if err != nil {
		return nil, err
	}

	return &PartyTagServiceImpl{
		tagSvc:   tagSvc,
		eventBus: eventBus,
	}, nil
}

type PartyTagServiceImpl struct {
	tagSvc   tag.TagService
	eventBus event.EventBus
}

func (this *PartyTagServiceImpl) TagSvc() tag.TagService {
	return this.tagSvc
}

func (this *PartyTagServiceImpl) CreatePartyTag(ctx context.Context, cmd it.CreatePartyTagCommand) (result *it.CreatePartyTagResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "create party tag"); e != nil {
			err = e
		}
	}()

	result, err = this.tagSvc.CreateTag(ctx, cmd.ToTagCommand())
	ft.PanicOnErr(err)

	return result, err
}

func (this *PartyTagServiceImpl) UpdatePartyTag(ctx context.Context, cmd it.UpdatePartyTagCommand) (result *it.UpdatePartyTagResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "update party tag"); e != nil {
			err = e
		}
	}()

	result, err = this.tagSvc.UpdateTag(ctx, cmd.ToTagCommand())
	ft.PanicOnErr(err)

	return result, err
}

func (this *PartyTagServiceImpl) DeletePartyTag(ctx context.Context, cmd it.DeletePartyTagCommand) (result *it.DeletePartyTagResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "delete party tag"); e != nil {
			err = e
		}
	}()

	result, err = this.tagSvc.DeleteTag(ctx, cmd.ToTagCommand())
	ft.PanicOnErr(err)

	return result, err
}

func (this *PartyTagServiceImpl) PartyTagExistsMulti(ctx context.Context, query it.PartyTagExistsMultiQuery) (result *it.PartyTagExistsMultiResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "check if party tags exist"); e != nil {
			err = e
		}
	}()

	result, err = this.tagSvc.TagExistsMulti(ctx, query.ToTagQuery())
	ft.PanicOnErr(err)

	return result, err
}

func (this *PartyTagServiceImpl) GetPartyTagById(ctx context.Context, query it.GetPartyByIdTagQuery) (result *it.GetPartyTagByIdResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "get party tag"); e != nil {
			err = e
		}
	}()

	result, err = this.tagSvc.GetTagById(ctx, query.ToTagQuery())
	ft.PanicOnErr(err)

	return result, err
}

func (this *PartyTagServiceImpl) ListPartyTags(ctx context.Context, query it.ListPartyTagsQuery) (result *it.ListPartyTagsResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "list party tags"); e != nil {
			err = e
		}
	}()

	result, err = this.tagSvc.ListTags(ctx, query.ToTagQuery())
	ft.PanicOnErr(err)

	return result, err
}
