package util

func IsEmptyStr(target *string) bool {
	return target == nil || len(*target) == 0
}
