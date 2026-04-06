package login

import (
	"time"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/authenticate/domain"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

var authenticateCommandType = cqrs.RequestType{
	Module:    "authenticate",
	Submodule: "login",
	Action:    "doAuthenticate",
}

type AuthenticateCommand struct {
	AttemptId model.Id          `json:"attempt_id"`
	Passwords map[string]string `json:"passwords"`
}

func (AuthenticateCommand) CqrsRequestType() cqrs.RequestType {
	return authenticateCommandType
}

func (this AuthenticateCommand) Validate() ft.ClientErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.AttemptId, true),
	}

	return domain.ValidationErrorsToClientErrors(val.ApiBased.ValidateStruct(&this, rules...))
}

type AuthenticateSuccessData struct {
	AccessToken           string    `json:"access_token"`
	AccessTokenExpiresAt  time.Time `json:"access_token_expires_at"`
	RefreshToken          string    `json:"refresh_token"`
	RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at"`
}

type AuthenticateResultData struct {
	Done     bool                     `json:"done"`
	NextStep *string                  `json:"next_step,omitempty"`
	Data     *AuthenticateSuccessData `json:"data,omitempty"`
}
type AuthenticateResult = dyn.OpResult[*AuthenticateResultData]

var createLoginAttemptCommandType = cqrs.RequestType{
	Module:    "authenticate",
	Submodule: "login",
	Action:    "createAttempt",
}

type CreateLoginAttemptCommand struct {
	DeviceIp         *string            `json:"device_ip,omitempty"`
	DeviceName       *string            `json:"device_name,omitempty"`
	DeviceLocation   *string            `json:"device_location,omitempty"`
	SubjectType      domain.SubjectType `json:"subject_type"`
	SubjectSourceRef *string            `json:"subject_source_ref,omitempty"`
	Username         string             `json:"username"`
}

func (CreateLoginAttemptCommand) CqrsRequestType() cqrs.RequestType {
	return createLoginAttemptCommandType
}

type CreateLoginAttemptResultData struct {
	Attempt     domain.LoginAttempt `json:"attempt"`
	SubjectName string              `json:"subject_name"`
}
type CreateLoginAttemptResult = dyn.OpResult[*CreateLoginAttemptResultData]

var updateLoginAttemptCommandType = cqrs.RequestType{
	Module:    "authenticate",
	Submodule: "login",
	Action:    "updateAttempt",
}

type UpdateLoginAttemptCommand struct {
	Id            model.Id              `json:"id"`
	IsGenuine     *bool                 `json:"is_genuine,omitempty"`
	CurrentMethod *string               `json:"current_method,omitempty"`
	Status        *domain.AttemptStatus `json:"status,omitempty"`
}

func (UpdateLoginAttemptCommand) CqrsRequestType() cqrs.RequestType {
	return updateLoginAttemptCommandType
}

type UpdateLoginAttemptResult = dyn.OpResult[*domain.LoginAttempt]

var startLoginFlowCommandType = cqrs.RequestType{
	Module:    "authenticate",
	Submodule: "login",
	Action:    "startLoginFlow",
}

type StartLoginFlowCommand struct {
	DeviceName       *string            `json:"device_name,omitempty"`
	SubjectType      domain.SubjectType `json:"subject_type"`
	SubjectSourceRef *string            `json:"subject_source_ref,omitempty"`
	Username         string             `json:"username"`
}

func (StartLoginFlowCommand) CqrsRequestType() cqrs.RequestType {
	return startLoginFlowCommandType
}

var getAttemptByIdQueryType = cqrs.RequestType{
	Module:    "authenticate",
	Submodule: "login",
	Action:    "getAttemptById",
}

type GetAttemptByIdQuery struct {
	Id model.Id `param:"id" json:"id"`
}

func (GetAttemptByIdQuery) CqrsRequestType() cqrs.RequestType {
	return getAttemptByIdQueryType
}

func (this GetAttemptByIdQuery) Validate() ft.ClientErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.Id, true),
	}

	return domain.ValidationErrorsToClientErrors(val.ApiBased.ValidateStruct(&this, rules...))
}

type GetAttemptByIdResult = dyn.OpResult[*domain.LoginAttempt]

var refreshTokenCommandType = cqrs.RequestType{
	Module:    "authenticate",
	Submodule: "login",
	Action:    "refreshToken",
}

type RefreshTokenCommand struct {
	RefreshToken string `json:"refresh_token"`
}

func (RefreshTokenCommand) CqrsRequestType() cqrs.RequestType {
	return refreshTokenCommandType
}

type RefreshTokenResultData struct {
	AccessToken           string    `json:"access_token"`
	AccessTokenExpiresAt  time.Time `json:"access_token_expires_at"`
	RefreshToken          string    `json:"refresh_token"`
	RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at"`
}
type RefreshTokenResult = dyn.OpResult[*RefreshTokenResultData]
