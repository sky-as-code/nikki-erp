package services

import (
	"testing"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// The fulfilment rules that decide what an order owes; the dispatch path is exercised live.

// THE fulfilment rule. An ACCEPTED request has stock held; only a COMPLETED one has goods moved.
func TestOnlyCompletedRequestsCount(t *testing.T) {
	cases := map[string]bool{
		string(models.SalesFulfillmentStatusCompleted): true,
		string(models.SalesFulfillmentStatusAccepted):  false,
		string(models.SalesFulfillmentStatusPending):   false,
		string(models.SalesFulfillmentStatusRejected):  false,
		string(models.SalesFulfillmentStatusCancelled): false,
	}

	for status, want := range cases {
		request := models.NewSalesFulfillmentRequestFrom(dmodel.DynamicFields{
			models.SalesFulfillmentRequestFieldStatus: status,
		})
		if got := request.IsCompleted(); got != want {
			t.Errorf("status %q counts = %v, want %v", status, got, want)
		}
	}
}

// A request with no status does not count: the field is required, and reading a missing one as
// completed would report goods shipped.
func TestARequestWithNoStatusDoesNotCount(t *testing.T) {
	request := models.NewSalesFulfillmentRequestFrom(dmodel.DynamicFields{})
	if request.IsCompleted() {
		t.Error("a request with no status must not count as completed")
	}
}

// requires_fulfillment defaults to TRUE when absent, deliberately asymmetric: a line wrongly
// needing goods holds an order open, while one wrongly needing none reports a sale as shipped.
func TestAbsentRequiresFulfillmentReadsAsTrue(t *testing.T) {
	cases := map[string]struct {
		record dmodel.DynamicFields
		want   bool
	}{
		"absent":     {dmodel.DynamicFields{}, true},
		"nil":        {dmodel.DynamicFields{models.SalesOrderLineFieldRequiresFulfillment: nil}, true},
		"true":       {dmodel.DynamicFields{models.SalesOrderLineFieldRequiresFulfillment: true}, true},
		"false":      {dmodel.DynamicFields{models.SalesOrderLineFieldRequiresFulfillment: false}, false},
		"wrong type": {dmodel.DynamicFields{models.SalesOrderLineFieldRequiresFulfillment: "no"}, true},
	}

	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			got := boolOrTrue(testCase.record, models.SalesOrderLineFieldRequiresFulfillment)
			if got != testCase.want {
				t.Errorf("%s reads as %v, want %v", name, got, testCase.want)
			}
		})
	}
}

// An order of only non-stocked lines is `not_required`, not pending forever.
func TestAnOrderOfServicesNeedsNoFulfilment(t *testing.T) {
	status := DeriveFulfillmentStatus([]LineQuantities{
		{Ordered: dec("1"), RequiresFulfillment: false},
		{Ordered: dec("2"), RequiresFulfillment: false},
	})

	if status != string(models.SalesOrderFulfillmentStatusNotRequired) {
		t.Errorf("an order owing no goods = %q, want not_required", status)
	}
}

// One stocked line among services still needs fulfilling.
func TestOneStockedLineMakesTheOrderFulfillable(t *testing.T) {
	status := DeriveFulfillmentStatus([]LineQuantities{
		{Ordered: dec("1"), RequiresFulfillment: false},
		{Ordered: dec("2"), Fulfilled: dec("0"), RequiresFulfillment: true},
	})

	if status == string(models.SalesOrderFulfillmentStatusNotRequired) {
		t.Error("an order with one stocked line must not read as not_required")
	}
}
