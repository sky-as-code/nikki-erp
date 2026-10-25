package dynamicmodel

import (
	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// EnabledLanguagesResolver reports the language codes the acting tenant has enabled.
//
// It is a function type for the same reason as LocaleResolver: the essential module, which owns
// the languages, imports THIS package, so an interface naming its types would close an import cycle.
type EnabledLanguagesResolver func(ctx corectx.Context) ([]model.LanguageCode, error)

// enabledLanguagesResolver is written once during module Init and only read afterwards.
var enabledLanguagesResolver EnabledLanguagesResolver

// SetEnabledLanguagesResolver installs the process-wide resolver.
//
// Call it once, during module Init, before any request is served.
func SetEnabledLanguagesResolver(fn EnabledLanguagesResolver) {
	enabledLanguagesResolver = fn
}

type enabledLanguagesCtxKeyType struct{}

var enabledLanguagesCtxKey = enabledLanguagesCtxKeyType{}

// cachedEnabledLanguages memoizes the error too: a request that writes several records must not
// re-read the languages table once per record only to fail the same way.
type cachedEnabledLanguages struct {
	value []model.LanguageCode
	err   error
}

// ResolveEnabledLanguages returns the acting tenant's enabled language codes, memoized for the request.
//
// A nil result with a nil error means the set is unknown -- no resolver installed (unit tests, a
// binary without essential) -- and callers must then skip anything that depends on it. Unlike
// ResolveLocale, a failed read is an error: a writer that drops languages cannot silently guess.
func ResolveEnabledLanguages(ctx corectx.Context) ([]model.LanguageCode, error) {
	if ctx == nil {
		return nil, nil
	}
	if cached, ok := ctx.Value(enabledLanguagesCtxKey).(cachedEnabledLanguages); ok {
		return cached.value, cached.err
	}

	resolved, err := callEnabledLanguagesResolver(ctx)
	ctx.WithValue(enabledLanguagesCtxKey, cachedEnabledLanguages{value: resolved, err: err})
	return resolved, err
}

func callEnabledLanguagesResolver(ctx corectx.Context) (resolved []model.LanguageCode, err error) {
	if enabledLanguagesResolver == nil {
		return nil, nil
	}
	defer func() {
		if recovered := recover(); recovered != nil {
			resolved = nil
			err = errors.Errorf("enabled languages resolver panicked: %v", recovered)
		}
	}()
	resolved, err = enabledLanguagesResolver(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "ResolveEnabledLanguages")
	}
	// A tenant with no enabled language is misconfigured, not asking for every translation to be
	// dropped, so it reads as "unknown".
	if len(resolved) == 0 {
		return nil, nil
	}
	return resolved, nil
}
