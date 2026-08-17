package services

import (
	"sort"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// Putaway: deciding where goods that have just arrived somewhere should be put next.
//
// The answer is a suggestion and nothing else. No quant changes, no move is created, nothing is
// reserved — the caller takes the destination and asks the Stock movement engine to do the moving.

// PutawayContext is what is known about the goods that just arrived.
//
// The three product references are optional. A rule that names one only matches when the caller
// supplies it, so a caller who knows nothing beyond the arrival location still gets whatever
// general rule applies.
type PutawayContext struct {
	WarehouseId       string
	ArrivalLocationId string
	ProductId         string
	ProductCategoryId string
	PackageTypeId     string
}

// PutawaySuggestion is the destination and the rule that chose it.
//
// MatchedRuleId is returned alongside the destination so a caller can explain the decision, which
// matters when a warehouse has many overlapping rules and someone asks why goods went where.
type PutawaySuggestion struct {
	DestinationLocationId string
	MatchedRuleId         string
}

// SuggestPutawayLocation returns where arriving goods should be put, or nothing if no rule applies.
//
// Rules are considered in priority order, lowest first, and the first one whose destination is
// actually usable wins. A rule pointing at a suspended or archived location is skipped rather than
// returned, because suggesting somewhere goods may not go is worse than suggesting nowhere.
func SuggestPutawayLocation(
	ctx corectx.Context, putawayCtx PutawayContext,
) (*PutawaySuggestion, error) {
	if putawayCtx.ArrivalLocationId == "" {
		return nil, nil
	}

	rules, err := findCandidatePutawayRules(ctx, putawayCtx)
	if err != nil {
		return nil, err
	}

	sortRulesByPriority(rules)

	for _, rule := range rules {
		if !ruleMatchesContext(rule, putawayCtx) {
			continue
		}

		destinationId := derefId(rule.GetDestinationLocationId())
		usable, err := isLocationUsableForPutaway(ctx, destinationId)
		if err != nil {
			return nil, err
		}
		if usable {
			return &PutawaySuggestion{
				DestinationLocationId: destinationId,
				MatchedRuleId:         derefString(rule.GetId()),
			}, nil
		}
	}
	return nil, nil
}

// sortRulesByPriority orders rules so the lowest priority is considered first.
//
// The sort is stable, so two rules of equal priority resolve in the order the repository returned
// them rather than at random: the same arrival should always produce the same suggestion.
func sortRulesByPriority(rules []models.PutawayRule) {
	sort.SliceStable(rules, func(i, j int) bool {
		return derefInt(rules[i].GetPriority()) < derefInt(rules[j].GetPriority())
	})
}

// findCandidatePutawayRules reads the unarchived rules for a warehouse and arrival location.
//
// Archived rules are excluded here rather than skipped later: an archived rule is out of the
// working set entirely, which is the whole of what archiving means for this resource.
func findCandidatePutawayRules(
	ctx corectx.Context, putawayCtx PutawayContext,
) ([]models.PutawayRule, error) {
	engine, err := engineFor(models.PutawayRuleSchemaName)
	if err != nil {
		return nil, err
	}

	graph := &dmodel.SearchGraph{}
	conditions := []dmodel.SearchNode{
		*dmodel.NewSearchNode().NewCondition(
			models.PutawayRuleFieldSourceLocationId, dmodel.Equals, putawayCtx.ArrivalLocationId),
	}
	if putawayCtx.WarehouseId != "" {
		conditions = append(conditions, *dmodel.NewSearchNode().NewCondition(
			models.PutawayRuleFieldWarehouseId, dmodel.Equals, putawayCtx.WarehouseId))
	}
	graph.And(conditions...)

	rules := make([]models.PutawayRule, 0)
	for page := 0; ; page++ {
		found, err := engine.ResourceRepository().Search(ctx, dyn.RepoSearchParam{
			Graph: graph,
			Page:  page,
			Size:  usageScanPageSize,
		})
		if err != nil {
			return nil, errors.Wrap(err, "findCandidatePutawayRules")
		}
		if found == nil || !found.HasData || len(found.Data.Items) == 0 {
			break
		}
		for _, row := range found.Data.Items {
			rules = append(rules, *models.NewPutawayRuleFrom(row))
		}
		if len(found.Data.Items) < usageScanPageSize {
			break
		}
	}
	return rules, nil
}

// ruleMatchesContext checks the optional product criteria.
//
// A criterion the rule leaves empty matches anything, which is what makes a general rule general.
// A criterion the rule names but the caller did not supply does not match: the rule asked about
// something the caller cannot answer, so it cannot be shown to apply.
func ruleMatchesContext(rule models.PutawayRule, putawayCtx PutawayContext) bool {
	return criterionMatches(derefId(rule.GetProductId()), putawayCtx.ProductId) &&
		criterionMatches(derefId(rule.GetProductCategoryId()), putawayCtx.ProductCategoryId) &&
		criterionMatches(derefId(rule.GetPackageTypeId()), putawayCtx.PackageTypeId)
}

func criterionMatches(ruleValue string, contextValue string) bool {
	if ruleValue == "" {
		return true
	}
	return ruleValue == contextValue
}

// isLocationUsableForPutaway reports whether goods may actually be put in a location.
//
// Suspended is excluded here, which is what suspension is for: the location still exists and still
// holds whatever it held, but nothing new is routed to it.
func isLocationUsableForPutaway(ctx corectx.Context, locationId string) (bool, error) {
	if locationId == "" {
		return false, nil
	}

	engine, err := engineFor(models.InventoryLocationSchemaName)
	if err != nil {
		return false, err
	}
	found, err := engine.ResourceRepository().FindByKeys(ctx, dmodel.DynamicFields{
		models.InventoryLocationFieldId: locationId,
	})
	if err != nil {
		return false, errors.Wrap(err, "isLocationUsableForPutaway")
	}
	if found == nil || !found.HasData {
		return false, nil
	}

	location := models.NewInventoryLocationFrom(found.Data)
	if location.GetIsArchived() != nil && *location.GetIsArchived() {
		return false, nil
	}
	if derefString(location.GetStatus()) != models.InventoryLocationStatusActive {
		return false, nil
	}
	// A purely organisational node holds nothing, so routing goods to it would be routing them
	// nowhere.
	return derefString(location.GetLocationUsage()) == models.InventoryLocationUsageInternal, nil
}
