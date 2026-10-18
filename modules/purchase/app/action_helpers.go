// Package app holds the Purchase application services. Authorization happens here and nowhere
// else: every built-in and custom action asserts its permission before delegating to a domain
// service, and domain services never authorize nor import this package.
package app

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// Param names the custom actions read from the request.
const (
	paramRecordId = "id"
	paramReason   = "reason"
	// paramAlternativeChoice carries the answer to the open-alternatives warning.
	paramAlternativeChoice = "alternative_choice"
	paramAlternativeVendor = "vendor_id"
	paramMergeOrderIds     = "order_ids"
)

// actionResult is the generic result every custom action answers with. The data is passed through
// as the response body, which is what the legacy engine actions returned.
type actionResult = dyn.OpResult[any]

// assertRecordAction asserts the permission in the caller's org, then that the record the path
// names belongs to that org.
func assertRecordAction(
	svc composable.CrudApplicationService, ctx corectx.Context, permission string, cmd dmodel.DynamicFields,
) (*ft.ClientErrors, error) {
	orgId, cErrs := svc.AssertAction(ctx, permission, cmd)
	if cErrs != nil || orgId == nil {
		return cErrs, nil
	}
	return svc.AssertRecordInOrg(ctx, cmd, *orgId)
}

func actionFailure(cErrs *ft.ClientErrors, err error) (*actionResult, error) {
	if err != nil {
		return nil, err
	}
	return &actionResult{ClientErrors: *cErrs}, nil
}

// toActionResult widens a typed result into the generic one. ClientErrors must survive: a refused
// operation reports its reason through them, not through err.
func toActionResult[TData any](result *dyn.OpResult[TData], err error) (*actionResult, error) {
	if err != nil {
		return nil, err
	}
	out := &actionResult{ClientErrors: result.ClientErrors, HasData: result.HasData}
	if result.HasData {
		out.Data = result.Data
	}
	return out, nil
}

func readStringParam(params dmodel.DynamicFields, field string) string {
	value, ok := params[field]
	if !ok || value == nil {
		return ""
	}
	if typed, ok := value.(string); ok {
		return typed
	}
	return ""
}

// readStringList reads a list of ids from the request body. A malformed list reads as empty rather
// than partially, so the merge refuses instead of silently merging a subset.
func readStringList(params dmodel.DynamicFields, field string) []string {
	raw, ok := params[field]
	if !ok || raw == nil {
		return nil
	}
	switch typed := raw.(type) {
	case []string:
		return typed
	case []any:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			if text, ok := item.(string); ok {
				out = append(out, text)
			}
		}
		return out
	default:
		return nil
	}
}
