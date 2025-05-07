package repository

import (
	"context"

	"github.com/sky-as-code/nikki-erp/common/crud"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
	"github.com/sky-as-code/nikki-erp/modules/identity/infra/ent"
	entGroup "github.com/sky-as-code/nikki-erp/modules/identity/infra/ent/group"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/user"
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
		SetCreatedBy(group.CreatedBy.String())

	if group.ParentId != nil {
		creation.SetParentID(group.ParentId.String())
	}

	entGroup, err := creation.Save(ctx)
	if err != nil {
		return nil, err
	}
	return entToGroup(entGroup), nil
}

func (this *GroupEntRepository) Update(ctx context.Context, group domain.Group) (*domain.Group, error) {
	update := this.client.Group.UpdateOneID(group.Id.String()).
		SetName(*group.Name).
		SetNillableDescription(group.Description).
		SetUpdatedBy(group.UpdatedBy.String())

	if group.ParentId != nil {
		update.SetParentID(group.ParentId.String())
	} else {
		update.ClearParentID()
	}

	entGroup, err := update.Save(ctx)
	if err != nil {
		return nil, err
	}
	return entToGroup(entGroup), nil
}

func (this *GroupEntRepository) Delete(ctx context.Context, id model.Id) error {
	err := this.client.Group.DeleteOneID(id.String()).
		Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (this *GroupEntRepository) FindById(ctx context.Context, id model.Id) (*domain.Group, error) {
	group, err := this.client.Group.Query().
		Where(entGroup.ID(id.String())).
		WithParent().
		WithSubgroups().
		WithUsers().
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return entToGroup(group), nil
}

func (this *GroupEntRepository) Search(
	ctx context.Context, criteria *orm.SearchGraph, opts *crud.PagingOptions,
) (*crud.PagedResult[*domain.Group], error) {
	predicate, err := criteria.ToPredicate()
	if err != nil {
		return nil, err
	}

	wholeQuery := this.client.Group.Query().
		Where(predicate)
	pagedQuery := wholeQuery.
		Offset(opts.Page * opts.Size).
		Limit(opts.Size)

	total, err := wholeQuery.Clone().Count(ctx)
	if err != nil {
		return nil, err
	}

	dbGroups, err := pagedQuery.All(ctx)
	if err != nil {
		return nil, err
	}

	return &crud.PagedResult[*domain.Group]{
		Items: entToGroups(dbGroups),
		Total: total,
	}, nil
}

func BuildGroupDescriptor() *orm.EntityDescriptor {
	entity := ent.Group{}
	builder := orm.DescribeEntity(entGroup.Label).
		Field(entGroup.FieldCreatedAt, entity.CreatedAt).
		Field(entGroup.FieldCreatedBy, entity.CreatedBy).
		Field(entGroup.FieldDescription, entity.Description).
		Field(entGroup.FieldID, entity.ID).
		Field(entGroup.FieldName, entity.Name).
		Field(entGroup.FieldParentID, entity.ParentID).
		Field(entGroup.FieldUpdatedAt, entity.UpdatedAt).
		Field(entGroup.FieldUpdatedBy, entity.UpdatedBy).
		Edge(entGroup.EdgeUsers, orm.ToEdgePredicate(entGroup.HasUsersWith)).
		Edge(entGroup.EdgeParent, orm.ToEdgePredicate(entGroup.HasParentWith)).
		Edge(entGroup.EdgeSubgroups, orm.ToEdgePredicate(entGroup.HasSubgroupsWith))

	return builder.Descriptor()
}
