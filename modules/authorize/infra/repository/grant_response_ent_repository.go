package repository

import (
	"context"
	"time"

	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	ent "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent"
	entGrantResponse "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/grantresponse"
	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/grant_response"
	"github.com/sky-as-code/nikki-erp/modules/core/database"
)

func NewGrantResponseEntRepository(client *ent.Client) it.GrantResponseRepository {
	return &GrantResponseEntRepository{
		client: client,
	}
}

func (this *GrantResponseEntRepository) Create(ctx context.Context, grantResponse domain.GrantResponse) (*domain.GrantResponse, error) {
	creation := this.client.GrantResponse.Create().
		SetID(*grantResponse.Id).
		SetRequestID(*grantResponse.RequestId).
		SetIsApproved(*grantResponse.IsApproved).
		SetResponderID(*grantResponse.ResponderId).
		SetNillableReason(grantResponse.Reason).
		SetCreatedAt(time.Now()).
		SetEtag(*grantResponse.Etag)

	return database.Mutate(ctx, creation, ent.IsNotFound, entToGrantResponse)
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
