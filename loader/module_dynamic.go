// This file is the dynamic-plugin loader, kept commented out until plugin loading is revived.
//
// When it is revived it needs its build constraint restored as the first line of the file: the
// go:build form naming the dynamicmods tag, plus the legacy paired line for older toolchains.
//
// The constraint is described here rather than written out, because vet matches a constraint-shaped
// comment anywhere in a file's leading comment block even when it is commented out. It then reports
// "possible malformed build comment" and fails `go test` for the entire loader package, while
// `go build` still passes — so the damage shows up only as a package nobody can test.

package loader

// import (
// 	"os"
// 	"path/filepath"
// 	"plugin"

// 	"github.com/sky-as-code/nikki-erp/common"
// 	"github.com/sky-as-code/nikki-erp/common/env"
// 	. "github.com/sky-as-code/nikki-erp/common/fault"
// 	mods "github.com/sky-as-code/nikki-erp/modules"
// 	"github.com/sky-as-code/nikki-erp/modules/core/logging"
// )

// func LoadModules(logger logging.LoggerService) ([]mods.NikkiModule, error) {
// 	// Common module is always loaded
// 	allMods := []mods.NikkiModule{
// 		common.ModuleSingleton,
// 	}
// 	otherMods, err := getDynamicModules(logger)
// 	if err != nil {
// 		return nil, err
// 	}
// 	allMods = append(allMods, otherMods...)
// 	return allMods, nil
// }

// // getDynamicModules loads .so files from modules directory
// func getDynamicModules(logger logging.LoggerService) ([]mods.NikkiModule, error) {
// 	entries, err := os.ReadDir(getPluginsDir())
// 	if err != nil {
// 		return handleReadDirError(err, logger)
// 	}

// 	return loadGoPlugins(entries, logger)
// }

// func handleReadDirError(err error, logger logging.LoggerService) ([]mods.NikkiModule, error) {
// 	if os.IsNotExist(err) {
// 		logger.Warn("Plugins directory not found, skipping plugin loading", err)
// 		return []mods.NikkiModule{}, nil
// 	}
// 	return nil, NewTechnicalError("failed to read plugins directory: %v", err)
// }

// func loadGoPlugins(entries []os.DirEntry, logger logging.LoggerService) ([]mods.NikkiModule, error) {
// 	modules := []mods.NikkiModule{}
// 	for _, entry := range entries {
// 		if isValidPlugin(entry) {
// 			if module, err := loadSinglePlugin(entry, logger); err == nil {
// 				modules = append(modules, module)
// 			}
// 		}
// 	}
// 	return modules, nil
// }

// func isValidPlugin(entry os.DirEntry) bool {
// 	return !entry.IsDir() && filepath.Ext(entry.Name()) == ".so"
// }

// func loadSinglePlugin(entry os.DirEntry, logger logging.LoggerService) (mods.NikkiModule, error) {
// 	pluginPath := filepath.Join(getPluginsDir(), entry.Name())
// 	p, err := plugin.Open(pluginPath)
// 	if err != nil {
// 		logger.Errorf("Failed to load plugin %s: %v", entry.Name(), err)
// 		return nil, err
// 	}

// 	return lookupModuleSymbol(p, entry.Name(), logger)
// }

// func lookupModuleSymbol(p *plugin.Plugin, pluginName string, logger logging.LoggerService) (mods.NikkiModule, error) {
// 	moduleSymbol, err := p.Lookup("Module")
// 	if err != nil {
// 		logger.Errorf("Plugin %s does not export 'Module' symbol: %v. Skip this plugin", pluginName, err)
// 		return nil, err
// 	}

// 	return validateModuleSymbol(moduleSymbol, pluginName, logger)
// }

// func validateModuleSymbol(moduleSymbol plugin.Symbol, pluginName string, logger logging.LoggerService) (mods.NikkiModule, error) {
// 	if module, ok := moduleSymbol.(mods.NikkiModule); ok {
// 		logger.Infof("Successfully loaded plugin: %s", module.Name())
// 		return module, nil
// 	}
// 	logger.Warnf("Plugin %s: 'Module' symbol does not implement Module interface. Skip this plugin", pluginName)
// 	return nil, nil
// }

// func getPluginsDir(workingDir ...string) string {
// 	cwd := env.Cwd()
// 	if len(workingDir) > 0 && workingDir[0] != "" {
// 		cwd = workingDir[0]
// 	}
// 	return filepath.Join(cwd, "modules")
// }
