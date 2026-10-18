package services

import (
	"github.com/sky-as-code/nikki-erp/common/safe"
	"github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/taxconfig"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// The domain services of the thirteen tax configuration resources. Each is the composable default
// with the resource's rules (tax_rules_*.go) chained into its own Create, Update and Delete, so the
// rules hold for a direct service call as well as for a REST one. A nil rule is skipped by
// ChainValidateExtra.

type taxRuleSet struct {
	create composable.ValidateExtraFn
	update composable.ValidateExtraFn
	delete composable.ValidateExtraFn
}

type ruledDomainService struct {
	composable.CrudDomainService
	rules taxRuleSet
}

func newRuled(base composable.CrudDomainService, rules taxRuleSet) ruledDomainService {
	return ruledDomainService{CrudDomainService: base, rules: rules}
}

func (this *ruledDomainService) Create(
	ctx corectx.Context, cmd composable.CreateCommand, options ...composable.CreateOptions,
) (*composable.CreateResult, error) {
	opts := safe.GetOptional(options, composable.CreateOptions{})
	opts.ValidateExtra = composable.ChainValidateExtra(opts.ValidateExtra, this.rules.create)
	return this.CrudDomainService.Create(ctx, cmd, opts)
}

func (this *ruledDomainService) Update(
	ctx corectx.Context, cmd composable.UpdateCommand, options ...composable.UpdateOptions,
) (*composable.MutateResult, error) {
	opts := safe.GetOptional(options, composable.UpdateOptions{})
	opts.ValidateExtra = composable.ChainValidateExtra(opts.ValidateExtra, this.rules.update)
	return this.CrudDomainService.Update(ctx, cmd, opts)
}

func (this *ruledDomainService) Delete(
	ctx corectx.Context, cmd composable.DeleteCommand, options ...composable.DeleteOptions,
) (*composable.MutateResult, error) {
	opts := safe.GetOptional(options, composable.DeleteOptions{})
	opts.ValidateExtra = composable.ChainValidateExtra(opts.ValidateExtra, this.rules.delete)
	return this.CrudDomainService.Delete(ctx, cmd, opts)
}

// Plain resources: the schema states every invariant.

type TaxDomainServiceImpl struct{ composable.CrudDomainService }

func NewTaxDomainService(base composable.CrudDomainService) taxconfig.TaxDomainService {
	return &TaxDomainServiceImpl{CrudDomainService: base}
}

type TaxGroupDomainServiceImpl struct{ composable.CrudDomainService }

func NewTaxGroupDomainService(base composable.CrudDomainService) taxconfig.TaxGroupDomainService {
	return &TaxGroupDomainServiceImpl{CrudDomainService: base}
}

type TaxProductClassificationDomainServiceImpl struct{ composable.CrudDomainService }

func NewTaxProductClassificationDomainService(
	base composable.CrudDomainService,
) taxconfig.TaxProductClassificationDomainService {
	return &TaxProductClassificationDomainServiceImpl{CrudDomainService: base}
}

// Versioned resources: lifecycle, immutability once published, and no overlapping published
// periods.

type TaxDefinitionVersionDomainServiceImpl struct{ ruledDomainService }

func NewTaxDefinitionVersionDomainService(
	base composable.CrudDomainService,
) taxconfig.TaxDefinitionVersionDomainService {
	return &TaxDefinitionVersionDomainServiceImpl{newRuled(base, taxRuleSet{
		create: validateDefinitionVersionCreate(base.Repository()),
		update: validateDefinitionVersionUpdate(base.Repository()),
		delete: validateDefinitionVersionDelete,
	})}
}

type TaxRateVersionDomainServiceImpl struct{ ruledDomainService }

func NewTaxRateVersionDomainService(base composable.CrudDomainService) taxconfig.TaxRateVersionDomainService {
	return &TaxRateVersionDomainServiceImpl{newRuled(base, taxRuleSet{
		create: validateRateVersionCreate(base.Repository()),
		update: validateRateVersionUpdate(base.Repository()),
		delete: validateRateVersionDelete,
	})}
}

type TaxRoundingPolicyDomainServiceImpl struct{ ruledDomainService }

func NewTaxRoundingPolicyDomainService(
	base composable.CrudDomainService,
) taxconfig.TaxRoundingPolicyDomainService {
	return &TaxRoundingPolicyDomainServiceImpl{newRuled(base, taxRuleSet{
		create: validateRoundingPolicyCreate,
		update: validateRoundingPolicyUpdate,
		delete: validateRoundingPolicyDelete,
	})}
}

// TaxComponentDomainServiceImpl reads the definition versions a component hangs off through their
// own repository: a repository answers only for its own schema.
type TaxComponentDomainServiceImpl struct{ ruledDomainService }

func NewTaxComponentDomainService(
	base composable.CrudDomainService, versions taxconfig.TaxDefinitionVersionRepository,
) taxconfig.TaxComponentDomainService {
	repos := componentRepos{Versions: versions, Components: base.Repository()}
	return &TaxComponentDomainServiceImpl{newRuled(base, taxRuleSet{
		create: validateComponentCreate(repos),
		update: validateComponentUpdate(repos),
	})}
}

type TaxJurisdictionDomainServiceImpl struct{ ruledDomainService }

func NewTaxJurisdictionDomainService(base composable.CrudDomainService) taxconfig.TaxJurisdictionDomainService {
	return &TaxJurisdictionDomainServiceImpl{newRuled(base, taxRuleSet{
		create: validateJurisdictionCreate(base.Repository()),
		update: validateJurisdictionUpdate(base.Repository()),
	})}
}
