package logging

import (
	"context"
	"sync"
)

const (
	FieldIdempotencyId = "idempotency_id"
	FieldTenantId      = "tenant_id"
	FieldOrgId         = "org_id"

	FieldServiceName = "service.name"
)

// contextFieldNames are the only keys a ContextFieldsFn may fill. Anything else it returns is
// dropped, so a hook cannot shadow a trace or resource field either.
var contextFieldNames = []string{FieldIdempotencyId, FieldTenantId, FieldOrgId}

var resourceFieldNames = []string{FieldServiceName}

// ContextFieldsFn reads operation fields (idempotency_id, tenant_id, org_id) from ctx. It exists
// because the packages that own those values import logging, so logging cannot import them.
type ContextFieldsFn func(ctx context.Context) map[string]any

var (
	fieldsMutex    sync.RWMutex
	contextHooks   []ContextFieldsFn
	resourceFields = map[string]string{}
)

func RegisterContextFields(hook ContextFieldsFn) {
	if hook == nil {
		return
	}
	fieldsMutex.Lock()
	defer fieldsMutex.Unlock()
	contextHooks = append(contextHooks, hook)
}

// SetResourceFields is called by telemetry once its resource is built; until then the
// resource fields are null.
func SetResourceFields(fields map[string]string) {
	fieldsMutex.Lock()
	defer fieldsMutex.Unlock()
	resourceFields = map[string]string{}
	for _, name := range resourceFieldNames {
		if value, ok := fields[name]; ok && value != "" {
			resourceFields[name] = value
		}
	}
}

func collectContextFields(ctx context.Context) map[string]any {
	result := map[string]any{}
	if ctx == nil {
		return result
	}
	fieldsMutex.RLock()
	hooks := contextHooks
	fieldsMutex.RUnlock()
	for _, hook := range hooks {
		fields := hook(ctx)
		for _, name := range contextFieldNames {
			if value, ok := fields[name]; ok && value != nil && value != "" {
				result[name] = value
			}
		}
	}
	return result
}

func snapshotResourceFields() map[string]string {
	fieldsMutex.RLock()
	defer fieldsMutex.RUnlock()
	return resourceFields
}

// HttpFields are the access-log fields. They reach the top level of a record only through
// WithHttpFields, never through caller data.
type HttpFields struct {
	Method     string
	Route      string
	StatusCode int
	DurationMs float64
}

type httpFieldsKey struct{}

func WithHttpFields(ctx context.Context, fields HttpFields) context.Context {
	return context.WithValue(ctx, httpFieldsKey{}, fields)
}

func HttpFieldsOf(ctx context.Context) (HttpFields, bool) {
	if ctx == nil {
		return HttpFields{}, false
	}
	fields, ok := ctx.Value(httpFieldsKey{}).(HttpFields)
	return fields, ok
}
