package repository

import (
	"context"

	"github.com/sky-as-code/nikki-erp/common/crud"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
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

func (this *GroupEntRepository) Create(ctx context.Context, group domain.Group) (*domain.GroupWithOrg, error) {
	creation := this.client.Group.Create().
		SetID(group.Id.String()).
		SetName(group.Name).
		SetNillableDescription(group.Description).
		SetCreatedBy(group.CreatedBy.String()).
		SetNillableOrgID(group.OrgId).
		SetEtag(group.Etag.String())

	return Mutate(ctx, creation, entToGroup)
}

func (this *GroupEntRepository) Update(ctx context.Context, group domain.Group) (*domain.GroupWithOrg, error) {
	update := this.client.Group.UpdateOneID(group.Id.String()).
		SetName(group.Name).
		SetNillableDescription(group.Description).
		SetEtag(group.Etag.String()).
		SetNillableOrgID(group.OrgId).
		SetUpdatedBy(group.UpdatedBy.String())

	return Mutate(ctx, update, entToGroup)
}

func (this *GroupEntRepository) Delete(ctx context.Context, id model.Id) error {
	return Delete[ent.Group](ctx, this.client.Group.DeleteOneID(id.String()))
}

func (this *GroupEntRepository) FindById(ctx context.Context, id model.Id, withOrg bool) (*domain.GroupWithOrg, error) {
	query := this.client.Group.Query().
		Where(entGroup.ID(id.String()))
	if withOrg {
		query = query.WithOrganization()
	}
	return FindOne(ctx, query, entToGroup)
}

func (this *GroupEntRepository) FindByName(ctx context.Context, name string) (*domain.GroupWithOrg, error) {
	query := this.client.Group.Query().
		Where(entGroup.Name(name))
	return FindOne(ctx, query, entToGroup)
}

func (this *GroupEntRepository) Search(
	ctx context.Context, criteria *orm.SearchGraph, opts *crud.PagingOptions,
) (*crud.PagedResult[domain.Group], error) {
	return Search(
		ctx,
		criteria,
		opts,
		entGroup.Label,
		this.client.Group.Query(),
		entToGroups,
	)
}

func BuildGroupDescriptor() *orm.EntityDescriptor {
	entity := ent.Group{}
	builder := orm.DescribeEntity(entGroup.Label).
		Field(entGroup.FieldCreatedAt, entity.CreatedAt).
		Field(entGroup.FieldCreatedBy, entity.CreatedBy).
		Field(entGroup.FieldDescription, entity.Description).
		Field(entGroup.FieldID, entity.ID).
		Field(entGroup.FieldName, entity.Name).
		Field(entGroup.FieldUpdatedAt, entity.UpdatedAt).
		Field(entGroup.FieldUpdatedBy, entity.UpdatedBy).
		Edge(entGroup.EdgeUsers, orm.ToEdgePredicate(entGroup.HasUsersWith))

	return builder.Descriptor()
}
