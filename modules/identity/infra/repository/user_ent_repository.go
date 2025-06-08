package repository

import (
	"context"
	"time"

	"github.com/sky-as-code/nikki-erp/common/crud"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
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
		SetID(user.Id.String()).
		SetNillableAvatarURL(user.AvatarUrl).
		SetDisplayName(*user.DisplayName).
		SetEtag(user.Etag.String()).
		SetEmail(*user.Email).
		SetMustChangePassword(*user.MustChangePassword).
		SetPasswordHash(*user.PasswordHash).
		SetPasswordChangedAt(*user.PasswordChangedAt).
		SetStatus(entUser.Status(*user.Status))

	return db.Mutate(ctx, creation, entToUser)
}

func (this *UserEntRepository) Update(ctx context.Context, user domain.User) (*domain.User, error) {
	update := this.client.User.UpdateOneID(user.Id.String()).
		SetNillableAvatarURL(user.AvatarUrl).
		SetNillableDisplayName(user.DisplayName).
		SetNillableEmail(user.Email).
		SetNillablePasswordHash(user.PasswordHash).
		SetNillablePasswordChangedAt(user.PasswordChangedAt).
		SetNillableStatus((*entUser.Status)(user.Status))

	if len(update.Mutation().Fields()) > 0 {
		update.
			SetEtag(user.Etag.String()).
			SetUpdatedAt(time.Now())
	}

	return db.Mutate(ctx, update, entToUser)
}

func (this *UserEntRepository) Delete(ctx context.Context, param it.DeleteUserParam) error {
	return db.Delete[ent.User](ctx, this.client.User.DeleteOneID(param.Id.String()))
}

func (this *UserEntRepository) FindById(ctx context.Context, param it.FindByIdParam) (*domain.User, error) {
	query := this.client.User.Query().
		Where(entUser.ID(param.Id.String())).
		WithGroups().
		WithOrgs()

	return db.FindOne(ctx, query, entToUser)
}

func (this *UserEntRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := this.client.User.Query().
		Where(entUser.EmailEQ(email)).
		WithGroups().
		WithOrgs()

	return db.FindOne(ctx, query, entToUser)
}

func (this *UserEntRepository) ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors) {
	return db.ParseSearchGraphStr[ent.User, domain.User](criteria)
}

func (this *UserEntRepository) Search(
	ctx context.Context,
	predicate *orm.Predicate,
	order []orm.OrderOption,
	opts crud.PagingOptions,
) (*crud.PagedResult[domain.User], error) {

	return db.Search(
		ctx,
		predicate,
		order,
		opts,
		this.client.User.Query(),
		entToUsers,
	)
}

func BuildUserDescriptor() *orm.EntityDescriptor {
	entity := ent.User{}
	builder := orm.DescribeEntity(entUser.Label).
		Field(entUser.FieldAvatarURL, entity.AvatarURL).
		Field(entUser.FieldCreatedAt, entity.CreatedAt).
		Field(entUser.FieldDisplayName, entity.DisplayName).
		Field(entUser.FieldEmail, entity.Email).
		Field(entUser.FieldEtag, entity.Etag).
		Field(entUser.FieldFailedLoginAttempts, entity.FailedLoginAttempts).
		Field(entUser.FieldID, entity.ID).
		Field(entUser.FieldLastLoginAt, entity.LastLoginAt).
		Field(entUser.FieldLockedUntil, entity.LockedUntil).
		Field(entUser.FieldMustChangePassword, entity.MustChangePassword).
		Field(entUser.FieldStatus, entity.Status).
		Field(entUser.FieldUpdatedAt, entity.UpdatedAt).
		Edge(entUser.EdgeGroups, orm.ToEdgePredicate(entUser.HasGroupsWith)).
		Edge(entUser.EdgeOrgs, orm.ToEdgePredicate(entUser.HasOrgsWith))

	return builder.Descriptor()
}
