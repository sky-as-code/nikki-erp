package domain

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	entGrantRequest "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/grantrequest"
)

type GrantRequest struct {
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

type GrantRequestTargetType entGrantRequest.TargetType

const (
	GrantRequestTargetTypeNikkiUser  = GrantRequestTargetType(entGrantRequest.TargetTypeRole)
	GrantRequestTargetTypeNikkiGroup = GrantRequestTargetType(entGrantRequest.TargetTypeSuite)
)

func (this GrantRequestTargetType) String() string {
	return string(this)
}

func WrapGrantTargetType(s string) *GrantRequestTargetType {
	st := GrantRequestTargetType(s)
	return &st
}

func WrapGrantRequestTargetTypeEnt(s entGrantRequest.TargetType) *GrantRequestTargetType {
	st := GrantRequestTargetType(s)
	return &st
}

type GrantRequestStatus entGrantRequest.Status

const (
	PendingGrantRequestStatus  = GrantRequestStatus(entGrantRequest.StatusPending)
	ApprovedGrantRequestStatus = GrantRequestStatus(entGrantRequest.StatusApproved)
	RejectedGrantRequestStatus = GrantRequestStatus(entGrantRequest.StatusRejected)
)

func (this GrantRequestStatus) String() string {
	return string(this)
}

func WrapGrantRequestStatus(s string) *GrantRequestStatus {
	st := GrantRequestStatus(s)
	return &st
}

func WrapGrantRequestStatusEnt(s entGrantRequest.Status) *GrantRequestStatus {
	st := GrantRequestStatus(s)
	return &st
}
