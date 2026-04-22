package pubsub

import (
	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
)

func InitSubModule() error {
	return deps.Register(
		NewRedisPubSub,
		func(r *RedisPubSub) Publisher {
			return r
		},
		func(r *RedisPubSub) Subcriber {
			return r
		},
	)
}
