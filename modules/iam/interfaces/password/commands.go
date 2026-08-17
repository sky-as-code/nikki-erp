package password

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
)

var createPasswordOtpCommandType = cqrs.RequestType{
	Module:    "iam",
	Submodule: "password",
	Action:    "createPasswordOtp",
}

type CreatePasswordOtpCommand struct {
	PrincipalType models.PrincipalType `json:"principal_type"`
	PrincipalId   model.Id             `json:"principal_id"`
}

func (CreatePasswordOtpCommand) CqrsRequestType() cqrs.RequestType {
	return createPasswordOtpCommandType
}

func (CreatePasswordOtpCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetOrRegisterSchema(
		"iam.create_password_otp_command",
		func() *dmodel.ModelSchemaBuilder {
			return dmodel.DefineModel("_").
				Field(models.DefinePrincipalTypeField("principal_type").RequiredAlways()).
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
	Module:    "iam",
	Submodule: "password",
	Action:    "confirmPasswordOtp",
}

type ConfirmPasswordOtpCommand struct {
	PrincipalType models.PrincipalType `json:"principal_type"`
	PrincipalId   model.Id             `json:"principal_id"`
	OtpCode       models.OtpCode       `json:"otp_code"`
}

func (ConfirmPasswordOtpCommand) CqrsRequestType() cqrs.RequestType {
	return confirmPasswordOtpCommandType
}

func (ConfirmPasswordOtpCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetOrRegisterSchema(
		"iam.confirm_password_otp_command",
		func() *dmodel.ModelSchemaBuilder {
			return dmodel.DefineModel("_").
				Field(models.DefinePrincipalTypeField("principal_type").RequiredAlways()).
				Field(basemodel.DefineFieldId("principal_id").RequiredAlways()).
				Field(models.DefinePasswordOtpField("otp_code").RequiredAlways())
		},
	)
}

type ConfirmPasswordOtpResultData struct {
	ConfirmedAt   model.ModelDateTime `json:"confirmed_at"`
	RecoveryCodes []string            `json:"recovery_codes"`
}
type ConfirmPasswordOtpResult = dyn.OpResult[ConfirmPasswordOtpResultData]

var createPasswordTempCommandType = cqrs.RequestType{
	Module:    "iam",
	Submodule: "password",
	Action:    "createPasswordTemp",
}

type CreatePasswordTempCommand struct {
	PrincipalType models.PrincipalType `json:"subject_type"`
	SendChannel   models.SendChannel   `json:"send_channel"`
	Username      string               `json:"username"`
}

func (CreatePasswordTempCommand) CqrsRequestType() cqrs.RequestType {
	return createPasswordTempCommandType
}

func (CreatePasswordTempCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetOrRegisterSchema(
		"iam.create_password_temp_command",
		func() *dmodel.ModelSchemaBuilder {
			return dmodel.DefineModel("_").
				Field(models.DefinePrincipalTypeField("principal_type").RequiredAlways()).
				Field(models.DefinePasswordSendChannelField("principal_id").RequiredAlways()).
				Field(models.DefinePasswordOtpField("otp_code").RequiredAlways())
		},
	)
}

type CreatePasswordTempResultData struct {
	CreatedAt model.ModelDateTime `json:"created_at"`
	ExpiresAt model.ModelDateTime `json:"expires_at"`
}
type CreatePasswordTempResult = dyn.OpResult[CreatePasswordTempResultData]

var setPasswordCommandType = cqrs.RequestType{
	Module:    "iam",
	Submodule: "password",
	Action:    "setPassword",
}

type SetPasswordCommand struct {
	PrincipalType   models.PrincipalType `json:"principal_type"`
	PrincipalId     model.Id             `json:"principal_id"`
	CurrentPassword *string              `json:"current_password"`
	NewPassword     string               `json:"new_password"`
}

func (SetPasswordCommand) CqrsRequestType() cqrs.RequestType {
	return setPasswordCommandType
}

func (SetPasswordCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetOrRegisterSchema(
		"iam.set_password_command",
		func() *dmodel.ModelSchemaBuilder {
			return dmodel.DefineModel("_").
				Field(models.DefinePrincipalTypeField("principal_type").RequiredAlways()).
				Field(basemodel.DefineFieldId("principal_id").RequiredAlways()).
				Field(models.DefinePasswordTextField("current_password")).
				Field(models.DefinePasswordTextField("new_password").RequiredAlways())
		},
	)
}

type SetPasswordResult = dyn.OpResult[dyn.MutateResultData]

var verifyPasswordQueryType = cqrs.RequestType{
	Module:    "iam",
	Submodule: "password",
	Action:    "verifyPassword",
}

type VerifyPasswordQuery struct {
	PrincipalType models.PrincipalType `json:"principal_type"`
	Username      string               `json:"username"`
	Password      string               `json:"password"`
}

func (VerifyPasswordQuery) CqrsRequestType() cqrs.RequestType {
	return verifyPasswordQueryType
}

func (VerifyPasswordQuery) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetOrRegisterSchema(
		"iam.verify_password_query",
		func() *dmodel.ModelSchemaBuilder {
			return dmodel.DefineModel("_").
				Field(models.DefinePrincipalTypeField("principal_type").RequiredAlways()).
				Field(models.DefinePrincipalUsernameField("username").RequiredAlways()).
				Field(models.DefinePasswordTextField("password").RequiredAlways())
		},
	)
}

type VerifyPasswordResultData struct {
	IsVerified   bool                `json:"is_verified"`
	FailedReason *ft.ClientErrorItem `json:"failed_reason,omitempty"`
}
type VerifyPasswordResult = dyn.OpResult[VerifyPasswordResultData]

var verifyPasswordOtpQueryType = cqrs.RequestType{
	Module:    "iam",
	Submodule: "password",
	Action:    "verifyPasswordOtp",
}

type VerifyPasswordOtpQuery struct {
	PrincipalType models.PrincipalType `json:"principal_type"`
	Username      string               `json:"username"`
	OtpCode       models.OtpCode       `json:"otp_code"`
}

func (VerifyPasswordOtpQuery) CqrsRequestType() cqrs.RequestType {
	return verifyPasswordOtpQueryType
}

func (VerifyPasswordOtpQuery) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetOrRegisterSchema(
		"iam.verify_password_otp_query",
		func() *dmodel.ModelSchemaBuilder {
			return dmodel.DefineModel("_").
				Field(models.DefinePrincipalTypeField("principal_type").RequiredAlways()).
				Field(models.DefinePrincipalUsernameField("username").RequiredAlways()).
				Field(models.DefinePasswordOtpField("otp_code").RequiredAlways())
		},
	)
}

type VerifyOtpCodeResult = VerifyPasswordResult
