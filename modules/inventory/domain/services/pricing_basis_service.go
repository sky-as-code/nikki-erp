package services

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

var _ itProduct.ProductPricingBasisService = (*ProductVariantDomainServiceImpl)(nil)

// maxCategoryDepth bounds the ancestor walk, so a cycle in parent_category_id — which the category
// engine refuses to create but a direct database edit could produce — cannot turn one price
// resolution into an infinite loop.
const maxCategoryDepth = 32

// GetPricingBasis reads the pricing inputs for a batch of variants in three reads regardless of
// batch size: the variants, their categories, then each ancestor level. Resolving one variant at a
// time would cost a round trip per line, and an order is repriced on every edit.
func (this *ProductVariantDomainServiceImpl) GetPricingBasis(
	ctx corectx.Context, query itProduct.GetPricingBasisQuery,
) (*itProduct.GetPricingBasisResult, error) {
	bases := map[string]itProduct.PricingBasis{}
	if len(query.ProductVariantIds) == 0 {
		return &itProduct.GetPricingBasisResult{
			Data: itProduct.GetPricingBasisResultData{Bases: bases}, HasData: true}, nil
	}

	variants, err := this.readPricingVariants(ctx, query.ProductVariantIds)
	if err != nil {
		return nil, err
	}

	// Ancestry is resolved once per distinct category: lines of one order routinely share a category,
	// and walking it per line would repeat the same reads.
	paths, err := this.resolveCategoryPaths(ctx, distinctCategoryIds(variants))
	if err != nil {
		return nil, err
	}

	for _, variant := range variants {
		variantId := basisString(variant, models.ProductVariantFieldId)
		if variantId == "" {
			continue
		}
		categoryId := basisString(variant, models.ProductVariantFieldTemplateCategoryId)
		cost, hasCost := basisDecimalString(variant, models.ProductVariantFieldCost)

		bases[variantId] = itProduct.PricingBasis{
			ProductVariantId:  variantId,
			ProductTemplateId: basisString(variant, models.ProductVariantFieldProductTemplateId),
			CategoryPath:      paths[categoryId],
			// The EFFECTIVE base price, not the template's raw one: it already includes what this
			// variant's attribute values add. Being a computed field, it is projected only because it
			// is asked for by name.
			EffectiveBaseSalesPrice: basisString(variant,
				models.ProductVariantFieldEffectiveBaseSalesPrice),
			Cost:    cost,
			HasCost: hasCost,
		}
	}

	return &itProduct.GetPricingBasisResult{
		Data:    itProduct.GetPricingBasisResultData{Bases: bases},
		HasData: true,
	}, nil
}

// readPricingVariants fetches only the fields pricing needs. Naming them explicitly is required,
// not an optimisation: effective_base_sales_price is a COMPUTED field, projected only when asked
// for by name, and reading the whole record leaves it empty so the price silently falls back to the
// catalogue.
func (this *ProductVariantDomainServiceImpl) readPricingVariants(
	ctx corectx.Context, variantIds []string,
) ([]dmodel.DynamicFields, error) {
	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(models.ProductVariantFieldId, dmodel.In, variantIds),
	)

	page, err := searchRows(ctx, models.ProductVariantSchemaName, dyn.SearchQuery{
		Graph: graph,
		Page:  0,
		Size:  len(variantIds),
		Fields: []string{
			models.ProductVariantFieldId,
			models.ProductVariantFieldProductTemplateId,
			models.ProductVariantFieldTemplateCategoryId,
			models.ProductVariantFieldEffectiveBaseSalesPrice,
			models.ProductVariantFieldCost,
		},
	}, "GetPricingBasis")
	if err != nil {
		return nil, err
	}
	if page == nil {
		return nil, nil
	}
	return page.Items, nil
}

// resolveCategoryPaths walks each starting category out to its root, level by level across the
// whole batch: siblings share ancestors, so ten variants in one department cost a handful of reads
// instead of ten walks.
func (this *ProductVariantDomainServiceImpl) resolveCategoryPaths(
	ctx corectx.Context, startIds []string,
) (map[string][]string, error) {
	paths := map[string][]string{}
	if len(startIds) == 0 {
		return paths, nil
	}

	// parents is filled as levels are read, so an ancestor already seen is never read twice.
	parents := map[string]string{}
	frontier := startIds

	for depth := 0; depth < maxCategoryDepth && len(frontier) > 0; depth++ {
		known, err := this.readCategoryParents(ctx, frontier)
		if err != nil {
			return nil, err
		}
		next := make([]string, 0, len(known))
		for categoryId, parentId := range known {
			parents[categoryId] = parentId
			if parentId != "" {
				if _, seen := parents[parentId]; !seen {
					next = append(next, parentId)
				}
			}
		}
		frontier = next
	}

	for _, startId := range startIds {
		paths[startId] = walkUp(startId, parents)
	}
	return paths, nil
}

func (this *ProductVariantDomainServiceImpl) readCategoryParents(
	ctx corectx.Context, categoryIds []string,
) (map[string]string, error) {
	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(models.ProductCategoryFieldId, dmodel.In, categoryIds),
	)

	page, err := searchRows(ctx, models.ProductCategorySchemaName, dyn.SearchQuery{
		Graph:  graph,
		Page:   0,
		Size:   len(categoryIds),
		Fields: []string{models.ProductCategoryFieldId, models.ProductCategoryFieldParentCategoryId},
	}, "readCategoryParents")
	if err != nil {
		return nil, err
	}

	parents := map[string]string{}
	if page == nil {
		return parents, nil
	}
	for _, record := range page.Items {
		parents[basisString(record, models.ProductCategoryFieldId)] =
			basisString(record, models.ProductCategoryFieldParentCategoryId)
	}
	return parents, nil
}

// walkUp turns the parent map into a path, nearest first. The visited set is load-bearing: a cycle
// in parent_category_id would otherwise loop forever, and this code must not assume the engine
// prevented every one.
func walkUp(startId string, parents map[string]string) []string {
	path := make([]string, 0, 4)
	visited := map[string]bool{}
	current := startId

	for current != "" && !visited[current] && len(path) < maxCategoryDepth {
		visited[current] = true
		path = append(path, current)
		current = parents[current]
	}
	return path
}

func distinctCategoryIds(variants []dmodel.DynamicFields) []string {
	seen := map[string]bool{}
	ids := make([]string, 0, len(variants))
	for _, variant := range variants {
		categoryId := basisString(variant, models.ProductVariantFieldTemplateCategoryId)
		if categoryId == "" || seen[categoryId] {
			continue
		}
		seen[categoryId] = true
		ids = append(ids, categoryId)
	}
	return ids
}

// basisString reads one field as a string, tolerating whatever shape it arrives in. Never a bare
// type assertion: a decimal column comes back as a decimal and an absent one as nil, so asserting
// would panic the request on data that is merely unset.
func basisString(record dmodel.DynamicFields, field string) string {
	value, present := record[field]
	if !present || value == nil {
		return ""
	}
	if text, ok := value.(string); ok {
		return text
	}
	if stringer, ok := value.(interface{ String() string }); ok {
		return stringer.String()
	}
	return ""
}

// basisDecimalString reads a decimal field, reporting whether it was set at all. An unset cost and
// a cost of zero must not be confused: zero is a real answer for a giveaway, while unset means a
// COST-based formula rule has nothing to compute from and must decline rather than price at
// nothing.
func basisDecimalString(record dmodel.DynamicFields, field string) (string, bool) {
	value, present := record[field]
	if !present || value == nil {
		return "", false
	}
	text := basisString(record, field)
	return text, text != ""
}
