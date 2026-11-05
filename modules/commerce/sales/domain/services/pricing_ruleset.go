package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"

	"github.com/sky-as-code/nikki-erp/modules/commerce/sales/domain/models"
	itCatalog "github.com/sky-as-code/nikki-erp/modules/commerce/sales/interfaces/catalog"
)

// Handing the whole pricing rule set to a caller that prices its own baskets.
//
// The reads run from the pricelists and programs outwards: each level's ids are what the next level
// is keyed by, so a level that comes back empty short-circuits the ones below it rather than
// scanning a table for rows that cannot be reachable.
//
// Every read is scoped to the org. The sales point narrows pricelists further. Promotions do not
// travel: the kiosk receives no promotion data, and Sales prices, confirms and cancels every kiosk
// order itself (PROMO-013). The rule set has no promotion field.

// NewPricingRuleSetExtService builds the port. It holds no state: every read goes through the
// resource hub at call time, like the rest of this package.
func NewPricingRuleSetExtService() itCatalog.PricingRuleSetExtService {
	return &pricingRuleSetExtServiceImpl{}
}

type pricingRuleSetExtServiceImpl struct{}

var _ itCatalog.PricingRuleSetExtService = (*pricingRuleSetExtServiceImpl)(nil)

func (this *pricingRuleSetExtServiceImpl) PricingRuleSetForPoint(
	ctx corectx.Context, query itCatalog.PricingRuleSetQuery,
) (*itCatalog.PricingRuleSetResult, error) {
	if query.OrgId == "" || query.SalesPointId == "" {
		return &itCatalog.PricingRuleSetResult{Data: itCatalog.NewPricingRuleSet()}, nil
	}

	point, err := loadRecord(ctx, models.SalesPointSchemaName,
		models.SalesPointFieldId, query.SalesPointId)
	if err != nil {
		return nil, err
	}
	// A point that is gone answers an empty set rather than an error: the caller is a machine
	// polling its own configuration, and a 500 would take its catalogue down over a row it cannot
	// do anything about.
	if point == nil || stringOf(point, basemodel.FieldOrgId) != query.OrgId {
		return &itCatalog.PricingRuleSetResult{Data: itCatalog.NewPricingRuleSet()}, nil
	}

	ruleSet := itCatalog.NewPricingRuleSet()
	channelId := stringOf(point, models.SalesPointFieldSalesChannelId)

	if err := loadPricelistRules(ctx, query.OrgId, query.SalesPointId, channelId, &ruleSet); err != nil {
		return nil, err
	}
	if err := loadComboRules(ctx, query.OrgId, &ruleSet); err != nil {
		return nil, err
	}

	return &itCatalog.PricingRuleSetResult{Data: ruleSet, HasData: true}, nil
}

// loadPricelistRules reads the lists that could rank here, then their items.
//
// The scope filter runs in Go rather than in the search graph because it tests NULL as a meaningful
// value — an unscoped list applies everywhere — and "column IS NULL OR column = x" is a shape the
// graph's operators do not express as one condition. A pricelist table is configuration and small,
// so reading the org's lists and deciding in memory costs one query either way.
func loadPricelistRules(
	ctx corectx.Context, orgId, salesPointId, channelId string, ruleSet *itCatalog.PricingRuleSet,
) error {
	records, err := searchLiveRows(ctx, models.SalesPricelistSchemaName, orgId, nil)
	if err != nil {
		return err
	}

	pricelistIds := make([]string, 0, len(records))
	for _, record := range records {
		if !pricelistAppliesAt(record, salesPointId, channelId) {
			continue
		}
		ruleSet.Pricelists = append(ruleSet.Pricelists, *models.NewSalesPricelistFrom(record))
		pricelistIds = append(pricelistIds, stringOf(record, models.SalesPricelistFieldId))
	}
	if len(pricelistIds) == 0 {
		return nil
	}

	items, err := searchLiveRows(ctx, models.SalesPricelistItemSchemaName, orgId,
		inCondition(models.SalesPricelistItemFieldSalesPricelistId, pricelistIds))
	if err != nil {
		return err
	}
	for _, item := range items {
		ruleSet.PricelistItems = append(ruleSet.PricelistItems, *models.NewSalesPricelistItemFrom(item))
	}
	return nil
}

