package repository

import (
	"time"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
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

func (this *GroupEntRepository) groupClient(ctx crud.Context) *ent.GroupClient {
	tx, isOk := ctx.GetDbTranx().(*ent.Tx)
	if isOk {
		return tx.Group
	}
	return this.client.Group
}

func (this *GroupEntRepository) Create(ctx crud.Context, group *domain.Group) (*domain.Group, error) {
	creation := this.groupClient(ctx).Create().
		SetID(*group.Id).
		SetName(*group.Name).
		SetNillableDescription(group.Description).
		SetNillableOrgID(group.OrgId).
		SetEtag(*group.Etag)

	return db.Mutate(ctx, creation, ent.IsNotFound, entToGroup)
}

func (this *GroupEntRepository) Update(ctx crud.Context, group *domain.Group, prevEtag model.Etag) (*domain.Group, error) {
	update := this.groupClient(ctx).UpdateOneID(*group.Id).
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

func (this *GroupEntRepository) DeleteHard(ctx crud.Context, param it.DeleteParam) (int, error) {
	return this.groupClient(ctx).Delete().
		Where(entGroup.ID(param.Id)).
		Exec(ctx)
}

func (this *GroupEntRepository) FindById(ctx crud.Context, param it.GetGroupByIdQuery) (*domain.Group, error) {
	dbQuery := this.groupClient(ctx).Query().
		Where(entGroup.ID(param.Id))
	if param.WithOrg != false {
		dbQuery = dbQuery.WithOrg()
	}
	return db.FindOne(ctx, dbQuery, ent.IsNotFound, entToGroup)
}

func (this *GroupEntRepository) FindByName(ctx crud.Context, param it.FindByNameParam) (*domain.Group, error) {
	return db.FindOne(
		ctx,
		this.groupClient(ctx).Query().Where(entGroup.Name(param.Name)),
		ent.IsNotFound,
		entToGroup,
	)
}

func (this *GroupEntRepository) ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors) {
	return db.ParseSearchGraphStr[ent.Group, domain.Group](criteria, entGroup.Label)
}

func (this *GroupEntRepository) Search(
	ctx crud.Context,
	param it.SearchParam,
) (*crud.PagedResult[domain.Group], error) {
	query := this.groupClient(ctx).Query()
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

func (this *GroupEntRepository) AddRemoveUsers(ctx crud.Context, param it.AddRemoveUsersParam) (*ft.ClientError, error) {
	if len(param.Add) == 0 && len(param.Remove) == 0 {
		return nil, nil
	}
	err := this.groupClient(ctx).UpdateOneID(param.GroupId).
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

func (this *GroupEntRepository) Exists(ctx crud.Context, param it.ExistsParam) (bool, error) {
	return this.client.Group.Query().
		Where(entGroup.ID(param.Id)).
		Exist(ctx)
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
		Edge(entGroup.EdgeUsers, orm.ToEdgePredicate(entGroup.HasUsersWith)).
		Edge(entGroup.EdgeOrg, orm.ToEdgePredicate(entGroup.HasOrgWith))

	return builder.Descriptor()
}
