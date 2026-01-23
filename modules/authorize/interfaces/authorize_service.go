package authorize

import (
	"regexp"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/common/validator"

	domain "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

func init() {
	var req cqrs.Request
	req = (*IsAuthorizedQuery)(nil)
	util.Unused(req)
}

type AuthorizeService interface {
	IsAuthorized(ctx crud.Context, query IsAuthorizedQuery) (*IsAuthorizedResult, error)
}

var isAuthorizedQueryType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "nil",
	Action:    "isAuthorized",
}

type IsAuthorizedQuery struct {
	ActionName   string               `json:"actionName"`
	ResourceName string               `json:"resourceName"`
	ScopeRef     string               `json:"scopeRef"`
	SubjectType  SubjectTypeAuthorize `json:"subjectType"`
	SubjectRef   string               `json:"subjectRef"`
}

func (IsAuthorizedQuery) CqrsRequestType() cqrs.RequestType {
	return isAuthorizedQueryType
}

func (this IsAuthorizedQuery) Validate() ft.ValidationErrors {
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
	Decision    *string         `json:"decision,omitempty"`
	ClientError *ft.ClientError `json:"error,omitempty"`
}

func (this IsAuthorizedResult) GetClientError() *ft.ClientError {
	return this.ClientError
}

func (this IsAuthorizedResult) GetHasData() bool {
	return this.Decision != nil
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
