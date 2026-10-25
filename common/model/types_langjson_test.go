package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLangJsonNormalize(t *testing.T) {
	enUsViVn := []LanguageCode{LanguageCodeViVn, LanguageCodeEnUs}
	cases := []struct {
		name    string
		input   LangJson
		enabled []LanguageCode
		want    LangJson
	}{
		{"nil stays nil", nil, enUsViVn, nil},
		{"empty stays empty", LangJson{}, enUsViVn, LangJson{}},
		{"ref untouched", LangJson{"$ref": "a.b", "fr-FR": "x"}, enUsViVn, LangJson{"$ref": "a.b", "fr-FR": "x"}},
		{"strips unsupported", LangJson{"en-US": "Hi", "fr-FR": "Salut"}, []LanguageCode{"en-US"}, LangJson{"en-US": "Hi"}},
		{"fills missing from before", LangJson{"en-US": "Hi"}, enUsViVn, LangJson{"en-US": "Hi", "vi-VN": "Hi"}},
		{"fills from after when nothing before", LangJson{"vi-VN": "Chào"}, enUsViVn,
			LangJson{"en-US": "Chào", "vi-VN": "Chào"}},
		{"fills empty values", LangJson{"en-US": "", "vi-VN": "Chào"}, enUsViVn,
			LangJson{"en-US": "Chào", "vi-VN": "Chào"}},
		{"nearest before wins", LangJson{"de-DE": "A", "en-US": "B"}, []LanguageCode{"de-DE", "en-US", "fr-FR", "vi-VN"},
			LangJson{"de-DE": "A", "en-US": "B", "fr-FR": "B", "vi-VN": "B"}},
		{"canonicalizes keys", LangJson{"en_us": "Hi"}, enUsViVn, LangJson{"en-US": "Hi", "vi-VN": "Hi"}},
		{"nil enabled keeps and fills present keys", LangJson{"fr-FR": "", "en-US": "Hi"}, nil,
			LangJson{"en-US": "Hi", "fr-FR": "Hi"}},
		{"nothing supported leaves empty", LangJson{"fr-FR": "Salut"}, enUsViVn, LangJson{}},
		{"all empty is only stripped", LangJson{"en-US": ""}, enUsViVn, LangJson{"en-US": ""}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.input.Normalize(tc.enabled))
		})
	}
}

func TestLangJsonNormalizeDoesNotMutateInput(t *testing.T) {
	input := LangJson{"en-US": "Hi"}
	_ = input.Normalize([]LanguageCode{LanguageCodeEnUs, LanguageCodeViVn})
	assert.Equal(t, LangJson{"en-US": "Hi"}, input)
}
