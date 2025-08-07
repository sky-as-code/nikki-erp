package domain

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	entGrantRequest "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/grantrequest"
)

type RevokeRequest struct {
	model.ModelBase
	model.AuditableBase

	Attachment  *string   `json:"attachment,omitempty"`
	Comment     *string   `json:"comment,omitempty"`
	ApprovalId  *model.Id `json:"approvalId,omitempty"`
	RequestorId *model.Id `json:"requestorId,omitempty"`
	ReceiverId  *model.Id `json:"receiverId,omitempty"`
	TargetType  *string   `json:"targetType,omitempty"` // "role" | "suite"
	TargetRef   *model.Id `json:"targetRef,omitempty"`
	ResponseId  *model.Id `json:"responseId,omitempty"` // Only set after response
	Status      *string   `json:"status,omitempty"`     // "pending" | "approved" | "rejected" | "cancelled"
	CreatedBy   *string   `json:"createdBy,omitempty"`

	Role      *Role      `json:"role,omitempty" model:"-"` // TODO: Handle copy
	RoleSuite *RoleSuite `json:"roleSuite,omitempty" model:"-"`
}

type RevokeRequestTargetType entGrantRequest.TargetType

const (
	RevokeTargetTypeNikkiUser  = RevokeRequestTargetType(entGrantRequest.TargetTypeRole)
	RevokeTargetTypeNikkiGroup = RevokeRequestTargetType(entGrantRequest.TargetTypeSuite)
)

func (this RevokeRequestTargetType) String() string {
	return string(this)
}

func WrapRevokeTargetType(s string) *RevokeRequestTargetType {
	st := RevokeRequestTargetType(s)
	return &st
}

func WrapRevokeRequestTargetTypeEnt(s entGrantRequest.TargetType) *RevokeRequestTargetType {
	st := RevokeRequestTargetType(s)
	return &st
}

type RevokeRequestStatus entGrantRequest.Status

const (
	PendingRevokeRequestStatus  = RevokeRequestStatus(entGrantRequest.StatusPending)
	ApprovedRevokeRequestStatus = RevokeRequestStatus(entGrantRequest.StatusApproved)
	RejectedRevokeRequestStatus = RevokeRequestStatus(entGrantRequest.StatusRejected)
)

func (this RevokeRequestStatus) String() string {
	return string(this)
}

func WrapRevokeRequestStatus(s string) *RevokeRequestStatus {
	st := RevokeRequestStatus(s)
	return &st
}

func WrapRevokeRequestStatusEnt(s entGrantRequest.Status) *RevokeRequestStatus {
	st := RevokeRequestStatus(s)
	return &st
}
