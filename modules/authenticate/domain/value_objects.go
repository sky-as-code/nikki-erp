package domain

import (
	"regexp"

	val "github.com/sky-as-code/nikki-erp/common/validator"
	c "github.com/sky-as-code/nikki-erp/modules/authenticate/constants"
	"go.bryk.io/pkg/errors"
)

type SubjectType string

const (
	SubjectTypeUser   = SubjectType("user")
	SubjectTypeCustom = SubjectType("custom")
)

func (this SubjectType) String() string {
	return string(this)
}

func WrapSubjectType(s string) *SubjectType {
	st := SubjectType(s)
	return &st
}

func SubjectTypeValidateRule(field *SubjectType, isRequired bool) *val.FieldRules {
	return val.Field(field,
		val.NotNilWhen(isRequired),
		val.When(field != nil,
			val.NotEmpty,
			val.OneOf(SubjectTypeUser, SubjectTypeCustom),
		),
	)
}

func SubjectTypePtrValidateRule(field **SubjectType, isRequired bool) *val.FieldRules {
	return val.Field(field,
		val.NotNilWhen(isRequired),
		val.When(*field != nil,
			val.NotEmpty,
			val.OneOf(SubjectTypeUser, SubjectTypeCustom),
		),
	)
}

type SendChannel string

const (
	SendChannelEmail = SendChannel("email")
	SendChannelSms   = SendChannel("sms")
)

func (this SendChannel) String() string {
	return string(this)
}

func WrapSendChannel(s string) *SendChannel {
	sc := SendChannel(s)
	return &sc
}

func SendChannelValidateRule(field *SendChannel) *val.FieldRules {
	return val.Field(field,
		val.NotEmpty,
		val.OneOf(SendChannelEmail, SendChannelSms),
	)
}

type OtpCode string

func (this OtpCode) String() string {
	return string(this)
}

var otpCodeRegexp = regexp.MustCompile(`^\d+$`)

func OtpCodeValidateRule(field *OtpCode, isRequired bool) *val.FieldRules {
	return val.Field(field,
		val.NotNilWhen(isRequired),
		val.When(field != nil,
			val.NotEmpty,
			val.By(func(value any) error {
				otpCode := *(value.(*OtpCode))
				if otpCodeRegexp.MatchString(otpCode.String()) && len(otpCode.String()) == c.OTP_CODE_LENGTH {
					return nil
				}
				return errors.New("invalid otp code")
			}),
		),
	)
}
