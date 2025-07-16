package repository

import (
	"context"
	"time"

	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/authenticate/domain"
	"github.com/sky-as-code/nikki-erp/modules/authenticate/infra/ent"
	entPass "github.com/sky-as-code/nikki-erp/modules/authenticate/infra/ent/passwordstore"
	it "github.com/sky-as-code/nikki-erp/modules/authenticate/interfaces/password"
	db "github.com/sky-as-code/nikki-erp/modules/core/database"
)

func NewPasswordStoreEntRepository(client *ent.Client) it.PasswordStoreRepository {
	return &PasswordStoreEntRepository{
		client: client,
	}
}

type PasswordStoreEntRepository struct {
	client *ent.Client
}

func (this *PasswordStoreEntRepository) Create(ctx context.Context, pass domain.PasswordStore) (*domain.PasswordStore, error) {
	creation := this.client.PasswordStore.Create().
		SetID(*pass.Id).
		SetNillablePasswordExpiredAt(pass.PasswordExpiredAt).
		SetNillablePasswordotp(pass.Passwordotp).
		SetNillablePasswordotpExpiredAt(pass.PasswordotpExpiredAt).
		SetPasswordotpRecovery(pass.PasswordotpRecovery).
		SetNillablePasswordtmp(pass.Passwordtmp).
		SetNillablePasswordtmpExpiredAt(pass.PasswordtmpExpiredAt).
		SetNillableSubjectSourceRef(pass.SubjectSourceRef).
		SetSubjectRef(*pass.SubjectRef).
		SetSubjectType(pass.SubjectType.String())

	if pass.Password != nil {
		creation = creation.
			SetPassword(*pass.Password).
			SetPasswordUpdatedAt(*pass.PasswordUpdatedAt)
	}

	return db.Mutate(ctx, creation, ent.IsNotFound, entToPasswordStore)
}

func (this *PasswordStoreEntRepository) Update(ctx context.Context, pass domain.PasswordStore) (*domain.PasswordStore, error) {
	update := this.client.PasswordStore.UpdateOneID(*pass.Id)

	if pass.Password != nil {
		pass.PasswordUpdatedAt = util.ToPtr(time.Now())
		update = update.
			SetPassword(*pass.Password).
			SetPasswordUpdatedAt(*pass.PasswordUpdatedAt).
			SetNillablePasswordExpiredAt(pass.PasswordExpiredAt)
	}

	if pass.Passwordotp != nil {
		if len(*pass.Passwordotp) > 0 {
			update = update.
				SetPasswordotp(*pass.Passwordotp).
				SetPasswordotpExpiredAt(*pass.PasswordotpExpiredAt)
		} else {
			update = update.ClearPasswordotp().ClearPasswordotpExpiredAt()
		}
	}

	// TODO: Refactor this
	if pass.PasswordotpRecovery != nil {
		if len(pass.PasswordotpRecovery) > 0 {
			update = update.
				SetPasswordotpRecovery(pass.PasswordotpRecovery)
		} else {
			update = update.ClearPasswordotpRecovery()
		}
	}

	if pass.Passwordtmp != nil {
		if len(*pass.Passwordtmp) > 0 {
			update = update.
				SetPasswordtmp(*pass.Passwordtmp).
				SetPasswordtmpExpiredAt(*pass.PasswordtmpExpiredAt)
		} else {
			update = update.ClearPasswordtmp().ClearPasswordtmpExpiredAt()
		}
	}

	if pass.PasswordExpiredAt != nil && *pass.PasswordExpiredAt == model.ZeroTime {
		update = update.ClearPasswordExpiredAt()
	}

	return db.Mutate(ctx, update, ent.IsNotFound, entToPasswordStore)
}

func (this *PasswordStoreEntRepository) FindBySubject(ctx context.Context, param it.FindBySubjectParam) (*domain.PasswordStore, error) {
	query := this.client.PasswordStore.Query().
		Where(
			entPass.SubjectRef(param.SubjectRef),
			entPass.SubjectType(param.SubjectType.String()),
		)

	return db.FindOne(ctx, query, ent.IsNotFound, entToPasswordStore)
}

func BuildPasswordStoreDescriptor() *orm.EntityDescriptor {
	entity := ent.PasswordStore{}
	builder := orm.DescribeEntity(entPass.Label).
		/*
		 * DO NOT include sensitive fields in the descriptor (passwords, recovery codes, etc.)
		 */
		Field(entPass.FieldID, entity.ID).
		Field(entPass.FieldPasswordExpiredAt, entity.PasswordExpiredAt).
		Field(entPass.FieldPasswordUpdatedAt, entity.PasswordUpdatedAt).
		Field(entPass.FieldPasswordotpExpiredAt, entity.PasswordotpExpiredAt).
		Field(entPass.FieldPasswordtmpExpiredAt, entity.PasswordtmpExpiredAt).
		Field(entPass.FieldSubjectRef, entity.SubjectRef).
		Field(entPass.FieldSubjectType, entity.SubjectType).
		Field(entPass.FieldSubjectSourceRef, entity.SubjectSourceRef)

	return builder.Descriptor()
}
