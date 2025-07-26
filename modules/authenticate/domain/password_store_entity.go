package domain

import (
	"time"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	c "github.com/sky-as-code/nikki-erp/modules/authenticate/constants"
)

type PasswordStore struct {
	model.ModelBase

	Password             *string      `json:"password"`
	PasswordExpiredAt    *time.Time   `json:"passwordExpiredAt"`
	PasswordUpdatedAt    *time.Time   `json:"passwordUpdatedAt"`
	Passwordtmp          *string      `json:"passwordTmp"`
	PasswordtmpExpiredAt *time.Time   `json:"passwordTmpExpiredAt"`
	Passwordotp          *string      `json:"passwordOtp"`
	PasswordotpExpiredAt *time.Time   `json:"passwordOtpExpiredAt"`
	PasswordotpRecovery  []string     `json:"passwordOtpRecovery"`
	SubjectType          *SubjectType `json:"subjectType"`
	SubjectRef           *model.Id    `json:"subjectRef"`
	SubjectSourceRef     *string      `json:"subjectSourceRef"`
}

func (this *PasswordStore) SetDefaults() {
	this.ModelBase.SetDefaults()
}

func (this *PasswordStore) Validate(forEdit bool) ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdPtrValidateRule(&this.Id, !forEdit),
		model.IdPtrValidateRule(&this.SubjectRef, true),
		SubjectTypePtrValidateRule(&this.SubjectType, true),
	}
	rules = append(rules, this.ModelBase.ValidateRules(forEdit)...)

	return val.ApiBased.ValidateStruct(this, rules...)
}

type OtpCode string

func (this OtpCode) String() string {
	return string(this)
}

func WrapOtpCode(s string) *OtpCode {
	st := OtpCode(s)
	return &st
}

func OtpCodeValidateRule(field *OtpCode, isRequired bool) *val.FieldRules {
	return val.Field(field,
		val.NotNilWhen(isRequired),
		val.When(field != nil,
			val.NotEmpty,
			val.Length(c.OTP_CODE_LENGTH, c.OTP_CODE_LENGTH),
		),
	)
}
