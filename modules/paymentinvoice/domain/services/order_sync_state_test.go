package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The sync log is what bounds the retry sweep and what a human reads when a tenant says it was
// never told a payment settled. These tests cover the two things that has to get right: not
// growing without bound, and surviving whatever shape the JSON column comes back in.

func TestAnOutcomeIsAppendedToAnEmptyLog(t *testing.T) {
	logs := appendSyncLog(nil, SyncOutcome{
		Status:   SyncStatusFailure,
		Attempts: 3,
		Detail:   "the ordering system answered 502 Bad Gateway",
		At:       time.Date(2026, 8, 18, 12, 0, 0, 0, time.UTC),
	})

	entries := syncLogEntriesOf(logs)
	require.Len(t, entries, 1)

	entry, ok := entries[0].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "paymentResult", entry["type"])
	assert.Equal(t, SyncStatusFailure, entry["status"])
	assert.Equal(t, 3, entry["attempts"])
	assert.Equal(t, "2026-08-18T12:00:00Z", entry["timestamp"])
}

func TestOutcomesAccumulateInOrder(t *testing.T) {
	logs := appendSyncLog(nil, SyncOutcome{Status: SyncStatusFailure, At: time.Now()})
	logs = appendSyncLog(logs, SyncOutcome{Status: SyncStatusSuccess, At: time.Now()})

	entries := syncLogEntriesOf(logs)
	require.Len(t, entries, 2)
	assert.Equal(t, SyncStatusFailure, entries[0].(map[string]any)["status"])
	assert.Equal(t, SyncStatusSuccess, entries[1].(map[string]any)["status"])
}

// An order whose tenant has been unreachable for days would otherwise grow an unbounded JSON blob
// in a column that is read every time the order is.
func TestTheLogIsCappedAndKeepsTheMostRecentEntries(t *testing.T) {
	var logs map[string]any
	for i := 0; i < syncLogLimit+10; i++ {
		logs = appendSyncLog(logs, SyncOutcome{
			Status:   SyncStatusFailure,
			Attempts: i,
			At:       time.Now(),
		})
	}

	entries := syncLogEntriesOf(logs)
	require.Len(t, entries, syncLogLimit)

	// The oldest were dropped, not the newest: the recent attempts are the ones that explain the
	// current state.
	assert.Equal(t, syncLogLimit+9, entries[len(entries)-1].(map[string]any)["attempts"])
	assert.Equal(t, 10, entries[0].(map[string]any)["attempts"])
}

// The column round-trips through JSON, so the entry list comes back as []any rather than the
// []map[string]any it was written as. Counting must survive that.
func TestAttemptsAreCountedFromADecodedColumn(t *testing.T) {
	decoded := map[string]any{
		syncLogEntriesKey: []any{
			map[string]any{"status": SyncStatusFailure},
			map[string]any{"status": SyncStatusFailure},
		},
	}

	assert.Equal(t, 2, syncAttemptsOf(decoded))
}

// A malformed or absent log must read as empty rather than panicking: a bad log must not stop a
// payment from being reported.
func TestAMalformedLogReadsAsEmpty(t *testing.T) {
	for name, logs := range map[string]map[string]any{
		"nil":            nil,
		"no entries key": {"something": "else"},
		"wrong type":     {syncLogEntriesKey: "not a list"},
		"a number":       {syncLogEntriesKey: 42},
	} {
		assert.Zero(t, syncAttemptsOf(logs), name)
		assert.Empty(t, syncLogEntriesOf(logs), name)
	}
}

// Appending to a malformed log must still produce a usable one, rather than carrying the damage
// forward forever.
func TestAppendingToAMalformedLogRepairsIt(t *testing.T) {
	logs := appendSyncLog(map[string]any{syncLogEntriesKey: "not a list"}, SyncOutcome{
		Status: SyncStatusSuccess,
		At:     time.Now(),
	})

	assert.Len(t, syncLogEntriesOf(logs), 1)
}

// The statuses reported to the ordering system must stay the module's own values, so there is one
// vocabulary and one translation rather than two sets of names drifting apart.
func TestTheSyncStatusesAreTheModulesOwn(t *testing.T) {
	assert.Equal(t, "payment_success", OrderStatusPaidForSync)
	assert.Equal(t, "payment_failed", OrderStatusFailedForSync)
	assert.Equal(t, "expired", OrderStatusExpiredForSync)
}
