package context

import (
	"context"

	"github.com/sky-as-code/nikki-erp/common/model"
)

// RequestSource says what started the execution: an HTTP request, a scheduled job, a message
// consumer, or the platform itself.
type RequestSource string

const (
	RequestSourceHttp   = RequestSource("http")
	RequestSourceJob    = RequestSource("job")
	RequestSourceEvent  = RequestSource("event")
	RequestSourceSystem = RequestSource("system")
)

// RequestMeta identifies one execution, so that the several records one call writes (an audit
// entry per resource it touched) can be correlated.
type RequestMeta struct {
	Id     string
	Source RequestSource
}

var ctxKeyRequestMeta = contextKey{"request_meta"}

// WithRequestMeta derives a plain context carrying meta, for an entry point that has no
// RequestContext yet: a corectx built on top of it sees the same meta.
func WithRequestMeta(ctx context.Context, meta RequestMeta) context.Context {
	return context.WithValue(ctx, ctxKeyRequestMeta, meta)
}

// SetRequestMeta stores meta on ctx in place. It lives in the inner context's values, so
// CloneRequestContext and every context derived from the inner one carry it along.
func SetRequestMeta(ctx Context, meta RequestMeta) {
	ctx.WithValue(ctxKeyRequestMeta, meta)
}

// RequestMetaOf returns the meta ctx carries, and false when none was set.
func RequestMetaOf(ctx context.Context) (RequestMeta, bool) {
	if ctx == nil {
		return RequestMeta{}, false
	}
	meta, ok := ctx.Value(ctxKeyRequestMeta).(RequestMeta)
	return meta, ok && meta.Id != ""
}

// NewRequestMeta mints a meta with a fresh id.
func NewRequestMeta(source RequestSource) RequestMeta {
	id, err := model.NewId()
	if err != nil || id == nil {
		return RequestMeta{Source: source}
	}
	return RequestMeta{Id: string(*id), Source: source}
}

// EnsureRequestMeta returns the meta ctx carries, or mints one with fallbackSource and stores it
// on ctx, so later writes through the same ctx share it. An entry point should set the meta
// itself: a meta minted here does not reach a context cloned from ctx before this call.
func EnsureRequestMeta(ctx Context, fallbackSource RequestSource) RequestMeta {
	if meta, ok := RequestMetaOf(ctx); ok {
		return meta
	}
	meta := NewRequestMeta(fallbackSource)
	if meta.Id != "" {
		SetRequestMeta(ctx, meta)
	}
	return meta
}

var ctxKeyRequestAction = contextKey{"request_action"}

// SetRequestAction labels the writes ctx makes from now on with a custom action code, such as
// "suspend": the audit trail records an update made under it as that action. The request guard
// sets it when a custom permission check passes, so a custom action needs no audit code of its
// own. Stored in place, like SetRequestMeta.
func SetRequestAction(ctx Context, actionCode string) {
	ctx.WithValue(ctxKeyRequestAction, actionCode)
}

// RequestActionOf returns the custom action code ctx was labelled with, or "".
func RequestActionOf(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	code, _ := ctx.Value(ctxKeyRequestAction).(string)
	return code
}
