package services

import (
	"time"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/services/pricing"
)

// Applying a voucher to a draft order. This file owns the ORDER the gates run in, not the rules,
// which live and are tested elsewhere. Re-pricing is deliberately NOT here: this returns the
// accepted program set so the caller can apply several vouchers and price once.

type ApplyVoucherResult struct {
	Redemption *models.SalesVoucherRedemption

	ProgramId string

	// AcceptedProgramIds can be SHORTER than what the order had, so the caller re-prices against this
	// list rather than appending to its own.
	AcceptedProgramIds []string

	// DisplacedProgramIds lets the till say what was pushed out rather than silently dropping a
	// discount the customer could already see.
	DisplacedProgramIds []string
}

// ApplyVoucher validates a code against an order and reserves its use. A refusal is never an error
// return: it comes back as ClientErrors carrying a machine-readable reason.
func ApplyVoucher(
	ctx corectx.Context, params ApplyVoucherParams,
) (*ApplyVoucherResult, *ft.ClientErrors, error) {
	code, vErrs, err := resolveCodeByString(ctx, params.Code)
	if err != nil || vErrs != nil {
		return nil, vErrs, err
	}

	if vErrs := assertCodeUsable(code, params.NowUnix); vErrs != nil {
		return nil, vErrs, nil
	}

	programId := ""
	if id := code.GetProgramId(); id != nil {
		programId = string(*id)
	}
	program, vErrs, err := loadVoucherProgram(ctx, programId, params.NowUnix)
	if err != nil || vErrs != nil {
		return nil, vErrs, err
	}

	// Same evaluator as the automatic promotions, so a voucher and an automatic program cannot reach
	// different conclusions about the same basket.
	if vErrs := assertProgramEligible(ctx, programId, params); vErrs != nil {
		return nil, vErrs, nil
	}

	// Checked BEFORE reserving, so a refused voucher leaves no row behind.
	accepted, displaced, vErrs, err := resolveWithVoucher(ctx, program, params.AppliedProgramIds)
	if err != nil || vErrs != nil {
		return nil, vErrs, err
	}

	// Last, because it is the only step that writes. Its unique index is what actually stops a
	// double-spend, so nothing may come between it and the discount being honoured.
	redemption, vErrs, err := ReserveVoucher(
		ctx, string(*code.GetId()), params.SalesOrderId, params.OrgId, params.NowUnix)
	if err != nil || vErrs != nil {
		return nil, vErrs, err
	}

	return &ApplyVoucherResult{
		Redemption:          redemption,
		ProgramId:           programId,
		AcceptedProgramIds:  accepted,
		DisplacedProgramIds: displaced,
	}, nil, nil
}

type ApplyVoucherParams struct {
	Code string

	SalesOrderId string
	OrgId        string

	SalesChannelId string
	SalesPointId   string

	// AppliedProgramIds holds vouchers and automatic promotions alike: the compatibility rules do not
	// distinguish them.
	AppliedProgramIds []string

	// Facts come from the pricing result the caller already holds: this service does not price.
	Facts pricing.BasketFacts

	// NowUnix is passed in rather than read from a clock, so applying a voucher is reproducible
	// without freezing time globally.
	NowUnix int64
}

func resolveCodeByString(
	ctx corectx.Context, code string,
) (*models.SalesVoucherCode, *ft.ClientErrors, error) {
	notFound := func() *ft.ClientErrors {
		vErrs := ft.NewClientErrors()
		vErrs.Append(*ft.NewBusinessViolation("code", ReasonCodeNotFound,
			"no voucher exists with that code"))
		return vErrs
	}

	if code == "" {
		return nil, notFound(), nil
	}

	engine, err := engineFor(models.SalesVoucherCodeSchemaName)
	if err != nil {
		return nil, nil, err
	}
	found, err := engine.ResourceRepository().FindByKeys(ctx, dmodel.DynamicFields{
		models.SalesVoucherCodeFieldCode: code,
	})
	if err != nil {
		return nil, nil, err
	}
	if found == nil || !found.HasData {
		return nil, notFound(), nil
	}
	return models.NewSalesVoucherCodeFrom(found.Data), nil, nil
}

