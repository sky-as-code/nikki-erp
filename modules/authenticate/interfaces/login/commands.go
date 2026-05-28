package login

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/authenticate/domain/models"
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
		LoginAttempt: *models.NewLoginAttempt(),
	}
}

type CreateLoginAttemptCommand struct {
	models.LoginAttempt
}

func (CreateLoginAttemptCommand) CqrsRequestType() cqrs.RequestType {
	return createLoginAttemptCommandType
}

func (this CreateLoginAttemptCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(models.LoginAttemptSchemaName)
}

type CreateLoginAttemptResultData struct {
	Attempt       models.LoginAttempt `json:"attempt"`
	PrincipalName string              `json:"principal_name"`
}

type CreateLoginAttemptResult = dyn.OpResult[CreateLoginAttemptResultData]

var updateLoginAttemptCommandType = cqrs.RequestType{
	Module:    "authenticate",
	Submodule: "login",
	Action:    "updateAttempt",
}

type UpdateLoginAttemptCommand struct {
	models.LoginAttempt
}

func (UpdateLoginAttemptCommand) CqrsRequestType() cqrs.RequestType {
	return updateLoginAttemptCommandType
}

func (this UpdateLoginAttemptCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(models.LoginAttemptSchemaName)
}

type UpdateLoginAttemptResult = dyn.OpResult[dyn.MutateResultData]

var startLoginFlowCommandType = cqrs.RequestType{
	Module:    "authenticate",
	Submodule: "login",
	Action:    "startLoginFlow",
}

type StartLoginFlowCommand struct {
	DeviceName    *string               `json:"device_name,omitempty"`
	PrincipalType *models.PrincipalType `json:"principal_type,omitempty"`
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
				Field(models.DefinePrincipalDeviceNameField()).
				Field(models.DefinePrincipalTypeField("principal_type").Default(models.PrincipalTypeNikkiUser)).
				Field(models.DefinePrincipalUsernameField("username").RequiredAlways())
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

type GetAttemptResult = dyn.OpResult[models.LoginAttempt]

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
