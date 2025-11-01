package repository

import (
	"time"

	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/core/database"

	domain "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	ent "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent"
	entEntitlement "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/entitlement"
	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/entitlement"
)

func NewEntitlementEntRepository(client *ent.Client) it.EntitlementRepository {
	return &EntitlementEntRepository{
		client: client,
	}
}

type EntitlementEntRepository struct {
	client *ent.Client
}

func (this *EntitlementEntRepository) BeginTransaction(ctx crud.Context) (*ent.Tx, error) {
	return this.client.Tx(ctx)
}

func (this *EntitlementEntRepository) entitlementClient(ctx crud.Context) *ent.EntitlementClient {
	tx, isOk := ctx.GetDbTranx().(*ent.Tx)
	if isOk {
		return tx.Entitlement
	}
	return this.client.Entitlement
}

func (this *EntitlementEntRepository) Create(ctx crud.Context, entitlement *domain.Entitlement) (*domain.Entitlement, error) {
	creation := this.entitlementClient(ctx).Create().
		SetID(*entitlement.Id).
		SetEtag(*entitlement.Etag).
		SetName(*entitlement.Name).
		SetNillableDescription(entitlement.Description).
		SetNillableResourceID(entitlement.ResourceId).
		SetNillableActionID(entitlement.ActionId).
		SetActionExpr(*entitlement.ActionExpr).
		SetNillableOrgID(entitlement.OrgId).
		SetCreatedBy(*entitlement.CreatedBy).
		SetCreatedAt(time.Now())

	return database.Mutate(ctx, creation, ent.IsNotFound, entToEntitlement)
}

func (this *EntitlementEntRepository) Update(ctx crud.Context, entitlement *domain.Entitlement, prevEtag model.Etag) (*domain.Entitlement, error) {
	updation := this.entitlementClient(ctx).UpdateOneID(*entitlement.Id).
		SetNillableDescription(entitlement.Description).
		Where(entEntitlement.EtagEQ(prevEtag))

	if len(updation.Mutation().Fields()) > 0 {
		updation.
			SetEtag(*entitlement.Etag)
	}

	return database.Mutate(ctx, updation, ent.IsNotFound, entToEntitlement)
}

func (this *EntitlementEntRepository) DeleteHard(ctx crud.Context, param it.DeleteParam) (int, error) {
	return this.entitlementClient(ctx).Delete().
		Where(entEntitlement.IDEQ(param.Id)).
		Exec(ctx)
}

func (this *EntitlementEntRepository) Exists(ctx crud.Context, param it.FindByIdParam) (bool, error) {
	return this.entitlementClient(ctx).Query().
		Where(entEntitlement.ID(param.Id)).
		Exist(ctx)
}

func (this *EntitlementEntRepository) FindById(ctx crud.Context, param it.FindByIdParam) (*domain.Entitlement, error) {
	query := this.entitlementClient(ctx).Query().
		Where(entEntitlement.IDEQ(param.Id)).
		WithAction().
		WithResource()

	return database.FindOne(ctx, query, ent.IsNotFound, entToEntitlement)
}

func (this *EntitlementEntRepository) FindByName(ctx crud.Context, param it.FindByNameParam) (*domain.Entitlement, error) {
	query := this.entitlementClient(ctx).Query().
		Where(entEntitlement.NameEQ(param.Name))

	if param.OrgId != nil {
		query = query.Where(entEntitlement.OrgIDEQ(*param.OrgId))
	} else {
		query = query.Where(entEntitlement.OrgIDIsNil())
	}

	return database.FindOne(ctx, query, ent.IsNotFound, entToEntitlement)
}

func (this *EntitlementEntRepository) FindAllByIds(ctx crud.Context, param it.FindAllByIdsParam) ([]domain.Entitlement, error) {
	query := this.entitlementClient(ctx).Query().
		Where(entEntitlement.IDIn(param.Ids...)).
		WithAction().
		WithResource()

	return database.List(ctx, query, entToEntitlements)
}

func (this *EntitlementEntRepository) FindByActionExpr(ctx crud.Context, param it.FindByActionExprParam) (*domain.Entitlement, error) {
	query := this.entitlementClient(ctx).Query().
		Where(entEntitlement.ActionExprEQ(param.ActionExpr))

	if param.OrgId != nil {
		query = query.Where(entEntitlement.OrgIDEQ(*param.OrgId))
	} else {
		query = query.Where(entEntitlement.OrgIDIsNil())
	}

	return database.FindOne(ctx, query, ent.IsNotFound, entToEntitlement)
}

func (this *EntitlementEntRepository) ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, fault.ValidationErrors) {
	return database.ParseSearchGraphStr[ent.Entitlement, domain.Entitlement](criteria, entEntitlement.Label)
}

func (this *EntitlementEntRepository) Search(
	ctx crud.Context,
	param it.SearchParam,
) (*crud.PagedResult[domain.Entitlement], error) {
	query := this.entitlementClient(ctx).Query().
		WithResource().
		WithAction()

	return database.Search(
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
		// Field(entEntitlement.FieldScopeRef, entity.ScopeRef).
		Field(entEntitlement.FieldResourceID, entity.ResourceID).
		Field(entEntitlement.FieldCreatedBy, entity.CreatedBy).
		Field(entEntitlement.FieldCreatedAt, entity.CreatedAt).
		Field(entEntitlement.FieldOrgID, entity.OrgID)

	return builder.Descriptor()
}
