package repository

import (
	"context"
	"time"

	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/crud"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	db "github.com/sky-as-code/nikki-erp/modules/core/database"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
	"github.com/sky-as-code/nikki-erp/modules/identity/infra/ent"
	entUser "github.com/sky-as-code/nikki-erp/modules/identity/infra/ent/user"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/user"
)

func NewUserEntRepository(client *ent.Client) it.UserRepository {
	return &UserEntRepository{
		client: client,
	}
}

type UserEntRepository struct {
	client *ent.Client
}

func (this *UserEntRepository) Create(ctx context.Context, user domain.User) (*domain.User, error) {
	creation := this.client.User.Create().
		SetID(*user.Id).
		SetNillableAvatarURL(user.AvatarUrl).
		SetDisplayName(*user.DisplayName).
		SetEtag(*user.Etag).
		SetEmail(*user.Email).
		SetStatus(string(*user.Status))

	return db.Mutate(ctx, creation, ent.IsNotFound, entToUser)
}

func (this *UserEntRepository) Update(ctx context.Context, user domain.User, prevEtag model.Etag) (*domain.User, error) {
	update := this.client.User.UpdateOneID(*user.Id).
		SetNillableAvatarURL(user.AvatarUrl).
		SetNillableDisplayName(user.DisplayName).
		SetNillableEmail(user.Email).
		SetNillableStatus((*string)(user.Status)).
		// IMPORTANT: Must have!
		Where(entUser.EtagEQ(prevEtag))

	if len(update.Mutation().Fields()) > 0 {
		update.
			SetEtag(*user.Etag).
			SetUpdatedAt(time.Now())
	}

	return db.Mutate(ctx, update, ent.IsNotFound, entToUser)
}

func (this *UserEntRepository) DeleteHard(ctx context.Context, param it.DeleteParam) (int, error) {
	return this.client.User.Delete().
		Where(entUser.ID(param.Id)).
		Exec(ctx)
}

func (this *UserEntRepository) Exists(ctx context.Context, id model.Id) (bool, error) {
	return this.client.User.Query().
		Where(entUser.ID(id)).
		Exist(ctx)
}

func (this *UserEntRepository) ExistsMulti(ctx context.Context, ids []model.Id) (existing []model.Id, notExisting []model.Id, err error) {
	dbEntities, err := this.client.User.Query().
		Where(entUser.IDIn(ids...)).
		Select(entUser.FieldID).
		All(ctx)

	if err != nil {
		return nil, nil, err
	}

	existing = array.Map(dbEntities, func(entity *ent.User) model.Id {
		return entity.ID
	})

	notExisting = array.Filter(ids, func(id model.Id) bool {
		return !array.Contains(existing, id)
	})

	return existing, notExisting, nil
}

func (this *UserEntRepository) FindById(ctx context.Context, param it.FindByIdParam) (*domain.User, error) {
	query := this.client.User.Query().
		Where(entUser.ID(param.Id)).
		WithGroups().
		WithOrgs()

	if param.Status != nil {
		query = query.Where(entUser.StatusEQ(string(*param.Status)))
	}

	return db.FindOne(ctx, query, ent.IsNotFound, entToUser)
}

func (this *UserEntRepository) FindByEmail(ctx context.Context, param it.FindByEmailParam) (*domain.User, error) {
	query := this.client.User.Query().
		Where(entUser.EmailEQ(param.Email)).
		WithGroups().
		WithOrgs()

	if param.Status != nil {
		query = query.Where(entUser.StatusEQ(string(*param.Status)))
	}

	return db.FindOne(ctx, query, ent.IsNotFound, entToUser)
}

func (this *UserEntRepository) ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors) {
	return db.ParseSearchGraphStr[ent.User, domain.User](criteria, entUser.Label)
}

func (this *UserEntRepository) Search(
	ctx context.Context,
	param it.SearchParam,
) (*crud.PagedResult[domain.User], error) {
	query := this.client.User.Query()

	if param.WithGroups {
		query = query.WithGroups()
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
		entToUsers,
	)
}

func BuildUserDescriptor() *orm.EntityDescriptor {
	entity := ent.User{}
	builder := orm.DescribeEntity(entUser.Label).
		Aliases("users").
		Field(entUser.FieldAvatarURL, entity.AvatarURL).
		Field(entUser.FieldCreatedAt, entity.CreatedAt).
		Field(entUser.FieldDisplayName, entity.DisplayName).
		Field(entUser.FieldEmail, entity.Email).
		Field(entUser.FieldEtag, entity.Etag).
		Field(entUser.FieldID, entity.ID).
		Field(entUser.FieldStatus, entity.Status).
		Field(entUser.FieldUpdatedAt, entity.UpdatedAt).
		Edge(entUser.EdgeGroups, orm.ToEdgePredicate(entUser.HasGroupsWith)).
		Edge(entUser.EdgeOrgs, orm.ToEdgePredicate(entUser.HasOrgsWith))
		// TODO: Use for hierarchy
		//OrderByEdge(entUser.EdgeUserStatus, entUser.NewUserStatusStepNikki)

	return builder.Descriptor()
}
