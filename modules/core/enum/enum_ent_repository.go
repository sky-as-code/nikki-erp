package enum

import (
	"context"

	"entgo.io/ent/dialect/sql/sqljson"

	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/crud"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	db "github.com/sky-as-code/nikki-erp/modules/core/database"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/ent"
	entEnum "github.com/sky-as-code/nikki-erp/modules/core/infra/ent/enum"
)

func NewEnumEntRepository(client *ent.Client) EnumRepository {
	return &EnumEntRepository{
		client: client,
	}
}

type EnumEntRepository struct {
	client *ent.Client
}

func (this *EnumEntRepository) Create(ctx context.Context, enum Enum) (*Enum, error) {
	creation := this.client.Enum.Create().
		SetID(*enum.Id).
		SetLabel(*enum.Label).
		SetValue(*enum.Value)

	return db.Mutate(ctx, creation, EntToEnum)
}

func (this *EnumEntRepository) Update(ctx context.Context, enum Enum) (*Enum, error) {
	update := this.client.Enum.UpdateOneID(*enum.Id).
		SetLabel(*enum.Label).
		SetValue(*enum.Value)

	return db.Mutate(ctx, update, EntToEnum)
}

func (this *EnumEntRepository) DeleteById(ctx context.Context, id model.Id) (int, error) {
	return db.DeleteMulti[ent.Enum](ctx, this.client.Enum.Delete().Where(entEnum.ID(id)))
}

func (this *EnumEntRepository) DeleteByType(ctx context.Context, enumType string) (int, error) {
	return db.DeleteMulti[ent.Enum](
		ctx,
		this.client.Enum.Delete().Where(entEnum.Type(enumType)),
	)
}

func (this *EnumEntRepository) Exists(ctx context.Context, id model.Id) (bool, error) {
	return this.client.Enum.Query().
		Where(entEnum.ID(id)).
		Exist(ctx)
}

func (this *EnumEntRepository) ExistsMulti(ctx context.Context, ids []model.Id) (existing []model.Id, notExisting []model.Id, err error) {
	dbEntities, err := this.client.Enum.Query().
		Where(entEnum.IDIn(ids...)).
		Select(entEnum.FieldID).
		All(ctx)

	if err != nil {
		return nil, nil, err
	}

	existing = array.Map(dbEntities, func(entity *ent.Enum) model.Id {
		return entity.ID
	})

	notExisting = array.Filter(ids, func(id model.Id) bool {
		return !array.Contains(existing, id)
	})

	return existing, notExisting, nil
}

func (this *EnumEntRepository) FindById(ctx context.Context, id model.Id) (*Enum, error) {
	query := this.client.Enum.Query().
		Where(entEnum.ID(id))

	return db.FindOne(ctx, query, ent.IsNotFound, EntToEnum)
}

func (this *EnumEntRepository) FindByValue(ctx context.Context, value string, enumType string) (*Enum, error) {
	query := this.client.Enum.Query().
		Where(entEnum.Value(value), entEnum.Type(enumType))

	return db.FindOne(ctx, query, ent.IsNotFound, EntToEnum)
}

func (this *EnumEntRepository) List(ctx context.Context, param ListParam) (*crud.PagedResult[Enum], error) {
	query := this.client.Enum.Query()

	if param.EnumType != nil {
		query = query.Where(entEnum.Type(*param.EnumType))
	}
	if param.SortedByLang != nil {
		query = query.Order(
			sqljson.OrderValue(entEnum.FieldLabel, sqljson.Path(string(*param.SortedByLang))),
		)
	}

	return db.Search(
		ctx,
		nil,
		nil,
		crud.PagingOptions{
			Page: *param.Page,
			Size: *param.Size,
		},
		query,
		EntToEnums,
	)
}

func BuildEnumDescriptor() *orm.EntityDescriptor {
	entity := ent.Enum{}
	builder := orm.DescribeEntity(entEnum.Label).
		Aliases("enums").
		Field(entEnum.FieldID, entity.ID).
		Field(entEnum.FieldLabel, entity.Label).
		Field(entEnum.FieldValue, entity.Value).
		Field(entEnum.FieldType, entity.Type)

	return builder.Descriptor()
}
