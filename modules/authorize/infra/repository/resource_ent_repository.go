package repository

// import (
// 	"time"

// 	"github.com/sky-as-code/nikki-erp/common/fault"
// 	"github.com/sky-as-code/nikki-erp/common/model"
// 	"github.com/sky-as-code/nikki-erp/common/orm"
// 	"github.com/sky-as-code/nikki-erp/modules/core/crud"
// 	"github.com/sky-as-code/nikki-erp/modules/core/database"

// 	domain "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
// 	ent "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent"
// 	entAction "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/action"
// 	entEntitlement "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/entitlement"
// 	entResource "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/resource"
// 	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/resource"
// )

// func NewResourceEntRepository(client *ent.Client) it.ResourceRepository {
// 	return &ResourceEntRepository{
// 		client: client,
// 	}
// }

// // Deprecated: Must create dynamic model repository instead
// type ResourceEntRepository struct {
// 	client *ent.Client
// }

// func (this *ResourceEntRepository) Create(ctx crud.Context, resource *domain.Resource) (*domain.Resource, error) {
// 	creation := this.client.Resource.Create().
// 		SetID(*resource.GetId()).
// 		SetName(*resource.GetName()).
// 		SetNillableDescription(resource.GetDescription()).
// 		SetResourceType(domain.ResourceOwnerType(*resource.GetResourceType()).String()).
// 		SetResourceRef(*resource.GetResourceRef()).
// 		SetScopeType(domain.ResourceScope(*resource.GetScopeType()).String()).
// 		SetEtag(*resource.GetEtag()).
// 		SetCreatedAt(time.Now())

// 	return database.Mutate(ctx, creation, ent.IsNotFound, entToResource)
// }

// func (this *ResourceEntRepository) FindById(ctx crud.Context, param it.FindByIdParam) (*domain.Resource, error) {
// 	query := this.client.Resource.Query().
// 		Where(entResource.IDEQ(param.Id))

// 	return database.FindOne(ctx, query, ent.IsNotFound, entToResource)
// }

// func (this *ResourceEntRepository) FindByName(ctx crud.Context, param it.FindByNameParam) (*domain.Resource, error) {
// 	query := this.client.Resource.Query().
// 		Where(entResource.NameEQ(param.Name))

// 	return database.FindOne(ctx, query, ent.IsNotFound, entToResource)
// }

// func (this *ResourceEntRepository) Update(ctx crud.Context, resource *domain.Resource, prevEtag model.Etag) (*domain.Resource, error) {
// 	update := this.client.Resource.UpdateOneID(*resource.GetId()).
// 		SetNillableDescription(resource.GetDescription()).
// 		Where(entResource.EtagEQ(prevEtag))

// 	if len(update.Mutation().Fields()) > 0 {
// 		update.
// 			SetEtag(*resource.GetEtag())
// 	}

// 	return database.Mutate(ctx, update, ent.IsNotFound, entToResource)
// }

// func (this *ResourceEntRepository) DeleteHard(ctx crud.Context, param it.DeleteParam) (int, error) {
// 	return this.client.Resource.Delete().
// 		Where(entResource.NameEQ(param.Name)).
// 		Exec(ctx)
// }

// func (this *ResourceEntRepository) ListActionNamesByResourceId(ctx crud.Context, resourceId model.Id) ([]string, error) {
// 	rows, err := this.client.Action.Query().Where(entAction.ResourceIDEQ(string(resourceId))).All(ctx)
// 	if err != nil {
// 		return nil, err
// 	}
// 	names := make([]string, 0, len(rows))
// 	for _, row := range rows {
// 		names = append(names, row.Name)
// 	}
// 	return names, nil
// }

// func (this *ResourceEntRepository) ListEntitlementNamesByResourceId(ctx crud.Context, resourceId model.Id) ([]string, error) {
// 	rows, err := this.client.Entitlement.Query().Where(entEntitlement.ResourceIDEQ(string(resourceId))).All(ctx)
// 	if err != nil {
// 		return nil, err
// 	}
// 	names := make([]string, 0, len(rows))
// 	for _, row := range rows {
// 		names = append(names, row.Name)
// 	}
// 	return names, nil
// }

// func (this *ResourceEntRepository) ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, fault.ValidationErrors) {
// 	return database.ParseSearchGraphStr[ent.Resource, domain.Resource](criteria, entResource.Label)
// }

// func (this *ResourceEntRepository) Search(
// 	ctx crud.Context,
// 	param it.SearchParam,
// ) (*crud.PagedResult[domain.Resource], error) {
// 	query := this.client.Resource.Query()
// 	if param.WithActions {
// 		query = query.WithActions()
// 	}

// 	return database.Search(
// 		ctx,
// 		param.Predicate,
// 		param.Order,
// 		crud.PagingOptions{
// 			Page: param.Page,
// 			Size: param.Size,
// 		},
// 		query,
// 		entToResources,
// 	)
// }

// func (this *ResourceEntRepository) Exists(ctx crud.Context, param it.ExistsParam) (bool, error) {
// 	return this.client.Resource.Query().
// 		Where(entResource.IDEQ(param.Id)).
// 		Exist(ctx)
// }

// func BuildResourceDescriptor() *orm.EntityDescriptor {
// 	entity := ent.Resource{}
// 	builder := orm.DescribeEntity(entResource.Label).
// 		Aliases("resources").
// 		Field(entResource.FieldID, entity.ID).
// 		Field(entResource.FieldName, entity.Name).
// 		Field(entResource.FieldDescription, entity.Description).
// 		Field(entResource.FieldResourceType, entity.ResourceType).
// 		Field(entResource.FieldResourceRef, entity.ResourceRef).
// 		Field(entResource.FieldScopeType, entity.ScopeType).
// 		Field(entResource.FieldEtag, entity.Etag).
// 		Field(entResource.FieldCreatedAt, entity.CreatedAt)

// 	return builder.Descriptor()
// }
