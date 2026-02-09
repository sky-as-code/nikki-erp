package repository

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	db "github.com/sky-as-code/nikki-erp/modules/core/database"
	"github.com/sky-as-code/nikki-erp/modules/inventory/infra/ent"
	entAttributeGroup "github.com/sky-as-code/nikki-erp/modules/inventory/infra/ent/attributegroup"
	"github.com/sky-as-code/nikki-erp/modules/inventory/product/domain"
	itAttributeGroup "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/attributegroup"
)

func NewAttributeGroupEntRepository(client *ent.Client) itAttributeGroup.AttributeGroupRepository {
	return &AttributeGroupEntRepository{
		client: client,
	}
}

type AttributeGroupEntRepository struct {
	client *ent.Client
}

func (r *AttributeGroupEntRepository) attributeGroupClient(ctx crud.Context) *ent.AttributeGroupClient {
	tx, isOk := ctx.GetDbTranx().(*ent.Tx)
	if isOk {
		return tx.AttributeGroup
	}
	return r.client.AttributeGroup
}

// ✅ Create AttributeGroup
func (r *AttributeGroupEntRepository) Create(ctx crud.Context, attributeGroup *domain.AttributeGroup) (*domain.AttributeGroup, error) {
	creation := r.client.AttributeGroup.Create().
		SetID(*attributeGroup.Id).
		SetName(*attributeGroup.Name).
		SetIndex(*attributeGroup.Index).
		SetNillableProductID(attributeGroup.ProductId)

	return db.Mutate(ctx, creation, ent.IsNotFound, itAttributeGroup.EntToAttributeGroup)
}

// ✅ Update AttributeGroup
func (r *AttributeGroupEntRepository) Update(ctx crud.Context, attributeGroup *domain.AttributeGroup, prevEtag model.Etag) (*domain.AttributeGroup, error) {
	update := r.client.AttributeGroup.UpdateOneID(*attributeGroup.Id).
		SetName(*attributeGroup.Name).
		SetIndex(*attributeGroup.Index).
		SetNillableProductID(attributeGroup.ProductId)
		// Note: Etag field doesn't exist in the schema, but keeping pattern for consistency

	return db.Mutate(ctx, update, ent.IsNotFound, itAttributeGroup.EntToAttributeGroup)
}

// ✅ Delete AttributeGroup by ID
func (r *AttributeGroupEntRepository) DeleteById(ctx crud.Context, id model.Id) (int, error) {
	return r.client.AttributeGroup.Delete().
		Where(entAttributeGroup.ID(id)).
		Exec(ctx)
}

// ✅ Find by ID
func (r *AttributeGroupEntRepository) FindById(ctx crud.Context, query itAttributeGroup.FindByIdParam) (*domain.AttributeGroup, error) {
	dbQuery := r.attributeGroupClient(ctx).Query().
		Where(entAttributeGroup.ID(query.Id))

	return db.FindOne(ctx, dbQuery, ent.IsNotFound, itAttributeGroup.EntToAttributeGroup)
}

// ✅ Search (advanced)
func (r *AttributeGroupEntRepository) Search(ctx crud.Context, param itAttributeGroup.SearchParam) (*crud.PagedResult[domain.AttributeGroup], error) {
	query := r.client.AttributeGroup.Query()

	// Add ProductId filter if provided
	if param.ProductId != nil {
		query = query.Where(entAttributeGroup.ProductID(*param.ProductId))
	}

	return db.Search(
		ctx,
		param.Predicate,
		param.Order,
		crud.PagingOptions{
			Page: param.Page,
			Size: param.Size,
		},
		query,
		itAttributeGroup.EntToAttributeGroups,
	)
}

func (r *AttributeGroupEntRepository) ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors) {
	return db.ParseSearchGraphStr[ent.AttributeGroup, domain.AttributeGroup](criteria, entAttributeGroup.Label)
}

func BuildAttributeGroupDescriptor() *orm.EntityDescriptor {
	entity := ent.AttributeGroup{}
	builder := orm.DescribeEntity(entAttributeGroup.Label).
		Aliases("attributegroups", "attribute_groups").
		Field(entAttributeGroup.FieldCreatedAt, entity.CreatedAt).
		Field(entAttributeGroup.FieldID, entity.ID).
		Field(entAttributeGroup.FieldName, entity.Name).
		Field(entAttributeGroup.FieldIndex, entity.Index).
		Field(entAttributeGroup.FieldProductID, entity.ProductID).
		Field(entAttributeGroup.FieldUpdatedAt, entity.UpdatedAt)

	return builder.Descriptor()
}
