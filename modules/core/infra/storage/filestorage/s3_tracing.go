package filestorage

import (
	"context"
	"strings"

	"github.com/aws/smithy-go/tracing"
	"github.com/aws/smithy-go/tracing/smithyoteltracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// s3OperationSpanPrefix marks the one CLIENT span the SDK opens per call ("S3.GetObject"). Its internal phase
// spans (serialize, sign, retry loop, attempt, HTTP request...) are folded into it to keep one span per call.
const s3OperationSpanPrefix = "S3."

// s3AllowedSpanAttributes is an allowlist, not a denylist: `exception.message` and `api.error_message` can
// quote the request URL, which carries the object key (a user's file name), so only fixed-shape values pass.
var s3AllowedSpanAttributes = map[attribute.Key]bool{
	"rpc.system": true, "rpc.service": true, "rpc.method": true,
	"error": true, "exception.type": true, "api.error_code": true, "api.error_fault": true,
	"aws.request_id": true, "operation.attempt": true,
	"http.method": true, "http.status_code": true, "net.protocol.version": true,
	"http.request_content_length": true, "http.response_content_length": true,
	"net.peer.name": true, "net.peer.host": true, "net.peer.port": true, "net.peer.addr": true,
}

// s3TracerProvider bridges the SDK's native tracing to OpenTelemetry. The SDK never injects a trace header
// into the S3 request (smithy tracing has no propagator), which is what we want: the storage endpoint is not
// a service of ours. Pass otel.GetTracerProvider(): the global delegates, so telemetry may be set up later.
func s3TracerProvider(provider oteltrace.TracerProvider) tracing.TracerProvider {
	return smithyoteltracing.Adapt(&s3FilteredTracerProvider{TracerProvider: provider})
}

type s3FilteredTracerProvider struct {
	oteltrace.TracerProvider
}

func (this *s3FilteredTracerProvider) Tracer(name string, opts ...oteltrace.TracerOption) oteltrace.Tracer {
	return &s3FilteredTracer{Tracer: this.TracerProvider.Tracer(name, opts...)}
}

type s3FilteredTracer struct {
	oteltrace.Tracer
}

func (this *s3FilteredTracer) Start(
	ctx context.Context, spanName string, opts ...oteltrace.SpanStartOption,
) (context.Context, oteltrace.Span) {
	if !strings.HasPrefix(spanName, s3OperationSpanPrefix) {
		folded := foldIntoOperation(ctx)
		return oteltrace.ContextWithSpan(ctx, folded), folded
	}
	ctx, span := this.Tracer.Start(ctx, spanName, opts...)
	filtered := &s3FilteredSpan{Span: span}
	return oteltrace.ContextWithSpan(ctx, filtered), filtered
}

type s3FilteredSpan struct {
	oteltrace.Span
}

func (this *s3FilteredSpan) SetAttributes(kv ...attribute.KeyValue) {
	this.Span.SetAttributes(allowedS3Attributes(kv)...)
}

func (this *s3FilteredSpan) AddEvent(name string, _ ...oteltrace.EventOption) {
	this.Span.AddEvent(name)
}

// foldIntoOperation stands in for an SDK phase span: it shares the operation span's context, so anything
// started under it stays a child of the operation, and forwards attributes (filtered) to the operation span.
func foldIntoOperation(ctx context.Context) oteltrace.Span {
	switch parent := oteltrace.SpanFromContext(ctx).(type) {
	case *s3FilteredSpan:
		return &s3FoldedSpan{operation: parent, Span: parent}
	case *s3FoldedSpan:
		return parent
	}
	parent := oteltrace.SpanContextFromContext(ctx)
	return oteltrace.SpanFromContext(oteltrace.ContextWithSpanContext(context.Background(), parent))
}

// s3FoldedSpan never ends, renames or sets status on the operation span; the SDK does that itself.
type s3FoldedSpan struct {
	oteltrace.Span
	operation *s3FilteredSpan
}

func (this *s3FoldedSpan) SetAttributes(kv ...attribute.KeyValue) {
	this.operation.SetAttributes(kv...)
}
func (this *s3FoldedSpan) End(...oteltrace.SpanEndOption)              {}
func (this *s3FoldedSpan) SetStatus(codes.Code, string)                {}
func (this *s3FoldedSpan) SetName(string)                              {}
func (this *s3FoldedSpan) AddEvent(string, ...oteltrace.EventOption)   {}
func (this *s3FoldedSpan) RecordError(error, ...oteltrace.EventOption) {}

func allowedS3Attributes(kv []attribute.KeyValue) []attribute.KeyValue {
	kept := make([]attribute.KeyValue, 0, len(kv))
	for _, attr := range kv {
		if s3AllowedSpanAttributes[attr.Key] {
			kept = append(kept, attr)
		}
	}
	return kept
}