// loadVoucherProgram checks the program half. A code pointing at an AUTOMATIC program is a
// configuration error but is still reported as a customer refusal, with the code named so an
// operator can find it.
func loadVoucherProgram(
	ctx corectx.Context, programId string, nowUnix int64,
) (dmodel.DynamicFields, *ft.ClientErrors, error) {
	refuse := func(reason, message string) *ft.ClientErrors {
		vErrs := ft.NewClientErrors()
		vErrs.Append(*ft.NewBusinessViolation("code", reason, message))
		return vErrs
	}

	record, err := loadRecord(ctx,
		models.SalesPromotionProgramSchemaName, models.SalesPromotionProgramFieldId, programId)
	if err != nil {
		return nil, nil, err
	}
	if record == nil {
		return nil, refuse(ReasonProgramUnavailable,
			"the promotion this voucher activates no longer exists"), nil
	}

	if archived, ok := record[basemodel.FieldIsArchived].(bool); ok && archived {
		return nil, refuse(ReasonProgramUnavailable,
			"the promotion this voucher activates has been withdrawn"), nil
	}

	if activation := stringOf(record, models.SalesPromotionProgramFieldActivationType); activation !=
		string(models.PromotionActivationVoucherCode) {
		return nil, refuse(ReasonProgramUnavailable,
			"this code points at a program that does not accept codes"), nil
	}

	// The PROGRAM's window, separate from the code's: a campaign can end while codes issued under it
	// are still in date, and the campaign governs.
	if from := dateTimeOf(record, models.SalesPromotionProgramFieldValidFrom); from != nil &&
		nowUnix < from.GoTime().Unix() {
		return nil, refuse(ReasonNotYetValid, "this promotion has not started"), nil
	}
	if until := dateTimeOf(record, models.SalesPromotionProgramFieldValidUntil); until != nil &&
		nowUnix >= until.GoTime().Unix() {
		return nil, refuse(ReasonExpired, "this promotion has ended"), nil
	}

	return record, nil, nil
}

// The refusal reasons this operation adds to the ones in voucher_redemption.go.
const (
	ReasonProgramUnavailable = "sales_voucher.program_unavailable"
	ReasonChannelNotEligible = "sales_voucher.channel_not_eligible"
	ReasonConditionsNotMet   = "sales_voucher.conditions_not_met"
	ReasonIncompatible       = "sales_voucher.incompatible"
)

// Channel and sales point are conditions like any other rather than special-cased columns, so a
// program cannot be restricted to a channel two ways that eventually disagree.
func assertProgramEligible(
	ctx corectx.Context, programId string, params ApplyVoucherParams,
) *ft.ClientErrors {
	groups, err := loadConditionGroups(ctx, programId)
	if err != nil {
		// Fail closed: a condition that cannot be read is not a condition that passes, matching the
		// evaluator's own reading of an unknown condition type.
		vErrs := ft.NewClientErrors()
		vErrs.Append(*ft.NewBusinessViolation("code", ReasonConditionsNotMet,
			"this voucher's conditions could not be evaluated"))
		return vErrs
	}

	facts := params.Facts
	facts.SalesChannelId = params.SalesChannelId
	facts.SalesPointId = params.SalesPointId
	facts.NowUnix = params.NowUnix

	if pricing.IsEligible(groups, facts) {
		return nil
	}

	// One reason for every unmet condition: distinguishing them means re-evaluating each group to
	// find the first that failed, deliberately not done yet.
	vErrs := ft.NewClientErrors()
	vErrs.Append(*ft.NewBusinessViolation("code", ReasonConditionsNotMet,
		"this voucher does not apply to this order"))
	return vErrs
}

