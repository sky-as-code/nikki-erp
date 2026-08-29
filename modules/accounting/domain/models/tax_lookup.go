package models

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

// TaxSearcher is the slice of a resource repository the tax lookups need. It is declared here rather
// than imported so the domain models do not depend on the engine that implements it.
type TaxSearcher interface {
	Search(ctx corectx.Context, param dyn.RepoSearchParam) (*dyn.OpResult[dyn.PagedResultData[dmodel.DynamicFields]], error)
}

func searchAll(
	ctx corectx.Context, repo TaxSearcher, graph *dmodel.SearchGraph, limit int, what string,
) ([]dmodel.DynamicFields, error) {
	found, err := repo.Search(ctx, dyn.RepoSearchParam{Graph: graph, Page: 0, Size: limit})
	if err != nil {
		return nil, errors.Wrap(err, what)
	}
	if found.ClientErrors.Count() > 0 {
		return nil, errors.Wrap(found.ClientErrors.ToError(), what)
	}
	if !found.HasData {
		return nil, nil
	}
	return found.Data.Items, nil
}

// FindJurisdictionById fetches one jurisdiction, or nil when it does not exist. The cycle check
// calls it per row because a single query cannot follow parent_id to an unknown depth.
func FindJurisdictionById(
	ctx corectx.Context, repo TaxSearcher, id string,
) (*TaxJurisdiction, error) {
	graph := &dmodel.SearchGraph{}
	graph.And(*dmodel.NewSearchNode().NewCondition(TaxJurisdictionFieldId, dmodel.Equals, id))

	items, err := searchAll(ctx, repo, graph, 1, "FindJurisdictionById")
	if err != nil || len(items) == 0 {
		return nil, err
	}
	return NewTaxJurisdictionFrom(items[0]), nil
}

// FindComponentsOfVersion returns the components declared on one tax definition version.
func FindComponentsOfVersion(
	ctx corectx.Context, repo TaxSearcher, parentVersionId string, limit int,
) ([]dmodel.DynamicFields, error) {
	graph := &dmodel.SearchGraph{}
	graph.And(*dmodel.NewSearchNode().NewCondition(
		TaxComponentFieldParentTaxDefinitionVersionId, dmodel.Equals, parentVersionId))

	return searchAll(ctx, repo, graph, limit, "FindComponentsOfVersion")
}

// FindPublishedVersionsOfTax returns the published definition versions of one tax. The caller
// filters the effective period in memory, because the nullable upper bound is more fragile to
// express in the search graph than to compare directly.
func FindPublishedVersionsOfTax(
	ctx corectx.Context, repo TaxSearcher, taxId string, limit int,
) ([]dmodel.DynamicFields, error) {
	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(TaxDefinitionVersionFieldTaxId, dmodel.Equals, taxId),
		*dmodel.NewSearchNode().NewCondition(
			TaxDefinitionVersionFieldLifecycleStatus, dmodel.Equals, string(LifecyclePublished)),
	)

	return searchAll(ctx, repo, graph, limit, "FindPublishedVersionsOfTax")
}

// FindPublishedRateVersionsOfTax returns the published rate versions of one tax.
func FindPublishedRateVersionsOfTax(
	ctx corectx.Context, repo TaxSearcher, taxId string, limit int,
) ([]dmodel.DynamicFields, error) {
	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(TaxRateVersionFieldTaxId, dmodel.Equals, taxId),
		*dmodel.NewSearchNode().NewCondition(
			TaxRateVersionFieldLifecycleStatus, dmodel.Equals, string(LifecyclePublished)),
	)

	return searchAll(ctx, repo, graph, limit, "FindPublishedRateVersionsOfTax")
}

// FindRateVersionsOfTax returns every rate version of a tax regardless of lifecycle state, because
// on a group or none-typed definition even a draft rate is a configuration error.
func FindRateVersionsOfTax(
	ctx corectx.Context, repo TaxSearcher, taxId string, limit int,
) ([]dmodel.DynamicFields, error) {
	graph := &dmodel.SearchGraph{}
	graph.And(*dmodel.NewSearchNode().NewCondition(TaxRateVersionFieldTaxId, dmodel.Equals, taxId))

	return searchAll(ctx, repo, graph, limit, "FindRateVersionsOfTax")
}

