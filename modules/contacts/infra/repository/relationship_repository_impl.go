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
	entRelationship "github.com/sky-as-code/nikki-erp/modules/contacts/infra/ent/relationship"
	rel "github.com/sky-as-code/nikki-erp/modules/contacts/interfaces/relationship"
	db "github.com/sky-as-code/nikki-erp/modules/core/database"
)

func NewRelationshipEntRepository(client *ent.Client) rel.RelationshipRepository {
	return &RelationshipEntRepository{
		client: client,
	}
}

type RelationshipEntRepository struct {
	client *ent.Client
}

func (this *RelationshipEntRepository) Create(ctx context.Context, relationship domain.Relationship) (*domain.Relationship, error) {
	creation := this.client.Relationship.Create().
		SetID(*relationship.Id).
		SetEtag(*relationship.Etag).
		SetTargetPartyID(*relationship.TargetPartyId)

	if relationship.Type != nil && relationship.Type.Value != nil {
		creation = creation.SetType(entRelationship.Type(*relationship.Type.Value))
	}

	if relationship.Note != nil {
		creation = creation.SetNote(*relationship.Note)
	}

	return db.Mutate(ctx, creation, ent.IsNotFound, entToRelationship)
}

func (this *RelationshipEntRepository) Update(ctx context.Context, relationship domain.Relationship, prevEtag model.Etag) (*domain.Relationship, error) {
	update := this.client.Relationship.UpdateOneID(*relationship.Id).
		SetNillableNote(relationship.Note).
		SetNillableTargetPartyID((*string)(relationship.TargetPartyId)).
		// IMPORTANT: Must have!
		Where(entRelationship.EtagEQ(prevEtag))

	if relationship.Type != nil && relationship.Type.Value != nil {
		update = update.SetType(entRelationship.Type(*relationship.Type.Value))
	}

	if len(update.Mutation().Fields()) > 0 {
		update = update.
			SetEtag(*relationship.Etag).
			SetUpdatedAt(time.Now())
	}

	return db.Mutate(ctx, update, ent.IsNotFound, entToRelationship)
}

func (this *RelationshipEntRepository) DeleteHard(ctx context.Context, param rel.DeleteParam) (int, error) {
	return this.client.Relationship.Delete().
		Where(entRelationship.ID(param.Id)).
		Exec(ctx)
}

func (this *RelationshipEntRepository) FindById(ctx context.Context, param rel.FindByIdParam) (*domain.Relationship, error) {
	query := this.client.Relationship.Query().
		Where(entRelationship.ID(param.Id))

	return db.FindOne(ctx, query, ent.IsNotFound, entToRelationship)
}

func (this *RelationshipEntRepository) FindByParty(ctx context.Context, param rel.FindByPartyParam) ([]*domain.Relationship, error) {
	query := this.client.Relationship.Query().
		Where(entRelationship.TargetPartyIDEQ(string(param.PartyId)))

	if param.Type != nil && param.Type.Value != nil {
		query = query.Where(entRelationship.TypeEQ(entRelationship.Type(*param.Type.Value)))
	}

	dbEntities, err := query.All(ctx)
	if err != nil {
		return nil, err
	}

	return entToRelationships(dbEntities), nil
}

func (this *RelationshipEntRepository) ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors) {
	return db.ParseSearchGraphStr[ent.Relationship, domain.Relationship](criteria, entRelationship.Label)
}

func (this *RelationshipEntRepository) Search(
	ctx context.Context,
	param rel.SearchParam,
) (*crud.PagedResult[domain.Relationship], error) {
	query := this.client.Relationship.Query()

	return db.Search(
		ctx,
		param.Predicate,
		param.Order,
		crud.PagingOptions{
			Page: param.Page,
			Size: param.Size,
		},
		query,
		entToRelationshipsNonPtr,
	)
}

func BuildRelationshipDescriptor() *orm.EntityDescriptor {
	entity := ent.Relationship{}
	builder := orm.DescribeEntity(entRelationship.Label).
		Aliases("relationships").
		Field(entRelationship.FieldCreatedAt, entity.CreatedAt).
		Field(entRelationship.FieldEtag, entity.Etag).
		Field(entRelationship.FieldID, entity.ID).
		Field(entRelationship.FieldNote, entity.Note).
		Field(entRelationship.FieldTargetPartyID, entity.TargetPartyID).
		Field(entRelationship.FieldType, entity.Type).
		Field(entRelationship.FieldUpdatedAt, entity.UpdatedAt)

	return builder.Descriptor()
}
