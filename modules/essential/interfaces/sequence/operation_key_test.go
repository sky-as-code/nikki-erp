package sequence

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

func TestOperationKeyIsTheFieldAloneForASingleNumber(t *testing.T) {
	ctx := corectx.NewRequestContext(t.Context())
	assert.Equal(t, "order_number", OperationKey(ctx, "order_number"))
	assert.Equal(t, "order_number", OperationKey(ctx, "order_number", ""))
}

func TestOperationKeyIsDeterministic(t *testing.T) {
	ctx := corectx.NewRequestContext(t.Context())
	first := OperationKey(ctx, "transfer_number", "backorder_of", "01KTRANSFER0000000000000001")
	second := OperationKey(ctx, "transfer_number", "backorder_of", "01KTRANSFER0000000000000001")
	assert.Equal(t, first, second)
	assert.Equal(t, "transfer_number:backorder_of:01KTRANSFER0000000000000001", first)
}

func TestOperationKeyAddsTheBatchRow(t *testing.T) {
	ctx := corectx.NewRequestContext(t.Context())
	corectx.SetBatchRow(ctx, 3)
	assert.Equal(t, "code:row:3", OperationKey(ctx, "code"))

	corectx.SetBatchRow(ctx, 0)
	assert.Equal(t, "code", OperationKey(ctx, "code"))
}

func TestOperationKeyHashesAnOverlongTail(t *testing.T) {
	ctx := corectx.NewRequestContext(t.Context())
	long := strings.Repeat("x", 200)
	key := OperationKey(ctx, "transfer_number", "correction", long)

	assert.LessOrEqual(t, len(key), MaxOperationKeyLength)
	assert.True(t, strings.HasPrefix(key, "transfer_number:"))
	assert.Equal(t, key, OperationKey(ctx, "transfer_number", "correction", long))
	assert.NotEqual(t, key, OperationKey(ctx, "transfer_number", "correction", long+"y"))
}
