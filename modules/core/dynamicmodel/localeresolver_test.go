package dynamicmodel

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// withResolver installs a resolver for one test and puts back whatever was there. The resolver is
// process-wide, so a test that left its own behind would change the meaning of every test after it.
func withResolver(t *testing.T, fn LocaleResolver) {
	t.Helper()
	previous := localeResolver
	SetLocaleResolver(fn)
	t.Cleanup(func() { SetLocaleResolver(previous) })
}

func newTestCtx() corectx.Context {
	return corectx.NewRequestContext(context.Background())
}

// Nothing installs a resolver in a binary that has no settings module, and a search there must
// still build its SQL.
func TestResolveLocale_NoResolverInstalledReturnsNil(t *testing.T) {
	withResolver(t, nil)

	assert.Nil(t, ResolveLocale(newTestCtx()))
}

func TestResolveLocale_ReturnsResolvedLocale(t *testing.T) {
	code := model.LanguageCode("vi-VN")
	withResolver(t, func(corectx.Context) *model.LanguageCode { return &code })

	resolved := ResolveLocale(newTestCtx())

	require.NotNil(t, resolved)
	assert.Equal(t, "vi-VN", string(*resolved))
}

// The memo is the whole reason this indirection exists: a search runs a count query and a list
// query, and the settings module has no cache of its own, so an unmemoized resolve would put two
// extra round-trips on every list request.
func TestResolveLocale_ResolvesOncePerContext(t *testing.T) {
	calls := 0
	code := model.LanguageCode("en-US")
	withResolver(t, func(corectx.Context) *model.LanguageCode {
		calls++
		return &code
	})

	ctx := newTestCtx()
	first := ResolveLocale(ctx)
	second := ResolveLocale(ctx)

	assert.Equal(t, 1, calls, "the resolver should be consulted once per request")
	assert.Equal(t, first, second)
}

// A user who has never chosen a language is the common case, so "no locale" must be memoized too --
// otherwise exactly the users who cost the most are the ones who pay on every call.
func TestResolveLocale_MemoizesTheAbsenceOfALocale(t *testing.T) {
	calls := 0
	withResolver(t, func(corectx.Context) *model.LanguageCode {
		calls++
		return nil
	})

	ctx := newTestCtx()
	assert.Nil(t, ResolveLocale(ctx))
	assert.Nil(t, ResolveLocale(ctx))

	assert.Equal(t, 1, calls, "a nil result should be cached, not re-resolved")
}

// Two requests are two resolves: the memo is per request, not process-wide, or one user's language
// would be served to the next.
func TestResolveLocale_DoesNotShareAcrossContexts(t *testing.T) {
	calls := 0
	code := model.LanguageCode("vi-VN")
	withResolver(t, func(corectx.Context) *model.LanguageCode {
		calls++
		return &code
	})

	ResolveLocale(newTestCtx())
	ResolveLocale(newTestCtx())

	assert.Equal(t, 2, calls)
}

// The resolver reaches the database. A panic there must cost the caller its localization, not its
// request.
func TestResolveLocale_RecoversFromAPanickingResolver(t *testing.T) {
	withResolver(t, func(corectx.Context) *model.LanguageCode {
		panic("settings read blew up")
	})

	assert.NotPanics(t, func() {
		assert.Nil(t, ResolveLocale(newTestCtx()))
	})
}

func TestResolveLocale_NilContextReturnsNil(t *testing.T) {
	withResolver(t, func(corectx.Context) *model.LanguageCode {
		t.Fatal("the resolver must not be called without a context")
		return nil
	})

	assert.Nil(t, ResolveLocale(nil))
}
