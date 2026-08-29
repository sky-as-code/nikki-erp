package services

import (
	"sort"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// Putaway decides where goods that have just arrived should be put next. The answer is a suggestion
// only: no quant changes, no move is created, nothing is reserved.

// PutawayContext is what is known about the goods that just arrived. The three product references
// are optional: a rule that names one matches only when the caller supplies it.
type PutawayContext struct {
	WarehouseId       string
	ArrivalLocationId string
	ProductId         string
	ProductCategoryId string
	PackageTypeId     string
}

// PutawaySuggestion is the destination and the rule that chose it. MatchedRuleId lets a caller
// explain the decision when many rules overlap.
type PutawaySuggestion struct {
	DestinationLocationId string
	MatchedRuleId         string
}

// SuggestPutawayLocation returns where arriving goods should be put, or nothing if no rule applies.
// Rules are considered lowest priority first, and the first whose destination is actually usable
// wins; one pointing at a suspended or archived location is skipped rather than returned.
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

// sortRulesByPriority orders rules lowest priority first. The sort must stay stable, so equal
// priorities resolve in repository order and the same arrival always produces the same suggestion.
func sortRulesByPriority(rules []models.PutawayRule) {
	sort.SliceStable(rules, func(i, j int) bool {
		return derefInt(rules[i].GetPriority()) < derefInt(rules[j].GetPriority())
	})
}

// findCandidatePutawayRules reads the unarchived rules for a warehouse and arrival location.
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

// ruleMatchesContext checks the optional product criteria. A criterion the rule leaves empty
// matches anything; one the rule names but the caller did not supply does not match, since the rule
// cannot be shown to apply.
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

// isLocationUsableForPutaway reports whether goods may actually be put in a location. Suspended is
// excluded: the location still holds what it held, but nothing new is routed to it.
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
	// A purely organisational node holds nothing, so routing goods there routes them nowhere.
	return derefString(location.GetLocationUsage()) == models.InventoryLocationUsageInternal, nil
}
