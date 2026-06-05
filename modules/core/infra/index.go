package infra

import (
	"github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/distributedlock"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/filestorage"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/pubsub"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/redisclient"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/storage/filestorage"
)

func InitSubModule() error {
	return deps_inject.Register(
		distributedlock.NewRedisDistributedLock,
		pubsub.NewRedisPubSub,
		redisclient.NewRedisClient,
		filestorage.NewS3Adapter,
	)
}
