package logging

import (
	stdJson "encoding/json"
	"fmt"

	"go.bryk.io/pkg/errors"

	fastJson "github.com/sky-as-code/nikki-erp/common/json"
)

// NewJsonCodec encodes error data as JSON documents. If `pretty`
// is set to `true` the output will be indented for readability.
func NewJsonCodec(pretty bool) errors.Codec {
	return &jsonCodec{pretty: pretty}
}

type errReport struct {
	Message    string   `json:"message,omitempty"`
	Timestamp  int64    `json:"timestamp,omitempty"`
	Stacktrace []string `json:"stacktrace,omitempty"`
}

type jsonCodec struct {
	pretty bool
}

func (c *jsonCodec) Marshal(err error) ([]byte, error) {
	rec := new(errReport)
	rec.Message = err.Error()
	var oe *errors.Error
	if errors.As(err, &oe) {
		rec.Timestamp = oe.Stamp()
		rec.Stacktrace = miniStackTrace(oe.PortableTrace())
	}
	if c.pretty {
		return stdJson.MarshalIndent(rec, "", "  ")
	}
	return fastJson.Marshal(rec)
}

func (c *jsonCodec) Unmarshal(src []byte) (bool, error) {
	return false, errors.New("not implemented")
}

func miniStackTrace(frames []errors.StackFrame) []string {
	lines := make([]string, len(frames))
	for i, frame := range frames {
		lines[i] = fmt.Sprintf("%s:%d %s",
			frame.File,
			frame.LineNumber,
			frame.Function,
		)
	}
	return lines
}
