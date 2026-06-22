package app

import (
	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
)

func InitServices() error {
	return deps.Register(NewLanguageServiceImpl)
}
