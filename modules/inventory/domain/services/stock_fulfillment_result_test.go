package services

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

func resultItem(sourceItemId string, quantity int64) itStock.FulfillmentResultItem {
	return itStock.FulfillmentResultItem{
		SourceItemId: sourceItemId,
		Quantity:     decimal.NewFromInt(quantity),
	}
}

func wellFormedResult() itStock.FulfillmentResultRequest {
	return itStock.FulfillmentResultRequest{
		EventId:         "evt-1",
		SourceType:      "sales_fulfillment",
		SourceId:        "ful-1",
		SuccessfulItems: []itStock.FulfillmentResultItem{resultItem("item-1", 2)},
	}
}

func TestAFulfillmentResultMustCarryItsEventId(t *testing.T) {
	// The event id is the whole of the idempotency guarantee: without it a redelivered result
	// would consume the goods a second time.
	request := wellFormedResult()
	request.EventId = ""

	refusal := assertFulfillmentResultWellFormed(request)

	require.NotNil(t, refusal)
	assert.Positive(t, refusal.ClientErrors.Count())
}

func TestAFulfillmentResultMustNameItsDemand(t *testing.T) {
	for _, missing := range []string{"type", "id"} {
		request := wellFormedResult()
		if missing == "type" {
			request.SourceType = ""
		} else {
			request.SourceId = ""
		}

		refusal := assertFulfillmentResultWellFormed(request)

		require.NotNil(t, refusal, "a result naming no %s must be refused", missing)
		assert.Positive(t, refusal.ClientErrors.Count())
	}
}

func TestAWellFormedFulfillmentResultIsNotRefused(t *testing.T) {
	assert.Nil(t, assertFulfillmentResultWellFormed(wellFormedResult()))
}

func TestNegativeQuantitiesAreRefusedOnBothLists(t *testing.T) {
	// A negative delivery is not a return: it would subtract from what the report says moved, and
	// the two lists mean opposite things, so both are checked.
	successful := wellFormedResult()
	successful.SuccessfulItems = []itStock.FulfillmentResultItem{resultItem("item-1", -1)}
	assert.NotNil(t, assertFulfillmentResultWellFormed(successful))

	failed := wellFormedResult()
	failed.FailedItems = []itStock.FulfillmentResultItem{resultItem("item-1", -1)}
	assert.NotNil(t, assertFulfillmentResultWellFormed(failed))
}

func TestAResultReportingNothingDeliveredIsWellFormed(t *testing.T) {
	// A total failure is a legitimate report — the machine dispensed none of it — and must reach
	// the ledger so the hold is given back.
	request := wellFormedResult()
	request.SuccessfulItems = nil
	request.FailedItems = []itStock.FulfillmentResultItem{resultItem("item-1", 2)}

	assert.Nil(t, assertFulfillmentResultWellFormed(request))
}

func TestDispensedQuantitiesAreSummedPerSourceItem(t *testing.T) {
	// A device reporting one item across several slots sends several lines for one move.
	request := wellFormedResult()
	request.SuccessfulItems = []itStock.FulfillmentResultItem{
		resultItem("item-1", 2),
		resultItem("item-1", 3),
		resultItem("item-2", 1),
	}

	dispensed := dispensedByItem(request)

	assert.True(t, dispensed["item-1"].Equal(decimal.NewFromInt(5)))
	assert.True(t, dispensed["item-2"].Equal(decimal.NewFromInt(1)))
}

func TestFailedItemsDoNotCountAsDispensed(t *testing.T) {
	// What failed to come out must never be consumed from stock; only the successful list moves
	// goods.
	request := wellFormedResult()
	request.SuccessfulItems = []itStock.FulfillmentResultItem{resultItem("item-1", 1)}
	request.FailedItems = []itStock.FulfillmentResultItem{resultItem("item-1", 4)}

	dispensed := dispensedByItem(request)

	assert.True(t, dispensed["item-1"].Equal(decimal.NewFromInt(1)))
}

func TestAnItemNobodyReportedIsAbsentRatherThanZero(t *testing.T) {
	dispensed := dispensedByItem(wellFormedResult())

	_, present := dispensed["item-unreported"]
	assert.False(t, present)
	// An absent entry still reads as zero, which is what the trim decision needs.
	assert.True(t, dispensed["item-unreported"].IsZero())
}

func TestAMoveThatDeliveredEverythingIsLeftAlone(t *testing.T) {
	trim := decideMoveTrim(decimal.NewFromInt(5), decimal.NewFromInt(5))

	assert.False(t, trim.Rewrite, "a fully delivered move must not have its demand rewritten")
	assert.False(t, trim.Rereserve)
}

func TestAPartiallyDeliveredMoveIsRewrittenAndReReserved(t *testing.T) {
	trim := decideMoveTrim(decimal.NewFromInt(5), decimal.NewFromInt(2))

	assert.True(t, trim.Rewrite)
	assert.True(t, trim.Rereserve, "the delivered part must be re-held so validate consumes it")
}

func TestAMoveThatDeliveredNothingReleasesItsHoldAndTakesNoneBack(t *testing.T) {
	// The goods never left the shelf, so the hold is given back in full and nothing is re-claimed.
	trim := decideMoveTrim(decimal.NewFromInt(5), decimal.Zero)

	assert.True(t, trim.Rewrite)
	assert.False(t, trim.Rereserve)
}

func TestOverDeliveryIsTreatedAsFullDelivery(t *testing.T) {
	// The goods are already in the customer's hands. Refusing the report would leave the ledger
	// believing they are still on the shelf, which is the worse of the two wrong answers.
	trim := decideMoveTrim(decimal.NewFromInt(5), decimal.NewFromInt(7))

	assert.False(t, trim.Rewrite)
	assert.False(t, trim.Rereserve)
}

func TestATrimDecisionHandlesFractionalQuantities(t *testing.T) {
	// Quantities are decimals, not counts: a move measured in kilograms can be partly delivered.
	trim := decideMoveTrim(decimal.RequireFromString("2.5"), decimal.RequireFromString("2.4"))

	assert.True(t, trim.Rewrite)
	assert.True(t, trim.Rereserve)
}

func TestAZeroDemandMoveIsLeftAlone(t *testing.T) {
	// Nothing was asked for, so nothing is owed and there is no hold to rewrite.
	trim := decideMoveTrim(decimal.Zero, decimal.Zero)

	assert.False(t, trim.Rewrite)
}
