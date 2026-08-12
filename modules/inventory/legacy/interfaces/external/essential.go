package external

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	itUom "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/uom"
)

// UomExtService is this module's local port onto Essential's UoM capability. Product code
// depends on this interface, never on the Essential service directly, so that splitting
// Essential into its own process only changes the binding in infra/external.
//
// It exposes the narrowest surface product actually needs. In particular it does not expose
// conversion arithmetic: a module that must convert calls Convert rather than reading
// factors and dividing them itself (BR-UOM-ESS-023, AC-UOM-33).
type UomExtService interface {
	GetUom(ctx corectx.Context, query GetUomQuery) (*GetUomResult, error)
}

type GetUomQuery = itUom.GetUomQuery
type GetUomResult = itUom.GetUomResult
