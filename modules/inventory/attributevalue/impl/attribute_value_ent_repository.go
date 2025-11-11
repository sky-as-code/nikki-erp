package impl

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	db "github.com/sky-as-code/nikki-erp/modules/core/database"
	itAttributeValue "github.com/sky-as-code/nikki-erp/modules/inventory/attributevalue/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/inventory/infra/ent"
	entAttributeValue "github.com/sky-as-code/nikki-erp/modules/inventory/infra/ent/attributevalue"
)

func NewAttributeValueEntRepository(client *ent.Client) itAttributeValue.AttributeValueRepository {
	return &AttributeValueEntRepository{
		client: client,
	}
}

type AttributeValueEntRepository struct {
	client *ent.Client
}

func (r *AttributeValueEntRepository) attributeValueClient(ctx crud.Context) *ent.AttributeValueClient {
	tx, isOk := ctx.GetDbTranx().(*ent.Tx)
	if isOk {
		return tx.AttributeValue
	}
	return r.client.AttributeValue
}

// ✅ Create AttributeValue
func (r *AttributeValueEntRepository) Create(ctx crud.Context, attributeValue *itAttributeValue.AttributeValue) (*itAttributeValue.AttributeValue, error) {
	creation := r.client.AttributeValue.Create().
		SetID(*attributeValue.Id).
		SetAttributeID(*attributeValue.AttributeId).
		SetVariantCount(0)

	if attributeValue.ValueText != nil {
		creation.SetValueText(*attributeValue.ValueText)
	}
	if attributeValue.ValueNumber != nil {
		creation.SetValueNumber(*attributeValue.ValueNumber)
	}
	if attributeValue.ValueBool != nil {
		creation.SetValueBool(*attributeValue.ValueBool)
	}
	if attributeValue.ValueRef != nil {
		creation.SetValueRef(*attributeValue.ValueRef)
	}
	if attributeValue.VariantCount != nil {
		creation.SetVariantCount(*attributeValue.VariantCount)
	}

	return db.Mutate(ctx, creation, ent.IsNotFound, itAttributeValue.EntToAttributeValue)
}

// ✅ Update AttributeValue
func (r *AttributeValueEntRepository) Update(ctx crud.Context, attributeValue *itAttributeValue.AttributeValue, prevEtag model.Etag) (*itAttributeValue.AttributeValue, error) {
	update := r.client.AttributeValue.UpdateOneID(*attributeValue.Id).
		Where(entAttributeValue.Etag(prevEtag))

	// AttributeId is immutable, cannot be updated
	if attributeValue.ValueText != nil {
		update.SetValueText(*attributeValue.ValueText)
	}
	if attributeValue.ValueNumber != nil {
		update.SetValueNumber(*attributeValue.ValueNumber)
	}
	if attributeValue.ValueBool != nil {
		update.SetValueBool(*attributeValue.ValueBool)
	}
	if attributeValue.ValueRef != nil {
		update.SetValueRef(*attributeValue.ValueRef)
	}
	if attributeValue.VariantCount != nil {
		update.SetVariantCount(*attributeValue.VariantCount)
	}

	if len(update.Mutation().Fields()) > 0 {
		update.SetEtag(*attributeValue.Etag)
	}

	return db.Mutate(ctx, update, ent.IsNotFound, itAttributeValue.EntToAttributeValue)
}

// ✅ Delete AttributeValue by ID
func (r *AttributeValueEntRepository) DeleteById(ctx crud.Context, id model.Id) (int, error) {
	return r.client.AttributeValue.Delete().
		Where(entAttributeValue.ID(id)).
		Exec(ctx)
}

// ✅ Find by ID
func (r *AttributeValueEntRepository) FindById(ctx crud.Context, query itAttributeValue.FindByIdParam) (*itAttributeValue.AttributeValue, error) {
	dbQuery := r.attributeValueClient(ctx).Query().
		Where(entAttributeValue.ID(query.Id))

	return db.FindOne(ctx, dbQuery, ent.IsNotFound, itAttributeValue.EntToAttributeValue)
}

// ✅ Search (advanced)
func (r *AttributeValueEntRepository) Search(ctx crud.Context, param itAttributeValue.SearchParam) (*crud.PagedResult[itAttributeValue.AttributeValue], error) {
	query := r.client.AttributeValue.Query()

	return db.Search(
		ctx,
		param.Predicate,
		param.Order,
		crud.PagingOptions{
			Page: param.Page,
			Size: param.Size,
		},
		query,
		itAttributeValue.EntToAttributeValues,
	)
}

func (this *AttributeValueEntRepository) ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors) {
	return db.ParseSearchGraphStr[ent.AttributeValue, itAttributeValue.AttributeValue](criteria, entAttributeValue.Label)
}

func BuildAttributeValueDescriptor() *orm.EntityDescriptor {
	entity := ent.AttributeValue{}
	builder := orm.DescribeEntity(entAttributeValue.Label).
		Aliases("attribute_values").
		Field(entAttributeValue.FieldCreatedAt, entity.CreatedAt).
		Field(entAttributeValue.FieldValueText, entity.ValueText).
		Field(entAttributeValue.FieldValueRef, entity.ValueRef).
		Field(entAttributeValue.FieldID, entity.ID).
		Field(entAttributeValue.FieldUpdatedAt, entity.UpdatedAt)

	return builder.Descriptor()
}
