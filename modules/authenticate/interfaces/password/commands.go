package password

import (
	"time"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/authenticate/domain"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

var createPasswordOtpCommandType = cqrs.RequestType{
	Module:    "authenticate",
	Submodule: "password",
	Action:    "createPasswordOtp",
}

type CreateOtpPasswordCommand struct {
	SubjectType domain.SubjectType `json:"subject_type"`
	SubjectRef  model.Id           `json:"subject_ref"`
}

func (CreateOtpPasswordCommand) CqrsRequestType() cqrs.RequestType {
	return createPasswordOtpCommandType
}

func (this CreateOtpPasswordCommand) Validate() ft.ClientErrors {
	rules := []*val.FieldRules{
		domain.SubjectTypeValidateRule(&this.SubjectType, true),
		model.IdValidateRule(&this.SubjectRef, true),
	}

	return domain.ValidationErrorsToClientErrors(val.ApiBased.ValidateStruct(&this, rules...))
}

type CreatePasswordOtpResultData struct {
	CreatedAt time.Time `json:"created_at"`
	ExpiredAt time.Time `json:"expired_at"`
	OtpUrl    string    `json:"otp_url"`
}
type CreateOtpPasswordResult = dyn.OpResult[*CreatePasswordOtpResultData]

var confirmOtpPasswordCommandType = cqrs.RequestType{
	Module:    "authenticate",
	Submodule: "password",
	Action:    "confirmOtpPassword",
}

type ConfirmOtpPasswordCommand struct {
	SubjectType domain.SubjectType `json:"subject_type"`
	SubjectRef  model.Id           `json:"subject_ref"`
	OtpCode     domain.OtpCode     `json:"otp_code"`
}

func (ConfirmOtpPasswordCommand) CqrsRequestType() cqrs.RequestType {
	return confirmOtpPasswordCommandType
}

func (this ConfirmOtpPasswordCommand) Validate() ft.ClientErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.SubjectRef, true),
		domain.SubjectTypeValidateRule(&this.SubjectType, true),
		domain.OtpCodeValidateRule(&this.OtpCode, true),
	}

	return domain.ValidationErrorsToClientErrors(val.ApiBased.ValidateStruct(&this, rules...))
}

type ConfirmOtpPasswordResultData struct {
	ConfirmedAt   time.Time `json:"confirmed_at"`
	RecoveryCodes []string  `json:"recovery_codes"`
}
type ConfirmOtpPasswordResult = dyn.OpResult[*ConfirmOtpPasswordResultData]

var createTempPasswordCommandType = cqrs.RequestType{
	Module:    "authenticate",
	Submodule: "password",
	Action:    "createTempPassword",
}

type CreateTempPasswordCommand struct {
	SubjectType domain.SubjectType `json:"subject_type"`
	SendChannel domain.SendChannel `json:"send_channel"`
	Username    string             `json:"username"`
}

func (CreateTempPasswordCommand) CqrsRequestType() cqrs.RequestType {
	return createTempPasswordCommandType
}

func (this CreateTempPasswordCommand) Validate() ft.ClientErrors {
	rules := []*val.FieldRules{
		val.Field(&this.Username,
			val.NotEmpty,
			val.Length(5, model.MODEL_RULE_USERNAME_LENGTH),
		),
		domain.SendChannelValidateRule(&this.SendChannel),
		domain.SubjectTypeValidateRule(&this.SubjectType, true),
	}

	return domain.ValidationErrorsToClientErrors(val.ApiBased.ValidateStruct(&this, rules...))
}

type CreateTempPasswordResultData struct {
	CreatedAt time.Time `json:"created_at"`
	ExpiredAt time.Time `json:"expired_at"`
}
type CreateTempPasswordResult = dyn.OpResult[*CreateTempPasswordResultData]

var setPasswordCommandType = cqrs.RequestType{
	Module:    "authenticate",
	Submodule: "password",
	Action:    "setPassword",
}

