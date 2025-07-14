package impl

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqljson"

	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/crud"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	db "github.com/sky-as-code/nikki-erp/modules/core/database"
	it "github.com/sky-as-code/nikki-erp/modules/core/enum/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/ent"
	entEnum "github.com/sky-as-code/nikki-erp/modules/core/infra/ent/enum"
)

func NewEnumEntRepository(client *ent.Client) it.EnumRepository {
	return &EnumEntRepository{
		client: client,
	}
}

type EnumEntRepository struct {
	client *ent.Client
}

func (this *EnumEntRepository) Create(ctx context.Context, enum it.Enum) (*it.Enum, error) {
	creation := this.client.Enum.Create().
		SetID(*enum.Id).
		SetEtag(*enum.Etag).
		SetLabel(*enum.Label).
		SetType(*enum.Type).
		SetNillableValue(enum.Value)

	return db.Mutate(ctx, creation, ent.IsNotFound, it.EntToEnum)
}

func (this *EnumEntRepository) Update(ctx context.Context, enum it.Enum, prevEtag model.Etag) (*it.Enum, error) {
	update := this.client.Enum.UpdateOneID(*enum.Id).
		SetLabel(*enum.Label).
		SetNillableValue(enum.Value).
		// IMPORTANT: Must have!
		Where(entEnum.EtagEQ(prevEtag))

	if len(update.Mutation().Fields()) > 0 {
		update.SetEtag(*enum.Etag)
	}

	return db.Mutate(ctx, update, ent.IsNotFound, it.EntToEnum)
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

func (this *EnumEntRepository) FindById(ctx context.Context, id model.Id) (*it.Enum, error) {
	query := this.client.Enum.Query().
		Where(entEnum.ID(id))

	return db.FindOne(ctx, query, ent.IsNotFound, it.EntToEnum)
}

func (this *EnumEntRepository) FindByValue(ctx context.Context, value string, enumType string) (*it.Enum, error) {
	query := this.client.Enum.Query().
		Where(entEnum.Value(value), entEnum.Type(enumType))

	return db.FindOne(ctx, query, ent.IsNotFound, it.EntToEnum)
}

func (this *EnumEntRepository) List(ctx context.Context, param it.ListParam) (*crud.PagedResult[it.Enum], error) {
	query := this.client.Enum.Query()

	if param.PartialLabel != nil && len(*param.PartialLabel) > 0 {
		query = query.Where(sql.FieldContainsFold(entEnum.FieldLabel, *param.PartialLabel))
	}
	if param.Type != nil {
		query = query.Where(entEnum.Type(*param.Type))
	}
	if param.SortByLang != nil {
		query = query.Order(
			sqljson.OrderValue(entEnum.FieldLabel, sqljson.Path(string(*param.SortByLang))),
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
		it.EntToEnums,
	)
}

func (this *EnumEntRepository) ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors) {
	return db.ParseSearchGraphStr[ent.Enum, it.Enum](criteria, entEnum.Label)
}

func (this *EnumEntRepository) Search(ctx context.Context, param it.SearchParam) (*crud.PagedResult[it.Enum], error) {
	query := this.client.Enum.Query()

	if param.TypePrefix != nil {
		query = query.Where(entEnum.TypeHasPrefix(*param.TypePrefix))
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
		it.EntToEnums,
	)
}

func BuildEnumDescriptor() *orm.EntityDescriptor {
	return GetEnumDescriptorBuilder(entEnum.Label).
		Aliases("enums").
		Descriptor()
}

func GetEnumDescriptorBuilder(entityName string) *orm.EntityDescriptorBuilder {
	entity := ent.Enum{}
	builder := orm.DescribeEntity(entityName).
		Field(entEnum.FieldID, entity.ID).
		Field(entEnum.FieldLabel, entity.Label).
		Field(entEnum.FieldValue, entity.Value).
		Field(entEnum.FieldType, entity.Type)

	return builder
}
