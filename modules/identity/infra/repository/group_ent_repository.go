package repository

import (
	"context"

	"github.com/sky-as-code/nikki-erp/common/crud"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	db "github.com/sky-as-code/nikki-erp/modules/core/database"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
	"github.com/sky-as-code/nikki-erp/modules/identity/infra/ent"
	entGroup "github.com/sky-as-code/nikki-erp/modules/identity/infra/ent/group"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/group"
)

func NewGroupEntRepository(client *ent.Client) it.GroupRepository {
	return &GroupEntRepository{
		client: client,
	}
}

type GroupEntRepository struct {
	client *ent.Client
}

func (this *GroupEntRepository) Create(ctx context.Context, group domain.Group) (*domain.Group, error) {
	creation := this.client.Group.Create().
		SetID(group.Id.String()).
		SetName(*group.Name).
		SetNillableDescription(group.Description).
		SetNillableOrgID(model.IdToNillableStr(group.OrgId)).
		SetEtag(group.Etag.String())

	return db.Mutate(ctx, creation, entToGroup)
}

func (this *GroupEntRepository) Update(ctx context.Context, group domain.Group) (*domain.Group, error) {
	update := this.client.Group.UpdateOneID(group.Id.String()).
		SetName(*group.Name).
		SetNillableDescription(group.Description).
		SetEtag(group.Etag.String()).
		SetNillableOrgID(model.IdToNillableStr(group.OrgId))

	return db.Mutate(ctx, update, entToGroup)
}

func (this *GroupEntRepository) Delete(ctx context.Context, id model.Id) error {
	return db.Delete[ent.Group](ctx, this.client.Group.DeleteOneID(id.String()))
}

func (this *GroupEntRepository) FindById(ctx context.Context, param it.GetGroupByIdQuery) (*domain.Group, error) {
	dbQuery := this.client.Group.Query().
		Where(entGroup.ID(param.Id.String()))
	if *param.WithOrg {
		dbQuery = dbQuery.WithOrg()
	}
	return db.FindOne(ctx, dbQuery, entToGroup)
}

func (this *GroupEntRepository) FindByName(ctx context.Context, name string) (*domain.Group, error) {
	return db.FindOne(
		ctx,
		this.client.Group.Query().Where(entGroup.Name(name)),
		entToGroup,
	)
}

func (this *GroupEntRepository) ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors) {
	return db.ParseSearchGraphStr[ent.Group, domain.Group](criteria)
}

func (this *GroupEntRepository) Search(
	ctx context.Context,
	predicate *orm.Predicate,
	order []orm.OrderOption,
	opts crud.PagingOptions,
) (*crud.PagedResult[domain.Group], error) {
	return db.Search(
		ctx,
		predicate,
		order,
		opts,
		this.client.Group.Query(),
		entToGroups,
	)
}

func BuildGroupDescriptor() *orm.EntityDescriptor {
	entity := ent.Group{}
	builder := orm.DescribeEntity(entGroup.Label).
		Field(entGroup.FieldCreatedAt, entity.CreatedAt).
		Field(entGroup.FieldDescription, entity.Description).
		Field(entGroup.FieldID, entity.ID).
		Field(entGroup.FieldName, entity.Name).
		Field(entGroup.FieldUpdatedAt, entity.UpdatedAt).
		Edge(entGroup.EdgeUsers, orm.ToEdgePredicate(entGroup.HasUsersWith))

	return builder.Descriptor()
}
