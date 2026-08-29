package services

import (
	"testing"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// The pure half of the currency guard.
//
// SetDefault and the repository-reading half of Update are exercised live rather than here:
// `engineFor` is stubbable, but inventing a fake registry for one test pins the fake instead of the
// behaviour. The decision itself is what is worth testing without a repository.

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
// every save, so treating present as changing would refuse every edit to a list that has rules.
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
// validation reports it. Deciding otherwise would answer "currency in use" for a request whose
// real problem is that it sent a number.
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

// A list that has no currency yet is being given one, not having one reinterpreted. That is a
// change, and the rule-count check decides whether it is allowed.
func TestSettingACurrencyOnAListThatHasNoneCountsAsAChange(t *testing.T) {
	params := dmodel.DynamicFields{models.SalesPricelistFieldCurrencyId: "VND"}

	if !isCurrencyChanging(params, dmodel.DynamicFields{}) {
		t.Fatal("giving a currency to a list that had none is a change")
	}
}
