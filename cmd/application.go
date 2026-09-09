package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"go.bryk.io/pkg/errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/orm"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules"
	apptraitconstants "github.com/sky-as-code/nikki-erp/modules/apptrait/constants"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	coreconstants "github.com/sky-as-code/nikki-erp/modules/core/constants"
	"github.com/sky-as-code/nikki-erp/modules/core/job"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

type ModuleLoader interface {
	LoadModules() ([]modules.InCodeModule, error)
	LoadModule(name string) (modules.InCodeModule, error)
}

func NewApplication(logger logging.LoggerService, moduleLoader ModuleLoader) *Application {
	return &Application{
		logger:       logger,
		moduleLoader: moduleLoader,
	}
}

type Application struct {
	modules      []modules.InCodeModule
	orderedMods  []modules.InCodeModule
	config       config.ConfigService
	logger       logging.LoggerService
	moduleLoader ModuleLoader
}

func (this *Application) Config() config.ConfigService {
	return this.config
}

func (this *Application) Logger() logging.LoggerService {
	return this.logger
}

func (this *Application) Start() {
	modules, err := this.moduleLoader.LoadModules()
	if err != nil {
		this.logger.Errorf("failed to load modules: %v", err)
	}

	this.modules = modules

	err = this.initModules()
	if err != nil {
		this.logger.Error("failed to initialize modules", err)
		os.Exit(1)
	}
	this.config = config.ConfigSvcSingleton()
}

// Stop drains modules in reverse init order under a shared ctx deadline; a failing or
// overrunning module is logged and skipped rather than aborting the rest.
func (this *Application) Stop(ctx context.Context) {
	if len(this.orderedMods) == 0 {
		return
	}
	this.logger.Info("Start stopping modules", nil)

	// Scheduler stops first so no sweep fires against a half-torn-down module; blocks until
	// in-flight jobs return.
	if err := job.GetCronjob().Stop(); err != nil {
		this.logger.Error("the cron scheduler did not stop cleanly", err)
	}

	for i := len(this.orderedMods) - 1; i >= 0; i-- {
		mod := this.orderedMods[i]
		modWithStopping, ok := mod.(modules.InCodeModuleAppStopping)
		if !ok {
			continue
		}

		if err := ctx.Err(); err != nil {
			this.logger.Errorf("shutdown budget spent; skipping OnAppStopping() for module %s", mod.Name())
			continue
		}

		if err := modWithStopping.OnAppStopping(ctx); err != nil {
			this.logger.Error(fmt.Sprintf("module %s failed to stop cleanly", mod.Name()), err)
			continue
		}
		this.logger.Debugf("Invoked OnAppStopping() on module %s", mod.Name())
	}
}

func (this *Application) GenSql(moduleName string, dialect string) string {
	registeredCount, err := this.registerModelsUpTo(moduleName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to register models: %v\n", err)
		os.Exit(1)
	}

	registry := dmodel.GetSchemaRegistry()
	if err := registry.FinalizeRelations(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to finalize schema relations: %v\n", err)
		os.Exit(1)
	}

	if err := orm.ValidateRelations(registry); err != nil {
		fmt.Fprintf(os.Stderr, "failed to validate relations: %v\n", err)
		os.Exit(1)
	}

	prefix := this.buildModuleMap()[moduleName].(modules.DynamicModule).ModelPrefix() + "_"
	queries, err := orm.GenCreateSql(registry, dialect, prefix)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to generate create SQL: %v\n", err)
		os.Exit(1)
	}

	// Schemas registered but none matched the prefix: ModelPrefix() is wrong for this module.
	if len(queries) == 0 && registeredCount > 0 {
		fmt.Fprintf(os.Stderr,
			"module '%s' registered %d schema(s) but none match '%s'; "+
				"fix ModelPrefix() in the module's index.go\n",
			moduleName, registeredCount, prefix)
		os.Exit(1)
	}
	return strings.Join(queries, ";\n")
}

// registerModelsUpTo registers models for moduleName and its transitive dependencies only
// (not registerModelInOrder's full set — GenSql runs with a nil logger, and base schemas
// like "core.basemodel.base_model" must exist before a JSON model can extend them).
// Registers the dependency closure, not a slice of the full topological order, since that
// order is non-deterministic across runs (topologicalSort seeds from a map range).
// Returns how many schemas moduleName itself added, so callers can distinguish "no tables"
// from "prefix filter matched nothing".
func (this *Application) registerModelsUpTo(moduleName string) (int, error) {
	mods, err := this.moduleLoader.LoadModules()
	if err != nil {
		return 0, errors.Wrap(err, "failed to load modules")
	}
	this.modules = mods

	moduleMap := this.buildModuleMap()
	if _, exists := moduleMap[moduleName]; !exists {
		return 0, errors.Errorf("module '%s' not found", moduleName)
	}
	if _, ok := moduleMap[moduleName].(modules.DynamicModule); !ok {
		return 0, errors.Errorf("module '%s' is not a dynamic module", moduleName)
	}

	depGraph, err := this.buildDependencyGraph(moduleMap)
	if err != nil {
		return 0, err
	}

	registerOrder, err := dependencyClosure(depGraph, moduleName)
	if err != nil {
		return 0, errors.Wrap(err, "failed to determine model registering order")
	}

	registry := dmodel.GetSchemaRegistry()
	targetSchemaCount := 0
	for _, modName := range registerOrder {
		dynamicMod, ok := moduleMap[modName].(modules.DynamicModule)
		if !ok {
			continue
		}
		before := countSchemas(registry)
		if err := dynamicMod.RegisterModels(); err != nil {
			return 0, errors.Wrapf(err, "module '%s'", modName)
		}
		if modName == moduleName {
			targetSchemaCount = countSchemas(registry) - before
		}
	}

	return targetSchemaCount, nil
}

