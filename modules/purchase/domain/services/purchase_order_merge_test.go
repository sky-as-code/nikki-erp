package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"

	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/models"
)

func mergeLine(variantId, uomId, discount string, arrival *time.Time) dmodel.DynamicFields {
	line := dmodel.DynamicFields{
		models.PurchaseOrderLineFieldLineType:         string(models.PurchaseOrderLineTypeProduct),
		models.PurchaseOrderLineFieldProductVariantId: variantId,
		models.PurchaseOrderLineFieldUomId:            uomId,
		models.PurchaseOrderLineFieldDiscountPercent:  dec(discount),
	}
	if arrival != nil {
		line[models.PurchaseOrderLineFieldExpectedArrival] = *arrival
	}
	return line
}

func at(day int, hour int) *time.Time {
	moment := time.Date(2026, 3, day, hour, 0, 0, 0, time.UTC)
	return &moment
}

// The line compatibility rules of §26.1, which are what decide whether two requests for the same
// goods become one line or stay two.
func TestLinesAreMergeable(t *testing.T) {
	testCases := []struct {
		name   string
		target dmodel.DynamicFields
		source dmodel.DynamicFields
		want   bool
	}{
		{
			"same product, unit, discount and arrival",
			mergeLine("01VAR", "01UOM", "0", at(10, 9)),
			mergeLine("01VAR", "01UOM", "0", at(10, 9)),
			true,
		},
		{
			// The window is what makes "Tuesday" and "Tuesday morning" one delivery.
			"arrivals within a day",
			mergeLine("01VAR", "01UOM", "0", at(10, 9)),
			mergeLine("01VAR", "01UOM", "0", at(11, 8)),
			true,
		},
		{
			"arrivals more than a day apart",
			mergeLine("01VAR", "01UOM", "0", at(10, 9)),
			mergeLine("01VAR", "01UOM", "0", at(12, 9)),
			false,
		},
		{
			"different products",
			mergeLine("01VAR", "01UOM", "0", at(10, 9)),
			mergeLine("01OTHER", "01UOM", "0", at(10, 9)),
			false,
		},
		{
			// Ten boxes and ten units are not twenty of anything.
			"different units",
			mergeLine("01VAR", "01BOX", "0", at(10, 9)),
			mergeLine("01VAR", "01UNIT", "0", at(10, 9)),
			false,
		},
		{
			// Two lines at different discounts are at different effective prices, so summing them
			// would produce a quantity at a price neither of them had.
			"different discounts",
			mergeLine("01VAR", "01UOM", "0", at(10, 9)),
			mergeLine("01VAR", "01UOM", "10", at(10, 9)),
			false,
		},
		{
			"neither states an arrival",
			mergeLine("01VAR", "01UOM", "0", nil),
			mergeLine("01VAR", "01UOM", "0", nil),
			true,
		},
		{
			// The request that named a day is asking for something the other is not.
			"one states an arrival and the other does not",
			mergeLine("01VAR", "01UOM", "0", at(10, 9)),
			mergeLine("01VAR", "01UOM", "0", nil),
			false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.want, linesAreMergeable(testCase.target, testCase.source))
		})
	}
}

// Two free-text lines have no product to compare, so "same product" would be vacuously true and two
// unrelated charges would be summed into one.
func TestFreeTextLinesAreNeverMerged(t *testing.T) {
	freight := mergeLine("", "01UOM", "0", nil)
	handling := mergeLine("", "01UOM", "0", nil)

	assert.False(t, linesAreMergeable(freight, handling))
}

// A section or note is document structure, not a quantity. Merging two headings would drop one.
func TestStructuralLinesAreNeverMerged(t *testing.T) {
	for _, lineType := range []models.PurchaseOrderLineType{
		models.PurchaseOrderLineTypeSection,
		models.PurchaseOrderLineTypeSubsection,
		models.PurchaseOrderLineTypeNote,
	} {
		t.Run(string(lineType), func(t *testing.T) {
			target := mergeLine("01VAR", "01UOM", "0", nil)
			source := mergeLine("01VAR", "01UOM", "0", nil)
			source[models.PurchaseOrderLineFieldLineType] = string(lineType)

			assert.Nil(t, findMergeableLine([]dmodel.DynamicFields{target}, source))

			target[models.PurchaseOrderLineFieldLineType] = string(lineType)
			source[models.PurchaseOrderLineFieldLineType] = string(models.PurchaseOrderLineTypeProduct)
			assert.Nil(t, findMergeableLine([]dmodel.DynamicFields{target}, source))
		})
	}
}

