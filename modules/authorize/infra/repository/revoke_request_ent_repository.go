package repository

import (
	"time"

	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/core/database"

	domain "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	ent "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent"
	entRevokeRequest "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/revokerequest"
	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/revoke_request"
)

func NewRevokeRequestEntRepository(client *ent.Client) it.RevokeRequestRepository {
	return &RevokeRequestEntRepository{
		client: client,
	}
}

type RevokeRequestEntRepository struct {
	client *ent.Client
}

func (this *RevokeRequestEntRepository) BeginTransaction(ctx crud.Context) (*ent.Tx, error) {
	return this.client.Tx(ctx)
}

func (this *RevokeRequestEntRepository) Create(ctx crud.Context, revokeRequest *domain.RevokeRequest) (*domain.RevokeRequest, error) {
	var creation *ent.RevokeRequestCreate
	tx := ctx.GetDbTranx()
	
	if tx != nil {
		creation = tx.(*ent.Tx).RevokeRequest.Create()
	} else {
		creation = this.client.RevokeRequest.Create()
	}

	creation = creation.
		SetID(*revokeRequest.Id).
		SetEtag(*revokeRequest.Etag).
		SetNillableAttachmentURL(revokeRequest.AttachmentUrl).
		SetNillableComment(revokeRequest.Comment).
		SetCreatedBy(*revokeRequest.RequestorId).
		SetReceiverType(entRevokeRequest.ReceiverType(*revokeRequest.ReceiverType)).
		SetReceiverID(*revokeRequest.ReceiverId).
		SetTargetType(entRevokeRequest.TargetType(*revokeRequest.TargetType)).
		SetNillableTargetRoleName(revokeRequest.TargetRoleName).
		SetNillableTargetSuiteName(revokeRequest.TargetSuiteName).
		SetCreatedAt(time.Now())

	switch *revokeRequest.TargetType {
	case domain.RevokeRequestTargetTypeNikkiRole:
		creation = creation.SetNillableTargetRoleID(revokeRequest.TargetRef)
	case domain.RevokeRequestTargetTypeNikkiSuite:
		creation = creation.SetNillableTargetSuiteID(revokeRequest.TargetRef)
	}

	return database.Mutate(ctx, creation, ent.IsNotFound, entToRevokeRequest)
}

func BuildRevokeRequestDescriptor() *orm.EntityDescriptor {
	entity := ent.RevokeRequest{}
	builder := orm.DescribeEntity(entRevokeRequest.Label).
		Aliases("revoke_requests").
		Field(entRevokeRequest.FieldID, entity.ID).
		Field(entRevokeRequest.FieldComment, entity.Comment).
		Field(entRevokeRequest.FieldCreatedAt, entity.CreatedAt).
		Field(entRevokeRequest.FieldCreatedBy, entity.CreatedBy).
		Field(entRevokeRequest.FieldEtag, entity.Etag).
		Field(entRevokeRequest.FieldReceiverID, entity.ReceiverID).
		Field(entRevokeRequest.FieldTargetType, entity.TargetType).
		Field(entRevokeRequest.FieldTargetRoleID, entity.TargetRoleID).
		Field(entRevokeRequest.FieldTargetSuiteID, entity.TargetSuiteID).
		Field(entRevokeRequest.FieldTargetRoleName, entity.TargetRoleName).
		Field(entRevokeRequest.FieldTargetSuiteName, entity.TargetSuiteName)

	return builder.Descriptor()
}