// FindEffectiveDefinitionVersion returns the published definition version of a tax whose effective
// period contains taxDate, and reports how many matched. The count matters as much as the row: zero
// means no definition on that date, and more than one is corrupt configuration the caller must
// surface as unresolved rather than settle by taking the newest.
func FindEffectiveDefinitionVersion(
	ctx corectx.Context, repo TaxSearcher, taxId string, taxDate string,
) (*TaxDefinitionVersion, int, error) {
	items, err := FindPublishedVersionsOfTax(ctx, repo, taxId, 100)
	if err != nil {
		return nil, 0, err
	}

	var match *TaxDefinitionVersion
	count := 0
	for _, item := range items {
		version := NewTaxDefinitionVersionFrom(item)
		if !periodContains(version.GetEffectiveFrom(), version.GetEffectiveTo(), taxDate) {
			continue
		}
		count++
		if match == nil {
			match = version
		}
	}
	return match, count, nil
}

// FindEffectiveRateVersion is the rate-version counterpart of FindEffectiveDefinitionVersion.
func FindEffectiveRateVersion(
	ctx corectx.Context, repo TaxSearcher, taxId string, taxDate string,
) (*TaxRateVersion, int, error) {
	items, err := FindPublishedRateVersionsOfTax(ctx, repo, taxId, 100)
	if err != nil {
		return nil, 0, err
	}

	var match *TaxRateVersion
	count := 0
	for _, item := range items {
		version := NewTaxRateVersionFrom(item)
		if !periodContains(version.GetEffectiveFrom(), version.GetEffectiveTo(), taxDate) {
			continue
		}
		count++
		if match == nil {
			match = version
		}
	}
	return match, count, nil
}

// FindDefinitionVersionById fetches one definition version, or nil when it does not exist.
func FindDefinitionVersionById(
	ctx corectx.Context, repo TaxSearcher, id string,
) (*TaxDefinitionVersion, error) {
	graph := &dmodel.SearchGraph{}
	graph.And(*dmodel.NewSearchNode().NewCondition(TaxDefinitionVersionFieldId, dmodel.Equals, id))

	items, err := searchAll(ctx, repo, graph, 1, "FindDefinitionVersionById")
	if err != nil || len(items) == 0 {
		return nil, err
	}
	return NewTaxDefinitionVersionFrom(items[0]), nil
}

// FindVersionsOfTaxAnyStatus returns every definition version of a tax, whatever its lifecycle. The
// cycle check uses this rather than the published-only lookup because a cycle assembled across draft
// versions is still a cycle.
func FindVersionsOfTaxAnyStatus(
	ctx corectx.Context, repo TaxSearcher, taxId string, limit int,
) ([]dmodel.DynamicFields, error) {
	graph := &dmodel.SearchGraph{}
	graph.And(*dmodel.NewSearchNode().NewCondition(TaxDefinitionVersionFieldTaxId, dmodel.Equals, taxId))

	return searchAll(ctx, repo, graph, limit, "FindVersionsOfTaxAnyStatus")
}

// FindMappingById fetches one tax mapping, or nil when it does not exist.
func FindMappingById(
	ctx corectx.Context, repo TaxSearcher, id string,
) (*TaxMapping, error) {
	graph := &dmodel.SearchGraph{}
	graph.And(*dmodel.NewSearchNode().NewCondition(TaxMappingFieldId, dmodel.Equals, id))

	items, err := searchAll(ctx, repo, graph, 1, "FindMappingById")
	if err != nil || len(items) == 0 {
		return nil, err
	}
	return NewTaxMappingFrom(items[0]), nil
}

