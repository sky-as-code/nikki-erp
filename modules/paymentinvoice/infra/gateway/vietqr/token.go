package vietqr

import (
	"sync"
	"time"
)

// tokenCache holds the bearer token this deployment presents to VietQR, and hands out a fresh one
// when it is close to expiring.
//
// The service this module supersedes logged in again before every single call — a QR generation
// was two round trips, a refund two, a status check two. Caching halves that and removes an
// authentication attempt from the critical path of taking a payment.
//
// The clock is injected rather than read directly so the expiry logic can be tested without
// sleeping.
type tokenCache struct {
	mutex     sync.Mutex
	token     string
	expiresAt time.Time
	now       func() time.Time
}

// expiryMargin is how long before its stated expiry a token is treated as spent.
//
// A token that expires while a request is in flight fails that request, so the margin has to
// cover the round trip plus any clock skew between us and the gateway. Thirty seconds is
// generous for the former and forgiving of the latter.
const expiryMargin = 30 * time.Second

func newTokenCache(now func() time.Time) *tokenCache {
	if now == nil {
		now = time.Now
	}
	return &tokenCache{now: now}
}

// get returns a usable token, calling login only when there is no live one.
//
// The lock is held across login, so concurrent callers arriving at an expired token produce one
// login between them rather than one each. That matters: a burst of payments would otherwise
// open a burst of authentications, which is exactly the behaviour being fixed here.
func (this *tokenCache) get(login func() (string, time.Duration, error)) (string, error) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	if this.token != "" && this.now().Before(this.expiresAt.Add(-expiryMargin)) {
		return this.token, nil
	}

	token, expiresIn, err := login()
	if err != nil {
		return "", err
	}

	this.token = token
	this.expiresAt = this.now().Add(expiresIn)
	return token, nil
}

// invalidate drops the cached token, so the next call authenticates again.
//
// Used when the gateway rejects a token we believed was live — its idea of the expiry is the one
// that counts, and it may end a session early.
func (this *tokenCache) invalidate() {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	this.token = ""
	this.expiresAt = time.Time{}
}
