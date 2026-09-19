package services

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.bryk.io/pkg/errors"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
)

func TestCapToWarehouseAvailabilityNeverClaimsWhatIsCommitted(t *testing.T) {
	// H=10, R=6 leaves A=4: a demand for 5 is capped to 4 (AC-02), one for 3 is untouched.
	assert.True(t, decimal.NewFromInt(4).Equal(capToWarehouseAvailability(decimal.NewFromInt(5), decimal.NewFromInt(4))))
	assert.True(t, decimal.NewFromInt(3).Equal(capToWarehouseAvailability(decimal.NewFromInt(3), decimal.NewFromInt(4))))
	assert.True(t, capToWarehouseAvailability(decimal.NewFromInt(3), decimal.NewFromInt(-2)).IsZero(),
		"a scope already short of its commitments gives nothing")
}

func TestProtectionViolationSurvivesWrappingAndBecomesAClientError(t *testing.T) {
	vErrs := ft.NewClientErrors()
	vErrs.Append(*ft.NewBusinessViolation("inventory_stock_reservation", ReasonProtectionViolation, "short"))
	wrapped := errors.Wrap(&ProtectionViolation{Errors: vErrs}, "withQuantTransaction")

	unwrapped, refused := ProtectionViolationOf(wrapped)
	require.True(t, refused, "the refusal must be recognised through the transaction wrapper")
	assert.Equal(t, 1, unwrapped.Count())

	result := protectionResultOf(wrapped)
	require.NotNil(t, result)
	assert.Equal(t, 1, result.ClientErrors.Count(), "answered as a 400, not a 500")

	assert.Nil(t, protectionResultOf(errors.New("database down")), "any other error stays an error")
	_, refused = ProtectionViolationOf(nil)
	assert.False(t, refused)
}