// FindTaxById fetches one tax's business identity, or nil when it does not exist. The identity
// carries no rate and no effective period; those live on the versions.
func FindTaxById(
	ctx corectx.Context, repo TaxSearcher, id string,
) (*Tax, error) {
	graph := &dmodel.SearchGraph{}
	graph.And(*dmodel.NewSearchNode().NewCondition(TaxFieldId, dmodel.Equals, id))

	items, err := searchAll(ctx, repo, graph, 1, "FindTaxById")
	if err != nil || len(items) == 0 {
		return nil, err
	}
	return NewTaxFrom(items[0]), nil
}

// FindPublishedRules returns every published tax rule, not only those matching a context: whether a
// rule applies is decided by evaluating its conditions against the request and cannot be expressed
// as a query. The caller filters the effective period, as with the version lookups.
func FindPublishedRules(
	ctx corectx.Context, repo TaxSearcher, limit int,
) ([]dmodel.DynamicFields, error) {
	graph := &dmodel.SearchGraph{}
	graph.And(*dmodel.NewSearchNode().NewCondition(
		TaxRuleFieldLifecycleStatus, dmodel.Equals, string(LifecyclePublished)))

	return searchAll(ctx, repo, graph, limit, "FindPublishedRules")
}

// FindConditionsOfRule returns the conditions attached to one rule.
func FindConditionsOfRule(
	ctx corectx.Context, repo TaxSearcher, ruleId string, limit int,
) ([]dmodel.DynamicFields, error) {
	graph := &dmodel.SearchGraph{}
	graph.And(*dmodel.NewSearchNode().NewCondition(
		TaxRuleConditionFieldTaxRuleId, dmodel.Equals, ruleId))

	return searchAll(ctx, repo, graph, limit, "FindConditionsOfRule")
}

// FindResultsOfRule returns the results a matching rule applies.
func FindResultsOfRule(
	ctx corectx.Context, repo TaxSearcher, ruleId string, limit int,
) ([]dmodel.DynamicFields, error) {
	graph := &dmodel.SearchGraph{}
	graph.And(*dmodel.NewSearchNode().NewCondition(
		TaxRuleResultFieldTaxRuleId, dmodel.Equals, ruleId))

	return searchAll(ctx, repo, graph, limit, "FindResultsOfRule")
}

// FindLinesOfMapping returns the substitution lines of one mapping.
func FindLinesOfMapping(
	ctx corectx.Context, repo TaxSearcher, mappingId string, limit int,
) ([]dmodel.DynamicFields, error) {
	graph := &dmodel.SearchGraph{}
	graph.And(*dmodel.NewSearchNode().NewCondition(
		TaxMappingLineFieldTaxMappingId, dmodel.Equals, mappingId))

	return searchAll(ctx, repo, graph, limit, "FindLinesOfMapping")
}

// FindRoundingPolicyByCode returns the published rounding policies with a code. A policy is
// versioned, so several rows may share a code and differ by effective period; the caller picks the
// one in force on the date.
func FindRoundingPolicyByCode(
	ctx corectx.Context, repo TaxSearcher, code string, limit int,
) ([]dmodel.DynamicFields, error) {
	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(TaxRoundingPolicyFieldCode, dmodel.Equals, code),
		*dmodel.NewSearchNode().NewCondition(
			TaxRoundingPolicyFieldLifecycleStatus, dmodel.Equals, string(LifecyclePublished)),
	)

	return searchAll(ctx, repo, graph, limit, "FindRoundingPolicyByCode")
}

// FindEffectiveRoundingPolicy returns the published policy with a code whose period contains the
// date, and how many matched. As with the version lookups, more than one is a configuration fault
// the caller must surface rather than resolve by recency.
func FindEffectiveRoundingPolicy(
	ctx corectx.Context, repo TaxSearcher, code string, taxDate string,
) (*TaxRoundingPolicy, int, error) {
	items, err := FindRoundingPolicyByCode(ctx, repo, code, 100)
	if err != nil {
		return nil, 0, err
	}

	var match *TaxRoundingPolicy
	count := 0
	for _, item := range items {
		policy := NewTaxRoundingPolicyFrom(item)
		if !periodContains(policy.GetEffectiveFrom(), policy.GetEffectiveTo(), taxDate) {
			continue
		}
		count++
		if match == nil {
			match = policy
		}
	}
	return match, count, nil
}
