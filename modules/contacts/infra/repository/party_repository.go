package repository

import (
	"time"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/contacts/domain"
	"github.com/sky-as-code/nikki-erp/modules/contacts/infra/ent"
	entParty "github.com/sky-as-code/nikki-erp/modules/contacts/infra/ent/party"
	pt "github.com/sky-as-code/nikki-erp/modules/contacts/interfaces/party"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	db "github.com/sky-as-code/nikki-erp/modules/core/database"
)

func NewPartyEntRepository(client *ent.Client) pt.PartyRepository {
	return &PartyEntRepository{
		client: client,
	}
}

type PartyEntRepository struct {
	client *ent.Client
}

func (this *PartyEntRepository) Create(ctx crud.Context, party *domain.Party) (*domain.Party, error) {
	creation := this.client.Party.Create().
		SetID(*party.Id).
		SetEtag(*party.Etag).
		SetDisplayName(*party.DisplayName).
		SetType(*party.Type).
		SetNillableAvatarUrl(party.AvatarUrl).
		SetNillableLegalName(party.LegalName).
		SetNillableLegalAddress(party.LegalAddress).
		SetNillableTaxID(party.TaxId).
		SetNillableTitle(party.Title).
		SetNillableJobPosition(party.JobPosition).
		SetNillableNote(party.Note).
		SetOrgID(*party.OrgId).
		SetNillableWebsite(party.Website)

	if party.Language != nil {
		creation = creation.SetNillableLanguageID((*string)(party.Language))
	}

	if party.Nationality != nil {
		creation = creation.SetNillableNationalityID((*string)(party.Nationality))
	}

	return db.Mutate(ctx, creation, ent.IsNotFound, entToParty)
}

func (this *PartyEntRepository) Update(ctx crud.Context, party *domain.Party, prevEtag model.Etag) (*domain.Party, error) {
	update := this.client.Party.UpdateOneID(*party.Id).
		SetNillableAvatarUrl(party.AvatarUrl).
		SetNillableDisplayName(party.DisplayName).
		SetNillableLegalName(party.LegalName).
		SetNillableLegalAddress(party.LegalAddress).
		SetNillableTaxID(party.TaxId).
		SetNillableJobPosition(party.JobPosition).
		SetNillableType(party.Type).
		SetNillableTitle(party.Title).
		SetNillableNote(party.Note).
		SetNillableWebsite(party.Website).
		// IMPORTANT: Must have!
		Where(entParty.EtagEQ(prevEtag))

	if party.Language != nil {
		update = update.SetNillableLanguageID((*string)(party.Language))
	}

	if party.Nationality != nil {
		update = update.SetNillableNationalityID((*string)(party.Nationality))
	}

	if len(update.Mutation().Fields()) > 0 {
		update = update.
			SetEtag(*party.Etag).
			SetUpdatedAt(time.Now())
	}

	return db.Mutate(ctx, update, ent.IsNotFound, entToParty)
}

func (this *PartyEntRepository) DeleteHard(ctx crud.Context, param pt.DeleteParam) (int, error) {
	return this.client.Party.Delete().
		Where(entParty.ID(param.Id)).
		Exec(ctx)
}

func (this *PartyEntRepository) FindById(ctx crud.Context, param pt.FindByIdParam) (*domain.Party, error) {
	query := this.client.Party.Query().
		Where(entParty.ID(param.Id))

	if param.WithCommChannels {
		query = query.WithCommChannels()
	}

	if param.WithRelationships {
		query = query.WithRelationshipsAsSource()
	}

	return db.FindOne(ctx, query, ent.IsNotFound, entToParty)
}

func (this *PartyEntRepository) FindByDisplayName(ctx crud.Context, param pt.FindByDisplayNameParam) (*domain.Party, error) {
	query := this.client.Party.Query().
		Where(entParty.DisplayName(param.DisplayName))

	if param.WithCommChannels {
		query = query.WithCommChannels()
	}

	if param.WithRelationships {
		query = query.WithRelationshipsAsSource()
	}

	return db.FindOne(ctx, query, ent.IsNotFound, entToParty)
}

func (this *PartyEntRepository) ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors) {
	return db.ParseSearchGraphStr[ent.Party, domain.Party](criteria, entParty.Label)
}

func (this *PartyEntRepository) Search(
	ctx crud.Context,
	param pt.SearchParam,
) (*crud.PagedResult[domain.Party], error) {
	query := this.client.Party.Query()

	if param.WithCommChannels {
		query = query.WithCommChannels()
	}

	if param.WithRelationships {
		query = query.WithRelationshipsAsSource()
	}

	return db.Search(
		ctx,
		param.Predicate,
		param.Order,
		crud.PagingOptions{
			Page: param.Page,
			Size: param.Size,
		},
		query,
		entToPartiesNonPtr,
	)
}

func BuildPartyDescriptor() *orm.EntityDescriptor {
	entity := ent.Party{}
	builder := orm.DescribeEntity(entParty.Label).
		Aliases("parties").
		Field(entParty.FieldAvatarUrl, entity.AvatarUrl).
		Field(entParty.FieldCreatedAt, entity.CreatedAt).
		Field(entParty.FieldDisplayName, entity.DisplayName).
		Field(entParty.FieldEtag, entity.Etag).
		Field(entParty.FieldID, entity.ID).
		Field(entParty.FieldJobPosition, entity.JobPosition).
		Field(entParty.FieldLegalAddress, entity.LegalAddress).
		Field(entParty.FieldLegalName, entity.LegalName).
		Field(entParty.FieldNote, entity.Note).
		Field(entParty.FieldTaxID, entity.TaxID).
		Field(entParty.FieldTitle, entity.Title).
		Field(entParty.FieldType, entity.Type).
		Field(entParty.FieldUpdatedAt, entity.UpdatedAt).
		Field(entParty.FieldWebsite, entity.Website).
		Edge(entParty.EdgeCommChannels, orm.ToEdgePredicate(entParty.HasCommChannelsWith)).
		Edge(entParty.EdgeRelationshipsAsSource, orm.ToEdgePredicate(entParty.HasRelationshipsAsSourceWith))

	return builder.Descriptor()
}
