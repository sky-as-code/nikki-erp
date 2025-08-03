package domain

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	entEntitlementAssignment "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/entitlementassignment"
)

type EntitlementAssignment struct {
	model.ModelBase

	SubjectType   *EntitlementAssignmentSubjectType `json:"subjectType,omitempty"`
	SubjectRef    *string                           `json:"subjectRef,omitempty"`
	ActionName    *string                           `json:"actionName,omitempty"`
	ResourceName  *string                           `json:"resourceName,omitempty"`
	ResolvedExpr  *string                           `json:"resolvedExpr,omitempty"`
	EntitlementId *model.Id                         `json:"entitlementId,omitempty"`

	Entitlement *Entitlement `json:"entitlement,omitempty" model:"-"` // TODO: Handle copy
	Role        *Role        `json:"role,omitempty" model:"-"`
}

type EntitlementAssignmentSubjectType entEntitlementAssignment.SubjectType

const (
	EntitlementAssignmentSubjectTypeNikkiUser  = EntitlementAssignmentSubjectType(entEntitlementAssignment.SubjectTypeNikkiUser)
	EntitlementAssignmentSubjectTypeNikkiGroup = EntitlementAssignmentSubjectType(entEntitlementAssignment.SubjectTypeNikkiGroup)
	EntitlementAssignmentSubjectTypeNikkiRole  = EntitlementAssignmentSubjectType(entEntitlementAssignment.SubjectTypeNikkiRole)
	EntitlementAssignmentSubjectTypeCustom     = EntitlementAssignmentSubjectType(entEntitlementAssignment.SubjectTypeCustom)
)

func (this EntitlementAssignmentSubjectType) String() string {
	return string(this)
}

func WrapEntitlementAssignmentSubjectType(s string) *EntitlementAssignmentSubjectType {
	st := EntitlementAssignmentSubjectType(s)
	return &st
}

func WrapEntitlementAssignmentSubjectTypeEnt(s entEntitlementAssignment.SubjectType) *EntitlementAssignmentSubjectType {
	st := EntitlementAssignmentSubjectType(s)
	return &st
}
