package services

import (
	"time"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"

	"github.com/sky-as-code/nikki-erp/modules/commerce/sales/domain/models"
	itExt "github.com/sky-as-code/nikki-erp/modules/commerce/sales/interfaces/external"
)

// The draft order's promotion actions (plan §3.2). Each changes what the draft holds — an attached
// voucher, a reward choice — then reprices and answers the PromotionQuote. None consumes usage or
// holds a voucher (BR-V07): that happens only inside confirm.

// PromotionActionDeps are the ports a reprice needs.
type PromotionActionDeps struct {
	TaxSvc   itExt.TaxCalculationExtService
	Policy   SalesPolicy
	BasisSvc itExt.ProductPricingBasisExtService
}

// The client error keys of the promotion actions, under the "sales" namespace.
const (
	PromotionErrInvalidVoucher         = "promotion.invalid_voucher"
	PromotionErrCustomerMismatch       = "promotion.voucher_customer_mismatch"
	PromotionErrContextRequired        = "promotion.context_required"
	PromotionErrScopeMismatch          = "promotion.scope_mismatch"
	PromotionErrNotAvailable           = "promotion.not_available"
	PromotionErrVoucherUsageExhausted  = "promotion.voucher_usage_exhausted"
	PromotionErrUsageExhausted         = "promotion.usage_exhausted"
	PromotionErrSelectionRequired      = "promotion.reward_selection_required"
	PromotionErrRequoteRequired        = "promotion.requote_required"
	PromotionErrVoucherReplaceRequired = "promotion.voucher_replace_required"
	PromotionErrSourceOrderVoucherUsed = "promotion.source_order_voucher_used"
)

func promotionRefusal(field, key, message string, vars ...map[string]any) *ft.ClientErrors {
	vErrs := ft.NewClientErrors()
	if field == "" {
		vErrs.Append(*ft.NewAnonymousBusinessViolation(ft.ErrorKey(key, "sales"), message, vars...))
		return vErrs
	}
	vErrs.Append(*ft.NewBusinessViolation(field, ft.ErrorKey(key, "sales"), message, vars...))
	return vErrs
}

// PreviewPromotions reprices the draft and answers its quote.
func PreviewPromotions(
	ctx corectx.Context, orderId string, deps PromotionActionDeps,
) (map[string]any, *ft.ClientErrors, error) {
	priced, vErrs, err := RepriceOrder(ctx, orderId, deps.TaxSvc, deps.Policy, deps.BasisSvc)
	if err != nil || vErrs != nil {
		return nil, vErrs, err
	}
	return priced.Promotion.View(), nil, nil
}

// loadDraftOrder reads an order a promotion action may change: it must exist and still be a draft.
func loadDraftOrder(ctx corectx.Context, orderId string) (dmodel.DynamicFields, *ft.ClientErrors, error) {
	order, err := loadRecord(ctx, models.SalesOrderSchemaName, models.SalesOrderFieldId, orderId)
	if err != nil {
		return nil, nil, err
	}
	if order == nil {
		return nil, OrderNotFoundErrors(orderId), nil
	}
	if !models.NewSalesOrderFrom(order).IsEditable() {
		return nil, promotionRefusal("status", "sales_order.not_editable",
			"only a draft sales order may change its promotions"), nil
	}
	return order, nil, nil
}

// ApplyOrderVoucher attaches a voucher to a draft (BR-V07, BR-R13) and answers the fresh quote. An
// unknown, other-org and archived code all answer the same invalid_voucher (AC-21).
func ApplyOrderVoucher(
	ctx corectx.Context, orderId, code string, replaceExisting bool, deps PromotionActionDeps,
) (map[string]any, *ft.ClientErrors, error) {
	order, vErrs, err := loadDraftOrder(ctx, orderId)
	if err != nil || vErrs != nil {
		return nil, vErrs, err
	}
	voucher, promotionRecord, vErrs, err := resolveVoucherForOrder(ctx, order, code)
	if err != nil || vErrs != nil {
		return nil, vErrs, err
	}
	if vErrs, err := attachVoucher(ctx, order, voucher, promotionRecord, replaceExisting); err != nil || vErrs != nil {
		return nil, vErrs, err
	}
	return PreviewPromotions(ctx, orderId, deps)
}

