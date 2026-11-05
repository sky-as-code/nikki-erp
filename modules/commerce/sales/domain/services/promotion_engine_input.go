package services

import (
	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"

	"github.com/sky-as-code/nikki-erp/modules/commerce/sales/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/commerce/sales/domain/services/pricing"
	"github.com/sky-as-code/nikki-erp/modules/commerce/sales/domain/services/pricing/promotion"
	itExt "github.com/sky-as-code/nikki-erp/modules/commerce/sales/interfaces/external"
)

// Building the engine's input from a priced draft. Every line is described as it stands after
// pricelist and manual discounts and before any promotion (BR-C05): N_i is its untaxed value and
// T_i the tax Accounting charged on it (BR-C06), since Sales prices tax-inclusive.

type engineInputParams struct {
	pc           *promotionContext
	order        dmodel.DynamicFields
	pricingInput pricing.Input
	base         pricing.Result
	baseTax      map[string]pricing.LineTax
	scale        int32
	facts        map[string]itExt.PromotionProductFacts
}

func buildEngineInput(ctx corectx.Context, params engineInputParams, programs []*promotionCandidate) promotion.Input {
	input := promotion.Input{
		OrderId:       params.pc.orderId,
		StageVersion:  int32Of(params.order, models.SalesOrderFieldStageVersion),
		CurrencyId:    params.pc.currencyId,
		CurrencyScale: params.scale,
		CustomerId:    params.pc.customerId,
		Selections:    params.pc.selections,
	}
	for _, candidate := range programs {
		input.Programs = append(input.Programs, candidate.program)
	}
	qualificationUoms := qualificationUomsOf(input.Programs)
	templates := templatesByLine(params.pricingInput)
	converter := newUomConverter()
	for _, line := range params.base.Lines {
		input.Lines = append(input.Lines,
			engineLineOf(ctx, line, params, templates[line.Key], qualificationUoms, converter))
	}
	for _, record := range params.pc.lineRecords {
		if stringOf(record, models.SalesOrderLineFieldLineType) == string(models.SalesOrderLineTypePromotionGift) {
			input.Lines = append(input.Lines, giftLineOf(record))
		}
	}
	return input
}

type lineProductFacts struct {
	templateId  string
	categoryIds []string
}

func templatesByLine(input pricing.Input) map[string]lineProductFacts {
	byLine := map[string]lineProductFacts{}
	for _, line := range input.Lines {
		byLine[line.Key] = lineProductFacts{templateId: line.ProductTemplateId, categoryIds: line.CategoryPath}
	}
	return byLine
}

func engineLineOf(
	ctx corectx.Context, line pricing.LineResult, params engineInputParams, basis lineProductFacts,
	qualificationUoms []string, converter *uomConverter,
) promotion.Line {
	tax := params.baseTax[line.Key].Amount
	facts, found := params.facts[line.ProductVariantId]
	if !found {
		facts = itExt.PromotionProductFacts{ProductTemplateId: basis.templateId, CategoryIds: basis.categoryIds}
	}
	if facts.ProductTemplateId == "" {
		facts.ProductTemplateId = basis.templateId
	}
	if len(facts.CategoryIds) == 0 {
		facts.CategoryIds = basis.categoryIds
	}
	engineLine := promotion.Line{
		Key: line.Key, LineNumber: line.LineNumber, LineType: line.LineType,
		ProductVariantId: line.ProductVariantId, ProductTemplateId: facts.ProductTemplateId,
		CategoryIds: facts.CategoryIds, PricelistId: line.PricelistId,
		UomId: line.UomId, Quantity: line.Quantity,
		BaseQuantity:  converter.convertOr(ctx, line.Quantity, line.UomId, facts.BaseUomId, line.Quantity),
		UntaxedAmount: line.NetAmount.Sub(tax),
		TaxAmount:     tax,
	}
	if len(qualificationUoms) > 0 {
		engineLine.QualificationQuantities = map[string]decimal.Decimal{}
		for _, uomId := range qualificationUoms {
			if quantity, ok := converter.convert(ctx, line.Quantity, line.UomId, uomId); ok {
				engineLine.QualificationQuantities[uomId] = quantity
			}
		}
	}
	return engineLine
}

