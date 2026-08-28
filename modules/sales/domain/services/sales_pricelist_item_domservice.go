package services

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// SalesPricelistItemDomainServiceImpl enforces the rules a pricing rule carries that the schema
// cannot express.
//
// All three are conditional on another field, which is exactly what `required_for_create` cannot
// say: which target column must be set depends on applies_to, which price fields are required
// depends on calculation_method, and whether a base pricelist is acceptable depends on the whole
// derivation graph. The schema declares the columns; this decides when each one means anything.
type SalesPricelistItemDomainServiceImpl struct {
	drif.DynamicResourceService
}

func NewSalesPricelistItemDomainService(
	base drif.DynamicResourceService,
) *SalesPricelistItemDomainServiceImpl {
	return &SalesPricelistItemDomainServiceImpl{DynamicResourceService: base}
}

func (this *SalesPricelistItemDomainServiceImpl) Create(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[dmodel.DynamicFields], error) {
	if vErrs := assertRuleConsistent(params); vErrs != nil {
		return &dyn.OpResult[dmodel.DynamicFields]{ClientErrors: *vErrs}, nil
	}
	if vErrs, err := this.assertNoCycle(ctx, params, params); err != nil || vErrs != nil {
		if err != nil {
			return nil, err
		}
		return &dyn.OpResult[dmodel.DynamicFields]{ClientErrors: *vErrs}, nil
	}
	return this.DynamicResourceService.Create(ctx, params)
}

// Update validates the record as it will BE, not as it was sent.
//
// A partial update names only the fields it changes, so validating the payload alone would reject
// a rule for lacking a target it already has. Merging the stored record with the change is what
// makes "exactly one target" mean the same thing on create and on update.
func (this *SalesPricelistItemDomainServiceImpl) Update(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	itemId := stringOf(params, models.SalesPricelistItemFieldId)
	stored, err := loadRecord(ctx, models.SalesPricelistItemSchemaName,
		models.SalesPricelistItemFieldId, itemId)
	if err != nil {
		return nil, err
	}
	if stored == nil {
		return notFoundResult(models.SalesPricelistItemSchemaName, itemId), nil
	}

	merged := mergedFields(stored, params)
	if vErrs := assertRuleConsistent(merged); vErrs != nil {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}, nil
	}
	if vErrs, err := this.assertNoCycle(ctx, merged, params); err != nil || vErrs != nil {
		if err != nil {
			return nil, err
		}
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}, nil
	}

	return this.DynamicResourceService.Update(ctx, params)
}

// assertNoCycle refuses a rule that would make pricelist derivation circular.
//
// Skipped entirely when the change does not touch the derivation: walking the graph costs a read
// per level, and a rule that only moves its price cannot create a loop.
func (this *SalesPricelistItemDomainServiceImpl) assertNoCycle(
	ctx corectx.Context, merged dmodel.DynamicFields, changed dmodel.DynamicFields,
) (*ft.ClientErrors, error) {
	_, touchesBase := changed[models.SalesPricelistItemFieldBasePricelistId]
	_, touchesSource := changed[models.SalesPricelistItemFieldBasePriceSource]
	if !touchesBase && !touchesSource {
		return nil, nil
	}
	if stringOf(merged, models.SalesPricelistItemFieldBasePriceSource) !=
		models.PricelistBaseSourceOtherPricelist {
		return nil, nil
	}

	baseId := stringOf(merged, models.SalesPricelistItemFieldBasePricelistId)
	if baseId == "" {
		return violationsOf(models.SalesPricelistItemFieldBasePricelistId,
			"sales_pricelist_item.base_pricelist_required",
			"a rule priced from another pricelist must name which one"), nil
	}

	engine, err := engineFor(models.SalesPricelistItemSchemaName)
	if err != nil {
		return nil, err
	}
	ownerId := stringOf(merged, models.SalesPricelistItemFieldSalesPricelistId)
	reader := newRepoPricelistBaseReader(ctx, engine.ResourceRepository())

	if err := AssertNoPricelistCycle(ownerId, baseId, reader); err != nil {
		// A cycle is the user's problem to fix, not a server fault: the message names what is
		// circular and the remedy is to point the rule somewhere else.
		return violationsOf(models.SalesPricelistItemFieldBasePricelistId,
			"sales_pricelist_item.base_pricelist_cycle", err.Error()), nil
	}
	return nil, nil
}

// assertRuleConsistent checks a rule against itself: the right target for its applies_to, and the
// right operands for its calculation_method.
//
// Pure, so the whole rule table is testable without a database. Every branch answers the same
// question — does this row say something coherent — and a row that does not is rejected at write
// time rather than silently never matching anything at resolution time. That distinction matters:
// a rule that never matches is invisible, and the operator who wrote it sees a price that did not
// change and no reason why.
func assertRuleConsistent(record dmodel.DynamicFields) *ft.ClientErrors {
	vErrs := ft.NewClientErrors()

	assertTargetConsistent(record, vErrs)
	assertCalculationConsistent(record, vErrs)
	assertValidityConsistent(record, vErrs)

	if vErrs.Count() == 0 {
		return nil
	}
	return vErrs
}

