// Package services holds the Accounting business rules that reach for stored configuration.
//
// The pure arithmetic lives one level down in services/tax: given a resolved rate, what does the
// tax come to. This package is what turns rows into those resolved inputs, and it is deliberately
// the only place in the domain that knows a database exists.
package services

import (
	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/modules/accounting/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
)

// engineFor resolves a resource engine from the shared registry.
//
// It is a var rather than a plain function so a test can substitute the registry without building
// one, which is how the resolver is tested without a database.
var engineFor = func(schemaName string) (drif.DynamicResourceEngine, error) {
	engine, ok := dynamicresource.Registry().GetEngine(schemaName)
	if !ok {
		return nil, errors.Errorf("no resource engine for '%s'", schemaName)
	}
	return engine, nil
}

// EngineFor exposes the registry lookup to the sibling packages that need another resource's
// engine — an application service reaching the repository of a resource it does not own.
func EngineFor(schemaName string) (drif.DynamicResourceEngine, error) {
	return engineFor(schemaName)
}

// RepoFor resolves the repository of one schema's engine.
//
// Every schema has its own engine, and an engine's repository answers only for the schema it was
// built for — asking the tax engine for a rounding policy fails with "field is not defined on this
// schema", because it is searching the wrong table. Each lookup therefore has to be handed the
// repository of the resource it reads.
func RepoFor(schemaName string) (models.TaxSearcher, error) {
	engine, err := engineFor(schemaName)
	if err != nil {
		return nil, err
	}
	return engine.ResourceRepository(), nil
}

// TaxRepos is the set of repositories one calculation reads.
//
// Gathered once at the start of a request and passed down, rather than resolved at each call site:
// the registry lookup is cheap but the mistake it invites is not, and a struct with named fields
// makes handing the wrong repository to a lookup a compile error rather than a runtime one.
type TaxRepos struct {
	Tax               models.TaxSearcher
	DefinitionVersion models.TaxSearcher
	RateVersion       models.TaxSearcher
	Component         models.TaxSearcher
	RoundingPolicy    models.TaxSearcher
	Rule              models.TaxSearcher
	RuleCondition     models.TaxSearcher
	RuleResult        models.TaxSearcher
	Mapping           models.TaxSearcher
	MappingLine       models.TaxSearcher
}

// NewTaxRepos resolves every repository a calculation needs.
func NewTaxRepos() (*TaxRepos, error) {
	repos := &TaxRepos{}
	bindings := map[string]*models.TaxSearcher{
		models.TaxSchemaName:                  &repos.Tax,
		models.TaxDefinitionVersionSchemaName: &repos.DefinitionVersion,
		models.TaxRateVersionSchemaName:       &repos.RateVersion,
		models.TaxComponentSchemaName:         &repos.Component,
		models.TaxRoundingPolicySchemaName:    &repos.RoundingPolicy,
		models.TaxRuleSchemaName:              &repos.Rule,
		models.TaxRuleConditionSchemaName:     &repos.RuleCondition,
		models.TaxRuleResultSchemaName:        &repos.RuleResult,
		models.TaxMappingSchemaName:           &repos.Mapping,
		models.TaxMappingLineSchemaName:       &repos.MappingLine,
	}

	for schemaName, target := range bindings {
		repo, err := RepoFor(schemaName)
		if err != nil {
			return nil, err
		}
		*target = repo
	}
	return repos, nil
}
