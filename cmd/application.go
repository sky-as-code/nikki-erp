package main

import (
	"fmt"
	"os"

	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/cmd/loader"
	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

func newApplication(logger logging.LoggerService) *Application {
	return &Application{
		logger: logger,
	}
}

type Application struct {
	modules []modules.NikkiModule
	config  config.ConfigService
	logger  logging.LoggerService
}

func (thisApp *Application) Config() config.ConfigService {
	return thisApp.config
}

func (thisApp *Application) Start() {
	var modules []modules.NikkiModule
	var err error

	modules, err = loader.LoadModules()
	if err != nil {
		thisApp.logger.Errorf("failed to load modules: %v", err)
	}

	thisApp.modules = append(thisApp.modules, modules...)

	err = thisApp.initModules()
	if err != nil {
		thisApp.logger.Error("failed to initialize modules", err)
		os.Exit(1)
	}
	thisApp.config = config.ConfigSvcSingleton()
}

func (thisApp *Application) initModules() error {
	moduleMap := thisApp.buildModuleMap()

	depGraph, err := thisApp.buildDependencyGraph(moduleMap)
	if err != nil {
		return err
	}

	if err := thisApp.validateDependencies(depGraph); err != nil {
		return err
	}

	return thisApp.initializeInOrder(moduleMap, depGraph)
}

func (thisApp *Application) buildModuleMap() map[string]modules.NikkiModule {
	moduleMap := make(map[string]modules.NikkiModule)
	for _, mod := range thisApp.modules {
		moduleMap[mod.Name()] = mod
	}
	return moduleMap
}

func (thisApp *Application) buildDependencyGraph(moduleMap map[string]modules.NikkiModule) (map[string][]string, error) {
	depGraph := make(map[string][]string)

	for _, mod := range thisApp.modules {
		deps := mod.Deps()
		for _, dep := range deps {
			if _, exists := moduleMap[dep]; !exists {
				return nil, errors.New(fmt.Errorf("Module '%s' requires '%s' but it's not loaded", mod.Name(), dep))
			}
		}
		depGraph[mod.Name()] = deps
	}

	return depGraph, nil
}

func (thisApp *Application) validateDependencies(depGraph map[string][]string) error {
	if hasCycle := detectCycle(depGraph); hasCycle {
		return errors.New("Modules have circular dependencies")
	}
	return nil
}

func (thisApp *Application) initializeInOrder(moduleMap map[string]modules.NikkiModule, depGraph map[string][]string) error {
	initOrder, err := topologicalSort(depGraph)
	if err != nil {
		return errors.Wrap(err, "Failed to determine module initialization order")
	}

	orderedMods := make([]modules.NikkiModule, 0)
	for _, modName := range initOrder {
		mod := moduleMap[modName]
		if err := mod.Init(); err != nil {
			return errors.Wrap(err, fmt.Sprintf("Failed to initialize module '%s'", mod.Name()))
		}
		orderedMods = append(orderedMods, mod)
		thisApp.logger.Infof("Initialized module %s done", mod.Name())
	}

	deps.Register(func() []modules.NikkiModule {
		return orderedMods
	})

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
