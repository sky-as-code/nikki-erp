package services

import (
	"time"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"

	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/models"
)

// Merging several draft requests for quotation into one (BR §26).
//
// The point is a real purchasing situation: three people each raised an RFQ for the same vendor,
// and sending three separate documents would get three separate quotes and three deliveries. Merge
// makes one document out of them.
//
// Only RFQ and RFQ_SENT may be merged. A confirmed order is a commitment the vendor is holding, and
// folding it into another document would change what was agreed after the fact.

// arrivalWindow is how far apart two lines' expected arrival dates may be and still merge (§26.1).
//
// A day is the resolution a purchasing decision is actually made at: goods wanted "Tuesday" and
// "Tuesday morning" are one delivery, and splitting them into two lines would ask the vendor for two
// shipments of the same product. Beyond a day they are genuinely different requests.
const arrivalWindow = 24 * time.Hour

// MergeOrders folds several draft orders into one (BR §26).
//
// The TARGET is the oldest by order deadline, and keeps its own code: it is the document most
// likely to have been quoted to the vendor already, and keeping the oldest reference means the
// vendor's own paperwork still matches. The sources are CANCELLED rather than deleted, so the trail
// of what was merged where survives.
func (this *PurchaseOrderDomainServiceImpl) MergeOrders(
	ctx corectx.Context, orderIds []string,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	if len(orderIds) < 2 {
		return orderViolationResult("purchase_order.merge_needs_two",
			"merging needs at least two purchase orders"), nil
	}

	var result *dyn.OpResult[dyn.MutateResultData]

	err := withOrderTransaction(ctx, func(tranxCtx corectx.Context) error {
		orders, refusal, err := loadMergeableOrders(tranxCtx, orderIds)
		if err != nil {
			return err
		}
		if refusal != nil {
			result = refusal
			return nil
		}

		target, sources := splitMergeTarget(orders)
		targetId := stringOf(target, models.PurchaseOrderFieldId)

		mergedFrom := make([]any, 0, len(sources))
		for _, source := range sources {
			sourceId := stringOf(source, models.PurchaseOrderFieldId)
			if err := this.mergeOneOrder(tranxCtx, source, targetId); err != nil {
				return err
			}
			mergedFrom = append(mergedFrom, sourceId)
		}

		if err := RecomputeOrderTotals(tranxCtx, targetId); err != nil {
			return err
		}

		result = mutateOk()
		// One event on the target, listing what came into it. The sources each carry their own
		// cancel event from mergeOneOrder, so the trail reads correctly from either end.
		return WriteAuditEvent(tranxCtx, AuditEntry{
			EntityType: models.PurchaseOrderSchemaName,
			EntityId:   targetId,
			Action:     AuditActionMerge,
			OrgId:      stringOf(target, basemodel.FieldOrgId),
			Metadata: map[string]any{
				"merged_from_purchase_order_ids": mergedFrom,
			},
		})
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

// loadMergeableOrders reads the orders and refuses a set that cannot be merged.
//
// Compatibility is vendor, currency and agreement (§26). Those three are what a purchase order
// commits to: merging across vendors would ask one supplier for another's goods, merging across
// currencies would produce a document whose total is in no currency at all, and merging across
// agreements would draw down a commitment the lines were never made under.
func loadMergeableOrders(
	ctx corectx.Context, orderIds []string,
) ([]dmodel.DynamicFields, *dyn.OpResult[dyn.MutateResultData], error) {
	orders := make([]dmodel.DynamicFields, 0, len(orderIds))
	seen := map[string]bool{}

	for _, orderId := range orderIds {
		if seen[orderId] {
			// The same order named twice would otherwise be merged into itself.
			continue
		}
		seen[orderId] = true

		order, err := loadOrder(ctx, orderId)
		if err != nil {
			return nil, nil, err
		}
		if order == nil {
			return nil, orderNotFoundResult(orderId), nil
		}

		status := stringOf(order, models.PurchaseOrderFieldStatus)
		if status != string(models.PurchaseOrderStatusRfq) &&
			status != string(models.PurchaseOrderStatusRfqSent) {
			return nil, orderViolationResult("purchase_order.not_mergeable",
				"only a draft or sent request for quotation can be merged; '"+
					stringOf(order, models.PurchaseOrderFieldCode)+"' is '"+status+"'"), nil
		}
		orders = append(orders, order)
	}

	if len(orders) < 2 {
		return nil, orderViolationResult("purchase_order.merge_needs_two",
			"merging needs at least two distinct purchase orders"), nil
	}
	if refusal := assertCompatibleOrders(orders); refusal != nil {
		return nil, refusal, nil
	}
	return orders, nil, nil
}

func assertCompatibleOrders(orders []dmodel.DynamicFields) *dyn.OpResult[dyn.MutateResultData] {
	first := orders[0]
	for _, field := range []struct {
		name    string
		key     string
		message string
	}{
		{"vendor", models.PurchaseOrderFieldVendorId,
			"every order in a merge must name the same vendor"},
		{"currency", models.PurchaseOrderFieldCurrencyId,
			"every order in a merge must be in the same currency"},
		{"agreement", models.PurchaseOrderFieldAgreementId,
			"every order in a merge must be drawn against the same agreement"},
	} {
		want := stringOf(first, field.key)
		for _, order := range orders[1:] {
			if stringOf(order, field.key) != want {
				return orderViolationResult(
					"purchase_order.merge_"+field.name+"_mismatch", field.message)
			}
		}
	}
	return nil
}

// splitMergeTarget picks the oldest order as the target and returns the rest as sources.
//
// Oldest by order deadline, falling back to the code when a deadline is absent. The fallback is not
// arbitrary: the code carries a ULID, which sorts by creation time, so "oldest" still means oldest
// even for orders that never had a deadline set.
func splitMergeTarget(
	orders []dmodel.DynamicFields,
) (dmodel.DynamicFields, []dmodel.DynamicFields) {
	targetIndex := 0
	for index := 1; index < len(orders); index++ {
		if isOlderOrder(orders[index], orders[targetIndex]) {
			targetIndex = index
		}
	}

	sources := make([]dmodel.DynamicFields, 0, len(orders)-1)
	for index, order := range orders {
		if index != targetIndex {
			sources = append(sources, order)
		}
	}
	return orders[targetIndex], sources
}

func isOlderOrder(candidate, current dmodel.DynamicFields) bool {
	candidateDeadline, candidateHas := timeOf(candidate, models.PurchaseOrderFieldOrderDeadline)
	currentDeadline, currentHas := timeOf(current, models.PurchaseOrderFieldOrderDeadline)

	switch {
	case candidateHas && currentHas:
		return candidateDeadline.Before(currentDeadline)
	case candidateHas:
		// A stated deadline beats none: an order somebody put a date on is the one being worked to.
		return true
	case currentHas:
		return false
	default:
		return stringOf(candidate, models.PurchaseOrderFieldCode) <
			stringOf(current, models.PurchaseOrderFieldCode)
	}
}

// mergeOneOrder moves one source order's lines onto the target and cancels the source.
func (this *PurchaseOrderDomainServiceImpl) mergeOneOrder(
	ctx corectx.Context, source dmodel.DynamicFields, targetId string,
) error {
	lineEngine, err := engineFor(models.PurchaseOrderLineSchemaName)
	if err != nil {
		return err
	}
	sourceId := stringOf(source, models.PurchaseOrderFieldId)

	sourceLines, err := models.FindOrderLines(
		ctx, lineEngine.ResourceRepository(), sourceId, models.MaxOrderLines)
	if err != nil {
		return err
	}
	targetLines, err := models.FindOrderLines(
		ctx, lineEngine.ResourceRepository(), targetId, models.MaxOrderLines)
	if err != nil {
		return err
	}

	for _, sourceLine := range sourceLines {
		match := findMergeableLine(targetLines, sourceLine)
		if match == nil {
			// No compatible line: the source line moves across as a line of its own (§26.1).
			if err := moveLineToOrder(ctx, lineEngine, sourceLine, targetId); err != nil {
				return err
			}
			continue
		}
		if err := addQuantityToLine(ctx, lineEngine, match, sourceLine); err != nil {
			return err
		}
	}

	// The source is cancelled, not deleted: what was merged where has to stay readable.
	if err := writeOrderChanges(ctx, source, dmodel.DynamicFields{
		models.PurchaseOrderFieldStatus: string(models.PurchaseOrderStatusCancelled),
	}); err != nil {
		return err
	}
	return WriteAuditEvent(ctx, AuditEntry{
		EntityType: models.PurchaseOrderSchemaName,
		EntityId:   sourceId,
		Action:     AuditActionMerge,
		FromStatus: stringOf(source, models.PurchaseOrderFieldStatus),
		ToStatus:   string(models.PurchaseOrderStatusCancelled),
		OrgId:      stringOf(source, basemodel.FieldOrgId),
		Metadata:   map[string]any{"merged_into_purchase_order_id": targetId},
	})
}

// findMergeableLine returns the target line the source may be added to, or nil.
//
// The rules are §26.1: same product, same unit, same discount, and expected arrival within a day.
// Discount is included because two lines at different discounts are at different effective prices,
// and summing them would produce a quantity at a price neither of them had.
func findMergeableLine(
	targetLines []dmodel.DynamicFields, sourceLine dmodel.DynamicFields,
) dmodel.DynamicFields {
	if !isMoneyBearingLine(sourceLine) {
		// A section or a note is a piece of document structure, not a quantity. Merging two
		// headings would silently drop one.
		return nil
	}
	for _, targetLine := range targetLines {
		if linesAreMergeable(targetLine, sourceLine) {
			return targetLine
		}
	}
	return nil
}

func linesAreMergeable(target, source dmodel.DynamicFields) bool {
	if !isMoneyBearingLine(target) {
		return false
	}

	targetProduct := stringOf(target, models.PurchaseOrderLineFieldProductVariantId)
	// Two free-text lines are never merged: they have no product to compare, so "same product"
	// would be vacuously true and two unrelated charges would be summed.
	if targetProduct == "" ||
		targetProduct != stringOf(source, models.PurchaseOrderLineFieldProductVariantId) {
		return false
	}
	if stringOf(target, models.PurchaseOrderLineFieldUomId) !=
		stringOf(source, models.PurchaseOrderLineFieldUomId) {
		return false
	}
	if !decimalOf(target, models.PurchaseOrderLineFieldDiscountPercent).Equal(
		decimalOf(source, models.PurchaseOrderLineFieldDiscountPercent)) {
		return false
	}
	return arrivalsAreClose(target, source)
}

// arrivalsAreClose reports whether two lines want their goods at about the same time.
//
// Two lines with NO stated arrival are close — neither has asked for a date, so there is nothing to
// disagree about. One with a date and one without are not: the request that named a day is asking
// for something the other is not.
func arrivalsAreClose(target, source dmodel.DynamicFields) bool {
	targetArrival, targetHas := timeOf(target, models.PurchaseOrderLineFieldExpectedArrival)
	sourceArrival, sourceHas := timeOf(source, models.PurchaseOrderLineFieldExpectedArrival)

	if !targetHas && !sourceHas {
		return true
	}
	if targetHas != sourceHas {
		return false
	}

	gap := targetArrival.Sub(sourceArrival)
	if gap < 0 {
		gap = -gap
	}
	return gap <= arrivalWindow
}

// addQuantityToLine sums the source line's quantity into the matching target line.
func addQuantityToLine(
	ctx corectx.Context, lineEngine drif.DynamicResourceEngine,
	target, source dmodel.DynamicFields,
) error {
	combined := decimalOf(target, models.PurchaseOrderLineFieldQuantity).Add(
		decimalOf(source, models.PurchaseOrderLineFieldQuantity))
	// The taxes are summed too: tax is an input per line (D9), so the merged line owes what the two
	// lines owed between them.
	combinedTax := decimalOf(target, models.PurchaseOrderLineFieldTaxAmount).Add(
		decimalOf(source, models.PurchaseOrderLineFieldTaxAmount))

	merged := dmodel.DynamicFields{}
	for key, value := range target {
		merged[key] = value
	}
	merged[models.PurchaseOrderLineFieldQuantity] = combined
	merged[models.PurchaseOrderLineFieldTaxAmount] = combinedTax
	StampLineTotals(merged)

	if _, err := lineEngine.ResourceRepository().Update(ctx, dmodel.DynamicFields{
		models.PurchaseOrderLineFieldId:        stringOf(target, models.PurchaseOrderLineFieldId),
		models.PurchaseOrderLineFieldQuantity:  combined,
		models.PurchaseOrderLineFieldTaxAmount: combinedTax,
		models.PurchaseOrderLineFieldSubtotal:  merged[models.PurchaseOrderLineFieldSubtotal],
		models.PurchaseOrderLineFieldTotal:     merged[models.PurchaseOrderLineFieldTotal],
		basemodel.FieldEtag:                    stringOf(target, basemodel.FieldEtag),
	}); err != nil {
		return err
	}
	// The in-memory copy is updated too, so a second source line matching the same target sums onto
	// the new quantity rather than the one it replaced.
	target[models.PurchaseOrderLineFieldQuantity] = combined
	target[models.PurchaseOrderLineFieldTaxAmount] = combinedTax

	// The source line is removed: it now lives inside the target's quantity, and leaving it on a
	// cancelled order would double-count if anyone read both.
	_, err := lineEngine.ResourceRepository().DeleteOne(ctx, dmodel.DynamicFields{
		models.PurchaseOrderLineFieldId: stringOf(source, models.PurchaseOrderLineFieldId),
	})
	return err
}

// moveLineToOrder repoints one line at another order.
func moveLineToOrder(
	ctx corectx.Context, lineEngine drif.DynamicResourceEngine,
	line dmodel.DynamicFields, targetId string,
) error {
	_, err := lineEngine.ResourceRepository().Update(ctx, dmodel.DynamicFields{
		models.PurchaseOrderLineFieldId:              stringOf(line, models.PurchaseOrderLineFieldId),
		models.PurchaseOrderLineFieldPurchaseOrderId: targetId,
		basemodel.FieldEtag:                          stringOf(line, basemodel.FieldEtag),
	})
	return err
}

// timeOf reads a time field, reporting whether one was there at all.
//
// The bool matters: "no arrival date" and "the zero time" mean different things to the merge rule,
// and collapsing them would make every dateless line look like it wanted delivery in year one. It
// matters again to vendor price validity (windowCovers), where an unread bound would read as
// open-ended and quietly resurrect an expired quote.
//
// model.ModelDateTime is handled because that is what the model layer's own SetModelDateTime
// stores, and DynamicFields.GetModelDateTime — the obvious reader — returns nil for the pointer
// form rather than reporting that it could not read it.
func timeOf(fields dmodel.DynamicFields, key string) (time.Time, bool) {
	value, ok := fields[key]
	if !ok || value == nil {
		return time.Time{}, false
	}
	switch typed := value.(type) {
	case model.ModelDateTime:
		return typed.GoTime(), !typed.GoTime().IsZero()
	case *model.ModelDateTime:
		if typed == nil {
			return time.Time{}, false
		}
		return typed.GoTime(), !typed.GoTime().IsZero()
	case time.Time:
		return typed, !typed.IsZero()
	case *time.Time:
		if typed == nil {
			return time.Time{}, false
		}
		return *typed, !typed.IsZero()
	case string:
		parsed, err := time.Parse(time.RFC3339, typed)
		if err != nil {
			return time.Time{}, false
		}
		return parsed, true
	default:
		return time.Time{}, false
	}
}