// resolveWithVoucher decides whether the new program may join the ones applied. The voucher may be
// refused by an incumbent, or DISPLACE one it outranks, hence a full accepted list not a boolean.
func resolveWithVoucher(
	ctx corectx.Context, program dmodel.DynamicFields, appliedProgramIds []string,
) (accepted []string, displaced []string, vErrs *ft.ClientErrors, err error) {
	programId := stringOf(program, models.SalesPromotionProgramFieldId)

	candidates := make([]CandidateProgram, 0, len(appliedProgramIds)+1)
	candidates = append(candidates, candidateFrom(program))

	for _, appliedId := range appliedProgramIds {
		if appliedId == programId {
			continue
		}
		record, loadErr := loadRecord(ctx,
			models.SalesPromotionProgramSchemaName, models.SalesPromotionProgramFieldId, appliedId)
		if loadErr != nil {
			return nil, nil, nil, loadErr
		}
		if record == nil {
			// An applied program that no longer exists cannot constrain anything.
			continue
		}
		candidates = append(candidates, candidateFrom(record))
	}

	rules, err := loadCompatibilityRules(ctx, candidates)
	if err != nil {
		return nil, nil, nil, err
	}

	survivors := ResolvePromotions(candidates, rules)
	acceptedIds := make([]string, 0, len(survivors))
	kept := make(map[string]bool, len(survivors))
	for _, survivor := range survivors {
		acceptedIds = append(acceptedIds, survivor.Id)
		kept[survivor.Id] = true
	}

	// The voucher itself losing is the refusal the customer sees.
	if !kept[programId] {
		vErrs := ft.NewClientErrors()
		vErrs.Append(*ft.NewBusinessViolation("code", ReasonIncompatible,
			"this voucher cannot be combined with a discount already on this order"))
		return nil, nil, vErrs, nil
	}

	for _, appliedId := range appliedProgramIds {
		if !kept[appliedId] {
			displaced = append(displaced, appliedId)
		}
	}
	return acceptedIds, displaced, nil, nil
}

func candidateFrom(record dmodel.DynamicFields) CandidateProgram {
	return CandidateProgram{
		Id:             stringOf(record, models.SalesPromotionProgramFieldId),
		Priority:       int32Of(record, models.SalesPromotionProgramFieldPriority),
		CreatedAt:      stringOf(record, basemodel.FieldCreatedAt),
		StackPolicy:    stringOf(record, models.SalesPromotionProgramFieldStackPolicy),
		ExclusiveGroup: stringOf(record, models.SalesPromotionProgramFieldExclusiveGroup),
	}
}

// loadCompatibilityRules loads every rule and filters here rather than querying per pair: the
// candidate set is small, so one read beats N-squared.
func loadCompatibilityRules(
	ctx corectx.Context, candidates []CandidateProgram,
) ([]CompatibilityRule, error) {
	if len(candidates) < 2 {
		// A single program has nothing to be incompatible with.
		return nil, nil
	}

	engine, err := engineFor(models.SalesPromotionCompatibilitySchemaName)
	if err != nil {
		return nil, err
	}

	found, err := engine.ResourceRepository().Search(ctx, dyn.RepoSearchParam{
		Page: 0,
		Size: model.MODEL_RULE_PAGE_MAX_SIZE,
	})
	if err != nil {
		return nil, err
	}
	if found == nil || !found.HasData {
		return nil, nil
	}

	inPlay := make(map[string]bool, len(candidates))
	for _, candidate := range candidates {
		inPlay[candidate.Id] = true
	}

	rules := make([]CompatibilityRule, 0, len(found.Data.Items))
	for _, item := range found.Data.Items {
		a := stringOf(item, models.SalesPromotionCompatibilityFieldProgramAId)
		b := stringOf(item, models.SalesPromotionCompatibilityFieldProgramBId)
		if !inPlay[a] || !inPlay[b] {
			continue
		}
		rules = append(rules, CompatibilityRule{
			ProgramAId: a,
			ProgramBId: b,
			// The column is an enum, not a boolean: `denied` wins, so anything not explicitly
			// `allowed` must not be read as permission.
			Allowed: stringOf(item, models.SalesPromotionCompatibilityFieldCompatibility) ==
				string(models.PromotionCompatibilityAllowed),
		})
	}
	return rules, nil
}

