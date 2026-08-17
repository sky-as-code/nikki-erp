package mpos

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/models"
	itGateway "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/gateway"
)

func newTestAdapter() *Adapter {
	return NewAdapter(Config{MerchantId: "M1", SecretKey: testSecretKey}, nil)
}

// The adapter code is what a payment-method row stores to select this implementation, so it is a
// wire identifier: changing it silently detaches every row that names it.
func TestAdapterCodeIsStable(t *testing.T) {
	assert.Equal(t, "mpos", newTestAdapter().AdapterCode())
	assert.Equal(t, models.AdapterCodeMpos, newTestAdapter().AdapterCode())
}

// This is the case the port's ValidateOrder exists for. A card-terminal payment that does not say
// which terminal is meaningless, and catching it here means the caller is told which field is
// missing before anything is written or any money asked for.
func TestOrderWithoutATerminalIsRejected(t *testing.T) {
	vErrs := &ft.ClientErrors{}

	require.NoError(t, newTestAdapter().ValidateOrder(nil, itGateway.OrderRequest{
		Amount:       decimal.RequireFromString("150000"),
		CurrencyCode: "VND",
	}, vErrs))

	assert.Equal(t, 1, vErrs.Count())
	assert.True(t, vErrs.Has(models.OrderFieldMetadata+"."+models.OrderMetaPosId))
}

func TestOrderWithATerminalIsAccepted(t *testing.T) {
	vErrs := &ft.ClientErrors{}

	require.NoError(t, newTestAdapter().ValidateOrder(nil, itGateway.OrderRequest{
		Amount:       decimal.RequireFromString("150000"),
		CurrencyCode: "VND",
		Metadata:     map[string]any{models.OrderMetaPosId: "POS-0001"},
	}, vErrs))

	assert.Zero(t, vErrs.Count())
}

// The terminal id is kept on the order because the watchdog needs it later to ask the gateway
// about an order that never received a callback.
func TestTerminalIdIsKeptOnTheOrder(t *testing.T) {
	metadata, err := newTestAdapter().PrepareMetadata(nil, itGateway.OrderRequest{
		Metadata: map[string]any{models.OrderMetaPosId: "POS-0001"},
	})

	require.NoError(t, err)
	assert.Equal(t, "POS-0001", metadata[models.OrderMetaPosId])
}

// Three statuses mean the money was taken, and the distinction between them is about settlement
// rather than outcome. The service this replaces compared only against APPROVED, which left a
// settled QR payment — status 104 — looking unpaid, expiring an order the customer had paid for.
func TestEverySuccessfulStatusCountsAsPaid(t *testing.T) {
	paid := []int{TransStatusApproved, TransStatusSettled, TransStatusPendingSignature}
	for _, status := range paid {
		assert.Truef(t, isPaid(status), "status %d means the money was taken", status)
		assert.Truef(t, isSettled(status), "status %d is a verdict", status)
	}

	notPaid := []int{
		TransStatusRejected, TransStatusFailed, TransStatusReversed,
		TransStatusVoided, TransStatusRefund,
	}
	for _, status := range notPaid {
		assert.Falsef(t, isPaid(status), "status %d does not mean the money was taken", status)
		assert.Truef(t, isSettled(status), "status %d is still a verdict", status)
	}
}

// A pending transaction is the customer still standing at the terminal. Reporting it as unpaid
// would expire an order that is seconds away from succeeding.
func TestPendingIsNotAVerdict(t *testing.T) {
	assert.False(t, isSettled(TransStatusPending))
	assert.False(t, isPaid(TransStatusPending))
}

func TestWebhookOutcomeReadsTheStatus(t *testing.T) {
	settled, paid := WebhookOutcome(WebhookPayload{TransStatus: TransStatusSettled})
	assert.True(t, settled)
	assert.True(t, paid)

	settled, paid = WebhookOutcome(WebhookPayload{TransStatus: TransStatusRejected})
	assert.True(t, settled)
	assert.False(t, paid)

	settled, paid = WebhookOutcome(WebhookPayload{TransStatus: TransStatusPending})
	assert.False(t, settled)
	assert.False(t, paid)
}

// Rounding would charge the customer standing at the terminal something other than what the order
// says, and the discrepancy would be invisible.
func TestFractionalAmountsAreRefused(t *testing.T) {
	_, err := toWholeUnits(decimal.RequireFromString("1500.75"))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "whole amounts only")
}

func TestWholeAmountsConvert(t *testing.T) {
	amount, err := toWholeUnits(decimal.RequireFromString("150000.000000"))

	require.NoError(t, err)
	assert.Equal(t, int64(150000), amount)
}

// Asking the gateway about an order it cannot identify would produce a meaningless answer, so the
// adapter refuses rather than sending a query with an empty terminal.
func TestCheckOrderNeedsTheTerminal(t *testing.T) {
	_, err := newTestAdapter().CheckOrder(nil, itGateway.CheckOrderRequest{
		OrderCode: "ORD1234ABCD5",
		Amount:    decimal.RequireFromString("150000"),
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "names no terminal")
}

// A refund has to name the payment it reverses; the gateway knows that payment by its own
// transaction code, which was recorded when the payment settled.
func TestRefundNeedsTheGatewayTransactionCode(t *testing.T) {
	_, err := newTestAdapter().Refund(nil, itGateway.RefundRequest{
		OrderCode: "ORD1234ABCD5",
		Amount:    decimal.RequireFromString("50000"),
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "transaction code")
}
