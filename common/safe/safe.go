package safe

import "time"

func SetDefaultStr[T string](target *T, defaultValue T) {
	if target == nil || len(*target) == 0 {
		target = &defaultValue
	}
}

// if target is nil, set it to defaultValue
// target must keep the value when this function is returned
func SetDefaultValue[T any](target **T, defaultValue T) {
	if *target == nil {
		*target = &defaultValue
	}
}

func GetVal[T any](source *T, fallbackValue T) T {
	if source != nil {
		return *source
	}
	return fallbackValue
}

func GetTimeUnix(source *time.Time) *int64 {
	if source != nil {
		unix := source.Unix()
		return &unix
	}
	return nil
}
