package domain

import (
	"github.com/thoas/go-funk"
	"go.bryk.io/pkg/errors"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	entHistory "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/permissionhistory"
)

type History struct {
	model.ModelBase
	model.AuditableBase

	ApproverId      *model.Id               `json:"approverId,omitempty"`
	ApproverEmail   *string                 `json:"approverEmail,omitempty"`
	Effect          *HistoryEffect          `json:"effect,omitempty"`
	EntitlementId   *model.Id               `json:"entitlementId,omitempty"`
	EntitlementExpr *string                 `json:"entitlementExpr,omitempty"`
	GrantRequestId  *model.Id               `json:"grantRequestId,omitempty"`
	RevokeRequestId *model.Id               `json:"revokeRequestId,omitempty"`
	Reason          *HistoryReason          `json:"reason,omitempty"`
	ReceiverId      *model.Id               `json:"receiverId,omitempty"`
	ReceiverEmail   *string                 `json:"receiverEmail,omitempty"`
	ResourceId      *model.Id               `json:"resourceId,omitempty"`
	RoleId          *model.Id               `json:"roleId,omitempty"`
	RoleName        *string                 `json:"roleName,omitempty"`
	RoleSuiteId     *model.Id               `json:"roleSuiteId,omitempty"`
	RoleSuiteName   *string                 `json:"roleSuiteName,omitempty"`
	SubjectType     *EntitlementSubjectType `json:"subjectType,omitempty"`
	SubjectRef      *string                 `json:"subjectRef,omitempty"`
	ScopeRef        *string                 `json:"scopeRef,omitempty"`
}

func (this *History) Validate(forEdit bool) ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdPtrValidateRule(&this.ApproverId, true),
		model.IdPtrValidateRule(&this.EntitlementId, true),
		model.IdPtrValidateRule(&this.GrantRequestId, true),
		model.IdPtrValidateRule(&this.RevokeRequestId, true),
		model.IdPtrValidateRule(&this.ReceiverId, true),
		model.IdPtrValidateRule(&this.ResourceId, true),
		model.IdPtrValidateRule(&this.RoleId, true),
		model.IdPtrValidateRule(&this.RoleSuiteId, true),
		HistoryEffectValidateRule(&this.Effect),
		HistoryReasonValidateRule(&this.Reason),
	}

	return val.ApiBased.ValidateStruct(this, rules...)
}

type HistoryEffect string

const (
	HistoryEffectGrant  = HistoryEffect(entHistory.EffectGrant)
	HistoryEffectRevoke = HistoryEffect(entHistory.EffectRevoke)
)

func (this HistoryEffect) Validate() error {
	switch this {
	case HistoryEffectGrant, HistoryEffectRevoke:
		return nil
	default:
		return errors.Errorf("invalid history effect value: %s", this)
	}
}

func (this HistoryEffect) String() string {
	return string(this)
}

func WrapHistoryEffect(s string) *HistoryEffect {
	st := HistoryEffect(s)
	return &st
}

func WrapHistoryEffectEnt(s entHistory.Effect) *HistoryEffect {
	st := HistoryEffect(s)
	return &st
}

func HistoryEffectValidateRule(field **HistoryEffect) *val.FieldRules {
	return val.Field(field,
		val.NotEmpty,
		val.OneOf(HistoryEffectGrant, HistoryEffectRevoke),
	)
}

type HistoryReason string

