package infra

import (
	"github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/distributedlock"
	mqttclient "github.com/sky-as-code/nikki-erp/modules/core/infra/mqtt"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/pubsub"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/redisclient"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/storage/filestorage"
)

func InitSubModule() error {
	return deps_inject.Register(
		mqttclient.NewVerneMq,
		filestorage.NewS3Adapter,
		distributedlock.NewRedisDistributedLock,
		// pubsub.NewRedisPubSub,
		redisclient.NewRedisClient,
		pubsub.NewVerneMq,
	)
}
