package app

import (
	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/modules"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/module"
)

// readableBySession is the empty action code: AssertPermission skips the entitlement check for
// it, leaving the session the route's middleware already verified.
const readableBySession = ""

// NewModuleApplicationService is handed the composable default by the module metadata onion.
func NewModuleApplicationService(base composable.CrudApplicationService) it.ModuleApplicationService {
	moduleSvc, ok := base.DomainService().(it.ModuleDomainService)
	if !ok {
		panic(errors.New("the module metadata onion must be built with NewModuleDomainService"))
	}
	return &ModuleApplicationServiceImpl{CrudApplicationService: base, moduleSvc: moduleSvc}
}

// ModuleApplicationServiceImpl serves the module catalogue every signed-in user's shell builds its
// navigation from. Reading it needs a session but no entitlement: no permission is seeded for
// essential_module_metadata, and asserting "read" would leave every non-owner with an empty shell.
// The reads still open with AssertAction, with an empty action code, so the org scoping and the
// call shape stay those of every other application method.
type ModuleApplicationServiceImpl struct {
	composable.CrudApplicationService
	moduleSvc it.ModuleDomainService
}

func (this *ModuleApplicationServiceImpl) Search(
	ctx corectx.Context, query it.SearchModulesQuery,
) (*it.SearchModulesResult, error) {
	if _, cErrs := this.AssertAction(ctx, readableBySession, query); cErrs != nil {
		return &it.SearchModulesResult{ClientErrors: *cErrs}, nil
	}
	return this.moduleSvc.Search(ctx, query)
}

func (this *ModuleApplicationServiceImpl) GetById(
	ctx corectx.Context, query it.GetModuleByIdQuery,
) (*it.GetModuleByIdResult, error) {
	if _, cErrs := this.AssertAction(ctx, readableBySession, query); cErrs != nil {
		return &it.GetModuleByIdResult{ClientErrors: *cErrs}, nil
	}
	return this.moduleSvc.GetById(ctx, query)
}

func (this *ModuleApplicationServiceImpl) GetSchema(
	ctx corectx.Context, query composable.GetSchemaQuery,
) (*composable.GetSchemaResult, error) {
	if _, cErrs := this.AssertAction(ctx, readableBySession, query); cErrs != nil {
		return &composable.GetSchemaResult{ClientErrors: *cErrs}, nil
	}
	return &composable.GetSchemaResult{Data: this.Schema().ToSimplized(), HasData: true}, nil
}

// SyncModuleMetadata runs once at start-up with no request behind it, so it asserts nothing.
func (this *ModuleApplicationServiceImpl) SyncModuleMetadata(
	ctx corectx.Context, installedModules []modules.InCodeModule,
) (bool, error) {
	return this.moduleSvc.SyncModuleMetadata(ctx, installedModules)
}

// ListModulesByDeps is a cross-module port (IAM's org lifecycle), so it asserts nothing.
func (this *ModuleApplicationServiceImpl) ListModulesByDeps(ctx corectx.Context) ([]models.ModuleMetadata, error) {
	return this.moduleSvc.ListModulesByDeps(ctx)
}
