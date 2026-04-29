package models

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

type RoleRequestStatus string
type RoleRequestType string

const (
	RoleReqStatusPending   = RoleRequestStatus("pending")
	RoleReqStatusApproved  = RoleRequestStatus("approved")
	RoleReqStatusRejected  = RoleRequestStatus("rejected")
	RoleReqStatusCancelled = RoleRequestStatus("cancelled")

	RoleReqTypeGrant  = RoleRequestType("grant")
	RoleReqTypeRevoke = RoleRequestType("revoke")
)

const (
	RoleRequestSchemaName = "authz.grant_request"

	RoleReqFieldId              = basemodel.FieldId
	RoleReqFieldAttachmentUrl   = "attachment_url"
	RoleReqFieldStatus          = "status"
	RoleReqFieldType            = "type"
	RoleReqFieldRoleId          = "role_id"
	RoleReqFieldGrantExpiresAt  = "grant_expires_at"
	RoleReqFieldRequestComment  = "request_comment"
	RoleReqFieldReceiverUserId  = "receiver_user_id"
	RoleReqFieldReceiverGroupId = "receiver_group_id"
	RoleReqFieldRequestorId     = "requestor_id"
	RoleReqFieldRejectionReason = "rejection_reason"
	RoleReqFieldRespondedAt     = "responded_at"
	RoleReqFieldResponderId     = "responder_id"

	RoleReqEdgeRole          = "role"
	RoleReqEdgeReceiverGroup = "receiver_group"
	RoleReqEdgeReceiverUser  = "receiver_user"
	RoleReqEdgeRequestor     = "requestor"
	RoleReqEdgeResponder     = "responder"
)

func RoleRequestSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(RoleRequestSchemaName).
		Label(model.LangJson{"en-US": "Grant Request"}).
		TableName("authz_grant_requests").
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Field(
			basemodel.DefineFieldId(RoleReqFieldRoleId).
				RequiredForCreate(),
		).
		Field(
			basemodel.DefineFieldId(RoleReqFieldReceiverGroupId),
		).
		Field(
			basemodel.DefineFieldId(RoleReqFieldReceiverUserId),
		).
		ExclusiveRequiredFields(RoleReqFieldReceiverGroupId, RoleReqFieldReceiverUserId).
		Field(
			dmodel.DefineField().Name(RoleReqFieldStatus).
				DataType(dmodel.FieldDataTypeEnumString([]string{
					string(RoleReqStatusPending), string(RoleReqStatusApproved), string(RoleReqStatusRejected), string(RoleReqStatusCancelled),
				})).
				RequiredForCreate().
				Default(string(RoleReqStatusPending)),
		).
		Field(
			dmodel.DefineField().Name(RoleReqFieldType).
				DataType(dmodel.FieldDataTypeEnumString([]string{
					string(RoleReqTypeGrant), string(RoleReqTypeRevoke),
				})).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().Name(RoleReqFieldAttachmentUrl).
				DataType(dmodel.FieldDataTypeUrl()),
		).
		Field(
			dmodel.DefineField().Name(RoleReqFieldGrantExpiresAt).
				DataType(dmodel.FieldDataTypeDateTime()).
				Description(model.LangJson{"en-US": "The date and time when the grant (if approved) expires."}),
		).
		Field(
			dmodel.DefineField().Name(RoleReqFieldRequestComment).
				DataType(dmodel.FieldDataTypeString(0, model.MODEL_RULE_COMMENT_LENGTH)),
		).
		Field(
			basemodel.DefineFieldId(RoleReqFieldRequestorId).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().Name(RoleReqFieldRejectionReason).
				DataType(dmodel.FieldDataTypeString(0, model.MODEL_RULE_COMMENT_LENGTH)),
		).
		Field(
			dmodel.DefineField().Name(RoleReqFieldRespondedAt).
				DataType(dmodel.FieldDataTypeDateTime()),
		).
		Field(
			basemodel.DefineFieldId(RoleReqFieldResponderId),
		).
		Extend(basemodel.AuditableModelSchemaBuilder()).
		Extend(basemodel.VersionedModelSchemaBuilder()).
		EdgeTo(
			dmodel.Edge(RoleReqEdgeRole).
				Label(model.LangJson{"en-US": "Requested Role"}).
				ManyToOne(RoleSchemaName, dmodel.DynamicFields{
					RoleReqFieldRoleId: RoleFieldId,
				}),
		).
		EdgeTo(
			dmodel.Edge(RoleReqEdgeReceiverGroup).
				Label(model.LangJson{"en-US": "Receiver Group"}).
				ManyToOne(GroupSchemaName, dmodel.DynamicFields{
					RoleReqFieldReceiverGroupId: GroupFieldId,
				}),
		).
		EdgeTo(
			dmodel.Edge(RoleReqEdgeReceiverUser).
				Label(model.LangJson{"en-US": "Receiver User"}).
				ManyToOne(UserSchemaName, dmodel.DynamicFields{
					RoleReqFieldReceiverUserId: UserFieldId,
				}),
		).
		EdgeTo(
			dmodel.Edge(RoleReqEdgeRequestor).
				Label(model.LangJson{"en-US": "Requestor"}).
				ManyToOne(UserSchemaName, dmodel.DynamicFields{
					RoleReqFieldRequestorId: UserFieldId,
				}),
		).
		EdgeTo(
			dmodel.Edge(RoleReqEdgeResponder).
				Label(model.LangJson{"en-US": "Responder"}).
				ManyToOne(UserSchemaName, dmodel.DynamicFields{
					RoleReqFieldResponderId: UserFieldId,
				}),
		)
}

// RoleRequest represents a request to assign a role to a receiver (user or group).
type RoleRequest struct {
	basemodel.DynamicModelBase
}

func NewRoleRequest() *RoleRequest {
	return &RoleRequest{basemodel.NewDynamicModel()}
}

func NewRoleRequestFrom(src dmodel.DynamicFields) *RoleRequest {
	return &RoleRequest{basemodel.NewDynamicModel(src)}
}
