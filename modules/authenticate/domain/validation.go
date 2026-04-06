package domain

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	val "github.com/sky-as-code/nikki-erp/common/validator"
)

type loginAttemptValid struct {
	DeviceIp       *string
	DeviceName     *string
	DeviceLocation *string
	Username       *string
	Id             *model.Id
	SubjectType    *SubjectType
	SubjectRef     *model.Id
	Status         *AttemptStatus
}

func ValidateLoginAttempt(this *LoginAttempt, forEdit bool) ft.ClientErrors {
	v := loginAttemptValid{
		DeviceIp:       this.GetDeviceIp(),
		DeviceName:     this.GetDeviceName(),
		DeviceLocation: this.GetDeviceLocation(),
		Username:       this.GetUsername(),
		Id:             this.GetId(),
		SubjectType:    this.GetSubjectType(),
		SubjectRef:     this.GetSubjectRef(),
		Status:         this.GetStatus(),
	}
	rules := []*val.FieldRules{
		val.Field(&v.DeviceIp, val.IsIp),
		val.Field(&v.DeviceName, val.Length(1, model.MODEL_RULE_LONG_NAME_LENGTH)),
		val.Field(&v.DeviceLocation, val.Length(1, model.MODEL_RULE_SHORT_NAME_LENGTH)),
		val.Field(&v.Username,
			val.NotNilWhen(v.SubjectType != nil && v.SubjectRef == nil),
			val.When(v.Username != nil,
				val.NotEmpty,
				val.Length(5, model.MODEL_RULE_USERNAME_LENGTH),
			),
		),
		model.IdPtrValidateRule(&v.Id, true),
		SubjectTypePtrValidateRule(&v.SubjectType, !forEdit),
		model.IdPtrValidateRule(&v.SubjectRef, v.SubjectType != nil && v.Username == nil),
		AttemptStatusValidateRule(&v.Status),
	}
	return ValidationErrorsToClientErrors(val.ApiBased.ValidateStruct(&v, rules...))
}

type passwordStoreValid struct {
	Id          *model.Id
	SubjectRef  *model.Id
	SubjectType *SubjectType
}

func ValidatePasswordStore(this *PasswordStore, forEdit bool) ft.ClientErrors {
	v := passwordStoreValid{
		Id:          this.GetId(),
		SubjectRef:  this.GetSubjectRef(),
		SubjectType: this.GetSubjectType(),
	}
	rules := []*val.FieldRules{
		model.IdPtrValidateRule(&v.Id, !forEdit),
		model.IdPtrValidateRule(&v.SubjectRef, true),
		SubjectTypePtrValidateRule(&v.SubjectType, true),
	}
	return ValidationErrorsToClientErrors(val.ApiBased.ValidateStruct(&v, rules...))
}
