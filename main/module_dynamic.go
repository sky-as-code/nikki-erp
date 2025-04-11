//go:build dynamicmods
// +build dynamicmods

package main

import (
	"os"
	"path/filepath"
	"plugin"

	"github.com/sky-as-code/nikki-erp/modules/shared"
	. "github.com/sky-as-code/nikki-erp/utility/fault"
)

func (thisApp *Application) getModules() ([]shared.NikkiModule, AppError) {
	return thisApp.getDynamicModules()
}

func (thisApp *Application) getDynamicModules() ([]shared.NikkiModule, AppError) {
	modules := []shared.NikkiModule{}
	pluginsDir := "./app/modules"
	entries, err := os.ReadDir(pluginsDir)
	if err != nil {
		if os.IsNotExist(err) {
			thisApp.logger.Warn("Plugins directory not found, skipping plugin loading")
			return modules, nil
		}
		return nil, NewTechnicalError("failed to read plugins directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".so" {
			pluginPath := filepath.Join(pluginsDir, entry.Name())
			p, err := plugin.Open(pluginPath)
			if err != nil {
				thisApp.logger.Errorf("Failed to load plugin %s: %v", entry.Name(), err)
				continue
			}

			moduleSymbol, err := p.Lookup("Module")
			if err != nil {
				thisApp.logger.Errorf("Plugin %s does not export 'Module' symbol: %v. Skip this plugin", entry.Name(), err)
				continue
			}

			if module, ok := moduleSymbol.(shared.NikkiModule); ok {
				modules = append(modules, module)
				thisApp.logger.Infof("Successfully loaded plugin: %s", module.Name())
			} else {
				thisApp.logger.Warnf("Plugin %s: 'Module' symbol does not implement Module interface. Skip this plugin", entry.Name())
			}
		}
	}
	return modules, nil
}
