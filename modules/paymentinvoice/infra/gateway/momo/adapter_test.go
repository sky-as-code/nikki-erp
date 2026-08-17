package momo

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/models"
	itGateway "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/gateway"
)

// MoMo deals in whole dong. Amounts are decimal because how many minor units a currency has is a
// property of the currency, so the narrowing happens here, at the boundary with the gateway that
// needs it.
func TestWholeAmountsConvert(t *testing.T) {
	amount, err := toWholeUnits(decimal.RequireFromString("150000"))

	require.NoError(t, err)
	assert.Equal(t, int64(150000), amount)
}

// Rounding here would charge the payer something they did not agree to, and the difference would
// be invisible: the order would say one figure and MoMo would collect another.
func TestFractionalAmountsAreRefused(t *testing.T) {
	_, err := toWholeUnits(decimal.RequireFromString("1500.75"))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "whole amounts only")
}

// A trailing zero is not a fraction: 150000.00 is a whole amount that merely came in with scale.
func TestTrailingZeroesAreNotFractional(t *testing.T) {
	amount, err := toWholeUnits(decimal.RequireFromString("150000.000000"))

	require.NoError(t, err)
	assert.Equal(t, int64(150000), amount)
}

// The adapter code is what a payment-method row stores to select this implementation, so it is a
// wire identifier: changing it silently detaches every row that names it.
func TestAdapterCodeIsStable(t *testing.T) {
	assert.Equal(t, "momo", newTestAdapter().AdapterCode())
	assert.Equal(t, models.AdapterCodeMomo, newTestAdapter().AdapterCode())
}

// MoMo asks nothing of the merchant at order time — the payer identifies themselves to MoMo — so
// an order with no metadata at all is valid. The mPOS adapter is where this is not true.
func TestNoOrderTimeInputIsRequired(t *testing.T) {
	request := itGateway.OrderRequest{
		Amount:       decimal.RequireFromString("150000"),
		CurrencyCode: "VND",
	}

	metadata, err := newTestAdapter().PrepareMetadata(nil, request)
	require.NoError(t, err)
	assert.Empty(t, metadata)

	vErrs := &ft.ClientErrors{}
	require.NoError(t, newTestAdapter().ValidateOrder(nil, request, vErrs))
	assert.Zero(t, vErrs.Count())
}
