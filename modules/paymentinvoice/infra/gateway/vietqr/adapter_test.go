package vietqr

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
	return NewAdapter(Config{
		Username:   testUsername,
		Password:   "test-password",
		SecretKey:  testSecretKey,
		BankCode:   "MB",
		BankNumber: testBankNumber,
		BankName:   "TEST MERCHANT",
	}, nil)
}

// The adapter code is what a payment-method row stores to select this implementation, so it is a
// wire identifier: changing it silently detaches every row that names it.
func TestAdapterCodeIsStable(t *testing.T) {
	assert.Equal(t, "vietqr", newTestAdapter().AdapterCode())
	assert.Equal(t, models.AdapterCodeVietQr, newTestAdapter().AdapterCode())
}

// A bank transfer needs nothing from the merchant at order time: the payer identifies themselves
// to their own bank. mPOS is the adapter where this is not true.
func TestNoOrderTimeInputIsRequired(t *testing.T) {
	request := itGateway.OrderRequest{
		Amount:       decimal.RequireFromString("150000"),
		CurrencyCode: "VND",
	}

	vErrs := &ft.ClientErrors{}
	require.NoError(t, newTestAdapter().ValidateOrder(nil, request, vErrs))
	assert.Zero(t, vErrs.Count())

	metadata, err := newTestAdapter().PrepareMetadata(nil, request)
	require.NoError(t, err)
	assert.Empty(t, metadata)
}

// Rounding would collect something other than what the order says, and the discrepancy would be
// invisible once the transfer had settled.
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

// The gateway rejects an over-long description outright rather than truncating it, which would
// fail a payment over nothing more than a wordy note.
func TestLongDescriptionsAreTrimmedBeforeSending(t *testing.T) {
	trimmed := trimContent("a description considerably longer than the gateway will accept")

	assert.Len(t, trimmed, qrContentMaxLength)
	assert.Equal(t, "a description consider", trimmed[:22])
}

func TestShortDescriptionsAreLeftAlone(t *testing.T) {
	assert.Equal(t, "ORD1234ABCD5", trimContent("ORD1234ABCD5"))
}

// A refund is filed against the original transfer, so without its reference number there is
// nothing to file against.
func TestRefundNeedsTheReferenceNumber(t *testing.T) {
	_, err := newTestAdapter().Refund(nil, itGateway.RefundRequest{
		OrderCode: "ORD1234ABCD5",
		Amount:    decimal.RequireFromString("50000"),
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "reference number")
}
