package services

import (
	"time"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/services/pricing"
)

// Mirrors pricing_ruleset.go, which hands the same rows to a device that prices its own basket:
// both readers must resolve to the same numbers.

const (
	scopeGlobal  int32 = 0
	scopeChannel int32 = 1
	scopePoint   int32 = 2
)

// An item whose list is out of its window is dropped rather than ranked at zero, which would let it
// lose to nothing and so outrank a global list.
func loadEnginePricelistItems(
	ctx corectx.Context, orgId, salesPointId, channelId string, at time.Time,
) ([]pricing.PricelistItem, error) {
	lists, err := searchLiveRows(ctx, models.SalesPricelistSchemaName, orgId, nil)
	if err != nil {
		return nil, err
	}

	type listRank struct{ specificity, priority int32 }
	ranks := make(map[string]listRank, len(lists))
	listIds := make([]string, 0, len(lists))
	for _, list := range lists {
		if !pricelistAppliesAt(list, salesPointId, channelId) {
			continue
		}
		if !withinWindow(list, models.SalesPricelistFieldValidFrom, models.SalesPricelistFieldValidUntil, at) {
			continue
		}
		id := stringOf(list, models.SalesPricelistFieldId)
		ranks[id] = listRank{
			specificity: scopeOfPricelist(list),
			priority:    int32Of(list, models.SalesPricelistFieldPriority),
		}
		listIds = append(listIds, id)
	}
	if len(listIds) == 0 {
		return nil, nil
	}

	rows, err := searchLiveRows(ctx, models.SalesPricelistItemSchemaName, orgId,
		inCondition(models.SalesPricelistItemFieldSalesPricelistId, listIds))
	if err != nil {
		return nil, err
	}

	items := make([]pricing.PricelistItem, 0, len(rows))
	for _, row := range rows {
		rank, found := ranks[stringOf(row, models.SalesPricelistItemFieldSalesPricelistId)]
		if !found {
			continue
		}
		if !withinWindow(row, models.SalesPricelistItemFieldValidFrom, models.SalesPricelistItemFieldValidTo, at) {
			continue
		}

		items = append(items, pricing.PricelistItem{
			Id:                stringOf(row, models.SalesPricelistItemFieldId),
			AppliesTo:         stringOf(row, models.SalesPricelistItemFieldAppliesTo),
			ProductVariantId:  stringOf(row, models.SalesPricelistItemFieldProductVariantId),
			ProductTemplateId: stringOf(row, models.SalesPricelistItemFieldProductTemplateId),
			ProductCategoryId: stringOf(row, models.SalesPricelistItemFieldProductCategoryId),
			UomId:             stringOf(row, models.SalesPricelistItemFieldUomId),
			UnitPrice:         decimalOf(row, models.SalesPricelistItemFieldPrice),
			MinQuantity:       decimalOf(row, models.SalesPricelistItemFieldMinQuantity),
			Specificity:       rank.specificity,
			Priority:          rank.priority,
			Sequence:          int32Of(row, models.SalesPricelistItemFieldSequence),
			CalculationMethod: stringOf(row, models.SalesPricelistItemFieldCalculationMethod),
			DiscountPercent:   decimalOf(row, models.SalesPricelistItemFieldDiscountPercent),
			BasePriceSource:   stringOf(row, models.SalesPricelistItemFieldBasePriceSource),
			// OTHER_PRICELIST needs a chain this read does not walk; unresolved, the line falls back.
			ResolvedBasePrice: decimal.Zero,
			HasResolvedBase:   false,
			SurchargeAmount:   decimalOf(row, models.SalesPricelistItemFieldSurchargeAmount),
			RoundingIncrement: decimalOf(row, models.SalesPricelistItemFieldRoundingIncrement),
		})
	}
	return items, nil
}

func scopeOfPricelist(record dmodel.DynamicFields) int32 {
	if stringOf(record, models.SalesPricelistFieldSalesPointId) != "" {
		return scopePoint
	}
	if stringOf(record, models.SalesPricelistFieldSalesChannelId) != "" {
		return scopeChannel
	}
	return scopeGlobal
}

// [from, until): the upper bound is exclusive, as everywhere else in this module. An absent bound
// is open.
func withinWindow(record dmodel.DynamicFields, fromField, untilField string, at time.Time) bool {
	if from := dateTimeOf(record, fromField); from != nil && from.AfterT(at) {
		return false
	}
	if until := dateTimeOf(record, untilField); until != nil && !until.AfterT(at) {
		return false
	}
	return true
}
