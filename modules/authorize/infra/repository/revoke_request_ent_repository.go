package repository

import (
	"time"

	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/core/database"

	domain "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	ent "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent"
	entGrantRequest "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/grantrequest"
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

func (this *RevokeRequestEntRepository) revokeRequestClient(ctx crud.Context) *ent.RevokeRequestClient {
	tx, isOk := ctx.GetDbTranx().(*ent.Tx)
	if isOk {
		return tx.RevokeRequest
	}
	return this.client.RevokeRequest
}

func (this *RevokeRequestEntRepository) Create(ctx crud.Context, revokeRequest *domain.RevokeRequest) (*domain.RevokeRequest, error) {
	creation := this.revokeRequestClient(ctx).Create().
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

func (this *RevokeRequestEntRepository) FindById(ctx crud.Context, param it.FindByIdParam) (*domain.RevokeRequest, error) {
	query := this.revokeRequestClient(ctx).Query().
		Where(entGrantRequest.IDEQ(param.Id)).
		WithRole().
		WithRoleSuite()

	return database.FindOne(ctx, query, ent.IsNotFound, entToRevokeRequest)
}

func (this *RevokeRequestEntRepository) FindAllByTarget(ctx crud.Context, param it.FindAllByTargetParam) ([]domain.RevokeRequest, error) {
	query := this.revokeRequestClient(ctx).Query()

	targetType := domain.GrantRequestTargetType(param.TargetType)
	switch targetType {
	case domain.GrantRequestTargetTypeRole:
		query = query.Where(
			entRevokeRequest.TargetTypeEQ(entRevokeRequest.TargetTypeRole),
			entRevokeRequest.TargetRoleIDEQ(param.TargetRef),
		)
	case domain.GrantRequestTargetTypeSuite:
		query = query.Where(
			entRevokeRequest.TargetTypeEQ(entRevokeRequest.TargetTypeSuite),
			entRevokeRequest.TargetSuiteIDEQ(param.TargetRef),
		)
	}

	return database.List(ctx, query, entToRevokeRequests)
}

func (this *RevokeRequestEntRepository) UpdateTargetFields(ctx crud.Context, revokeRequest *domain.RevokeRequest, prevEtag model.Etag) error {
	update := this.revokeRequestClient(ctx).Update().
		Where(entRevokeRequest.IDEQ(*revokeRequest.Id)).
		Where(entRevokeRequest.EtagEQ(prevEtag))

	switch *revokeRequest.TargetType {
	case domain.RevokeRequestTargetTypeNikkiRole:
		if revokeRequest.TargetRoleName != nil {
			update = update.SetTargetRoleName(*revokeRequest.TargetRoleName)
		}
	case domain.RevokeRequestTargetTypeNikkiSuite:
		if revokeRequest.TargetSuiteName != nil {
			update = update.SetTargetSuiteName(*revokeRequest.TargetSuiteName)
		}
	}

	if len(update.Mutation().Fields()) > 0 {
		update = update.SetEtag(*revokeRequest.Etag)
	}

	return update.Exec(ctx)
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
