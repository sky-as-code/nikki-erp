package dynamicresource

import (
	"context"
	"net/http"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.bryk.io/pkg/errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

func installSpanRecorder(t *testing.T) *tracetest.InMemoryExporter {
	t.Helper()
	exporter := tracetest.NewInMemoryExporter()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	otel.SetTracerProvider(provider)
	t.Cleanup(func() { _ = provider.Shutdown(context.Background()) })
	return exporter
}

// spanProbeDomainService answers Create with failErr, recording the span current below the
// application layer.
type spanProbeDomainService struct {
	*fakeDomainService
	failErr error
	seen    trace.SpanContext
}

func (this *spanProbeDomainService) Create(
	ctx corectx.Context, cmd CreateCommand, options ...CreateOptions,
) (*CreateResult, error) {
	this.seen = trace.SpanContextFromContext(ctx)
	if this.failErr != nil {
		return nil, this.failErr
	}
	return this.fakeDomainService.Create(ctx, cmd, options...)
}

func newSpanProbe(failErr error) (*spanProbeDomainService, CrudApplicationService) {
	domSvc := &spanProbeDomainService{fakeDomainService: newFakeDomainService(newPlainSchema()), failErr: failErr}
	return domSvc, NewDefaultApplicationService(NewAppServiceParam{DomainService: domSvc})
}

func withRequestSpan(ctx corectx.Context) trace.Span {
	spanCtx, span := otel.Tracer("test").Start(ctx.InnerContext(), "POST /cmp_plain_resource")
	corectx.ReplaceInner(ctx, spanCtx)
	return span
}

// BE-05 / AC-13: one INTERNAL span named {schema}.{action}, current for the layers below, and
// the request's span current again once the action returns.
func TestBuiltInActionOpensOneSpanNamedSchemaAction(t *testing.T) {
	exporter := installSpanRecorder(t)
	domSvc, appSvc := newSpanProbe(nil)
	ctx := ownerContext()
	request := withRequestSpan(ctx)

	_, err := appSvc.Create(ctx, dmodel.DynamicFields{"name": "x"})
	require.NoError(t, err)

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)
	assert.Equal(t, "cmp_plain_resource.create", spans[0].Name)
	assert.Equal(t, trace.SpanKindInternal, spans[0].SpanKind)
	assert.Equal(t, request.SpanContext().SpanID(), spans[0].Parent.SpanID())
	assert.Equal(t, spans[0].SpanContext.SpanID(), domSvc.seen.SpanID())
	assert.Equal(t, codes.Unset, spans[0].Status.Code)
	assert.Equal(t, request.SpanContext().SpanID(), trace.SpanContextFromContext(ctx).SpanID())
}

func TestClientErrorLeavesTheStatusUnset(t *testing.T) {
	exporter := installSpanRecorder(t)
	_, appSvc := newSpanProbe(nil)

	result, err := appSvc.Create(deniedContext(), dmodel.DynamicFields{"name": "x"})
	require.NoError(t, err)
	require.Positive(t, result.ClientErrors.Count())

	require.Len(t, exporter.GetSpans(), 1)
	assert.Equal(t, codes.Unset, exporter.GetSpans()[0].Status.Code)
}

func TestInternalErrorMarksTheSpan(t *testing.T) {
	exporter := installSpanRecorder(t)
	_, appSvc := newSpanProbe(errors.New("database is down"))

	_, err := appSvc.Create(ownerContext(), dmodel.DynamicFields{"name": "x"})
	require.Error(t, err)

	require.Len(t, exporter.GetSpans(), 1)
	assert.Equal(t, codes.Error, exporter.GetSpans()[0].Status.Code)
}

func TestCustomActionNameHasNoIds(t *testing.T) {
	assert.Equal(t, "suspend", customActionName(RouteDefinition{Path: ":id/suspend"}))
	assert.Equal(t, "lines.approve", customActionName(RouteDefinition{Path: "lines/:line_id/approve"}))
	assert.Equal(t, "generic", customActionName(RouteDefinition{ActionType: ActionTypeGeneric}))
}

func TestCustomRouteOpensTheActionSpan(t *testing.T) {
	exporter := installSpanRecorder(t)
	restEngine, _ := newRestEngine(newOrgSchema())
	restEngine.
		AddRoute(RouteDefinition{Path: ":id/suspend", ActionType: ActionTypeGeneric, HandlerFn: okHandler}).
		AddRoute(RouteDefinition{Path: ":id/resume", ActionType: ActionTypeGeneric,
			HandlerFn: func(echoCtx *echo.Context, _ map[string]any) error {
				return echoCtx.NoContent(http.StatusInternalServerError)
			}})

	serveWithContext(t, restEngine, memberContext(),
		jsonRequest(http.MethodPost, "/cmp_org_resource/rec_1/suspend?org_id=org_mine", `{}`))
	serveWithContext(t, restEngine, memberContext(),
		jsonRequest(http.MethodPost, "/cmp_org_resource/rec_1/resume?org_id=org_mine", `{}`))

	spans := exporter.GetSpans()
	require.Len(t, spans, 2)
	assert.Equal(t, "cmp_org_resource.suspend", spans[0].Name)
	assert.Equal(t, codes.Unset, spans[0].Status.Code)
	assert.Equal(t, "cmp_org_resource.resume", spans[1].Name)
	assert.Equal(t, codes.Error, spans[1].Status.Code)
}
