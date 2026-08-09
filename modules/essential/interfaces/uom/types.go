// Package uom declares the UoM capability that Essential offers to other modules.
//
// CRUD on the UoM resources themselves goes through the dynamic resource engine and needs
// no types here. What lives in this package is the conversion capability of BR-UOM-ESS-013,
// which business modules (stock, purchase, sales) must consume instead of reimplementing.
//
// A consuming module never imports this package from its domain or application layer.
// It declares a local port in its own interfaces/external/ and binds it once in
// infra/external/index.go — see docs/01 "Microservice-ready Monolith".
package uom

import (
	"github.com/shopspring/decimal"

	"github.com/sky-as-code/nikki-erp/common/model"
)

// ConvertQuantityQuery asks for a quantity expressed in SourceUomId to be re-expressed in
// TargetUomId. Both UoMs must belong to the same UoM Category (BR-UOM-ESS-012).
type ConvertQuantityQuery struct {
	Quantity    decimal.Decimal
	SourceUomId model.Id
	TargetUomId model.Id
}

// ConvertQuantityResultData carries the converted quantity in both its exact and rounded
// forms. Rounding is applied only at the end, per BR-UOM-ESS-018.
type ConvertQuantityResultData struct {
	// Quantity is the result after the target UoM's rounding precision is applied. This is
	// the value to store on a document or display to a user.
	Quantity decimal.Decimal

	// ExactQuantity is the unrounded result, for callers that chain further calculations
	// and must not accumulate rounding error.
	ExactQuantity decimal.Decimal
}

// ToReferenceQuery converts a quantity into the Reference UoM of its own category, the
// normalized form business modules store stock in (BR-UOM-ESS-007).
type ToReferenceQuery struct {
	Quantity    decimal.Decimal
	SourceUomId model.Id
}

// ToReferenceResultData carries the normalized quantity and the UoM it is expressed in.
type ToReferenceResultData struct {
	Quantity       decimal.Decimal
	ExactQuantity  decimal.Decimal
	ReferenceUomId model.Id
}

// GetUomQuery fetches a single UoM, for consumers that must validate a UoM reference they
// hold. HasData false on the result means the UoM does not exist.
type GetUomQuery struct {
	Id model.Id
}

// GetUomResultData exposes the parts of a UoM a consuming module may legitimately depend
// on. It deliberately does not expose the whole record: a consumer that needs to convert
// must call Convert rather than reimplement the arithmetic (BR-UOM-ESS-023, AC-UOM-33).
type GetUomResultData struct {
	Id         model.Id
	Symbol     string
	CategoryId model.Id
	IsArchived bool
}
