package repository

import (
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/contacts/domain"
	"github.com/sky-as-code/nikki-erp/modules/contacts/infra/ent"
	entRelationship "github.com/sky-as-code/nikki-erp/modules/contacts/infra/ent/relationship"
	rel "github.com/sky-as-code/nikki-erp/modules/contacts/interfaces/relationship"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
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

func (this *RelationshipEntRepository) Create(ctx crud.Context, relationship *domain.Relationship) (*domain.Relationship, error) {
	creation := this.client.Relationship.Create().
		SetID(*relationship.Id).
		SetEtag(*relationship.Etag).
		SetType(*relationship.Type).
		SetPartyID(*relationship.PartyId).
		SetTargetPartyID(*relationship.TargetPartyId).
		SetNillableNote(relationship.Note)

	return db.Mutate(ctx, creation, ent.IsNotFound, entToRelationship)
}

func (this *RelationshipEntRepository) FindById(ctx crud.Context, param rel.FindByIdParam) (*domain.Relationship, error) {
	query := this.client.Relationship.Query().
		Where(entRelationship.ID(param.Id))

	return db.FindOne(ctx, query, ent.IsNotFound, entToRelationship)
}

func (this *RelationshipEntRepository) FindByPartyIds(ctx crud.Context, param rel.FindByPartyIdsParam) (*domain.Relationship, error) {
	query := this.client.Relationship.Query().
		Where(entRelationship.PartyID(param.PartyId)).
		Where(entRelationship.TargetPartyID(param.TargetPartyID)).
		Where(entRelationship.Type(param.Type))

	return db.FindOne(ctx, query, ent.IsNotFound, entToRelationship)

}

func (this *RelationshipEntRepository) FindByParty(ctx crud.Context, param rel.FindByPartyParam) ([]*domain.Relationship, error) {
	query := this.client.Relationship.Query().
		Where(entRelationship.TargetPartyIDEQ(string(param.PartyId)))

	if param.Type != nil && param.Type.Value != nil {
		query = query.Where(entRelationship.Type(*param.Type.Value))
	}

	dbEntities, err := query.All(ctx)
	if err != nil {
		return nil, err
	}

	return entToRelationships(dbEntities), nil
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
