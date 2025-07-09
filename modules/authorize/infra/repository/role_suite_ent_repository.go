package repository

import (
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent"
	entRoleSuite "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/rolesuite"
	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/role_suite"
)

func NewRoleSuiteEntRepository(client *ent.Client) it.RoleSuiteRepository {
	return &RoleSuiteEntRepository{
		client: client,
	}
}

type RoleSuiteEntRepository struct {
	client *ent.Client
}

// func (this *ResourceEntRepository) Create(ctx context.Context, resource domain.Resource) (*domain.Resource, error) {
// 	creation := this.client.Resource.Create().
// 		SetID(*resource.Id).
// 		SetName(*resource.Name).
// 		SetDescription(*resource.Description).
// 		SetResourceType(entResource.ResourceType(*resource.ResourceType)).
// 		SetResourceRef(*resource.ResourceRef).
// 		SetScopeType(entResource.ScopeType(*resource.ScopeType)).
// 		SetEtag(*resource.Etag)

// 	return db.Mutate(ctx, creation, entToResource)
// }

// func (this *ResourceEntRepository) FindById(ctx context.Context, param it.FindByIdParam) (*domain.Resource, error) {
// 	query := this.client.Resource.Query().
// 		Where(entResource.IDEQ(param.Id))

// 	return db.FindOne(ctx, query, nil, entToResource)
// }

// func (this *ResourceEntRepository) FindByName(ctx context.Context, param it.FindByNameParam) (*domain.Resource, error) {
// 	query := this.client.Resource.Query().
// 		Where(entResource.NameEQ(param.Name))

// 	return db.FindOne(ctx, query, nil, entToResource)
// }

// func (this *ResourceEntRepository) Update(ctx context.Context, resource domain.Resource) (*domain.Resource, error) {
// 	update := this.client.Resource.UpdateOneID(*resource.Id).
// 		SetDescription(*resource.Description)

// 	if len(update.Mutation().Fields()) > 0 {
// 		update.
// 			SetEtag(*resource.Etag)
// 	}

// 	return db.Mutate(ctx, update, entToResource)
// }

// func (this *ResourceEntRepository) ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors) {
// 	return db.ParseSearchGraphStr[ent.Resource, domain.Resource](criteria, entResource.Label)
// }

// func (this *ResourceEntRepository) Search(
// 	ctx context.Context,
// 	param it.SearchParam,
// ) (*crud.PagedResult[domain.Resource], error) {
// 	query := this.client.Resource.Query()
// 	if param.WithActions {
// 		query = query.WithActions()
// 	}

// 	return db.Search(
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

func BuildRoleSuiteDescriptor() *orm.EntityDescriptor {
	entity := ent.RoleSuite{}
	builder := orm.DescribeEntity(entRoleSuite.Label).
		Aliases("role_suites").
		Field(entRoleSuite.FieldID, entity.ID).
		Field(entRoleSuite.FieldName, entity.Name).
		Field(entRoleSuite.FieldDescription, entity.Description).
		Field(entRoleSuite.FieldEtag, entity.Etag)

	return builder.Descriptor()
}
