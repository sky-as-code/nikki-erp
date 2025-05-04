package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/sky-as-code/nikki-erp/common/httpserver"
	"github.com/sky-as-code/nikki-erp/common/logging"
	util "github.com/sky-as-code/nikki-erp/common/util"
	deps "github.com/sky-as-code/nikki-erp/common/util/deps_inject"
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
