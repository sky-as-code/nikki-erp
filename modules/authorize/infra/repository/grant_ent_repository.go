package repository

import (
	"github.com/sky-as-code/nikki-erp/common/orm"

	ent "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent"
	entGrantRequest "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/grantrequest"
	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/grant_request"
)

func NewGrantRequestEntRepository(client *ent.Client) it.GrantRequestRepository {
	return &GrantRequestEntRepository{
		client: client,
	}
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
		Field(entGrantRequest.FieldStatus, entity.Status)

	return builder.Descriptor()
}
