package app

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

const sampleTraceId = "4bf92f3577b34da6a3ce929d0e0e4736"

func contextWithSpan(t *testing.T) context.Context {
	traceId, err := trace.TraceIDFromHex(sampleTraceId)
	require.NoError(t, err)
	spanId, err := trace.SpanIDFromHex("00f067aa0ba902b7")
	require.NoError(t, err)
	spanCtx := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: traceId, SpanID: spanId, TraceFlags: trace.FlagsSampled,
	})
	return trace.ContextWithSpanContext(context.Background(), spanCtx)
}

func TestActiveTraceIdReadsTheActiveSpan(t *testing.T) {
	ctx := corectx.NewRequestContext(contextWithSpan(t))

	traceId := activeTraceId(ctx)
	row, err := buildActionRow(itAuditEntry(), actorSnapshot{kind: "system"}, metaWithoutSource(), traceId)

	require.NoError(t, err)
	assert.Equal(t, sampleTraceId, row["trace_id"])
}

func TestActiveTraceIdIsEmptyWithoutASpan(t *testing.T) {
	ctx := corectx.NewRequestContext(context.Background())

	assert.Equal(t, "", activeTraceId(ctx))
}
