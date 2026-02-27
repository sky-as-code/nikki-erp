package enum_util

import "bytes"

func DescriptionFromMap[T ~uint8](m map[T]string) string {
	nameBuffer := bytes.NewBufferString("")
	valueBuffer := bytes.NewBufferString("")
	for k, v := range m {
		nameBuffer.WriteByte(byte(k))
		nameBuffer.WriteString(" ")

		valueBuffer.WriteString(v)
		valueBuffer.WriteString(" ")
	}

	buf := bytes.NewBufferString("[ ")
	buf.WriteString(nameBuffer.String())
	buf.WriteString(valueBuffer.String())
	buf.WriteString("]")

	return buf.String()
}
