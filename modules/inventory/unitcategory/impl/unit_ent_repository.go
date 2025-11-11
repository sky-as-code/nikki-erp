package impl

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	db "github.com/sky-as-code/nikki-erp/modules/core/database"
	"github.com/sky-as-code/nikki-erp/modules/inventory/infra/ent"
	entUnitCategory "github.com/sky-as-code/nikki-erp/modules/inventory/infra/ent/unitcategory"
	itUnitCategory "github.com/sky-as-code/nikki-erp/modules/inventory/unitcategory/interfaces"
)

func NewUnitCategoryEntRepository(client *ent.Client) itUnitCategory.UnitCategoryRepository {
	return &UnitCategoryEntRepository{
		client: client,
	}
}

type UnitCategoryEntRepository struct {
	client *ent.Client
}

func (r *UnitCategoryEntRepository) unitCategoryClient(ctx crud.Context) *ent.UnitCategoryClient {
	tx, isOk := ctx.GetDbTranx().(*ent.Tx)
	if isOk {
		return tx.UnitCategory
	}
	return r.client.UnitCategory
}

// ✅ Create UnitCategory
func (r *UnitCategoryEntRepository) Create(ctx crud.Context, unitCategory *itUnitCategory.UnitCategory) (*itUnitCategory.UnitCategory, error) {
	creation := r.client.UnitCategory.Create().
		SetID(*unitCategory.Id).
		SetOrgID(*unitCategory.OrgId).
		SetName(*unitCategory.Name).
		SetDescription(*unitCategory.Description).
		SetNillableStatus(unitCategory.Status).
		SetNillableThumbnailURL(unitCategory.ThumbnailUrl)

	return db.Mutate(ctx, creation, ent.IsNotFound, itUnitCategory.EntToUnitCategory)
}

// ✅ Update UnitCategory
func (r *UnitCategoryEntRepository) Update(ctx crud.Context, unitCategory *itUnitCategory.UnitCategory, prevEtag model.Etag) (*itUnitCategory.UnitCategory, error) {
	update := r.client.UnitCategory.UpdateOneID(*unitCategory.Id).
		SetName(*unitCategory.Name).
		SetDescription(*unitCategory.Description).
		SetNillableStatus(unitCategory.Status).
		SetNillableThumbnailURL(unitCategory.ThumbnailUrl).
		Where(entUnitCategory.Etag(prevEtag))

	if len(update.Mutation().Fields()) > 0 {
		update.SetEtag(*unitCategory.Etag)
	}

	return db.Mutate(ctx, update, ent.IsNotFound, itUnitCategory.EntToUnitCategory)
}

// ✅ Delete UnitCategory by ID
func (r *UnitCategoryEntRepository) DeleteById(ctx crud.Context, id model.Id) (int, error) {
	return r.client.UnitCategory.Delete().
		Where(entUnitCategory.ID(id)).
		Exec(ctx)
}

// ✅ Find by ID
func (r *UnitCategoryEntRepository) FindById(ctx crud.Context, query itUnitCategory.FindByIdParam) (*itUnitCategory.UnitCategory, error) {
	dbQuery := r.unitCategoryClient(ctx).Query().
		Where(entUnitCategory.ID(query.Id))

	return db.FindOne(ctx, dbQuery, ent.IsNotFound, itUnitCategory.EntToUnitCategory)
}

// ✅ Search (advanced)
func (r *UnitCategoryEntRepository) Search(ctx crud.Context, param itUnitCategory.SearchParam) (*crud.PagedResult[itUnitCategory.UnitCategory], error) {
	query := r.client.UnitCategory.Query()

	return db.Search(
		ctx,
		param.Predicate,
		param.Order,
		crud.PagingOptions{
			Page: param.Page,
			Size: param.Size,
		},
		query,
		itUnitCategory.EntToUnitCategories,
	)
}

func (r *UnitCategoryEntRepository) ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors) {
	return db.ParseSearchGraphStr[ent.UnitCategory, itUnitCategory.UnitCategory](criteria, entUnitCategory.Label)
}

func BuildUnitCategoryDescriptor() *orm.EntityDescriptor {
	entity := ent.UnitCategory{}
	builder := orm.DescribeEntity(entUnitCategory.Label).
		Aliases("unitcategories", "unit_categories").
		Field(entUnitCategory.FieldCreatedAt, entity.CreatedAt).
		Field(entUnitCategory.FieldID, entity.ID).
		Field(entUnitCategory.FieldName, entity.Name).
		Field(entUnitCategory.FieldDescription, entity.Description).
		Field(entUnitCategory.FieldStatus, entity.Status).
		Field(entUnitCategory.FieldUpdatedAt, entity.UpdatedAt)

	return builder.Descriptor()
}
