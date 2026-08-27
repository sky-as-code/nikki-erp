package v1

import (
	"github.com/shopspring/decimal"

	it "github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/tax"
)

// The reversal endpoints take the original snapshot from the caller rather than looking it up.
//
// Accounting stores no transaction and no reversal state of its own (BR-TAX-ESS-SUP-025): the
// snapshot and the running totals of prior refunds belong to the module that made the sale. The
// alternative would require Tax to track every downstream transaction, which is precisely the
// dependency the module is arranged to avoid.

type ReverseFullRequest struct {
	OriginalSnapshot TaxSnapshotRequest `json:"original_snapshot"`

	// TaxDate is the refund's own date. It does NOT re-resolve the rate: a full reversal negates
	// what was charged, whatever the rate happens to be today.
	TaxDate string `json:"tax_date"`

	Components []ReversalComponentRequest `json:"components"`
}

func (this ReverseFullRequest) ToCommand() it.FullReversalRequest {
	return it.FullReversalRequest{
		OriginalSnapshot: this.OriginalSnapshot.toSnapshot(),
		TaxDate:          this.TaxDate,
		Components:       toReversalComponents(this.Components),
	}
}

type ReversePartialRequest struct {
	OriginalSnapshot TaxSnapshotRequest `json:"original_snapshot"`
	TaxDate          string             `json:"tax_date"`

	RoundingPolicyCode string `json:"rounding_policy_code"`

	Components []ReversalComponentRequest `json:"components"`
}

func (this ReversePartialRequest) ToCommand() it.PartialReversalRequest {
	return it.PartialReversalRequest{
		OriginalSnapshot:   this.OriginalSnapshot.toSnapshot(),
		TaxDate:            this.TaxDate,
		RoundingPolicyCode: this.RoundingPolicyCode,
		Components:         toReversalComponents(this.Components),
	}
}

// TaxSnapshotRequest is the caller handing back the snapshot it stored at the time of sale.
//
// Only the fields a reversal actually reads are bound. The rest of the stored snapshot is the
// caller's record and Tax has no use for it here.
type TaxSnapshotRequest struct {
	SchemaVersion string `json:"schema_version"`
	CurrencyCode  string `json:"currency_code"`

	RoundingPolicyId      string          `json:"rounding_policy_id"`
	RoundingPolicyVersion int32           `json:"rounding_policy_version"`
	RoundingScope         string          `json:"rounding_scope"`
	RoundingMethod        string          `json:"rounding_method"`
	RoundingIncrement     decimal.Decimal `json:"rounding_increment"`
}

func (this TaxSnapshotRequest) toSnapshot() it.Snapshot {
	return it.Snapshot{
		SchemaVersion:         this.SchemaVersion,
		CurrencyCode:          this.CurrencyCode,
		RoundingPolicyId:      this.RoundingPolicyId,
		RoundingPolicyVersion: this.RoundingPolicyVersion,
		RoundingScope:         roundingScopeOf(this.RoundingScope),
		RoundingMethod:        roundingMethodOf(this.RoundingMethod),
		RoundingIncrement:     this.RoundingIncrement,
	}
}

// ReversalComponentRequest is the caller's reversal state for one original component.
//
// The amounts are strings for the same reason every other money field is: a float64 cannot hold
// them exactly, and a reversal that is off by a fraction will not close to zero against the
// original charge.
type ReversalComponentRequest struct {
	OriginalComponentReference string `json:"original_component_reference"`
	OriginalReversibleBasis    string `json:"original_reversible_basis"`
	OriginalTaxAmount          string `json:"original_tax_amount"`
	AlreadyReversedBasis       string `json:"already_reversed_basis"`
	AlreadyReversedTaxAmount   string `json:"already_reversed_tax_amount"`
	RequestedReversalBasis     string `json:"requested_reversal_basis"`

	// IsFinalReversal marks the last refund against this component. It absorbs any rounding
	// residual, which is what makes a sequence of partial refunds sum to exactly the original
	// charge rather than approximately (BR-TAX-ESS-033).
	IsFinalReversal bool `json:"is_final_reversal"`
}

func toReversalComponents(requested []ReversalComponentRequest) []it.ReversalComponentRequest {
	components := make([]it.ReversalComponentRequest, 0, len(requested))
	for _, component := range requested {
		components = append(components, it.ReversalComponentRequest{
			OriginalComponentReference: component.OriginalComponentReference,
			OriginalReversibleBasis:    component.OriginalReversibleBasis,
			OriginalTaxAmount:          component.OriginalTaxAmount,
			AlreadyReversedBasis:       component.AlreadyReversedBasis,
			AlreadyReversedTaxAmount:   component.AlreadyReversedTaxAmount,
			RequestedReversalBasis:     component.RequestedReversalBasis,
			IsFinalReversal:            component.IsFinalReversal,
		})
	}
	return components
}

type ReverseTaxResponse struct {
	Status string `json:"status"`

	// TotalReversedTax is negative, so a caller summing charges and refunds gets the net.
	TotalReversedTax string `json:"total_reversed_tax"`

	Components []ReversalComponentResponse `json:"components"`

	// Snapshot is the refund's own frozen record, referencing the original by the logical identity
	// the caller supplied (BR-TAX-ESS-055).
	Snapshot TaxSnapshotResponse `json:"snapshot"`
}

type ReversalComponentResponse struct {
	OriginalComponentReference string `json:"original_component_reference"`
	ReversalTaxAmount          string `json:"reversal_tax_amount"`
	RemainingTaxAmount         string `json:"remaining_tax_amount"`
}

func NewReverseTaxResponse(data it.ReversalResult) ReverseTaxResponse {
	components := make([]ReversalComponentResponse, 0, len(data.Components))
	for _, component := range data.Components {
		components = append(components, ReversalComponentResponse{
			OriginalComponentReference: component.OriginalComponentReference,
			ReversalTaxAmount:          component.ReversalTaxAmount,
			RemainingTaxAmount:         component.RemainingTaxAmount,
		})
	}
	return ReverseTaxResponse{
		Status:           string(data.Status),
		TotalReversedTax: data.TotalReversedTax,
		Components:       components,
		Snapshot:         newTaxSnapshotResponse(data.Snapshot),
	}
}
