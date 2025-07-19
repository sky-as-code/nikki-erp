package domain

import (
	"time"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	val "github.com/sky-as-code/nikki-erp/common/validator"
)

type LoginAttempt struct {
	model.ModelBase
	model.AuditableBase

	Methods          []string       `json:"methods"`
	CurrentMethod    *string        `json:"currentMethod"`
	DeviceIp         *string        `json:"deviceIP"`
	DeviceName       *string        `json:"deviceName"`
	DeviceLocation   *string        `json:"deviceLocation"`
	ExpiredAt        *time.Time     `json:"expiredAt"`
	IsGenuine        *bool          `json:"isGenuine"`
	SubjectType      *SubjectType   `json:"subjectType"`
	SubjectRef       *model.Id      `json:"subjectRef"`
	SubjectSourceRef *string        `json:"subjectSourceRef"`
	Status           *AttemptStatus `json:"status"`
	Username         *string        `json:"username"`
}

func (this *LoginAttempt) NextMethod() *string {
	if this.CurrentMethod == nil {
		return util.ToPtr(this.Methods[0])
	}
	total := len(this.Methods)
	for i := 0; i < total-1; i++ {
		if this.Methods[i] == *this.CurrentMethod {
			return util.ToPtr(this.Methods[i+1])
		}
	}
	return nil
}

func (this *LoginAttempt) SetDefaults() {
	this.ModelBase.SetDefaults()
	this.IsGenuine = util.ToPtr(false)
	this.Status = util.ToPtr(AttemptStatusPending)
}

func (this *LoginAttempt) Validate(forEdit bool) ft.ValidationErrors {
	rules := []*val.FieldRules{
		val.Field(&this.DeviceIp,
			val.IsIp,
		),
		val.Field(&this.DeviceName,
			val.Length(1, model.MODEL_RULE_LONG_NAME_LENGTH),
		),
		val.Field(&this.DeviceLocation,
			val.Length(1, model.MODEL_RULE_SHORT_NAME_LENGTH),
		),
		val.Field(&this.Username,
			val.NotNilWhen(this.SubjectType != nil && this.SubjectRef == nil),
			val.When(this.Username != nil,
				val.NotEmpty,
				val.Length(5, model.MODEL_RULE_USERNAME_LENGTH),
			),
		),

		model.IdPtrValidateRule(&this.Id, true),
		SubjectTypePtrValidateRule(&this.SubjectType, !forEdit),
		model.IdPtrValidateRule(&this.SubjectRef, this.SubjectType != nil && this.Username == nil),
		AttemptStatusValidateRule(&this.Status),
	}
	rules = append(rules, this.AuditableBase.ValidateRules(forEdit)...)

	return val.ApiBased.ValidateStruct(this, rules...)
}

type AttemptStatus string

const (
	AttemptStatusPending = AttemptStatus("pending")
	AttemptStatusSuccess = AttemptStatus("success")
	AttemptStatusFailed  = AttemptStatus("failed")
)

func (this AttemptStatus) String() string {
	return string(this)
}

func WrapAttemptStatus(s string) *AttemptStatus {
	st := AttemptStatus(s)
	return &st
}

func AttemptStatusValidateRule(field **AttemptStatus) *val.FieldRules {
	return val.Field(field,
		val.When(*field != nil,
			val.NotEmpty,
			val.OneOf(AttemptStatusPending, AttemptStatusSuccess, AttemptStatusFailed),
		),
	)
}
