package vietqr

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.bryk.io/pkg/errors"
)

// The service this module supersedes logged in again before every single call. These tests are
// what hold the improvement in place.

func TestTokenIsReusedWhileItIsLive(t *testing.T) {
	clock := &fakeClock{now: time.Unix(1_700_000_000, 0)}
	cache := newTokenCache(clock.Now)
	logins := 0

	login := func() (string, time.Duration, error) {
		logins++
		return "token-1", time.Hour, nil
	}

	for range 5 {
		token, err := cache.get(login)
		require.NoError(t, err)
		assert.Equal(t, "token-1", token)
	}

	assert.Equal(t, 1, logins, "a live token must not be re-fetched")
}

func TestTokenIsReplacedOnceItExpires(t *testing.T) {
	clock := &fakeClock{now: time.Unix(1_700_000_000, 0)}
	cache := newTokenCache(clock.Now)
	logins := 0

	login := func() (string, time.Duration, error) {
		logins++
		return "token-" + string(rune('0'+logins)), time.Hour, nil
	}

	first, err := cache.get(login)
	require.NoError(t, err)

	clock.advance(2 * time.Hour)

	second, err := cache.get(login)
	require.NoError(t, err)

	assert.NotEqual(t, first, second)
	assert.Equal(t, 2, logins)
}

// A token that expires while a request is in flight fails that request, so one close to expiry is
// treated as already spent.
func TestATokenAboutToExpireIsNotHandedOut(t *testing.T) {
	clock := &fakeClock{now: time.Unix(1_700_000_000, 0)}
	cache := newTokenCache(clock.Now)
	logins := 0

	login := func() (string, time.Duration, error) {
		logins++
		return "token", time.Minute, nil
	}

	_, err := cache.get(login)
	require.NoError(t, err)

	// Inside the margin, though the token has not formally expired.
	clock.advance(time.Minute - expiryMargin/2)

	_, err = cache.get(login)
	require.NoError(t, err)
	assert.Equal(t, 2, logins, "a token inside the expiry margin must be replaced")
}

// A burst of payments arriving on an expired token must produce one login between them, not one
// each — which is the behaviour this cache exists to prevent.
func TestConcurrentCallersShareOneLogin(t *testing.T) {
	clock := &fakeClock{now: time.Unix(1_700_000_000, 0)}
	cache := newTokenCache(clock.Now)

	var mutex sync.Mutex
	logins := 0
	login := func() (string, time.Duration, error) {
		mutex.Lock()
		defer mutex.Unlock()
		logins++
		return "token", time.Hour, nil
	}

	var waitGroup sync.WaitGroup
	for range 20 {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			_, _ = cache.get(login)
		}()
	}
	waitGroup.Wait()

	assert.Equal(t, 1, logins)
}

// The gateway's idea of when a session ends is the one that counts: it may drop a token we still
// believe is live, so a rejected token has to be discardable.
func TestInvalidateForcesAFreshLogin(t *testing.T) {
	clock := &fakeClock{now: time.Unix(1_700_000_000, 0)}
	cache := newTokenCache(clock.Now)
	logins := 0

	login := func() (string, time.Duration, error) {
		logins++
		return "token", time.Hour, nil
	}

	_, err := cache.get(login)
	require.NoError(t, err)

	cache.invalidate()

	_, err = cache.get(login)
	require.NoError(t, err)
	assert.Equal(t, 2, logins)
}

// A failed login must not be cached as though it had produced a token, or every later call would
// present an empty bearer.
func TestAFailedLoginIsNotCached(t *testing.T) {
	clock := &fakeClock{now: time.Unix(1_700_000_000, 0)}
	cache := newTokenCache(clock.Now)

	_, err := cache.get(func() (string, time.Duration, error) {
		return "", 0, errors.New("gateway unreachable")
	})
	require.Error(t, err)

	token, err := cache.get(func() (string, time.Duration, error) {
		return "token", time.Hour, nil
	})
	require.NoError(t, err)
	assert.Equal(t, "token", token)
}

type fakeClock struct {
	mutex sync.Mutex
	now   time.Time
}

func (this *fakeClock) Now() time.Time {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	return this.now
}

func (this *fakeClock) advance(d time.Duration) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	this.now = this.now.Add(d)
}
