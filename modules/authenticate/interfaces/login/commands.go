package login

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/authenticate/domain"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
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

func (this AuthenticateCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetOrRegisterSchema(
		"authenticate.authenticate_command",
		func() *dmodel.ModelSchemaBuilder {
			return dmodel.DefineModel("_").
				Field(basemodel.DefineFieldId("attempt_id").RequiredAlways()).
				Field(
					dmodel.DefineField().
						Name("passwords").
						DataType(dmodel.FieldDataTypeModel()).
						RequiredAlways(),
				)
		},
	)
}

type AuthenticateSuccessData struct {
	AccessToken           string              `json:"access_token"`
	AccessTokenExpiresAt  model.ModelDateTime `json:"access_token_expires_at"`
	RefreshToken          string              `json:"refresh_token"`
	RefreshTokenExpiresAt model.ModelDateTime `json:"refresh_token_expires_at"`
}

type AuthenticateResultData struct {
	Done     bool                     `json:"done"`
	NextStep *string                  `json:"next_step,omitempty"`
	Data     *AuthenticateSuccessData `json:"data,omitempty"`
}
type AuthenticateResult = dyn.OpResult[AuthenticateResultData]

var createLoginAttemptCommandType = cqrs.RequestType{
	Module:    "authenticate",
	Submodule: "login",
	Action:    "createAttempt",
}

func NewCreateLoginAttemptCommand() CreateLoginAttemptCommand {
	return CreateLoginAttemptCommand{
		LoginAttempt: *domain.NewLoginAttempt(),
	}
}

type CreateLoginAttemptCommand struct {
	domain.LoginAttempt
}

func (CreateLoginAttemptCommand) CqrsRequestType() cqrs.RequestType {
	return createLoginAttemptCommandType
}

func (this CreateLoginAttemptCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.LoginAttemptSchemaName)
}

type CreateLoginAttemptResultData struct {
	Attempt       domain.LoginAttempt `json:"attempt"`
	PrincipalName string              `json:"principal_name"`
}

type CreateLoginAttemptResult = dyn.OpResult[CreateLoginAttemptResultData]

var updateLoginAttemptCommandType = cqrs.RequestType{
	Module:    "authenticate",
	Submodule: "login",
	Action:    "updateAttempt",
}

type UpdateLoginAttemptCommand struct {
	domain.LoginAttempt
}

func (UpdateLoginAttemptCommand) CqrsRequestType() cqrs.RequestType {
	return updateLoginAttemptCommandType
}

func (this UpdateLoginAttemptCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.LoginAttemptSchemaName)
}

type UpdateLoginAttemptResult = dyn.OpResult[dyn.MutateResultData]

var startLoginFlowCommandType = cqrs.RequestType{
	Module:    "authenticate",
	Submodule: "login",
	Action:    "startLoginFlow",
}

type StartLoginFlowCommand struct {
	DeviceName    *string               `json:"device_name,omitempty"`
	PrincipalType *domain.PrincipalType `json:"principal_type,omitempty"`
	Username      string                `json:"username"`
}

func (StartLoginFlowCommand) CqrsRequestType() cqrs.RequestType {
	return startLoginFlowCommandType
}

func (this StartLoginFlowCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetOrRegisterSchema(
		"authenticate.start_login_flow_command",
		func() *dmodel.ModelSchemaBuilder {
			return dmodel.DefineModel("_").
				Field(domain.DefinePrincipalDeviceNameField()).
				Field(domain.DefinePrincipalTypeField("principal_type").Default(domain.PrincipalTypeNikkiUser)).
				Field(domain.DefinePrincipalUsernameField("username").RequiredAlways())
		},
	)
}

var getAttemptByIdQueryType = cqrs.RequestType{
	Module:    "authenticate",
	Submodule: "login",
	Action:    "getAttemptById",
}

type GetAttemptQuery dyn.GetOneQuery

func (this GetAttemptQuery) CqrsRequestType() cqrs.RequestType {
	return getAttemptByIdQueryType
}

type GetAttemptResult = dyn.OpResult[domain.LoginAttempt]

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

func (this RefreshTokenCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetOrRegisterSchema(
		"authenticate.refresh_token_command",
		func() *dmodel.ModelSchemaBuilder {
			return dmodel.DefineModel("_").
				Field(
					dmodel.DefineField().
						Name("refresh_token").
						DataType(dmodel.FieldDataTypeString(1, 1000)).
						RequiredAlways(),
				)
		},
	)
}

type RefreshTokenResultData = AuthenticateSuccessData
type RefreshTokenResult = dyn.OpResult[RefreshTokenResultData]
