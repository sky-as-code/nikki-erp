package repository

import (
	"context"

	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	"github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent"
	entEntitlement "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/entitlement"
	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/entitlement"
	db "github.com/sky-as-code/nikki-erp/modules/core/database"
)

func NewEntitlementEntRepository(client *ent.Client) it.EntitlementRepository {
	return &EntitlementEntRepository{
		client: client,
	}
}

type EntitlementEntRepository struct {
	client *ent.Client
}

func (this *EntitlementEntRepository) Create(ctx context.Context, entitlement domain.Entitlement) (*domain.Entitlement, error) {
	creation := this.client.Entitlement.Create().
		SetID(*entitlement.Id).
		SetEtag(*entitlement.Etag).
		SetCreatedAt(*entitlement.CreatedAt).
		SetNillableActionID(entitlement.ActionId).
		SetActionExpr(*entitlement.ActionExpr).
		SetName(*entitlement.Name).
		SetNillableDescription(entitlement.Description).
		// SetSubjectType(entEntitlement.SubjectType(*entitlement.SubjectType)).
		// SetSubjectRef(*entitlement.SubjectRef).
		SetNillableResourceID(entitlement.ResourceId).
		SetNillableScopeRef(entitlement.ScopeRef).
		SetCreatedBy(*entitlement.CreatedBy)

	return db.Mutate(ctx, creation, entToEntitlement)
}

func (this *EntitlementEntRepository) FindByName(ctx context.Context, param it.FindByNameParam) (*domain.Entitlement, error) {
	query := this.client.Entitlement.Query().
		Where(entEntitlement.NameEQ(param.Name))

	return db.FindOne(ctx, query, ent.IsNotFound, entToEntitlement)
}

// func (this *EntitlementEntRepository) getUserEffectiveEntitlements(ctx context.Context, subject domain.Subject) ([]domain.Entitlement, error) {
// func (this *EntitlementEntRepository) getUserEffectiveEntitlements(ctx context.Context, userId model.Id) ([]domain.Entitlement, error) {
// 	effectiveEnts, err := this.client.EffectiveEntitlement.
// 		Query().
// 		Where(entEff.UserIDEQ(userId.String())).
// 		All(ctx)
// }

func BuildEntitlementDescriptor() *orm.EntityDescriptor {
	entity := ent.Entitlement{}
	builder := orm.DescribeEntity(entEntitlement.Label).
		Aliases("entitlements").
		Field(entEntitlement.FieldID, entity.ID).
		Field(entEntitlement.FieldActionID, entity.ActionID).
		Field(entEntitlement.FieldActionExpr, entity.ActionExpr).
		Field(entEntitlement.FieldName, entity.Name).
		Field(entEntitlement.FieldDescription, entity.Description).
		// Field(entEntitlement.FieldSubjectType, entity.SubjectType).
		// Field(entEntitlement.FieldSubjectRef, entity.SubjectRef).
		Field(entEntitlement.FieldScopeRef, entity.ScopeRef).
		Field(entEntitlement.FieldResourceID, entity.ResourceID).
		Field(entEntitlement.FieldEtag, entity.Etag).
		Field(entEntitlement.FieldCreatedAt, entity.CreatedAt)

	return builder.Descriptor()
}
