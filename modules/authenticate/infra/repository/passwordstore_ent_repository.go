package repository

import (
	"time"

	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/authenticate/domain"
	"github.com/sky-as-code/nikki-erp/modules/authenticate/infra/ent"
	entPass "github.com/sky-as-code/nikki-erp/modules/authenticate/infra/ent/passwordstore"
	it "github.com/sky-as-code/nikki-erp/modules/authenticate/interfaces/password"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
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

func (this *PasswordStoreEntRepository) Create(ctx crud.Context, pass domain.PasswordStore) (*domain.PasswordStore, error) {
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

func (this *PasswordStoreEntRepository) Update(ctx crud.Context, pass domain.PasswordStore) (*domain.PasswordStore, error) {
	update := this.client.PasswordStore.UpdateOneID(*pass.Id)

	if pass.Password != nil {
		pass.PasswordUpdatedAt = util.ToPtr(time.Now())
		update = update.
			SetPassword(*pass.Password).
			SetPasswordUpdatedAt(*pass.PasswordUpdatedAt)
	}

	// Password expiration may not be set together with Password in case a new policy
	// enforces with or without password expiration.
	if pass.PasswordExpiredAt != nil {
		if !model.ZeroTime.Equal(*pass.PasswordExpiredAt) {
			update = update.SetPasswordExpiredAt(*pass.PasswordExpiredAt)
		} else {
			update = update.ClearPasswordExpiredAt()
		}
	}

	// Setting or deleting OTP secret always requires the same action on OTP expiration.
	if pass.Passwordotp != nil {
		if len(*pass.Passwordotp) > 0 {
			update = update.
				SetPasswordotp(*pass.Passwordotp).
				SetPasswordotpExpiredAt(*pass.PasswordotpExpiredAt)
		} else {
			update = update.ClearPasswordotp().ClearPasswordotpExpiredAt()
		}
	}

	// OTP expiration is deleted when OTP confirmation step is done.
	if pass.PasswordotpExpiredAt != nil && model.ZeroTime.Equal(*pass.PasswordotpExpiredAt) {
		update = update.ClearPasswordotpExpiredAt()
	}

	// OTP recovery codes are cleared after all codes are used up.
	if pass.PasswordotpRecovery != nil {
		if len(pass.PasswordotpRecovery) > 0 {
			update = update.
				SetPasswordotpRecovery(pass.PasswordotpRecovery)
		} else {
			update = update.ClearPasswordotpRecovery()
		}
	}

	// Setting or deleting temp password always requires the same action on temp password expiration.
	if pass.Passwordtmp != nil {
		if len(*pass.Passwordtmp) > 0 {
			update = update.
				SetPasswordtmp(*pass.Passwordtmp).
				SetPasswordtmpExpiredAt(*pass.PasswordtmpExpiredAt)
		} else {
			update = update.ClearPasswordtmp().ClearPasswordtmpExpiredAt()
		}
	}

	return db.Mutate(ctx, update, ent.IsNotFound, entToPasswordStore)
}

func (this *PasswordStoreEntRepository) FindBySubject(ctx crud.Context, param it.FindBySubjectParam) (*domain.PasswordStore, error) {
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
