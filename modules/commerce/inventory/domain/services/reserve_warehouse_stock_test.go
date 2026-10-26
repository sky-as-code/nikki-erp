package services

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

func reserveRequest() itStock.ReserveWarehouseStockRequest {
	return itStock.ReserveWarehouseStockRequest{
		OrgId:          "org-1",
		WarehouseId:    "wh-1",
		SourceModule:   "sales",
		SourceType:     "sales_order_fulfillment",
		SourceId:       "ful-1",
		SourceRevision: 1,
		IdempotencyKey: "key-1",
		Lines: []itStock.WarehouseReservationLine{
			{SourceLineId: "L1", ProductVariantId: "var-1", Quantity: decimal.NewFromInt(3)},
			{SourceLineId: "L2", ProductVariantId: "var-2", Quantity: decimal.NewFromInt(1)},
		},
	}
}

func TestReserveRequestRefusesWhatCannotHoldAnything(t *testing.T) {
	assert.Zero(t, assertReserveRequestWellFormed(reserveRequest()).Count())

	noLines := reserveRequest()
	noLines.Lines = nil
	assert.Positive(t, assertReserveRequestWellFormed(noLines).Count())

	zero := reserveRequest()
	zero.Lines[0].Quantity = decimal.Zero
	assert.Positive(t, assertReserveRequestWellFormed(zero).Count(), "a hold of nothing is refused")

	noKey := reserveRequest()
	noKey.IdempotencyKey = ""
	assert.Positive(t, assertReserveRequestWellFormed(noKey).Count())

	anonymous := reserveRequest()
	anonymous.SourceModule = ""
	assert.Positive(t, assertReserveRequestWellFormed(anonymous).Count(), "a hold nobody owns is unfindable")

	twice := reserveRequest()
	twice.Lines[1].SourceLineId = "L1"
	assert.Positive(t, assertReserveRequestWellFormed(twice).Count())
}

func TestReservationFingerprintIgnoresLineOrderAndSeesEveryChange(t *testing.T) {
	request := reserveRequest()
	lines := []normalisedLine{
		{SourceLineId: "L1", ProductVariantId: "var-1", BaseUomId: "uom-1", Quantity: decimal.NewFromInt(3)},
		{SourceLineId: "L2", ProductVariantId: "var-2", BaseUomId: "uom-1", Quantity: decimal.NewFromInt(1)},
	}
	reversed := []normalisedLine{lines[1], lines[0]}

	assert.Equal(t, reservationFingerprint(request, lines), reservationFingerprint(request, reversed))

	moreOfOne := []normalisedLine{lines[0], {SourceLineId: "L2", ProductVariantId: "var-2", BaseUomId: "uom-1", Quantity: decimal.NewFromInt(2)}}
	assert.NotEqual(t, reservationFingerprint(request, lines), reservationFingerprint(request, moreOfOne))

	later := reserveRequest()
	deadline := time.Date(2026, 9, 19, 10, 15, 0, 0, time.UTC)
	later.ReservedUntil = &deadline
	assert.NotEqual(t, reservationFingerprint(request, lines), reservationFingerprint(later, lines))
}

func storedReservation(key, fingerprint, status string) models.StockReservation {
	return *models.NewStockReservationFrom(dmodel.DynamicFields{
		models.StockReservationFieldId:                 "rsv-1",
		models.StockReservationFieldProductVariantId:   "var-1",
		models.StockReservationFieldSourceLineId:       "L1",
		models.StockReservationFieldQuantity:           decimal.NewFromInt(3),
		models.StockReservationFieldConsumedQuantity:   decimal.NewFromInt(1),
		models.StockReservationFieldReleasedQuantity:   decimal.Zero,
		models.StockReservationFieldStatus:             status,
		models.StockReservationFieldIdempotencyKey:     key,
		models.StockReservationFieldRequestFingerprint: fingerprint,
	})
}

func TestReplayAnswersFromStoredRowsOrRefusesADifferentRequest(t *testing.T) {
	now := time.Date(2026, 9, 19, 10, 0, 0, 0, time.UTC)
	request := reserveRequest()

	replayed := replayReservation([]models.StockReservation{storedReservation("key-1", "fp", "active")}, request, "fp", now)
	assert.True(t, replayed.Replayed)
	assert.False(t, replayed.Refused())
	require.Len(t, replayed.Reservations, 1)
	assert.Equal(t, model.Id("rsv-1"), replayed.Reservations[0].ReservationId)
	assert.True(t, decimal.NewFromInt(2).Equal(replayed.Reservations[0].RemainingQuantity), "current figures, not the original promise")

	otherPayload := replayReservation([]models.StockReservation{storedReservation("key-1", "fp", "active")}, request, "fp-2", now)
	assert.True(t, otherPayload.Refused(), "same key, different payload is a conflict")

	otherKey := replayReservation([]models.StockReservation{storedReservation("key-2", "fp", "active")}, request, "fp", now)
	assert.True(t, otherKey.Refused(), "a new request for the same demand revision is a conflict")

	released := replayReservation([]models.StockReservation{storedReservation("key-1", "fp", "released")}, request, "fp", now)
	assert.Equal(t, "released", released.Reservations[0].EffectiveStatus, "a replay never revives a released hold")
}

func TestEffectiveStatusOfReportsExpiryWithoutStoringIt(t *testing.T) {
	deadline := time.Date(2026, 9, 19, 10, 15, 0, 0, time.UTC)
	active := reservationRow("active", "3", "0", "0", &deadline)

	assert.Equal(t, "active", EffectiveStatusOf(active, deadline.Add(-time.Second)))
	assert.Equal(t, "expired", EffectiveStatusOf(active, deadline))
	assert.Equal(t, "active", *active.GetStatus(), "the stored status is untouched")
	assert.Equal(t, "consumed", EffectiveStatusOf(reservationRow("consumed", "3", "3", "0", &deadline), deadline))
}
