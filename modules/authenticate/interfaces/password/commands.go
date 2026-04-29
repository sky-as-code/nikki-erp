package password

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/authenticate/domain"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

var createPasswordOtpCommandType = cqrs.RequestType{
	Module:    "authenticate",
	Submodule: "password",
	Action:    "createPasswordOtp",
}

type CreatePasswordOtpCommand struct {
	PrincipalType domain.PrincipalType `json:"principal_type"`
	PrincipalId   model.Id             `json:"principal_id"`
}

func (CreatePasswordOtpCommand) CqrsRequestType() cqrs.RequestType {
	return createPasswordOtpCommandType
}

func (CreatePasswordOtpCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetOrRegisterSchema(
		"authenticate.create_password_otp_command",
		func() *dmodel.ModelSchemaBuilder {
			return dmodel.DefineModel("_").
				Field(domain.DefinePrincipalTypeField("principal_type").RequiredAlways()).
				Field(basemodel.DefineFieldId("principal_id").RequiredAlways())
		},
	)
}

type CreatePasswordOtpResultData struct {
	CreatedAt model.ModelDateTime `json:"created_at"`
	ExpiredAt model.ModelDateTime `json:"expired_at"`
	OtpUrl    string              `json:"otp_url"`
}
type CreatePasswordOtpResult = dyn.OpResult[CreatePasswordOtpResultData]

var confirmPasswordOtpCommandType = cqrs.RequestType{
	Module:    "authenticate",
	Submodule: "password",
	Action:    "confirmPasswordOtp",
}

type ConfirmPasswordOtpCommand struct {
	PrincipalType domain.PrincipalType `json:"principal_type"`
	PrincipalId   model.Id             `json:"principal_id"`
	OtpCode       domain.OtpCode       `json:"otp_code"`
}

func (ConfirmPasswordOtpCommand) CqrsRequestType() cqrs.RequestType {
	return confirmPasswordOtpCommandType
}

func (ConfirmPasswordOtpCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetOrRegisterSchema(
		"authenticate.confirm_password_otp_command",
		func() *dmodel.ModelSchemaBuilder {
			return dmodel.DefineModel("_").
				Field(domain.DefinePrincipalTypeField("principal_type").RequiredAlways()).
				Field(basemodel.DefineFieldId("principal_id").RequiredAlways()).
				Field(domain.DefinePasswordOtpField("otp_code").RequiredAlways())
		},
	)
}

type ConfirmPasswordOtpResultData struct {
	ConfirmedAt   model.ModelDateTime `json:"confirmed_at"`
	RecoveryCodes []string            `json:"recovery_codes"`
}
type ConfirmPasswordOtpResult = dyn.OpResult[ConfirmPasswordOtpResultData]

var createPasswordTempCommandType = cqrs.RequestType{
	Module:    "authenticate",
	Submodule: "password",
	Action:    "createPasswordTemp",
}

type CreatePasswordTempCommand struct {
	PrincipalType domain.PrincipalType `json:"subject_type"`
	SendChannel   domain.SendChannel   `json:"send_channel"`
	Username      string               `json:"username"`
}

func (CreatePasswordTempCommand) CqrsRequestType() cqrs.RequestType {
	return createPasswordTempCommandType
}

func (CreatePasswordTempCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetOrRegisterSchema(
		"authenticate.create_password_temp_command",
		func() *dmodel.ModelSchemaBuilder {
			return dmodel.DefineModel("_").
				Field(domain.DefinePrincipalTypeField("principal_type").RequiredAlways()).
				Field(domain.DefinePasswordSendChannelField("principal_id").RequiredAlways()).
				Field(domain.DefinePasswordOtpField("otp_code").RequiredAlways())
		},
	)
}

type CreatePasswordTempResultData struct {
	CreatedAt model.ModelDateTime `json:"created_at"`
	ExpiresAt model.ModelDateTime `json:"expires_at"`
}
type CreatePasswordTempResult = dyn.OpResult[CreatePasswordTempResultData]

var setPasswordCommandType = cqrs.RequestType{
	Module:    "authenticate",
	Submodule: "password",
	Action:    "setPassword",
}

type SetPasswordCommand struct {
	PrincipalType   domain.PrincipalType `json:"principal_type"`
	PrincipalId     model.Id             `json:"principal_id"`
	CurrentPassword *string              `json:"current_password"`
	NewPassword     string               `json:"new_password"`
}

func (SetPasswordCommand) CqrsRequestType() cqrs.RequestType {
	return setPasswordCommandType
}

func (SetPasswordCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetOrRegisterSchema(
		"authenticate.set_password_command",
		func() *dmodel.ModelSchemaBuilder {
			return dmodel.DefineModel("_").
				Field(domain.DefinePrincipalTypeField("principal_type").RequiredAlways()).
				Field(basemodel.DefineFieldId("principal_id").RequiredAlways()).
				Field(domain.DefinePasswordTextField("current_password")).
				Field(domain.DefinePasswordTextField("new_password").RequiredAlways())
		},
	)
}

type SetPasswordResult = dyn.OpResult[dyn.MutateResultData]

var verifyPasswordQueryType = cqrs.RequestType{
	Module:    "authenticate",
	Submodule: "password",
	Action:    "verifyPassword",
}

type VerifyPasswordQuery struct {
	PrincipalType domain.PrincipalType `json:"principal_type"`
	Username      string               `json:"username"`
	Password      string               `json:"password"`
}

func (VerifyPasswordQuery) CqrsRequestType() cqrs.RequestType {
	return verifyPasswordQueryType
}

func (VerifyPasswordQuery) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetOrRegisterSchema(
		"authenticate.verify_password_query",
		func() *dmodel.ModelSchemaBuilder {
			return dmodel.DefineModel("_").
				Field(domain.DefinePrincipalTypeField("principal_type").RequiredAlways()).
				Field(domain.DefinePrincipalUsernameField("username").RequiredAlways()).
				Field(domain.DefinePasswordTextField("password").RequiredAlways())
		},
	)
}

type VerifyPasswordResultData struct {
	IsVerified   bool                `json:"is_verified"`
	FailedReason *ft.ClientErrorItem `json:"failed_reason,omitempty"`
}
type VerifyPasswordResult = dyn.OpResult[VerifyPasswordResultData]

var verifyPasswordOtpQueryType = cqrs.RequestType{
	Module:    "authenticate",
	Submodule: "password",
	Action:    "verifyPasswordOtp",
}

type VerifyPasswordOtpQuery struct {
	PrincipalType domain.PrincipalType `json:"principal_type"`
	Username      string               `json:"username"`
	OtpCode       domain.OtpCode       `json:"otp_code"`
}

func (VerifyPasswordOtpQuery) CqrsRequestType() cqrs.RequestType {
	return verifyPasswordOtpQueryType
}

func (VerifyPasswordOtpQuery) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetOrRegisterSchema(
		"authenticate.verify_password_otp_query",
		func() *dmodel.ModelSchemaBuilder {
			return dmodel.DefineModel("_").
				Field(domain.DefinePrincipalTypeField("principal_type").RequiredAlways()).
				Field(domain.DefinePrincipalUsernameField("username").RequiredAlways()).
				Field(domain.DefinePasswordOtpField("otp_code").RequiredAlways())
		},
	)
}

type VerifyOtpCodeResult = VerifyPasswordResult
