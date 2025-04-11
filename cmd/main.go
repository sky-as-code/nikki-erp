package main

import (
	"flag"

	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	util "github.com/sky-as-code/nikki-erp/utility"
)

func main() {
	isDbMigrate := flag.Bool("migrate", false, "Whether to start in database auto migration mode")
	flag.Parse()
	logging.InitLogger()
	util.Unused(isDbMigrate)
	// mainHttpServer := SetupMainHttpServer()

	// startService(mainHttpServer)
}

func startService() {
	// application := pkg.NewWebApp(logging.Logger())
	// application.Start()

	// pkg.SetupMainRoutes(mainHttpServer)
	// prometheusHttpServer := pkg.SetupPrometheusHttpServer(mainHttpServer)

	// config := application.Config()
	// go mainHttpServer.Start(config.GetInt32(c.HttpPortMain))
	// go prometheusHttpServer.Start(config.GetInt32(c.HttpPortPrometheus))

	// <-pkg.AwaitOsTerminateSignal()

	// mainHttpServer.Shutdown()
	// prometheusHttpServer.Shutdown()
}
