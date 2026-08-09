package cmd

import (
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
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	essentialconstants "github.com/sky-as-code/nikki-erp/modules/essential/constants"
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

	prefixes := schemaPrefixesOf(moduleName)
	queries, err := orm.GenCreateSql(registry, dialect, prefixes...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to generate create SQL: %v\n", err)
		os.Exit(1)
	}

	// A module that registered schemas but matched no prefix means schemaPrefixesOf is wrong
	// for it, which would otherwise surface as a silently empty migration. A module that
	// registered nothing at all (report, for example) legitimately produces no SQL.
	if len(queries) == 0 && registeredCount > 0 {
		fmt.Fprintf(os.Stderr,
			"module '%s' registered %d schema(s) but none match [%s]; "+
				"fix schemaPrefixesOf in cmd/application.go\n",
			moduleName, registeredCount, strings.Join(prefixes, ", "))
		os.Exit(1)
	}
	return strings.Join(queries, ";\n")
}

// schemaPrefixesOf maps a module name to the prefixes its schemas are named with, so
// -createsql emits only that module's tables while its dependencies stay registered as FK
// targets.
//
// Most modules name their schemas "{module}_...", but not all, and a wrong prefix silently
// drops tables from the migration rather than failing loudly. Two exceptions today:
//
//   - vending_machine names its schemas "vending_..."
//   - iam owns "authenticate_password_store" as well as its "iam_" schemas, left over from
//     the Identity/Authenticate merge
//
// Add an entry when a new module's schema names do not all start with its module name.
func schemaPrefixesOf(moduleName string) []string {
	switch moduleName {
	case "vending_machine":
		return []string{"vending_"}
	case "iam":
		return []string{"iam_", "authenticate_"}
	default:
		return []string{moduleName + "_"}
	}
}

// registerModelsUpTo registers models for the target module and its transitive
// dependencies, in dependency order.
//
// The normal app path does this via registerModelInOrder, but -createsql cannot reuse it:
// that method logs, and GenSql runs with a nil logger (see runCreateSql). Registering only
// the target module is not enough either — base schemas such as "core.basemodel.base_model"
// are put into the builder registry by CoreModule.RegisterModels, and a JSON model that
// names one in "extend_before" panics when it has not run yet.
//
// Only the dependency closure is registered. Selecting by position in the full topological
// order would be non-deterministic: topologicalSort seeds its walk from a map range, so
// unrelated modules land before the target in some runs and after it in others, making the
// generated SQL differ between two runs of the same command.
//
// Returns how many schemas moduleName itself added to the registry, which lets the caller
// tell "this module owns no tables" from "the prefix filter matched nothing".
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

// dependencyClosure returns root and everything it transitively depends on, ordered
// dependencies-first, so "core" registers its base schemas before any module that extends
// them by name. Modules outside root's closure are excluded entirely.
//
// The result is deterministic for a given graph: the walk starts at root rather than at an
// arbitrary map key, and each node's dependencies are visited in their declared order.
// Each module appears exactly once, which matters because RegisterSchemaBuilderFn rejects
// duplicate registration.
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
			modName != coreconstants.CoreModuleName &&
			modName != essentialconstants.EssentialModuleName {
			deps = append(deps, essentialconstants.EssentialModuleName)
		}
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

	return nil
}

func (this *Application) registerModelInOrder(moduleMap map[string]modules.InCodeModule, depGraph map[string][]string) error {
	this.logger.Info("Start registering models for modules", nil)

	initOrder, err := topologicalSort(depGraph)
	if err != nil {
		return errors.Wrap(err, "failed to determine model registering order")
	}

	// No need to force "core" to the front: buildDependencyGraph makes every other module
	// implicitly depend on it, and core itself depends on "apptrait", so the sort already
	// yields ["apptrait", "core", ...]. Prepending it here would visit it twice and
	// register its schemas twice.
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
