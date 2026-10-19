package services

import (
	"sort"

	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
)

// computeDepLevels gives each loaded module its depth in the dependency graph: 0 when it depends
// on no loaded module, otherwise one more than its deepest loaded dependency. A dependency the
// binary did not load is ignored; the loader has already refused a missing required one.
func computeDepLevels(installed []modules.InCodeModule) (map[string]int32, error) {
	depsOf := make(map[string][]string, len(installed))
	for _, mod := range installed {
		depsOf[mod.Name()] = mod.Deps()
	}

	levels := make(map[string]int32, len(installed))
	visiting := make(map[string]bool, len(installed))
	var levelOf func(name string) (int32, error)
	levelOf = func(name string) (int32, error) {
		if level, done := levels[name]; done {
			return level, nil
		}
		if visiting[name] {
			return 0, errors.Errorf("module dependency cycle through %s", name)
		}
		visiting[name] = true
		level := int32(0)
		for _, dep := range depsOf[name] {
			if _, loaded := depsOf[dep]; !loaded {
				continue
			}
			depLevel, err := levelOf(dep)
			if err != nil {
				return 0, err
			}
			level = max(level, depLevel+1)
		}
		visiting[name] = false
		levels[name] = level
		return level, nil
	}

	for name := range depsOf {
		if _, err := levelOf(name); err != nil {
			return nil, err
		}
	}
	return levels, nil
}

// ListModulesByDeps returns every module this deployment has loaded - internal and disabled ones
// included, orphaned ones not, since nothing loaded would answer for them - ordered by
// dependency level, then name. Unlike Search it hides nothing: it serves cross-module
// orchestration, not a screen.
func (this *ModuleDomainServiceImpl) ListModulesByDeps(ctx corectx.Context) ([]models.ModuleMetadata, error) {
	all, err := this.listAllModules(ctx)
	if err != nil {
		return nil, err
	}

	loaded := make([]models.ModuleMetadata, 0, len(all))
	for _, item := range all {
		if !util.ValueOrZeroOf(item.GetIsOrphaned()) {
			loaded = append(loaded, item)
		}
	}
	sort.SliceStable(loaded, func(i, j int) bool {
		levelI, levelJ := util.ValueOrZeroOf(loaded[i].GetDepLevel()), util.ValueOrZeroOf(loaded[j].GetDepLevel())
		if levelI != levelJ {
			return levelI < levelJ
		}
		return util.ValueOrZeroOf(loaded[i].GetName()) < util.ValueOrZeroOf(loaded[j].GetName())
	})
	return loaded, nil
}
