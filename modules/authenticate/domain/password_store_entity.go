package domain

import (
	"time"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	val "github.com/sky-as-code/nikki-erp/common/validator"
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
		SubjectTypeValidateRule(&this.SubjectType),
	}
	rules = append(rules, this.ModelBase.ValidateRules(forEdit)...)

	return val.ApiBased.ValidateStruct(this, rules...)
}
