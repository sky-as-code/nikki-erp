package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func returnable(t *testing.T, completed, alreadyReturned string) ReturnableLine {
	return ReturnableLine{
		MoveId:    "mv-1",
		Completed: dec(t, completed),

		AlreadyReturned: dec(t, alreadyReturned),
	}
}

func TestReturnableIsCompletedMinusAlreadyReturned(t *testing.T) {
	assert.Equal(t, "80", returnable(t, "80", "0").Returnable().String())
	assert.Equal(t, "50", returnable(t, "80", "30").Returnable().String())
	assert.True(t, returnable(t, "80", "80").Returnable().IsZero())
}

func TestReturnableComesFromWhatShippedNotWhatWasDemanded(t *testing.T) {
	// BR §4.2.10.3's own example: a transfer of 100 that shipped only 80 has 80 returnable, because
	// the other 20 never left the building. Computing from demand_quantity would let a customer
	// return goods they were never sent, which is the failure this pins.
	partialShipment := returnable(t, "80", "0")

	assert.Equal(t, "80", partialShipment.Returnable().String())
}

func TestPartialReturnOfAPartialShipment(t *testing.T) {
	// The case that catches sign and source errors: 80 shipped of 100 demanded, 30 already back,
	// so 50 remains. Getting either operand wrong yields a plausible number here, which is why it
	// is a table row rather than left to inspection.
	assert.Equal(t, "50", returnable(t, "80", "30").Returnable().String())
}

func TestReturnableNeverGoesNegative(t *testing.T) {
	// More returned than shipped is a data problem, not a licence to send stock back out.
	assert.True(t, returnable(t, "80", "100").Returnable().IsZero())
}

func TestReturnIsRefusedBeyondTheReturnable(t *testing.T) {
	line := returnable(t, "80", "30")

	assert.Equal(t, 0, AssertReturnable(line, dec(t, "50")).Count())
	assert.Equal(t, 1, AssertReturnable(line, dec(t, "51")).Count())
}

func TestReturnIsRefusedForANonPositiveQuantity(t *testing.T) {
	line := returnable(t, "80", "0")

	assert.Equal(t, 1, AssertReturnable(line, dec(t, "0")).Count())
	assert.Equal(t, 1, AssertReturnable(line, dec(t, "-5")).Count())
}

func TestFullyReturnedMoveAcceptsNothingFurther(t *testing.T) {
	line := returnable(t, "80", "80")

	assert.Equal(t, 1, AssertReturnable(line, dec(t, "1")).Count())
}

func TestTotalReturnableSumsEveryLine(t *testing.T) {
	lines := []ReturnableLine{returnable(t, "80", "30"), returnable(t, "20", "0")}

	assert.Equal(t, "70", TotalReturnable(lines).String())
}

func TestTotalReturnableIsZeroOnceEverythingIsBack(t *testing.T) {
	lines := []ReturnableLine{returnable(t, "80", "80"), returnable(t, "20", "20")}

	assert.True(t, TotalReturnable(lines).IsZero())
}

func TestReverseOperationCodeInvertsDirection(t *testing.T) {
	// A return of a delivery is a receipt: the goods travel the other way. Getting this wrong
	// produces a transfer that moves stock in the direction it already went.
	assert.Equal(t, "incoming", reverseOperationCode("outgoing"))
	assert.Equal(t, "outgoing", reverseOperationCode("incoming"))
}

func TestReverseOfAnInternalTransferIsStillInternal(t *testing.T) {
	// Both ends are the company's own locations, so there is no direction to invert.
	assert.Equal(t, "internal", reverseOperationCode("internal"))
}

func TestDefaultFullReturnTakesBackEverythingRemaining(t *testing.T) {
	// What the contextual action sends: no per-line payload, so the whole remaining quantity is
	// proposed and the user trims the draft afterwards (F9).
	lines := []ReturnableLine{returnable(t, "80", "30"), returnable(t, "20", "0")}

	resolved := defaultFullReturn(lines)

	assert.Len(t, resolved, 2)
	assert.Equal(t, "50", resolved[0].Quantity.String())
	assert.Equal(t, "20", resolved[1].Quantity.String())
}

func TestDefaultFullReturnSkipsFullyReturnedLines(t *testing.T) {
	// A line with nothing left must not become a zero-quantity move, which would validate to
	// nothing while still appearing in the document as if something came back.
	lines := []ReturnableLine{returnable(t, "80", "80"), returnable(t, "20", "5")}

	resolved := defaultFullReturn(lines)

	assert.Len(t, resolved, 1)
	assert.Equal(t, "15", resolved[0].Quantity.String())
}

func TestRequestedReturnIsRefusedForAnUnknownMove(t *testing.T) {
	lines := []ReturnableLine{returnable(t, "80", "0")}

	_, vErrs := resolveRequestedReturns(lines, ReturnRequest{
		Lines: []ReturnLineRequest{{MoveId: "mv-other", Quantity: dec(t, "1")}},
	})

	assert.Equal(t, 1, vErrs.Count())
}

func TestRequestedReturnIsCappedWithNoOverride(t *testing.T) {
	// AC-STOCK-022. There is deliberately no waiver parameter to pass here — the cap can only be
	// lifted by changing this code, not by anything a caller sends.
	lines := []ReturnableLine{returnable(t, "80", "30")}

	_, vErrs := resolveRequestedReturns(lines, ReturnRequest{
		Lines: []ReturnLineRequest{{MoveId: "mv-1", Quantity: dec(t, "51")}},
	})

	assert.Equal(t, 1, vErrs.Count())
}
