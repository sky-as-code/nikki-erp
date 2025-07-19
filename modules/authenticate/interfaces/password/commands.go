package password

import (
	"time"

	"github.com/sky-as-code/nikki-erp/common/crud"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/authenticate/domain"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
)

var createPasswordOtpCommandType = cqrs.RequestType{
	Module:    "authenticate",
	Submodule: "password",
	Action:    "createPasswordOtp",
}

type CreateOtpPasswordCommand struct {
	SubjectType domain.SubjectType `json:"subjectType"`
	SubjectRef  model.Id           `json:"subjectRef"`
}

func (CreateOtpPasswordCommand) CqrsRequestType() cqrs.RequestType {
	return createPasswordOtpCommandType
}

func (this CreateOtpPasswordCommand) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		domain.SubjectTypeValidateRule(&this.SubjectType, true),
		model.IdValidateRule(&this.SubjectRef, true),
	}

	return val.ApiBased.ValidateStruct(&this, rules...)
}

type CreatePasswordOtpResultData struct {
	CreatedAt time.Time `json:"createdAt"`
	ExpiredAt time.Time `json:"expiredAt"`
	OtpUrl    string    `json:"otpUrl"`
}
type CreateOtpPasswordResult = crud.OpResult[*CreatePasswordOtpResultData]

var confirmOtpPasswordCommandType = cqrs.RequestType{
	Module:    "authenticate",
	Submodule: "password",
	Action:    "confirmOtpPassword",
}

type ConfirmOtpPasswordCommand struct {
	SubjectType domain.SubjectType `json:"subjectType"`
	SubjectRef  model.Id           `json:"subjectRef"`
	OtpCode     domain.OtpCode     `json:"otpCode"`
}

func (ConfirmOtpPasswordCommand) CqrsRequestType() cqrs.RequestType {
	return confirmOtpPasswordCommandType
}

func (this ConfirmOtpPasswordCommand) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.SubjectRef, true),
		domain.SubjectTypeValidateRule(&this.SubjectType, true),
		domain.OtpCodeValidateRule(&this.OtpCode, true),
	}

	return val.ApiBased.ValidateStruct(&this, rules...)
}

type ConfirmOtpPasswordResultData struct {
	ConfirmedAt   time.Time `json:"confirmedAt"`
	RecoveryCodes []string  `json:"recoveryCodes"`
}
type ConfirmOtpPasswordResult = crud.OpResult[*ConfirmOtpPasswordResultData]

var createTempPasswordCommandType = cqrs.RequestType{
	Module:    "authenticate",
	Submodule: "password",
	Action:    "createTempPassword",
}

type CreateTempPasswordCommand struct {
	SubjectType domain.SubjectType `json:"subjectType"`
	SendChannel domain.SendChannel `json:"sendChannel"`
	Username    string             `json:"username"`
}

func (CreateTempPasswordCommand) CqrsRequestType() cqrs.RequestType {
	return createTempPasswordCommandType
}

func (this CreateTempPasswordCommand) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		val.Field(&this.Username,
			val.NotEmpty,
			val.Length(5, model.MODEL_RULE_USERNAME_LENGTH),
		),
		domain.SendChannelValidateRule(&this.SendChannel),
		domain.SubjectTypeValidateRule(&this.SubjectType, true),
	}

	return val.ApiBased.ValidateStruct(&this, rules...)
}

type CreateTempPasswordResultData struct {
	CreatedAt time.Time `json:"createdAt"`
	ExpiredAt time.Time `json:"expiredAt"`
}
type CreateTempPasswordResult = crud.OpResult[*CreateTempPasswordResultData]

var setPasswordCommandType = cqrs.RequestType{
	Module:    "authenticate",
	Submodule: "password",
	Action:    "setPassword",
}

type SetPasswordCommand struct {
	SubjectType     domain.SubjectType `json:"subjectType"`
	SubjectRef      model.Id           `json:"subjectRef"`
	CurrentPassword *string            `json:"currentPassword"`
	NewPassword     string             `json:"newPassword"`
}

func (SetPasswordCommand) CqrsRequestType() cqrs.RequestType {
	return setPasswordCommandType
}

func (this SetPasswordCommand) Validate() ft.ValidationErrors {

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

	return val.ApiBased.ValidateStruct(&this, rules...)
}

type SetPasswordResultData struct {
	UpdatedAt time.Time `json:"updatedAt"`
}
type SetPasswordResult = crud.OpResult[*SetPasswordResultData]

var verifyPasswordQueryType = cqrs.RequestType{
	Module:    "authenticate",
	Submodule: "password",
	Action:    "verifyPassword",
}

type VerifyPasswordQuery struct {
	SubjectType domain.SubjectType `json:"subjectType"`
	Username    string             `json:"username"`
	Password    string             `json:"password"`
}

func (VerifyPasswordQuery) CqrsRequestType() cqrs.RequestType {
	return verifyPasswordQueryType
}

func (this VerifyPasswordQuery) Validate() ft.ValidationErrors {
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

	return val.ApiBased.ValidateStruct(&this, rules...)
}

type VerifyPasswordResultData struct {
	IsVerified   bool   `json:"isVerified"`
	FailedReason string `json:"failedReason,omitempty"`
}
type VerifyPasswordResult = crud.OpResult[*VerifyPasswordResultData]

var verifyOtpCodeQueryType = cqrs.RequestType{
	Module:    "authenticate",
	Submodule: "password",
	Action:    "isPasswordMatched",
}

type VerifyOtpCodeQuery struct {
	SubjectType domain.SubjectType `json:"subjectType"`
	Username    string             `json:"username"`
	OtpCode     domain.OtpCode     `json:"otpCode"`
}

func (VerifyOtpCodeQuery) CqrsRequestType() cqrs.RequestType {
	return verifyOtpCodeQueryType
}

func (this VerifyOtpCodeQuery) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		val.Field(&this.Username,
			val.NotEmpty,
			val.Length(5, model.MODEL_RULE_USERNAME_LENGTH),
		),
		domain.SubjectTypeValidateRule(&this.SubjectType, true),
		domain.OtpCodeValidateRule(&this.OtpCode, true),
	}

	return val.ApiBased.ValidateStruct(&this, rules...)
}

type VerifyOtpCodeResult = VerifyPasswordResult
