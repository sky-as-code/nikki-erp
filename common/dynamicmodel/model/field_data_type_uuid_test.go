package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// lib/pq hands a uuid column back as its text form in a []byte; reading it must not fail.
func TestUuid_TryConvertAcceptsDriverBytes(t *testing.T) {
	text := "01a1a2e7-de4d-7b7b-9b90-8bec8ce4418f"
	converted, err := FieldDataTypeUuid().TryConvert([]byte(text), nil)
	require.NoError(t, err)
	assert.Equal(t, text, *converted.Get())
}
