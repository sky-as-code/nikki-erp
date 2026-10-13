package dynamicmodel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSearchQueryFieldsAcceptTheWildcard(t *testing.T) {
	q := SearchQuery{Page: 0, Size: 10, Fields: []string{"*", "template.name"}}

	_, cErrs := q.GetSchema().ValidateStruct(q)

	assert.Equal(t, 0, cErrs.Count(), "%v", cErrs)
}

func TestSearchQueryFieldsStillRejectOneCharacterNames(t *testing.T) {
	q := SearchQuery{Page: 0, Size: 10, Fields: []string{"x"}}

	_, cErrs := q.GetSchema().ValidateStruct(q)

	assert.Greater(t, cErrs.Count(), 0)
}
