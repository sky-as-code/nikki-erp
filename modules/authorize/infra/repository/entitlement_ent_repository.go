package repository

import (
	"context"

	"github.com/sky-as-code/nikki-erp/common/crud"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
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
		SetName(*entitlement.Name).
		SetNillableDescription(entitlement.Description).
		SetNillableResourceID(entitlement.ResourceId).
		SetNillableActionID(entitlement.ActionId).
		SetNillableScopeRef(entitlement.ScopeRef).
		SetActionExpr(*entitlement.ActionExpr).
		SetCreatedBy(*entitlement.CreatedBy)

	return db.Mutate(ctx, creation, ent.IsNotFound, entToEntitlement)
}

func (this *EntitlementEntRepository) Update(ctx context.Context, entitlement domain.Entitlement, prevEtag model.Etag) (*domain.Entitlement, error) {
	updation := this.client.Entitlement.UpdateOneID(*entitlement.Id).
		SetEtag(*entitlement.Etag).
		SetNillableDescription(entitlement.Description).
		Where(entEntitlement.EtagEQ(prevEtag))

	if len(updation.Mutation().Fields()) > 0 {
		updation.
			SetEtag(*entitlement.Etag)
	}

	return db.Mutate(ctx, updation, ent.IsNotFound, entToEntitlement)
}

func (this *EntitlementEntRepository) Exists(ctx context.Context, param it.FindByIdParam) (bool, error) {
	return this.client.Entitlement.Query().
		Where(entEntitlement.ID(param.Id)).
		Exist(ctx)
}

func (this *EntitlementEntRepository) FindById(ctx context.Context, param it.FindByIdParam) (*domain.Entitlement, error) {
	query := this.client.Entitlement.Query().
		Where(entEntitlement.IDEQ(param.Id)).
		WithAction().
		WithResource()

	return db.FindOne(ctx, query, ent.IsNotFound, entToEntitlement)
}

func (this *EntitlementEntRepository) FindByName(ctx context.Context, param it.FindByNameParam) (*domain.Entitlement, error) {
	query := this.client.Entitlement.Query().
		Where(entEntitlement.NameEQ(param.Name))

	return db.FindOne(ctx, query, ent.IsNotFound, entToEntitlement)
}

func (this *EntitlementEntRepository) FindAllByIds(ctx context.Context, param it.FindAllByIdsParam) ([]*domain.Entitlement, error) {
	query := this.client.Entitlement.Query().
		Where(entEntitlement.IDIn(param.Ids...))

	return db.List(ctx, query, entToEntitlements)
}

func (this *EntitlementEntRepository) ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors) {
	return db.ParseSearchGraphStr[ent.Entitlement, domain.Entitlement](criteria, entEntitlement.Label)
}

func (this *EntitlementEntRepository) Search(
	ctx context.Context,
	param it.SearchParam,
) (*crud.PagedResult[*domain.Entitlement], error) {
	query := this.client.Entitlement.Query().
		WithResource().
		WithAction()

	return db.Search(
		ctx,
		param.Predicate,
		param.Order,
		crud.PagingOptions{
			Page: param.Page,
			Size: param.Size,
		},
		query,
		entToEntitlements,
	)
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
		Field(entEntitlement.FieldEtag, entity.Etag).
		Field(entEntitlement.FieldCreatedAt, entity.CreatedAt).
		Field(entEntitlement.FieldName, entity.Name).
		Field(entEntitlement.FieldDescription, entity.Description).
		Field(entEntitlement.FieldActionID, entity.ActionID).
		Field(entEntitlement.FieldActionExpr, entity.ActionExpr).
		Field(entEntitlement.FieldScopeRef, entity.ScopeRef).
		Field(entEntitlement.FieldResourceID, entity.ResourceID).
		Field(entEntitlement.FieldCreatedBy, entity.CreatedBy)

	return builder.Descriptor()
}
