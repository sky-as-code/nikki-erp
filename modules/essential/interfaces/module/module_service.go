package module

import (
	"github.com/sky-as-code/nikki-erp/common/module"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

type ModuleService interface {
	CreateModule(ctx crud.Context, cmd CreateModuleCommand) (*CreateModuleResult, error)
	DeleteModule(ctx crud.Context, cmd DeleteModuleCommand) (*DeleteModuleResult, error)
	ModuleExists(ctx crud.Context, cmd ModuleExistsQuery) (*ModuleExistsResult, error)
	GetModule(ctx crud.Context, query GetModuleByIdQuery) (result *GetModuleResult, err error)
	ListModules(ctx crud.Context, query ListModulesQuery) (result *ListModulesResult, err error)
	UpdateModule(ctx crud.Context, cmd UpdateModuleCommand) (*UpdateModuleResult, error)
	SyncModuleMetadata(ctx crud.Context, installedModules []module.InCodeModule) (bool, error)
}
