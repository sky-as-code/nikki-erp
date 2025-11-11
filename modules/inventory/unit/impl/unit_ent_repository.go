package impl

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	db "github.com/sky-as-code/nikki-erp/modules/core/database"
	"github.com/sky-as-code/nikki-erp/modules/inventory/infra/ent"
	entUnit "github.com/sky-as-code/nikki-erp/modules/inventory/infra/ent/unit"
	itUnit "github.com/sky-as-code/nikki-erp/modules/inventory/unit/interfaces"
)

func NewUnitEntRepository(client *ent.Client) itUnit.UnitRepository {
	return &UnitEntRepository{
		client: client,
	}
}

type UnitEntRepository struct {
	client *ent.Client
}

func (r *UnitEntRepository) unitClient(ctx crud.Context) *ent.UnitClient {
	tx, isOk := ctx.GetDbTranx().(*ent.Tx)
	if isOk {
		return tx.Unit
	}
	return r.client.Unit
}

// Create Unit
func (r *UnitEntRepository) Create(ctx crud.Context, unit *itUnit.Unit) (*itUnit.Unit, error) {
	creation := r.client.Unit.Create().
		SetID(*unit.Id).
		SetName(*unit.Name).
		SetSymbol(*unit.Symbol).
		SetNillableCategoryID(unit.CategoryId).
		SetNillableMultiplier(unit.Multiplier).
		SetNillableOrgID(unit.OrgId).
		SetNillableStatus(unit.Status).
		SetNillableBaseUnit(unit.BaseUnit).
		SetEtag(*unit.Etag)

	return db.Mutate(ctx, creation, ent.IsNotFound, itUnit.EntToUnit)
}

// Update Unit
func (r *UnitEntRepository) Update(ctx crud.Context, unit *itUnit.Unit, prevEtag model.Etag) (*itUnit.Unit, error) {
	update := r.client.Unit.UpdateOneID(*unit.Id).
		SetName(*unit.Name).
		SetSymbol(*unit.Symbol).
		SetNillableCategoryID(unit.CategoryId).
		SetNillableMultiplier(unit.Multiplier).
		SetNillableStatus(unit.Status).
		Where(entUnit.Etag(prevEtag))

	if unit.BaseUnit != nil {
		update.SetBaseUnit(*unit.BaseUnit)
	}

	if len(update.Mutation().Fields()) > 0 {
		update.SetEtag(*unit.Etag)
	}

	return db.Mutate(ctx, update, ent.IsNotFound, itUnit.EntToUnit)
}

// Delete Unit by ID
func (r *UnitEntRepository) DeleteById(ctx crud.Context, id model.Id) (int, error) {
	return r.client.Unit.Delete().
		Where(entUnit.ID(id)).
		Exec(ctx)
}

// Find by ID
func (r *UnitEntRepository) FindById(ctx crud.Context, query itUnit.FindByIdParam) (*itUnit.Unit, error) {
	dbQuery := r.unitClient(ctx).Query().
		Where(entUnit.ID(query.Id))

	return db.FindOne(ctx, dbQuery, ent.IsNotFound, itUnit.EntToUnit)
}

// func (r *UnitEntRepository) FindByName(ctx crud.Context, name model.LangJson) (*itUnit.Unit, error) {
// 	dbQuery := r.unitClient(ctx).Query().
// 		Where(entUnit.)

// 	return db.FindOne(ctx, dbQuery, ent.IsNotFound, itUnit.EntToUnit)
// }

// Search (advanced)
func (r *UnitEntRepository) Search(ctx crud.Context, param itUnit.SearchParam) (*crud.PagedResult[itUnit.Unit], error) {
	query := r.client.Unit.Query()

	return db.Search(
		ctx,
		param.Predicate,
		param.Order,
		crud.PagingOptions{
			Page: param.Page,
			Size: param.Size,
		},
		query,
		itUnit.EntToUnits,
	)
}

func (this *UnitEntRepository) ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors) {
	return db.ParseSearchGraphStr[ent.Unit, itUnit.Unit](criteria, entUnit.Label)
}

func BuildUnitDescriptor() *orm.EntityDescriptor {
	entity := ent.Unit{}
	builder := orm.DescribeEntity(entUnit.Label).
		Aliases("units").
		Field(entUnit.FieldCreatedAt, entity.CreatedAt).
		Field(entUnit.FieldID, entity.ID).
		Field(entUnit.FieldName, entity.Name).
		Field(entUnit.FieldSymbol, entity.Symbol).
		Field(entUnit.FieldBaseUnit, entity.BaseUnit).
		Field(entUnit.FieldUpdatedAt, entity.UpdatedAt)

	return builder.Descriptor()
}
