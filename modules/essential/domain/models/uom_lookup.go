package models

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)


// UomSearcher is the slice of a resource repository the UoM lookups need. Declaring it
// here rather than importing the dynamicresource interface keeps the domain model free of
// a dependency on the engine that happens to implement it.
type UomSearcher interface {
	Search(ctx corectx.Context, param dyn.RepoSearchParam) (*dyn.OpResult[dyn.PagedResultData[dmodel.DynamicFields]], error)
}

// FindCategoryReferenceUoms returns up to limit Reference UoMs of the given category.
//
// It searches rather than fetching one: {category_id, uom_type} is not a unique key group,
// and the repository's GetOne rejects any filter that is not one. Callers asking "does this
// category already have a reference?" pass limit 2, so a stored violation of
// BR-UOM-ESS-005 is visible rather than silently truncated to the first row.
func FindCategoryReferenceUoms(
	ctx corectx.Context, repo UomSearcher, categoryId string, limit int,
) ([]dmodel.DynamicFields, error) {
	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(UomFieldCategoryId, dmodel.Equals, categoryId),
		*dmodel.NewSearchNode().NewCondition(UomFieldUomType, dmodel.Equals, UomTypeReference.String()),
	)

	found, err := repo.Search(ctx, dyn.RepoSearchParam{
		Graph: graph,
		Page:  0,
		Size:  limit,
	})
	if err != nil {
		return nil, errors.Wrap(err, "FindCategoryReferenceUoms")
	}
	if found.ClientErrors.Count() > 0 {
		return nil, errors.Wrap(found.ClientErrors.ToError(), "FindCategoryReferenceUoms")
	}
	if !found.HasData {
		return nil, nil
	}
	return found.Data.Items, nil
}
