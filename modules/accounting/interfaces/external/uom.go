package external

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	itUom "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/uom"
)

// UomExtService is Accounting's port onto Essential's UoM conversion capability. A fixed tax is
// quoted per unit while a line may be priced in any unit of the same dimension, so quantities must
// be converted here. Accounting must not do the arithmetic itself: the conversion factors live in
// Essential, and a second implementation would eventually disagree with it.
type UomExtService interface {
	Convert(ctx corectx.Context, query ConvertQuantityQuery) (*ConvertQuantityResult, error)
	GetUom(ctx corectx.Context, query GetUomQuery) (*GetUomResult, error)
}

type ConvertQuantityQuery = itUom.ConvertQuantityQuery
type ConvertQuantityResult = itUom.ConvertQuantityResult
type GetUomQuery = itUom.GetUomQuery
type GetUomResult = itUom.GetUomResult
