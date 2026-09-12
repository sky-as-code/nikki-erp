package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	c "github.com/sky-as-code/nikki-erp/modules/core/constants"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	"github.com/sky-as-code/nikki-erp/modules/core/job"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

type StartableApp interface {
	Start()
	Stop(ctx context.Context)
	Logger() logging.LoggerService
	GenSql(module string, dialect string, withDeps bool) string
	GenSqlExcludes(module string) []string
}

// defaultShutdownGraceSecs matches config.default.yaml. It is used only if the configuration
// key is missing, which would otherwise panic inside GetInt.
const defaultShutdownGraceSecs = "30"

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
	withDeps := flag.Bool("withdeps", false,
		"Also emit the tables of the module's dependencies, so a cross-module foreign key has a table to reference")
	listDepTables := flag.Bool("listdeptables", false,
		"Print the dependency table names -withdeps adds, one per line, for the caller to exclude from the diff")
	job := flag.String("job", "", "Run job")
	jobArgs := flag.String("jobArgs", "", "Job Args")
	flag.Parse()

	if *listDepTables {
		runListDepTables(param.CreateAppFn, *module)
		return
	}

	if *isCreateSql {
		runCreateSql(param.CreateAppFn, *module, *dialect, *withDeps)
		return
	}

	logging.InitSubModule()

	app := param.CreateAppFn(logging.Logger())
	app.Start()

	if *job != "" {
		if jobArgs != nil && *jobArgs != "" {
			runHandleJob(*job, jobArgs)
		} else {
			runHandleJob(*job, nil)
		}
		return
	}

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

	// Modules drain first, then the HTTP server closes. The order matters: a module draining
	// background work may still need the services the server's shutdown would tear down, and
	// nothing new arrives in the meantime because the signal has already stopped the traffic
	// that would create it.
	graceSecs := config.ConfigSvcSingleton().GetInt(c.ShutdownGraceSecs, defaultShutdownGraceSecs)
	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Duration(graceSecs)*time.Second)
	defer cancel()
	app.Stop(shutdownCtx)

	// server is assigned by the goroutine above and is nil if startup failed before it got
	// that far, in which case there is nothing to shut down.
	if server != nil {
		server.Shutdown()
	}
}

func runCreateSql(createAppFn CreateAppFn, module string, dialect string, withDeps bool) {
	if module == "" || dialect == "" {
		fmt.Fprintln(os.Stderr, "error: -createsql requires both -module and -dialect to have values")
		os.Exit(1)
	}
	app := createAppFn(nil)
	sql := app.GenSql(module, dialect, withDeps)
	fmt.Print(sql)
}

func runListDepTables(createAppFn CreateAppFn, module string) {
	if module == "" {
		fmt.Fprintln(os.Stderr, "error: -listdeptables requires -module to have a value")
		os.Exit(1)
	}
	for _, table := range createAppFn(nil).GenSqlExcludes(module) {
		fmt.Println(table)
	}
}

func runHandleJob(jobName string, jobArgs *string) {
	err := deps.Invoke(func(jobManager *job.JobManager) error {
		jobManager.HandleJob(jobName, jobArgs)
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to run job %q: %v\n", jobName, err)
		os.Exit(1)
	}
}

func awaitOsTerminateSignal() chan os.Signal {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	return signalChan
}
