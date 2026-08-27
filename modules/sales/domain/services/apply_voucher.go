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

// Applying a voucher to a draft order (BR 71).
//
// The operation is an orchestration rather than a rule: every rule it applies lives somewhere else
// and is tested there. What this file owns is the ORDER the gates run in, and the decision of what
// to do when one fails.
//
//	code exists, usable                -> voucher_redemption.go assertCodeUsable
//	program exists, is a voucher program, in date
//	channel and sales point eligible   -> eligibility conditions on the program
//	compatible with vouchers already applied -> ResolvePromotions (BR 29)
//	reserve the use                    -> voucher_redemption.go ReserveVoucher
//
// Re-pricing is deliberately NOT here. This returns the accepted program set and the caller re-runs
// the pricing engine with it. Keeping the two apart is what lets a caller apply several vouchers and
// price once, rather than pricing once per voucher and discarding all but the last answer.

// ApplyVoucherResult is what the caller needs to re-price and to answer the customer.
type ApplyVoucherResult struct {
	// Redemption is the reservation just taken. The caller stores nothing extra: the row IS the
	// record that this order holds a use.
	Redemption *models.SalesVoucherRedemption

	// ProgramId is the program the code activated, which is what the pricing engine applies.
	ProgramId string

	// AcceptedProgramIds is every program that survives conflict resolution once this one joins,
	// in application order. It can be SHORTER than what the order had before: an exclusive voucher
	// displaces the programs it cannot combine with, and the caller must re-price against this list
	// rather than appending to its own.
	AcceptedProgramIds []string

	// DisplacedProgramIds names what this voucher pushed out, so the till can say so rather than
	// silently dropping a discount the customer could already see.
	DisplacedProgramIds []string
}

// ApplyVoucher validates a code against an order and reserves its use.
//
// Returns ClientErrors for every refusal the customer could have caused, each carrying one of BR
// 71's machine-readable reasons. A refusal is never an error return: the customer typed a code that
// did not work, which is a 400 the till renders, not a fault.
//
// The gates run cheapest-first and stop at the first failure, so a customer is told the most specific
// true thing about why the code was refused rather than a list.
func ApplyVoucher(
	ctx corectx.Context, params ApplyVoucherParams,
) (*ApplyVoucherResult, *ft.ClientErrors, error) {
	code, vErrs, err := resolveCodeByString(ctx, params.Code)
	if err != nil || vErrs != nil {
		return nil, vErrs, err
	}

	// The code's own gates: status, archival, window, uses left.
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

	// Channel and sales point eligibility, plus every other condition the program declares. The
	// evaluator is the same one the automatic promotions use, so a voucher and an automatic program
	// cannot come to different conclusions about the same basket.
	if vErrs := assertProgramEligible(ctx, programId, params); vErrs != nil {
		return nil, vErrs, nil
	}

	// Compatibility with what is already on the order (BR 29). Checked BEFORE reserving, so a
	// refused voucher leaves no row behind.
	accepted, displaced, vErrs, err := resolveWithVoucher(ctx, program, params.AppliedProgramIds)
	if err != nil || vErrs != nil {
		return nil, vErrs, err
	}

	// Last, because it is the only step that writes. Its unique index is what actually stops a
	// double-spend (D-43), so nothing may come between it and the discount being honoured.
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

// ApplyVoucherParams is what applying a voucher needs to know.
type ApplyVoucherParams struct {
	// Code is what the customer typed, not an id. Resolving it here rather than in the caller is
	// what keeps "no such code" a single answer.
	Code string

	SalesOrderId string
	OrgId        string

	SalesChannelId string
	SalesPointId   string

	// AppliedProgramIds is what the order already has, vouchers and automatic promotions alike.
	// Both are passed because BR 29's compatibility rules do not distinguish them — a voucher can
	// be incompatible with an automatic promotion just as easily.
	AppliedProgramIds []string

	// Facts describes the basket, for the program's conditions. Supplied by the caller because this
	// service does not price: the facts come from the pricing result the caller already holds, via
	// pricing.FactsFromLines.
	Facts pricing.BasketFacts

	// NowUnix is the evaluation instant, passed in rather than read from a clock so that applying a
	// voucher is reproducible and testable without freezing time globally.
	NowUnix int64
}

// resolveCodeByString finds a code by the string a customer typed.
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

// loadVoucherProgram fetches the program a code points at, and checks the two things about it that
// are the program's business rather than the code's.
//
// A code pointing at an AUTOMATIC program is a configuration error, not a customer error, but it is
// still reported as a refusal: the customer cannot be told "your code is fine but misconfigured",
// and an operator finds it from the code, which the message names.
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

	// The PROGRAM's window, which is separate from the code's. A campaign can end while individual
	// codes issued under it are still notionally in date, and the campaign is what governs.
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

// assertProgramEligible runs the program's own conditions against the basket.
//
// Channel and sales point are conditions like any other rather than special-cased columns, which is
// why SALES-018 gave them condition types. Special-casing them here would mean a program could be
// restricted to a channel two different ways, and the two would eventually disagree.
func assertProgramEligible(
	ctx corectx.Context, programId string, params ApplyVoucherParams,
) *ft.ClientErrors {
	groups, err := loadConditionGroups(ctx, programId)
	if err != nil {
		// A condition that cannot be read is not a condition that passes. Refusing is the safe
		// direction: honouring a discount whose restrictions nobody could evaluate is the expensive
		// mistake, and the same fail-closed reading the evaluator itself takes for an unknown
		// condition type.
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

	// One reason for every unmet condition rather than naming which. BR 71 lists
	// CHANNEL_NOT_ELIGIBLE and MINIMUM_AMOUNT_NOT_MET separately, and distinguishing them means
	// re-evaluating each group to find the first that failed — worth doing when a till needs it,
	// and deliberately not done yet rather than guessed at.
	vErrs := ft.NewClientErrors()
	vErrs.Append(*ft.NewBusinessViolation("code", ReasonConditionsNotMet,
		"this voucher does not apply to this order"))
	return vErrs
}

// resolveWithVoucher asks BR 29 whether the new program may join the ones already applied.
//
// Both directions matter and both are the resolver's job: the voucher may be refused because an
// incumbent excludes it, or it may DISPLACE an incumbent because it has the better priority and
// excludes that one. The second case is why this returns a full accepted list rather than a boolean —
// the caller must re-price against what survived, not append to what it had.
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
			// An applied program that no longer exists cannot constrain anything. Skipping it is
			// right: it is already not being applied.
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

// candidateFrom reads the four fields conflict resolution sorts and compares on.
func candidateFrom(record dmodel.DynamicFields) CandidateProgram {
	return CandidateProgram{
		Id:             stringOf(record, models.SalesPromotionProgramFieldId),
		Priority:       int32Of(record, models.SalesPromotionProgramFieldPriority),
		CreatedAt:      stringOf(record, basemodel.FieldCreatedAt),
		StackPolicy:    stringOf(record, models.SalesPromotionProgramFieldStackPolicy),
		ExclusiveGroup: stringOf(record, models.SalesPromotionProgramFieldExclusiveGroup),
	}
}

// loadCompatibilityRules fetches the explicit pairwise directives among the candidates.
//
// Every rule touching any candidate is loaded and the resolver filters, rather than querying per
// pair: the candidate set is small, and one read is cheaper than N-squared.
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
			// The column is an enum, not a boolean: `denied` wins over everything (D-09), so
			// anything that is not explicitly `allowed` must not be read as permission.
			Allowed: stringOf(item, models.SalesPromotionCompatibilityFieldCompatibility) ==
				string(models.PromotionCompatibilityAllowed),
		})
	}
	return rules, nil
}

