package services

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

func reservationRow(status string, quantity, consumed, released string, until *time.Time) models.StockReservation {
	fields := dmodel.DynamicFields{
		models.StockReservationFieldStatus:           status,
		models.StockReservationFieldQuantity:         decimal.RequireFromString(quantity),
		models.StockReservationFieldConsumedQuantity: decimal.RequireFromString(consumed),
		models.StockReservationFieldReleasedQuantity: decimal.RequireFromString(released),
	}
	if until != nil {
		fields[models.StockReservationFieldReservedUntil] = model.WrapModelDateTime(*until)
	}
	return *models.NewStockReservationFrom(fields)
}

func TestReservationEffectiveIntervalIsHalfOpen(t *testing.T) {
	deadline := time.Date(2026, 9, 19, 10, 15, 0, 0, time.UTC)

	beforeDeadline := reservationRow("active", "3", "0", "0", &deadline)
	assert.True(t, IsReservationEffective(beforeDeadline, deadline.Add(-time.Second)))
	assert.False(t, IsReservationEffective(beforeDeadline, deadline), "at exactly the deadline it has lapsed")
	assert.False(t, IsReservationEffective(beforeDeadline, deadline.Add(time.Second)))

	noDeadline := reservationRow("active", "3", "0", "0", nil)
	assert.True(t, IsReservationEffective(noDeadline, deadline.Add(365*24*time.Hour)), "null never lapses")

	for _, status := range []string{"consumed", "released"} {
		assert.False(t, IsReservationEffective(reservationRow(status, "3", "3", "0", nil), deadline))
	}
}

func TestEffectiveReservedQuantityIsTheRemainderOnlyWhileInForce(t *testing.T) {
	deadline := time.Date(2026, 9, 19, 10, 15, 0, 0, time.UTC)
	partlyConsumed := reservationRow("active", "6", "2", "0", &deadline)

	assert.True(t, decimal.NewFromInt(4).Equal(EffectiveReservedQuantity(partlyConsumed, deadline.Add(-time.Minute))))
	assert.True(t, EffectiveReservedQuantity(partlyConsumed, deadline).IsZero(), "a lapsed hold commits nothing")
	assert.True(t, decimal.NewFromInt(4).Equal(partlyConsumed.RemainingQuantity()), "the remainder itself stays on record")
}

func TestShortageIsNeverNegative(t *testing.T) {
	availability := WarehouseAvailability{Available: decimal.NewFromInt(4)}

	assert.True(t, decimal.NewFromInt(3).Equal(availability.Shortage(decimal.NewFromInt(7))))
	assert.True(t, availability.Shortage(decimal.NewFromInt(4)).IsZero())
	assert.True(t, availability.Shortage(decimal.NewFromInt(1)).IsZero())
}
