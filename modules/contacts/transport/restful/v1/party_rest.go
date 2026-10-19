package v1

import (
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// PartyEngineName is the container name of the onion. A struct tag cannot call
// composable.EngineDependencyName, so the literal is declared once here and checked by a test.
const PartyEngineName = "dynengine_contacts_party"

type partyRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_contacts_party"`
}

func NewPartyRest(params partyRestParams) *PartyRest {
	rest := &PartyRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

// PartyRest serves the built-in CRUD of the resource; it has no route of its own.
type PartyRest struct {
	composable.CrudRestBase
}
