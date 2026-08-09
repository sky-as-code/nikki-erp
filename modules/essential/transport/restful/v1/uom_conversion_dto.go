package v1

import (
	"github.com/shopspring/decimal"

	"github.com/sky-as-code/nikki-erp/common/model"
	itUom "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/uom"
)

// Quantities travel as JSON strings, not numbers. A JSON number is parsed as a float64 by
// most clients, which would lose the precision BR-UOM-ESS-018 requires before the value
// ever reaches this service.

type ConvertQuantityRequest struct {
	Quantity    decimal.Decimal `json:"quantity"`
	SourceUomId model.Id        `json:"source_uom_id"`
	TargetUomId model.Id        `json:"target_uom_id"`
}

func (this ConvertQuantityRequest) ToQuery() itUom.ConvertQuantityQuery {
	return itUom.ConvertQuantityQuery{
		Quantity:    this.Quantity,
		SourceUomId: this.SourceUomId,
		TargetUomId: this.TargetUomId,
	}
}

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

type ToReferenceRequest struct {
	Quantity    decimal.Decimal `json:"quantity"`
	SourceUomId model.Id        `json:"source_uom_id"`
}

func (this ToReferenceRequest) ToQuery() itUom.ToReferenceQuery {
	return itUom.ToReferenceQuery{
		Quantity:    this.Quantity,
		SourceUomId: this.SourceUomId,
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
