package domain

import (
	"github.com/thoas/go-funk"
	"go.bryk.io/pkg/errors"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	entPermissionHistory "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/permissionhistory"
)

type PermissionHistory struct {
	model.ModelBase
	model.AuditableBase

	ApproverId              *model.Id                `json:"approverId,omitempty"`
	ApproverEmail           *string                  `json:"approverEmail,omitempty"`
	Effect                  *PermissionHistoryEffect `json:"effect,omitempty"`
	Reason                  *PermissionHistoryReason `json:"reason,omitempty"`
	EntitlementId           *model.Id                `json:"entitlementId,omitempty"`
	EntitlementExpr         *string                  `json:"entitlementExpr,omitempty"`
	EntitlementAssignmentId *model.Id                `json:"assignmentId,omitempty"`
	ResolvedExpr            *string                  `json:"resolvedExpr,omitempty"`
	ReceiverId              *model.Id                `json:"receiverId,omitempty"`
	ReceiverEmail           *string                  `json:"receiverEmail,omitempty"`
	GrantRequestId          *model.Id                `json:"grantRequestId,omitempty"`
	RevokeRequestId         *model.Id                `json:"revokeRequestId,omitempty"`
	ResourceId              *model.Id                `json:"resourceId,omitempty"`
	RoleId                  *model.Id                `json:"roleId,omitempty"`
	RoleName                *string                  `json:"roleName,omitempty"`
	RoleSuiteId             *model.Id                `json:"roleSuiteId,omitempty"`
	RoleSuiteName           *string                  `json:"roleSuiteName,omitempty"`
	ScopeRef                *string                  `json:"scopeRef,omitempty"`
	SubjectRef              *string                  `json:"subjectRef,omitempty"`
	// SubjectType     *EntitlementSubjectType  `json:"subjectType,omitempty"`
}

func (this *PermissionHistory) Validate(forEdit bool) ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdPtrValidateRule(&this.ApproverId, true),
		model.IdPtrValidateRule(&this.EntitlementId, true),
		model.IdPtrValidateRule(&this.EntitlementAssignmentId, true),
		model.IdPtrValidateRule(&this.GrantRequestId, true),
		model.IdPtrValidateRule(&this.RevokeRequestId, true),
		model.IdPtrValidateRule(&this.ReceiverId, true),
		model.IdPtrValidateRule(&this.ResourceId, true),
		model.IdPtrValidateRule(&this.RoleId, true),
		model.IdPtrValidateRule(&this.RoleSuiteId, true),
		PermissionHistoryEffectValidateRule(&this.Effect),
		HistoryReasonValidateRule(&this.Reason),
	}

	return val.ApiBased.ValidateStruct(this, rules...)
}

type PermissionHistoryEffect string

const (
	PermissionHistoryEffectGrant  = PermissionHistoryEffect(entPermissionHistory.EffectGrant)
	PermissionHistoryEffectRevoke = PermissionHistoryEffect(entPermissionHistory.EffectRevoke)
)

func (this PermissionHistoryEffect) Validate() error {
	switch this {
	case PermissionHistoryEffectGrant, PermissionHistoryEffectRevoke:
		return nil
	default:
		return errors.Errorf("invalid history effect value: %s", this)
	}
}

func (this PermissionHistoryEffect) String() string {
	return string(this)
}

func WrapHistoryEffect(s string) *PermissionHistoryEffect {
	st := PermissionHistoryEffect(s)
	return &st
}

func WrapHistoryEffectEnt(s entPermissionHistory.Effect) *PermissionHistoryEffect {
	st := PermissionHistoryEffect(s)
	return &st
}

func PermissionHistoryEffectValidateRule(field **PermissionHistoryEffect) *val.FieldRules {
	return val.Field(field,
		val.NotEmpty,
		val.OneOf(PermissionHistoryEffectGrant, PermissionHistoryEffectRevoke),
	)
}

type PermissionHistoryReason string

