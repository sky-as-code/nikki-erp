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
	entResource "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/resource"
	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/resource"
)

func NewResourceEntRepository(client *ent.Client) it.ResourceRepository {
	return &ResourceEntRepository{
		client: client,
	}
}

type ResourceEntRepository struct {
	client *ent.Client
}

func (this *ResourceEntRepository) Create(ctx context.Context, resource domain.Resource) (*domain.Resource, error) {
	creation := this.client.Resource.Create().
		SetID(*resource.Id).
		SetName(*resource.Name).
		SetNillableDescription(resource.Description).
		SetResourceType(entResource.ResourceType(*resource.ResourceType)).
		SetResourceRef(*resource.ResourceRef).
		SetScopeType(entResource.ScopeType(*resource.ScopeType)).
		SetEtag(*resource.Etag).
		SetCreatedAt(time.Now())

	return database.Mutate(ctx, creation, ent.IsNotFound, entToResource)
}

func (this *ResourceEntRepository) FindById(ctx context.Context, param it.FindByIdParam) (*domain.Resource, error) {
	query := this.client.Resource.Query().
		Where(entResource.IDEQ(param.Id))

	return database.FindOne(ctx, query, ent.IsNotFound, entToResource)
}

func (this *ResourceEntRepository) FindByName(ctx context.Context, param it.FindByNameParam) (*domain.Resource, error) {
	query := this.client.Resource.Query().
		Where(entResource.NameEQ(param.Name)).
		WithActions()

	return database.FindOne(ctx, query, ent.IsNotFound, entToResource)
}

func (this *ResourceEntRepository) Update(ctx context.Context, resource domain.Resource, prevEtag model.Etag) (*domain.Resource, error) {
	update := this.client.Resource.UpdateOneID(*resource.Id).
		SetDescription(*resource.Description).
		Where(entResource.EtagEQ(prevEtag))

	if len(update.Mutation().Fields()) > 0 {
		update.
			SetEtag(*resource.Etag)
	}

	return database.Mutate(ctx, update, ent.IsNotFound, entToResource)
}

func (this *ResourceEntRepository) ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, fault.ValidationErrors) {
	return database.ParseSearchGraphStr[ent.Resource, domain.Resource](criteria, entResource.Label)
}

func (this *ResourceEntRepository) Search(
	ctx context.Context,
	param it.SearchParam,
) (*crud.PagedResult[domain.Resource], error) {
	query := this.client.Resource.Query()
	if param.WithActions {
		query = query.WithActions()
	}

	return database.Search(
		ctx,
		param.Predicate,
		param.Order,
		crud.PagingOptions{
			Page: param.Page,
			Size: param.Size,
		},
		query,
		entToResources,
	)
}

func (this *ResourceEntRepository) Exist(ctx context.Context, param it.ExistParam) (bool, error) {
	return this.client.Resource.Query().
		Where(entResource.IDEQ(param.Id)).
		Exist(ctx)
}

func BuildResourceDescriptor() *orm.EntityDescriptor {
	entity := ent.Resource{}
	builder := orm.DescribeEntity(entResource.Label).
		Aliases("resources").
		Field(entResource.FieldID, entity.ID).
		Field(entResource.FieldName, entity.Name).
		Field(entResource.FieldDescription, entity.Description).
		Field(entResource.FieldResourceType, entity.ResourceType).
		Field(entResource.FieldResourceRef, entity.ResourceRef).
		Field(entResource.FieldScopeType, entity.ScopeType).
		Field(entResource.FieldEtag, entity.Etag).
		Field(entResource.FieldCreatedAt, entity.CreatedAt)

	return builder.Descriptor()
}
