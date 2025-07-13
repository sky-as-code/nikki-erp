package repository

import (
	"context"
	"time"

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
		SetID(*group.Id).
		SetName(*group.Name).
		SetNillableDescription(group.Description).
		SetNillableOrgID(group.OrgId).
		SetEtag(*group.Etag)

	return db.Mutate(ctx, creation, ent.IsNotFound, entToGroup)
}

func (this *GroupEntRepository) Update(ctx context.Context, group domain.Group, prevEtag model.Etag) (*domain.Group, error) {
	update := this.client.Group.UpdateOneID(*group.Id).
		SetNillableName(group.Name).
		SetNillableDescription(group.Description).
		SetEtag(*group.Etag).
		SetNillableOrgID(group.OrgId).
		// IMPORTANT: Must have!
		Where(entGroup.EtagEQ(prevEtag))

	if len(update.Mutation().Fields()) > 0 {
		update.
			SetEtag(*group.Etag).
			SetUpdatedAt(time.Now())
	}

	return db.Mutate(ctx, update, ent.IsNotFound, entToGroup)
}

func (this *GroupEntRepository) Delete(ctx context.Context, param it.DeleteParam) error {
	return db.Delete[ent.Group](ctx, this.client.Group.DeleteOneID(param.Id))
}

func (this *GroupEntRepository) FindById(ctx context.Context, param it.GetGroupByIdQuery) (*domain.Group, error) {
	dbQuery := this.client.Group.Query().
		Where(entGroup.ID(param.Id))
	if param.WithOrg != nil && *param.WithOrg {
		dbQuery = dbQuery.WithOrg()
	}
	return db.FindOne(ctx, dbQuery, ent.IsNotFound, entToGroup)
}

func (this *GroupEntRepository) FindByName(ctx context.Context, param it.FindByNameParam) (*domain.Group, error) {
	return db.FindOne(
		ctx,
		this.client.Group.Query().Where(entGroup.Name(param.Name)),
		ent.IsNotFound,
		entToGroup,
	)
}

func (this *GroupEntRepository) ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors) {
	return db.ParseSearchGraphStr[ent.Group, domain.Group](criteria, entGroup.Label)
}

func (this *GroupEntRepository) Search(
	ctx context.Context,
	param it.SearchParam,
) (*crud.PagedResult[domain.Group], error) {
	query := this.client.Group.Query()
	if param.WithOrg {
		query = query.WithOrg()
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
		entToGroups,
	)
}

func (this *GroupEntRepository) AddRemoveUsers(ctx context.Context, param it.AddRemoveUsersParam) (*ft.ClientError, error) {
	if len(param.Add) == 0 && len(param.Remove) == 0 {
		return nil, nil
	}
	err := this.client.Group.UpdateOneID(param.GroupId).
		AddUserIDs(param.Add...).
		RemoveUserIDs(param.Remove...).
		SetEtag(param.Etag).
		Exec(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return &ft.ClientError{
				Code:    "not_found",
				Details: "some resource doesn't exist",
			}, nil
		}
		return nil, err
	}

	return nil, nil
}

func BuildGroupDescriptor() *orm.EntityDescriptor {
	entity := ent.Group{}
	builder := orm.DescribeEntity(entGroup.Label).
		Aliases("groups").
		Field(entGroup.FieldCreatedAt, entity.CreatedAt).
		Field(entGroup.FieldDescription, entity.Description).
		Field(entGroup.FieldID, entity.ID).
		Field(entGroup.FieldName, entity.Name).
		Field(entGroup.FieldUpdatedAt, entity.UpdatedAt).
		Edge(entGroup.EdgeUsers, orm.ToEdgePredicate(entGroup.HasUsersWith))

	return builder.Descriptor()
}
