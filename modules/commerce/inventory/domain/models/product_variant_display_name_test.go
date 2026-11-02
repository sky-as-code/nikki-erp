package models

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sky-as-code/nikki-erp/common/model"
)

func TestFormatVariantDisplayNameSkipsEmptyValues(t *testing.T) {
	assert.Equal(t, "Coca Cola (No sugar, Big)", FormatVariantDisplayName("Coca Cola", []string{"No sugar", "", "Big"}))
	assert.Equal(t, "Coca Cola", FormatVariantDisplayName("Coca Cola", nil))
	assert.Equal(t, "Coca Cola", FormatVariantDisplayName("Coca Cola", []string{""}))
}

// Every enabled locale gets its own name; a value translated into fewer languages borrows its
// neighbour's text rather than leaving a gap.
func TestBuildVariantDisplayNameRendersEveryLocale(t *testing.T) {
	locales := []model.LanguageCode{model.LanguageCodeEnUs, model.LanguageCodeViVn}
	name := model.LangJson{model.LanguageCodeEnUs: "Coca Cola", model.LanguageCodeViVn: "Cô ca"}
	values := []model.LangJson{
		{model.LanguageCodeEnUs: "No sugar", model.LanguageCodeViVn: "Không đường"},
		{model.LanguageCodeEnUs: "Big"},
	}

	got := BuildVariantDisplayName(name, values, locales, "SKU-1")

	assert.Equal(t, model.LangJson{
		model.LanguageCodeEnUs: "Coca Cola (No sugar, Big)",
		model.LanguageCodeViVn: "Cô ca (Không đường, Big)",
	}, got)
}

// An unknown enabled set falls back to the template name's own locales.
func TestBuildVariantDisplayNameWithoutEnabledLocales(t *testing.T) {
	got := BuildVariantDisplayName(model.LangJson{model.LanguageCodeViVn: "Trà"}, nil, nil, "")

	assert.Equal(t, model.LangJson{model.LanguageCodeViVn: "Trà"}, got)
}

// A template with no name is named by the fallback (the SKU), never by an empty string.
func TestBuildVariantDisplayNameFallsBackWithoutATemplateName(t *testing.T) {
	got := BuildVariantDisplayName(nil, []model.LangJson{{model.LanguageCodeEnUs: "Big"}},
		[]model.LanguageCode{model.LanguageCodeEnUs}, "CC-BIG")

	assert.Equal(t, model.LangJson{model.LanguageCodeEnUs: "CC-BIG (Big)"}, got)
}
