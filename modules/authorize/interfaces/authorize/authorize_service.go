package authorize

import (
	"context"

	"github.com/sky-as-code/nikki-erp/modules/authorize/domain"
)

type AuthorizeService interface {
	IsAuthorized(ctx context.Context, query IsAuthorizedQuery) (*IsAuthorizedResult, error)
}

type IsAuthorizedQuery struct {
	ActionName   string  `json:"actionName"`
	ResourceName string  `json:"resourceName"`
	ScopeRef     string  `json:"scopeRef"`
	Subjects     Subject `json:"subjects"`
}

type IsAuthorizedResult struct {
	Decision string `json:"decision"`
}

const (
	DecisionAllow = "allow"
	DecisionDeny  = "deny"
)

type Subject struct {
	Type *SubjectTypeAuthorize `param:"type" json:"type"`
	Ref  string                `param:"ref" json:"ref"`
}

type SubjectTypeAuthorize string

const (
	SubjectTypeUser   = SubjectTypeAuthorize("nikki_user")
	SubjectTypeGroup  = SubjectTypeAuthorize("nikki_group")
	SubjectTypeRole   = SubjectTypeAuthorize("nikki_role")
	SubjectTypeSuite  = SubjectTypeAuthorize("nikki_suite")
	SubjectTypeCustom = SubjectTypeAuthorize("custom")
)

func (this SubjectTypeAuthorize) String() string {
	return string(this)
}

func WrapSubjectType(s string) *SubjectTypeAuthorize {
	st := SubjectTypeAuthorize(s)
	return &st
}

func WrapSubjectTypeEnt(s domain.EntitlementAssignmentSubjectType) *SubjectTypeAuthorize {
	st := SubjectTypeAuthorize(s)
	return &st
}
