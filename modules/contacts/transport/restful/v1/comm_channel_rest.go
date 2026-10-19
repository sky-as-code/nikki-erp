package v1

import (
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// CommChannelEngineName is the container name of the onion. A struct tag cannot call
// composable.EngineDependencyName, so the literal is declared once here and checked by a test.
const CommChannelEngineName = "dynengine_contacts_comm_channel"

type commChannelRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_contacts_comm_channel"`
}

func NewCommChannelRest(params commChannelRestParams) *CommChannelRest {
	rest := &CommChannelRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

// CommChannelRest serves the built-in CRUD of the resource; it has no route of its own.
type CommChannelRest struct {
	composable.CrudRestBase
}
