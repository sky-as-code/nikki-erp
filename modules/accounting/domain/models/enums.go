package models

// The enum values every tax schema shares, as typed constants rather than bare strings, declared
// once because several appear in more than one schema. Each Wrap function returns nil for a value
// outside the set, so a caller reading a corrupt row gets nil rather than an unhandled value.

type LifecycleStatus string

const (
	// LifecycleDraft configuration is invisible to the engine and freely editable.
	LifecycleDraft LifecycleStatus = "draft"
	// LifecyclePublished configuration is what the engine calculates with; its material fields are
	// frozen from this point on.
	LifecyclePublished LifecycleStatus = "published"
	// LifecycleWithdrawn configuration is retired from new determination but kept for audit, and
	// can never return to draft.
	LifecycleWithdrawn LifecycleStatus = "withdrawn"
)

func WrapLifecycleStatus(v string) *LifecycleStatus {
	switch LifecycleStatus(v) {
	case LifecycleDraft, LifecyclePublished, LifecycleWithdrawn:
		s := LifecycleStatus(v)
		return &s
	}
	return nil
}

type TaxUsage string

const (
	TaxUsageSale     TaxUsage = "sale"
	TaxUsagePurchase TaxUsage = "purchase"
	TaxUsageBoth     TaxUsage = "both"
	TaxUsageNone     TaxUsage = "none"
)

func WrapTaxUsage(v string) *TaxUsage {
	switch TaxUsage(v) {
	case TaxUsageSale, TaxUsagePurchase, TaxUsageBoth, TaxUsageNone:
		s := TaxUsage(v)
		return &s
	}
	return nil
}

type TaxKind string

const (
	TaxKindVat           TaxKind = "vat"
	TaxKindGst           TaxKind = "gst"
	TaxKindSalesTax      TaxKind = "sales_tax"
	TaxKindExcise        TaxKind = "excise"
	TaxKindEnvironmental TaxKind = "environmental"
	// TaxKindWithholding classifies a tax as withholding for reporting only; the engine cannot
	// calculate withholding, whose base, sign, gross/net effect and reversal are undefined.
	TaxKindWithholding TaxKind = "withholding"
	TaxKindOther       TaxKind = "other"
)

func WrapTaxKind(v string) *TaxKind {
	switch TaxKind(v) {
	case TaxKindVat, TaxKindGst, TaxKindSalesTax, TaxKindExcise,
		TaxKindEnvironmental, TaxKindWithholding, TaxKindOther:
		s := TaxKind(v)
		return &s
	}
	return nil
}

type CalculationType string

const (
	// CalculationPercentage computes tax as base x rate / 100.
	CalculationPercentage CalculationType = "percentage"
	// CalculationDivision is percentage-of-total.
	CalculationDivision CalculationType = "division"
	// CalculationFixed charges a money amount per unit of quantity.
	CalculationFixed CalculationType = "fixed"
	// CalculationGroup has no rate of its own and delegates to its components.
	CalculationGroup CalculationType = "group"
	// CalculationNone carries legal semantics without producing an amount. Valid only with exempt,
	// non_taxable or out_of_scope, never with zero_rated, which is a real 0% tax.
	CalculationNone CalculationType = "none"
)

func WrapCalculationType(v string) *CalculationType {
	switch CalculationType(v) {
	case CalculationPercentage, CalculationDivision, CalculationFixed,
		CalculationGroup, CalculationNone:
		s := CalculationType(v)
		return &s
	}
	return nil
}

type TaxTreatment string

const (
	TaxTreatmentTaxable TaxTreatment = "taxable"
	// TaxTreatmentZeroRated is a real tax charged at 0%, with a tax code on the invoice. It is not an
	// absence of tax and may never be expressed as no_tax_applicable.
	TaxTreatmentZeroRated  TaxTreatment = "zero_rated"
	TaxTreatmentExempt     TaxTreatment = "exempt"
	TaxTreatmentNonTaxable TaxTreatment = "non_taxable"
	TaxTreatmentOutOfScope TaxTreatment = "out_of_scope"
)

