// Package tax is the port other modules consume to have tax determined and calculated. Everything
// here is a plain value type; the caller stores the returned snapshot itself, as Accounting keeps no
// record of the transaction and holds no foreign key into it.
package tax

import (
	"github.com/shopspring/decimal"

	"github.com/sky-as-code/nikki-erp/modules/accounting/domain/models"
)

// CalculationRequest is one document's worth of tax to compute. It is document-level because
// rounding may be document-scoped, and rounding each line then summing yields a different number.
type CalculationRequest struct {
	OrgId string

	// OperationType must be sale or sale_refund; the purchase values are reserved and rejected here.
	OperationType models.OperationType

	// TaxDate decides which configuration is in force and is mandatory: it is never defaulted from
	// the server clock, or a sale would be priced against processing time rather than its legal date.
	TaxDate string

	CurrencyCode string

	Seller TaxPartyContext
	Buyer  TaxPartyContext

	ShipFromJurisdictionId string
	ShipToJurisdictionId   string

	// PriceMode is the document-wide tax-inclusive/exclusive default that a tax whose inclusion mode
	// is "inherit" resolves against.
	PriceMode models.PriceInclusionMode

	// BusinessChannelCode and OutletReference are opaque: carried for rule conditions and audit,
	// never resolved against Sales.
	BusinessChannelCode string
	OutletReference     string

	// RoundingPolicyCode names the policy to apply. Empty falls back to the organization default;
	// when neither names one the result is unresolved rather than a guessed scale.
	RoundingPolicyCode string

	Lines []CalculationLine
}

// TaxPartyContext is the tax-relevant facts about one side of a transaction. PartyReference is
// opaque and for tracing only: Tax must not query Contacts with it, or the dependency runs backwards.
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
	// LineReference is the caller's own line id, echoed back and never interpreted.
	LineReference string

	ProductReference         string
	ProductTaxClassification string

	Quantity decimal.Decimal
	UomId    string

	UnitPrice      decimal.Decimal
	DiscountAmount decimal.Decimal

	// CommercialBaseAmount is the taxable base as the caller computed it, already net of discount,
	// and is taken as given.
	CommercialBaseAmount decimal.Decimal

	// CandidateTaxIds is the caller's proposal; rules may add to it, remove from it, or empty it.
	CandidateTaxIds []string

	// OverrideTaxIds substitutes the determined set; it requires the accounting_tax:override
	// entitlement and a reason. A raw amount or arbitrary rate is not permitted.
	OverrideTaxIds []string
	OverrideReason string
}

// CalculationResult is what the engine concluded for the whole document.
type CalculationResult struct {
	// Status is unresolved if any line is, so a caller cannot store a total that silently omits a
	// line's tax.
	Status models.DeterminationStatus

	TotalExcluded      decimal.Decimal
	TotalTax           decimal.Decimal
	TotalIncluded      decimal.Decimal
	RoundingAdjustment decimal.Decimal

	AppliedRuleIds    []string
	AppliedMappingIds []string

	Lines []LineResult

	// Snapshot is the immutable payload the caller stores on its own transaction; Tax defines the
	// contract but does not own the stored copy.
	Snapshot Snapshot
}

// LineResult is the outcome for one line.
type LineResult struct {
	LineReference string
	Status        models.DeterminationStatus

	// ErrorCode explains an unresolved line; a code rather than a message so callers can branch on it.
	ErrorCode string

	// Treatment is the line's legal character when the status is no_tax_applicable.
	Treatment models.TaxTreatment

	BaseAmount    decimal.Decimal
	TotalExcluded decimal.Decimal
	TotalTax      decimal.Decimal
	TotalIncluded decimal.Decimal

	Components []ComponentResult
}

// ComponentResult is the full detail of one tax applied to one line: the base it was computed on,
// the configuration versions that produced it, and its legal basis, all of which invoicing, refunds
// and audit need.
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

	// Both amounts are carried: their difference is RoundingAdjustment, and a refund must reverse the
	// rounded figure actually charged, not the exact one.
	UnroundedTaxAmount decimal.Decimal
	TaxAmount          decimal.Decimal
	RoundingAdjustment decimal.Decimal

	LegalReference string
}

// Snapshot is the immutable record of how a tax outcome was reached. It must stay self-contained:
// readers use it alone and never the current tax master, so a later rate change cannot reinterpret a
// historical sale.
type Snapshot struct {
	// SchemaVersion tells a consumer holding an older snapshot which shape it has.
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

	// Override records what a manual substitution replaced, so the original determination survives.
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

// SnapshotSchemaVersion changes only when the payload changes incompatibly, so a consumer holding an
// old snapshot reads it under the old rules.
const SnapshotSchemaVersion = "1.0"
