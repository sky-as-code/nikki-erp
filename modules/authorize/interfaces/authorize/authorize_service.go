package authorize

import (
	"context"
	"regexp"

	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/validator"

	domain "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
)

type AuthorizeService interface {
	IsAuthorized(ctx context.Context, query IsAuthorizedQuery) (*IsAuthorizedResult, error)
}

type IsAuthorizedQuery struct {
	ActionName   string               `json:"actionName"`
	ResourceName string               `json:"resourceName"`
	ScopeRef     string               `json:"scopeRef"`
	SubjectType  SubjectTypeAuthorize `json:"subjectType"`
	SubjectRef   string               `json:"subjectRef"`
}

func (this IsAuthorizedQuery) Validate() fault.ValidationErrors {
	rules := []*validator.FieldRules{
		validator.Field(&this.ActionName,
			validator.NotEmpty,
			validator.RegExp(regexp.MustCompile(`^[a-zA-Z0-9]+$`)), // alphanumeric
			validator.Length(1, model.MODEL_RULE_TINY_NAME_LENGTH),
		),
		validator.Field(&this.ResourceName,
			validator.NotEmpty,
			validator.RegExp(regexp.MustCompile(`^[a-zA-Z0-9]+$`)), // alphanumeric
			validator.Length(1, model.MODEL_RULE_TINY_NAME_LENGTH),
		),
		validator.Field(&this.ScopeRef,
			validator.When(this.ResourceName != "",
				validator.RegExp(regexp.MustCompile(`^[a-zA-Z0-9]+$`)), // alphanumeric
				validator.Length(0, model.MODEL_RULE_ULID_LENGTH),
			),
		),
		SubjectTypeValidateRule(&this.SubjectType),
		model.IdValidateRule(&this.SubjectRef, true),
	}

	return validator.ApiBased.ValidateStruct(&this, rules...)
}

func SubjectTypeValidateRule(field *SubjectTypeAuthorize) *validator.FieldRules {
	return validator.Field(field,
		validator.NotEmpty,
		validator.OneOf(
			SubjectTypeUser,
			SubjectTypeGroup,
			SubjectTypeRole,
			SubjectTypeSuite,
			SubjectTypeCustom,
		),
	)
}

type IsAuthorizedResult struct {
	Decision    *string            `json:"decision,omitempty"`
	ClientError *fault.ClientError `json:"error,omitempty"`
}

const (
	DecisionAllow = "allow"
	DecisionDeny  = "deny"
)

type Subject struct {
	Type SubjectTypeAuthorize `param:"type" json:"type"`
	Ref  string               `param:"ref" json:"ref"`
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
