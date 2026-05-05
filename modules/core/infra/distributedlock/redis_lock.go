package distributedlock

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	c "github.com/sky-as-code/nikki-erp/modules/core/constants"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

type RedisDistributedLock struct {
	logger      logging.LoggerService
	redisClient *redis.Client
}

func NewRedisDistributedLock(logger logging.LoggerService, cfg config.ConfigService) DistributedLock {
	return &RedisDistributedLock{
		logger: logger,
		redisClient: redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%s", cfg.GetStr(c.DistributedLockRedisHost), cfg.GetStr(c.DistributedLockRedisPort)),
			Password: cfg.GetStr(c.DistributedLockRedisPassword),
			DB:       cfg.GetInt(c.DistributedLockRedisDB),
		}),
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
