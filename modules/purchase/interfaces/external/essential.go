// Package external declares Purchase's local ports onto the capabilities other modules offer, so
// that splitting a module into its own process changes only the binding in infra/external. They are
// narrowed local interfaces, not aliases of the upstream service: an alias would re-export every
// method the upstream adds later.
package external

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	itUom "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/uom"
)

// UomExtService is Purchase's port onto Essential's unit-of-measure capability. It exposes
// conversion but no factors: callers must use Convert rather than dividing factors themselves, or
// the two would give different answers to the same question.
type UomExtService interface {
	// GetUom resolves one unit, so a line can be checked against the product's inventory unit
	// before it is priced. HasData false means the unit does not exist.
	GetUom(ctx corectx.Context, query GetUomQuery) (*GetUomResult, error)

	// Convert re-expresses a quantity in another unit of the same category.
	Convert(ctx corectx.Context, query ConvertQuantityQuery) (*ConvertQuantityResult, error)
}

type GetUomQuery = itUom.GetUomQuery
type GetUomResult = itUom.GetUomResult
type ConvertQuantityQuery = itUom.ConvertQuantityQuery
type ConvertQuantityResult = itUom.ConvertQuantityResult
