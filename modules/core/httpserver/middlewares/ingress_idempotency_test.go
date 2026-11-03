package middlewares

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

const callerIdempotencyId = "019a46c2-2a40-7b35-9e0f-8e2fd22b9b6c"

// ID-AC-06: a direct call without the header gets one fresh ID, shared by context, response and span.
func TestIdempotency_MissingHeaderMintsOne(t *testing.T) {
	fixture := newIngressFixture(t, sdktrace.AlwaysSample())
	recorder := fixture.serve("/items/1", nil)

	require.Equal(t, http.StatusOK, recorder.Code)
	id := recorder.Header().Get(HeaderIdempotencyId)
	parsed, err := corectx.ParseIdempotencyId(id)
	require.NoError(t, err)
	assert.Equal(t, id, parsed)
	assert.Equal(t, id, fixture.seen.idempotencyId)
	attr, ok := stubAttribute(fixture.onlySpan(t).Attributes, attrIdempotencyId)
	assert.True(t, ok)
	assert.Equal(t, id, attr.AsString())
}

// ID-AC-05 (backend side): an ID the BFF already assigned is kept, not re-minted.
func TestIdempotency_ValidHeaderIsKept(t *testing.T) {
	fixture := newIngressFixture(t, sdktrace.AlwaysSample())
	recorder := fixture.serve("/items/1", http.Header{HeaderIdempotencyId: {callerIdempotencyId}})

	assert.Equal(t, callerIdempotencyId, recorder.Header().Get(HeaderIdempotencyId))
	assert.Equal(t, callerIdempotencyId, fixture.seen.idempotencyId)
}

// ID-AC-07: uppercase hex is the same UUID, normalised to lowercase.
func TestIdempotency_UppercaseIsNormalised(t *testing.T) {
	fixture := newIngressFixture(t, sdktrace.AlwaysSample())
	upper := "019A46C2-2A40-7B35-9E0F-8E2FD22B9B6C"
	recorder := fixture.serve("/items/1", http.Header{"x-idempotency-id": {upper}})

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, callerIdempotencyId, recorder.Header().Get(HeaderIdempotencyId))
	assert.Equal(t, callerIdempotencyId, fixture.seen.idempotencyId)
}

// ID-AC-08: empty, wrong version, wrong format or ambiguous headers are a 400 before the handler,
// with no echo of the value in the response or the log.
func TestIdempotency_InvalidHeaderIsRejected(t *testing.T) {
	cases := map[string][]string{
		"empty":        {""},
		"uuid v4":      {"9b2f1e64-3c1d-4b7a-9f0e-8e2fd22b9b6c"},
		"bad format":   {"not-a-uuid-at-all-not-a-uuid-at-all!"},
		"braces":       {"{" + callerIdempotencyId + "}"},
		"duplicate":    {callerIdempotencyId, callerIdempotencyId},
		"comma list":   {callerIdempotencyId + "," + callerIdempotencyId},
		"nil uuid":     {"00000000-0000-0000-0000-000000000000"},
		"spaced comma": {callerIdempotencyId + ", " + callerIdempotencyId},
		"urn":          {"urn:uuid:" + callerIdempotencyId},
	}
	for name, values := range cases {
		t.Run(name, func(t *testing.T) {
			fixture := newIngressFixture(t, sdktrace.AlwaysSample())
			recorder := fixture.serve("/items/1", http.Header{HeaderIdempotencyId: values})

			assert.Equal(t, http.StatusBadRequest, recorder.Code)
			assert.Equal(t, 0, fixture.seen.calls)
			assert.Empty(t, recorder.Header().Values(HeaderIdempotencyId))
			assert.Regexp(t, traceIdPattern, recorder.Header().Get(HeaderTraceId))
			assertInvalidIdempotencyBody(t, recorder.Body.Bytes())
			assertNoRawValueLogged(t, fixture.logger, values)
		})
	}
}

func assertInvalidIdempotencyBody(t *testing.T, body []byte) {
	t.Helper()
	items := []map[string]any{}
	require.NoError(t, json.Unmarshal(body, &items))
	require.Len(t, items, 1)
	assert.Equal(t, "X-Idempotency-ID", items[0]["field"])
	assert.Equal(t, "common:err_invalid_idempotency_id", items[0]["key"])
	assert.Equal(t, "validation", items[0]["type"])
}

func assertNoRawValueLogged(t *testing.T, logger *recordingLogger, values []string) {
	t.Helper()
	rejections := logger.find("rejected X-Idempotency-ID header")
	require.Len(t, rejections, 1)
	for _, entry := range logger.entries {
		for _, value := range values {
			if value == "" {
				continue
			}
			assert.NotContains(t, fmt.Sprint(entry.data), value)
		}
	}
	access := logger.find(accessLogMessage)
	require.Len(t, access, 1)
	assert.Equal(t, http.StatusBadRequest, access[0].data["http.response.status_code"])
}

// ID-AC-14: two calls without the header are two requests, not a guessed retry.
func TestIdempotency_TwoMissingHeadersGetTwoIds(t *testing.T) {
	fixture := newIngressFixture(t, sdktrace.AlwaysSample())
	first := fixture.serve("/items/1", nil).Header().Get(HeaderIdempotencyId)
	second := fixture.serve("/items/1", nil).Header().Get(HeaderIdempotencyId)

	assert.NotEmpty(t, first)
	assert.NotEmpty(t, second)
	assert.NotEqual(t, first, second)
}
