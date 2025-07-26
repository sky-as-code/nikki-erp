package validator

import (
	"regexp"

	"go.bryk.io/pkg/errors"

	tagVal "github.com/go-playground/validator/v10"
	apiVal "github.com/invopop/validation"
	"github.com/invopop/validation/is"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
)

var TagBased = TagBasedValidator{
	validator: tagVal.New(tagVal.WithRequiredStructEnabled()),
}

type TagBasedValidator struct {
	validator *tagVal.Validate
}

func (v TagBasedValidator) Validate(i any) error {
	// TODO: Must convert to more friendly error struct
	return v.validator.Struct(i)
}

var (
	ApiBased       = newApiBasedValidator()
	NotEmpty       = apiVal.Required
	NotNil         = apiVal.NotNil
	IsAlpha        = is.Alpha
	IsAlphaNumeric = is.Alphanumeric
	IsDomain       = is.Domain
	IsEmail        = is.EmailFormat
	IsIp           = is.IP
	IsNumeric      = is.Digit
	IsUuid         = is.UUID
	IsUrl          = is.URL
)

func newApiBasedValidator() ApiBasedValidator {
	return ApiBasedValidator{}
}

type EachRule = apiVal.EachRule
type FieldRules = apiVal.FieldRules
type KeyRules = apiVal.KeyRules
type MapRule = apiVal.MapRule
type LengthRule = apiVal.LengthRule
type MatchRule = apiVal.MatchRule
type OneOfRule = apiVal.InRule[any]
type Rule = apiVal.Rule
type ThresholdRule = apiVal.ThresholdRule
type WhenRule = apiVal.WhenRule

type ApiBasedValidator struct {
}

func (v ApiBasedValidator) ValidateRaw(value interface{}, rules ...apiVal.Rule) error {
	return apiVal.Validate(value, rules...)
}

func (v ApiBasedValidator) ValidateStruct(structPtr any, fields ...*apiVal.FieldRules) ft.ValidationErrors {
	err := apiVal.ValidateStruct(structPtr, fields...)
	if err != nil {
		invopopErr, isOk := err.(apiVal.Errors)
		if isOk {
			return ft.NewValidationErrorsFromInvopop(invopopErr)
		} else {
			panic(errors.Wrap(err, "failed to validate struct"))
		}
	}
	return ft.NewValidationErrors()
}

func By(fn func(value any) error) Rule {
	return apiVal.By(fn)
}

func Each(rules ...Rule) EachRule {
	return apiVal.Each(rules...)
}

func Field(fieldPtr any, rules ...Rule) *FieldRules {
	return apiVal.Field(fieldPtr, rules...)
}

func Map(keys ...*KeyRules) MapRule {
	return apiVal.Map(keys...)
}

func Key(key any, rules ...Rule) *KeyRules {
	return apiVal.Key(key, rules...)
}

func OneOf(values ...any) OneOfRule {
	return apiVal.In(values...)
}

func Length(min, max int) LengthRule {
	return apiVal.Length(min, max)
}

func Max(max any) ThresholdRule {
	return apiVal.Max(max)
}

func Min(min any) ThresholdRule {
	return apiVal.Min(min)
}

func NotEmptyWhen(condition bool) WhenRule {
	return apiVal.When(condition, NotEmpty)
}

func NotNilWhen(condition bool) WhenRule {
	return apiVal.When(condition, NotNil)
}

func RegExp(re *regexp.Regexp) MatchRule {
	return apiVal.Match(re)
}

func When(condition bool, rules ...Rule) WhenRule {
	return apiVal.When(condition, rules...)
}
