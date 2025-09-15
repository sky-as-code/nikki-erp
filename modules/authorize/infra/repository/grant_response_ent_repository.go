package repository

import (
	"time"

	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/core/database"

	"github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	ent "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent"
	entGrantResponse "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/grantresponse"
	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/grant_response"
)

func NewGrantResponseEntRepository(client *ent.Client) it.GrantResponseRepository {
	return &GrantResponseEntRepository{
		client: client,
	}
}

func (this *GrantResponseEntRepository) Create(ctx crud.Context, grantResponse domain.GrantResponse) (*domain.GrantResponse, error) {
	var creation *ent.GrantResponseCreate

	if tx := ctx.GetDbTranx().(*ent.Tx); tx != nil {
		creation = tx.GrantResponse.Create()
	} else {
		creation = this.client.GrantResponse.Create()
	}

	creation = creation.
		SetID(*grantResponse.Id).
		SetRequestID(*grantResponse.RequestId).
		SetIsApproved(*grantResponse.IsApproved).
		SetResponderID(*grantResponse.ResponderId).
		SetNillableReason(grantResponse.Reason).
		SetCreatedAt(time.Now()).
		SetEtag(*grantResponse.Etag)

	return database.Mutate(ctx, creation, ent.IsNotFound, entToGrantResponse)
}

func (this *GrantResponseEntRepository) FindByRequestId(ctx crud.Context, requestId model.Id) ([]domain.GrantResponse, error) {
	query := this.client.GrantResponse.Query().
		Where(
			entGrantResponse.RequestIDEQ(requestId),
		)

	return database.List(ctx, query, entToGrantResponses)
}

type GrantResponseEntRepository struct {
	client *ent.Client
}

func BuildGrantResponseDescriptor() *orm.EntityDescriptor {
	entity := ent.GrantResponse{}
	builder := orm.DescribeEntity(entGrantResponse.Label).
		Aliases("grant_responses").
		Field(entGrantResponse.FieldID, entity.ID).
		Field(entGrantResponse.FieldRequestID, entity.RequestID).
		Field(entGrantResponse.FieldIsApproved, entity.IsApproved).
		Field(entGrantResponse.FieldReason, entity.Reason).
		Field(entGrantResponse.FieldResponderID, entity.ResponderID).
		Field(entGrantResponse.FieldCreatedAt, entity.CreatedAt).
		Field(entGrantResponse.FieldEtag, entity.Etag)

	return builder.Descriptor()
}
