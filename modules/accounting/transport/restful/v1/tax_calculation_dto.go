package v1

import (
	"github.com/shopspring/decimal"

	"github.com/sky-as-code/nikki-erp/modules/accounting/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/tax"
)

// Money and quantities travel as JSON strings, never as JSON numbers.
//
// A JSON number is parsed as a float64 by most clients, and a float64 cannot hold a decimal
// fraction exactly. On a tax figure that is not an academic point: 0.1 + 0.2 is famously not 0.3,
// and an invoice total that disagrees with the sum of its lines by a cent is a defect an auditor
// will find. Requests bind decimal.Decimal, which parses a JSON string exactly; responses emit
// .String().

type CalculateTaxRequest struct {
	OperationType string `json:"operation_type"`

	// TaxDate is mandatory and never defaulted from the server clock (BR-TAX-ESS-SUP-020).
	TaxDate      string `json:"tax_date"`
	CurrencyCode string `json:"currency_code"`

	Seller TaxPartyRequest `json:"seller"`
	Buyer  TaxPartyRequest `json:"buyer"`

	ShipFromJurisdictionId string `json:"ship_from_jurisdiction_id"`
	ShipToJurisdictionId   string `json:"ship_to_jurisdiction_id"`

	PriceMode string `json:"price_mode"`

	BusinessChannelCode string `json:"business_channel_code"`
	OutletReference     string `json:"outlet_reference"`

	RoundingPolicyCode string `json:"rounding_policy_code"`

	Lines []CalculateTaxLineRequest `json:"lines"`
}

func (this CalculateTaxRequest) ToCommand() it.CalculationRequest {
	lines := make([]it.CalculationLine, 0, len(this.Lines))
	for _, line := range this.Lines {
		lines = append(lines, line.toLine())
	}

	return it.CalculationRequest{
		OperationType:          models.OperationType(this.OperationType),
		TaxDate:                this.TaxDate,
		CurrencyCode:           this.CurrencyCode,
		Seller:                 this.Seller.toContext(),
		Buyer:                  this.Buyer.toContext(),
		ShipFromJurisdictionId: this.ShipFromJurisdictionId,
		ShipToJurisdictionId:   this.ShipToJurisdictionId,
		PriceMode:              models.PriceInclusionMode(this.PriceMode),
		BusinessChannelCode:    this.BusinessChannelCode,
		OutletReference:        this.OutletReference,
		RoundingPolicyCode:     this.RoundingPolicyCode,
		Lines:                  lines,
	}
}

type TaxPartyRequest struct {
	PartyReference         string                   `json:"party_reference"`
	PartyTaxClassification string                   `json:"party_tax_classification"`
	PrimaryJurisdictionId  string                   `json:"primary_jurisdiction_id"`
	TaxRegistrations       []TaxRegistrationRequest `json:"tax_registrations"`
}

func (this TaxPartyRequest) toContext() it.TaxPartyContext {
	registrations := make([]it.TaxRegistration, 0, len(this.TaxRegistrations))
	for _, registration := range this.TaxRegistrations {
		registrations = append(registrations, it.TaxRegistration{
			JurisdictionId:   registration.JurisdictionId,
			RegistrationType: registration.RegistrationType,
			IsRegistered:     registration.IsRegistered,
		})
	}
	return it.TaxPartyContext{
		PartyReference:         this.PartyReference,
		PartyTaxClassification: this.PartyTaxClassification,
		PrimaryJurisdictionId:  this.PrimaryJurisdictionId,
		TaxRegistrations:       registrations,
	}
}

type TaxRegistrationRequest struct {
	JurisdictionId   string `json:"jurisdiction_id"`
	RegistrationType string `json:"registration_type"`
	IsRegistered     bool   `json:"is_registered"`
}

type CalculateTaxLineRequest struct {
	LineReference            string `json:"line_reference"`
	ProductReference         string `json:"product_reference"`
	ProductTaxClassification string `json:"product_tax_classification"`

	Quantity decimal.Decimal `json:"quantity"`
	UomId    string          `json:"uom_id"`

	UnitPrice      decimal.Decimal `json:"unit_price"`
	DiscountAmount decimal.Decimal `json:"discount_amount"`

	// CommercialBaseAmount is the taxable base as the caller computed it, already net of discount.
	// Tax takes it as given rather than deriving it (TAX-INV-17).
	CommercialBaseAmount decimal.Decimal `json:"commercial_base_amount"`

	CandidateTaxIds []string `json:"candidate_tax_ids"`

	OverrideTaxIds []string `json:"override_tax_ids"`
	OverrideReason string   `json:"override_reason"`
}

