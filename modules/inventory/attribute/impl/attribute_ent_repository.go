package impl

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	db "github.com/sky-as-code/nikki-erp/modules/core/database"
	itAttribute "github.com/sky-as-code/nikki-erp/modules/inventory/attribute/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/inventory/infra/ent"
	entAttribute "github.com/sky-as-code/nikki-erp/modules/inventory/infra/ent/attribute"
)

func NewAttributeEntRepository(client *ent.Client) itAttribute.AttributeRepository {
	return &AttributeEntRepository{
		client: client,
	}
}

type AttributeEntRepository struct {
	client *ent.Client
}

func (r *AttributeEntRepository) attributeClient(ctx crud.Context) *ent.AttributeClient {
	tx, isOk := ctx.GetDbTranx().(*ent.Tx)
	if isOk {
		return tx.Attribute
	}
	return r.client.Attribute
}

// ✅ Create Attribute
func (r *AttributeEntRepository) Create(ctx crud.Context, attribute *itAttribute.Attribute) (*itAttribute.Attribute, error) {
	creation := r.client.Attribute.Create().
		SetID(*attribute.Id).
		SetProductID(*attribute.ProductId).
		SetCodeName(*attribute.CodeName).
		SetDataType(*attribute.DataType).
		SetNillableIsRequired(attribute.IsRequired).
		SetNillableIsEnum(attribute.IsEnum).
		SetNillableEnumValueSort(attribute.EnumValueSort).
		SetNillableSortIndex(attribute.SortIndex).
		SetNillableGroupID(attribute.GroupId).
		SetEtag(*attribute.Etag)

	if attribute.DisplayName != nil {
		creation.SetDisplayName(*attribute.DisplayName)
	}

	if attribute.EnumValue != nil {
		creation.SetEnumValue(*attribute.EnumValue)
	}

	return db.Mutate(ctx, creation, ent.IsNotFound, itAttribute.EntToAttribute)
}

// ✅ Update Attribute
func (r *AttributeEntRepository) Update(ctx crud.Context, attribute *itAttribute.Attribute, prevEtag model.Etag) (*itAttribute.Attribute, error) {
	update := r.client.Attribute.UpdateOneID(*attribute.Id).
		SetNillableCodeName(attribute.CodeName).
		SetNillableIsEnum(attribute.IsEnum).
		SetNillableIsRequired(attribute.IsRequired).
		SetNillableDataType(attribute.DataType).
		SetNillableEnumValueSort(attribute.EnumValueSort).
		SetNillableProductID(attribute.ProductId).
		Where(entAttribute.Etag(prevEtag))

	// CodeName is immutable, cannot be updated
	if attribute.DisplayName != nil {
		update.SetDisplayName(*attribute.DisplayName)
	}

	if attribute.EnumValue != nil {
		update.SetEnumValue(*attribute.EnumValue)
	}

	if len(update.Mutation().Fields()) > 0 {
		update.SetEtag(*attribute.Etag)
	}

	return db.Mutate(ctx, update, ent.IsNotFound, itAttribute.EntToAttribute)
}

// ✅ Delete Attribute by ID
func (r *AttributeEntRepository) DeleteById(ctx crud.Context, id model.Id) (int, error) {
	return r.client.Attribute.Delete().
		Where(entAttribute.ID(id)).
		Exec(ctx)
}

// ✅ Find by ID
func (r *AttributeEntRepository) FindById(ctx crud.Context, query itAttribute.FindByIdParam) (*itAttribute.Attribute, error) {
	dbQuery := r.attributeClient(ctx).Query().
		Where(entAttribute.ID(query.Id))

	return db.FindOne(ctx, dbQuery, ent.IsNotFound, itAttribute.EntToAttribute)
}

// ✅ Search (advanced)
func (r *AttributeEntRepository) Search(ctx crud.Context, param itAttribute.SearchParam) (*crud.PagedResult[itAttribute.Attribute], error) {
	query := r.client.Attribute.Query()

	return db.Search(
		ctx,
		param.Predicate,
		param.Order,
		crud.PagingOptions{
			Page: param.Page,
			Size: param.Size,
		},
		query,
		itAttribute.EntToAttributes,
	)
}

func (this *AttributeEntRepository) ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors) {
	return db.ParseSearchGraphStr[ent.Attribute, itAttribute.Attribute](criteria, entAttribute.Label)
}

func BuildAttributeDescriptor() *orm.EntityDescriptor {
	entity := ent.Attribute{}
	builder := orm.DescribeEntity(entAttribute.Label).
		Aliases("attributes").
		Field(entAttribute.FieldCreatedAt, entity.CreatedAt).
		Field(entAttribute.FieldCodeName, entity.CodeName).
		// Field(entAttribute.FieldDisplayName, entity.DisplayName).
		Field(entAttribute.FieldID, entity.ID).
		Field(entAttribute.FieldDataType, entity.DataType).
		Field(entAttribute.FieldUpdatedAt, entity.UpdatedAt)

	return builder.Descriptor()
}
