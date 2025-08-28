package repository

import (
	"context"
	"time"

	"github.com/sky-as-code/nikki-erp/common/crud"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/contacts/domain"
	"github.com/sky-as-code/nikki-erp/modules/contacts/infra/ent"
	entCommChannel "github.com/sky-as-code/nikki-erp/modules/contacts/infra/ent/commchannel"
	cc "github.com/sky-as-code/nikki-erp/modules/contacts/interfaces/comm_channel"
	db "github.com/sky-as-code/nikki-erp/modules/core/database"
)

func NewCommChannelEntRepository(client *ent.Client) cc.CommChannelRepository {
	return &CommChannelEntRepository{
		client: client,
	}
}

type CommChannelEntRepository struct {
	client *ent.Client
}

func (this *CommChannelEntRepository) Create(ctx context.Context, commChannel domain.CommChannel) (*domain.CommChannel, error) {
	creation := this.client.CommChannel.Create().
		SetID(*commChannel.Id).
		SetEtag(*commChannel.Etag).
		SetPartyID(string(*commChannel.PartyId)).
		SetNillableNote(commChannel.Note).
		SetNillableValue(commChannel.Value)

	if commChannel.Type != nil && commChannel.Type.Value != nil {
		creation = creation.SetType(entCommChannel.Type(*commChannel.Type.Value))
	}

	if commChannel.ValueJson != nil {
		creation = creation.SetValueJSON(*commChannel.ValueJson)
	}

	return db.Mutate(ctx, creation, ent.IsNotFound, entToCommChannel)
}

func (this *CommChannelEntRepository) Update(ctx context.Context, commChannel domain.CommChannel, prevEtag model.Etag) (*domain.CommChannel, error) {
	update := this.client.CommChannel.UpdateOneID(*commChannel.Id).
		SetNillableNote(commChannel.Note).
		SetNillableValue(commChannel.Value).
		SetNillablePartyID((*string)(commChannel.PartyId)).
		// IMPORTANT: Must have!
		Where(entCommChannel.EtagEQ(prevEtag))

	if commChannel.Type != nil && commChannel.Type.Value != nil {
		update = update.SetType(entCommChannel.Type(*commChannel.Type.Value))
	}

	if commChannel.ValueJson != nil {
		update = update.SetValueJSON(*commChannel.ValueJson)
	}

	if len(update.Mutation().Fields()) > 0 {
		update = update.
			SetEtag(*commChannel.Etag).
			SetUpdatedAt(time.Now())
	}

	return db.Mutate(ctx, update, ent.IsNotFound, entToCommChannel)
}

func (this *CommChannelEntRepository) DeleteHard(ctx context.Context, param cc.DeleteParam) (int, error) {
	return this.client.CommChannel.Delete().
		Where(entCommChannel.ID(param.Id)).
		Exec(ctx)
}

func (this *CommChannelEntRepository) FindById(ctx context.Context, param cc.FindByIdParam) (*domain.CommChannel, error) {
	query := this.client.CommChannel.Query().
		Where(entCommChannel.ID(param.Id))

	if param.WithParty {
		query = query.WithParty()
	}

	return db.FindOne(ctx, query, ent.IsNotFound, entToCommChannel)
}

func (this *CommChannelEntRepository) FindByParty(ctx context.Context, param cc.FindByPartyParam) ([]*domain.CommChannel, error) {
	query := this.client.CommChannel.Query().
		Where(entCommChannel.PartyIDEQ(string(param.PartyId)))

	if param.Type != nil && param.Type.Value != nil {
		query = query.Where(entCommChannel.TypeEQ(entCommChannel.Type(*param.Type.Value)))
	}

	if param.WithParty {
		query = query.WithParty()
	}

	dbEntities, err := query.All(ctx)
	if err != nil {
		return nil, err
	}

	return entToCommChannels(dbEntities), nil
}

func (this *CommChannelEntRepository) ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors) {
	return db.ParseSearchGraphStr[ent.CommChannel, domain.CommChannel](criteria, entCommChannel.Label)
}

func (this *CommChannelEntRepository) Search(
	ctx context.Context,
	param cc.SearchParam,
) (*crud.PagedResult[domain.CommChannel], error) {
	query := this.client.CommChannel.Query()

	if param.WithParty {
		query = query.WithParty()
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
		entToCommChannelsNonPtr,
	)
}

func BuildCommChannelDescriptor() *orm.EntityDescriptor {
	entity := ent.CommChannel{}
	builder := orm.DescribeEntity(entCommChannel.Label).
		Aliases("comm_channels").
		Field(entCommChannel.FieldCreatedAt, entity.CreatedAt).
		Field(entCommChannel.FieldEtag, entity.Etag).
		Field(entCommChannel.FieldID, entity.ID).
		Field(entCommChannel.FieldNote, entity.Note).
		Field(entCommChannel.FieldPartyID, entity.PartyID).
		Field(entCommChannel.FieldType, entity.Type).
		Field(entCommChannel.FieldUpdatedAt, entity.UpdatedAt).
		Field(entCommChannel.FieldValue, entity.Value).
		Field(entCommChannel.FieldValueJSON, entity.ValueJSON).
		Edge(entCommChannel.EdgeParty, orm.ToEdgePredicate(entCommChannel.HasPartyWith))

	return builder.Descriptor()
}
