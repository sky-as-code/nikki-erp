// Package dynamicengines declares the resource engines the Contacts module serves through the
// dynamic resource engine, and creates them during the module's Init().
//
// Its imports point one way only: the domain, the module's own interfaces, and the dynamicresource
// module — never app/, infra/ or transport/. That keeps the package importable by both contacts
// (which creates the engines) and contacts/transport/restful (which registers their routes) without
// a cycle.
//
// All three resources are plain CRUD: there is no lifecycle, no state machine, and nothing a client
// can ask a party to do beyond creating, reading, updating, archiving and deleting it. So no spec
// defines actions and no derived resource service is installed.
package dynamicengines

import (
	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/array"
	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/contacts/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
)

// engineSpec declares one resource engine the Contacts module owns.
type engineSpec struct {
	// SchemaName is the dynamic-model schema the engine serves. It must be an XSchemaName
	// constant, never a string derived from the resource path.
	SchemaName string

	// DefaultFields is the field set a listing search returns. Primary key fields are always
	// included by the query builder, so listing them here is redundant.
	DefaultFields []string

	// DefineActions adds resource-specific actions and validation on top of the built-in CRUD
	// ones. It is optional: a resource without custom behavior leaves it nil.
	DefineActions func(drif.DynamicResourceEngine) error
}

// engineSpecs lists the resources this module serves through the dynamic resource engine, each
// with the field set its listing UI needs.
var engineSpecs = []engineSpec{
	partyEngineSpec(),
	commChannelEngineSpec(),
	relationshipEngineSpec(),
	vendorProfileEngineSpec(),
}

func partyEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.PartySchemaName,
		DefaultFields: []string{
			models.PartyFieldDisplayName,
			models.PartyFieldLegalName,
			models.PartyFieldType,
			models.PartyFieldTaxId,
			models.PartyFieldOrgId,
			models.PartyFieldIsArchived,
		},
	}
}

func commChannelEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.CommChannelSchemaName,
		DefaultFields: []string{
			models.CommChannelFieldPartyId,
			models.CommChannelFieldType,
			models.CommChannelFieldValue,
			models.CommChannelFieldOrgId,
			models.CommChannelFieldIsArchived,
		},
	}
}

func relationshipEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.RelationshipSchemaName,
		DefaultFields: []string{
			models.RelationshipFieldPartyId,
			models.RelationshipFieldTargetPartyId,
			models.RelationshipFieldType,
			models.RelationshipFieldIsArchived,
		},
	}
}

// vendorProfileEngineSpec serves the supplier-specific facts hanging off a party.
//
// Plain CRUD like its siblings: qualification status is an ordinary enum field in this phase, not a
// state machine, so there is no transition for a domain action to protect. The rule that only an
// active vendor may be ordered from belongs to Purchase and is enforced there, at confirmation.
func vendorProfileEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.VendorProfileSchemaName,
		DefaultFields: []string{
			models.VendorProfileFieldPartyId,
			models.VendorProfileFieldStatus,
			models.VendorProfileFieldDefaultCurrencyId,
			models.VendorProfileFieldPaymentTerms,
			models.VendorProfileFieldLeadTimeDays,
			models.VendorProfileFieldOrgId,
			models.VendorProfileFieldIsArchived,
		},
	}
}

// EngineSchemaNames lists the schemas this module creates an engine for, so that route
// registration and engine creation cannot drift apart.
func EngineSchemaNames() []string {
	return array.Map(engineSpecs, func(spec engineSpec) string {
		return spec.SchemaName
	})
}

// InitDynamicEngines creates the resource engines this module owns and publishes them into the
// dependency container, so that other modules can inject them by name.
func InitDynamicEngines() error {
	for _, spec := range engineSpecs {
		if err := initEngine(spec); err != nil {
			return err
		}
	}
	return nil
}

func initEngine(spec engineSpec) error {
	engine, err := dynamicresource.Registry().NewEngine(spec.SchemaName, drif.NewEngineOptions{
		DefaultSearchFields: spec.DefaultFields,
	})
	if err != nil {
		return errors.Wrapf(err, "failed to create the '%s' resource engine", spec.SchemaName)
	}

	if spec.DefineActions != nil {
		if err := spec.DefineActions(engine); err != nil {
			return errors.Wrapf(err, "failed to define actions of the '%s' resource engine", spec.SchemaName)
		}
	}

	err = deps.RegisterNamed(
		dynamicresource.EngineDependencyName(spec.SchemaName),
		func() drif.DynamicResourceEngine { return engine },
	)
	return errors.Wrapf(err, "failed to register the '%s' resource engine", spec.SchemaName)
}