// pricelistAppliesAt decides whether one list could ever rank at this point.
//
// Nullability IS the specificity mechanism: a list naming this point applies, a list naming no
// point but this channel applies, a list naming neither applies everywhere. A list naming another
// point or another channel can never win here, and shipping it would invite the caller to rank a
// rule it must never choose. Which of the survivors wins is the CALLER's decision, made from the
// scope columns that travel with them.
func pricelistAppliesAt(record dmodel.DynamicFields, salesPointId, channelId string) bool {
	if pointId := stringOf(record, models.SalesPricelistFieldSalesPointId); pointId != "" {
		return pointId == salesPointId
	}
	if listChannel := stringOf(record, models.SalesPricelistFieldSalesChannelId); listChannel != "" {
		return listChannel == channelId
	}
	return true
}

// loadComboRules reads the org's bundles and the components of the ones that survived.
//
// Combos carry no scope columns: a bundle is offered by the organization, not by a place, so every
// live combo travels. Components are keyed by combo, so an org with no combos costs one query.
func loadComboRules(
	ctx corectx.Context, orgId string, ruleSet *itCatalog.PricingRuleSet,
) error {
	combos, err := searchLiveRows(ctx, models.SalesComboSchemaName, orgId, nil)
	if err != nil {
		return err
	}

	comboIds := make([]string, 0, len(combos))
	for _, combo := range combos {
		ruleSet.Combos = append(ruleSet.Combos, *models.NewSalesComboFrom(combo))
		comboIds = append(comboIds, stringOf(combo, models.SalesComboFieldId))
	}
	if len(comboIds) == 0 {
		return nil
	}

	components, err := searchLiveRows(ctx, models.SalesComboComponentSchemaName, orgId,
		inCondition(models.SalesComboComponentFieldSalesComboId, comboIds))
	if err != nil {
		return err
	}
	for _, component := range components {
		ruleSet.ComboComponents = append(ruleSet.ComboComponents,
			*models.NewSalesComboComponentFrom(component))
	}
	return nil
}

// searchLiveRows reads every non-archived row of a schema in one org, optionally narrowed further.
//
// It pages to exhaustion rather than trusting one read: a busy organization's pricelist items and
// voucher codes both run past a single page, and a truncated rule set is worse than no rule set
// because the caller cannot tell it was truncated.
//
// Archived rows are excluded through IncludeArchived rather than through a condition of our own,
// because the repository prepends that condition ONLY to schemas carrying the column. Five of these
// eleven tables have no is_archived — a condition is withdrawn by archiving the program above it —
// and a hand-written condition would have to carry a list of which, to be wrong the day one gains
// the mixin.
func searchLiveRows(
	ctx corectx.Context, schemaName, orgId string, extra *dmodel.SearchNode,
) ([]dmodel.DynamicFields, error) {
	engineRepo, err := repoFor(schemaName)
	if err != nil {
		return nil, err
	}

	graph := &dmodel.SearchGraph{}
	// One And call: SearchGraph.And replaces the conditions of an earlier call, which silently
	// dropped the org filter whenever an extra condition was given.
	nodes := []dmodel.SearchNode{*dmodel.NewSearchNode().NewCondition(basemodel.FieldOrgId, dmodel.Equals, orgId)}
	if extra != nil {
		nodes = append(nodes, *extra)
	}
	graph.And(nodes...)

	rows := make([]dmodel.DynamicFields, 0, model.MODEL_RULE_PAGE_MAX_SIZE)
	for page := 0; page < maxRuleSetPages; page++ {
		found, err := engineRepo.Search(ctx, dyn.RepoSearchParam{
			Graph:           graph,
			Page:            page,
			Size:            model.MODEL_RULE_PAGE_MAX_SIZE,
			IncludeArchived: util.ToPtr(false),
		})
		if err != nil {
			return nil, errors.Wrapf(err, "reading the pricing rules of '%s'", schemaName)
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

// maxRuleSetPages caps one table's scan. Twenty pages is forty times the largest rule table seen in
// practice; reaching it means a rule table has grown into a transaction log, which is a data problem
// to find rather than a page count to raise.
const maxRuleSetPages = 20

// inCondition builds an IN over a list of ids, or nil for an empty list — which every caller has
// already short-circuited, since an empty IN matches nothing and asking for it is a wasted query.
func inCondition(field string, ids []string) *dmodel.SearchNode {
	if len(ids) == 0 {
		return nil
	}
	values := make([]any, 0, len(ids))
	for _, id := range ids {
		values = append(values, id)
	}
	return dmodel.NewSearchNode().NewCondition(field, dmodel.In, values...)
}
