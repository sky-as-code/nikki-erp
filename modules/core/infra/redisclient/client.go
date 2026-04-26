package redisclient

import (
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	c "github.com/sky-as-code/nikki-erp/modules/core/constants"
)

func NewRedisClient(cfg config.ConfigService) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.GetStr(c.EventBusRedisHost), cfg.GetStr(c.EventBusRedisPort)),
		Password: cfg.GetStr(c.EventBusRedisPassword),
		DB:       cfg.GetInt(c.EventBusRedisDB),
	})
}
