package context

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const validIdempotencyId = "019a46c2-2a40-7b35-9e0f-8e2fd22b9b6c"

func TestParseIdempotencyId_AcceptsCanonicalV7(t *testing.T) {
	id, err := ParseIdempotencyId(validIdempotencyId)
	require.NoError(t, err)
	assert.Equal(t, validIdempotencyId, id)
}

func TestParseIdempotencyId_LowercasesUppercase(t *testing.T) {
	id, err := ParseIdempotencyId("019A46C2-2A40-7B35-9E0F-8E2FD22B9B6C")
	require.NoError(t, err)
	assert.Equal(t, validIdempotencyId, id)
}

func TestParseIdempotencyId_Rejects(t *testing.T) {
	cases := []struct {
		name    string
		raw     string
		wantErr error
	}{
		{"empty", "", ErrIdempotencyIdLength},
		{"nil uuid", "00000000-0000-0000-0000-000000000000", ErrIdempotencyIdVersion},
		{"max uuid", "ffffffff-ffff-ffff-ffff-ffffffffffff", ErrIdempotencyIdVersion},
		{"uuid v4", "9b2f1e64-3c1d-4b7a-9f0e-8e2fd22b9b6c", ErrIdempotencyIdVersion},
		{"wrong variant", "019a46c2-2a40-7b35-1e0f-8e2fd22b9b6c", ErrIdempotencyIdVariant},
		{"urn", "urn:uuid:" + validIdempotencyId, ErrIdempotencyIdLength},
		{"braces", "{" + validIdempotencyId + "}", ErrIdempotencyIdLength},
		{"list", validIdempotencyId + "," + validIdempotencyId, ErrIdempotencyIdLength},
		{"no hyphens", "019a46c22a407b359e0f8e2fd22b9b6c", ErrIdempotencyIdLength},
		{"misplaced hyphen", "019a46c-22a40-7b35-9e0f-8e2fd22b9b6c", ErrIdempotencyIdFormat},
		{"non hex", "019a46c2-2a40-7b35-9e0f-8e2fd22b9b6z", ErrIdempotencyIdFormat},
		{"padded", " 019a46c2-2a40-7b35-9e0f-8e2fd22b9b6", ErrIdempotencyIdFormat},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			id, err := ParseIdempotencyId(tc.raw)
			assert.ErrorIs(t, err, tc.wantErr)
			assert.Empty(t, id)
		})
	}
}

func TestNewIdempotencyId_IsValidAndDistinct(t *testing.T) {
	first, err := NewIdempotencyId()
	require.NoError(t, err)
	second, err := NewIdempotencyId()
	require.NoError(t, err)

	parsed, err := ParseIdempotencyId(first)
	require.NoError(t, err)
	assert.Equal(t, first, parsed)
	assert.NotEqual(t, first, second)
}

func TestIdempotencyIdOf_NeverGenerates(t *testing.T) {
	id, ok := IdempotencyIdOf(context.Background())
	assert.False(t, ok)
	assert.Empty(t, id)

	_, ok = IdempotencyIdOf(nil) //nolint:staticcheck
	assert.False(t, ok)
}

func TestIdempotencyId_SurvivesRequestContextAndClone(t *testing.T) {
	inner := WithIdempotencyId(context.Background(), validIdempotencyId)
	reqCtx := NewRequestContext(inner)
	reqCtx.WithValue("other", "value")

	for _, ctx := range []context.Context{reqCtx, CloneRequestContext(reqCtx)} {
		id, ok := IdempotencyIdOf(ctx)
		assert.True(t, ok)
		assert.Equal(t, validIdempotencyId, id)
	}
}
