package external

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	itUom "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/uom"
)

// UomExtService is Accounting's port onto Essential's UoM conversion capability.
//
// A fixed tax is quoted per unit — 4,000 VND per litre — and a transaction line may be priced in
// any unit of the same dimension. Turning the second into the first is what this port is for, and
// BR-TAX-ESS-SUP-014 is explicit that Accounting must not do the arithmetic itself: the conversion
// factors live in Essential, and a second implementation of them would eventually disagree with
// the first about what a litre is.
//
// The port is narrowed to the two questions a tax calculation asks: convert this quantity, and
// does this UoM exist. Aliasing the whole upstream contract would re-export every method added to
// it later.
type UomExtService interface {
	Convert(ctx corectx.Context, query ConvertQuantityQuery) (*ConvertQuantityResult, error)
	GetUom(ctx corectx.Context, query GetUomQuery) (*GetUomResult, error)
}

type ConvertQuantityQuery = itUom.ConvertQuantityQuery
type ConvertQuantityResult = itUom.ConvertQuantityResult
type GetUomQuery = itUom.GetUomQuery
type GetUomResult = itUom.GetUomResult
