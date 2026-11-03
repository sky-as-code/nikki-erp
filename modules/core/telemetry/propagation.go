package telemetry

import (
	"context"
	"sort"
	"strings"

	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/propagation"
)

const (
	BaggageMaxEntries = 16
	BaggageMaxBytes   = 2048

	baggageHeader = "baggage"
	// baggageMaxRawBytes bounds the work done on an inbound header before any member is kept.
	baggageMaxRawBytes = 8192
)

// NewPropagator returns the configured propagators. Baggage, when enabled, is always the
// allowlisting one: the SDK's own would carry every key through.
func NewPropagator(names []string, baggageAllowlist []string) propagation.TextMapPropagator {
	propagators := []propagation.TextMapPropagator{}
	for _, name := range names {
		switch strings.ToLower(name) {
		case "tracecontext":
			propagators = append(propagators, propagation.TraceContext{})
		case "baggage":
			propagators = append(propagators, NewAllowlistBaggage(baggageAllowlist))
		}
	}
	return propagation.NewCompositeTextMapPropagator(propagators...)
}

// AllowlistBaggage keeps only allowlisted keys, at most BaggageMaxEntries members and
// BaggageMaxBytes encoded. Inbound members are kept in header order; outbound in key order, as
// a context's baggage has no order of its own. Each member is parsed by the SDK's parser.
type AllowlistBaggage struct {
	allowed map[string]struct{}
}

func NewAllowlistBaggage(allowlist []string) AllowlistBaggage {
	allowed := make(map[string]struct{}, len(allowlist))
	for _, key := range allowlist {
		allowed[key] = struct{}{}
	}
	return AllowlistBaggage{allowed: allowed}
}

func (this AllowlistBaggage) Fields() []string {
	return []string{baggageHeader}
}

func (this AllowlistBaggage) Inject(ctx context.Context, carrier propagation.TextMapCarrier) {
	members := baggage.FromContext(ctx).Members()
	sort.Slice(members, func(i, j int) bool { return members[i].Key() < members[j].Key() })
	kept := this.limit(members)
	if len(kept) == 0 {
		return
	}
	encoded := make([]string, len(kept))
	for i, member := range kept {
		encoded[i] = member.String()
	}
	carrier.Set(baggageHeader, strings.Join(encoded, ","))
}

func (this AllowlistBaggage) Extract(ctx context.Context, carrier propagation.TextMapCarrier) context.Context {
	raw := readBaggageHeader(carrier)
	if raw == "" || len(raw) > baggageMaxRawBytes {
		return ctx
	}
	members := []baggage.Member{}
	for _, segment := range strings.Split(raw, ",") {
		if parsed, err := baggage.Parse(strings.TrimSpace(segment)); err == nil && parsed.Len() == 1 {
			members = append(members, parsed.Members()[0])
		}
	}
	kept := this.limit(members)
	if len(kept) == 0 {
		return ctx
	}
	bag, err := baggage.New(kept...)
	if err != nil {
		return ctx
	}
	return baggage.ContextWithBaggage(ctx, bag)
}

// limit drops a member that is not allowlisted, repeats a key, or would break a bound; the
// members after it are still considered, so the outcome depends only on the input order.
func (this AllowlistBaggage) limit(members []baggage.Member) []baggage.Member {
	kept := []baggage.Member{}
	seen := map[string]struct{}{}
	totalBytes := 0
	for _, member := range members {
		if _, ok := this.allowed[member.Key()]; !ok {
			continue
		}
		if _, dup := seen[member.Key()]; dup || len(kept) == BaggageMaxEntries {
			continue
		}
		size := len(member.String())
		if len(kept) > 0 {
			size++
		}
		if totalBytes+size > BaggageMaxBytes {
			continue
		}
		seen[member.Key()] = struct{}{}
		kept = append(kept, member)
		totalBytes += size
	}
	return kept
}

func readBaggageHeader(carrier propagation.TextMapCarrier) string {
	if multi, ok := carrier.(propagation.ValuesGetter); ok {
		return strings.Join(multi.Values(baggageHeader), ",")
	}
	return carrier.Get(baggageHeader)
}
