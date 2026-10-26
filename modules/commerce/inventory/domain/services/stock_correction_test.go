package services

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func correctionRequest(quantity, source, destination string) CorrectionRequest {
	return CorrectionRequest{
		OrgId:                 "org-1",
		ProductVariantId:      "pvar-1",
		Quantity:              decimal.RequireFromString(quantity),
		SourceLocationId:      source,
		DestinationLocationId: destination,
	}
}

func TestCorrectionAcceptsAPositiveQuantityBetweenTwoLocations(t *testing.T) {
	vErrs := assertCorrectable(correctionRequest("3", "loc-stock", "loc-loss"))

	assert.Equal(t, 0, vErrs.Count())
}

func TestCorrectionRefusesANonPositiveQuantity(t *testing.T) {
	// Zero is the case that matters: a variance of zero is a legitimate count result, but it must
	// be handled by not generating a movement at all rather than by generating an empty one.
	// A zero-quantity move would close as `done` having shipped nothing, which reads in the
	// movement history as a correction that happened when none did.
	for _, quantity := range []string{"0", "-5"} {
		vErrs := assertCorrectable(correctionRequest(quantity, "loc-stock", "loc-loss"))

		assert.Equal(t, 1, vErrs.Count(), "quantity %s should be refused", quantity)
	}
}

func TestCorrectionRefusesAMoveToTheSameLocation(t *testing.T) {
	// Both sides equal would decrement and increment the same balance, netting to nothing while
	// still producing a done transfer. The net-zero write is harmless; the misleading document is
	// not.
	vErrs := assertCorrectable(correctionRequest("3", "loc-stock", "loc-stock"))

	assert.Equal(t, 1, vErrs.Count())
}

func TestCorrectionIsCompleteOnlyWhenNothingIsLeftOver(t *testing.T) {
	// With backorder policy forced to `never` a shortfall has nowhere to go, so it must fail the
	// transaction rather than be dropped. Silently applying part of a variance would leave the
	// balance between what it was and what the count said, with no record of the difference.
	assert.NoError(t, assertCorrectionComplete([]moveOutcome{outcome("10", "10")}))
	assert.Error(t, assertCorrectionComplete([]moveOutcome{outcome("10", "7")}))
}

func TestCorrectionCompletenessIgnoresOverDelivery(t *testing.T) {
	// Shortfall() floors at zero, so processing more than demanded is not a shortfall. This pins
	// that assertCorrectionComplete inherits that flooring rather than comparing the two directly.
	assert.NoError(t, assertCorrectionComplete([]moveOutcome{outcome("10", "12")}))
}
