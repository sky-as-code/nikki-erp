package enum_util

import (
	"bytes"
	"fmt"
	"strconv"
)

func UnmarshalJSON[T ~uint8](data []byte, valueMap map[string]T, nameMap map[T]string) (T, error) {
	data = bytes.TrimSpace(data)

	if len(data) > 0 && data[0] != byte('"') {
		i, err := strconv.ParseUint(string(data), 10, 8)
		if err != nil {
			return 0, fmt.Errorf("invalid enum number: %s", data)
		}

		v := T(i)
		if _, ok := nameMap[v]; !ok {
			return 0, fmt.Errorf(
				"enum '%d' is not registered, must be one of: %v",
				i,
				DescriptionFromMap(nameMap),
			)
		}

		return v, nil
	}

	data = bytes.Trim(data, "\"")
	v, ok := valueMap[string(data)]
	if !ok {
		return 0, fmt.Errorf(
			"enum '%d' is not registered, must be one of: %v",
			data,
			DescriptionFromMap(nameMap),
		)
	}

	return v, nil
}

func MarshalJSON[T ~uint8](e *T, nameMap map[T]string) ([]byte, error) {
	v, ok := nameMap[*e]
	if !ok {
		return fmt.Appendf(nil, "\"%s\"", ""), nil
	}

	buffer := bytes.NewBufferString("\"")
	buffer.WriteString(v)
	buffer.WriteString("\"")
	return buffer.Bytes(), nil
}
