package repository

import (
	"github.com/sky-as-code/nikki-erp/common/orm"

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
		Field(entRevokeRequest.FieldStatus, entity.Status).
		Field(entRevokeRequest.FieldTargetRoleName, entity.TargetRoleName).
		Field(entRevokeRequest.FieldTargetSuiteName, entity.TargetSuiteName)

	return builder.Descriptor()
}
