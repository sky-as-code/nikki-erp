// Package external declares the ports Inventory consumes from other modules. A port names only
// what Inventory needs, so a module split into its own process rebinds infra/external and nothing
// else.
package external

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	itUom "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/uom"
)

type (
	ConvertQuantityQuery  = itUom.ConvertQuantityQuery
	ConvertQuantityResult = itUom.ConvertQuantityResult
)

// UomConversionExtService re-expresses a quantity in another unit of the same category. Stock
// keeps every quantity in the variant's base unit; a caller reserving in its own unit is converted
// through this before anything is compared with a balance.
type UomConversionExtService interface {
	Convert(ctx corectx.Context, query ConvertQuantityQuery) (*ConvertQuantityResult, error)
}