func countSchemas(registry *dmodel.SchemaRegistry) int {
	count := 0
	_ = registry.ForEach(func(_ string, _ *dmodel.ModelSchema) error {
		count++
		return nil
	})
	return count
}

// dependencyClosure returns root and its transitive deps only, dependencies-first (so base
// schemas register before modules extending them), each module exactly once (repeat
// registration is rejected), deterministically (walk starts at root, not a map key).
func dependencyClosure(graph map[string][]string, root string) ([]string, error) {
	visited := make(map[string]bool)
	inProgress := make(map[string]bool)
	order := make([]string, 0)

	var visit func(string) error
	visit = func(node string) error {
		if inProgress[node] {
			return errors.Errorf("cycle detected at module '%s'", node)
		}
		if visited[node] {
			return nil
		}
		inProgress[node] = true
		for _, dep := range graph[node] {
			if err := visit(dep); err != nil {
				return err
			}
		}
		inProgress[node] = false
		visited[node] = true
		order = append(order, node)
		return nil
	}

	if err := visit(root); err != nil {
		return nil, err
	}
	return order, nil
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

func (this *Application) buildModuleMap() map[string]modules.InCodeModule {
	moduleMap := make(map[string]modules.InCodeModule)
	for _, mod := range this.modules {
		moduleMap[mod.Name()] = mod
	}
	return moduleMap
}

func (this *Application) buildDependencyGraph(moduleMap map[string]modules.InCodeModule) (map[string][]string, error) {
	depGraph := make(map[string][]string)

	for _, mod := range this.modules {
		modName := mod.Name()
		deps := mod.Deps()
		if modName != apptraitconstants.AppTraitModuleName &&
			modName != coreconstants.CoreModuleName {
			deps = append(deps, coreconstants.CoreModuleName)
		}
		for _, dep := range deps {
			if _, exists := moduleMap[dep]; !exists {
				return nil, errors.New(fmt.Errorf("module '%s' requires '%s' but it's not loaded", mod.Name(), dep))
			}
		}
		depGraph[modName] = deps
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

	orderedMods := make([]modules.InCodeModule, 0)
	for _, modName := range initOrder {
		mod := moduleMap[modName]
		if err := this.initModule(mod); err != nil {
			return err
		}
		orderedMods = append(orderedMods, mod)
		this.logger.Infof("Initialized module %s", mod.Name())
	}

	this.orderedMods = orderedMods
	deps.Register(func() []modules.InCodeModule {
		return orderedMods
	})

	for _, mod := range orderedMods {
		modWithAppStarted, ok := mod.(modules.InCodeModuleAppStarted)
		if ok {
			if err := modWithAppStarted.OnAppStarted(); err != nil {
				return err
			}
			this.logger.Debugf("Invoked OnAppStarted() on module %s", mod.Name())
		}
	}

	// Starts last, after all modules register their cron jobs in OnAppStarted — gocron fires
	// on schedule immediately, so starting earlier would silently skip a later module's jobs
	// until the next tick. Failure is logged, not fatal: degraded background sweeps beat not
	// booting at all.
	if err := job.GetCronjob().Start(); err != nil {
		this.logger.Error("failed to start the cron scheduler; background sweeps will not run", err)
	}

	return nil
}

func (this *Application) registerModelInOrder(moduleMap map[string]modules.InCodeModule, depGraph map[string][]string) error {
	this.logger.Info("Start registering models for modules", nil)

	initOrder, err := topologicalSort(depGraph)
	if err != nil {
		return errors.Wrap(err, "failed to determine model registering order")
	}

	// "core" is already first: every module implicitly depends on it via buildDependencyGraph.
	for _, modName := range initOrder {
		mod := moduleMap[modName]
		if mod == nil {
			continue
		}
		modWithDynamic, ok := mod.(modules.DynamicModule)
		if !ok {
			continue
		}
		if err := modWithDynamic.RegisterModels(); err != nil {
			return err
		}
		this.logger.Infof("Registered models for module %s", mod.Name())
	}

	if err := dmodel.GetSchemaRegistry().FinalizeRelations(); err != nil {
		return errors.Wrap(err, "FinalizeRelations")
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
