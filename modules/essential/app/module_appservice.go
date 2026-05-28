package app

import (
	"github.com/sky-as-code/nikki-erp/modules"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	it "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/module"
)

func NewModuleApplicationServiceImpl(moduleSvc it.ModuleDomainService) it.ModuleAppService {
	return &ModuleApplicationServiceImpl{moduleSvc: moduleSvc}
}

type ModuleApplicationServiceImpl struct {
	moduleSvc it.ModuleDomainService
}

func (this *ModuleApplicationServiceImpl) CreateModule(ctx corectx.Context, cmd it.CreateModuleCommand) (*it.CreateModuleResult, error) {
	return this.moduleSvc.CreateModule(ctx, cmd)
}

func (this *ModuleApplicationServiceImpl) DeleteModule(ctx corectx.Context, cmd it.DeleteModuleCommand) (*it.DeleteModuleResult, error) {
	return this.moduleSvc.DeleteModule(ctx, cmd)
}

func (this *ModuleApplicationServiceImpl) ModuleExists(ctx corectx.Context, query it.ModuleExistsQuery) (*it.ModuleExistsResult, error) {
	return this.moduleSvc.ModuleExists(ctx, query)
}

func (this *ModuleApplicationServiceImpl) GetModule(ctx corectx.Context, query it.GetModuleQuery) (*it.GetModuleResult, error) {
	return this.moduleSvc.GetModule(ctx, query)
}

func (this *ModuleApplicationServiceImpl) SearchModules(ctx corectx.Context, query it.SearchModulesQuery) (*it.SearchModulesResult, error) {
	return this.moduleSvc.SearchModules(ctx, query)
}

func (this *ModuleApplicationServiceImpl) UpdateModule(ctx corectx.Context, cmd it.UpdateModuleCommand) (*it.UpdateModuleResult, error) {
	return this.moduleSvc.UpdateModule(ctx, cmd)
}

func (this *ModuleApplicationServiceImpl) SyncModuleMetadata(ctx corectx.Context, installedModules []modules.InCodeModule) (bool, error) {
	return this.moduleSvc.SyncModuleMetadata(ctx, installedModules)
}
