package model

import (
	"encoding/json"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/common/model"
)

func validateLangJson(t *testing.T, dataType FieldDataType, input model.LangJson) (model.LangJson, string) {
	t.Helper()
	result, cErr := dataType.Validate(Value(input))
	if cErr != nil {
		return nil, cErr.Field
	}
	langObj, ok := (*result.Get()).(model.LangJson)
	require.True(t, ok, "a valid LangJson must come back as model.LangJson")
	return langObj, ""
}

// The engine hook strips unsupported languages from the tenant's enabled list; the data type no
// longer carries a hard-coded one, so any well-formed language survives validation.
func TestLangJsonKeepsLanguagesOutsideTheFormerWhitelist(t *testing.T) {
	got, errField := validateLangJson(t, FieldDataTypeLangJson(1, 50), model.LangJson{"en-US": "Hi", "fr-FR": "Salut"})
	require.Empty(t, errField)
	assert.Equal(t, model.LangJson{"en-US": "Hi", "fr-FR": "Salut"}, got)
}

func TestLangJsonHonoursAnExplicitWhitelist(t *testing.T) {
	dataType := FieldDataTypeLangJson(1, 50).(fieldDataTypeLangJson)
	dataType.options[FieldDataTypeOptLangJsonWhitelist] = []model.LanguageCode{model.LanguageCodeEnUs}
	got, errField := validateLangJson(t, dataType, model.LangJson{"en-US": "Hi", "fr-FR": "Salut"})
	require.Empty(t, errField)
	assert.Equal(t, model.LangJson{"en-US": "Hi"}, got)
}

func TestLangJsonAppliesLengthToEveryTranslation(t *testing.T) {
	_, errField := validateLangJson(t, FieldDataTypeLangJson(1, 5), model.LangJson{"en-US": "Hi", "vi-VN": "Too long"})
	assert.Equal(t, ".vi-VN", errField)
}

func TestLangJsonAppliesPatternToEveryTranslation(t *testing.T) {
	dataType := FieldDataTypeLangJson(1, 50, FieldDataTypeStringOpts{Regex: regexp.MustCompile(`^[A-Za-z ]+$`)})

	_, errField := validateLangJson(t, dataType, model.LangJson{"en-US": "Hello", "vi-VN": "Hello 2"})
	assert.Equal(t, ".vi-VN", errField)

	_, errField = validateLangJson(t, dataType, model.LangJson{"en-US": "Hello", "vi-VN": "Xin chao"})
	assert.Empty(t, errField)
}

func TestLangJsonExemptsTheRefKeyFromTextRules(t *testing.T) {
	dataType := FieldDataTypeLangJson(1, 5, FieldDataTypeStringOpts{Regex: regexp.MustCompile(`^[A-Z]+$`)})
	got, errField := validateLangJson(t, dataType, model.LangJson{"$ref": "inventory.product.long_key"})
	require.Empty(t, errField)
	assert.Equal(t, "inventory.product.long_key", got.TranslationKey())
}

func TestLangJsonFromJsonModelCarriesThePattern(t *testing.T) {
	dataType := constructDataType(&dataTypeJsonDto{Type: "langjson", Min: 1.0, Max: 10.0, Regex: "^[a-z]+$"}, "name")
	_, hasPattern := dataType.Options()[FieldDataTypeOptPattern]
	assert.True(t, hasPattern)
}

// meta/schema serializes ToSimplized; a compiled pattern must reach clients as its source string.
func TestLangJsonSimplizedExposesThePatternSource(t *testing.T) {
	dataType := FieldDataTypeLangJson(1, 50, FieldDataTypeStringOpts{Regex: regexp.MustCompile(`^[a-z]+$`)})
	raw, err := json.Marshal(dataType.ToSimplized())
	require.NoError(t, err)
	assert.Contains(t, string(raw), `"pattern":"^[a-z]+$"`)
	_, isCompiled := dataType.Options()[FieldDataTypeOptPattern].(*regexp.Regexp)
	assert.True(t, isCompiled, "simplizing must not mutate the validator's options")
}

func TestLangJsonModelJsonRoundTripsThePattern(t *testing.T) {
	dataType := constructDataType(&dataTypeJsonDto{Type: "langjson", Min: 1.0, Max: 10.0, Regex: "^[a-z]+$"}, "name")
	object, err := dataTypeToModelJson(dataType)
	require.NoError(t, err)
	assert.Equal(t, "^[a-z]+$", object.(map[string]any)["regex"])
}
