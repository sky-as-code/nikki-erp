package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	util "github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

func main() {
	isDbMigrate := flag.Bool("migrate", false, "Whether to start in database auto migration mode")
	flag.Parse()
	logging.InitSubModule()
	util.Unused(isDbMigrate)

	app := newApplication(logging.Logger())
	app.Start()

	var server *httpserver.HttpServer
	go func() {
		err := deps.Invoke(func(s *httpserver.HttpServer) error {
			server = s
			return server.Start()
		})
		if err != nil {
			app.logger.Error("failed to start HTTP server", err)
			os.Exit(1)
		}
	}()

	<-awaitOsTerminateSignal()
	server.Shutdown()
}

func awaitOsTerminateSignal() chan os.Signal {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	return signalChan
}
