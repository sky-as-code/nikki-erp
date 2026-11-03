package filestorage

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	oteltrace "go.opentelemetry.io/otel/trace"
)

const s3TestObjectKey = "org/01H/payroll-of-jane-doe.pdf"

type s3TestHeaders struct {
	seen []http.Header
}

func newS3TestClient(t *testing.T, provider oteltrace.TracerProvider, status int) (*s3.Client, *s3TestHeaders) {
	headers := &s3TestHeaders{}
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		headers.seen = append(headers.seen, req.Header.Clone())
		writer.Header().Set("x-amz-request-id", "REQ123")
		writer.WriteHeader(status)
		if status >= 400 {
			_, _ = writer.Write([]byte(`<Error><Code>AccessDenied</Code><Message>denied ` +
				s3TestObjectKey + `</Message></Error>`))
		}
	}))
	t.Cleanup(server.Close)

	client := s3.New(s3.Options{
		Region:           "us-east-1",
		BaseEndpoint:     aws.String(server.URL),
		UsePathStyle:     true,
		Credentials:      aws.AnonymousCredentials{},
		RetryMaxAttempts: 1,
		TracerProvider:   s3TracerProvider(provider),
	})
	return client, headers
}

func runTracedS3Call(t *testing.T, status int) ([]sdktrace.ReadOnlySpan, *s3TestHeaders, error) {
	recorder := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	client, headers := newS3TestClient(t, provider, status)

	ctx, request := provider.Tracer("test").Start(context.Background(), "request")
	_, err := client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String("files"), Key: aws.String(s3TestObjectKey),
	})
	request.End()
	return recorder.Ended(), headers, err
}

func TestS3OneClientSpanPerOperationUnderTheRequestSpan(t *testing.T) {
	spans, _, err := runTracedS3Call(t, http.StatusNoContent)
	require.NoError(t, err)

	require.Len(t, spans, 2, "SDK phase spans must be folded into the operation span")
	operation, request := spans[0], spans[1]
	assert.Equal(t, "S3.DeleteObject", operation.Name())
	assert.Equal(t, oteltrace.SpanKindClient, operation.SpanKind())
	assert.Equal(t, request.SpanContext().SpanID(), operation.Parent().SpanID())
	assert.Equal(t, request.SpanContext().TraceID(), operation.SpanContext().TraceID())
}

func TestS3SpanCarriesNoObjectKeyEvenOnError(t *testing.T) {
	spans, _, err := runTracedS3Call(t, http.StatusForbidden)
	require.Error(t, err)

	attrs := map[string]string{}
	for _, span := range spans {
		for _, attr := range span.Attributes() {
			attrs[string(attr.Key)] = attr.Value.Emit()
		}
		for _, event := range span.Events() {
			assert.NotContains(t, event.Name, "jane-doe")
		}
	}
	assert.Equal(t, "DeleteObject", attrs["rpc.method"])
	assert.Equal(t, "AccessDenied", attrs["api.error_code"])
	assert.NotContains(t, attrs, "exception.message")
	assert.NotContains(t, attrs, "api.error_message")
	for key, value := range attrs {
		assert.NotContains(t, value, "jane-doe", "attribute %s carries the object key", key)
	}
}

func TestS3RequestCarriesNoTraceOrIdempotencyHeader(t *testing.T) {
	_, headers, err := runTracedS3Call(t, http.StatusNoContent)
	require.NoError(t, err)

	require.NotEmpty(t, headers.seen)
	for _, header := range headers.seen {
		for name := range header {
			lower := strings.ToLower(name)
			assert.NotContains(t, []string{"traceparent", "tracestate", "baggage", "x-idempotency-id"}, lower)
		}
	}
}
