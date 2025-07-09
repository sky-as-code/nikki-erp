package repository

import (
	"context"

	"github.com/sky-as-code/nikki-erp/common/crud"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	"github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent"

	entAction "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/action"
	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/action"
	db "github.com/sky-as-code/nikki-erp/modules/core/database"
)

func NewActionEntRepository(client *ent.Client) it.ActionRepository {
	return &ActionEntRepository{
		client: client,
	}
}

type ActionEntRepository struct {
	client *ent.Client
}

func (this *ActionEntRepository) Create(ctx context.Context, action domain.Action) (*domain.Action, error) {
	creation := this.client.Action.Create().
		SetID(*action.Id).
		SetEtag(*action.Etag).
		SetName(*action.Name).
		SetResourceID(*action.ResourceId).
		SetNillableDescription(action.Description).
		SetCreatedBy(*action.CreatedBy)

	return db.Mutate(ctx, creation, entToAction)
}

func (this *ActionEntRepository) FindById(ctx context.Context, param it.FindByIdParam) (*domain.Action, error) {
	query := this.client.Action.Query().
		Where(entAction.IDEQ(param.Id)).
		WithResource()

	return db.FindOne(ctx, query, ent.IsNotFound, entToAction)
}

func (this *ActionEntRepository) FindByName(ctx context.Context, param it.FindByNameParam) (*domain.Action, error) {
	query := this.client.Action.Query().
		Where(entAction.NameEQ(param.Name))

	return db.FindOne(ctx, query, ent.IsNotFound, entToAction)
}

func (this *ActionEntRepository) Update(ctx context.Context, action domain.Action) (*domain.Action, error) {
	update := this.client.Action.UpdateOneID(*action.Id).
		SetDescription(*action.Description)

	if len(update.Mutation().Fields()) > 0 {
		update.
			SetEtag(*action.Etag)
	}

	return db.Mutate(ctx, update, entToAction)
}

func (this *ActionEntRepository) ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors) {
	return db.ParseSearchGraphStr[ent.Action, domain.Action](criteria, entAction.Label)
}

func (this *ActionEntRepository) Search(
	ctx context.Context,
	param it.SearchParam,
) (*crud.PagedResult[domain.Action], error) {
	query := this.client.Action.Query().
		WithResource()

	return db.Search(
		ctx,
		param.Predicate,
		param.Order,
		crud.PagingOptions{
			Page: param.Page,
			Size: param.Size,
		},
		query,
		entToActions,
	)
}

func BuildActionDescriptor() *orm.EntityDescriptor {
	entity := ent.Action{}
	builder := orm.DescribeEntity(entAction.Label).
		Aliases("actions").
		Field(entAction.FieldID, entity.ID).
		Field(entAction.FieldName, entity.Name).
		Field(entAction.FieldDescription, entity.Description).
		Field(entAction.FieldEtag, entity.Etag).
		Field(entAction.FieldResourceID, entity.ResourceID)

	return builder.Descriptor()
}
