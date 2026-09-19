package computed

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/common/model"
)

// A loaded row carries date/time columns as model.ModelDateTime, which is what a computed
// `case/when` over reserved_until or valid_until compares against `now`.
func TestCoerceTime_AcceptsModelDateTime(t *testing.T) {
	at := time.Date(2026, 9, 19, 10, 45, 58, 0, time.UTC)
	wrapped := model.ModelDateTime(at)

	for _, value := range []any{at, &at, wrapped, &wrapped} {
		got, err := coerceTime(value)
		require.NoError(t, err)
		assert.True(t, got.Equal(at))
	}

	_, err := coerceTime("2026-09-19T10:45:58Z")
	assert.Error(t, err)
}
