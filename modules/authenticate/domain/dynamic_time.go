package domain

import (
	"time"

	"github.com/sky-as-code/nikki-erp/common/model"
)

func AnyToTimePtr(v any) *time.Time {
	if v == nil {
		return nil
	}
	switch t := v.(type) {
	case model.ModelDateTime:
		gt := t.GoTime()
		return &gt
	case *model.ModelDateTime:
		if t == nil {
			return nil
		}
		gt := t.GoTime()
		return &gt
	case time.Time:
		return &t
	case *time.Time:
		return t
	default:
		return nil
	}
}

func TimePtrToModelDateTime(v *time.Time) model.ModelDateTime {
	if v == nil {
		return model.ModelDateTime{}
	}
	return model.ModelDateTime(*v)
}
