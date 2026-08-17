package services

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/models"
)

// The refund guard rails, one test per rule.
//
// Every one of them is a client error rather than a Go error, because each describes something
// about the caller's request. Answering 500 would tell the caller to retry a request that will
// never be accepted, and would put a gateway refusal in the same bucket as a bug here.

// orderWith builds an order in the state a test needs, without touching a database.
func orderWith(status string, amount string, refunded string) models.Order {
	return *models.NewOrderFrom(dmodel.DynamicFields{
		models.OrderFieldId:           "01ORDER00000000000000000000",
		models.OrderFieldOrderId:      "VDMCMOM0Q8HABCDEFGH",
		models.OrderFieldStatus:       status,
		models.OrderFieldAmount:       decimal.RequireFromString(amount),
		models.OrderFieldRefundAmount: decimal.RequireFromString(refunded),
	})
}

func TestPaidOrderIsRefundable(t *testing.T) {
	vErrs := ft.NewClientErrors()

	ok := assertRefundable(orderWith(models.OrderStatusPaymentSuccess, "150000", "0"),
		decimal.RequireFromString("150000"), vErrs)

	assert.True(t, ok)
	assert.Zero(t, vErrs.Count())
}

func TestAlreadyRefundedOrderIsRefused(t *testing.T) {
	vErrs := ft.NewClientErrors()

	ok := assertRefundable(orderWith(models.OrderStatusRefundSuccess, "150000", "150000"),
		decimal.RequireFromString("50000"), vErrs)

	assert.False(t, ok)
	assert.Equal(t, 1, vErrs.Count())
}

// A previous attempt the gateway refused is a reason to try again, not a reason the order is
// closed. The service this module supersedes treated refund_failed as terminal, which left an
// order stuck whenever the gateway had been briefly unavailable.
func TestAFailedRefundMayBeRetried(t *testing.T) {
	vErrs := ft.NewClientErrors()

	ok := assertRefundable(orderWith(models.OrderStatusRefundFailed, "150000", "0"),
		decimal.RequireFromString("150000"), vErrs)

	assert.True(t, ok)
	assert.Zero(t, vErrs.Count())
}

// A canceled or expired order collected nothing, so there is nothing to give back. Asking the
// gateway anyway gets an opaque refusal that says far less than this does.
func TestCanceledAndExpiredOrdersAreRefused(t *testing.T) {
	for _, status := range []string{models.OrderStatusCanceled, models.OrderStatusExpired} {
		vErrs := ft.NewClientErrors()

		ok := assertRefundable(orderWith(status, "150000", "0"),
			decimal.RequireFromString("150000"), vErrs)

		assert.False(t, ok, status)
		assert.Equal(t, 1, vErrs.Count(), status)
	}
}

// An order still being paid has no settled payment to reverse. Refunding one would return money
// the business has not received, and the payment could still land afterwards.
func TestUnpaidOrdersAreRefused(t *testing.T) {
	for _, status := range []string{
		models.OrderStatusPending,
		models.OrderStatusProcessing,
		models.OrderStatusPaymentFailed,
	} {
		vErrs := ft.NewClientErrors()

		ok := assertRefundable(orderWith(status, "150000", "0"),
			decimal.RequireFromString("150000"), vErrs)

		assert.False(t, ok, status)
		assert.Equal(t, 1, vErrs.Count(), status)
	}
}

func TestZeroAndNegativeRefundsAreRefused(t *testing.T) {
	for _, amount := range []string{"0", "-1", "-150000"} {
		vErrs := ft.NewClientErrors()

		ok := assertRefundable(orderWith(models.OrderStatusPaymentSuccess, "150000", "0"),
			decimal.RequireFromString(amount), vErrs)

		assert.False(t, ok, amount)
		assert.Equal(t, 1, vErrs.Count(), amount)
	}
}

func TestRefundingMoreThanTheOrderIsRefused(t *testing.T) {
	vErrs := ft.NewClientErrors()

	ok := assertRefundable(orderWith(models.OrderStatusPaymentSuccess, "150000", "0"),
		decimal.RequireFromString("150001"), vErrs)

	assert.False(t, ok)
	assert.Equal(t, 1, vErrs.Count())
}

// This is the rule the old service did not have. It compared each refund against the order total
// in isolation, so an order could be refunded for its full amount repeatedly — every one of them
// accepted, and the business paying out several times over.
func TestPartialRefundsAreCountedAgainstTheRunningTotal(t *testing.T) {
	vErrs := ft.NewClientErrors()
	order := orderWith(models.OrderStatusPaymentSuccess, "150000", "100000")

	assert.True(t, assertRefundable(order, decimal.RequireFromString("50000"), vErrs),
		"the remaining 50000 must still be refundable")
	assert.Zero(t, vErrs.Count())

	vErrs = ft.NewClientErrors()
	assert.False(t, assertRefundable(order, decimal.RequireFromString("50001"), vErrs),
		"one unit past the remainder must be refused")
	assert.Equal(t, 1, vErrs.Count())
}

// Refunding exactly what is left closes the order, and must not be caught by the bound above.
func TestRefundingTheExactRemainderIsAllowed(t *testing.T) {
	vErrs := ft.NewClientErrors()

	ok := assertRefundable(orderWith(models.OrderStatusPaymentSuccess, "150000", "149999"),
		decimal.RequireFromString("1"), vErrs)

	assert.True(t, ok)
	assert.Zero(t, vErrs.Count())
}

// The rules are checked in a fixed order so the caller is told the first thing wrong with their
// request, rather than a list whose later entries are consequences of the first.
func TestTheFirstBrokenRuleIsTheOneReported(t *testing.T) {
	vErrs := ft.NewClientErrors()

	// Both the status and the amount are wrong; the status is the one that matters.
	ok := assertRefundable(orderWith(models.OrderStatusRefundSuccess, "150000", "150000"),
		decimal.RequireFromString("999999"), vErrs)

	assert.False(t, ok)
	assert.Equal(t, 1, vErrs.Count())
}
