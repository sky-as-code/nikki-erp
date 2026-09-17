package models

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sky-as-code/nikki-erp/common/util"
)

// The iso code is the join between this table and everything that names a language: the stored
// `language` setting, the LangJson document keys, i18next on the frontend. All of them spell it
// with a hyphen, so a row spelled with an underscore would be unreachable from any of them.
func TestLanguageIsoCodeAcceptsHyphenForm(t *testing.T) {
	requireBaseSchemasRegistered(t)
	field := requireField(t, LanguageSchemaBuilder().Build(), LanguageFieldIsoCode)

	for _, code := range []string{"en-US", "vi-VN"} {
		_, clientErr := field.Validate(code)

		assert.Nilf(t, clientErr, "iso code %q must validate", code)
	}
}

// Pins the decision rather than the mechanism: the underscore form was the original spelling and
// matched no consumer, so a well-meaning revert has to fail here rather than in production.
func TestLanguageIsoCodeRejectsUnderscoreForm(t *testing.T) {
	requireBaseSchemasRegistered(t)
	field := requireField(t, LanguageSchemaBuilder().Build(), LanguageFieldIsoCode)

	_, clientErr := field.Validate("en_US")

	assert.NotNil(t, clientErr)
}

func TestLanguageGettersRoundTrip(t *testing.T) {
	lang := NewLanguage()

	lang.SetName(util.ToPtr("English"))
	lang.SetIsoCode(util.ToPtr("en-US"))
	lang.SetDirection(util.ToPtr(string(LanguageDirectionLtr)))
	lang.SetDecimalSeparator(util.ToPtr("."))
	lang.SetThousandsSeparator(util.ToPtr(","))
	lang.SetDateFormat(util.ToPtr("MM/dd/yyyy"))
	lang.SetTimeFormat(util.ToPtr("HH:mm:ss"))
	lang.SetShortTimeFormat(util.ToPtr("HH:mm"))
	lang.SetFirstDayOfWeek(util.ToPtr(string(WeekDaySunday)))

	assert.Equal(t, "English", *lang.GetName())
	assert.Equal(t, "en-US", *lang.GetIsoCode())
	assert.Equal(t, "ltr", *lang.GetDirection())
	assert.Equal(t, ".", *lang.GetDecimalSeparator())
	assert.Equal(t, ",", *lang.GetThousandsSeparator())
	assert.Equal(t, "MM/dd/yyyy", *lang.GetDateFormat())
	assert.Equal(t, "HH:mm:ss", *lang.GetTimeFormat())
	assert.Equal(t, "HH:mm", *lang.GetShortTimeFormat())
	assert.Equal(t, "sunday", *lang.GetFirstDayOfWeek())
}
