package models

// The tax context: the facts a determination rule is allowed to test.
//
// This is a closed list on purpose (BR-TAX-ESS-SUP-007). A rule condition names a field key, an
// operator and a value, and nothing else — no SQL, no expression to evaluate, no path into an
// arbitrary object graph. Keeping the vocabulary finite is what makes rule evaluation something the
// engine can reason about and validate ahead of time, rather than something it discovers at
// calculation time on a customer's order.
//
// The keys mirror the request contract of BR-TAX-ESS-007 and BR-TAX-ESS-025/026. Adding one means
// the engine can actually supply it; a key that nothing populates would match nothing and read as
// a rule that mysteriously never fires.

const (
	CtxOperationType            = "operation_type"
	CtxTaxDate                  = "tax_date"
	CtxCurrencyCode             = "currency_code"
	CtxProductTaxClassification = "product_tax_classification"
	CtxPartyTaxClassification   = "party_tax_classification"
	CtxSellerJurisdictionId     = "seller_jurisdiction_id"
	CtxBuyerJurisdictionId      = "buyer_jurisdiction_id"
	CtxShipFromJurisdictionId   = "ship_from_jurisdiction_id"
	CtxShipToJurisdictionId     = "ship_to_jurisdiction_id"
	CtxBuyerIsTaxRegistered     = "buyer_is_tax_registered"
	CtxSellerIsTaxRegistered    = "seller_is_tax_registered"
	CtxCandidateTaxId           = "candidate_tax_id"
	CtxCommercialBaseAmount     = "commercial_base_amount"
	CtxBusinessChannelCode      = "business_channel_code"
	CtxProductReference         = "product_reference"
)

// ContextFieldType is the datatype the engine knows a context field to hold.
//
// It exists so that condition validation can reject an ordering comparison against something that
// has no ordering — "greater than" applied to a jurisdiction id is meaningless, and catching it at
// configuration time is cheaper than discovering it produced no match on an invoice.
type ContextFieldType string

const (
	ContextTypeString  ContextFieldType = "string"
	ContextTypeBoolean ContextFieldType = "boolean"
	ContextTypeMoney   ContextFieldType = "money"
	ContextTypeDate    ContextFieldType = "date"
)

// contextFieldTypes is the whitelist itself, mapping each testable field to its type.
var contextFieldTypes = map[string]ContextFieldType{
	CtxOperationType:            ContextTypeString,
	CtxTaxDate:                  ContextTypeDate,
	CtxCurrencyCode:             ContextTypeString,
	CtxProductTaxClassification: ContextTypeString,
	CtxPartyTaxClassification:   ContextTypeString,
	CtxSellerJurisdictionId:     ContextTypeString,
	CtxBuyerJurisdictionId:      ContextTypeString,
	CtxShipFromJurisdictionId:   ContextTypeString,
	CtxShipToJurisdictionId:     ContextTypeString,
	CtxBuyerIsTaxRegistered:     ContextTypeBoolean,
	CtxSellerIsTaxRegistered:    ContextTypeBoolean,
	CtxCandidateTaxId:           ContextTypeString,
	CtxCommercialBaseAmount:     ContextTypeMoney,
	CtxBusinessChannelCode:      ContextTypeString,
	CtxProductReference:         ContextTypeString,
}

// IsKnownContextField reports whether a rule may test this field.
func IsKnownContextField(fieldKey string) bool {
	_, known := contextFieldTypes[fieldKey]
	return known
}

// ContextFieldTypeOf returns the datatype of a context field, and whether it is known at all.
func ContextFieldTypeOf(fieldKey string) (ContextFieldType, bool) {
	fieldType, known := contextFieldTypes[fieldKey]
	return fieldType, known
}

// IsOrderableContextField reports whether gte and lte mean anything against this field.
func IsOrderableContextField(fieldKey string) bool {
	fieldType, known := contextFieldTypes[fieldKey]
	if !known {
		return false
	}
	return fieldType == ContextTypeMoney || fieldType == ContextTypeDate
}

// IsMoneyContextField reports whether a condition on this field must carry a currency.
//
// A money threshold without one is not comparable: "over 1,000,000" means something different in
// VND and in USD, and V1 does not convert between them (BR-TAX-ESS-SUP-007).
func IsMoneyContextField(fieldKey string) bool {
	fieldType, known := contextFieldTypes[fieldKey]
	return known && fieldType == ContextTypeMoney
}

// IsNullaryOperator reports whether the operator takes no comparison value.
func IsNullaryOperator(op ConditionOperator) bool {
	return op == OperatorIsNull || op == OperatorIsNotNull
}

// IsArrayOperator reports whether the operator compares against a list.
func IsArrayOperator(op ConditionOperator) bool {
	return op == OperatorIn || op == OperatorNotIn
}

// IsOrderingOperator reports whether the operator needs an orderable field.
func IsOrderingOperator(op ConditionOperator) bool {
	return op == OperatorGte || op == OperatorLte
}
