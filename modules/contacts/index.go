// Package contacts holds the people and organizations the business deals with.
//
// One resource, contacts_party, covers both — a company is frequently both a customer and a
// supplier, and duplicating it would give it two addresses and two tax ids that drift apart. What a
// party *is to the business* is expressed by profile records that hang off it rather than by a type
// column on the party itself.
package contacts

import (
	stdErr "errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/semver"
	"github.com/sky-as-code/nikki-erp/modules"
	modconstants "github.com/sky-as-code/nikki-erp/modules/contacts/constants"
	"github.com/sky-as-code/nikki-erp/modules/contacts/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/contacts/domain/services"
	"github.com/sky-as-code/nikki-erp/modules/contacts/dynamicengines"
	"github.com/sky-as-code/nikki-erp/modules/contacts/transport/restful"
)

// ModuleSingleton is the exported symbol that will be looked up by the plugin loader.
//
// It is typed DynamicModule rather than InCodeModule so that dropping RegisterModels fails the
// build. Under the wider interface the method is found by a type assertion instead, and a module
// that has lost it still compiles, still loads, and silently registers no schemas at all.
var ModuleSingleton modules.DynamicModule = &ContactsModule{}

type ContactsModule struct {
}

// LabelKey implements NikkiModule.
func (*ContactsModule) LabelKey() string {
	return "contacts.moduleLabel"
}

// Name implements NikkiModule.
func (*ContactsModule) Name() string {
	return modconstants.ContactsModuleName
}

// Deps implements NikkiModule.
func (*ContactsModule) Deps() []string {
	return []string{
		"dynamicresource",
	}
}

// IsInternal implements InCodeModule.
func (*ContactsModule) IsInternal() bool {
	return false
}

// Version implements NikkiModule.
func (*ContactsModule) Version() semver.SemVer {
	return *semver.MustParseSemVer("v1.1.0")
}

// Init implements NikkiModule.
//
// The order is load-bearing: the engines must exist before the vendor service that reads through
// them, and before transport registers their routes.
func (*ContactsModule) Init() error {
	if err := dynamicengines.InitDynamicEngines(); err != nil {
		return err
	}
	if err := services.InitDomainServices(); err != nil {
		return err
	}
	return restful.InitRestfulHandlers()
}

// RegisterModels implements DynamicModule.
//
// Schemas are registered referenced-before-referencing: the party is pointed at by the
// communication channel, the relationship and the vendor profile, and an edge is resolved against
// the schema registry at registration time.
func (*ContactsModule) RegisterModels() error {
	return stdErr.Join(
		dmodel.RegisterSchemaB(models.PartySchemaBuilder()),
		dmodel.RegisterSchemaB(models.CommChannelSchemaBuilder()),
		dmodel.RegisterSchemaB(models.RelationshipSchemaBuilder()),
		dmodel.RegisterSchemaB(models.VendorProfileSchemaBuilder()),
	)
}
