package v1

import (
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// VendorProfileEngineName is the container name of the onion. A struct tag cannot call
// composable.EngineDependencyName, so the literal is declared once here and checked by a test.
const VendorProfileEngineName = "dynengine_contacts_vendor_profile"

type vendorProfileRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_contacts_vendor_profile"`
}

func NewVendorProfileRest(params vendorProfileRestParams) *VendorProfileRest {
	rest := &VendorProfileRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

// VendorProfileRest serves the built-in CRUD of the resource; it has no route of its own.
type VendorProfileRest struct {
	composable.CrudRestBase
}
