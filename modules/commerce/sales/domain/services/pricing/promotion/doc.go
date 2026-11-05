// Package promotion is the pure Sales promotion engine (BR_Sales_Promotion_Voucher, plan §3.1). It
// performs no I/O and imports no model, repository or the parent pricing package: the caller loads
// candidates and resolves products, categories, UoM and tax, and passes plain values in.
//
// # Calling sequence (reprice and confirm)
//
//  1. Build Input from the order AFTER pricelist and manual discounts, BEFORE promotions:
//     - Lines: every order line. Line.LineType decides what counts — product/combo lines with a
//     positive quantity are purchase lines; promotion_gift lines are recognised as existing gifts
//     (SourcePromotionId + SourceRewardId) and never qualify nor receive a discount.
//     - Line.UntaxedAmount is N_i, Line.TaxAmount is T_i from Accounting (first ResolveBasketTax).
//     - Line.CategoryIds is the variant's category plus all ancestors.
//     - Line.BaseQuantity is the quantity in the product base UoM; Line.QualificationQuantities maps
//     each BXGY rule's qualification_uom_id to the line quantity in it (omit when not convertible).
//     - Line.PricelistId is the pricelist that priced the line (D2).
//     - Programs: candidates already filtered by BR-C01/BR-S01 (active, window, channel AND point,
//     usage). Voucher-gated types carry the attached VoucherId (and VoucherSourceOrderId).
//     - Selections: persisted draft choices (D6).
//  2. result := Apply(input). Read result.Programs (one per candidate, in (sequence, id) order),
//     result.Lines (per-line untaxed promotion discount), result.Gifts (upsert intents; an empty
//     ExistingLineKey means create) and result.RemovedGiftLineKeys.
//  3. Resolve tax again on the discounted lines and call AttributeTax(&result, before, after, scale)
//     to fill tax_reduction_amount per allocation and per program.
//  4. reference := CalculationReference(ReferenceInput{...}) with the order's grand total; store it
//     on reprice, recompute and compare on confirm (D4, BR-T02).
//  5. On confirm only: EvaluateIssues(input) for next_order_coupon issuance, and IssueValidity for
//     the voucher's valid_from/expires_at (BR-V09, BR-V10).
//
// # Outcomes
//
// applied (has allocations and/or a gift; becomes an application on confirm), selection_required
// (BR-R12; never blocks confirm), declined, not_eligible (with a reason code: voucher_required,
// voucher_source_order, currency_mismatch, no_rule, no_eligible_line, pricelist_mismatch,
// condition_not_met) and not_applicable (unknown_program_type, no_reward, no_shipping_line,
// zero_benefit, order_discount_not_best). Only applied programs consume usage or vouchers.
//
// # Extension
//
// A new program_type is a ProgramEvaluator (and optionally an IssueEvaluator) registered on a
// Registry; Apply and EvaluateIssues need no edit.
//
// The engine has no point, credit or loyalty concept by design (AC-01): Buy X Get Y counts sets in
// one order only and next-order coupons are concrete vouchers.
package promotion