func resolveVoucherForOrder(
	ctx corectx.Context, order dmodel.DynamicFields, code string,
) (dmodel.DynamicFields, dmodel.DynamicFields, *ft.ClientErrors, error) {
	invalid := promotionRefusal(paramVoucherField, PromotionErrInvalidVoucher, "this code is not valid")
	normalized := NormalizeVoucherCode(code)
	if !voucherCodePattern.MatchString(normalized) {
		return nil, nil, invalid, nil
	}
	voucher, err := FindVoucherByCode(ctx, stringOf(order, basemodel.FieldOrgId), normalized)
	if err != nil || voucher == nil || boolOf(voucher, basemodel.FieldIsArchived) {
		return nil, nil, invalidUnlessErr(invalid, err), err
	}
	promotionRecord, err := loadRecord(ctx, models.SalesPromotionSchemaName, basemodel.FieldId,
		stringOf(voucher, models.SalesVoucherFieldPromotionId))
	if err != nil || promotionRecord == nil || boolOf(promotionRecord, basemodel.FieldIsArchived) {
		return nil, nil, invalidUnlessErr(invalid, err), err
	}
	if vErrs, err := assertVoucherApplicable(ctx, order, voucher, promotionRecord); err != nil || vErrs != nil {
		return nil, nil, vErrs, err
	}
	return voucher, promotionRecord, nil, nil
}

func invalidUnlessErr(invalid *ft.ClientErrors, err error) *ft.ClientErrors {
	if err != nil {
		return nil
	}
	return invalid
}

// paramVoucherField is the input a voucher refusal names.
const paramVoucherField = "code"

// assertVoucherApplicable explains, to someone allowed to handle the order, why an in-org code
// cannot be used now (BR-V07). The engine still re-checks everything on every reprice.
func assertVoucherApplicable(
	ctx corectx.Context, order, voucher, promotionRecord dmodel.DynamicFields,
) (*ft.ClientErrors, error) {
	if vErrs := voucherStateRefusal(order, voucher, promotionRecord, time.Now().UTC()); vErrs != nil {
		return vErrs, nil
	}
	single := map[string]dmodel.DynamicFields{stringOf(promotionRecord, basemodel.FieldId): promotionRecord}
	inScope, _, err := promotionsInScope(ctx, order, single)
	if err != nil {
		return nil, err
	}
	if !inScope[stringOf(promotionRecord, basemodel.FieldId)] {
		return promotionRefusal(paramVoucherField, PromotionErrScopeMismatch,
			"this promotion does not run at this channel and point"), nil
	}
	return nil, nil
}

func voucherStateRefusal(order, voucher, promotionRecord dmodel.DynamicFields, now time.Time) *ft.ClientErrors {
	refuse := func(key, message string) *ft.ClientErrors {
		return promotionRefusal(paramVoucherField, key, message)
	}
	ownSource := stringOf(voucher, models.SalesVoucherFieldSourceOrderId) == stringOf(order, models.SalesOrderFieldId)
	running := stringOf(promotionRecord, models.SalesPromotionFieldStatus) == string(models.PromotionStatusActive) &&
		withinWindow(promotionRecord, models.SalesPromotionFieldStartsAt, models.SalesPromotionFieldEndsAt, now)
	switch {
	case stringOf(order, models.SalesOrderFieldSalesChannelId) == "" ||
		stringOf(order, models.SalesOrderFieldSalesPointId) == "":
		return promotionRefusal("", PromotionErrContextRequired, "the order needs a sales channel and a sales point")
	case !voucherCustomerMatches(voucher, stringOf(order, models.SalesOrderFieldSoldToPartyId)):
		return refuse(PromotionErrCustomerMismatch, "this code belongs to another customer")
	case voucherExhausted(voucher):
		return refuse(PromotionErrVoucherUsageExhausted, "this code has been used up")
	case !VoucherUsableAt(voucher, now) || ownSource || !running:
		return refuse(PromotionErrNotAvailable, "this code cannot be used now")
	case !promotionLiveAt(promotionRecord, now):
		return refuse(PromotionErrUsageExhausted, "this promotion has been used up")
	}
	return nil
}

func voucherExhausted(voucher dmodel.DynamicFields) bool {
	limit := int32Of(voucher, models.SalesVoucherFieldUsageLimit)
	return limit > 0 && int32Of(voucher, models.SalesVoucherFieldUsedCount) >= limit
}

