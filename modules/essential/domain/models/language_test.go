package models

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sky-as-code/nikki-erp/common/util"
)

// The locale code is the join between this table and everything that names a language: the stored
// `language` setting, the LangJson document keys, i18next on the frontend. All of them spell it
// with a hyphen, so a row spelled with an underscore would be unreachable from any of them.
func TestLanguageLocaleCodeAcceptsHyphenForm(t *testing.T) {
	requireBaseSchemasRegistered(t)
	field := requireField(t, LanguageSchemaBuilder().Build(), LanguageFieldLocaleCode)

	for _, code := range []string{"en-US", "vi-VN"} {
		_, clientErr := field.Validate(code)

		assert.Nilf(t, clientErr, "locale code %q must validate", code)
	}
}

// Pins the decision rather than the mechanism: the underscore form was the original spelling and
// matched no consumer, so a well-meaning revert has to fail here rather than in production.
func TestLanguageLocaleCodeRejectsUnderscoreForm(t *testing.T) {
	requireBaseSchemasRegistered(t)
	field := requireField(t, LanguageSchemaBuilder().Build(), LanguageFieldLocaleCode)

	_, clientErr := field.Validate("en_US")

	assert.NotNil(t, clientErr)
}

func TestLanguageGettersRoundTrip(t *testing.T) {
	lang := NewLanguage()

	lang.SetName(util.ToPtr("English"))
	lang.SetLocaleCode(util.ToPtr("en-US"))
	lang.SetDirection(util.ToPtr(string(LanguageDirectionLtr)))
	lang.SetDecimalSeparator(util.ToPtr("."))
	lang.SetThousandsSeparator(util.ToPtr(","))
	lang.SetGrouping(util.ToPtr("[3,0]"))
	lang.SetDateFormat(util.ToPtr("%m/%d/%Y"))
	lang.SetTimeFormat(util.ToPtr("%H:%M:%S"))
	lang.SetShortTimeFormat(util.ToPtr("%H:%M"))
	lang.SetFirstDayOfWeek(util.ToPtr(string(WeekDaySunday)))

	assert.Equal(t, "English", *lang.GetName())
	assert.Equal(t, "en-US", *lang.GetLocaleCode())
	assert.Equal(t, "ltr", *lang.GetDirection())
	assert.Equal(t, ".", *lang.GetDecimalSeparator())
	assert.Equal(t, ",", *lang.GetThousandsSeparator())
	assert.Equal(t, "[3,0]", *lang.GetGrouping())
	assert.Equal(t, "%m/%d/%Y", *lang.GetDateFormat())
	assert.Equal(t, "%H:%M:%S", *lang.GetTimeFormat())
	assert.Equal(t, "%H:%M", *lang.GetShortTimeFormat())
	assert.Equal(t, "sunday", *lang.GetFirstDayOfWeek())
}
