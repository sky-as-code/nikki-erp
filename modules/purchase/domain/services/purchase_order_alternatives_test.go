package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"

	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/models"
)

func alternativeOrder(code, vendorId, currencyId, total string) dmodel.DynamicFields {
	return dmodel.DynamicFields{
		models.PurchaseOrderFieldId:          "id-" + code,
		models.PurchaseOrderFieldCode:        code,
		models.PurchaseOrderFieldVendorId:    vendorId,
		models.PurchaseOrderFieldCurrencyId:  currencyId,
		models.PurchaseOrderFieldStatus:      string(models.PurchaseOrderStatusRfqSent),
		models.PurchaseOrderFieldTotalAmount: dec(total),
	}
}

// The comparison names the cheapest quote.
func TestTheCheapestAlternativeIsMarked(t *testing.T) {
	comparison := buildComparison([]dmodel.DynamicFields{
		alternativeOrder("PO-A", "01V1", "01USD", "1200"),
		alternativeOrder("PO-B", "01V2", "01USD", "950"),
		alternativeOrder("PO-C", "01V3", "01USD", "1100"),
	})

	require.True(t, comparison.ComparableByPrice)
	require.Len(t, comparison.Alternatives, 3)

	cheapest := []string{}
	for _, row := range comparison.Alternatives {
		if row.IsCheapest {
			cheapest = append(cheapest, row.Code)
		}
	}
	assert.Equal(t, []string{"PO-B"}, cheapest, "exactly one alternative is the cheapest")
}

// No exchange rate model exists, so quotes in different currencies cannot be ranked: comparing 100
// USD against 100 VND by their numbers would name the wrong winner with complete confidence.
func TestAlternativesInDifferentCurrenciesAreNotRankedByPrice(t *testing.T) {
	comparison := buildComparison([]dmodel.DynamicFields{
		alternativeOrder("PO-A", "01V1", "01USD", "100"),
		alternativeOrder("PO-B", "01V2", "01VND", "100"),
	})

	assert.False(t, comparison.ComparableByPrice)
	for _, row := range comparison.Alternatives {
		assert.False(t, row.IsCheapest, "nothing is marked cheapest when prices cannot be compared")
	}
}

// A tie marks one row rather than both, and the first of the tied rows wins so the same comparison
// asked twice names the same vendor.
func TestATieMarksTheFirstAlternativeOnly(t *testing.T) {
	comparison := buildComparison([]dmodel.DynamicFields{
		alternativeOrder("PO-A", "01V1", "01USD", "500"),
		alternativeOrder("PO-B", "01V2", "01USD", "500"),
		alternativeOrder("PO-C", "01V3", "01USD", "500"),
	})

	marked := []string{}
	for _, row := range comparison.Alternatives {
		if row.IsCheapest {
			marked = append(marked, row.Code)
		}
	}
	assert.Equal(t, []string{"PO-A"}, marked)
}

func TestAnEmptyComparisonIsNotAnError(t *testing.T) {
	comparison := buildComparison(nil)

	assert.Empty(t, comparison.Alternatives)
	assert.True(t, comparison.ComparableByPrice)
}

// Alternatives are raised while the requirement is still being quoted; once confirmed, the decision
// has been made.
func TestOnlyQuotableOrdersMayRaiseAlternatives(t *testing.T) {
	testCases := []struct {
		status models.PurchaseOrderStatus
		want   bool
	}{
		{models.PurchaseOrderStatusRfq, true},
		{models.PurchaseOrderStatusRfqSent, true},
		{models.PurchaseOrderStatusToApprove, false},
		{models.PurchaseOrderStatusPurchaseOrder, false},
		{models.PurchaseOrderStatusCancelled, false},
	}

	for _, testCase := range testCases {
		t.Run(string(testCase.status), func(t *testing.T) {
			assert.Equal(t, testCase.want, isQuotableStatus(string(testCase.status)))
		})
	}
}

// The warning names the codes rather than only counting them, because the buyer's decision depends
// on which vendors are still being asked.
func TestTheOpenAlternativesWarningNamesTheCodes(t *testing.T) {
	open := []dmodel.DynamicFields{
		alternativeOrder("PO-B", "01V2", "01USD", "950"),
		alternativeOrder("PO-C", "01V3", "01USD", "1100"),
	}

	result := alternativesWarningResult(open)

	require.Equal(t, 1, result.ClientErrors.Count())
	violation := result.ClientErrors[0]
	assert.Equal(t, "purchase_order.open_alternatives", violation.Key)
	assert.Contains(t, violation.Message, "PO-B")
	assert.Contains(t, violation.Message, "PO-C")

	// The machine-readable form carries both the codes and the two answers the caller may give, so a
	// client does not have to parse the sentence to build the prompt.
	require.NotNil(t, violation.Vars)
	assert.Equal(t, []string{"PO-B", "PO-C"}, violation.Vars["alternative_codes"])
	assert.Equal(t,
		[]string{AlternativeChoiceKeep, AlternativeChoiceCancel},
		violation.Vars["choices"])
}

// The two choices are the ones the confirm path accepts, so a client echoing one back cannot be
// refused for using a value the server offered it.
func TestTheOfferedChoicesAreTheAcceptedOnes(t *testing.T) {
	violation := alternativesWarningResult(
		[]dmodel.DynamicFields{alternativeOrder("PO-B", "01V", "01USD", "1")}).ClientErrors[0]

	offered, ok := violation.Vars["choices"].([]string)
	require.True(t, ok)

	for _, choice := range offered {
		assert.True(t, choice == AlternativeChoiceKeep || choice == AlternativeChoiceCancel,
			"the warning offers %q, which Confirm does not accept", choice)
	}
}
