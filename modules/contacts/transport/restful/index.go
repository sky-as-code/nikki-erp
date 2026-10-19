package restful

import (
	stdErr "errors"

	"github.com/labstack/echo/v5"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/contacts/domain/models"
	v1 "github.com/sky-as-code/nikki-erp/modules/contacts/transport/restful/v1"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

func InitRestfulHandlers() error {
	err := deps.Register(
		v1.NewPartyRest,
		v1.NewCommChannelRest,
		v1.NewRelationshipRest,
		v1.NewVendorProfileRest,
	)
	return stdErr.Join(err, initContactsV1())
}

// initContactsV1 mounts every Contacts resource at /v1/contacts/{schema_name} with the built-in
// route table. A party's channels and relationships are not nested paths: they are graph filters
// on party_id, and the organization is the org_id query/body parameter.
func initContactsV1() error {
	return deps.Invoke(func(route *echo.Group) error {
		routeV1 := route.Group("/v1/contacts")
		return deps.Invoke(func(
			party *v1.PartyRest, channel *v1.CommChannelRest,
			relationship *v1.RelationshipRest, vendorProfile *v1.VendorProfileRest,
		) error {
			return stdErr.Join(
				registerCrud(routeV1, models.PartySchemaName, party),
				registerCrud(routeV1, models.CommChannelSchemaName, channel),
				registerCrud(routeV1, models.RelationshipSchemaName, relationship),
				registerCrud(routeV1, models.VendorProfileSchemaName, vendorProfile),
			)
		})
	})
}

func registerCrud(route *echo.Group, schemaName string, handlers composable.CrudRestHandlers) error {
	return composable.NewRestEngine(schemaName, handlers).AddCrudRoutes().RegisterRoutes(route)
}
