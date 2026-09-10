package composable

import (
	"github.com/labstack/echo/v5"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
)

// ServeAction is the body of a custom route's handler: it hands the bound payload to an
// application-service method and answers the way the built-in handlers do. A client error
// answers 400, a result without data answers 400 with a not-found payload, and a Go error
// bubbles up as 500.
//
//	func (this *WarehouseRest) Suspend(echoCtx *echo.Context, payload map[string]any) error {
//		return composable.ServeAction(echoCtx, "suspend warehouse", payload, this.warehouseSvc.Suspend, composable.MutateResponse)
//	}
func ServeAction[TData any](
	echoCtx *echo.Context,
	action string,
	payload map[string]any,
	call func(ctx corectx.Context, params dmodel.DynamicFields) (*dyn.OpResult[TData], error),
	shape func(data TData) any,
) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST "+action); e != nil {
			err = e
		}
	}()

	reqCtx, err := corectx.AsRequestContext(echoCtx)
	if err != nil {
		return err
	}
	if payload == nil {
		payload = map[string]any{}
	}

	result, err := call(reqCtx, payload)
	if err != nil {
		return err
	}
	if result.ClientErrors != nil && result.ClientErrors.Count() > 0 {
		return httpserver.JsonBadRequest(echoCtx, result.ClientErrors)
	}
	if !result.HasData {
		return httpserver.JsonBadRequest(echoCtx, ft.ClientErrors{*ft.NewAnonymousNotFoundError()})
	}
	return httpserver.JsonOk(echoCtx, shape(result.Data))
}

// Identity passes the result data through as the response body.
func Identity[TData any](data TData) any {
	return data
}

// MutateResponse shapes a mutation result the way the built-in update and delete answer.
func MutateResponse(mutation dyn.MutateResultData) any {
	return httpserver.NewRestMutateResponse(mutation)
}
