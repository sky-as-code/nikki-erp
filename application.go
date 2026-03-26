package main

import (
	"fmt"
	"os"
	"strings"

	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/array"
	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/orm"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/loader"
	"github.com/sky-as-code/nikki-erp/modules"
	"github.com/sky-as-code/nikki-erp/modules/core"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

func newApplication(logger logging.LoggerService) *Application {
	return &Application{
		logger: logger,
	}
}

type Application struct {
	modules []modules.InCodeModule
	config  config.ConfigService
	logger  logging.LoggerService
}

func (this *Application) Config() config.ConfigService {
	return this.config
}

func (this *Application) Start() {
	modules := []modules.InCodeModule{}
	var err error

	modules, err = loader.LoadModules()
	if err != nil {
		this.logger.Errorf("failed to load modules: %v", err)
	}

	this.modules = append(this.modules, modules...)

	err = this.initModules()
	if err != nil {
		this.logger.Error("failed to initialize modules", err)
		os.Exit(1)
	}
	this.config = config.ConfigSvcSingleton()
}

func (this *Application) GenSql(moduleName string, dialect string) string {
	module, err := loader.LoadModule(moduleName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load modules: %v\n", err)
	}

	dynamicModule, ok := module.(modules.DynamicModule)
	if !ok {
		fmt.Fprintf(os.Stderr, "module '%s' is not a dynamic module\n", moduleName)
		os.Exit(1)
	}

	err = dynamicModule.RegisterModels()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to register entities: %v\n", err)
		os.Exit(1)
	}

	registry := dmodel.GetSchemaRegistry()
	err = orm.ValidateRelations(registry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to validate relations: %v\n", err)
		os.Exit(1)
	}

	queries, err := orm.GenCreateSql(registry, dialect)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to generate create SQL: %v\n", err)
		os.Exit(1)
	}
	return strings.Join(queries, ";\n")
}

func (this *Application) initModules() error {
	moduleMap := this.buildModuleMap()

	depGraph, err := this.buildDependencyGraph(moduleMap)
	if err != nil {
		return err
	}

	if err := this.validateDependencies(depGraph); err != nil {
		return err
	}

	if err := this.registerModelInOrder(moduleMap, depGraph); err != nil {
		return err
	}

	return this.initializeInOrder(moduleMap, depGraph)
}

func (this *Application) registerEntities() error {
	moduleMap := this.buildModuleMap()

	depGraph, err := this.buildDependencyGraph(moduleMap)
	if err != nil {
		return err
	}

	if err := this.validateDependencies(depGraph); err != nil {
		return err
	}

	return this.registerModelInOrder(moduleMap, depGraph)
}

func (this *Application) buildModuleMap() map[string]modules.InCodeModule {
	moduleMap := make(map[string]modules.InCodeModule)
	moduleMap["core"] = core.ModuleSingleton
	for _, mod := range this.modules {
		moduleMap[mod.Name()] = mod
	}
	return moduleMap
}

func (this *Application) buildDependencyGraph(moduleMap map[string]modules.InCodeModule) (map[string][]string, error) {
	depGraph := make(map[string][]string)

	for _, mod := range this.modules {
		deps := mod.Deps()
		for _, dep := range deps {
			if _, exists := moduleMap[dep]; !exists {
				return nil, errors.New(fmt.Errorf("module '%s' requires '%s' but it's not loaded", mod.Name(), dep))
			}
		}
		depGraph[mod.Name()] = deps
	}

	return depGraph, nil
}

func (this *Application) validateDependencies(depGraph map[string][]string) error {
	if hasCycle := detectCycle(depGraph); hasCycle {
		return errors.New("modules have circular dependencies")
	}
	return nil
}

func (this *Application) initializeInOrder(moduleMap map[string]modules.InCodeModule, depGraph map[string][]string) error {
	this.logger.Info("Start initializing modules", nil)

	initOrder, err := topologicalSort(depGraph)
	if err != nil {
		return errors.Wrap(err, "failed to determine module initialization order")
	}

	initOrder = array.Prepend(initOrder, "core")
	orderedMods := make([]modules.InCodeModule, 0)
	for _, modName := range initOrder {
		mod := moduleMap[modName]
		if err := this.initModule(mod); err != nil {
			return err
		}
		orderedMods = append(orderedMods, mod)
		this.logger.Infof("Initialized module %s", mod.Name())
	}

	deps.Register(func() []modules.InCodeModule {
		return orderedMods
	})

	for _, mod := range orderedMods {
		modWithAppStarted, ok := mod.(modules.InCodeModuleAppStarted)
		if ok {
			if err := modWithAppStarted.OnAppStarted(); err != nil {
				return err
			}
		}
		this.logger.Debugf("Invoked OnAppStarted() on module %s", mod.Name())
	}

	return nil
}

func (this *Application) registerModelInOrder(moduleMap map[string]modules.InCodeModule, depGraph map[string][]string) error {
	this.logger.Info("Start registering models for modules", nil)

	initOrder, err := topologicalSort(depGraph)
	if err != nil {
		return errors.Wrap(err, "failed to determine model registering order")
	}

	initOrder = array.Prepend(initOrder, "core")
	for _, modName := range initOrder {
		mod := moduleMap[modName]
		modWithDynamic, ok := mod.(modules.DynamicModule)
		if !ok {
			continue
		}
		if err := modWithDynamic.RegisterModels(); err != nil {
			return err
		}
		this.logger.Infof("Registered models for module %s", mod.Name())
	}

	return nil
}

func (this *Application) initModule(mod modules.InCodeModule) (err error) {
	defer func() {
		if e := ft.RecoverPanicf(recover(), "failed to initialize module '%s'", mod.Name()); e != nil {
			err = e
		}
	}()
	if err := mod.Init(); err != nil {
		panic(err)
	}
	return nil
}

func detectCycle(graph map[string][]string) bool {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var visit func(string) bool
	visit = func(node string) bool {
		if !visited[node] {
			visited[node] = true
			recStack[node] = true

			for _, dep := range graph[node] {
				if !visited[dep] && visit(dep) {
					return true
				} else if recStack[dep] {
					return true
				}
			}
		}
		recStack[node] = false
		return false
	}

	for node := range graph {
		if !visited[node] && visit(node) {
			return true
		}
	}
	return false
}

func topologicalSort(graph map[string][]string) ([]string, error) {
	visited := make(map[string]bool)
	temp := make(map[string]bool)
	order := make([]string, 0)

	var visit func(string) error
	visit = func(node string) error {
		if temp[node] {
			return fmt.Errorf("cycle detected at module '%s'", node)
		}
		if !visited[node] {
			temp[node] = true
			for _, dep := range graph[node] {
				if err := visit(dep); err != nil {
					return err
				}
			}
			visited[node] = true
			temp[node] = false
			// Changed: append to end instead of prepending
			order = append(order, node)
		}
		return nil
	}

	for node := range graph {
		if !visited[node] {
			if err := visit(node); err != nil {
				return nil, err
			}
		}
	}

	return order, nil
}