func (this CalculateTaxLineRequest) toLine() it.CalculationLine {
	return it.CalculationLine{
		LineReference:            this.LineReference,
		ProductReference:         this.ProductReference,
		ProductTaxClassification: this.ProductTaxClassification,
		Quantity:                 this.Quantity,
		UomId:                    this.UomId,
		UnitPrice:                this.UnitPrice,
		DiscountAmount:           this.DiscountAmount,
		CommercialBaseAmount:     this.CommercialBaseAmount,
		CandidateTaxIds:          this.CandidateTaxIds,
		OverrideTaxIds:           this.OverrideTaxIds,
		OverrideReason:           this.OverrideReason,
	}
}

type CalculateTaxResponse struct {
	Status string `json:"status"`

	TotalExcluded      string `json:"total_excluded"`
	TotalTax           string `json:"total_tax"`
	TotalIncluded      string `json:"total_included"`
	RoundingAdjustment string `json:"rounding_adjustment"`

	AppliedRuleIds    []string `json:"applied_rule_ids"`
	AppliedMappingIds []string `json:"applied_mapping_ids"`

	Lines []TaxLineResponse `json:"lines"`

	// Snapshot is the payload the caller stores on its own transaction. Accounting keeps no copy
	// and holds no foreign key into the caller's schema (BR-TAX-ESS-030).
	Snapshot TaxSnapshotResponse `json:"snapshot"`
}

func NewCalculateTaxResponse(data it.CalculationResult) CalculateTaxResponse {
	return CalculateTaxResponse{
		Status:             string(data.Status),
		TotalExcluded:      data.TotalExcluded.String(),
		TotalTax:           data.TotalTax.String(),
		TotalIncluded:      data.TotalIncluded.String(),
		RoundingAdjustment: data.RoundingAdjustment.String(),
		AppliedRuleIds:     data.AppliedRuleIds,
		AppliedMappingIds:  data.AppliedMappingIds,
		Lines:              newTaxLineResponses(data.Lines),
		Snapshot:           newTaxSnapshotResponse(data.Snapshot),
	}
}

type TaxLineResponse struct {
	LineReference string `json:"line_reference"`
	Status        string `json:"status"`

	// ErrorCode explains an unresolved line — a missing rate, an ambiguous one, an impossible UoM
	// conversion. It is a code rather than prose so a caller can branch on it.
	ErrorCode string `json:"error_code,omitempty"`
	Treatment string `json:"treatment,omitempty"`

	BaseAmount    string `json:"base_amount"`
	TotalExcluded string `json:"total_excluded"`
	TotalTax      string `json:"total_tax"`
	TotalIncluded string `json:"total_included"`

	Components []TaxComponentResponse `json:"components"`
}

func newTaxLineResponses(lines []it.LineResult) []TaxLineResponse {
	responses := make([]TaxLineResponse, 0, len(lines))
	for _, line := range lines {
		components := make([]TaxComponentResponse, 0, len(line.Components))
		for _, component := range line.Components {
			components = append(components, newTaxComponentResponse(component))
		}
		responses = append(responses, TaxLineResponse{
			LineReference: line.LineReference,
			Status:        string(line.Status),
			ErrorCode:     line.ErrorCode,
			Treatment:     string(line.Treatment),
			BaseAmount:    line.BaseAmount.String(),
			TotalExcluded: line.TotalExcluded.String(),
			TotalTax:      line.TotalTax.String(),
			TotalIncluded: line.TotalIncluded.String(),
			Components:    components,
		})
	}
	return responses
}

// TaxComponentResponse is the full detail of one tax applied to one line.
//
// It is this detailed because "tax = 8%" is not enough to issue a VAT invoice, reverse a refund or
// answer an auditor: each of those needs the base it was computed on, the version of the
// configuration that produced it, and its legal basis (BR-TAX-ESS-028).
type TaxComponentResponse struct {
	TaxId   string `json:"tax_id"`
	TaxCode string `json:"tax_code"`
	TaxName string `json:"tax_name"`

	TaxDefinitionVersionId string `json:"tax_definition_version_id"`
	TaxDefinitionVersionNo int32  `json:"tax_definition_version_no"`
	TaxRateVersionId       string `json:"tax_rate_version_id"`
	TaxRateVersionNo       int32  `json:"tax_rate_version_no"`

	Rate        string `json:"rate"`
	FixedAmount string `json:"fixed_amount"`

	TaxGroupId     string `json:"tax_group_id,omitempty"`
	Treatment      string `json:"treatment"`
	JurisdictionId string `json:"jurisdiction_id,omitempty"`

	CalculationType    string `json:"calculation_type"`
	PriceInclusionMode string `json:"price_inclusion_mode"`
	Sequence           int32  `json:"sequence"`

	TaxableBase string `json:"taxable_base"`

	// UnroundedTaxAmount and TaxAmount are both carried: the difference between them is what
	// rounding_adjustment accounts for, and a refund must reverse the rounded figure that was
	// actually charged rather than the exact one that was not.
	UnroundedTaxAmount string `json:"unrounded_tax_amount"`
	TaxAmount          string `json:"tax_amount"`
	RoundingAdjustment string `json:"rounding_adjustment"`

	LegalReference string `json:"legal_reference,omitempty"`
}