// attachVoucher links the voucher, one per promotion per order: a second code of the same promotion
// must replace the first, explicitly (BR-R13).
func attachVoucher(
	ctx corectx.Context, order, voucher, promotionRecord dmodel.DynamicFields, replaceExisting bool,
) (*ft.ClientErrors, error) {
	orderId, promotionId := stringOf(order, models.SalesOrderFieldId), stringOf(promotionRecord, basemodel.FieldId)
	voucherId := stringOf(voucher, basemodel.FieldId)
	existing, err := orderVoucherLinks(ctx, orderId, promotionId)
	if err != nil {
		return nil, err
	}
	for _, link := range existing {
		if stringOf(link, models.SalesOrderVoucherFieldVoucherId) == voucherId {
			return nil, nil
		}
	}
	if len(existing) > 0 && !replaceExisting {
		return promotionRefusal(paramVoucherField, PromotionErrVoucherReplaceRequired,
			"another code of this promotion is already on the order; replace it explicitly",
			map[string]any{"promotion_id": promotionId}), nil
	}
	return nil, withTransaction(ctx, models.SalesOrderVoucherSchemaName, func(tranxCtx corectx.Context) error {
		if err := deleteRows(tranxCtx, models.SalesOrderVoucherSchemaName, existing); err != nil {
			return err
		}
		return insertOrderVoucher(tranxCtx, order, voucherId, promotionId)
	})
}

func orderVoucherLinks(ctx corectx.Context, orderId, promotionId string) ([]dmodel.DynamicFields, error) {
	return searchRows(ctx, models.SalesOrderVoucherSchemaName, false,
		equalsCondition(models.SalesOrderVoucherFieldOrderId, orderId),
		equalsCondition(models.SalesOrderVoucherFieldPromotionId, promotionId))
}

func insertOrderVoucher(ctx corectx.Context, order dmodel.DynamicFields, voucherId, promotionId string) error {
	repo, err := repoFor(models.SalesOrderVoucherSchemaName)
	if err != nil {
		return err
	}
	id, err := model.NewId()
	if err != nil {
		return err
	}
	fields := dmodel.DynamicFields{
		basemodel.FieldId:                        string(*id),
		basemodel.FieldOrgId:                     stringOf(order, basemodel.FieldOrgId),
		models.SalesOrderVoucherFieldOrderId:     stringOf(order, models.SalesOrderFieldId),
		models.SalesOrderVoucherFieldVoucherId:   voucherId,
		models.SalesOrderVoucherFieldPromotionId: promotionId,
		models.SalesOrderVoucherFieldAttachedAt:  model.ModelDateTime(time.Now().UTC()),
	}
	if actor := actorIdOf(ctx); actor != nil {
		fields[models.SalesOrderVoucherFieldAttachedBy] = *actor
	}
	_, err = repo.Insert(ctx, fields)
	return err
}

func deleteRows(ctx corectx.Context, schemaName string, rows []dmodel.DynamicFields) error {
	if len(rows) == 0 {
		return nil
	}
	repo, err := repoFor(schemaName)
	if err != nil {
		return err
	}
	for _, row := range rows {
		keys := dmodel.DynamicFields{basemodel.FieldId: stringOf(row, basemodel.FieldId)}
		if _, err := repo.DeleteOne(ctx, keys); err != nil {
			return err
		}
	}
	return nil
}

// RemoveOrderVoucher detaches a promotion's voucher; its benefit goes with it on the reprice.
func RemoveOrderVoucher(
	ctx corectx.Context, orderId, promotionId string, deps PromotionActionDeps,
) (map[string]any, *ft.ClientErrors, error) {
	if promotionId == "" {
		return nil, promotionRefusal("promotion_id", "promotion.promotion_required", "promotion_id is required"), nil
	}
	if _, vErrs, err := loadDraftOrder(ctx, orderId); err != nil || vErrs != nil {
		return nil, vErrs, err
	}
	links, err := orderVoucherLinks(ctx, orderId, promotionId)
	if err != nil {
		return nil, nil, err
	}
	err = withTransaction(ctx, models.SalesOrderVoucherSchemaName, func(tranxCtx corectx.Context) error {
		return deleteRows(tranxCtx, models.SalesOrderVoucherSchemaName, links)
	})
	if err != nil {
		return nil, nil, err
	}
	return PreviewPromotions(ctx, orderId, deps)
}
