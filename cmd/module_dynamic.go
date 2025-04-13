//go:build dynamicmods
// +build dynamicmods

package main

import (
	"os"
	"path/filepath"
	"plugin"

	"github.com/sky-as-code/nikki-erp/common"
	"github.com/sky-as-code/nikki-erp/common/util/env"
	. "github.com/sky-as-code/nikki-erp/common/util/fault"
	mods "github.com/sky-as-code/nikki-erp/modules"
)

func (thisApp *Application) getModules() ([]mods.NikkiModule, AppError) {
	// Common module is always loaded
	allMods := []mods.NikkiModule{
		common.ModuleSingleton,
	}
	otherMods, err := thisApp.getDynamicModules()
	if err != nil {
		return nil, err
	}
	allMods = append(allMods, otherMods...)
	return allMods, nil
}

// getDynamicModules loads .so files from modules directory
func (thisApp *Application) getDynamicModules() ([]mods.NikkiModule, AppError) {
	entries, err := os.ReadDir(thisApp.getPluginsDir())
	if err != nil {
		return thisApp.handleReadDirError(err)
	}

	return thisApp.loadGoPlugins(entries)
}

func (thisApp *Application) handleReadDirError(err error) ([]mods.NikkiModule, AppError) {
	if os.IsNotExist(err) {
		thisApp.logger.Warn("Plugins directory not found, skipping plugin loading")
		return []mods.NikkiModule{}, nil
	}
	return nil, NewTechnicalError("failed to read plugins directory: %v", err)
}

func (thisApp *Application) loadGoPlugins(entries []os.DirEntry) ([]mods.NikkiModule, AppError) {
	modules := []mods.NikkiModule{}
	for _, entry := range entries {
		if thisApp.isValidPlugin(entry) {
			if module, err := thisApp.loadSinglePlugin(entry); err == nil {
				modules = append(modules, module)
			}
		}
	}
	return modules, nil
}

func (thisApp *Application) isValidPlugin(entry os.DirEntry) bool {
	return !entry.IsDir() && filepath.Ext(entry.Name()) == ".so"
}

func (thisApp *Application) loadSinglePlugin(entry os.DirEntry) (mods.NikkiModule, error) {
	pluginPath := filepath.Join(thisApp.getPluginsDir(), entry.Name())
	p, err := plugin.Open(pluginPath)
	if err != nil {
		thisApp.logger.Errorf("Failed to load plugin %s: %v", entry.Name(), err)
		return nil, err
	}

	return thisApp.lookupModuleSymbol(p, entry.Name())
}

func (thisApp *Application) lookupModuleSymbol(p *plugin.Plugin, pluginName string) (mods.NikkiModule, error) {
	moduleSymbol, err := p.Lookup("Module")
	if err != nil {
		thisApp.logger.Errorf("Plugin %s does not export 'Module' symbol: %v. Skip this plugin", pluginName, err)
		return nil, err
	}

	return thisApp.validateModuleSymbol(moduleSymbol, pluginName)
}

func (thisApp *Application) validateModuleSymbol(moduleSymbol plugin.Symbol, pluginName string) (mods.NikkiModule, error) {
	if module, ok := moduleSymbol.(mods.NikkiModule); ok {
		thisApp.logger.Infof("Successfully loaded plugin: %s", module.Name())
		return module, nil
	}
	thisApp.logger.Warnf("Plugin %s: 'Module' symbol does not implement Module interface. Skip this plugin", pluginName)
	return nil, nil
}

func (thisApp *Application) getPluginsDir(workingDir ...string) string {
	cwd := env.Cwd()
	if len(workingDir) > 0 && workingDir[0] != "" {
		cwd = workingDir[0]
	}
	return filepath.Join(cwd, "modules")
}
