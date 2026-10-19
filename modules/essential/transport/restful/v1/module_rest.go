package v1

import (
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// ModuleMetadataEngineName is the container name of the module metadata onion, checked by a test.
const ModuleMetadataEngineName = "dynengine_essential_module_metadata"

type moduleRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_essential_module_metadata"`
}

func NewModuleRest(params moduleRestParams) *ModuleRest {
	rest := &ModuleRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

// ModuleRest serves the reads of the module catalogue. There is no write route: the rows belong
// to the boot-time sync, which writes them through the domain service.
type ModuleRest struct {
	composable.CrudRestBase
}
