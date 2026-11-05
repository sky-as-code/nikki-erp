package services

import (
	"sort"
	"time"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"

	"github.com/sky-as-code/nikki-erp/modules/commerce/sales/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/commerce/sales/domain/services/pricing/promotion"
)

// Loading the promotion candidates of a draft (BR-C01, BR-C02, BR-S01). Only promotions that pass
// the base conditions reach the engine: active, not archived, in the order's org, inside
// [starts_at, ends_at) at server time, usage left, and linked to BOTH the order's channel AND its
// point. A voucher-gated promotion is a candidate only through an attached, still-available voucher.

// Warning and reason codes of the PromotionQuote (plan §3.2).
const (
	PromotionWarningContextRequired     = "context_required"
	PromotionWarningCouponNeedsCustomer = "coupon_needs_customer"

	PromotionReasonContextRequired  = "context_required"
	PromotionReasonScopeMismatch    = "scope_mismatch"
	PromotionReasonNotAvailable     = "not_available"
	PromotionReasonConditionNotMet  = "condition_not_met"
	PromotionReasonCustomerMismatch = "voucher_customer_mismatch"
)

// promotionCandidate is one promotion as the order flow carries it around the engine.
type promotionCandidate struct {
	record      dmodel.DynamicFields
	rewards     []dmodel.DynamicFields
	program     promotion.Program
	voucher     dmodel.DynamicFields
	voucherCode string

	// rejectReason is set for an attached voucher whose promotion cannot even be a candidate; the
	// quote still lists it so the cashier sees why the code gives nothing.
	rejectReason string
}

// promotionContext is everything loaded once per reprice.
type promotionContext struct {
	orderId    string
	orgId      string
	customerId string
	currencyId string
	now        time.Time

	candidates  []*promotionCandidate
	rejected    []*promotionCandidate
	issuers     []*promotionCandidate
	selections  []promotion.Selection
	lineRecords []dmodel.DynamicFields
	warnings    []PromotionWarning
}

// PromotionWarning is one quote warning.
type PromotionWarning struct {
	Code        string `json:"code"`
	PromotionId string `json:"promotion_id,omitempty"`
}

func (this *promotionContext) hasWork() bool {
	if len(this.candidates) > 0 || len(this.rejected) > 0 || len(this.issuers) > 0 {
		return true
	}
	for _, line := range this.lineRecords {
		if stringOf(line, models.SalesOrderLineFieldLineType) == string(models.SalesOrderLineTypePromotionGift) {
			return true
		}
	}
	return false
}

func (this *promotionContext) byPromotionId(promotionId string) *promotionCandidate {
	for _, group := range [][]*promotionCandidate{this.candidates, this.rejected, this.issuers} {
		for _, candidate := range group {
			if candidate.program.Id == promotionId {
				return candidate
			}
		}
	}
	return nil
}

func loadPromotionContext(
	ctx corectx.Context, order dmodel.DynamicFields, now time.Time,
) (*promotionContext, error) {
	orderId := stringOf(order, models.SalesOrderFieldId)
	pc := &promotionContext{
		orderId: orderId, orgId: stringOf(order, basemodel.FieldOrgId), now: now,
		customerId: stringOf(order, models.SalesOrderFieldSoldToPartyId),
		currencyId: orderCurrencyId(ctx, order),
	}
	var err error
	if pc.lineRecords, err = searchBy(ctx, models.SalesOrderLineSchemaName,
		models.SalesOrderLineFieldSalesOrderId, orderId); err != nil {
		return nil, err
	}
	active, err := livePromotions(ctx, pc.orgId, now)
	if err != nil {
		return nil, err
	}
	inScope, contextOk, err := promotionsInScope(ctx, order, active)
	if err != nil {
		return nil, err
	}
	if !contextOk {
		pc.warnings = append(pc.warnings, PromotionWarning{Code: PromotionWarningContextRequired})
	}
	if err := pc.collect(ctx, active, inScope, contextOk); err != nil {
		return nil, err
	}
	if pc.selections, err = loadSelections(ctx, orderId); err != nil {
		return nil, err
	}
	return pc, pc.loadChildren(ctx)
}

func orderCurrencyId(ctx corectx.Context, order dmodel.DynamicFields) string {
	code := stringOf(order, models.SalesOrderFieldCurrencyCode)
	if code == "" || promotionPorts.Currencies == nil {
		return ""
	}
	id, err := promotionPorts.Currencies.CurrencyIdByCode(ctx, code)
	if err != nil {
		return ""
	}
	return string(id)
}

// livePromotions are the org's active, unarchived promotions inside their window with usage left.
func livePromotions(ctx corectx.Context, orgId string, now time.Time) (map[string]dmodel.DynamicFields, error) {
	rows, err := searchLiveRows(ctx, models.SalesPromotionSchemaName, orgId,
		equalsCondition(models.SalesPromotionFieldStatus, string(models.PromotionStatusActive)))
	if err != nil {
		return nil, err
	}
	live := map[string]dmodel.DynamicFields{}
	for _, row := range rows {
		if promotionLiveAt(row, now) {
			live[stringOf(row, basemodel.FieldId)] = row
		}
	}
	return live, nil
}

