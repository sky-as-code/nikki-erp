package repository

import (
	"context"

	"github.com/sky-as-code/nikki-erp/common/crud"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
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
		SetCreatedBy(user.CreatedBy.String()).
		SetDisplayName(*user.DisplayName).
		SetEtag(user.Etag.String()).
		SetEmail(*user.Email).
		SetMustChangePassword(*user.MustChangePassword).
		SetPasswordHash(*user.PasswordHash).
		SetPasswordChangedAt(*user.PasswordChangedAt).
		SetStatus(entUser.Status(*user.Status))

	entUser, err := creation.Save(ctx)
	if err != nil {
		return nil, err
	}
	newUser := entToUser(entUser)

	return newUser, nil
}

func (this *UserEntRepository) Update(ctx context.Context, user domain.User) (*domain.User, error) {
	update := this.client.User.UpdateOneID(user.Id.String()).
		SetNillableAvatarURL(user.AvatarUrl).
		SetNillableDisplayName(user.DisplayName).
		SetEtag(user.Etag.String()).
		SetNillableEmail(user.Email).
		SetNillablePasswordHash(user.PasswordHash).
		SetNillablePasswordChangedAt(user.PasswordChangedAt).
		SetNillableStatus((*entUser.Status)(user.Status)).
		SetUpdatedBy(user.UpdatedBy.String())

	entUser, err := update.Save(ctx)
	if err != nil {
		return nil, err
	}
	modifiedUser := entToUser(entUser)
	return modifiedUser, nil
}

func (this *UserEntRepository) Delete(ctx context.Context, id model.Id) error {
	err := this.client.User.DeleteOneID(id.String()).
		Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (this *UserEntRepository) FindById(ctx context.Context, id model.Id) (*domain.User, error) {
	user, err := this.client.User.Query().
		Where(entUser.ID(id.String())).
		WithGroups().
		WithOrgs().
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return entToUser(user), nil
}

func (this *UserEntRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	user, err := this.client.User.Query().
		Where(entUser.EmailEQ(email)).
		WithGroups().
		WithOrgs().
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return entToUser(user), nil
}

func (this *UserEntRepository) Search(
	ctx context.Context, criteria *orm.SearchGraph, opts *crud.PagingOptions,
) (*crud.PagedResult[*domain.User], error) {
	predicate, err := criteria.ToPredicate()
	if err != nil {
		return nil, err
	}

	wholeQuery := this.client.User.Query().
		Where(predicate)
	pagedQuery := wholeQuery.
		Offset(opts.Page * opts.Size).
		Limit(opts.Size)

	total, err := wholeQuery.Clone().Count(ctx)
	if err != nil {
		return nil, err
	}

	dbUsers, err := pagedQuery.All(ctx)
	if err != nil {
		return nil, err
	}

	return &crud.PagedResult[*domain.User]{
		Items: entToUsers(dbUsers),
		Total: total,
	}, nil
}

func BuildUserDescriptor() *orm.EntityDescriptor {
	entity := ent.User{}
	builder := orm.DescribeEntity(entUser.Label).
		Field(entUser.FieldAvatarURL, entity.AvatarURL).
		Field(entUser.FieldCreatedAt, entity.CreatedAt).
		Field(entUser.FieldCreatedBy, entity.CreatedBy).
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
		Field(entUser.FieldUpdatedBy, entity.UpdatedBy).
		Edge(entUser.EdgeGroups, orm.ToEdgePredicate(entUser.HasGroupsWith)).
		Edge(entUser.EdgeOrgs, orm.ToEdgePredicate(entUser.HasOrgsWith))

	return builder.Descriptor()
}
