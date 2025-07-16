package convert

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"time"
)

// FromBytes converts byte array to a struct.
// Output parameter must be a pointer.
func FromBytes(input []byte, output any) error {
	var buf bytes.Buffer
	buf.Write(input)
	return FromByteBuffer(&buf, output)
}

// FromBytes converts byte array to a struct.
// Output parameter must be a pointer.
func FromByteBuffer(input *bytes.Buffer, output any) error {
	enc := gob.NewDecoder(input)
	err := enc.Decode(output)
	return err
}

// ToBytes converts an interface to bytes.
// Input must be a pointer.
func ToBytes(input any) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(input)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// StringToDate attempts to parse a string into a time.Time type using a
// predefined list of formats.  If no suitable format is found, an error is
// returned.
func StringToDate(s string) (time.Time, error) {
	return parseDateWith(s, []string{
		time.RFC3339,
		"2006-01-02T15:04:05", // iso8601 without timezone
		time.RFC1123Z,
		time.RFC1123,
		time.RFC822Z,
		time.RFC822,
		time.RFC850,
		time.ANSIC,
		time.UnixDate,
		time.RubyDate,
		"2006-01-02 15:04:05.999999999 -0700 MST", // Time.String()
		"2006-01-02",
		"02 Jan 2006",
		"2006-01-02T15:04:05-0700", // RFC3339 without timezone hh:mm colon
		"2006-01-02 15:04:05 -07:00",
		"2006-01-02 15:04:05 -0700",
		"2006-01-02 15:04:05Z07:00", // RFC3339 without T
		"2006-01-02 15:04:05Z0700",  // RFC3339 without T or timezone hh:mm colon
		"2006-01-02 15:04:05",
		time.Kitchen,
		time.Stamp,
		time.StampMilli,
		time.StampMicro,
		time.StampNano,
	})
}

func parseDateWith(s string, dates []string) (d time.Time, e error) {
	for _, dateType := range dates {
		if d, e = time.Parse(dateType, s); e == nil {
			return
		}
	}
	return d, fmt.Errorf("unable to parse date: %s", s)
}

func GetStringValue(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}

func GetStringValueByEnum(e any) string {
	if ptr, ok := e.(*string); ok {
		return GetStringValue(ptr)
	}
	return ""
}

func GetStringValueByTime(t *time.Time) string {
	if t != nil {
		return t.String()
	}
	return ""
}

func StringPtrToBytes(s *string) []byte {
	if s == nil {
		return nil
	}
	return []byte(*s)
}
