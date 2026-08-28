package services

import (
	"testing"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// The pure half of the currency guard (section 8, BR-PRICE-CUR-004).
//
// SetDefault and the repository-reading half of Update are exercised live rather than here, for the
// reason recorded in apply_voucher_test.go: `engineFor` is stubbable, but inventing a fake registry
// for one test pins the fake instead of the behaviour. What is worth testing without a repository
// is the decision itself, which is why it was split out.

func storedWithCurrency(code string) dmodel.DynamicFields {
	return dmodel.DynamicFields{models.SalesPricelistFieldCurrencyId: code}
}

func TestCurrencyIsNotChangingWhenAbsentFromTheUpdate(t *testing.T) {
	params := dmodel.DynamicFields{models.SalesPricelistFieldName: "Retail"}

	if isCurrencyChanging(params, storedWithCurrency("VND")) {
		t.Fatal("an update that never mentions the currency does not change it")
	}
}

// The case that matters most in practice: a client PUTting the whole record sends currency_id on
// every save. Treating "present" as "changing" would refuse every edit to a list that has rules —
// renaming it, describing it — which is not what section 8 forbids.
func TestCurrencyIsNotChangingWhenTheSameValueIsResubmitted(t *testing.T) {
	params := dmodel.DynamicFields{
		models.SalesPricelistFieldCurrencyId: "VND",
		models.SalesPricelistFieldName:       "Retail, renamed",
	}

	if isCurrencyChanging(params, storedWithCurrency("VND")) {
		t.Fatal("resubmitting the currency it already has is not a change")
	}
}

func TestCurrencyIsChangingWhenADifferentCodeIsGiven(t *testing.T) {
	params := dmodel.DynamicFields{models.SalesPricelistFieldCurrencyId: "USD"}

	if !isCurrencyChanging(params, storedWithCurrency("VND")) {
		t.Fatal("moving from VND to USD is exactly the change the guard exists for")
	}
}

// A value of the wrong type is not a currency change; it is invalid input, and the engine's own
// validation reports it. Deciding otherwise here would answer with a confusing "currency in use"
// for a request whose real problem is that it sent a number.
func TestCurrencyIsNotChangingWhenTheValueIsNotAString(t *testing.T) {
	for name, value := range map[string]any{
		"number": 704,
		"nil":    nil,
		"empty":  "",
	} {
		params := dmodel.DynamicFields{models.SalesPricelistFieldCurrencyId: value}

		if isCurrencyChanging(params, storedWithCurrency("VND")) {
			t.Fatalf("a %s currency value must not be treated as a change", name)
		}
	}
}

// A list that has no currency yet — written before this change request added the field — is being
// given one, not having one reinterpreted. That is a change, and the guard should let the
// rule-count check decide whether it is allowed.
func TestSettingACurrencyOnAListThatHasNoneCountsAsAChange(t *testing.T) {
	params := dmodel.DynamicFields{models.SalesPricelistFieldCurrencyId: "VND"}

	if !isCurrencyChanging(params, dmodel.DynamicFields{}) {
		t.Fatal("giving a currency to a list that had none is a change")
	}
}
