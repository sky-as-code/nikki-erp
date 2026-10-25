package dynamicmodel

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

func withEnabledLanguagesResolver(t *testing.T, fn EnabledLanguagesResolver) {
	t.Helper()
	previous := enabledLanguagesResolver
	SetEnabledLanguagesResolver(fn)
	t.Cleanup(func() { SetEnabledLanguagesResolver(previous) })
}

func TestResolveEnabledLanguages_NoResolverMeansUnknown(t *testing.T) {
	withEnabledLanguagesResolver(t, nil)

	codes, err := ResolveEnabledLanguages(newTestCtx())

	require.NoError(t, err)
	assert.Nil(t, codes)
}

func TestResolveEnabledLanguages_EmptyListMeansUnknown(t *testing.T) {
	withEnabledLanguagesResolver(t, func(corectx.Context) ([]model.LanguageCode, error) {
		return []model.LanguageCode{}, nil
	})

	codes, err := ResolveEnabledLanguages(newTestCtx())

	require.NoError(t, err)
	assert.Nil(t, codes)
}

func TestResolveEnabledLanguages_ResolvesOncePerContext(t *testing.T) {
	calls := 0
	withEnabledLanguagesResolver(t, func(corectx.Context) ([]model.LanguageCode, error) {
		calls++
		return []model.LanguageCode{"en-US", "vi-VN"}, nil
	})

	ctx := newTestCtx()
	first, _ := ResolveEnabledLanguages(ctx)
	second, _ := ResolveEnabledLanguages(ctx)

	assert.Equal(t, 1, calls)
	assert.Equal(t, []model.LanguageCode{"en-US", "vi-VN"}, first)
	assert.Equal(t, first, second)
}

func TestResolveEnabledLanguages_ErrorIsReturnedAndMemoized(t *testing.T) {
	calls := 0
	withEnabledLanguagesResolver(t, func(corectx.Context) ([]model.LanguageCode, error) {
		calls++
		return nil, errors.New("db down")
	})

	ctx := newTestCtx()
	_, firstErr := ResolveEnabledLanguages(ctx)
	_, secondErr := ResolveEnabledLanguages(ctx)

	assert.Error(t, firstErr)
	assert.Error(t, secondErr)
	assert.Equal(t, 1, calls)
}

func TestResolveEnabledLanguages_PanicBecomesError(t *testing.T) {
	withEnabledLanguagesResolver(t, func(corectx.Context) ([]model.LanguageCode, error) {
		panic("boom")
	})

	codes, err := ResolveEnabledLanguages(newTestCtx())

	assert.Error(t, err)
	assert.Nil(t, codes)
}
