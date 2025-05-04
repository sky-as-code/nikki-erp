package cqrs

import (
	deps "github.com/sky-as-code/nikki-erp/common/util/deps_inject"
)

func InitSubModule() error {
	return deps.Register(NewWatermillCqrsBus)
}
