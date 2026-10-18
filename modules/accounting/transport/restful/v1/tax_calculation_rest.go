package v1

import (
	"encoding/json"

	"github.com/labstack/echo/v5"
	"go.uber.org/dig"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	it "github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/tax"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

type taxCalculationRestParams struct {
	dig.In

	CalculationSvc it.TaxCalculationAppService
}

func NewTaxCalculationRest(params taxCalculationRestParams) *TaxCalculationRest {
	return &TaxCalculationRest{
		calculationSvc: params.CalculationSvc,
	}
}

// TaxCalculationRest serves the tax engine as four collection actions of accounting_tax
// (calculate, simulate, reverse_full, reverse_partial) for the frontend and out-of-process
// consumers. In-process Go modules depend on it.TaxCalculationAppService through their own
// interfaces/external port instead, which keeps a later process split to one file.
//
// The application service authorizes every call (read, simulate or override on accounting_tax in
// the request's org); this layer only turns the bound payload into the typed request.
type TaxCalculationRest struct {
	calculationSvc it.TaxCalculationAppService
}

// Calculate prices a whole document. It is a POST despite only reading, because the request is a
// multi-line document that does not fit a query string. It has no side effects at all.
func (this *TaxCalculationRest) Calculate(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "calculate tax", payload,
		func(ctx corectx.Context, params dmodel.DynamicFields) (*it.CalculateResult, error) {
			request := CalculateTaxRequest{}
			if cErrs := decodePayload(params, &request); cErrs != nil {
				return &it.CalculateResult{ClientErrors: *cErrs}, nil
			}
			return this.calculationSvc.Calculate(ctx, request.ToCommand())
		},
		func(data it.CalculationResult) any { return NewCalculateTaxResponse(data) })
}

// Simulate runs the same pipeline and additionally returns the trace. It is separate from Calculate
// because the explanation is expensive to assemble and pointless on the pricing hot path.
func (this *TaxCalculationRest) Simulate(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "simulate tax", payload,
		func(ctx corectx.Context, params dmodel.DynamicFields) (*it.SimulateResult, error) {
			request := SimulateTaxRequest{}
			if cErrs := decodePayload(params, &request); cErrs != nil {
				return &it.SimulateResult{ClientErrors: *cErrs}, nil
			}
			return this.calculationSvc.Simulate(ctx, request.ToCommand())
		},
		func(data it.SimulationResult) any { return NewSimulateTaxResponse(data) })
}

// ReverseFull negates an entire original charge from its frozen snapshot.
func (this *TaxCalculationRest) ReverseFull(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "reverse tax in full", payload,
		func(ctx corectx.Context, params dmodel.DynamicFields) (*it.ReverseResult, error) {
			request := ReverseFullRequest{}
			if cErrs := decodePayload(params, &request); cErrs != nil {
				return &it.ReverseResult{ClientErrors: *cErrs}, nil
			}
			return this.calculationSvc.ReverseFull(ctx, request.ToCommand())
		},
		reverseResponse)
}

// ReversePartial reverses part of an original charge.
func (this *TaxCalculationRest) ReversePartial(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "reverse tax partially", payload,
		func(ctx corectx.Context, params dmodel.DynamicFields) (*it.ReverseResult, error) {
			request := ReversePartialRequest{}
			if cErrs := decodePayload(params, &request); cErrs != nil {
				return &it.ReverseResult{ClientErrors: *cErrs}, nil
			}
			return this.calculationSvc.ReversePartial(ctx, request.ToCommand())
		},
		reverseResponse)
}

func reverseResponse(data it.ReversalResult) any {
	return NewReverseTaxResponse(data)
}

// decodePayload re-reads the bound payload into a typed request. Amounts travel as JSON strings
// (see CalculateTaxRequest), which survive the round trip exactly; a malformed one is the caller's
// mistake and answers 400, as the typed binder used to.
func decodePayload(params dmodel.DynamicFields, target any) *ft.ClientErrors {
	encoded, err := json.Marshal(params)
	if err == nil {
		err = json.Unmarshal(encoded, target)
	}
	if err == nil {
		return nil
	}
	cErrs := ft.ClientErrors{}
	cErrs.Append(*ft.NewValidationError("body", string(ft.ErrorKey("err_malformed_request")), err.Error()))
	return &cErrs
}