type SetPasswordCommand struct {
	SubjectType     domain.SubjectType `json:"subject_type"`
	SubjectRef      model.Id           `json:"subject_ref"`
	CurrentPassword *string            `json:"current_password"`
	NewPassword     string             `json:"new_password"`
}

func (SetPasswordCommand) CqrsRequestType() cqrs.RequestType {
	return setPasswordCommandType
}

func (this SetPasswordCommand) Validate() ft.ClientErrors {

	rules := []*val.FieldRules{
		domain.SubjectTypeValidateRule(&this.SubjectType, true),
		model.IdValidateRule(&this.SubjectRef, true),
		val.Field(&this.CurrentPassword,
			val.Length(model.MODEL_RULE_PASSWORD_MIN_LENGTH, model.MODEL_RULE_PASSWORD_MAX_LENGTH),
		),
		val.Field(&this.NewPassword,
			val.NotEmpty,
			val.Length(model.MODEL_RULE_PASSWORD_MIN_LENGTH, model.MODEL_RULE_PASSWORD_MAX_LENGTH),
		),
	}

	return domain.ValidationErrorsToClientErrors(val.ApiBased.ValidateStruct(&this, rules...))
}

type SetPasswordResultData struct {
	UpdatedAt time.Time `json:"updated_at"`
}
type SetPasswordResult = dyn.OpResult[*SetPasswordResultData]

var verifyPasswordQueryType = cqrs.RequestType{
	Module:    "authenticate",
	Submodule: "password",
	Action:    "verifyPassword",
}

type VerifyPasswordQuery struct {
	SubjectType domain.SubjectType `json:"subject_type"`
	Username    string             `json:"username"`
	Password    string             `json:"password"`
}

func (VerifyPasswordQuery) CqrsRequestType() cqrs.RequestType {
	return verifyPasswordQueryType
}

func (this VerifyPasswordQuery) Validate() ft.ClientErrors {
	rules := []*val.FieldRules{
		val.Field(&this.Username,
			val.NotEmpty,
			val.Length(5, model.MODEL_RULE_USERNAME_LENGTH),
		),
		val.Field(&this.Password,
			val.NotEmpty,
			val.Length(model.MODEL_RULE_PASSWORD_MIN_LENGTH, model.MODEL_RULE_PASSWORD_MAX_LENGTH),
		),
		domain.SubjectTypeValidateRule(&this.SubjectType, true),
	}

	return domain.ValidationErrorsToClientErrors(val.ApiBased.ValidateStruct(&this, rules...))
}

type VerifyPasswordResultData struct {
	IsVerified   bool   `json:"is_verified"`
	FailedReason string `json:"failed_reason,omitempty"`
}
type VerifyPasswordResult = dyn.OpResult[*VerifyPasswordResultData]

var verifyOtpCodeQueryType = cqrs.RequestType{
	Module:    "authenticate",
	Submodule: "password",
	Action:    "isPasswordMatched",
}

type VerifyOtpCodeQuery struct {
	SubjectType domain.SubjectType `json:"subject_type"`
	Username    string             `json:"username"`
	OtpCode     domain.OtpCode     `json:"otp_code"`
}

func (VerifyOtpCodeQuery) CqrsRequestType() cqrs.RequestType {
	return verifyOtpCodeQueryType
}

func (this VerifyOtpCodeQuery) Validate() ft.ClientErrors {
	rules := []*val.FieldRules{
		val.Field(&this.Username,
			val.NotEmpty,
			val.Length(5, model.MODEL_RULE_USERNAME_LENGTH),
		),
		domain.SubjectTypeValidateRule(&this.SubjectType, true),
		domain.OtpCodeValidateRule(&this.OtpCode, true),
	}

	return domain.ValidationErrorsToClientErrors(val.ApiBased.ValidateStruct(&this, rules...))
}

type VerifyOtpCodeResult = VerifyPasswordResult
