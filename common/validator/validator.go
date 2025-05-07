package validator

import (
	"regexp"

	tagVal "github.com/go-playground/validator/v10"
	apiVal "github.com/invopop/validation"
	"github.com/invopop/validation/is"
	. "github.com/sky-as-code/nikki-erp/common/fault"
)

var TagBased = TagBasedValidator{
	validator: tagVal.New(tagVal.WithRequiredStructEnabled()),
}

type TagBasedValidator struct {
	validator *tagVal.Validate
}

func (v TagBasedValidator) Validate(i interface{}) error {
	// TODO: Must convert to more friendly error struct
	return v.validator.Struct(i)
}

var (
	ApiBased       = newApiBasedValidator()
	Required       = apiVal.Required
	IsAlpha        = is.Alpha
	IsAlphaNumeric = is.Alphanumeric
	IsDomain       = is.Domain
	IsEmail        = is.EmailFormat
	IsNumeric      = is.Digit
	IsUuid         = is.UUID
	IsUrl          = is.URL
)

func newApiBasedValidator() ApiBasedValidator {
	apiVal.ErrorTag = "label"
	return ApiBasedValidator{}
}

type FieldRules = apiVal.FieldRules
type LengthRule = apiVal.LengthRule
type MatchRule = apiVal.MatchRule
type OneOfRule = apiVal.InRule[any]
type Rule = apiVal.Rule
type ThresholdRule = apiVal.ThresholdRule
type WhenRule = apiVal.WhenRule

type ApiBasedValidator struct {
}

func (v ApiBasedValidator) ValidateStruct(structPtr interface{}, fields ...*apiVal.FieldRules) error {
	err := apiVal.ValidateStruct(structPtr, fields...)
	if err != nil {
		return NewValidationErrorFromOzzo(err.(apiVal.Errors))
	}
	return nil
}

func Field(fieldPtr interface{}, rules ...Rule) *FieldRules {
	return apiVal.Field(fieldPtr, rules...)
}

func OneOf(values ...interface{}) OneOfRule {
	return apiVal.In(values...)
}

func Length(min, max int) LengthRule {
	return apiVal.Length(min, max)
}

func Max(max interface{}) ThresholdRule {
	return apiVal.Max(max)
}

func Min(min interface{}) ThresholdRule {
	return apiVal.Min(min)
}

func RegExp(re *regexp.Regexp) MatchRule {
	return apiVal.Match(re)
}

func RequiredWhen(condition bool) WhenRule {
	return apiVal.When(condition, Required)
}

func When(condition bool, rules ...Rule) WhenRule {
	return apiVal.When(condition, rules...)
}
