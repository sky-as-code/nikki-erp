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

type MainParam struct {
	CreateAppFn       CreateAppFn
	BeforeHttpStartFn BeforeHttpStartFn
}
type CreateAppFn func(logging.LoggerService) StartableApp
type BeforeHttpStartFn func() error

func Main(param MainParam) {
	isCreateSql := flag.Bool("createsql", false, "Generate CREATE SQL for model schemas and write to stdout")
	module := flag.String("module", "", "Module name (required when -createsql is set)")
	dialect := flag.String("dialect", "", "SQL dialect (required when -createsql is set)")
	flag.Parse()

	if *isCreateSql {
		runCreateSql(param.CreateAppFn, *module, *dialect)
		return
	}

	logging.InitSubModule()

	app := param.CreateAppFn(logging.Logger())
	app.Start()

	var server *httpserver.HttpServer
	go func() {
		var err error
		if param.BeforeHttpStartFn != nil {
			err = param.BeforeHttpStartFn()
		}

		if err == nil {
			err = deps.Invoke(func(s *httpserver.HttpServer) error {
				server = s
				return server.Start()
			})
		}
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
