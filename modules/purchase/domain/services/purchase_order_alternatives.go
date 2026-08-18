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

// Alternatives: asking several vendors for the same thing and comparing what comes back
// (BR §27-§31).
//
// A sourcing group is a technical record with no meaning of its own (§28): it exists only to say
// that these orders are alternatives for the same requirement. It carries no fields beyond the base
// ones, which is why the engine refuses direct creation — a hand-made group would be an empty
// container that nothing reaps.

// The answers to the confirm-time warning of §31.
const (
	// AlternativeChoiceKeep leaves the other alternatives open. The buyer may be placing more than
	// one of them, or wants to keep quoting until the goods arrive.
	AlternativeChoiceKeep = "keep_alternatives"
	// AlternativeChoiceCancel cancels the others, which is the usual outcome: the requirement is
	// met by the order being confirmed and the remaining quotes are no longer wanted.
	AlternativeChoiceCancel = "cancel_alternatives"
)

// CreateAlternative raises a second order for the same requirement, against a different vendor
// (BR §27).
//
// The two orders end up in one sourcing group, creating it if the source is not already in one. The
// new order copies the source's LINES but not its vendor: the whole point is to ask somebody else,
// and copying the vendor would produce two identical requests to the same supplier.
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
		// The source's currency came from ITS vendor, so it is dropped: the new vendor's own
		// default applies, and carrying the old one would quote a second supplier in a currency
		// they may not trade in.
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

// isQuotableStatus reports whether an order is still at the stage where alternatives make sense.
//
// Once confirmed, the decision has been made: raising an alternative to an order already placed
// would be quoting for goods the business has already committed to buying.
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
	// sourcing group, and it must keep refusing it while the system still creates its own.
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

// AlternativeComparison is one row of the comparison view (BR §30).
type AlternativeComparison struct {
	OrderId     string
	Code        string
	VendorId    string
	CurrencyId  string
	Status      string
	TotalAmount decimal.Decimal

	// IsCheapest marks the lowest total in the group. It is computed here rather than left to the
	// caller so that "cheapest" means the same thing in the API and the UI.
	//
	// It is only meaningful when every alternative is in the same currency, which is why
	// ComparableByPrice exists: comparing 100 USD against 100 VND by their numbers would name the
	// wrong winner with complete confidence.
	IsCheapest bool
}

// AlternativeComparisonResult is the comparison of one sourcing group.
type AlternativeComparisonResult struct {
	Alternatives []AlternativeComparison

	// ComparableByPrice is false when the alternatives are not all in one currency. No exchange
	// rate model exists (D5), so the totals genuinely cannot be ranked, and saying so is better
	// than ranking them wrongly.
	ComparableByPrice bool
}

// CompareAlternatives lists the orders in a sourcing group side by side (BR §30).
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
		// An order with no alternatives compares to itself, which is a legitimate answer rather
		// than an error: the caller asked what the alternatives were, and there are none.
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
//
// This is what the confirm-time warning of §31 is built on: confirming one alternative leaves the
// others quoting for a requirement that has just been met, and the buyer has to say what happens
// to them.
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

// CancelOpenAlternatives cancels the other still-open orders in the group (§31,
// CANCEL_ALTERNATIVES).
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

// ReapSourcingGroup removes a group that no longer has two alternatives to compare (§28).
//
// A group of one is not a comparison, and leaving it behind would show a lone order as though it
// were being weighed against something. The surviving order's pointer is cleared first, so nothing
// is left referencing a group that has gone.
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
