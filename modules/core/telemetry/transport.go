package telemetry

import (
	"net/http"
	"strings"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/propagation"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

const HeaderIdempotencyId = "X-Idempotency-ID"

var strippedHeaders = []string{"traceparent", "tracestate", "baggage", HeaderIdempotencyId}

// NewTransport instruments an outbound transport. Per hop it strips the propagation headers a
// request inherited (net/http copies headers onto redirects), opens a CLIENT span through
// otelhttp, and injects the span context and X-Idempotency-ID only when the hop's host is
// allowlisted, so a redirect to another host is judged again.
func NewTransport(base http.RoundTripper, allowedHosts []string, options ...otelhttp.Option) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}
	injecting := &allowlistInjectTransport{next: base, hosts: newHostMatcher(allowedHosts)}
	options = append(options, otelhttp.WithPropagators(propagation.NewCompositeTextMapPropagator()))
	return &stripTransport{next: otelhttp.NewTransport(injecting, options...)}
}

type stripTransport struct {
	next http.RoundTripper
}

func (this *stripTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	clone := request.Clone(request.Context())
	for _, name := range strippedHeaders {
		clone.Header.Del(name)
	}
	return this.next.RoundTrip(clone)
}

type allowlistInjectTransport struct {
	next  http.RoundTripper
	hosts hostMatcher
}

func (this *allowlistInjectTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	if !this.hosts.matches(request.URL) {
		return this.next.RoundTrip(request)
	}
	clone := request.Clone(request.Context())
	Propagator().Inject(clone.Context(), propagation.HeaderCarrier(clone.Header))
	if id, ok := corectx.IdempotencyIdOf(clone.Context()); ok {
		clone.Header.Set(HeaderIdempotencyId, id)
	}
	return this.next.RoundTrip(clone)
}

// hostMatcher entries are "host" (any port), "host:port" (exact) or "*.suffix" (subdomains only).
type hostMatcher struct {
	exact    map[string]struct{}
	suffixes []string
}

func newHostMatcher(entries []string) hostMatcher {
	matcher := hostMatcher{exact: map[string]struct{}{}}
	for _, entry := range entries {
		entry = strings.ToLower(strings.TrimSpace(entry))
		if strings.HasPrefix(entry, "*.") {
			matcher.suffixes = append(matcher.suffixes, entry[1:])
		} else if entry != "" {
			matcher.exact[entry] = struct{}{}
		}
	}
	return matcher
}

func (this hostMatcher) matches(target interface {
	Hostname() string
	Port() string
}) bool {
	hostname := strings.ToLower(target.Hostname())
	if hostname == "" {
		return false
	}
	if _, ok := this.exact[hostname]; ok {
		return true
	}
	if port := target.Port(); port != "" {
		if _, ok := this.exact[hostname+":"+port]; ok {
			return true
		}
	}
	for _, suffix := range this.suffixes {
		if strings.HasSuffix(hostname, suffix) {
			return true
		}
	}
	return false
}
