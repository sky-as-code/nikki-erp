package dynamicmodel

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// LocaleResolver reports the effective display locale for the acting user, or nil when none can be
// determined.
//
// It is a function type rather than an interface because the settings module -- the only thing that
// can answer this -- imports THIS package. An interface would have to name the settings types to be
// useful, which would close the loop into an import cycle and abort the build. A func value carries
// no package dependency on its provider, so the edge stays one-directional and the module that owns
// both ends pushes the implementation in at boot.
type LocaleResolver func(ctx corectx.Context) *model.LanguageCode

// localeResolver is written once during module Init and only read afterwards. It is deliberately
// not guarded: installing it mid-flight would be the bug, not the missing lock.
var localeResolver LocaleResolver

// SetLocaleResolver installs the process-wide resolver.
//
// Call it once, during module Init, before any request is served.
func SetLocaleResolver(fn LocaleResolver) {
	localeResolver = fn
}

type localeCtxKeyType struct{}

// localeCtxKey is an unexported zero-size type so no other package can collide with it.
var localeCtxKey = localeCtxKeyType{}

// cachedLocale wraps the result so that a memoized "this user has no locale" (a nil pointer) is
// distinguishable from "not resolved yet" (the key absent). Without the wrapper, every user without
// a stored locale would re-read the settings tables on each call -- which is precisely the cost the
// memo exists to avoid.
type cachedLocale struct {
	value *model.LanguageCode
}

// ResolveLocale returns the acting user's effective locale, memoized for the request.
//
// Returns nil whenever the locale cannot be established -- no resolver installed, no acting user, a
// failed or panicking read. Every caller must treat nil as "do not localize" rather than as an
// error: a list that sorts by the raw column is worse than one sorted by the reader's language, but
// it is a great deal better than a 500.
//
// The memo assumes a request context is used by one goroutine at a time, which is how corectx is
// used throughout and what every other WithValue caller already assumes.
func ResolveLocale(ctx corectx.Context) *model.LanguageCode {
	if ctx == nil {
		return nil
	}
	if cached, ok := ctx.Value(localeCtxKey).(cachedLocale); ok {
		return cached.value
	}

	resolved := callLocaleResolver(ctx)
	ctx.WithValue(localeCtxKey, cachedLocale{value: resolved})
	return resolved
}

// callLocaleResolver isolates the installed resolver's failure modes.
//
// A settings read reaches the database, and a panic in it would otherwise travel up through the
// query builder and take down a request that only wanted a sorted list.
func callLocaleResolver(ctx corectx.Context) (resolved *model.LanguageCode) {
	if localeResolver == nil {
		return nil
	}
	defer func() {
		if recovered := recover(); recovered != nil {
			resolved = nil
		}
	}()
	return localeResolver(ctx)
}
