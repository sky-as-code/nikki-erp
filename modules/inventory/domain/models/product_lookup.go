package models

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// ProductSearcher is declared here rather than imported from dynamicresource so the domain model
// does not depend on the engine that implements it.
type ProductSearcher interface {
	Search(ctx corectx.Context, param dyn.RepoSearchParam) (*dyn.OpResult[dyn.PagedResultData[dmodel.DynamicFields]], error)
}

// searchAll unwraps the three-way search result into a plain slice.
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

// FindVariantsByCombination returns up to limit variants of a template holding the combination
// key. It searches rather than fetching one because callers pass limit 2 to distinguish the record
// being updated from a genuine conflict, even though {product_template_id, combination_key} is
// unique.
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

// FindActiveTemplateVariants returns up to limit non-archived variants. Active means not archived:
// a discontinued but unarchived variant still counts, because it stays deliberately visible.
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

// FindChildCategories returns up to limit direct children. Cycle detection walks upwards from a
// proposed parent; this is the downward counterpart used by delete guards.
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

func FindTemplatesByCategory(
	ctx corectx.Context, repo ProductSearcher, categoryId string, limit int,
) ([]dmodel.DynamicFields, error) {
	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(ProductTemplateFieldCategoryId, dmodel.Equals, categoryId),
	)
	return searchAll(ctx, repo, graph, limit, "FindTemplatesByCategory")
}