// promotionLiveAt is BR-C02 ([starts_at, ends_at), null open) plus usage left (BR-T03).
func promotionLiveAt(row dmodel.DynamicFields, now time.Time) bool {
	if !withinWindow(row, models.SalesPromotionFieldStartsAt, models.SalesPromotionFieldEndsAt, now) {
		return false
	}
	limit := int32Of(row, models.SalesPromotionFieldUsageLimit)
	return limit <= 0 || int32Of(row, models.SalesPromotionFieldUsedCount) < limit
}

// promotionsInScope is BR-S01: a channel link AND a point link, never one inferred from the other.
// contextOk is false when the order lacks either value (AC-05): no promotion applies then.
func promotionsInScope(
	ctx corectx.Context, order dmodel.DynamicFields, active map[string]dmodel.DynamicFields,
) (map[string]bool, bool, error) {
	channelId := stringOf(order, models.SalesOrderFieldSalesChannelId)
	pointId := stringOf(order, models.SalesOrderFieldSalesPointId)
	inScope := map[string]bool{}
	if channelId == "" || pointId == "" {
		return inScope, false, nil
	}
	if len(active) == 0 {
		return inScope, true, nil
	}
	ids := sortedKeys(active)
	channels, err := searchRows(ctx, models.SalesPromotionChannelSchemaName, false,
		inCondition(models.SalesPromotionChannelFieldPromotionId, ids),
		equalsCondition(models.SalesPromotionChannelFieldChannelId, channelId))
	if err != nil {
		return nil, true, err
	}
	points, err := searchRows(ctx, models.SalesPromotionPointSchemaName, false,
		inCondition(models.SalesPromotionPointFieldPromotionId, ids),
		equalsCondition(models.SalesPromotionPointFieldPointId, pointId))
	if err != nil {
		return nil, true, err
	}
	withPoint := map[string]bool{}
	for _, row := range points {
		withPoint[stringOf(row, models.SalesPromotionPointFieldPromotionId)] = true
	}
	for _, row := range channels {
		if id := stringOf(row, models.SalesPromotionChannelFieldPromotionId); withPoint[id] {
			inScope[id] = true
		}
	}
	return inScope, true, nil
}

func sortedKeys[V any](values map[string]V) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// collect sorts the live promotions into engine candidates and issue-phase programs, then adds the
// voucher-gated ones the attached vouchers unlock.
func (this *promotionContext) collect(
	ctx corectx.Context, active map[string]dmodel.DynamicFields, inScope map[string]bool, contextOk bool,
) error {
	for _, id := range sortedKeys(active) {
		if !inScope[id] {
			continue
		}
		switch models.PromotionProgramType(stringOf(active[id], models.SalesPromotionFieldProgramType)) {
		case models.PromotionTypeAutomatic, models.PromotionTypeBuyXGetY:
			this.candidates = append(this.candidates, newCandidate(active[id]))
		case models.PromotionTypeNextOrderCoupon:
			this.issuers = append(this.issuers, newCandidate(active[id]))
		}
	}
	return this.collectVouchers(ctx, active, inScope, contextOk)
}

func newCandidate(record dmodel.DynamicFields) *promotionCandidate {
	return &promotionCandidate{record: record, program: promotion.Program{
		Id:            stringOf(record, basemodel.FieldId),
		ProgramType:   stringOf(record, models.SalesPromotionFieldProgramType),
		Sequence:      int32Of(record, models.SalesPromotionFieldSequence),
		ConfigVersion: int32Of(record, models.SalesPromotionFieldConfigVersion),
		CurrencyId:    stringOf(record, models.SalesPromotionFieldCurrencyId),
	}}
}

func loadSelections(ctx corectx.Context, orderId string) ([]promotion.Selection, error) {
	rows, err := searchBy(ctx, models.SalesOrderPromotionSelectionSchemaName,
		models.SalesOrderPromotionSelectionFieldOrderId, orderId)
	if err != nil {
		return nil, err
	}
	selections := make([]promotion.Selection, 0, len(rows))
	for _, row := range rows {
		selections = append(selections, promotion.Selection{
			PromotionId: stringOf(row, models.SalesOrderPromotionSelectionFieldPromotionId),
			RewardId:    stringOf(row, models.SalesOrderPromotionSelectionFieldRewardId),
			Declined:    stringOf(row, models.SalesOrderPromotionSelectionFieldDecision) == PromotionDecisionDeclined,
		})
	}
	return selections, nil
}

// The decision values of sales_order_promotion_selection.
const (
	PromotionDecisionAccepted = "accepted"
	PromotionDecisionDeclined = "declined"
)
