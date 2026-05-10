package distributedlock

import (
	"context"
	"errors"
	"time"
)

var (
	ErrAcquireLockFailed = errors.New("acquire lock failed")
)

type DistributedLock interface {
	Acquire(ctx context.Context, key string, ttl time.Duration) (bool, error)
	AcquireWithRetry(
		ctx context.Context,
		key string,
		ttl time.Duration,
		retryCount int,
		retryDelay time.Duration,
	) (bool, error)

	Release(ctx context.Context, key string) error
}

