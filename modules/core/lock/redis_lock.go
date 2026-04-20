package lock

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
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

type RedisDistributedLock struct {
	logger      logging.LoggerService
	redisClient *redis.Client
}

func NewRedisDistributedLock(logger logging.LoggerService, redisClient *redis.Client) *RedisDistributedLock {
	return &RedisDistributedLock{
		logger:      logger,
		redisClient: redisClient,
	}
}

func (this *RedisDistributedLock) Acquire(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	return this.redisClient.SetNX(ctx, key, 1, ttl).Result()
}

func (this *RedisDistributedLock) AcquireWithRetry(
	ctx context.Context,
	key string,
	ttl time.Duration,
	retryCount int,
	retryDelay time.Duration,
) (bool, error) {
	for range retryCount {
		acquired, err := this.Acquire(ctx, key, ttl)
		if err != nil {
			return false, err
		}

		if acquired {
			return true, nil
		}

		time.Sleep(retryDelay)
	}
	return false, nil
}

func (this *RedisDistributedLock) Release(ctx context.Context, key string) error {
	return this.redisClient.Del(ctx, key).Err()
}
