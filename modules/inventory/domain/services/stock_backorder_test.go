package services

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

func outcome(demand, processed string) moveOutcome {
	return moveOutcome{
		MoveId:    "mv-1",
		Demand:    decimal.RequireFromString(demand),
		Processed: decimal.RequireFromString(processed),
	}
}

func TestShortfallIsWhatWasNotDelivered(t *testing.T) {
	assert.Equal(t, "30", outcome("100", "70").Shortfall().String())
	assert.True(t, outcome("100", "100").Shortfall().IsZero())
}

func TestShortfallNeverGoesNegative(t *testing.T) {
	// Processing more than demanded should not read as a negative remainder, which would look like
	// a backorder owing stock back to the warehouse.
	assert.True(t, outcome("100", "120").Shortfall().IsZero())
}

func TestDecideBackorderSkipsTheQuestionWhenEverythingShipped(t *testing.T) {
	// With no remainder there is nothing to decide, so even the `ask` policy must not demand an
	// answer — a fully-delivered transfer would otherwise fail for want of an irrelevant flag.
	decision, vErrs := DecideBackorder(models.StockBackorderPolicyAsk, []moveOutcome{outcome("100", "100")}, nil)

	assert.Equal(t, BackorderNone, decision)
	assert.Equal(t, 0, vErrs.Count())
}

func TestDecideBackorderFollowsTheAlwaysPolicy(t *testing.T) {
	decision, vErrs := DecideBackorder(models.StockBackorderPolicyAlways, []moveOutcome{outcome("100", "70")}, nil)

	assert.Equal(t, BackorderCreate, decision)
	assert.Equal(t, 0, vErrs.Count())
}

func TestDecideBackorderFollowsTheNeverPolicy(t *testing.T) {
	decision, vErrs := DecideBackorder(models.StockBackorderPolicyNever, []moveOutcome{outcome("100", "70")}, nil)

	assert.Equal(t, BackorderDrop, decision)
	assert.Equal(t, 0, vErrs.Count())
}

func TestDecideBackorderRequiresAnAnswerUnderAsk(t *testing.T) {
	// Defaulting either way would make the `ask` setting meaningless: one default silently drops a
	// commitment to the customer, the other silently creates paperwork nobody asked for.
	decision, vErrs := DecideBackorder(models.StockBackorderPolicyAsk, []moveOutcome{outcome("100", "70")}, nil)

	assert.Equal(t, BackorderNone, decision)
	require.Equal(t, 1, vErrs.Count(), "a missing decision must be a client error, not a guess")
	assert.Equal(t, "stock_transfer.backorder_decision_required", (*vErrs)[0].Key)
}

func TestDecideBackorderHonoursAnExplicitAnswerUnderAsk(t *testing.T) {
	yes, vErrs := DecideBackorder(models.StockBackorderPolicyAsk, []moveOutcome{outcome("100", "70")}, boolPtr(true))
	assert.Equal(t, BackorderCreate, yes)
	assert.Equal(t, 0, vErrs.Count())

	no, vErrs := DecideBackorder(models.StockBackorderPolicyAsk, []moveOutcome{outcome("100", "70")}, boolPtr(false))
	assert.Equal(t, BackorderDrop, no)
	assert.Equal(t, 0, vErrs.Count())
}

func TestDecideBackorderRefusesAnUnknownPolicy(t *testing.T) {
	// Guessing here could drop a remainder that should have been backordered, losing a commitment
	// with no record that it existed.
	decision, vErrs := DecideBackorder("whatever", []moveOutcome{outcome("100", "70")}, nil)

	assert.Equal(t, BackorderNone, decision)
	require.Equal(t, 1, vErrs.Count())
	assert.Equal(t, "stock_transfer.unknown_backorder_policy", (*vErrs)[0].Key)
}

func TestDecideBackorderSeesAShortfallInAnyMove(t *testing.T) {
	outcomes := []moveOutcome{outcome("100", "100"), outcome("50", "20")}

	decision, vErrs := DecideBackorder(models.StockBackorderPolicyAlways, outcomes, nil)

	assert.Equal(t, BackorderCreate, decision)
	assert.Equal(t, 0, vErrs.Count())
}

func TestDecideBackorderTreatsAWhollyUnfilledMoveAsAShortfall(t *testing.T) {
	// A move that reserved nothing processes nothing. Its whole demand is outstanding, and that is
	// the case a backorder most obviously exists for.
	decision, _ := DecideBackorder(models.StockBackorderPolicyAlways, []moveOutcome{outcome("100", "0")}, nil)

	assert.Equal(t, BackorderCreate, decision)
}
