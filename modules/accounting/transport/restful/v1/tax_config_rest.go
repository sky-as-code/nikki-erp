package v1

import (
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// The REST structs of the thirteen tax configuration resources: the built-in CRUD over each
// onion's application service. The container names are literals because a struct tag cannot call
// dynamicresource.EngineDependencyName; routes_test.go checks them against it.

const TaxJurisdictionEngineName = "dynengine_accounting_tax_jurisdiction"

type taxJurisdictionRestParams struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_accounting_tax_jurisdiction"`
}

type TaxJurisdictionRest struct {
	dynamicresource.CrudRestBase
}

func NewTaxJurisdictionRest(params taxJurisdictionRestParams) *TaxJurisdictionRest {
	rest := &TaxJurisdictionRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

const TaxGroupEngineName = "dynengine_accounting_tax_group"

type taxGroupRestParams struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_accounting_tax_group"`
}

type TaxGroupRest struct {
	dynamicresource.CrudRestBase
}

func NewTaxGroupRest(params taxGroupRestParams) *TaxGroupRest {
	rest := &TaxGroupRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

const TaxRoundingPolicyEngineName = "dynengine_accounting_tax_rounding_policy"

type taxRoundingPolicyRestParams struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_accounting_tax_rounding_policy"`
}

type TaxRoundingPolicyRest struct {
	dynamicresource.CrudRestBase
}

func NewTaxRoundingPolicyRest(params taxRoundingPolicyRestParams) *TaxRoundingPolicyRest {
	rest := &TaxRoundingPolicyRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

const TaxProductClassificationEngineName = "dynengine_accounting_tax_product_classification"

type taxProductClassificationRestParams struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_accounting_tax_product_classification"`
}

type TaxProductClassificationRest struct {
	dynamicresource.CrudRestBase
}

func NewTaxProductClassificationRest(params taxProductClassificationRestParams) *TaxProductClassificationRest {
	rest := &TaxProductClassificationRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

const TaxEngineName = "dynengine_accounting_tax"

type taxRestParams struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_accounting_tax"`
}

type TaxRest struct {
	dynamicresource.CrudRestBase
}

func NewTaxRest(params taxRestParams) *TaxRest {
	rest := &TaxRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

const TaxDefinitionVersionEngineName = "dynengine_accounting_tax_definition_version"

type taxDefinitionVersionRestParams struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_accounting_tax_definition_version"`
}

type TaxDefinitionVersionRest struct {
	dynamicresource.CrudRestBase
}

func NewTaxDefinitionVersionRest(params taxDefinitionVersionRestParams) *TaxDefinitionVersionRest {
	rest := &TaxDefinitionVersionRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

const TaxRateVersionEngineName = "dynengine_accounting_tax_rate_version"

type taxRateVersionRestParams struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_accounting_tax_rate_version"`
}

type TaxRateVersionRest struct {
	dynamicresource.CrudRestBase
}

func NewTaxRateVersionRest(params taxRateVersionRestParams) *TaxRateVersionRest {
	rest := &TaxRateVersionRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

const TaxComponentEngineName = "dynengine_accounting_tax_component"

type taxComponentRestParams struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_accounting_tax_component"`
}

type TaxComponentRest struct {
	dynamicresource.CrudRestBase
}

func NewTaxComponentRest(params taxComponentRestParams) *TaxComponentRest {
	rest := &TaxComponentRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

const TaxMappingEngineName = "dynengine_accounting_tax_mapping"

type taxMappingRestParams struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_accounting_tax_mapping"`
}

type TaxMappingRest struct {
	dynamicresource.CrudRestBase
}

func NewTaxMappingRest(params taxMappingRestParams) *TaxMappingRest {
	rest := &TaxMappingRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

const TaxMappingLineEngineName = "dynengine_accounting_tax_mapping_line"

type taxMappingLineRestParams struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_accounting_tax_mapping_line"`
}

type TaxMappingLineRest struct {
	dynamicresource.CrudRestBase
}

func NewTaxMappingLineRest(params taxMappingLineRestParams) *TaxMappingLineRest {
	rest := &TaxMappingLineRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

const TaxRuleEngineName = "dynengine_accounting_tax_rule"

type taxRuleRestParams struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_accounting_tax_rule"`
}

type TaxRuleRest struct {
	dynamicresource.CrudRestBase
}

func NewTaxRuleRest(params taxRuleRestParams) *TaxRuleRest {
	rest := &TaxRuleRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

const TaxRuleConditionEngineName = "dynengine_accounting_tax_rule_condition"

type taxRuleConditionRestParams struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_accounting_tax_rule_condition"`
}

type TaxRuleConditionRest struct {
	dynamicresource.CrudRestBase
}

func NewTaxRuleConditionRest(params taxRuleConditionRestParams) *TaxRuleConditionRest {
	rest := &TaxRuleConditionRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

const TaxRuleResultEngineName = "dynengine_accounting_tax_rule_result"

type taxRuleResultRestParams struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_accounting_tax_rule_result"`
}

type TaxRuleResultRest struct {
	dynamicresource.CrudRestBase
}

func NewTaxRuleResultRest(params taxRuleResultRestParams) *TaxRuleResultRest {
	rest := &TaxRuleResultRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}
