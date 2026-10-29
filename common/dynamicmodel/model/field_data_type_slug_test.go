package model

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidatePattern_NilPatternAcceptsAndMismatchRejects(t *testing.T) {
	assert.True(t, ValidatePattern("anything", nil))
	assert.True(t, ValidatePattern("abc", regexp.MustCompile(`^[a-z]+$`)))
	assert.False(t, ValidatePattern("ABC", regexp.MustCompile(`^[a-z]+$`)))
}

func TestSlug_AcceptsNormalInputAfterConversion(t *testing.T) {
	cases := map[string]string{
		"my-company":       "my-company",
		"  CoreMart  ":     "coremart",
		"Tech Solutions":   "tech-solutions",
		"lifecycle-org-a1": "lifecycle-org-a1",
	}
	for input, expected := range cases {
		validated, clientErr := FieldDataTypeSlug().Validate(Value(input))
		require.Nilf(t, clientErr, "%q must be accepted", input)
		assert.Equal(t, expected, *validated.Get())
	}
}

func TestSlug_RejectsBadSlug(t *testing.T) {
	for _, input := range []string{"my_company", "my.company", "-abc", "abc-", "a--bc", "café-shop", "my/org"} {
		_, clientErr := FieldDataTypeSlug().Validate(Value(input))
		require.NotNilf(t, clientErr, "%q must be rejected", input)
		assert.Equal(t, "common:err_invalid_data_type", clientErr.Key)
	}
}

func TestSlug_FieldErrorNamesTheField(t *testing.T) {
	schema := ParseModelJson(`{
		"name": "test_slug",
		"fields": [{"name": "slug", "data_type": "slug"}]
	}`).Build()

	_, clientErrs := schema.Validate(map[string]any{"slug": "bad_slug"})
	require.Len(t, clientErrs, 1)
	assert.Equal(t, "slug", clientErrs[0].Field)
}
