package event

import (
	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
)

func InitSubModule() error {
	return deps.Register(NewRedisEventBus)
}
