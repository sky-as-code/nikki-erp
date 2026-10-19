package v1

import (
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// RelationshipEngineName is the container name of the onion. A struct tag cannot call
// composable.EngineDependencyName, so the literal is declared once here and checked by a test.
const RelationshipEngineName = "dynengine_contacts_relationship"

type relationshipRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_contacts_relationship"`
}

func NewRelationshipRest(params relationshipRestParams) *RelationshipRest {
	rest := &RelationshipRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

// RelationshipRest serves the built-in CRUD of the resource; it has no route of its own.
type RelationshipRest struct {
	composable.CrudRestBase
}