func WrapTaxTreatment(v string) *TaxTreatment {
	switch TaxTreatment(v) {
	case TaxTreatmentTaxable, TaxTreatmentZeroRated, TaxTreatmentExempt,
		TaxTreatmentNonTaxable, TaxTreatmentOutOfScope:
		s := TaxTreatment(v)
		return &s
	}
	return nil
}

// ZeroAmountTreatments are the four treatments that all produce a tax amount of zero. The amount
// cannot distinguish them, and zero-rated, exempt, non-taxable and out-of-scope are four different
// statements in law, so the treatment must be carried alongside it.
var ZeroAmountTreatments = []TaxTreatment{
	TaxTreatmentZeroRated,
	TaxTreatmentExempt,
	TaxTreatmentNonTaxable,
	TaxTreatmentOutOfScope,
}

type PriceInclusionMode string

const (
	// PriceInclusionInherit defers to the request's price_mode, letting one tax serve both
	// tax-inclusive and tax-exclusive pricing.
	PriceInclusionInherit  PriceInclusionMode = "inherit"
	PriceInclusionIncluded PriceInclusionMode = "included"
	PriceInclusionExcluded PriceInclusionMode = "excluded"
)

func WrapPriceInclusionMode(v string) *PriceInclusionMode {
	switch PriceInclusionMode(v) {
	case PriceInclusionInherit, PriceInclusionIncluded, PriceInclusionExcluded:
		s := PriceInclusionMode(v)
		return &s
	}
	return nil
}

type JurisdictionLevel string

const (
	JurisdictionCountry  JurisdictionLevel = "country"
	JurisdictionState    JurisdictionLevel = "state"
	JurisdictionProvince JurisdictionLevel = "province"
	JurisdictionCounty   JurisdictionLevel = "county"
	JurisdictionCity     JurisdictionLevel = "city"
	JurisdictionSpecial  JurisdictionLevel = "special"
)

func WrapJurisdictionLevel(v string) *JurisdictionLevel {
	switch JurisdictionLevel(v) {
	case JurisdictionCountry, JurisdictionState, JurisdictionProvince,
		JurisdictionCounty, JurisdictionCity, JurisdictionSpecial:
		s := JurisdictionLevel(v)
		return &s
	}
	return nil
}

type RoundingScope string

const (
	RoundingScopeLine RoundingScope = "line"
	// RoundingScopeDocument requires the engine to receive every line in one call, since rounding
	// each line and summing produces a different number.
	RoundingScopeDocument RoundingScope = "document"
)

func WrapRoundingScope(v string) *RoundingScope {
	switch RoundingScope(v) {
	case RoundingScopeLine, RoundingScopeDocument:
		s := RoundingScope(v)
		return &s
	}
	return nil
}

type RoundingMethod string

const (
	RoundingHalfUp   RoundingMethod = "half_up"
	RoundingHalfEven RoundingMethod = "half_even"
	RoundingUp       RoundingMethod = "up"
	RoundingDown     RoundingMethod = "down"
)

func WrapRoundingMethod(v string) *RoundingMethod {
	switch RoundingMethod(v) {
	case RoundingHalfUp, RoundingHalfEven, RoundingUp, RoundingDown:
		s := RoundingMethod(v)
		return &s
	}
	return nil
}

type ConditionOperator string

const (
	OperatorEq        ConditionOperator = "eq"
	OperatorNotEq     ConditionOperator = "not_eq"
	OperatorIn        ConditionOperator = "in"
	OperatorNotIn     ConditionOperator = "not_in"
	OperatorIsNull    ConditionOperator = "is_null"
	OperatorIsNotNull ConditionOperator = "is_not_null"
	OperatorGte       ConditionOperator = "gte"
	OperatorLte       ConditionOperator = "lte"
)

func WrapConditionOperator(v string) *ConditionOperator {
	switch ConditionOperator(v) {
	case OperatorEq, OperatorNotEq, OperatorIn, OperatorNotIn,
		OperatorIsNull, OperatorIsNotNull, OperatorGte, OperatorLte:
		s := ConditionOperator(v)
		return &s
	}
	return nil
}

