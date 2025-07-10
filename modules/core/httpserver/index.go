package httpserver

import (
	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
)

func InitSubModule() error {
	err := deps.Register(func(params httpServerParams) httpServerResult {
		return initHttpServer(params)
	})
	return err
}
