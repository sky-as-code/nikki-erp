package repository

import (
	"time"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	db "github.com/sky-as-code/nikki-erp/modules/core/database"
	"github.com/sky-as-code/nikki-erp/modules/inventory/infra/ent"
	entProductCategory "github.com/sky-as-code/nikki-erp/modules/inventory/infra/ent/productcategory"
	"github.com/sky-as-code/nikki-erp/modules/inventory/product/domain"
	itProductCategory "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/productcategory"
)

func NewProductCategoryEntRepository(client *ent.Client) itProductCategory.ProductCategoryRepository {
	return &ProductCategoryEntRepository{
		client: client,
	}
}

type ProductCategoryEntRepository struct {
	client *ent.Client
}

func (r *ProductCategoryEntRepository) BeginTransaction(ctx crud.Context) (*ent.Tx, error) {
	return r.client.Tx(ctx)
}

func (r *ProductCategoryEntRepository) productCategoryClient(ctx crud.Context) *ent.ProductCategoryClient {
	tx, isOk := ctx.GetDbTranx().(*ent.Tx)
	if isOk {
		return tx.ProductCategory
	}
	return r.client.ProductCategory
}

// ✅ Create ProductCategory
func (r *ProductCategoryEntRepository) Create(ctx crud.Context, productCategory *domain.ProductCategory) (*domain.ProductCategory, error) {
	creation := r.productCategoryClient(ctx).Create().
		SetID(*productCategory.Id).
		SetOrgID(*productCategory.OrgId).
		SetName(*productCategory.Name).
		SetNillableParentID(productCategory.ParentId).
		SetEtag(*productCategory.Etag)

	return db.Mutate(ctx, creation, ent.IsNotFound, itProductCategory.EntToProductCategory)
}

// ✅ Update ProductCategory
func (r *ProductCategoryEntRepository) Update(ctx crud.Context, productCategory *domain.ProductCategory, prevEtag model.Etag) (*domain.ProductCategory, error) {
	update := r.productCategoryClient(ctx).UpdateOneID(*productCategory.Id).
		SetNillableParentID(productCategory.ParentId).
		Where(entProductCategory.Etag(prevEtag))

	if productCategory.Name != nil {
		update.SetName(*productCategory.Name)
	}

	if len(update.Mutation().Fields()) > 0 {
		update.SetEtag(*productCategory.Etag)
		update.SetUpdatedAt(time.Now())
	}

	return db.Mutate(ctx, update, ent.IsNotFound, itProductCategory.EntToProductCategory)
}

// ✅ Delete ProductCategory by ID
func (r *ProductCategoryEntRepository) DeleteById(ctx crud.Context, id model.Id) (int, error) {
	return r.client.ProductCategory.Delete().
		Where(entProductCategory.ID(id)).
		Exec(ctx)
}

// ✅ Find by ID
func (r *ProductCategoryEntRepository) FindById(ctx crud.Context, query itProductCategory.FindByIdParam) (*domain.ProductCategory, error) {
	dbQuery := r.productCategoryClient(ctx).Query().
		Where(entProductCategory.ID(query.Id))

	return db.FindOne(ctx, dbQuery, ent.IsNotFound, itProductCategory.EntToProductCategory)
}

// ✅ Search (advanced)
func (r *ProductCategoryEntRepository) Search(ctx crud.Context, param itProductCategory.SearchParam) (*crud.PagedResult[domain.ProductCategory], error) {
	query := r.client.ProductCategory.Query()

	return db.Search(
		ctx,
		param.Predicate,
		param.Order,
		crud.PagingOptions{
			Page: param.Page,
			Size: param.Size,
		},
		query,
		itProductCategory.EntToProductCategories,
	)
}

func (r *ProductCategoryEntRepository) ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors) {
	return db.ParseSearchGraphStr[ent.ProductCategory, domain.ProductCategory](criteria, entProductCategory.Label)
}

func BuildProductCategoryDescriptor() *orm.EntityDescriptor {
	entity := ent.ProductCategory{}
	builder := orm.DescribeEntity(entProductCategory.Label).
		Aliases("product_categories", "productCategories").
		Field(entProductCategory.FieldCreatedAt, entity.CreatedAt).
		Field(entProductCategory.FieldID, entity.ID).
		Field(entProductCategory.FieldName, entity.Name).
		Field(entProductCategory.FieldParentID, entity.ParentID).
		Field(entProductCategory.FieldUpdatedAt, entity.UpdatedAt).
		Edge(entProductCategory.EdgeChildren, orm.ToEdgePredicate(entProductCategory.HasChildrenWith)).
		Edge(entProductCategory.EdgeParent, orm.ToEdgePredicate(entProductCategory.HasParentWith)).
		Edge(entProductCategory.EdgeProduct, orm.ToEdgePredicate(entProductCategory.HasProductWith))

	return builder.Descriptor()
}
