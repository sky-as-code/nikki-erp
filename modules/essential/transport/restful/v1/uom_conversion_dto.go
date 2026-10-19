package v1

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	itUom "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/uom"
)

// Quantities travel as JSON strings, not numbers. A JSON number is parsed as a float64 by
// most clients, which would lose the precision BR-UOM-ESS-018 requires before the value
// ever reaches this service. The request is {quantity, source_uom_id[, target_uom_id]}, read by
// the application actions.

type ConvertQuantityResponse struct {
	// Quantity carries the target UoM's rounding precision already applied.
	Quantity string `json:"quantity"`

	// ExactQuantity is the unrounded result, for callers chaining further calculations.
	ExactQuantity string `json:"exact_quantity"`
}

func NewConvertQuantityResponse(data itUom.ConvertQuantityResultData) ConvertQuantityResponse {
	return ConvertQuantityResponse{
		Quantity:      data.Quantity.String(),
		ExactQuantity: data.ExactQuantity.String(),
	}
}

type ToReferenceResponse struct {
	Quantity       string   `json:"quantity"`
	ExactQuantity  string   `json:"exact_quantity"`
	ReferenceUomId model.Id `json:"reference_uom_id"`
}

func NewToReferenceResponse(data itUom.ToReferenceResultData) ToReferenceResponse {
	return ToReferenceResponse{
		Quantity:       data.Quantity.String(),
		ExactQuantity:  data.ExactQuantity.String(),
		ReferenceUomId: data.ReferenceUomId,
	}
}

func NewConvertQuantityResponseAny(data itUom.ConvertQuantityResultData) any {
	return NewConvertQuantityResponse(data)
}

func NewToReferenceResponseAny(data itUom.ToReferenceResultData) any {
	return NewToReferenceResponse(data)
}
