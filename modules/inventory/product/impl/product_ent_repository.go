package impl

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	db "github.com/sky-as-code/nikki-erp/modules/core/database"
	"github.com/sky-as-code/nikki-erp/modules/inventory/infra/ent"
	entProduct "github.com/sky-as-code/nikki-erp/modules/inventory/infra/ent/product"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces"
)

func NewProductEntRepository(client *ent.Client) itProduct.ProductRepository {
	return &ProductEntRepository{
		client: client,
	}
}

type ProductEntRepository struct {
	client *ent.Client
}

func (r *ProductEntRepository) productClient(ctx crud.Context) *ent.ProductClient {
	tx, isOk := ctx.GetDbTranx().(*ent.Tx)
	if isOk {
		return tx.Product
	}
	return r.client.Product
}

// ✅ Create Product
func (r *ProductEntRepository) Create(ctx crud.Context, product *itProduct.Product) (*itProduct.Product, error) {
	creation := r.client.Product.Create().
		SetID(*product.Id).
		SetOrgID(*product.OrgId).
		SetName(*product.Name).
		SetUnitID(*product.Unit).
		SetNillableStatus(product.Status).
		SetNillableDefaultVariantID(product.DefaultsVariantId).
		SetNillableThumbnailURL(product.ThumbnailUrl).
		SetEtag(*product.Etag)
		// SetNillableTagIDs(product.TagIds)

	if product.Description != nil {
		creation.SetDescription(*product.Description)
	}

	return db.Mutate(ctx, creation, ent.IsNotFound, itProduct.EntToProduct)
	// return nil, nil
}

// ✅ Update Product
func (r *ProductEntRepository) Update(ctx crud.Context, product *itProduct.Product, prevEtag model.Etag) (*itProduct.Product, error) {
	update := r.client.Product.UpdateOneID(*product.Id).
		SetNillableDefaultVariantID(product.DefaultsVariantId).
		SetNillableThumbnailURL(product.ThumbnailUrl).
		// SetNillableTagIDs(product.TagIds).
		Where(entProduct.Etag(prevEtag))

	if product.Name != nil {
		update.SetName(*product.Name)
	}

	if product.Description != nil {
		update.SetDescription(*product.Description)
	}

	if product.Status != nil {
		update.SetStatus(*product.Status)
	}

	if product.Unit != nil {
		update.SetUnitID(*product.Unit)
	}

	if len(update.Mutation().Fields()) > 0 {
		update.SetEtag(*product.Etag)
	}

	return db.Mutate(ctx, update, ent.IsNotFound, itProduct.EntToProduct)
}

// ✅ Delete Product by ID
func (r *ProductEntRepository) DeleteById(ctx crud.Context, id model.Id) (int, error) {
	return r.client.Product.Delete().
		Where(entProduct.ID(id)).
		Exec(ctx)
}

// ✅ Find by ID
func (r *ProductEntRepository) FindById(ctx crud.Context, query itProduct.FindByIdParam) (*itProduct.Product, error) {
	dbQuery := r.productClient(ctx).Query().
		Where(entProduct.ID(query.Id))

	if query.WithVariants {
		dbQuery.WithVariant()
	}

	return db.FindOne(ctx, dbQuery, ent.IsNotFound, itProduct.EntToProduct)
}

// ✅ Search (advanced)
func (r *ProductEntRepository) Search(ctx crud.Context, param itProduct.SearchParam) (*crud.PagedResult[itProduct.Product], error) {
	query := r.client.Product.Query()

	return db.Search(
		ctx,
		param.Predicate,
		param.Order,
		crud.PagingOptions{
			Page: param.Page,
			Size: param.Size,
		},
		query,
		itProduct.EntToProducts,
	)
}

func (this *ProductEntRepository) ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors) {
	return db.ParseSearchGraphStr[ent.Product, itProduct.Product](criteria, entProduct.Label)
}

func BuildProductDescriptor() *orm.EntityDescriptor {
	entity := ent.Product{}
	builder := orm.DescribeEntity(entProduct.Label).
		Aliases("products").
		Field(entProduct.FieldCreatedAt, entity.CreatedAt).
		Field(entProduct.FieldDescription, entity.Description).
		Field(entProduct.FieldID, entity.ID).
		Field(entProduct.FieldName, entity.Name).
		Field(entProduct.FieldUpdatedAt, entity.UpdatedAt).
		Edge(entProduct.EdgeVariant, orm.ToEdgePredicate(entProduct.HasVariantWith))

	return builder.Descriptor()
}
