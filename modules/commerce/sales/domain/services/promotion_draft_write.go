package services

import (
	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"

	"github.com/sky-as-code/nikki-erp/modules/commerce/sales/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/commerce/sales/domain/services/pricing/promotion"
)

// Writing a draft's promotion state (D3): the gift lines and the calculation reference. Applications
// and allocations are written only on confirm. Runs inside the reprice transaction.

func writePromotionDraft(ctx corectx.Context, orderId string, quote *PromotionQuoteResult) error {
	if quote == nil {
		return nil
	}
	if err := removeGiftLines(ctx, quote.outcome.RemovedGiftLineKeys); err != nil {
		return err
	}
	if err := upsertGiftLines(ctx, orderId, quote); err != nil {
		return err
	}
	repo, err := repoFor(models.SalesOrderSchemaName)
	if err != nil {
		return err
	}
	_, err = repo.Update(ctx, dmodel.DynamicFields{
		models.SalesOrderFieldId:                     orderId,
		models.SalesOrderFieldPromotionCalcReference: quote.Reference,
	})
	return err
}

// removeGiftLines drops the gift lines no applied promotion wants any more (BR-R09).
func removeGiftLines(ctx corectx.Context, lineKeys []string) error {
	if len(lineKeys) == 0 {
		return nil
	}
	repo, err := repoFor(models.SalesOrderLineSchemaName)
	if err != nil {
		return err
	}
	for _, key := range lineKeys {
		if _, err := repo.DeleteOne(ctx, dmodel.DynamicFields{models.SalesOrderLineFieldId: key}); err != nil {
			return err
		}
	}
	return nil
}

// upsertGiftLines updates the gift line a promotion already has, or adds one: a reprice never adds
// a second line for the same gift (BR-R09). The quantity is the server's (BR-T01).
func upsertGiftLines(ctx corectx.Context, orderId string, quote *PromotionQuoteResult) error {
	if len(quote.outcome.Gifts) == 0 {
		return nil
	}
	repo, err := repoFor(models.SalesOrderLineSchemaName)
	if err != nil {
		return err
	}
	nextNumber := nextOrderLineNumber(quote.pc.lineRecords)
	for _, gift := range quote.outcome.Gifts {
		fields := giftLineFields(gift, quote)
		if gift.ExistingLineKey != "" {
			fields[models.SalesOrderLineFieldId] = gift.ExistingLineKey
			if _, err := repo.Update(ctx, fields); err != nil {
				return err
			}
			quote.giftLineIds[gift.PromotionId] = gift.ExistingLineKey
			continue
		}
		id, err := model.NewId()
		if err != nil {
			return err
		}
		addNewGiftLineFields(fields, string(*id), orderId, quote.pc.orgId, nextNumber)
		nextNumber++
		if _, err := repo.Insert(ctx, fields); err != nil {
			return err
		}
		quote.giftLineIds[gift.PromotionId] = string(*id)
	}
	return nil
}

func giftLineFields(gift promotion.GiftIntent, quote *PromotionQuoteResult) dmodel.DynamicFields {
	fields := dmodel.DynamicFields{
		models.SalesOrderLineFieldProductVariantId:        gift.ProductVariantId,
		models.SalesOrderLineFieldUomId:                   gift.UomId,
		models.SalesOrderLineFieldOrderedQuantity:         gift.Quantity,
		models.SalesOrderLineFieldSourcePromotionId:       gift.PromotionId,
		models.SalesOrderLineFieldSourcePromotionRewardId: gift.RewardId,
	}
	if facts, found := quote.facts[gift.ProductVariantId]; found {
		if facts.Code != "" {
			fields[models.SalesOrderLineFieldProductCodeSnapshot] = facts.Code
		}
		if facts.Name != "" {
			fields[models.SalesOrderLineFieldProductNameSnapshot] = facts.Name
		}
	}
	return fields
}

// addNewGiftLineFields completes a new gift line: a real line at price zero that Inventory must
// fulfil, never a rewrite of the line the customer bought (BR-R07).
func addNewGiftLineFields(fields dmodel.DynamicFields, lineId, orderId, orgId string, lineNumber int32) {
	fields[models.SalesOrderLineFieldId] = lineId
	fields[models.SalesOrderLineFieldSalesOrderId] = orderId
	fields[models.SalesOrderLineFieldLineNumber] = lineNumber
	fields[models.SalesOrderLineFieldLineType] = string(models.SalesOrderLineTypePromotionGift)
	fields[models.SalesOrderLineFieldRequiresFulfillment] = true
	for _, field := range []string{
		models.SalesOrderLineFieldFulfilledQuantity, models.SalesOrderLineFieldReturnedQuantity,
		models.SalesOrderLineFieldBaseUnitPrice, models.SalesOrderLineFieldEffectiveUnitPrice,
		models.SalesOrderLineFieldGrossAmount, models.SalesOrderLineFieldDiscountAmount,
		models.SalesOrderLineFieldNetAmount, models.SalesOrderLineFieldTaxRateSnapshot,
		models.SalesOrderLineFieldTaxAmount, models.SalesOrderLineFieldFinalAmount,
	} {
		fields[field] = decimal.Zero
	}
	fields[models.SalesOrderLineFieldPricingSource] = promotionGiftPricingSource
	fields[basemodel.FieldOrgId] = orgId
	fields[basemodel.FieldIsArchived] = false
}

// promotionGiftPricingSource is the pricing_source enum value of a gift line.
const promotionGiftPricingSource = "promotion_reward"

func nextOrderLineNumber(lines []dmodel.DynamicFields) int32 {
	highest := int32(0)
	for _, line := range lines {
		if number := int32Of(line, models.SalesOrderLineFieldLineNumber); number > highest {
			highest = number
		}
	}
	return highest + 1
}
