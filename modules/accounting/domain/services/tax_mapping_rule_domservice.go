package services

import (
	"github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/taxconfig"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// The domain services of the determination resources: mappings and rules (ACC-02).

type TaxMappingDomainServiceImpl struct{ ruledDomainService }

func NewTaxMappingDomainService(base composable.CrudDomainService) taxconfig.TaxMappingDomainService {
	return &TaxMappingDomainServiceImpl{newRuled(base, taxRuleSet{
		create: validateMappingCreate,
		update: validateMappingUpdate,
		delete: validateMappingDelete,
	})}
}

// TaxMappingLineDomainServiceImpl reads the parent mapping through the mapping repository: a
// repository answers only for its own schema.
type TaxMappingLineDomainServiceImpl struct{ ruledDomainService }

func NewTaxMappingLineDomainService(
	base composable.CrudDomainService, mappings taxconfig.TaxMappingRepository,
) taxconfig.TaxMappingLineDomainService {
	return &TaxMappingLineDomainServiceImpl{newRuled(base, taxRuleSet{
		create: validateMappingLineCreate,
		update: validateMappingLineUpdate(mappings),
	})}
}

type TaxRuleDomainServiceImpl struct{ ruledDomainService }

func NewTaxRuleDomainService(base composable.CrudDomainService) taxconfig.TaxRuleDomainService {
	return &TaxRuleDomainServiceImpl{newRuled(base, taxRuleSet{
		create: validateRuleCreate,
		update: validateRuleUpdate,
		delete: validateRuleDelete,
	})}
}

type TaxRuleConditionDomainServiceImpl struct{ ruledDomainService }

func NewTaxRuleConditionDomainService(base composable.CrudDomainService) taxconfig.TaxRuleConditionDomainService {
	return &TaxRuleConditionDomainServiceImpl{newRuled(base, taxRuleSet{
		create: validateRuleConditionWrite,
		update: validateRuleConditionUpdate,
	})}
}

type TaxRuleResultDomainServiceImpl struct{ ruledDomainService }

func NewTaxRuleResultDomainService(base composable.CrudDomainService) taxconfig.TaxRuleResultDomainService {
	return &TaxRuleResultDomainServiceImpl{newRuled(base, taxRuleSet{
		create: validateRuleResultWrite,
		update: validateRuleResultUpdate,
	})}
}