func newTaxComponentResponse(component it.ComponentResult) TaxComponentResponse {
	return TaxComponentResponse{
		TaxId:                  component.TaxId,
		TaxCode:                component.TaxCode,
		TaxName:                component.TaxName,
		TaxDefinitionVersionId: component.TaxDefinitionVersionId,
		TaxDefinitionVersionNo: component.TaxDefinitionVersionNo,
		TaxRateVersionId:       component.TaxRateVersionId,
		TaxRateVersionNo:       component.TaxRateVersionNo,
		Rate:                   component.Rate.String(),
		FixedAmount:            component.FixedAmount.String(),
		TaxGroupId:             component.TaxGroupId,
		Treatment:              string(component.Treatment),
		JurisdictionId:         component.JurisdictionId,
		CalculationType:        string(component.CalculationType),
		PriceInclusionMode:     string(component.PriceInclusionMode),
		Sequence:               component.Sequence,
		TaxableBase:            component.TaxableBase.String(),
		UnroundedTaxAmount:     component.UnroundedTaxAmount.String(),
		TaxAmount:              component.TaxAmount.String(),
		RoundingAdjustment:     component.RoundingAdjustment.String(),
		LegalReference:         component.LegalReference,
	}
}

type TaxSnapshotResponse struct {
	SchemaVersion string `json:"schema_version"`

	TaxDate      string `json:"tax_date"`
	CalculatedAt string `json:"calculated_at"`
	Status       string `json:"status"`
	CurrencyCode string `json:"currency_code"`

	RoundingPolicyId      string `json:"rounding_policy_id,omitempty"`
	RoundingPolicyVersion int32  `json:"rounding_policy_version,omitempty"`
	RoundingScope         string `json:"rounding_scope"`
	RoundingMethod        string `json:"rounding_method"`
	RoundingIncrement     string `json:"rounding_increment"`

	AppliedRuleIds []string `json:"applied_rule_ids"`

	Lines []TaxLineResponse `json:"lines"`
}

func newTaxSnapshotResponse(snapshot it.Snapshot) TaxSnapshotResponse {
	return TaxSnapshotResponse{
		SchemaVersion:         snapshot.SchemaVersion,
		TaxDate:               snapshot.TaxDate,
		CalculatedAt:          snapshot.CalculatedAt,
		Status:                string(snapshot.Status),
		CurrencyCode:          snapshot.CurrencyCode,
		RoundingPolicyId:      snapshot.RoundingPolicyId,
		RoundingPolicyVersion: snapshot.RoundingPolicyVersion,
		RoundingScope:         string(snapshot.RoundingScope),
		RoundingMethod:        string(snapshot.RoundingMethod),
		RoundingIncrement:     snapshot.RoundingIncrement.String(),
		AppliedRuleIds:        snapshot.AppliedRuleIds,
		Lines:                 newTaxLineResponses(snapshot.Lines),
	}
}

// SimulateTaxRequest is the same input as a calculation.
//
// A distinct type rather than an alias so the simulator's request can gain a field — a hypothetical
// rate to test against, say — without changing the contract every Sales order already depends on.
type SimulateTaxRequest struct {
	CalculateTaxRequest
}

func (this SimulateTaxRequest) ToCommand() it.CalculationRequest {
	return this.CalculateTaxRequest.ToCommand()
}

type SimulateTaxResponse struct {
	Calculation CalculateTaxResponse `json:"calculation"`
	Trace       []TaxTraceResponse   `json:"trace"`
}

// TaxTraceResponse is one stage of the pipeline, as it happened.
type TaxTraceResponse struct {
	Stage   string   `json:"stage"`
	Detail  string   `json:"detail"`
	TaxIds  []string `json:"tax_ids,omitempty"`
	RuleIds []string `json:"rule_ids,omitempty"`
}

func NewSimulateTaxResponse(data it.SimulationResult) SimulateTaxResponse {
	trace := make([]TaxTraceResponse, 0, len(data.Trace))
	for _, step := range data.Trace {
		trace = append(trace, TaxTraceResponse{
			Stage:   step.Stage,
			Detail:  step.Detail,
			TaxIds:  step.TaxIds,
			RuleIds: step.RuleIds,
		})
	}
	return SimulateTaxResponse{
		Calculation: NewCalculateTaxResponse(data.Calculation),
		Trace:       trace,
	}
}
