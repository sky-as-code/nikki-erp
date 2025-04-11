package json

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/tidwall/gjson"
)

var (
	JSON = jsoniter.ConfigCompatibleWithStandardLibrary
)

// UseNumber solve very big int64 digits loss.
func UseNumber() {
	JSON = jsoniter.Config{
		UseNumber:              true,
		EscapeHTML:             false,
		SortMapKeys:            true,
		ValidateJsonRawMessage: true,
	}.Froze()
}

// Marshal returns the JSON encoding of v.
func Marshal(v interface{}) ([]byte, error) {
	return JSON.Marshal(v)
}

// MustMarshal must returns the JSON encoding of v.
// func MustMarshal(v interface{}) []byte {
// 	data, _ := JSON.Marshal(v)
// 	return data
// }

// MarshalToString returns the JSON encoding to string of v.
func MarshalToString(v interface{}) (string, error) {
	return JSON.MarshalToString(v)
}

// MustMarshalToString must returns the JSON encoding to string of v.
// func MustMarshalToString(v interface{}) string {
// 	str, _ := JSON.MarshalToString(v)
// 	return str
// }

// Unmarshal parses the JSON-encoded data and stores the result
// in the value pointed to by v.
func Unmarshal(data []byte, v interface{}) error {
	return JSON.Unmarshal(data, v)
}

// UnmarshalStr unmarshal string to v.
func UnmarshalStr(str string, v interface{}) error {
	return JSON.UnmarshalFromString(str, v)
}

// UnmarshalFromJson unmarshal imini to v.
func UnmarshalFromJson(b interface{}, v interface{}) error {
	data, err := JSON.Marshal(b)
	if err != nil {
		return err
	}
	return JSON.Unmarshal(data, v)
}

// IsValidBytes validates JSON data.
func IsValidBytes(data []byte) bool {
	return gjson.ValidBytes(data)
}

// IsValidStr validates JSON string.
func IsValidStr(str string) bool {
	return gjson.Valid(str)
}

// ParseBytes parses the JSON bytes and returns a gjson.Result which allows chaining further operations.
func ParseBytes(data []byte) gjson.Result {
	return gjson.ParseBytes(data)
}

// ParseStr parses the JSON string and returns a gjson.Result which allows chaining further operations.
func ParseStr(data string) gjson.Result {
	return gjson.Parse(data)
}
