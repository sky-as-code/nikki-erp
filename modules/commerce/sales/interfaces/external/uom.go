package external

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	itUom "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/uom"
)

type (
	ConvertQuantityQuery  = itUom.ConvertQuantityQuery
	ConvertQuantityResult = itUom.ConvertQuantityResult
	GetUomQuery           = itUom.GetUomQuery
	GetUomResult          = itUom.GetUomResult
)

// UomConversionExtService is Essential's unit conversion, used by the promotion engine to count a
// line in its product's base unit and in a Buy X Get Y rule's qualification unit (BR-C07). Sales
// never converts units itself.
type UomConversionExtService interface {
	Convert(ctx corectx.Context, query ConvertQuantityQuery) (*ConvertQuantityResult, error)
	GetUom(ctx corectx.Context, query GetUomQuery) (*GetUomResult, error)
}
