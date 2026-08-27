package v1

import (
	"github.com/labstack/echo/v5"
	"go.uber.org/dig"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	it "github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/tax"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
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

// TaxCalculationRest exposes the tax engine over HTTP, for the frontend and for out-of-process
// consumers. In-process Go modules should depend on it.TaxCalculationAppService through their own
// interfaces/external port instead, which is what keeps a later split into separate processes a
// change to one file.
type TaxCalculationRest struct {
	calculationSvc it.TaxCalculationAppService
}

// Calculate prices a whole document.
//
// It is a POST despite reading rather than writing, because the request is a document — many lines,
// each with its own context — and that does not fit a query string. It has no business side effects
// whatsoever: no invoice, no posting, no change to tax master data, so calling it twice with the
// same input produces the same answer and changes nothing (AC-TAX-35). That is what lets a draft
// order recalculate on every edit.
func (this TaxCalculationRest) Calculate(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST calculate tax"); e != nil {
			err = e
		}
	}()

	return httpserver.ServeRequest2(
		echoCtx,
		this.calculationSvc.Calculate,
		CalculateTaxRequest.ToCommand,
		NewCalculateTaxResponse,
		httpserver.JsonOk,
	)
}

// Simulate runs the same pipeline and additionally returns the trace of how it got there.
//
// Separate from Calculate because the explanation is expensive to assemble and pointless on the hot
// path: an order being priced needs the number, a tax administrator debugging a rule needs the
// reasoning (BR-TAX-ESS-051).
func (this TaxCalculationRest) Simulate(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST simulate tax"); e != nil {
			err = e
		}
	}()

	return httpserver.ServeRequest2(
		echoCtx,
		this.calculationSvc.Simulate,
		SimulateTaxRequest.ToCommand,
		NewSimulateTaxResponse,
		httpserver.JsonOk,
	)
}

// ReverseFull negates an entire original charge from its frozen snapshot.
func (this TaxCalculationRest) ReverseFull(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST reverse tax in full"); e != nil {
			err = e
		}
	}()

	return httpserver.ServeRequest2(
		echoCtx,
		this.calculationSvc.ReverseFull,
		ReverseFullRequest.ToCommand,
		NewReverseTaxResponse,
		httpserver.JsonOk,
	)
}

// ReversePartial reverses part of an original charge.
func (this TaxCalculationRest) ReversePartial(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST reverse tax partially"); e != nil {
			err = e
		}
	}()

	return httpserver.ServeRequest2(
		echoCtx,
		this.calculationSvc.ReversePartial,
		ReversePartialRequest.ToCommand,
		NewReverseTaxResponse,
		httpserver.JsonOk,
	)
}
