package services

import (
	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/models"
)

// The rules a vendor price carries that need nothing but the row itself.
//
// Separated from VendorPriceValidator so the whole rule table is testable without ports or a
// database. What is left there is only what genuinely needs to ask another module a question.

// assertVendorPriceSelfConsistent checks a row against itself.
//
// The schema already bounds each number on its own — min and max on the decimal types — so what is
// left here is what the schema cannot say: how two fields relate. A validity window is the whole of
// it, and a row that fails it is not rejected by anything else.
func assertVendorPriceSelfConsistent(record dmodel.DynamicFields, vErrs *ft.ClientErrors) {
	assertVendorPriceValidity(record, vErrs)
}

// assertVendorPriceValidity refuses a window that never opens.
//
// A price whose valid_from is after its valid_to applies on no date at all. Nothing fails: the row
// is written, looks ordinary in a listing, and simply never resolves — so an order is priced from
// some other row, or refused for having no vendor price, and the reason is invisible.
func assertVendorPriceValidity(record dmodel.DynamicFields, vErrs *ft.ClientErrors) {
	from := recordString(record, models.VendorProductPriceFieldValidFrom)
	to := recordString(record, models.VendorProductPriceFieldValidTo)
	if from == "" || to == "" {
		return // An open-ended window is the normal case, not an incomplete one.
	}
	// Both are ISO-8601 in UTC, so lexical order is chronological order.
	if from > to {
		appendLineViolation(vErrs, models.VendorProductPriceFieldValidTo,
			"purchase_vendor_product_price.validity_inverted",
			"this price stops applying before it starts, so it would never apply at all")
	}
}

// recordString reads one field as a string, whatever shape it arrived in.
//
// Never a bare type assertion: values come from a JSON body, and a client sending a number where an
// id belongs would panic the request rather than be told what is wrong with it.
func recordString(record dmodel.DynamicFields, field string) string {
	value, present := record[field]
	if !present || value == nil {
		return ""
	}
	if text, ok := value.(string); ok {
		return text
	}
	if stringer, ok := value.(interface{ String() string }); ok {
		return stringer.String()
	}
	return ""
}

// oneUnit is the probe quantity the UoM compatibility check converts.
//
// One rather than zero: zero converts successfully between any two units, including incompatible
// ones, so it would pass the very check it is meant to fail.
func oneUnit() decimal.Decimal {
	return decimal.NewFromInt(1)
}