const (
	HistoryReasonEntAdded   = HistoryReason(entHistory.ReasonEntAdded)
	HistoryReasonEntRemoved = HistoryReason(entHistory.ReasonEntRemoved)
	HistoryReasonEntDeleted = HistoryReason(entHistory.ReasonEntDeleted)

	HistoryReasonEntAddedGroup   = HistoryReason(entHistory.ReasonEntAddedGroup)
	HistoryReasonEntRemovedGroup = HistoryReason(entHistory.ReasonEntRemovedGroup)
	HistoryReasonEntDeletedGroup = HistoryReason(entHistory.ReasonEntDeletedGroup)

	HistoryReasonEntAddedRole   = HistoryReason(entHistory.ReasonEntAddedRole)
	HistoryReasonEntRemovedRole = HistoryReason(entHistory.ReasonEntRemovedRole)
	HistoryReasonEntDeletedRole = HistoryReason(entHistory.ReasonEntDeletedRole)

	HistoryReasonEntAddedRoleGroup    = HistoryReason(entHistory.ReasonEntAddedRoleGroup)
	HistoryReasonEntRemovedRoleGroup  = HistoryReason(entHistory.ReasonEntRemovedRoleGroup)
	HistoryReasonEntDeletedRoleGroup  = HistoryReason(entHistory.ReasonEntDeletedRoleGroup)
	HistoryReasonEntDeletedSuiteGroup = HistoryReason(entHistory.ReasonEntDeletedSuiteGroup)

	HistoryReasonRoleAdded   = HistoryReason(entHistory.ReasonRoleAdded)
	HistoryReasonRoleRemoved = HistoryReason(entHistory.ReasonRoleRemoved)
	HistoryReasonRoleDeleted = HistoryReason(entHistory.ReasonRoleDeleted)

	HistoryReasonRoleAddedGroup   = HistoryReason(entHistory.ReasonRoleAddedGroup)
	HistoryReasonRoleRemovedGroup = HistoryReason(entHistory.ReasonRoleRemovedGroup)
	HistoryReasonRoleDeletedGroup = HistoryReason(entHistory.ReasonRoleDeletedGroup)

	HistoryReasonSuiteAdded   = HistoryReason(entHistory.ReasonSuiteAdded)
	HistoryReasonSuiteRemoved = HistoryReason(entHistory.ReasonSuiteRemoved)
	HistoryReasonSuiteDeleted = HistoryReason(entHistory.ReasonSuiteDeleted)

	HistoryReasonSuiteMoreRole      = HistoryReason(entHistory.ReasonSuiteMoreRole)
	HistoryReasonSuiteLessRole      = HistoryReason(entHistory.ReasonSuiteLessRole)
	HistoryReasonSuiteMoreRoleGroup = HistoryReason(entHistory.ReasonSuiteMoreRoleGroup)
	HistoryReasonSuiteLessRoleGroup = HistoryReason(entHistory.ReasonSuiteLessRoleGroup)

	HistoryReasonSuiteAddedGroup   = HistoryReason(entHistory.ReasonSuiteAddedGroup)
	HistoryReasonSuiteRemovedGroup = HistoryReason(entHistory.ReasonSuiteRemovedGroup)
	HistoryReasonSuiteDeletedGroup = HistoryReason(entHistory.ReasonSuiteDeletedGroup)
)

var reasonValues = []any{
	HistoryReasonEntAdded, HistoryReasonEntRemoved, HistoryReasonEntDeleted,
	HistoryReasonEntAddedGroup, HistoryReasonEntRemovedGroup, HistoryReasonEntDeletedGroup,
	HistoryReasonEntAddedRole, HistoryReasonEntRemovedRole, HistoryReasonEntDeletedRole,
	HistoryReasonEntAddedRoleGroup, HistoryReasonEntRemovedRoleGroup, HistoryReasonEntDeletedRoleGroup, HistoryReasonEntDeletedSuiteGroup,
	HistoryReasonRoleAdded, HistoryReasonRoleRemoved, HistoryReasonRoleDeleted,
	HistoryReasonRoleAddedGroup, HistoryReasonRoleRemovedGroup, HistoryReasonRoleDeletedGroup,
	HistoryReasonSuiteAdded, HistoryReasonSuiteRemoved, HistoryReasonSuiteDeleted,
	HistoryReasonSuiteMoreRole, HistoryReasonSuiteLessRole, HistoryReasonSuiteMoreRoleGroup, HistoryReasonSuiteLessRoleGroup,
	HistoryReasonSuiteAddedGroup, HistoryReasonSuiteRemovedGroup, HistoryReasonSuiteDeletedGroup,
}

func (this HistoryReason) Validate() error {
	if !funk.Contains(reasonValues, this) {
		return errors.Errorf("invalid history reason value: %s", this)
	}
	return nil
}

func (this HistoryReason) String() string {
	return string(this)
}

func WrapHistoryReason(s string) *HistoryReason {
	st := HistoryReason(s)
	return &st
}

func WrapHistoryReasonEnt(s entHistory.Reason) *HistoryReason {
	st := HistoryReason(s)
	return &st
}

func HistoryReasonValidateRule(field **HistoryReason) *val.FieldRules {
	return val.Field(field,
		val.NotEmpty,
		val.OneOf(reasonValues...),
	)
}
