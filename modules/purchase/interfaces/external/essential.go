// Package external declares Purchase's local ports onto the capabilities other modules offer.
//
// Purchase code depends on these interfaces and never on another module's service directly, so
// that splitting a module into its own process changes only the binding in infra/external. See
// docs/wiki/01 "Microservice-ready Monolith".
//
// They are narrowed local interfaces rather than aliases of the upstream service. The difference
// matters: an alias would re-export every method the upstream adds later, so Purchase would
// silently gain the ability to reach into another module's data the day that module grew a method.
package external

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	itUom "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/uom"
)

// UomExtService is Purchase's port onto Essential's unit-of-measure capability.
//
// It exposes conversion but no arithmetic: a module that must convert calls Convert rather than
// reading factors and dividing them itself (BR-UOM-ESS-023, AC-UOM-33). Reimplementing the
// arithmetic here would mean two answers to the same question, and the one Purchase stored would
// be the one nobody could reproduce.
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
