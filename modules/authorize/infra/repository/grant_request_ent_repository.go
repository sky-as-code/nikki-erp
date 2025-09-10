package repository

import (
	"time"

	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/core/database"

	"github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	ent "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent"
	entGrantRequest "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/grantrequest"
	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/grant_request"
)

func NewGrantRequestEntRepository(client *ent.Client) it.GrantRequestRepository {
	return &GrantRequestEntRepository{
		client: client,
	}
}

func (this *GrantRequestEntRepository) BeginTransaction(ctx crud.Context) (*ent.Tx, error) {
	return this.client.Tx(ctx)
}

func (this *GrantRequestEntRepository) Create(ctx crud.Context, grantRequest domain.GrantRequest) (*domain.GrantRequest, error) {
	creation := this.client.GrantRequest.Create().
		SetID(*grantRequest.Id).
		SetEtag(*grantRequest.Etag).
		SetNillableAttachmentURL(grantRequest.AttachmentUrl).
		SetNillableComment(grantRequest.Comment).
		SetCreatedBy(*grantRequest.RequestorId).
		SetReceiverType(entGrantRequest.ReceiverType(*grantRequest.ReceiverType)).
		SetReceiverID(*grantRequest.ReceiverId).
		SetTargetType(entGrantRequest.TargetType(*grantRequest.TargetType)).
		SetStatus(entGrantRequest.Status(*grantRequest.Status)).
		SetNillableTargetRoleName(grantRequest.TargetRoleName).
		SetNillableTargetSuiteName(grantRequest.TargetSuiteName).
		SetCreatedAt(time.Now())

	switch *grantRequest.TargetType {
	case domain.GrantRequestTargetTypeRole:
		creation = creation.SetNillableTargetRoleID(grantRequest.TargetRef)
	case domain.GrantRequestTargetTypeSuite:
		creation = creation.SetNillableTargetSuiteID(grantRequest.TargetRef)
	}

	return database.Mutate(ctx, creation, ent.IsNotFound, entToGrantRequest)
}

func (this *GrantRequestEntRepository) FindById(ctx crud.Context, param it.FindByIdParam) (*domain.GrantRequest, error) {
	query := this.client.GrantRequest.Query().
		Where(entGrantRequest.IDEQ(param.Id))

	return database.FindOne(ctx, query, ent.IsNotFound, entToGrantRequest)
}

func (this *GrantRequestEntRepository) Update(ctx crud.Context, grantRequest domain.GrantRequest, prevEtag model.Etag) (*domain.GrantRequest, error) {
	var update *ent.GrantRequestUpdateOne

	if tx := ctx.GetDbTranx().(*ent.Tx); tx != nil {
		update = tx.GrantRequest.UpdateOneID(*grantRequest.Id)
	} else {
		update = this.client.GrantRequest.UpdateOneID(*grantRequest.Id)
	}

	update = update.
		SetStatus(entGrantRequest.Status(*grantRequest.Status)).
		Where(entGrantRequest.EtagEQ(prevEtag))

	if len(update.Mutation().Fields()) > 0 {
		update = update.SetEtag(*grantRequest.Etag)
	}

	return database.Mutate(ctx, update, ent.IsNotFound, entToGrantRequest)
}

func (this *GrantRequestEntRepository) Delete(ctx crud.Context, id model.Id) error {
	return this.client.GrantRequest.DeleteOneID(id).Exec(ctx)
}

func (this *GrantRequestEntRepository) FindPendingByReceiverAndTarget(ctx crud.Context, receiverId model.Id, targetId model.Id, targetType domain.GrantRequestTargetType) ([]*domain.GrantRequest, error) {
	query := this.client.GrantRequest.Query().
		Where(
			entGrantRequest.ReceiverIDEQ(receiverId),
			entGrantRequest.StatusEQ(entGrantRequest.StatusPending),
		)

	switch targetType {
	case domain.GrantRequestTargetTypeRole:
		query = query.Where(
			entGrantRequest.TargetTypeEQ(entGrantRequest.TargetTypeRole),
			entGrantRequest.TargetRoleIDEQ(targetId),
		)
	case domain.GrantRequestTargetTypeSuite:
		query = query.Where(
			entGrantRequest.TargetTypeEQ(entGrantRequest.TargetTypeSuite),
			entGrantRequest.TargetSuiteIDEQ(targetId),
		)
	}

	entResults, err := query.All(ctx)
	if err != nil {
		return nil, err
	}

	var results []*domain.GrantRequest
	for _, entResult := range entResults {
		results = append(results, entToGrantRequest(entResult))
	}

	return results, nil
}

type GrantRequestEntRepository struct {
	client *ent.Client
}

func BuildGrantRequestDescriptor() *orm.EntityDescriptor {
	entity := ent.GrantRequest{}
	builder := orm.DescribeEntity(entGrantRequest.Label).
		Aliases("grant_requests").
		Field(entGrantRequest.FieldID, entity.ID).
		Field(entGrantRequest.FieldComment, entity.Comment).
		Field(entGrantRequest.FieldCreatedAt, entity.CreatedAt).
		Field(entGrantRequest.FieldCreatedBy, entity.CreatedBy).
		Field(entGrantRequest.FieldEtag, entity.Etag).
		Field(entGrantRequest.FieldReceiverID, entity.ReceiverID).
		Field(entGrantRequest.FieldTargetType, entity.TargetType).
		Field(entGrantRequest.FieldTargetRoleID, entity.TargetRoleID).
		Field(entGrantRequest.FieldTargetSuiteID, entity.TargetSuiteID).
		Field(entGrantRequest.FieldStatus, entity.Status).
		Field(entGrantRequest.FieldTargetRoleName, entity.TargetRoleName).
		Field(entGrantRequest.FieldTargetSuiteName, entity.TargetSuiteName)

	return builder.Descriptor()
}
