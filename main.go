package main

import (
	"github.com/sky-as-code/nikki-erp/cmd"
	"github.com/sky-as-code/nikki-erp/loader"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

func main() {
	cmd.Main(func(logger logging.LoggerService) cmd.StartableApp {
		return cmd.NewApplication(logger, loader.StaticModuleLoader{})
	})
}
