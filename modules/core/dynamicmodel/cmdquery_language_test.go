package dynamicmodel

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/common/model"
)

func TestSearchQueryLanguageSurvivesValidation(t *testing.T) {
	code := model.LanguageCode("vi-VN")
	q := SearchQuery{Page: 0, Size: 10, Language: &code}

	sanitized, cErrs := q.GetSchema().ValidateStruct(q)
	require.Equal(t, 0, cErrs.Count(), "%v", cErrs)

	out := *(sanitized.(*SearchQuery))
	require.NotNil(t, out.Language, "language was dropped by validation")
	assert.Equal(t, "vi-VN", string(*out.Language))
}

func TestSearchQueryLanguageNilStaysNil(t *testing.T) {
	q := SearchQuery{Page: 0, Size: 10}
	sanitized, cErrs := q.GetSchema().ValidateStruct(q)
	require.Equal(t, 0, cErrs.Count(), "%v", cErrs)
	assert.Nil(t, (*(sanitized.(*SearchQuery))).Language)
}

func TestSearchQueryLanguageMalformedIsRejected(t *testing.T) {
	code := model.LanguageCode("klingon!!")
	q := SearchQuery{Page: 0, Size: 10, Language: &code}
	_, cErrs := q.GetSchema().ValidateStruct(q)
	assert.Greater(t, cErrs.Count(), 0, "a malformed locale should be a client error")
}
