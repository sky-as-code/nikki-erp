package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"

	itExt "github.com/sky-as-code/nikki-erp/modules/commerce/sales/interfaces/external"
)

// PromotionPorts are what the order-side promotion flow reads from other modules. A package hook,
// like DocumentNumbering, because reprice, confirm and cancel are free functions reached from many
// call chains. A nil port degrades safely: no product facts means no selector matches, no unit
// conversion means only same-unit lines count, no currency lookup means no promotion applies.
type PromotionPorts struct {
	Products   itExt.ProductPromotionExtService
	Uoms       itExt.UomConversionExtService
	Currencies itExt.OrgDefaultsLookupExtService
}

var promotionPorts PromotionPorts

// SetPromotionPorts installs the ports, from the module's OnAppStarted, and wires the rule's
// mixed-unit warning to the product port.
func SetPromotionPorts(ports PromotionPorts) {
	promotionPorts = ports
	if ports.Products != nil {
		SetPromotionUomCategoryResolver(ports.Products.ResolveUomCategories)
	}
}

// searchRows reads every row of a schema matching all the given conditions, without the org filter
// searchLiveRows adds: the promotion junctions carry no org_id. Archived rows are included only when
// asked.
func searchRows(
	ctx corectx.Context, schemaName string, includeArchived bool, nodes ...*dmodel.SearchNode,
) ([]dmodel.DynamicFields, error) {
	repo, err := repoFor(schemaName)
	if err != nil {
		return nil, err
	}
	graph := &dmodel.SearchGraph{}
	conditions := make([]dmodel.SearchNode, 0, len(nodes))
	for _, node := range nodes {
		if node == nil {
			return nil, nil
		}
		conditions = append(conditions, *node)
	}
	// One And call: SearchGraph.And replaces the conditions of an earlier call.
	graph.And(conditions...)
	rows := []dmodel.DynamicFields{}
	for page := 0; page < maxRuleSetPages; page++ {
		found, err := repo.Search(ctx, dyn.RepoSearchParam{
			Graph: graph, Page: page, Size: model.MODEL_RULE_PAGE_MAX_SIZE, IncludeArchived: util.ToPtr(includeArchived),
		})
		if err != nil {
			return nil, errors.Wrapf(err, "reading '%s'", schemaName)
		}
		if found == nil || !found.HasData || len(found.Data.Items) == 0 {
			break
		}
		rows = append(rows, found.Data.Items...)
		if len(found.Data.Items) < model.MODEL_RULE_PAGE_MAX_SIZE {
			break
		}
	}
	return rows, nil
}

func equalsCondition(field string, value any) *dmodel.SearchNode {
	return dmodel.NewSearchNode().NewCondition(field, dmodel.Equals, value)
}

// groupBy buckets rows by one field, keeping each bucket in read order.
func groupBy(rows []dmodel.DynamicFields, field string) map[string][]dmodel.DynamicFields {
	grouped := map[string][]dmodel.DynamicFields{}
	for _, row := range rows {
		key := stringOf(row, field)
		grouped[key] = append(grouped[key], row)
	}
	return grouped
}

func idsOf(rows []dmodel.DynamicFields, field string) []string {
	seen := map[string]bool{}
	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		if id := stringOf(row, field); id != "" && !seen[id] {
			seen[id] = true
			ids = append(ids, id)
		}
	}
	return ids
}
