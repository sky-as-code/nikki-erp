package impl

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	db "github.com/sky-as-code/nikki-erp/modules/core/database"
	"github.com/sky-as-code/nikki-erp/modules/inventory/infra/ent"
	entVariant "github.com/sky-as-code/nikki-erp/modules/inventory/infra/ent/variant"
	itVariant "github.com/sky-as-code/nikki-erp/modules/inventory/variant/interfaces"
)

func NewVariantEntRepository(client *ent.Client) itVariant.VariantRepository {
	return &VariantEntRepository{
		client: client,
	}
}

type VariantEntRepository struct {
	client *ent.Client
}

func (r *VariantEntRepository) variantClient(ctx crud.Context) *ent.VariantClient {
	tx, isOk := ctx.GetDbTranx().(*ent.Tx)
	if isOk {
		return tx.Variant
	}
	return r.client.Variant
}

// ✅ Create Variant
func (r *VariantEntRepository) Create(ctx crud.Context, variant *itVariant.Variant) (*itVariant.Variant, error) {
	creation := r.client.Variant.Create().
		SetID(*variant.Id).
		SetProductID(*variant.ProductId).
		SetSku(*variant.Sku).
		SetStatus("active")

	if variant.Barcode != nil {
		creation.SetBarcode(*variant.Barcode)
	}
	if variant.ProposedPrice != nil {
		creation.SetProposedPrice(*variant.ProposedPrice)
	}
	if variant.Status != nil {
		creation.SetStatus(*variant.Status)
	}

	return db.Mutate(ctx, creation, ent.IsNotFound, itVariant.EntToVariant)
}

// ✅ Update Variant
func (r *VariantEntRepository) Update(ctx crud.Context, variant *itVariant.Variant, prevEtag model.Etag) (*itVariant.Variant, error) {
	update := r.client.Variant.UpdateOneID(*variant.Id).
		Where(entVariant.Etag(prevEtag))

	// SKU is immutable, cannot be updated
	if variant.Barcode != nil {
		update.SetBarcode(*variant.Barcode)
	}
	if variant.ProposedPrice != nil {
		update.SetProposedPrice(*variant.ProposedPrice)
	}
	if variant.Status != nil {
		update.SetStatus(*variant.Status)
	}

	if len(update.Mutation().Fields()) > 0 {
		update.SetEtag(*variant.Etag)
	}

	return db.Mutate(ctx, update, ent.IsNotFound, itVariant.EntToVariant)
}

// ✅ Delete Variant by ID
func (r *VariantEntRepository) DeleteById(ctx crud.Context, id model.Id) (int, error) {
	return r.client.Variant.Delete().
		Where(entVariant.ID(id)).
		Exec(ctx)
}

// ✅ Find by ID
func (r *VariantEntRepository) FindById(ctx crud.Context, query itVariant.FindByIdParam) (*itVariant.Variant, error) {
	dbQuery := r.variantClient(ctx).Query().
		Where(entVariant.ID(query.Id))

	return db.FindOne(ctx, dbQuery, ent.IsNotFound, itVariant.EntToVariant)
}

// ✅ Search (advanced)
func (r *VariantEntRepository) Search(ctx crud.Context, param itVariant.SearchParam) (*crud.PagedResult[itVariant.Variant], error) {
	query := r.client.Variant.Query()

	return db.Search(
		ctx,
		param.Predicate,
		param.Order,
		crud.PagingOptions{
			Page: param.Page,
			Size: param.Size,
		},
		query,
		itVariant.EntToVariants,
	)
}

func (this *VariantEntRepository) ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors) {
	return db.ParseSearchGraphStr[ent.Variant, itVariant.Variant](criteria, entVariant.Label)
}

func BuildVariantDescriptor() *orm.EntityDescriptor {
	entity := ent.Variant{}
	builder := orm.DescribeEntity(entVariant.Label).
		Aliases("variants").
		Field(entVariant.FieldCreatedAt, entity.CreatedAt).
		Field(entVariant.FieldSku, entity.Sku).
		Field(entVariant.FieldBarcode, entity.Barcode).
		Field(entVariant.FieldID, entity.ID).
		Field(entVariant.FieldStatus, entity.Status).
		Field(entVariant.FieldUpdatedAt, entity.UpdatedAt)

	return builder.Descriptor()
}