// NullaryOperators take no operand at all, so a condition using one must carry no value.
var NullaryOperators = []ConditionOperator{OperatorIsNull, OperatorIsNotNull}

// ArrayOperators require the condition value to be a JSON array.
var ArrayOperators = []ConditionOperator{OperatorIn, OperatorNotIn}

// OrderableOperators are valid only against a field whose type has an ordering.
var OrderableOperators = []ConditionOperator{OperatorGte, OperatorLte}

type RuleResultAction string

const (
	ActionAddTax          RuleResultAction = "add_tax"
	ActionRemoveTax       RuleResultAction = "remove_tax"
	ActionApplyMapping    RuleResultAction = "apply_mapping"
	ActionNoTaxApplicable RuleResultAction = "no_tax_applicable"
)

func WrapRuleResultAction(v string) *RuleResultAction {
	switch RuleResultAction(v) {
	case ActionAddTax, ActionRemoveTax, ActionApplyMapping, ActionNoTaxApplicable:
		s := RuleResultAction(v)
		return &s
	}
	return nil
}

// DeterminationStatus is the outcome of determination and is never inferred from an amount: a zero
// tax amount is compatible with all three, so the status is carried explicitly through the result
// and into the snapshot.
type DeterminationStatus string

const (
	// DeterminationResolved means the engine could conclude; the conclusion may still be that no tax
	// is charged, when the treatment is zero_rated or exempt.
	DeterminationResolved DeterminationStatus = "resolved"
	// DeterminationNoTaxApplicable means the engine positively established the transaction sits
	// outside the applicable tax set.
	DeterminationNoTaxApplicable DeterminationStatus = "no_tax_applicable"
	// DeterminationUnresolved means the engine could not conclude, and must never be quietly turned
	// into 0%.
	DeterminationUnresolved DeterminationStatus = "unresolved"
)

// The error codes an unresolved determination reports, so a caller can tell a missing rate from an
// ambiguous one without parsing a message.
const (
	ErrCodeTaxRateMissing      = "tax_rate_missing"
	ErrCodeTaxRateAmbiguous    = "tax_rate_ambiguous"
	ErrCodeMultipleTaxMappings = "multiple_tax_mappings"

	// A named tax may have no definition in force on the date. That is distinct from a missing rate:
	// the tax exists but says nothing about that day, usually a wrong effective period.
	ErrCodeTaxNotFound            = "tax_not_found"
	ErrCodeTaxDefinitionMissing   = "tax_definition_missing"
	ErrCodeTaxDefinitionAmbiguous = "tax_definition_ambiguous"

	// ErrCodeTaxConfigurationInvalid covers a configuration validation should have refused, such as a
	// group with no components. It is reported rather than panicked on: a bad row must not take the
	// process down.
	ErrCodeTaxConfigurationInvalid = "tax_configuration_invalid"
	ErrCodeUomConversion           = "tax_uom_conversion_unavailable"
	ErrCodeFixedTaxCurrency        = "fixed_tax_currency_conversion_required"
	ErrCodeRoundingPolicyMissing   = "rounding_policy_missing"
	ErrCodeClassificationMissing   = "tax_classification_missing"
	ErrCodeNoApplicableTax         = "no_applicable_tax_determined"
)

// OperationType is which side of trade a calculation request represents.
type OperationType string

const (
	OperationSale       OperationType = "sale"
	OperationSaleRefund OperationType = "sale_refund"
	// OperationPurchase and OperationPurchaseRefund are reserved contract values; the engine rejects
	// them rather than guessing at purchase-side semantics.
	OperationPurchase       OperationType = "purchase"
	OperationPurchaseRefund OperationType = "purchase_refund"
)

func WrapOperationType(v string) *OperationType {
	switch OperationType(v) {
	case OperationSale, OperationSaleRefund, OperationPurchase, OperationPurchaseRefund:
		s := OperationType(v)
		return &s
	}
	return nil
}

// IsImplementedOperation reports whether V1 can actually calculate this operation type.
func IsImplementedOperation(v OperationType) bool {
	return v == OperationSale || v == OperationSaleRefund
}
