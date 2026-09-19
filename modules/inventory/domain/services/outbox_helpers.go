package services

import (
	"time"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
)

// Defensive readers over a row that came back through the repository. A value that crossed a
// jsonb column or a generic decoder may arrive as any of several Go types; these settle on one
// rather than type-asserting and panicking a background sweep.

func NowUnix() int64 {
	return time.Now().Unix()
}

func stringOf(record dmodel.DynamicFields, field string) string {
	if record == nil {
		return ""
	}
	value, ok := record[field]
	if !ok || value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return typed
	case *string:
		if typed != nil {
			return *typed
		}
	}
	return ""
}

func int32Of(record dmodel.DynamicFields, field string) int32 {
	value, present := record[field]
	if !present || value == nil {
		return 0
	}
	switch typed := value.(type) {
	case int32:
		return typed
	case *int32:
		if typed != nil {
			return *typed
		}
	case int64:
		return int32(typed)
	case int:
		return int32(typed)
	case float64:
		return int32(typed)
	}
	return 0
}

func dateTimeOf(record dmodel.DynamicFields, field string) *model.ModelDateTime {
	value, present := record[field]
	if !present || value == nil {
		return nil
	}
	switch typed := value.(type) {
	case model.ModelDateTime:
		return &typed
	case *model.ModelDateTime:
		return typed
	case time.Time:
		wrapped := model.WrapModelDateTime(typed)
		return &wrapped
	}
	return nil
}

// truncateError fits a failure message into the outbox's last_error column.
func truncateError(message string) string {
	const maxLength = 1000
	if len(message) <= maxLength {
		return message
	}
	return message[:maxLength]
}
