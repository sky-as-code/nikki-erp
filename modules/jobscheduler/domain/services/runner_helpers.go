package services

import (
	"encoding/json"
	stdErr "errors"
	"strings"
	"time"

	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/jobscheduler/domain/models"
)

func strPtrOf(v string) *string {
	return &v
}

func int32PtrOf(v int32) *int32 {
	return &v
}

func goTimePtr(v *model.ModelDateTime) *time.Time {
	if v == nil {
		return nil
	}
	converted := v.GoTime().UTC()
	return &converted
}

func modelDateTimePtr(v *time.Time) *model.ModelDateTime {
	if v == nil {
		return nil
	}
	return toModelDateTime(*v)
}

func maxTime(a time.Time, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

// truncate bounds a value to what its column accepts.
//
// It cuts at a rune boundary rather than a byte one: slicing a UTF-8 string mid-rune produces
// bytes that are not valid text, which some drivers reject outright and others store as a
// mojibake tail nobody can read.
func truncate(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	runes := []rune(value)
	for limit > 0 && len(string(runes[:min(len(runes), limit)])) > limit {
		limit--
	}
	return strings.ToValidUTF8(string(runes[:min(len(runes), limit)]), "")
}

// snapshotAsMap converts the snapshot to what the jsonmap column stores.
//
// It round-trips through JSON rather than reflecting over the struct so that the stored shape is
// exactly what the json tags declare - the same bytes a reader will decode with.
func snapshotAsMap(snapshot JobSnapshot) (map[string]any, error) {
	encoded, err := json.Marshal(snapshot)
	if err != nil {
		return nil, err
	}
	var asMap map[string]any
	if err := json.Unmarshal(encoded, &asMap); err != nil {
		return nil, err
	}
	return asMap, nil
}

// snapshotOf reads the frozen configuration back off an execution row.
func snapshotOf(execution models.Execution) (JobSnapshot, error) {
	var snapshot JobSnapshot
	raw := execution.GetJobSnapshot()
	if raw == nil {
		return snapshot, nil
	}
	encoded, err := json.Marshal(raw)
	if err != nil {
		return snapshot, err
	}
	err = json.Unmarshal(encoded, &snapshot)
	return snapshot, err
}

// isDuplicateKey reports whether err is a unique-constraint violation.
//
// It matches on the message because the repository layer wraps the driver error and the sqlstate
// is no longer reachable. That is worth knowing about: a driver changing its wording would make
// a duplicate execution_key surface as a tick error rather than as the no-op it should be. The
// consequence is a logged error, not incorrect data, because the row is not written either way.
func isDuplicateKey(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "duplicate key") ||
		strings.Contains(message, "unique constraint") ||
		strings.Contains(message, "sqlstate 23505")
}

// joinTickErrors collapses the per-phase errors into one, or nil.
//
// The engine logs what a tick returns and carries on regardless, so this is for the operator
// rather than for control flow: a tick that failed three ways should say so once.
func joinTickErrors(errs []error) error {
	if len(errs) == 0 {
		return nil
	}
	return stdErr.Join(errs...)
}

// durationMsSince is how long an attempt took, in milliseconds.
//
// It clamps at zero rather than reporting a negative duration: the two instants can come from
// different sources, and a clock stepping backwards between them would otherwise store a
// negative value that every reader would have to defend against.
func durationMsSince(startedAt time.Time, finishedAt time.Time) int64 {
	elapsed := finishedAt.Sub(startedAt)
	if elapsed < 0 {
		return 0
	}
	return elapsed.Milliseconds()
}
