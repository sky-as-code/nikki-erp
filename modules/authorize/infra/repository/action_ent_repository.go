package repository

import (
	"context"
	"time"

	"github.com/sky-as-code/nikki-erp/common/crud"
	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/database"

	domain "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	ent "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent"
	entAction "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/action"
	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/action"
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
		SetCreatedBy(*action.CreatedBy).
		SetCreatedAt(time.Now())

	return database.Mutate(ctx, creation, ent.IsNotFound, entToAction)
}

func (this *ActionEntRepository) FindById(ctx context.Context, param it.FindByIdParam) (*domain.Action, error) {
	query := this.client.Action.Query().
		Where(entAction.IDEQ(param.Id)).
		WithResource()

	return database.FindOne(ctx, query, ent.IsNotFound, entToAction)
}

func (this *ActionEntRepository) FindByName(ctx context.Context, param it.FindByNameParam) (*domain.Action, error) {
	query := this.client.Action.Query().
		Where(entAction.NameEQ(param.Name)).
		Where(entAction.ResourceIDEQ(param.ResourceId))

	return database.FindOne(ctx, query, ent.IsNotFound, entToAction)
}

func (this *ActionEntRepository) Update(ctx context.Context, action domain.Action, prevEtag model.Etag) (*domain.Action, error) {
	update := this.client.Action.UpdateOneID(*action.Id).
		SetNillableDescription(action.Description).
		Where(entAction.EtagEQ(prevEtag))

	if len(update.Mutation().Fields()) > 0 {
		update.
			SetEtag(*action.Etag)
	}

	return database.Mutate(ctx, update, ent.IsNotFound, entToAction)
}

func (this *ActionEntRepository) ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, fault.ValidationErrors) {
	return database.ParseSearchGraphStr[ent.Action, domain.Action](criteria, entAction.Label)
}

func (this *ActionEntRepository) Search(
	ctx context.Context,
	param it.SearchParam,
) (*crud.PagedResult[domain.Action], error) {
	query := this.client.Action.Query().
		WithResource()

	return database.Search(
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
		Field(entAction.FieldResourceID, entity.ResourceID).
		Field(entAction.FieldCreatedAt, entity.CreatedAt)

	return builder.Descriptor()
}
