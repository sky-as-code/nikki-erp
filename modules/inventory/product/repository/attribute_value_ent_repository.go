package repository

import (
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqljson"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	db "github.com/sky-as-code/nikki-erp/modules/core/database"
	"github.com/sky-as-code/nikki-erp/modules/inventory/infra/ent"
	entAttributeValue "github.com/sky-as-code/nikki-erp/modules/inventory/infra/ent/attributevalue"
	entVariantAttributeRel "github.com/sky-as-code/nikki-erp/modules/inventory/infra/ent/variantattributerel"
	"github.com/sky-as-code/nikki-erp/modules/inventory/product/domain"
	itAttributeValue "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/attributevalue"
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

func (r *AttributeValueEntRepository) variantAttributeRelClient(ctx crud.Context) *ent.VariantAttributeRelClient {
	tx, isOk := ctx.GetDbTranx().(*ent.Tx)
	if isOk {
		return tx.VariantAttributeRel
	}
	return r.client.VariantAttributeRel
}

// ✅ Create AttributeValue
func (r *AttributeValueEntRepository) Create(ctx crud.Context, attributeValue *domain.AttributeValue) (*domain.AttributeValue, error) {
	creation := r.client.AttributeValue.Create().
		SetID(*attributeValue.Id).
		SetAttributeID(*attributeValue.AttributeId).
		SetEtag(*attributeValue.Etag)

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
	} else {
		creation.SetVariantCount(0)
	}

	return db.Mutate(ctx, creation, ent.IsNotFound, itAttributeValue.EntToAttributeValue)
}

func (r *AttributeValueEntRepository) CreateAndLinkVariant(ctx crud.Context, attributeValue *domain.AttributeValue, variantId model.Id) (*domain.AttributeValue, error) {
	creation := r.attributeValueClient(ctx).Create().
		SetID(*attributeValue.Id).
		SetAttributeID(*attributeValue.AttributeId).
		SetVariantCount(1).
		SetEtag(*attributeValue.Etag).
		AddVariantIDs(variantId)

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

func (r *AttributeValueEntRepository) LinkVariantToExisting(ctx crud.Context, attributeValueId model.Id, variantId model.Id, prevEtag model.Etag) (*domain.AttributeValue, bool, error) {
	exists, err := r.variantAttributeRelClient(ctx).Query().
		Where(
			entVariantAttributeRel.VariantID(variantId),
			entVariantAttributeRel.AttributeValueID(attributeValueId),
		).
		Exist(ctx)
	if err != nil {
		return nil, false, err
	}
	if exists {
		return nil, false, nil
	}

	newEtag := model.NewEtag()
	update := r.attributeValueClient(ctx).UpdateOneID(attributeValueId).
		Where(entAttributeValue.Etag(prevEtag)).
		AddVariantIDs(variantId).
		AddVariantCount(1).
		SetEtag(*newEtag)

	updated, err := db.Mutate(ctx, update, ent.IsNotFound, itAttributeValue.EntToAttributeValue)
	if err != nil {
		return nil, false, err
	}
	return updated, true, nil
}

// ✅ Update AttributeValue
func (r *AttributeValueEntRepository) Update(ctx crud.Context, attributeValue *domain.AttributeValue, prevEtag model.Etag) (*domain.AttributeValue, error) {
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
func (r *AttributeValueEntRepository) FindById(ctx crud.Context, query itAttributeValue.FindByIdParam) (*domain.AttributeValue, error) {
	dbQuery := r.attributeValueClient(ctx).Query().
		Where(entAttributeValue.ID(query.Id))

	return db.FindOne(ctx, dbQuery, ent.IsNotFound, itAttributeValue.EntToAttributeValue)
}

func (r *AttributeValueEntRepository) FindByValueRef(ctx crud.Context, attributeValue *domain.AttributeValue, dataType string) (*domain.AttributeValue, error) {
	dbQuery := r.attributeValueClient(ctx).Query().
		Where(entAttributeValue.AttributeID(*attributeValue.AttributeId))

	switch dataType {
	case "text":
		if attributeValue.ValueText != nil {
			dbQuery = dbQuery.Where(func(s *sql.Selector) {
				s.Where(
					sqljson.ValueEQ(
						entAttributeValue.FieldValueText,
						*attributeValue.ValueText,
					),
				)
			})
		}
	case "number":
		if attributeValue.ValueNumber != nil {
			dbQuery = dbQuery.Where(entAttributeValue.ValueNumber(*attributeValue.ValueNumber))
		}
	case "boolean":
		if attributeValue.ValueBool != nil {
			dbQuery = dbQuery.Where(entAttributeValue.ValueBool(*attributeValue.ValueBool))
		}
	default:
		if attributeValue.ValueRef != nil {
			dbQuery = dbQuery.Where(entAttributeValue.ValueRef(*attributeValue.ValueRef))
		}
	}

	return db.FindOne(ctx, dbQuery, ent.IsNotFound, itAttributeValue.EntToAttributeValue)
}

// ✅ Search (advanced)
func (r *AttributeValueEntRepository) Search(ctx crud.Context, param itAttributeValue.SearchParam) (*crud.PagedResult[domain.AttributeValue], error) {
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
	return db.ParseSearchGraphStr[ent.AttributeValue, domain.AttributeValue](criteria, entAttributeValue.Label)
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