// assertTargetConsistent enforces exactly one target, chosen by applies_to (section 12).
//
// Both directions are checked. A missing target makes the rule match nothing; a SURPLUS target
// makes the row lie about what it prices — a variant rule that also carries a category id reads,
// to anyone browsing the table, as though it applied to the category too.
func assertTargetConsistent(record dmodel.DynamicFields, vErrs *ft.ClientErrors) {
	variantId := stringOf(record, models.SalesPricelistItemFieldProductVariantId)
	templateId := stringOf(record, models.SalesPricelistItemFieldProductTemplateId)
	categoryId := stringOf(record, models.SalesPricelistItemFieldProductCategoryId)

	appliesTo := stringOf(record, models.SalesPricelistItemFieldAppliesTo)
	if appliesTo == "" {
		appliesTo = models.PricelistAppliesToVariant
	}

	var required string
	switch appliesTo {
	case models.PricelistAppliesToVariant:
		required = models.SalesPricelistItemFieldProductVariantId
	case models.PricelistAppliesToTemplate:
		required = models.SalesPricelistItemFieldProductTemplateId
	case models.PricelistAppliesToCategory:
		required = models.SalesPricelistItemFieldProductCategoryId
	case models.PricelistAppliesToAllProducts:
		required = ""
	default:
		vErrs.Append(*ft.NewBusinessViolation(models.SalesPricelistItemFieldAppliesTo,
			"sales_pricelist_item.applies_to_unknown",
			"'"+appliesTo+"' is not a target a pricing rule can have"))
		return
	}

	// A slice rather than a map: two surplus targets would otherwise be reported in whichever
	// order the map happened to iterate, so the same bad row would produce different errors on
	// different runs — and a client showing "the first problem" would show a different one each
	// time.
	present := []struct {
		field string
		value string
	}{
		{models.SalesPricelistItemFieldProductVariantId, variantId},
		{models.SalesPricelistItemFieldProductTemplateId, templateId},
		{models.SalesPricelistItemFieldProductCategoryId, categoryId},
	}

	for _, target := range present {
		switch {
		case target.field == required && target.value == "":
			vErrs.Append(*ft.NewBusinessViolation(target.field,
				"sales_pricelist_item.target_required",
				"a rule that applies to "+appliesTo+" must name which one"))
		case target.field != required && target.value != "":
			vErrs.Append(*ft.NewBusinessViolation(target.field,
				"sales_pricelist_item.target_not_allowed",
				"a rule that applies to "+appliesTo+" must not also name a "+target.field))
		}
	}
}

// assertCalculationConsistent checks that a rule carries the operands its method needs (§13, §14).
func assertCalculationConsistent(record dmodel.DynamicFields, vErrs *ft.ClientErrors) {
	method := stringOf(record, models.SalesPricelistItemFieldCalculationMethod)
	if method == "" {
		method = models.PricelistMethodFixedPrice
	}

	switch method {
	case models.PricelistMethodFixedPrice:
		// The price may be zero — a giveaway is a real price — so its ABSENCE is what is refused,
		// not its value. A fixed-price rule with no price would match and then quote nothing.
		if _, present := record[models.SalesPricelistItemFieldPrice]; !present {
			vErrs.Append(*ft.NewBusinessViolation(models.SalesPricelistItemFieldPrice,
				"sales_pricelist_item.price_required",
				"a fixed-price rule must state its price"))
		}
		if stringOf(record, models.SalesPricelistItemFieldUomId) == "" {
			vErrs.Append(*ft.NewBusinessViolation(models.SalesPricelistItemFieldUomId,
				"sales_pricelist_item.uom_required",
				"a fixed price is per a unit, so the rule must say which"))
		}

	case models.PricelistMethodDiscount:
		if _, present := record[models.SalesPricelistItemFieldDiscountPercent]; !present {
			vErrs.Append(*ft.NewBusinessViolation(models.SalesPricelistItemFieldDiscountPercent,
				"sales_pricelist_item.discount_required",
				"a discount rule must state its percentage"))
		}

	case models.PricelistMethodFormula:
		source := stringOf(record, models.SalesPricelistItemFieldBasePriceSource)
		if source == "" {
			vErrs.Append(*ft.NewBusinessViolation(models.SalesPricelistItemFieldBasePriceSource,
				"sales_pricelist_item.base_source_required",
				"a formula rule must say what it prices from"))
			return
		}
		if source == models.PricelistBaseSourceOtherPricelist &&
			stringOf(record, models.SalesPricelistItemFieldBasePricelistId) == "" {
			vErrs.Append(*ft.NewBusinessViolation(models.SalesPricelistItemFieldBasePricelistId,
				"sales_pricelist_item.base_pricelist_required",
				"a rule priced from another pricelist must name which one"))
		}

	default:
		vErrs.Append(*ft.NewBusinessViolation(models.SalesPricelistItemFieldCalculationMethod,
			"sales_pricelist_item.method_unknown",
			"'"+method+"' is not a way a rule can compute a price"))
	}
}

// assertValidityConsistent refuses a window that never opens.
//
// A rule whose valid_from is after its valid_to matches on no date at all. Nothing would fail; the
// rule would simply never apply, and its author would be left looking for a price change that
// never happened.
func assertValidityConsistent(record dmodel.DynamicFields, vErrs *ft.ClientErrors) {
	from := stringOf(record, models.SalesPricelistItemFieldValidFrom)
	to := stringOf(record, models.SalesPricelistItemFieldValidTo)
	if from == "" || to == "" {
		return // An open-ended window is the normal case, not an incomplete one.
	}
	// Both are ISO-8601 in UTC, so lexical order is chronological order.
	if from > to {
		vErrs.Append(*ft.NewBusinessViolation(models.SalesPricelistItemFieldValidTo,
			"sales_pricelist_item.validity_inverted",
			"this rule stops applying before it starts, so it would never apply at all"))
	}
}

// violationsOf wraps a single business violation as a ClientErrors.
func violationsOf(field, key, message string) *ft.ClientErrors {
	vErrs := ft.NewClientErrors()
	vErrs.Append(*ft.NewBusinessViolation(field, key, message))
	return vErrs
}
