package context

import (
	"context"
	"strings"

	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

const idempotencyIdLength = 36

var (
	ErrIdempotencyIdLength  = errors.New("idempotency id: not 36 characters")
	ErrIdempotencyIdFormat  = errors.New("idempotency id: not a canonical hex UUID")
	ErrIdempotencyIdVersion = errors.New("idempotency id: not UUID version 7")
	ErrIdempotencyIdVariant = errors.New("idempotency id: not the RFC 9562 variant")
)

var ctxKeyIdempotencyId = contextKey{"idempotency_id"}

// WithIdempotencyId derives a context carrying id. Set it on the inner context before a
// RequestContext wraps it, so CloneRequestContext and every derived context carry it too.
func WithIdempotencyId(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ctxKeyIdempotencyId, id)
}

// IdempotencyIdOf never generates one: an ID minted here would differ per call and hide a lost ID.
func IdempotencyIdOf(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}
	id, ok := ctx.Value(ctxKeyIdempotencyId).(string)
	return id, ok && id != ""
}

func NewIdempotencyId() (string, error) {
	id, err := model.NewUUID()
	if err != nil {
		return "", errors.Wrap(err, "generate idempotency id")
	}
	return *id, nil
}

// ParseIdempotencyId accepts only the canonical 8-4-4-4-12 form of a UUIDv7 and returns it
// lowercased. URN, braces, a list and the nil UUID all fail one of the checks below.
func ParseIdempotencyId(raw string) (string, error) {
	if len(raw) != idempotencyIdLength {
		return "", ErrIdempotencyIdLength
	}
	lower := strings.ToLower(raw)
	for i := 0; i < len(lower); i++ {
		if !isCanonicalUuidChar(i, lower[i]) {
			return "", ErrIdempotencyIdFormat
		}
	}
	if lower[14] != '7' {
		return "", ErrIdempotencyIdVersion
	}
	if !strings.ContainsRune("89ab", rune(lower[19])) {
		return "", ErrIdempotencyIdVariant
	}
	return lower, nil
}

func isCanonicalUuidChar(pos int, char byte) bool {
	if pos == 8 || pos == 13 || pos == 18 || pos == 23 {
		return char == '-'
	}
	return (char >= '0' && char <= '9') || (char >= 'a' && char <= 'f')
}

// The logger cannot import this package (this package imports it), so the field is handed over
// through a hook. Registered at package init so every binary and test that links corectx has it.
func init() {
	logging.RegisterContextFields(func(ctx context.Context) map[string]any {
		if id, ok := IdempotencyIdOf(ctx); ok {
			return map[string]any{logging.FieldIdempotencyId: id}
		}
		return nil
	})
}
