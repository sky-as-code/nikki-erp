package v1

import (
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// SpaceEngineName is the container name of the onion, checked against
// dynamicresource.EngineDependencyName by a test.
const SpaceEngineName = "dynengine_itstorage_space"

type spaceRestParams struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_itstorage_space"`
}

func NewSpaceRest(params spaceRestParams) *SpaceRest {
	rest := &SpaceRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

// SpaceRest serves the read-only built-ins: spaces are system-managed.
type SpaceRest struct {
	dynamicresource.CrudRestBase
}
