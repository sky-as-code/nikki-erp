package context

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequestMetaTravelsWithClones(t *testing.T) {
	ctx := NewRequestContext(context.Background())
	SetRequestMeta(ctx, RequestMeta{Source: RequestSourceHttp})

	meta, ok := RequestMetaOf(CloneRequestContext(ctx))

	assert.True(t, ok)
	assert.Equal(t, RequestMeta{Source: RequestSourceHttp}, meta)
}

func TestEnsureRequestMetaStoresTheFallbackOnce(t *testing.T) {
	ctx := NewRequestContext(context.Background())

	first := EnsureRequestMeta(ctx, RequestSourceSystem)
	second := EnsureRequestMeta(ctx, RequestSourceEvent)

	assert.Equal(t, first, second)
	assert.Equal(t, RequestSourceSystem, second.Source)
}

func TestPlainContextMetaReachesARequestContext(t *testing.T) {
	inner := WithRequestMeta(context.Background(), NewRequestMeta(RequestSourceJob))

	meta, ok := RequestMetaOf(NewRequestContext(inner))

	assert.True(t, ok)
	assert.Equal(t, RequestSourceJob, meta.Source)
}
