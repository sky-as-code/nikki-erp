package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	itVendor "github.com/sky-as-code/nikki-erp/modules/contacts/interfaces/vendor"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	itCurrency "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/currency"
	itExt "github.com/sky-as-code/nikki-erp/modules/purchase/interfaces/external"

	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/models"
)

type stubVendors struct {
	found           bool
	orderable       bool
	defaultCurrency string
}

func (this *stubVendors) GetVendor(
	_ corectx.Context, query itExt.GetVendorQuery,
) (*itExt.GetVendorResult, error) {
	if !this.found {
		return &itExt.GetVendorResult{}, nil
	}
	data := itVendor.GetVendorResultData{
		PartyId:     query.PartyId,
		OrgId:       query.OrgId,
		IsOrderable: this.orderable,
	}
	if this.defaultCurrency != "" {
		currencyId := model.Id(this.defaultCurrency)
		data.DefaultCurrencyId = &currencyId
	}
	return &itExt.GetVendorResult{HasData: true, Data: data}, nil
}

func (this *stubVendors) AssertOrderable(
	_ corectx.Context, query itExt.AssertOrderableQuery,
) (*itExt.AssertOrderableResult, error) {
	if this.found && this.orderable {
		return &itExt.AssertOrderableResult{}, nil
	}
	vErrs := ft.NewClientErrors()
	vErrs.Append(*ft.NewBusinessViolation(query.Field, "vendor.not_orderable",
		"this vendor is not active and cannot be selected for a new order"))
	return &itExt.AssertOrderableResult{ClientErrors: *vErrs}, nil
}

type stubCurrencies struct {
	found         bool
	usable        bool
	decimalPlaces int32
}

func (this *stubCurrencies) GetCurrency(
	_ corectx.Context, query itExt.GetCurrencyQuery,
) (*itExt.GetCurrencyResult, error) {
	if !this.found {
		return &itExt.GetCurrencyResult{}, nil
	}
	return &itExt.GetCurrencyResult{
		HasData: true,
		Data: itCurrency.GetCurrencyResultData{
			Id:            query.Id,
			DecimalPlaces: this.decimalPlaces,
			IsActive:      this.usable,
		},
	}, nil
}

func (this *stubCurrencies) AssertUsable(
	_ corectx.Context, query itExt.AssertUsableQuery,
) (*itExt.AssertUsableResult, error) {
	if this.found && this.usable {
		return &itExt.AssertUsableResult{}, nil
	}
	vErrs := ft.NewClientErrors()
	vErrs.Append(*ft.NewBusinessViolation(query.Field, "currency.not_active",
		"this currency is not active"))
	return &itExt.AssertUsableResult{ClientErrors: *vErrs}, nil
}

func usableVendor(defaultCurrency string) *stubVendors {
	return &stubVendors{found: true, orderable: true, defaultCurrency: defaultCurrency}
}

func usableCurrency(places int32) *stubCurrencies {
	return &stubCurrencies{found: true, usable: true, decimalPlaces: places}
}

func orderParams(vendorId, currencyId string) dmodel.DynamicFields {
	params := dmodel.DynamicFields{
		models.PurchaseOrderFieldVendorId: vendorId,
	}
	if currencyId != "" {
		params[models.PurchaseOrderFieldCurrencyId] = currencyId
	}
	return params
}

// "Is a vendor" means "has a vendor profile", which Contacts answers; Purchase does not compare a
// status itself, so "may be ordered from" has exactly one definition.
func TestAnUnorderableVendorIsRefused(t *testing.T) {
	validator := NewOrderReferenceValidator(
		&stubVendors{found: true, orderable: false}, usableCurrency(2))
	vErrs := ft.NewClientErrors()

	require.NoError(t, validator.PrepareOrder(nil, orderParams("01PARTY", "01USD"), vErrs))

	require.Equal(t, 1, vErrs.Count())
	// Contacts' own reason is carried through rather than restated.
	assert.Equal(t, "vendor.not_orderable", (*vErrs)[0].Key)
	// It points at Purchase's own field, not at Contacts'.
	assert.Equal(t, models.PurchaseOrderFieldVendorId, (*vErrs)[0].Field)
}

func TestAnUnusableCurrencyIsRefused(t *testing.T) {
	validator := NewOrderReferenceValidator(
		usableVendor(""), &stubCurrencies{found: true, usable: false})
	vErrs := ft.NewClientErrors()

	require.NoError(t, validator.PrepareOrder(nil, orderParams("01PARTY", "01OLD"), vErrs))

	require.Equal(t, 1, vErrs.Count())
	assert.Equal(t, "currency.not_active", (*vErrs)[0].Key)
	assert.Equal(t, models.PurchaseOrderFieldCurrencyId, (*vErrs)[0].Field)
}