// SalesOrderForVoucher is the slice of an order the apply-voucher action needs.
//
// A narrow struct rather than the record, because the action is transport and should not be handed
// a mutable row it might write back. Everything here is read-only input to a decision.
type SalesOrderForVoucher struct {
	Id             string
	OrgId          string
	SalesChannelId string
	SalesPointId   string

	Subtotal      decimal.Decimal
	TotalQuantity decimal.Decimal

	// AppliedProgramIds is every program already on the order, derived from its reserved and
	// redeemed voucher redemptions plus the automatic programs its adjustments name.
	AppliedProgramIds []string
}

// LoadSalesOrderForVoucher reads what applying a voucher needs to know about an order.
//
// Returns nil when there is no such order, rather than an error: a caller naming a record that does
// not exist has made a mistake it can correct, and that is a 400, not a 500.
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

// readOrderBasket sums the order's quantities and collects the programs already applied to it.
//
// The programs come from the REDEMPTIONS rather than from the adjustments, for vouchers: an
// adjustment records what a program did to a number, and a voucher that was applied to a draft has
// reserved a use without yet producing one. Reading the adjustments alone would let a customer apply
// two mutually exclusive vouchers to the same draft, because neither had priced yet.
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
		// A giveaway line is excluded: counting a free item toward a spend threshold would let one
		// promotion qualify the basket for another it did not earn. Same rule as
		// pricing.FactsFromLines, and it must stay the same rule.
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

// OrderNotFoundErrors is the refusal for an order id that names nothing.
func OrderNotFoundErrors(orderId string) *ft.ClientErrors {
	vErrs := ft.NewClientErrors()
	vErrs.Append(*ft.NewBusinessViolation("id", "sales_order.not_found",
		"no sales order exists with id '"+orderId+"'"))
	return vErrs
}

// NowUnix is the current instant, for a caller at the edge that has no better source.
//
// It exists so the clock is read in exactly one place per request and then passed inward as data.
// Every rule beneath the transport layer takes the instant as a parameter, which is what makes them
// testable without freezing time and what stops two gates in one request disagreeing about the hour.
func NowUnix() int64 {
	return time.Now().Unix()
}
