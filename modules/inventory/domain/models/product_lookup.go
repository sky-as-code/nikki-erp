package models

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// ProductSearcher is the slice of a resource repository the Products lookups need. Declaring it
// here rather than importing the dynamicresource interface keeps the domain model free of a
// dependency on the engine that happens to implement it.
type ProductSearcher interface {
	Search(ctx corectx.Context, param dyn.RepoSearchParam) (*dyn.OpResult[dyn.PagedResultData[dmodel.DynamicFields]], error)
}

// searchAll runs a graph search and unwraps the usual three-way result into a plain slice.
func searchAll(
	ctx corectx.Context, repo ProductSearcher, graph *dmodel.SearchGraph, limit int, what string,
) ([]dmodel.DynamicFields, error) {
	found, err := repo.Search(ctx, dyn.RepoSearchParam{
		Graph: graph,
		Page:  0,
		Size:  limit,
	})
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

// FindVariantsByCombination returns up to limit variants of a template holding the given
// combination key.
//
// It searches rather than fetching one: although {product_template_id, combination_key} is a
// unique group, callers ask "does another variant already hold this?" and pass limit 2, so the
// record being updated can be told apart from a genuine conflict. See BR-PROD-VAR-002.
func FindVariantsByCombination(
	ctx corectx.Context, repo ProductSearcher, templateId string, combinationKey string, limit int,
) ([]dmodel.DynamicFields, error) {
	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(ProductVariantFieldProductTemplateId, dmodel.Equals, templateId),
		*dmodel.NewSearchNode().NewCondition(ProductVariantFieldCombinationKey, dmodel.Equals, combinationKey),
	)
	return searchAll(ctx, repo, graph, limit, "FindVariantsByCombination")
}

// FindTemplateVariants returns up to limit variants of a template, archived ones included.
func FindTemplateVariants(
	ctx corectx.Context, repo ProductSearcher, templateId string, limit int,
) ([]dmodel.DynamicFields, error) {
	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(ProductVariantFieldProductTemplateId, dmodel.Equals, templateId),
	)
	return searchAll(ctx, repo, graph, limit, "FindTemplateVariants")
}

// FindActiveTemplateVariants returns up to limit non-archived variants of a template.
//
// "Active" here means not archived, which is what decides whether a template still has a
// selectable product. A discontinued-but-unarchived variant still counts, because the business
// deliberately keeps it visible. See BR-PROD-VAR-006.
func FindActiveTemplateVariants(
	ctx corectx.Context, repo ProductSearcher, templateId string, limit int,
) ([]dmodel.DynamicFields, error) {
	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(ProductVariantFieldProductTemplateId, dmodel.Equals, templateId),
		*dmodel.NewSearchNode().NewCondition(basemodel.FieldIsArchived, dmodel.Equals, false),
	)
	return searchAll(ctx, repo, graph, limit, "FindActiveTemplateVariants")
}

// FindTemplateAttributes returns the attribute configuration rows of a template.
func FindTemplateAttributes(
	ctx corectx.Context, repo ProductSearcher, templateId string, limit int,
) ([]dmodel.DynamicFields, error) {
	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(
			ProductTemplateAttributeFieldProductTemplateId, dmodel.Equals, templateId),
	)
	return searchAll(ctx, repo, graph, limit, "FindTemplateAttributes")
}

// FindTemplateAttributeValues returns the allowed values of one template attribute.
func FindTemplateAttributeValues(
	ctx corectx.Context, repo ProductSearcher, templateAttributeId string, limit int,
) ([]dmodel.DynamicFields, error) {
	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(
			ProductTemplateAttributeValueFieldTemplateAttributeId, dmodel.Equals, templateAttributeId),
	)
	return searchAll(ctx, repo, graph, limit, "FindTemplateAttributeValues")
}

// FindChildCategories returns up to limit direct children of a category. Cycle detection walks
// upwards from a proposed parent, so this is its downward counterpart for delete guards.
func FindChildCategories(
	ctx corectx.Context, repo ProductSearcher, categoryId string, limit int,
) ([]dmodel.DynamicFields, error) {
	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(
			ProductCategoryFieldParentCategoryId, dmodel.Equals, categoryId),
	)
	return searchAll(ctx, repo, graph, limit, "FindChildCategories")
}

// FindTemplatesByCategory returns up to limit templates classified under a category.
func FindTemplatesByCategory(
	ctx corectx.Context, repo ProductSearcher, categoryId string, limit int,
) ([]dmodel.DynamicFields, error) {
	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(ProductTemplateFieldCategoryId, dmodel.Equals, categoryId),
	)
	return searchAll(ctx, repo, graph, limit, "FindTemplatesByCategory")
}