func giftLineOf(record dmodel.DynamicFields) promotion.Line {
	return promotion.Line{
		Key:               stringOf(record, basemodel.FieldId),
		LineNumber:        int32Of(record, models.SalesOrderLineFieldLineNumber),
		LineType:          promotion.LineTypePromotionGift,
		ProductVariantId:  stringOf(record, models.SalesOrderLineFieldProductVariantId),
		UomId:             stringOf(record, models.SalesOrderLineFieldUomId),
		Quantity:          decimalOf(record, models.SalesOrderLineFieldOrderedQuantity),
		SourcePromotionId: stringOf(record, models.SalesOrderLineFieldSourcePromotionId),
		SourceRewardId:    stringOf(record, models.SalesOrderLineFieldSourcePromotionRewardId),
	}
}

func qualificationUomsOf(programs []promotion.Program) []string {
	seen := map[string]bool{}
	uoms := []string{}
	for _, program := range programs {
		if program.ProgramType != promotion.ProgramBuyXGetY {
			continue
		}
		for _, rule := range program.Rules {
			if rule.QualificationUomId != "" && !seen[rule.QualificationUomId] {
				seen[rule.QualificationUomId] = true
				uoms = append(uoms, rule.QualificationUomId)
			}
		}
	}
	return uoms
}

// resolvePromotionFacts reads the product facts of every line's variant and every gift variant in
// one batch. A failed read degrades to the pricing basis already on the line.
func resolvePromotionFacts(
	ctx corectx.Context, lines []pricing.LineResult, programs []*promotionCandidate,
) map[string]itExt.PromotionProductFacts {
	if promotionPorts.Products == nil {
		return map[string]itExt.PromotionProductFacts{}
	}
	ids := []string{}
	for _, line := range lines {
		ids = append(ids, line.ProductVariantId)
	}
	for _, candidate := range programs {
		for _, reward := range candidate.program.Rewards {
			if reward.GiftProductVariantId != "" {
				ids = append(ids, reward.GiftProductVariantId)
			}
		}
	}
	facts, err := promotionPorts.Products.ResolveForPromotion(ctx, uniqueStrings(ids))
	if err != nil || facts == nil {
		return map[string]itExt.PromotionProductFacts{}
	}
	return facts
}

func uniqueStrings(values []string) []string {
	seen := map[string]bool{}
	unique := make([]string, 0, len(values))
	for _, value := range values {
		if value != "" && !seen[value] {
			seen[value] = true
			unique = append(unique, value)
		}
	}
	return unique
}

// uomConverter converts through Essential, remembering each answer: the same pair recurs on every
// line of a basket.
type uomConverter struct {
	cache map[[3]string]*decimal.Decimal
}

func newUomConverter() *uomConverter {
	return &uomConverter{cache: map[[3]string]*decimal.Decimal{}}
}

func (this *uomConverter) convertOr(
	ctx corectx.Context, quantity decimal.Decimal, from, to string, fallback decimal.Decimal,
) decimal.Decimal {
	if converted, ok := this.convert(ctx, quantity, from, to); ok {
		return converted
	}
	return fallback
}

// convert answers false when the units are in different categories or cannot be read: such a line
// does not count toward a quantity in that unit (BR-C07).
func (this *uomConverter) convert(
	ctx corectx.Context, quantity decimal.Decimal, from, to string,
) (decimal.Decimal, bool) {
	if from == to && from != "" {
		return quantity, true
	}
	if from == "" || to == "" || promotionPorts.Uoms == nil {
		return decimal.Zero, false
	}
	key := [3]string{quantity.String(), from, to}
	if cached, seen := this.cache[key]; seen {
		return decimalOrZero(cached), cached != nil
	}
	result, err := promotionPorts.Uoms.Convert(ctx, itExt.ConvertQuantityQuery{
		Quantity: quantity, SourceUomId: model.Id(from), TargetUomId: model.Id(to),
	})
	if err != nil || result == nil || !result.HasData || result.ClientErrors.Count() > 0 {
		this.cache[key] = nil
		return decimal.Zero, false
	}
	converted := result.Data.ExactQuantity
	this.cache[key] = &converted
	return converted, true
}
