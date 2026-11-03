package telemetry

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const sampleTraceparent = "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"

func extractBaggage(t *testing.T, allowlist []string, header string) baggage.Baggage {
	t.Helper()
	carrier := propagation.HeaderCarrier(http.Header{})
	carrier.Set("baggage", header)
	ctx := NewAllowlistBaggage(allowlist).Extract(context.Background(), carrier)
	return baggage.FromContext(ctx)
}

func TestAllowlistBaggage_EmptyAllowlistDropsEverything(t *testing.T) {
	bag := extractBaggage(t, nil, "tenant=t1,feature=x")
	assert.Equal(t, 0, bag.Len())
}

func TestAllowlistBaggage_DropsKeysOutsideAllowlist(t *testing.T) {
	bag := extractBaggage(t, []string{"feature"}, "user_id=42,feature=beta,org_id=7")
	assert.Equal(t, 1, bag.Len())
	assert.Equal(t, "beta", bag.Member("feature").Value())
}

func TestAllowlistBaggage_SkipsMalformedMembersKeepsValid(t *testing.T) {
	bag := extractBaggage(t, []string{"feature", "bad key"}, "bad key=1,feature=beta,=x,;;")
	assert.Equal(t, 1, bag.Len())
	assert.Equal(t, "beta", bag.Member("feature").Value())
}

func TestAllowlistBaggage_KeepsAtMostSixteenInHeaderOrder(t *testing.T) {
	keys := []string{}
	members := []string{}
	for i := 0; i < 20; i++ {
		keys = append(keys, fmt.Sprintf("k%02d", i))
		members = append(members, fmt.Sprintf("k%02d=v", i))
	}
	bag := extractBaggage(t, keys, strings.Join(members, ","))
	assert.Equal(t, BaggageMaxEntries, bag.Len())
	assert.Equal(t, "v", bag.Member("k15").Value())
	assert.Empty(t, bag.Member("k16").Key())
}

func TestAllowlistBaggage_BoundsEncodedBytesDeterministically(t *testing.T) {
	big := strings.Repeat("a", 1500)
	header := "first=" + big + ",second=" + big + ",third=small"
	for i := 0; i < 5; i++ {
		bag := extractBaggage(t, []string{"first", "second", "third"}, header)
		assert.Equal(t, 2, bag.Len())
		assert.Equal(t, big, bag.Member("first").Value())
		assert.Empty(t, bag.Member("second").Key())
		assert.Equal(t, "small", bag.Member("third").Value())
	}
}

func TestAllowlistBaggage_InjectFiltersAndBounds(t *testing.T) {
	allowed, err := baggage.NewMember("feature", "beta")
	require.NoError(t, err)
	secret, err := baggage.NewMember("token", "abc")
	require.NoError(t, err)
	bag, err := baggage.New(allowed, secret)
	require.NoError(t, err)

	carrier := propagation.MapCarrier{}
	NewAllowlistBaggage([]string{"feature"}).Inject(baggage.ContextWithBaggage(context.Background(), bag), carrier)
	assert.Equal(t, "feature=beta", carrier.Get("baggage"))

	empty := propagation.MapCarrier{}
	NewAllowlistBaggage(nil).Inject(baggage.ContextWithBaggage(context.Background(), bag), empty)
	_, present := empty["baggage"]
	assert.False(t, present)
}

// A broken baggage header must not cost the valid traceparent (BR §3.2).
func TestPropagator_BrokenBaggageKeepsTraceparent(t *testing.T) {
	header := http.Header{}
	header.Set("traceparent", sampleTraceparent)
	header.Set("baggage", "%%%,===")
	ctx := NewPropagator([]string{"tracecontext", "baggage"}, []string{"feature"}).
		Extract(context.Background(), propagation.HeaderCarrier(header))

	spanCtx := trace.SpanContextFromContext(ctx)
	assert.True(t, spanCtx.IsValid())
	assert.Equal(t, "4bf92f3577b34da6a3ce929d0e0e4736", spanCtx.TraceID().String())
	assert.Equal(t, 0, baggage.FromContext(ctx).Len())
}

func TestPropagator_FieldsCoverW3cHeaders(t *testing.T) {
	fields := NewPropagator([]string{"tracecontext", "baggage"}, nil).Fields()
	assert.ElementsMatch(t, []string{"traceparent", "tracestate", "baggage"}, fields)
	assert.Empty(t, NewPropagator([]string{"none"}, nil).Fields())
}
