package cmd

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

type StartableApp interface {
	Start()
	Logger() logging.LoggerService
	GenSql(module string, dialect string) string
}

type CreateAppFn func(logging.LoggerService) StartableApp

func Main(createAppFn CreateAppFn) {
	isCreateSql := flag.Bool("createsql", false, "Generate CREATE SQL for model schemas and write to stdout")
	module := flag.String("module", "", "Module name (required when -createsql is set)")
	dialect := flag.String("dialect", "", "SQL dialect (required when -createsql is set)")
	flag.Parse()

	if *isCreateSql {
		runCreateSql(createAppFn, *module, *dialect)
		return
	}

	logging.InitSubModule()

	app := createAppFn(logging.Logger())
	app.Start()

	var server *httpserver.HttpServer
	go func() {
		err := deps.Invoke(func(s *httpserver.HttpServer) error {
			server = s
			return server.Start()
		})
		if err != nil {
			app.Logger().Error("failed to start HTTP server", err)
			os.Exit(1)
		}
	}()

	<-awaitOsTerminateSignal()
	server.Shutdown()
}

func runCreateSql(createAppFn CreateAppFn, module string, dialect string) {
	if module == "" || dialect == "" {
		fmt.Fprintln(os.Stderr, "error: -createsql requires both -module and -dialect to have values")
		os.Exit(1)
	}
	app := createAppFn(nil)
	sql := app.GenSql(module, dialect)
	fmt.Print(sql)
}

func awaitOsTerminateSignal() chan os.Signal {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	return signalChan
}
