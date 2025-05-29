package util

func IsEmptyStr(target *string) bool {
	return target == nil || len(*target) == 0
}

func SetDefaultStr[T string](target *T, defaultValue T) {
	if target == nil || len(*target) == 0 {
		target = &defaultValue
	}
}

func SetDefaultValue[T any](target *T, defaultValue T) {
	if target == nil {
		target = &defaultValue
	}
}

func SafeVal[T any](source *T, fallbackValue T) T {
	if source != nil {
		return *source
	}
	return fallbackValue
}
