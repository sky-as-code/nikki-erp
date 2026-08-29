package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"

	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

func transferWith(status, idempotencyKey string) models.StockTransfer {
	fields := dmodel.DynamicFields{
		models.StockTransferFieldStatus: status,
	}
	if idempotencyKey != "" {
		fields[models.StockTransferFieldIdempotencyKey] = idempotencyKey
	}
	return *models.NewStockTransferFrom(fields)
}

func TestReplayedValidateReturnsThePriorResult(t *testing.T) {
	// The retry-after-timeout case: the client never saw the response, sends the same request again,
	// and must not ship the goods a second time.
	transfer := transferWith(models.StockTransferStatusDone, "key-1")

	result, err := replayedValidate(transfer, "key-1")

	require.NoError(t, err)
	require.NotNil(t, result, "a replayed validate must be recognised")
	assert.Equal(t, 1, result.Data.AffectedCount)
}

func TestReplayedValidateIgnoresATransferStillInFlight(t *testing.T) {
	// A key on a transfer that is not done means the earlier attempt did not complete. Treating it
	// as a replay would report a success that never happened and leave the stock unmoved.
	transfer := transferWith(models.StockTransferStatusReady, "key-1")

	result, err := replayedValidate(transfer, "key-1")

	require.NoError(t, err)
	assert.Nil(t, result, "an incomplete attempt must be retried, not reported as done")
}

func TestReplayedValidateIgnoresADifferentKey(t *testing.T) {
	// A different key is a different operation, even on a transfer that is already done. The
	// already-closed check refuses it afterwards; this must not short-circuit into a false success.
	transfer := transferWith(models.StockTransferStatusDone, "key-1")

	result, err := replayedValidate(transfer, "key-2")

	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestReplayedValidateIsSkippedWithoutAKey(t *testing.T) {
	// Idempotency is opt-in: a caller that sends no key gets no replay protection, and must not
	// accidentally match a transfer that happens to carry one.
	transfer := transferWith(models.StockTransferStatusDone, "key-1")

	result, err := replayedValidate(transfer, "")

	require.NoError(t, err)
	assert.Nil(t, result)
}
