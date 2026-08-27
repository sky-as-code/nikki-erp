// Package tax is the port other modules consume to have tax determined and calculated.
//
// Everything here is a plain value type. A consuming module builds a request, receives a result and
// a snapshot, and stores the snapshot itself — Accounting keeps no record of the transaction and
// holds no foreign key into it (BR-TAX-ESS-030, TAX-SUP-INV-01). That is what makes it possible to
// split this module into its own process later without any caller changing.
package tax

import (
	"github.com/shopspring/decimal"

	"github.com/sky-as-code/nikki-erp/modules/accounting/domain/models"
)

// CalculationRequest is one document's worth of tax to compute.
//
// Document-level rather than line-level because rounding may be document-scoped, and a per-line API
// could only fake that by rounding each line and summing — which is a different number, and not the
// one the law asks for (BR-TAX-ESS-022).
type CalculationRequest struct {
	OrgId string

	// OperationType is sale or sale_refund. The purchase values exist in the enum as a reserved
	// contract and are rejected here: V1 has no purchase-side semantics defined.
	OperationType models.OperationType

	// TaxDate decides which configuration is in force, and is MANDATORY. It is never defaulted from
	// the server clock: BR-TAX-ESS-SUP-020 deleted that fallback, because a request that forgot the
	// date would otherwise be priced against whatever happened to be effective at the moment it was
	// processed rather than the date the sale legally occurred.
	TaxDate string

	CurrencyCode string

	Seller TaxPartyContext
	Buyer  TaxPartyContext

	ShipFromJurisdictionId string
	ShipToJurisdictionId   string

	// PriceMode is the document default a tax whose inclusion mode is "inherit" resolves against.
	PriceMode models.PriceInclusionMode

	// BusinessChannelCode and OutletReference are opaque context. Tax carries them for rule
	// conditions and audit, and never resolves them against Sales (BR-TAX-ESS-025).
	BusinessChannelCode string
	OutletReference     string

	// RoundingPolicyCode names the policy to apply. Empty falls back to the organization's default
	// setting, and when neither names one the result is unresolved rather than a guessed scale.
	RoundingPolicyCode string

	Lines []CalculationLine
}

// TaxPartyContext is the normalized tax-relevant facts about one side of a transaction.
//
// Deliberately a fixed, minimal schema rather than a free-form profile (BR-TAX-ESS-SUP-022). The
// PartyReference is opaque and exists for tracing only: Tax must not query Contacts with it, or the
// dependency would run the wrong way.
type TaxPartyContext struct {
	PartyReference         string
	PartyTaxClassification string
	PrimaryJurisdictionId  string
	TaxRegistrations       []TaxRegistration
}

// TaxRegistration is one jurisdiction a party is registered in.
type TaxRegistration struct {
	JurisdictionId   string
	RegistrationType string
	IsRegistered     bool
}

// CalculationLine is one line to tax.
type CalculationLine struct {
	// LineReference identifies the line in the caller's own document. Tax echoes it back and never
	// interprets it.
	LineReference string

	ProductReference         string
	ProductTaxClassification string

	Quantity decimal.Decimal
	UomId    string

	UnitPrice      decimal.Decimal
	DiscountAmount decimal.Decimal

	// CommercialBaseAmount is the taxable base as Sales computed it, already net of discount. Tax
	// takes it as given: discounts and promotions are Sales' business (TAX-INV-17).
	CommercialBaseAmount decimal.Decimal

	// CandidateTaxIds is the caller's proposal, typically a product's default tax. Rules may add to
	// it, remove from it, or empty it entirely (BR-TAX-ESS-SUP-010).
	CandidateTaxIds []string

	// OverrideTaxIds substitutes the determined set. Requires the accounting_tax:override
	// entitlement and a reason; V1 permits no raw amount or arbitrary rate.
	OverrideTaxIds []string
	OverrideReason string
}