// The vendor's currency defaults a new order.
func TestTheCurrencyDefaultsFromTheVendor(t *testing.T) {
	validator := NewOrderReferenceValidator(usableVendor("01VND"), usableCurrency(0))
	params := orderParams("01PARTY", "")
	vErrs := ft.NewClientErrors()

	require.NoError(t, validator.PrepareOrder(nil, params, vErrs))

	require.Equal(t, 0, vErrs.Count())
	assert.Equal(t, "01VND", params[models.PurchaseOrderFieldCurrencyId])
}

// An explicit choice wins over the vendor's default, or the order would be silently redenominated.
func TestAnExplicitCurrencyIsNotOverriddenByTheVendorDefault(t *testing.T) {
	validator := NewOrderReferenceValidator(usableVendor("01VND"), usableCurrency(2))
	params := orderParams("01PARTY", "01USD")
	vErrs := ft.NewClientErrors()

	require.NoError(t, validator.PrepareOrder(nil, params, vErrs))

	require.Equal(t, 0, vErrs.Count())
	assert.Equal(t, "01USD", params[models.PurchaseOrderFieldCurrencyId])
}

// A vendor with no stated terms leaves the order without a currency, which is legitimate: the
// schema makes currency_id optional and a draft may not have one yet.
func TestAVendorWithNoTermsLeavesTheCurrencyUnset(t *testing.T) {
	validator := NewOrderReferenceValidator(usableVendor(""), usableCurrency(2))
	params := orderParams("01PARTY", "")
	vErrs := ft.NewClientErrors()

	require.NoError(t, validator.PrepareOrder(nil, params, vErrs))

	assert.Equal(t, 0, vErrs.Count())
	assert.NotContains(t, params, models.PurchaseOrderFieldCurrencyId)
}

// The rounding scale comes from the currency's decimal_places.
func TestScaleForReadsTheCurrencysDecimalPlaces(t *testing.T) {
	testCases := []struct {
		name       string
		currencyId string
		currencies *stubCurrencies
		want       int32
	}{
		{"VND has no fractional digits", "01VND", usableCurrency(0), 0},
		{"USD has two", "01USD", usableCurrency(2), 2},
		{"KWD has three", "01KWD", usableCurrency(3), 3},
		{
			// An order with no currency yet is an ordinary draft, not an error.
			"no currency falls back to two", "", usableCurrency(3), defaultScale,
		},
		{
			// Failing a totals recompute because a currency could not be read would make the money
			// unreadable over a problem that is not about money.
			"an unresolvable currency falls back to two", "01GONE",
			&stubCurrencies{found: false}, defaultScale,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			validator := NewOrderReferenceValidator(usableVendor(""), testCase.currencies)

			assert.Equal(t, testCase.want, validator.ScaleFor(nil, testCase.currencyId))
		})
	}
}

// An inactive currency still rounds: amounts already recorded in it must keep reconciling, and
// reading a scale is not selecting a currency for something new.
func TestAnInactiveCurrencyStillRounds(t *testing.T) {
	validator := NewOrderReferenceValidator(
		usableVendor(""), &stubCurrencies{found: true, usable: false, decimalPlaces: 3})

	assert.Equal(t, int32(3), validator.ScaleFor(nil, "01OLD"))
}

// The scale reaches the arithmetic: a JPY order must not be rounded to two places it does not have.
func TestTheCurrencyScaleReachesTheLineArithmetic(t *testing.T) {
	line := productLine("3", "33.3333", "0", "0")

	assert.True(t, ComputeLineTotals(line, 0).Subtotal.Equal(dec("100")),
		"a zero-decimal currency rounds to whole units")
	assert.True(t, ComputeLineTotals(line, 2).Subtotal.Equal(dec("100.00")))
	assert.True(t, ComputeLineTotals(line, 3).Subtotal.Equal(dec("99.9999").Round(3)))
}

// Unset, the resolver hook leaves every order at the fallback scale, so a deployment that never
// installs it is not silently broken.
func TestTheScaleHookDefaultsToTheFallback(t *testing.T) {
	original := orderScaleResolver
	t.Cleanup(func() { orderScaleResolver = original })

	orderScaleResolver = nil
	assert.Equal(t, defaultScale, scaleForOrder(nil, dmodel.DynamicFields{}))

	SetOrderScaleResolver(func(_ corectx.Context, currencyId string) int32 {
		if currencyId == "01VND" {
			return 0
		}
		return 2
	})
	assert.Equal(t, int32(0), scaleForOrder(nil, dmodel.DynamicFields{
		models.PurchaseOrderFieldCurrencyId: "01VND",
	}))
}
