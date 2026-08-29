package services

import (
	"github.com/shopspring/decimal"
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"

	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/models"
)

// Alternatives: asking several vendors for the same thing and comparing what comes back. A
// sourcing group only records that these orders answer the same requirement; it carries no fields
// of its own, and the engine refuses direct creation because a hand-made group would be an empty
// container nothing reaps.

// The answers to the confirm-time alternatives warning.
const (
	// AlternativeChoiceKeep leaves the other alternatives open.
	AlternativeChoiceKeep = "keep_alternatives"
	// AlternativeChoiceCancel cancels the others; this is the usual outcome.
	AlternativeChoiceCancel = "cancel_alternatives"
)

// CreateAlternative raises a second order for the same requirement against a different vendor. Both
// orders end up in one sourcing group, created if the source has none. The new order copies the
// source's lines but not its vendor.
func (this *PurchaseOrderDomainServiceImpl) CreateAlternative(
	ctx corectx.Context, orderId string, vendorId string,
) (*dyn.OpResult[dmodel.DynamicFields], error) {
	var result *dyn.OpResult[dmodel.DynamicFields]

	err := withOrderTransaction(ctx, func(tranxCtx corectx.Context) error {
		source, err := loadOrder(tranxCtx, orderId)
		if err != nil {
			return err
		}
		if source == nil {
			notFound := orderNotFoundResult(orderId)
			result = &dyn.OpResult[dmodel.DynamicFields]{ClientErrors: notFound.ClientErrors}
			return nil
		}
		if vendorId == "" {
			vErrs := ft.NewClientErrors()
			vErrs.Append(*ft.NewBusinessViolation(models.PurchaseOrderFieldVendorId,
				"purchase_order.alternative_vendor_required",
				"an alternative must name the vendor to ask instead"))
			result = &dyn.OpResult[dmodel.DynamicFields]{ClientErrors: *vErrs}
			return nil
		}

		status := stringOf(source, models.PurchaseOrderFieldStatus)
		if !isQuotableStatus(status) {
			vErrs := ft.NewClientErrors()
			vErrs.Append(*ft.NewBusinessViolation(models.PurchaseOrderSchemaName,
				"purchase_order.not_alternatable",
				"alternatives can only be raised while the order is still a request for "+
					"quotation; this one is '"+status+"'"))
			result = &dyn.OpResult[dmodel.DynamicFields]{ClientErrors: *vErrs}
			return nil
		}

		groupId, err := this.ensureSourcingGroup(tranxCtx, source)
		if err != nil {
			return err
		}

		params := copyableOrderFields(source)
		params[models.PurchaseOrderFieldVendorId] = vendorId
		params[models.PurchaseOrderFieldSourcingGroupId] = groupId
		// The source's currency came from its own vendor; dropping it lets the new vendor's default
		// apply instead of quoting them in a currency they may not trade in.
		delete(params, models.PurchaseOrderFieldCurrencyId)

		created, err := this.Create(tranxCtx, params)
		if err != nil || created.ClientErrors.Count() > 0 {
			result = created
			return err
		}
		result = created

		newOrderId := stringOf(created.Data, models.PurchaseOrderFieldId)
		if err := this.copyOrderLines(tranxCtx, orderId, newOrderId); err != nil {
			return err
		}
		return RecomputeOrderTotals(tranxCtx, newOrderId)
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

// isQuotableStatus reports whether an order is still at the stage where alternatives make sense;
// once confirmed, the business has already committed to buying.
func isQuotableStatus(status string) bool {
	return status == string(models.PurchaseOrderStatusRfq) ||
		status == string(models.PurchaseOrderStatusRfqSent)
}

// ensureSourcingGroup returns the source's group, creating one if it has none.
func (this *PurchaseOrderDomainServiceImpl) ensureSourcingGroup(
	ctx corectx.Context, source dmodel.DynamicFields,
) (string, error) {
	if existing := stringOf(source, models.PurchaseOrderFieldSourcingGroupId); existing != "" {
		return existing, nil
	}

	groupEngine, err := engineFor(models.SourcingGroupSchemaName)
	if err != nil {
		return "", err
	}
	id, err := model.NewId()
	if err != nil {
		return "", errors.Wrap(err, "ensureSourcingGroup")
	}

	group := dmodel.DynamicFields{models.SourcingGroupFieldId: *id}
	if orgId := stringOf(source, basemodel.FieldOrgId); orgId != "" {
		group[basemodel.FieldOrgId] = orgId
	}
	// Through the repository, not the service: the engine's guard refuses client creation of a
	// sourcing group and must keep doing so while the system still creates its own.
	if _, err := groupEngine.ResourceRepository().Insert(ctx, group); err != nil {
		return "", errors.Wrap(err, "ensureSourcingGroup")
	}

	if err := writeOrderChanges(ctx, source, dmodel.DynamicFields{
		models.PurchaseOrderFieldSourcingGroupId: *id,
	}); err != nil {
		return "", err
	}
	return string(*id), nil
}

// AlternativeComparison is one row of the comparison view.
type AlternativeComparison struct {
	OrderId     string
	Code        string
	VendorId    string
	CurrencyId  string
	Status      string
	TotalAmount decimal.Decimal

	// IsCheapest marks the lowest total in the group. It is only meaningful when ComparableByPrice
	// is true: totals in different currencies compared by their numbers name the wrong winner.
	IsCheapest bool
}

// AlternativeComparisonResult is the comparison of one sourcing group.
type AlternativeComparisonResult struct {
	Alternatives []AlternativeComparison

	// ComparableByPrice is false when the alternatives are not all in one currency. There is no
	// exchange rate model, so such totals cannot be ranked at all.
	ComparableByPrice bool
}

// CompareAlternatives lists the orders in a sourcing group side by side.
func CompareAlternatives(
	ctx corectx.Context, orderId string,
) (*AlternativeComparisonResult, error) {
	order, err := loadOrder(ctx, orderId)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return &AlternativeComparisonResult{}, nil
	}

	groupId := stringOf(order, models.PurchaseOrderFieldSourcingGroupId)
	if groupId == "" {
		// An order with no alternatives compares to itself; that is an answer, not an error.
		return &AlternativeComparisonResult{
			Alternatives:      []AlternativeComparison{comparisonRow(order)},
			ComparableByPrice: true,
		}, nil
	}

	orders, err := loadSourcingGroupOrders(ctx, groupId)
	if err != nil {
		return nil, err
	}
	return buildComparison(orders), nil
}

func buildComparison(orders []dmodel.DynamicFields) *AlternativeComparisonResult {
	result := &AlternativeComparisonResult{
		Alternatives:      make([]AlternativeComparison, 0, len(orders)),
		ComparableByPrice: true,
	}

	currency := ""
	for index, order := range orders {
		row := comparisonRow(order)
		if index == 0 {
			currency = row.CurrencyId
		} else if row.CurrencyId != currency {
			result.ComparableByPrice = false
		}
		result.Alternatives = append(result.Alternatives, row)
	}

	if !result.ComparableByPrice || len(result.Alternatives) == 0 {
		return result
	}

	cheapest := 0
	for index := 1; index < len(result.Alternatives); index++ {
		if result.Alternatives[index].TotalAmount.LessThan(
			result.Alternatives[cheapest].TotalAmount) {
			cheapest = index
		}
	}
	result.Alternatives[cheapest].IsCheapest = true
	return result
}

func comparisonRow(order dmodel.DynamicFields) AlternativeComparison {
	return AlternativeComparison{
		OrderId:     stringOf(order, models.PurchaseOrderFieldId),
		Code:        stringOf(order, models.PurchaseOrderFieldCode),
		VendorId:    stringOf(order, models.PurchaseOrderFieldVendorId),
		CurrencyId:  stringOf(order, models.PurchaseOrderFieldCurrencyId),
		Status:      stringOf(order, models.PurchaseOrderFieldStatus),
		TotalAmount: decimalOf(order, models.PurchaseOrderFieldTotalAmount),
	}
}

// OpenAlternativesOf returns the other orders in this order's sourcing group that are still open.
// The confirm-time warning is built on this: confirming one alternative leaves the others quoting
// for a requirement already met.
func OpenAlternativesOf(ctx corectx.Context, order dmodel.DynamicFields) ([]dmodel.DynamicFields, error) {
	groupId := stringOf(order, models.PurchaseOrderFieldSourcingGroupId)
	if groupId == "" {
		return nil, nil
	}

	siblings, err := loadSourcingGroupOrders(ctx, groupId)
	if err != nil {
		return nil, err
	}

	orderId := stringOf(order, models.PurchaseOrderFieldId)
	open := make([]dmodel.DynamicFields, 0, len(siblings))
	for _, sibling := range siblings {
		if stringOf(sibling, models.PurchaseOrderFieldId) == orderId {
			continue
		}
		if isQuotableStatus(stringOf(sibling, models.PurchaseOrderFieldStatus)) {
			open = append(open, sibling)
		}
	}
	return open, nil
}

// CancelOpenAlternatives cancels the other still-open orders in the group.
func (this *PurchaseOrderDomainServiceImpl) CancelOpenAlternatives(
	ctx corectx.Context, order dmodel.DynamicFields, confirmedOrderId string,
) error {
	open, err := OpenAlternativesOf(ctx, order)
	if err != nil {
		return err
	}

	for _, alternative := range open {
		if err := writeOrderChanges(ctx, alternative, dmodel.DynamicFields{
			models.PurchaseOrderFieldStatus: string(models.PurchaseOrderStatusCancelled),
		}); err != nil {
			return err
		}
		if err := WriteAuditEvent(ctx, AuditEntry{
			EntityType: models.PurchaseOrderSchemaName,
			EntityId:   stringOf(alternative, models.PurchaseOrderFieldId),
			Action:     AuditActionCancel,
			FromStatus: stringOf(alternative, models.PurchaseOrderFieldStatus),
			ToStatus:   string(models.PurchaseOrderStatusCancelled),
			Reason:     "an alternative for the same requirement was confirmed",
			OrgId:      stringOf(alternative, basemodel.FieldOrgId),
			Metadata:   map[string]any{"confirmed_alternative_id": confirmedOrderId},
		}); err != nil {
			return err
		}
	}
	return ReapSourcingGroup(ctx, stringOf(order, models.PurchaseOrderFieldSourcingGroupId))
}

// ReapSourcingGroup removes a group that no longer has two alternatives to compare. The surviving
// orders' pointers are cleared first so nothing references a group that has gone.
func ReapSourcingGroup(ctx corectx.Context, groupId string) error {
	if groupId == "" {
		return nil
	}

	orders, err := loadSourcingGroupOrders(ctx, groupId)
	if err != nil {
		return err
	}

	live := make([]dmodel.DynamicFields, 0, len(orders))
	for _, order := range orders {
		if IsOrderOpen(stringOf(order, models.PurchaseOrderFieldStatus)) {
			live = append(live, order)
		}
	}
	if len(live) > 1 {
		return nil
	}

	for _, order := range orders {
		if err := writeOrderChanges(ctx, order, dmodel.DynamicFields{
			models.PurchaseOrderFieldSourcingGroupId: nil,
		}); err != nil {
			return err
		}
	}

	groupEngine, err := engineFor(models.SourcingGroupSchemaName)
	if err != nil {
		return err
	}
	_, err = groupEngine.ResourceRepository().DeleteOne(ctx, dmodel.DynamicFields{
		models.SourcingGroupFieldId: groupId,
	})
	return errors.Wrap(err, "ReapSourcingGroup")
}

func loadSourcingGroupOrders(
	ctx corectx.Context, groupId string,
) ([]dmodel.DynamicFields, error) {
	engine, err := engineFor(models.PurchaseOrderSchemaName)
	if err != nil {
		return nil, err
	}
	return models.FindOrdersInSourcingGroup(
		ctx, engine.ResourceRepository(), groupId, models.MaxSourcingGroupOrders)
}
