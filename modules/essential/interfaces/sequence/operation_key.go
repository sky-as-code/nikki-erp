package sequence

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

const (
	operationKeySeparator = ":"
	operationKeyBatchRow  = "row"
	// MaxOperationKeyLength is the longest operation key NextValue accepts.
	MaxOperationKeyLength = 128
)

// OperationKey is the one way a consumer names a NextValue call. Together with the request's
// idempotency id (which the port takes from ctx when the command leaves it nil) it is the issuance's
// identity, so it must be the same on every attempt of the same request: a retry then replays the
// number it was issued instead of drawing a second one.
//
// The key is the numbered field, followed by the discriminators that tell apart several numbers
// one request draws from the same definition (the source document of a backorder, a return or a
// correction), and by the batch row when ctx is writing one row of a bulk create or an import.
// Never pass a freshly minted id, including the id of the record being created: it differs per
// attempt. A key too long for NextValue keeps the field and hashes the rest.
func OperationKey(ctx context.Context, field string, discriminators ...string) string {
	parts := make([]string, 0, len(discriminators)+2)
	for _, part := range discriminators {
		if part != "" {
			parts = append(parts, part)
		}
	}
	if row, ok := corectx.BatchRowOf(ctx); ok {
		parts = append(parts, operationKeyBatchRow, strconv.Itoa(row))
	}
	if len(parts) == 0 {
		return field
	}
	tail := strings.Join(parts, operationKeySeparator)
	key := field + operationKeySeparator + tail
	if len(key) <= MaxOperationKeyLength {
		return key
	}
	digest := sha256.Sum256([]byte(tail))
	return field + operationKeySeparator + hex.EncodeToString(digest[:])
}
