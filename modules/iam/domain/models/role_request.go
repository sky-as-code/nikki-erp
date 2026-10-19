package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
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
	RoleRequestSchemaName = "iam_grant_request"

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

//go:embed role_request.json
var roleRequestSchemaJson string

func RoleRequestSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(roleRequestSchemaJson)
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