const (
	PermissionHistoryReasonEntAdded   = PermissionHistoryReason(entPermissionHistory.ReasonEntAdded)
	PermissionHistoryReasonEntRemoved = PermissionHistoryReason(entPermissionHistory.ReasonEntRemoved)
	PermissionHistoryReasonEntDeleted = PermissionHistoryReason(entPermissionHistory.ReasonEntDeleted)

	PermissionHistoryReasonEntAddedGroup   = PermissionHistoryReason(entPermissionHistory.ReasonEntAddedGroup)
	PermissionHistoryReasonEntRemovedGroup = PermissionHistoryReason(entPermissionHistory.ReasonEntRemovedGroup)
	PermissionHistoryReasonEntDeletedGroup = PermissionHistoryReason(entPermissionHistory.ReasonEntDeletedGroup)

	PermissionHistoryReasonEntAddedRole   = PermissionHistoryReason(entPermissionHistory.ReasonEntAddedRole)
	PermissionHistoryReasonEntRemovedRole = PermissionHistoryReason(entPermissionHistory.ReasonEntRemovedRole)
	PermissionHistoryReasonEntDeletedRole = PermissionHistoryReason(entPermissionHistory.ReasonEntDeletedRole)

	PermissionHistoryReasonEntAddedRoleGroup    = PermissionHistoryReason(entPermissionHistory.ReasonEntAddedRoleGroup)
	PermissionHistoryReasonEntRemovedRoleGroup  = PermissionHistoryReason(entPermissionHistory.ReasonEntRemovedRoleGroup)
	PermissionHistoryReasonEntDeletedRoleGroup  = PermissionHistoryReason(entPermissionHistory.ReasonEntDeletedRoleGroup)
	PermissionHistoryReasonEntDeletedSuiteGroup = PermissionHistoryReason(entPermissionHistory.ReasonEntDeletedSuiteGroup)

	PermissionHistoryReasonRoleAdded   = PermissionHistoryReason(entPermissionHistory.ReasonRoleAdded)
	PermissionHistoryReasonRoleRemoved = PermissionHistoryReason(entPermissionHistory.ReasonRoleRemoved)
	PermissionHistoryReasonRoleDeleted = PermissionHistoryReason(entPermissionHistory.ReasonRoleDeleted)

	PermissionHistoryReasonRoleAddedGroup   = PermissionHistoryReason(entPermissionHistory.ReasonRoleAddedGroup)
	PermissionHistoryReasonRoleRemovedGroup = PermissionHistoryReason(entPermissionHistory.ReasonRoleRemovedGroup)
	PermissionHistoryReasonRoleDeletedGroup = PermissionHistoryReason(entPermissionHistory.ReasonRoleDeletedGroup)

	PermissionHistoryReasonSuiteAdded   = PermissionHistoryReason(entPermissionHistory.ReasonSuiteAdded)
	PermissionHistoryReasonSuiteRemoved = PermissionHistoryReason(entPermissionHistory.ReasonSuiteRemoved)
	PermissionHistoryReasonSuiteDeleted = PermissionHistoryReason(entPermissionHistory.ReasonSuiteDeleted)

	PermissionHistoryReasonSuiteMoreRole      = PermissionHistoryReason(entPermissionHistory.ReasonSuiteMoreRole)
	PermissionHistoryReasonSuiteLessRole      = PermissionHistoryReason(entPermissionHistory.ReasonSuiteLessRole)
	PermissionHistoryReasonSuiteMoreRoleGroup = PermissionHistoryReason(entPermissionHistory.ReasonSuiteMoreRoleGroup)
	PermissionHistoryReasonSuiteLessRoleGroup = PermissionHistoryReason(entPermissionHistory.ReasonSuiteLessRoleGroup)

	PermissionHistoryReasonSuiteAddedGroup   = PermissionHistoryReason(entPermissionHistory.ReasonSuiteAddedGroup)
	PermissionHistoryReasonSuiteRemovedGroup = PermissionHistoryReason(entPermissionHistory.ReasonSuiteRemovedGroup)
	PermissionHistoryReasonSuiteDeletedGroup = PermissionHistoryReason(entPermissionHistory.ReasonSuiteDeletedGroup)
)

var reasonValues = []any{
	PermissionHistoryReasonEntAdded, PermissionHistoryReasonEntRemoved, PermissionHistoryReasonEntDeleted,
	PermissionHistoryReasonEntAddedGroup, PermissionHistoryReasonEntRemovedGroup, PermissionHistoryReasonEntDeletedGroup,
	PermissionHistoryReasonEntAddedRole, PermissionHistoryReasonEntRemovedRole, PermissionHistoryReasonEntDeletedRole,
	PermissionHistoryReasonEntAddedRoleGroup, PermissionHistoryReasonEntRemovedRoleGroup, PermissionHistoryReasonEntDeletedRoleGroup, PermissionHistoryReasonEntDeletedSuiteGroup,
	PermissionHistoryReasonRoleAdded, PermissionHistoryReasonRoleRemoved, PermissionHistoryReasonRoleDeleted,
	PermissionHistoryReasonRoleAddedGroup, PermissionHistoryReasonRoleRemovedGroup, PermissionHistoryReasonRoleDeletedGroup,
	PermissionHistoryReasonSuiteAdded, PermissionHistoryReasonSuiteRemoved, PermissionHistoryReasonSuiteDeleted,
	PermissionHistoryReasonSuiteMoreRole, PermissionHistoryReasonSuiteLessRole, PermissionHistoryReasonSuiteMoreRoleGroup, PermissionHistoryReasonSuiteLessRoleGroup,
	PermissionHistoryReasonSuiteAddedGroup, PermissionHistoryReasonSuiteRemovedGroup, PermissionHistoryReasonSuiteDeletedGroup,
}

func (this PermissionHistoryReason) Validate() error {
	if !funk.Contains(reasonValues, this) {
		return errors.Errorf("invalid history reason value: %s", this)
	}
	return nil
}

func (this PermissionHistoryReason) String() string {
	return string(this)
}

func WrapHistoryReason(s string) *PermissionHistoryReason {
	st := PermissionHistoryReason(s)
	return &st
}

func WrapHistoryReasonEnt(s entPermissionHistory.Reason) *PermissionHistoryReason {
	st := PermissionHistoryReason(s)
	return &st
}

func HistoryReasonValidateRule(field **PermissionHistoryReason) *val.FieldRules {
	return val.Field(field,
		val.NotEmpty,
		val.OneOf(reasonValues...),
	)
}
