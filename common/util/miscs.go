package util

func IsEmptyStr(target *string) bool {
	return target == nil || len(*target) == 0
}

// CopyMap returns a shallow copy of src. Returns nil if src is nil.
func CopyMap[K comparable, V any](src map[K]V) map[K]V {
	if src == nil {
		return nil
	}
	dst := make(map[K]V, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
