package infra

import (
	"github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/distributedlock"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/pubsub"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/redisclient"
)

func InitSubModule() error {
	return deps_inject.Register(
		distributedlock.NewRedisDistributedLock,
		pubsub.NewRedisPubSub,
		redisclient.NewRedisClient,
	)
}
