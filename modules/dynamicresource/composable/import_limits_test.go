package composable

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadImportLimitsWithoutConfigUsesDefaults(t *testing.T) {
	limits := readImportLimits(nil)

	assert.Equal(t, defaultImportMaxFileSizeBytes, limits.MaxFileSizeBytes)
	assert.Equal(t, defaultImportMaxRows, limits.MaxRows)
	assert.Equal(t, []string{"xlsx", "csv"}, limits.AllowedExtensions)
}

func TestNormalizeExtensions(t *testing.T) {
	assert.Equal(t, []string{"xlsx", "csv"}, normalizeExtensions([]string{" .XLSX", "csv", "", " "}))
}

func TestImportLimitsIsAllowed(t *testing.T) {
	limits := ImportLimits{AllowedExtensions: []string{"xlsx", "csv"}}

	assert.True(t, limits.IsAllowed(".CSV"))
	assert.True(t, limits.IsAllowed("xlsx"))
	assert.False(t, limits.IsAllowed("xls"))
	assert.False(t, limits.IsAllowed(""))
}
