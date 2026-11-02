package dynamicmodel

import (
	"maps"

	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// ComputedContextOrgId is the `${ctx.org_id}` a computed field may read: the organization the
// request acts in, stamped by the server once the caller's membership has been checked.
const ComputedContextOrgId = "org_id"

type actingOrgCtxKey struct{}

// WithActingOrg records the organization a request acts in. Every query the request runs then
// renders `${ctx.org_id}` with it, so a per-org computed field (an activation flag) answers for
// that org whatever the caller put in its own search context.
func WithActingOrg(ctx corectx.Context, orgId model.Id) {
	if ctx != nil {
		ctx.WithValue(actingOrgCtxKey{}, orgId)
	}
}

// ActingOrgOf returns the organization recorded by WithActingOrg, if any.
func ActingOrgOf(ctx corectx.Context) (model.Id, bool) {
	if ctx == nil {
		return "", false
	}
	orgId, ok := ctx.Value(actingOrgCtxKey{}).(model.Id)
	return orgId, ok && orgId != ""
}

// MergeComputedContext is explicit with the acting org laid over it: a client-sent `org_id` never
// wins over the one the server authorized. A read that acts in no org (the sign-in page, a locale
// lookup) binds an empty org, so a per-org value such as an activation reads false rather than
// failing the whole query.
func MergeComputedContext(ctx corectx.Context, explicit map[string]any) map[string]any {
	merged := make(map[string]any, len(explicit)+1)
	maps.Copy(merged, explicit)
	if orgId, ok := ActingOrgOf(ctx); ok {
		merged[ComputedContextOrgId] = string(orgId)
	} else if _, sent := merged[ComputedContextOrgId]; !sent {
		merged[ComputedContextOrgId] = ""
	}
	return merged
}
