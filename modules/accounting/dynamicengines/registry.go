// Package dynamicengines declares the thirteen tax configuration onions Accounting serves through
// the composable resource engine, and registers them into the dependency container during Init().
//
// Each file wires the module's own repository, domain service (which carries the resource's rules)
// and application service onto the composable defaults, and publishes the typed layers so that
// transport, the calculation services and the usage probe inject them by type. Nothing is built
// eagerly: an onion is a container constructor, resolved the first time a consumer asks.
package dynamicengines

import (
	stdErr "errors"

	"github.com/sky-as-code/nikki-erp/modules/accounting/domain/models"
)

// InitDynamicEngines registers every resource onion this module owns.
func InitDynamicEngines() error {
	return stdErr.Join(
		registerTaxJurisdictionEngine(),
		registerTaxGroupEngine(),
		registerTaxRoundingPolicyEngine(),
		registerTaxProductClassificationEngine(),
		registerTaxEngine(),
		registerTaxDefinitionVersionEngine(),
		registerTaxRateVersionEngine(),
		registerTaxComponentEngine(),
		registerTaxMappingEngine(),
		registerTaxMappingLineEngine(),
		registerTaxRuleEngine(),
		registerTaxRuleConditionEngine(),
		registerTaxRuleResultEngine(),
	)
}

// EngineSchemaNames lists the resources Accounting serves, in the dependency order of the schemas.
func EngineSchemaNames() []string {
	return []string{
		models.TaxJurisdictionSchemaName,
		models.TaxGroupSchemaName,
		models.TaxRoundingPolicySchemaName,
		models.TaxProductClassificationSchemaName,
		models.TaxSchemaName,
		models.TaxDefinitionVersionSchemaName,
		models.TaxRateVersionSchemaName,
		models.TaxComponentSchemaName,
		models.TaxMappingSchemaName,
		models.TaxMappingLineSchemaName,
		models.TaxRuleSchemaName,
		models.TaxRuleConditionSchemaName,
		models.TaxRuleResultSchemaName,
	}
}
