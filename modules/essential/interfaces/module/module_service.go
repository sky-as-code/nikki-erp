// Package module declares the layers of the essential_module_metadata resource: one row per
// module the binary has ever loaded, kept in step with the loader by SyncModuleMetadata at boot.
package module

import (
	"github.com/sky-as-code/nikki-erp/modules"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
)

// ModuleRepository adds the advisory lock that keeps two instances from syncing at once.
type ModuleRepository interface {
	composable.CrudRepository
	AcquireLock(ctx corectx.Context) (bool, error)
	ReleaseLock(ctx corectx.Context) error
}

type ModuleDomainService interface {
	composable.CrudDomainService
	ModuleCatalogService
	SyncModuleMetadata(ctx corectx.Context, installedModules []modules.InCodeModule) (bool, error)
}

// ModuleCatalogService lists the loaded modules for cross-module orchestration (org lifecycle).
// Service-only: it has no REST route and asserts nothing.
type ModuleCatalogService interface {
	ListModulesByDeps(ctx corectx.Context) ([]models.ModuleMetadata, error)
}

// ModuleAppService is what module start-up binds to.
type ModuleAppService interface {
	SyncModuleMetadata(ctx corectx.Context, installedModules []modules.InCodeModule) (bool, error)
}

// ModuleApplicationService is the whole application layer: the reads the REST handler serves
// (writes are service-only, the rows belong to the boot-time sync) plus the sync itself.
type ModuleApplicationService interface {
	composable.CrudApplicationService
	ModuleAppService
	ModuleCatalogService
}

type (
	CreateModuleCommand = composable.CreateCommand
	UpdateModuleCommand = composable.UpdateCommand
	GetModuleByIdQuery  = composable.GetByIdQuery
	SearchModulesQuery  = composable.SearchQuery
	ModuleExistsQuery   = composable.ExistsQuery
)

type (
	CreateModuleResult  = composable.CreateResult
	UpdateModuleResult  = composable.MutateResult
	GetModuleByIdResult = composable.GetOneResult
	SearchModulesResult = composable.SearchResult
	ModuleExistsResult  = composable.ExistsResult
)