// SalesOrderForVoucher is read-only and narrow, so transport is never handed a mutable row it
// might write back.
type SalesOrderForVoucher struct {
	Id             string
	OrgId          string
	SalesChannelId string
	SalesPointId   string

	Subtotal      decimal.Decimal
	TotalQuantity decimal.Decimal

	AppliedProgramIds []string
}

// LoadSalesOrderForVoucher returns nil, not an error, when there is no such order: naming a
// missing record is a 400, not a 500.
func LoadSalesOrderForVoucher(
	ctx corectx.Context, orderId string,
) (*SalesOrderForVoucher, error) {
	record, err := loadRecord(ctx,
		models.SalesOrderSchemaName, models.SalesOrderFieldId, orderId)
	if err != nil || record == nil {
		return nil, err
	}

	order := &SalesOrderForVoucher{
		Id:             orderId,
		OrgId:          stringOf(record, basemodel.FieldOrgId),
		SalesChannelId: stringOf(record, models.SalesOrderFieldSalesChannelId),
		SalesPointId:   stringOf(record, models.SalesOrderFieldSalesPointId),
		Subtotal:       decimalOf(record, models.SalesOrderFieldSubtotal),
	}

	quantity, programIds, err := readOrderBasket(ctx, orderId)
	if err != nil {
		return nil, err
	}
	order.TotalQuantity = quantity
	order.AppliedProgramIds = programIds
	return order, nil
}

// readOrderBasket's voucher programs come from the REDEMPTIONS, not the adjustments: a voucher
// applied to a draft has reserved a use without yet producing an adjustment, so reading
// adjustments alone would let two mutually exclusive vouchers onto the same draft.
func readOrderBasket(
	ctx corectx.Context, orderId string,
) (decimal.Decimal, []string, error) {
	total := decimal.Zero

	lines, err := searchBy(ctx,
		models.SalesOrderLineSchemaName, models.SalesOrderLineFieldSalesOrderId, orderId)
	if err != nil {
		return total, nil, err
	}
	for _, line := range lines {
		// Giveaway lines are excluded: counting a free item toward a spend threshold would let one
		// promotion qualify the basket for another. Must match pricing.FactsFromLines.
		if stringOf(line, models.SalesOrderLineFieldLineType) ==
			string(models.SalesOrderLineTypePromotionReward) {
			continue
		}
		total = total.Add(decimalOf(line, models.SalesOrderLineFieldOrderedQuantity))
	}

	redemptions, err := searchBy(ctx,
		models.SalesVoucherRedemptionSchemaName,
		models.SalesVoucherRedemptionFieldSalesOrderId, orderId)
	if err != nil {
		return total, nil, err
	}

	seen := make(map[string]bool, len(redemptions))
	programIds := make([]string, 0, len(redemptions))
	for _, redemption := range redemptions {
		if !models.NewSalesVoucherRedemptionFrom(redemption).HoldsAUse() {
			continue
		}
		codeId := stringOf(redemption, models.SalesVoucherRedemptionFieldVoucherCodeId)
		codeRecord, err := loadRecord(ctx,
			models.SalesVoucherCodeSchemaName, models.SalesVoucherCodeFieldId, codeId)
		if err != nil {
			return total, nil, err
		}
		if codeRecord == nil {
			continue
		}
		programId := stringOf(codeRecord, models.SalesVoucherCodeFieldProgramId)
		if programId == "" || seen[programId] {
			continue
		}
		seen[programId] = true
		programIds = append(programIds, programId)
	}
	return total, programIds, nil
}

func OrderNotFoundErrors(orderId string) *ft.ClientErrors {
	vErrs := ft.NewClientErrors()
	vErrs.Append(*ft.NewBusinessViolation("id", "sales_order.not_found",
		"no sales order exists with id '"+orderId+"'"))
	return vErrs
}

// NowUnix reads the clock once per request at the edge; every rule beneath transport takes the
// instant as a parameter, so two gates in one request cannot disagree about the hour.
func NowUnix() int64 {
	return time.Now().Unix()
}
