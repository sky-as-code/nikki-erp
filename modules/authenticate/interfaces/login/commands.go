package login

import (
	"time"

	"github.com/sky-as-code/nikki-erp/common/crud"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/authenticate/domain"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
)

var authenticateCommandType = cqrs.RequestType{
	Module:    "authenticate",
	Submodule: "login",
	Action:    "doAuthenticate",
}

type AuthenticateCommand struct {
	AttemptId model.Id `json:"attemptId"`
	Username  string   `json:"username"`
	Password  string   `json:"password"`
}

func (AuthenticateCommand) CqrsRequestType() cqrs.RequestType {
	return authenticateCommandType
}

type AuthenticateSuccessData struct {
	AccessToken           string    `json:"accessToken"`
	AccessTokenExpiredAt  time.Time `json:"accessTokenExpiredAt"`
	RefreshToken          string    `json:"refreshToken"`
	RefreshTokenExpiredAt time.Time `json:"refreshTokenExpiredAt"`
}

type AuthenticateResultData struct {
	Done     bool                     `json:"done"`
	NextStep *string                  `json:"nextStep,omitempty"`
	Data     *AuthenticateSuccessData `json:"data,omitempty"`
}
type AuthenticateResult = crud.OpResult[*AuthenticateResultData]

var createLoginAttemptCommandType = cqrs.RequestType{
	Module:    "authenticate",
	Submodule: "login",
	Action:    "createAttempt",
}

type CreateLoginAttemptCommand struct {
	DeviceIp         *string            `json:"deviceIp,omitempty"`
	DeviceName       *string            `json:"deviceName,omitempty"`
	DeviceLocation   *string            `json:"deviceLocation,omitempty"`
	SubjectType      domain.SubjectType `json:"subjectType"`
	SubjectSourceRef *string            `json:"subjectSourceRef,omitempty"`
	Username         string             `json:"username"`
}

func (CreateLoginAttemptCommand) CqrsRequestType() cqrs.RequestType {
	return createLoginAttemptCommandType
}

type CreateLoginAttemptResultData struct {
	Attempt     domain.LoginAttempt `json:"attempt"`
	SubjectName string              `json:"subjectName"`
}
type CreateLoginAttemptResult = crud.OpResult[*CreateLoginAttemptResultData]

var updateLoginAttemptCommandType = cqrs.RequestType{
	Module:    "authenticate",
	Submodule: "login",
	Action:    "updateAttempt",
}

type UpdateLoginAttemptCommand struct {
	Id            model.Id              `json:"id"`
	IsGenuine     *bool                 `json:"isGenuine,omitempty"`
	CurrentMethod *string               `json:"currentMethod,omitempty"`
	Status        *domain.AttemptStatus `json:"status,omitempty"`
}

func (UpdateLoginAttemptCommand) CqrsRequestType() cqrs.RequestType {
	return updateLoginAttemptCommandType
}

type UpdateLoginAttemptResult = crud.OpResult[*domain.LoginAttempt]

var startLoginFlowCommandType = cqrs.RequestType{
	Module:    "authenticate",
	Submodule: "login",
	Action:    "startLoginFlow",
}

type StartLoginFlowCommand struct {
	DeviceName       *string            `json:"deviceName,omitempty"`
	SubjectType      domain.SubjectType `json:"subjectType"`
	SubjectSourceRef *string            `json:"subjectSourceRef,omitempty"`
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

func (this GetAttemptByIdQuery) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.Id, true),
	}

	return val.ApiBased.ValidateStruct(&this, rules...)
}

type GetAttemptByIdResult = crud.OpResult[*domain.LoginAttempt]
