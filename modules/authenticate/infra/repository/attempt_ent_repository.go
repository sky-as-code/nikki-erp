package repository

import (
	"context"
	"time"

	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/authenticate/domain"
	"github.com/sky-as-code/nikki-erp/modules/authenticate/infra/ent"
	entAtp "github.com/sky-as-code/nikki-erp/modules/authenticate/infra/ent/loginattempt"
	it "github.com/sky-as-code/nikki-erp/modules/authenticate/interfaces/login"
	db "github.com/sky-as-code/nikki-erp/modules/core/database"
)

func NewAttemptEntRepository(client *ent.Client) it.AttemptRepository {
	return &AttemptEntRepository{
		client: client,
	}
}

type AttemptEntRepository struct {
	client *ent.Client
}

func (this *AttemptEntRepository) Create(ctx context.Context, attempt domain.LoginAttempt) (*domain.LoginAttempt, error) {
	creation := this.client.LoginAttempt.Create().
		SetID(*attempt.Id).
		SetNillableCurrentMethod(attempt.CurrentMethod).
		SetNillableDeviceIP(attempt.DeviceIp).
		SetNillableDeviceLocation(attempt.DeviceLocation).
		SetNillableDeviceName(attempt.DeviceName).
		SetExpiredAt(*attempt.ExpiredAt).
		SetIsGenuine(*attempt.IsGenuine).
		SetMethods(attempt.Methods).
		SetStatus(attempt.Status.String()).
		SetSubjectRef(*attempt.SubjectRef).
		SetSubjectType(attempt.SubjectType.String()).
		SetNillableSubjectSourceRef(attempt.SubjectSourceRef)

	return db.Mutate(ctx, creation, ent.IsNotFound, entToAttempt)
}

func (this *AttemptEntRepository) Update(ctx context.Context, attempt domain.LoginAttempt) (*domain.LoginAttempt, error) {
	update := this.client.LoginAttempt.UpdateOneID(*attempt.Id).
		SetNillableIsGenuine(attempt.IsGenuine).
		SetNillableStatus(util.ToPtr(attempt.Status.String()))

	if attempt.CurrentMethod == nil || len(*attempt.CurrentMethod) == 0 {
		update = update.ClearCurrentMethod()
	} else {
		update = update.SetNillableCurrentMethod(attempt.CurrentMethod)
	}

	if len(update.Mutation().Fields()) > 0 {
		update = update.SetUpdatedAt(time.Now())
	}

	return db.Mutate(ctx, update, ent.IsNotFound, entToAttempt)
}

func (this *AttemptEntRepository) FindById(ctx context.Context, param it.FindByIdParam) (*domain.LoginAttempt, error) {
	query := this.client.LoginAttempt.Query().
		Where(entAtp.ID(param.Id))

	return db.FindOne(ctx, query, ent.IsNotFound, entToAttempt)
}

func BuildAttemptDescriptor() *orm.EntityDescriptor {
	entity := ent.LoginAttempt{}
	builder := orm.DescribeEntity(entAtp.Label).
		Aliases("login_attempts").
		Field(entAtp.FieldCreatedAt, entity.CreatedAt).
		Field(entAtp.FieldCurrentMethod, entity.CurrentMethod).
		Field(entAtp.FieldDeviceIP, entity.DeviceIP).
		Field(entAtp.FieldDeviceLocation, entity.DeviceLocation).
		Field(entAtp.FieldDeviceName, entity.DeviceName).
		Field(entAtp.FieldExpiredAt, entity.ExpiredAt).
		Field(entAtp.FieldID, entity.ID).
		Field(entAtp.FieldIsGenuine, entity.IsGenuine).
		Field(entAtp.FieldMethods, entity.Methods).
		Field(entAtp.FieldStatus, entity.Status).
		Field(entAtp.FieldSubjectRef, entity.SubjectRef).
		Field(entAtp.FieldSubjectType, entity.SubjectType).
		Field(entAtp.FieldSubjectSourceRef, entity.SubjectSourceRef).
		Field(entAtp.FieldUpdatedAt, entity.UpdatedAt)

	return builder.Descriptor()
}
