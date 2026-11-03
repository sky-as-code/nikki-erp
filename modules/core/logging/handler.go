package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"
	"time"

	"go.opentelemetry.io/otel/trace"
)

const (
	timestampLayout = "2006-01-02T15:04:05.000000Z07:00"

	attrKeyData      = "data"
	attrKeySource    = "source"
	attrKeyErrorType = "error.type"
	attrKeyForeign   = "attrs"
)

type logField struct {
	key   string
	value any
}

// telemetryHandler writes one JSON object per line. Telemetry, operation and resource fields are
// written by the handler itself at the top level; whatever a caller passes stays nested under
// `data` (or `attrs` for third-party slog users), so it can never replace them (LOG-05).
type telemetryHandler struct {
	level  slog.Leveler
	out    io.Writer
	mutex  *sync.Mutex
	attrs  []slog.Attr
	groups []string
}

func NewTelemetryHandler(out io.Writer, level slog.Leveler) slog.Handler {
	return &telemetryHandler{level: level, out: out, mutex: &sync.Mutex{}}
}

func (this *telemetryHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= this.level.Level()
}

func (this *telemetryHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	clone := *this
	clone.attrs = append(append([]slog.Attr{}, this.attrs...), this.prefixed(attrs)...)
	return &clone
}

func (this *telemetryHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return this
	}
	clone := *this
	clone.groups = append(append([]string{}, this.groups...), name)
	return &clone
}

func (this *telemetryHandler) Handle(ctx context.Context, record slog.Record) error {
	if ctx == nil {
		ctx = context.Background()
	}
	fields := this.headerFields(ctx, record)
	fields = append(fields, operationFields(ctx)...)
	fields = append(fields, resourceFieldList()...)
	fields = append(fields, httpFieldList(ctx)...)
	fields = append(fields, this.attrFields(record)...)
	line := encodeFields(fields)

	this.mutex.Lock()
	defer this.mutex.Unlock()
	_, err := this.out.Write(line)
	return err
}

func (this *telemetryHandler) headerFields(ctx context.Context, record slog.Record) []logField {
	stamp := record.Time
	if stamp.IsZero() {
		stamp = time.Now()
	}
	fields := []logField{
		{"timestamp", stamp.UTC().Format(timestampLayout)},
		{"severity", record.Level.String()},
		{"message", record.Message},
	}
	// An invalid span context yields null, never an invented id: a fake trace per line would
	// hide that the operation lost its context.
	spanCtx := trace.SpanContextFromContext(ctx)
	if !spanCtx.IsValid() {
		return append(fields, logField{"trace_id", nil}, logField{"span_id", nil})
	}
	return append(fields,
		logField{"trace_id", spanCtx.TraceID().String()},
		logField{"span_id", spanCtx.SpanID().String()},
	)
}

func operationFields(ctx context.Context) []logField {
	values := collectContextFields(ctx)
	fields := make([]logField, 0, len(contextFieldNames))
	for _, name := range contextFieldNames {
		fields = append(fields, logField{name, values[name]})
	}
	return fields
}

func resourceFieldList() []logField {
	values := snapshotResourceFields()
	fields := make([]logField, 0, len(resourceFieldNames))
	for _, name := range resourceFieldNames {
		var value any
		if text, ok := values[name]; ok {
			value = text
		}
		fields = append(fields, logField{name, value})
	}
	return fields
}

func httpFieldList(ctx context.Context) []logField {
	httpFields, ok := HttpFieldsOf(ctx)
	if !ok {
		return nil
	}
	return []logField{
		{"http.request.method", httpFields.Method},
		{"http.route", httpFields.Route},
		{"http.response.status_code", httpFields.StatusCode},
		{"duration_ms", httpFields.DurationMs},
	}
}

// attrFields sorts the record attributes: the logger's own `error.type`, `source` and `data`
// keep their place, anything else (third-party slog users) is nested under `attrs`.
func (this *telemetryHandler) attrFields(record slog.Record) []logField {
	var errorType, source, data any
	foreign := map[string]any{}
	for _, attr := range this.attrs {
		foreign[attr.Key] = attrValue(attr.Value)
	}
	ownAttrs := len(this.groups) == 0
	record.Attrs(func(attr slog.Attr) bool {
		switch {
		case ownAttrs && attr.Key == attrKeyErrorType:
			errorType = attrValue(attr.Value)
		case ownAttrs && attr.Key == attrKeySource:
			source = attrValue(attr.Value)
		case ownAttrs && attr.Key == attrKeyData:
			data = attrValue(attr.Value)
		default:
			for _, prefixedAttr := range this.prefixed([]slog.Attr{attr}) {
				foreign[prefixedAttr.Key] = attrValue(prefixedAttr.Value)
			}
		}
		return true
	})
	return ownFieldList(errorType, source, data, foreign)
}

func ownFieldList(errorType, source, data any, foreign map[string]any) []logField {
	fields := []logField{}
	if errorType != nil {
		fields = append(fields, logField{attrKeyErrorType, errorType})
	}
	if source != nil {
		fields = append(fields, logField{attrKeySource, source})
	}
	if data != nil {
		fields = append(fields, logField{attrKeyData, data})
	}
	if len(foreign) > 0 {
		fields = append(fields, logField{attrKeyForeign, foreign})
	}
	return fields
}

func (this *telemetryHandler) prefixed(attrs []slog.Attr) []slog.Attr {
	if len(this.groups) == 0 {
		return attrs
	}
	prefix := strings.Join(this.groups, ".") + "."
	result := make([]slog.Attr, len(attrs))
	for i, attr := range attrs {
		result[i] = slog.Attr{Key: prefix + attr.Key, Value: attr.Value}
	}
	return result
}

func attrValue(value slog.Value) any {
	value = value.Resolve()
	if value.Kind() == slog.KindGroup {
		group := map[string]any{}
		for _, attr := range value.Group() {
			group[attr.Key] = attrValue(attr.Value)
		}
		return group
	}
	return jsonFriendly(value.Any())
}

// jsonFriendly turns errors into their message: encoding/json writes an error value as `{}`.
func jsonFriendly(value any) any {
	switch typed := value.(type) {
	case error:
		return typed.Error()
	case Attr:
		return jsonFriendlyMap(typed)
	case map[string]any:
		return jsonFriendlyMap(typed)
	}
	return value
}

func jsonFriendlyMap(source map[string]any) map[string]any {
	if source == nil {
		return nil
	}
	result := make(map[string]any, len(source))
	for key, value := range source {
		if err, ok := value.(error); ok {
			result[key] = err.Error()
			continue
		}
		result[key] = value
	}
	return result
}

func encodeFields(fields []logField) []byte {
	buffer := bytes.Buffer{}
	buffer.WriteByte('{')
	for i, field := range fields {
		if i > 0 {
			buffer.WriteByte(',')
		}
		key, _ := json.Marshal(field.key)
		buffer.Write(key)
		buffer.WriteByte(':')
		buffer.Write(encodeValue(field.value))
	}
	buffer.WriteString("}\n")
	return buffer.Bytes()
}

func encodeValue(value any) []byte {
	encoded, err := json.Marshal(value)
	if err == nil {
		return encoded
	}
	fallback, _ := json.Marshal(fmt.Sprintf("unencodable log value: %v", err))
	return fallback
}
