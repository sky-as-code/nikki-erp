package module

import (
	"github.com/sky-as-code/nikki-erp/modules"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

type ModuleService interface {
	CreateModule(ctx corectx.Context, cmd CreateModuleCommand) (*CreateModuleResult, error)
	DeleteModule(ctx corectx.Context, cmd DeleteModuleCommand) (*DeleteModuleResult, error)
	ModuleExists(ctx corectx.Context, query ModuleExistsQuery) (*ModuleExistsResult, error)
	GetModule(ctx corectx.Context, query GetModuleQuery) (*GetModuleResult, error)
	SearchModules(ctx corectx.Context, query SearchModulesQuery) (*SearchModulesResult, error)
	UpdateModule(ctx corectx.Context, cmd UpdateModuleCommand) (*UpdateModuleResult, error)
	SyncModuleMetadata(ctx corectx.Context, installedModules []modules.InCodeModule) (bool, error)
}
