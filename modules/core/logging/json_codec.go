package logging

import (
	"fmt"

	"go.bryk.io/pkg/errors"

	fastJson "github.com/sky-as-code/nikki-erp/common/json"
)

// NewJsonCodec encodes error data as JSON documents. If `pretty`
// is set to `true` the output will be indented for readability.
func NewJsonCodec(skipStacktrace bool) errors.Codec {
	return &jsonCodec{skipStacktrace}
}

type errReport struct {
	Message    string   `json:"message,omitempty"`
	Timestamp  int64    `json:"timestamp,omitempty"`
	Stacktrace []string `json:"stacktrace,omitempty"`
}

type jsonCodec struct {
	skipStacktrace bool
}

func (this *jsonCodec) Marshal(err error) ([]byte, error) {
	rec := NewErrReport(err, this.skipStacktrace)
	return fastJson.Marshal(rec)
}

func (this *jsonCodec) Unmarshal(src []byte) (bool, error) {
	return false, errors.New("not implemented")
}

func NewErrReport(err error, skipStacktrace bool) *errReport {
	rec := new(errReport)
	rec.Message = err.Error()
	var oe *errors.Error
	if errors.As(err, &oe) {
		rec.Timestamp = oe.Stamp()
		if !skipStacktrace {
			rec.Stacktrace = miniStackTrace(oe.PortableTrace())
		}
	}
	return rec
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