// CalculationResult is what the engine concluded for the whole document.
type CalculationResult struct {
	// Status is the document's overall outcome. A document is unresolved if any line is: a partial
	// answer would let a caller store a total that silently omits a line's tax.
	Status models.DeterminationStatus

	TotalExcluded      decimal.Decimal
	TotalTax           decimal.Decimal
	TotalIncluded      decimal.Decimal
	RoundingAdjustment decimal.Decimal

	AppliedRuleIds    []string
	AppliedMappingIds []string

	Lines []LineResult

	// Snapshot is the immutable payload the caller stores on its own transaction. Tax defines the
	// contract and does not own the stored copy (BR-TAX-ESS-030).
	Snapshot Snapshot
}

// LineResult is the outcome for one line.
type LineResult struct {
	LineReference string
	Status        models.DeterminationStatus

	// ErrorCode explains an unresolved line — a missing rate, an ambiguous one, an impossible UoM
	// conversion. It is a code rather than a message so a caller can branch on it.
	ErrorCode string

	// Treatment is the line's legal character when the status is no_tax_applicable.
	Treatment models.TaxTreatment

	BaseAmount    decimal.Decimal
	TotalExcluded decimal.Decimal
	TotalTax      decimal.Decimal
	TotalIncluded decimal.Decimal

	Components []ComponentResult
}

// ComponentResult is the full detail of one tax applied to one line.
//
// It is this detailed because "tax = 8%" is not enough to issue a VAT invoice, reverse a refund or
// answer an auditor (BR-TAX-ESS-028): each of those needs the base it was computed on, the version
// of the configuration that produced it, and the legal basis for it.
type ComponentResult struct {
	TaxId   string
	TaxCode string
	TaxName string

	TaxDefinitionVersionId string
	TaxDefinitionVersionNo int32
	TaxRateVersionId       string
	TaxRateVersionNo       int32

	Rate        decimal.Decimal
	FixedAmount decimal.Decimal

	TaxGroupId     string
	TaxGroupName   string
	Treatment      models.TaxTreatment
	JurisdictionId string

	CalculationType    models.CalculationType
	PriceInclusionMode models.PriceInclusionMode
	Sequence           int32

	TaxableBase decimal.Decimal

	// UnroundedTaxAmount and TaxAmount are both carried: the difference between them is what
	// RoundingAdjustment accounts for, and a refund needs to reverse the rounded figure that was
	// actually charged rather than the exact one that was not.
	UnroundedTaxAmount decimal.Decimal
	TaxAmount          decimal.Decimal
	RoundingAdjustment decimal.Decimal

	LegalReference string
}

// Snapshot is the immutable record of how a tax outcome was reached.
//
// It must be self-contained (BR-TAX-ESS-SUP-032): a screen showing a three-year-old invoice reads
// this and nothing else, never the current tax master. That is the whole mechanism by which a rate
// change cannot reinterpret a historical sale — the old numbers are not recomputed, they are read.
type Snapshot struct {
	// SchemaVersion lets a consumer that stored an older snapshot know which shape it holds.
	SchemaVersion string

	TaxDate      string
	CalculatedAt string
	Status       models.DeterminationStatus
	CurrencyCode string

	RoundingPolicyId      string
	RoundingPolicyVersion int32
	RoundingScope         models.RoundingScope
	RoundingMethod        models.RoundingMethod
	RoundingIncrement     decimal.Decimal

	AppliedRuleIds        []string
	AppliedRuleVersions   map[string]int32
	AppliedMappingId      string
	AppliedMappingVersion int32

	Lines []LineResult

	// Override records what a manual substitution replaced, so the original determination is not
	// lost (BR-TAX-ESS-SUP-023).
	Override *OverrideRecord
}

// OverrideRecord is the audit trail of a manual tax substitution.
type OverrideRecord struct {
	PreOverrideTaxIds []string
	FinalTaxIds       []string
	Reason            string
	PerformedBy       string
	PerformedAt       string
}

// SnapshotSchemaVersion is the current snapshot shape.
//
// It changes only when the payload changes in a way a stored snapshot could not be read under, so
// that a consumer holding an old one knows to read it with the old rules rather than misinterpreting
// a field that has since changed meaning.
const SnapshotSchemaVersion = "1.0"
