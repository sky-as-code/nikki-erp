package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

func expiringOrder(paymentStatus string, validUntil *time.Time, expiredAt *time.Time) dmodel.DynamicFields {
	record := dmodel.DynamicFields{
		models.SalesOrderFieldStatus:        string(models.SalesOrderStatusConfirmed),
		models.SalesOrderFieldPaymentStatus: paymentStatus,
	}
	if validUntil != nil {
		record[models.SalesOrderFieldValidUntil] = model.WrapModelDateTime(*validUntil)
	}
	if expiredAt != nil {
		record[models.SalesOrderFieldExpiredAt] = model.WrapModelDateTime(*expiredAt)
	}
	return record
}

// AC-19, AC-20, AC-24: an unpaid or partly paid order is expired the instant the deadline passes;
// a paid one never is, and a refund afterwards does not make it so (AC-21, §7.3).
func TestOrderExpiryIsDecidedOnTheClockAndThePaidLatch(t *testing.T) {
	deadline := time.Date(2026, 9, 19, 10, 15, 0, 0, time.UTC)
	before, at := deadline.Add(-time.Second), deadline

	unpaid := expiringOrder(string(models.SalesOrderPaymentStatusUnpaid), &deadline, nil)
	assert.False(t, IsOrderExpired(unpaid, before))
	assert.True(t, IsOrderExpired(unpaid, at), "at exactly the deadline the order has expired")

	partial := expiringOrder(string(models.SalesOrderPaymentStatusPartiallyPaid), &deadline, nil)
	assert.True(t, IsOrderExpired(partial, at), "a partial payment does not save the order")

	for _, paid := range []string{
		string(models.SalesOrderPaymentStatusPaid),
		string(models.SalesOrderPaymentStatusOverpaid),
		string(models.SalesOrderPaymentStatusRefunded),
		string(models.SalesOrderPaymentStatusPartiallyRefunded),
	} {
		assert.Falsef(t, IsOrderExpired(expiringOrder(paid, &deadline, nil), at.Add(time.Hour)),
			"%s is behind the paid latch", paid)
	}

	noDeadline := expiringOrder(string(models.SalesOrderPaymentStatusUnpaid), nil, nil)
	assert.False(t, IsOrderExpired(noDeadline, at.Add(365*24*time.Hour)), "null never expires")

	stamped := expiringOrder(string(models.SalesOrderPaymentStatusUnpaid), &deadline, &deadline)
	assert.True(t, IsOrderExpired(stamped, before), "an expiry already written cannot be undone by the clock")
}

func TestValidUntilMustLieInTheFuture(t *testing.T) {
	now := time.Date(2026, 9, 19, 10, 0, 0, 0, time.UTC)
	past, future := now.Add(-time.Minute), now.Add(time.Minute)

	assert.Nil(t, assertValidUntilInFuture(nil, now))
	assert.Nil(t, assertValidUntilInFuture(&future, now))
	assert.NotNil(t, assertValidUntilInFuture(&past, now))
	assert.NotNil(t, assertValidUntilInFuture(&now, now), "a deadline that is now is already gone")
}

// The hold lapses exactly when the order can no longer be paid: the deadline is passed through
// with no grace, and only an order without one falls back to the method's TTL.
func TestHoldDeadlineFollowsTheOrderBeforeTheMethod(t *testing.T) {
	now := time.Date(2026, 9, 19, 10, 0, 0, 0, time.UTC)
	validUntil := model.WrapModelDateTime(now.Add(15 * time.Minute))
	ttl := int32(1440)

	assert.Equal(t, &validUntil, holdDeadline(&validUntil, &ttl, now))
	assert.Equal(t, now.Add(24*time.Hour), holdDeadline(nil, &ttl, now).GoTime())
	assert.Nil(t, holdDeadline(nil, nil, now))
}

// AC-17, AC-18: a paid order is cancellable exactly when its own snapshot says so, or when the
// system cancels it; the channel's current setting plays no part.
func TestPaidOrderCancellationFollowsTheOrderSnapshot(t *testing.T) {
	paid := func(autoConfirm bool) dmodel.DynamicFields {
		return dmodel.DynamicFields{
			models.SalesOrderFieldStatus:            string(models.SalesOrderStatusConfirmed),
			models.SalesOrderFieldPaymentStatus:     string(models.SalesOrderPaymentStatusPaid),
			models.SalesOrderFieldFulfillmentStatus: string(models.SalesOrderFulfillmentStatusPending),
			models.SalesOrderFieldAutoConfirmOrder:  autoConfirm,
		}
	}

	assert.NotNil(t, assertCancellableWith(paid(false), CancelOrderOptions{}), "the manual channel keeps the refund workflow")
	assert.Nil(t, assertCancellableWith(paid(true), CancelOrderOptions{}), "the snapshot admits the cancel")
	assert.Nil(t, assertCancellableWith(paid(false), CancelOrderOptions{SystemInitiated: true}),
		"the system may cancel a paid order whose stock is gone")

	delivered := paid(true)
	delivered[models.SalesOrderFieldFulfillmentStatus] = string(models.SalesOrderFulfillmentStatusFulfilled)
	assert.NotNil(t, assertCancellableWith(delivered, CancelOrderOptions{}), "delivered goods go through a return")

	partly := paid(true)
	partly[models.SalesOrderFieldFulfillmentStatus] = string(models.SalesOrderFulfillmentStatusPartiallyFulfilled)
	assert.Nil(t, assertCancellableWith(partly, CancelOrderOptions{}), "the undelivered remainder may be cancelled (AC-27)")
}

func TestWarehouseHoldKeyIsPerDemandRevision(t *testing.T) {
	assert.NotEqual(t, warehouseHoldKey("ful-1", 1), warehouseHoldKey("ful-1", 2))
	assert.Equal(t, "fulfillment:ful-1:rev:12", warehouseHoldKey("ful-1", 12))
}
