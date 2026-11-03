package redisclient

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"

	"github.com/sky-as-code/nikki-erp/modules/core/config"
	c "github.com/sky-as-code/nikki-erp/modules/core/constants"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

func NewRedisClient(cfg config.ConfigService, logger logging.LoggerService) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.GetStr(c.EventBusRedisHost), cfg.GetStr(c.EventBusRedisPort)),
		Password: cfg.GetStr(c.EventBusRedisPassword),
		DB:       cfg.GetInt(c.EventBusRedisDB),
	})
	InstrumentTracing(client, logger)
	return client
}

// InstrumentTracing adds a CLIENT span per command. The span names the command only: the
// statement (keys and values, which may hold tokens or personal data) and the calling code
// location are not recorded (BE-04). A failure leaves the client working, untraced.
func InstrumentTracing(client redis.UniversalClient, logger logging.LoggerService) {
	err := redisotel.InstrumentTracing(client,
		redisotel.WithDBStatement(false),
		redisotel.WithCallerEnabled(false),
	)
	if err != nil {
		logger.Warn(context.Background(), "redis client is not traced", logging.Attr{"error": err.Error()})
	}
}
