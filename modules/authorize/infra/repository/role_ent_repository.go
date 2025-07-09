package repository

import (
	"context"

	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	"github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent"
	entRole "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/role"
	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/role"
	db "github.com/sky-as-code/nikki-erp/modules/core/database"
)

func NewRoleEntRepository(client *ent.Client) it.RoleRepository {
	return &RoleEntRepository{
		client: client,
	}
}

type RoleEntRepository struct {
	client *ent.Client
}

func (this *RoleEntRepository) Create(ctx context.Context, role domain.Role) (*domain.Role, error) {
	creation := this.client.Role.Create().
		SetID(*role.Id).
		SetEtag(*role.Etag).
		SetName(*role.Name).
		SetNillableDescription(role.Description).
		SetOwnerType(entRole.OwnerType(*role.OwnerType)).
		SetOwnerRef(*role.OwnerRef).
		SetIsRequestable(*role.IsRequestable).
		SetIsRequiredAttachment(*role.IsRequiredAttachment).
		SetIsRequiredComment(*role.IsRequiredComment).
		SetCreatedBy(*role.CreatedBy)

	return db.Mutate(ctx, creation, entToRole)
}

// func (this *ResourceEntRepository) FindById(ctx context.Context, param it.FindByIdParam) (*domain.Resource, error) {
// 	query := this.client.Resource.Query().
// 		Where(entResource.IDEQ(param.Id))

// 	return db.FindOne(ctx, query, nil, entToResource)
// }

func (this *RoleEntRepository) FindByName(ctx context.Context, param it.FindByNameParam) (*domain.Role, error) {
	query := this.client.Role.Query().
		Where(entRole.NameEQ(param.Name))

	return db.FindOne(ctx, query, ent.IsNotFound, entToRole)
}

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

func BuildRoleDescriptor() *orm.EntityDescriptor {
	entity := ent.Role{}
	builder := orm.DescribeEntity(entRole.Label).
		Aliases("roles").
		Field(entRole.FieldID, entity.ID).
		Field(entRole.FieldEtag, entity.Etag).
		Field(entRole.FieldName, entity.Name).
		Field(entRole.FieldDescription, entity.Description).
		Field(entRole.FieldOwnerType, entity.OwnerType).
		Field(entRole.FieldOwnerRef, entity.OwnerRef).
		Field(entRole.FieldIsRequestable, entity.IsRequestable).
		Field(entRole.FieldIsRequiredAttachment, entity.IsRequiredAttachment).
		Field(entRole.FieldIsRequiredComment, entity.IsRequiredComment).
		Field(entRole.FieldCreatedBy, entity.CreatedBy).
		Field(entRole.FieldCreatedAt, entity.CreatedAt)

	return builder.Descriptor()
}