func orderFor(code, vendorId, currencyId, agreementId string, deadline *time.Time) dmodel.DynamicFields {
	order := dmodel.DynamicFields{
		models.PurchaseOrderFieldId:          "id-" + code,
		models.PurchaseOrderFieldCode:        code,
		models.PurchaseOrderFieldVendorId:    vendorId,
		models.PurchaseOrderFieldCurrencyId:  currencyId,
		models.PurchaseOrderFieldAgreementId: agreementId,
	}
	if deadline != nil {
		order[models.PurchaseOrderFieldOrderDeadline] = *deadline
	}
	return order
}

// §26: merging must not cross vendor, currency or agreement. Each would produce a document that
// commits to something none of the sources did.
func TestIncompatibleOrdersAreRefused(t *testing.T) {
	testCases := []struct {
		name    string
		second  dmodel.DynamicFields
		wantKey string
	}{
		{
			"different vendors", orderFor("PO-2", "01OTHER", "01USD", "", nil),
			"purchase_order.merge_vendor_mismatch",
		},
		{
			"different currencies", orderFor("PO-2", "01VENDOR", "01EUR", "", nil),
			"purchase_order.merge_currency_mismatch",
		},
		{
			"different agreements", orderFor("PO-2", "01VENDOR", "01USD", "01AGR", nil),
			"purchase_order.merge_agreement_mismatch",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			orders := []dmodel.DynamicFields{
				orderFor("PO-1", "01VENDOR", "01USD", "", nil),
				testCase.second,
			}

			refusal := assertCompatibleOrders(orders)

			require.NotNil(t, refusal)
			require.Equal(t, 1, refusal.ClientErrors.Count())
			assert.Equal(t, testCase.wantKey, refusal.ClientErrors[0].Key)
		})
	}
}

func TestCompatibleOrdersPass(t *testing.T) {
	orders := []dmodel.DynamicFields{
		orderFor("PO-1", "01VENDOR", "01USD", "01AGR", nil),
		orderFor("PO-2", "01VENDOR", "01USD", "01AGR", nil),
		orderFor("PO-3", "01VENDOR", "01USD", "01AGR", nil),
	}

	assert.Nil(t, assertCompatibleOrders(orders))
}

// The oldest order is the target and keeps its code, so the vendor's own paperwork still matches.
func TestTheOldestOrderIsTheMergeTarget(t *testing.T) {
	t.Run("by deadline", func(t *testing.T) {
		orders := []dmodel.DynamicFields{
			orderFor("PO-B", "01V", "01C", "", at(12, 0)),
			orderFor("PO-A", "01V", "01C", "", at(10, 0)),
			orderFor("PO-C", "01V", "01C", "", at(15, 0)),
		}

		target, sources := splitMergeTarget(orders)

		assert.Equal(t, "PO-A", stringOf(target, models.PurchaseOrderFieldCode))
		assert.Len(t, sources, 2)
	})

	t.Run("a stated deadline beats none", func(t *testing.T) {
		orders := []dmodel.DynamicFields{
			orderFor("PO-A", "01V", "01C", "", nil),
			orderFor("PO-B", "01V", "01C", "", at(20, 0)),
		}

		target, _ := splitMergeTarget(orders)

		assert.Equal(t, "PO-B", stringOf(target, models.PurchaseOrderFieldCode))
	})

	t.Run("with no deadlines, the code decides", func(t *testing.T) {
		// The code carries a ULID, which sorts by creation time, so this still means oldest.
		orders := []dmodel.DynamicFields{
			orderFor("PO-02", "01V", "01C", "", nil),
			orderFor("PO-01", "01V", "01C", "", nil),
		}

		target, _ := splitMergeTarget(orders)

		assert.Equal(t, "PO-01", stringOf(target, models.PurchaseOrderFieldCode))
	})
}

// timeOf must distinguish "no date" from "the zero time", or every dateless line would look like it
// wanted delivery in year one — and every pair of them would merge on that basis.
func TestTimeOfDistinguishesAbsentFromZero(t *testing.T) {
	moment := time.Date(2026, 3, 10, 9, 0, 0, 0, time.UTC)

	testCases := []struct {
		name    string
		value   any
		wantHas bool
	}{
		{"a time", moment, true},
		{"a pointer to a time", &moment, true},
		{"an RFC3339 string", "2026-03-10T09:00:00Z", true},
		{"absent", nil, false},
		{"the zero time", time.Time{}, false},
		{"a nil pointer", (*time.Time)(nil), false},
		{"an unparseable string", "last Tuesday", false},
		{"the wrong type entirely", 42, false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			fields := dmodel.DynamicFields{"when": testCase.value}

			_, has := timeOf(fields, "when")

			assert.Equal(t, testCase.wantHas, has)
		})
	}

	_, has := timeOf(dmodel.DynamicFields{}, "missing")
	assert.False(t, has)
}
