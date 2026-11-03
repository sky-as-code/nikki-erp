package context

import (
	"context"
)

// ReplaceInner swaps ctx's inner context in place and reports whether it could. It is for a
// layer that scopes a value to the call it wraps, such as the current span, on a context its
// callee mutates in place (acting org, request action): deriving a clone instead would hide
// those mutations from the caller. The owning-goroutine rule of the setters applies.
func ReplaceInner(ctx Context, inner context.Context) bool {
	reqCtx, isReqCtx := ctx.(*RequestContext)
	if !isReqCtx || inner == nil {
		return false
	}
	reqCtx.setInner(inner)
	return true
}
