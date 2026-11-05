package services

import (
	"regexp"
	"strings"
	"time"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"

	"github.com/sky-as-code/nikki-erp/modules/commerce/sales/domain/models"
)

// Vouchers (BR §6.1, §8.7). A voucher only unlocks its promotion. Its type follows the promotion's
// (BR-V02), its code is compared normalised (BR-V03), and only discount_code, coupon and
// next_order_coupon promotions take one (BR-V01). Codes never go into logs: log the voucher id.

// voucherCodePattern is the normalised form: 6-32 of A-Z, 0-9 and '-' (BR-V03).
var voucherCodePattern = regexp.MustCompile(`^[A-Z0-9-]{6,32}$`)

// NormalizeVoucherCode trims the ends and uppercases; inner spaces and look-alike characters are
// kept, so a code that differs from a valid one is refused rather than silently matched.
func NormalizeVoucherCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}

type SalesVoucherDomainServiceImpl struct {
	dynamicresource.CrudDomainService
}

func NewSalesVoucherDomainService(base dynamicresource.CrudDomainService) *SalesVoucherDomainServiceImpl {
	return &SalesVoucherDomainServiceImpl{CrudDomainService: base}
}

// voucherServerFields are never taken from the command; source_order_id is kept here because the
// next-order issuing flow sets it, and the application service drops a client's value instead.
var voucherServerFields = []string{
	models.SalesVoucherFieldNormalizedCode, models.SalesVoucherFieldVoucherType, models.SalesVoucherFieldStatus,
	models.SalesVoucherFieldUsedCount, models.SalesVoucherFieldRevokedAt, models.SalesVoucherFieldRevokedBy,
	models.SalesVoucherFieldRevocationReason, models.SalesVoucherFieldRevocationNote,
}

func (this *SalesVoucherDomainServiceImpl) Create(
	ctx corectx.Context, cmd dynamicresource.CreateCommand, options ...dynamicresource.CreateOptions,
) (*dynamicresource.CreateResult, error) {
	resetServerFields(cmd, voucherServerFields)
	vErrs, err := stampNewVoucher(ctx, cmd)
	if err != nil || vErrs != nil {
		return createRefusal(vErrs, err)
	}
	result, err := this.CrudDomainService.Create(ctx, cmd, options...)
	if err != nil && isUniqueViolation(err) {
		return createRefusal(violationsOf(models.SalesVoucherFieldCode, "promotion.voucher_code_taken",
			"this code is already used in the organization"), nil)
	}
	return result, err
}

// Update keeps the identity of a voucher: its code, type, promotion, customer and source never
// change, and a revoked voucher is not edited (BR-L03).
func (this *SalesVoucherDomainServiceImpl) Update(
	ctx corectx.Context, cmd dynamicresource.UpdateCommand, options ...dynamicresource.UpdateOptions,
) (*dynamicresource.MutateResult, error) {
	resetServerFields(cmd, append(voucherServerFields, models.SalesVoucherFieldSourceOrderId,
		models.SalesVoucherFieldIssuedAt))
	stored, err := loadRecord(ctx, models.SalesVoucherSchemaName, basemodel.FieldId, stringOf(cmd, basemodel.FieldId))
	if err != nil || stored == nil {
		return this.CrudDomainService.Update(ctx, cmd, options...)
	}
	if stringOf(stored, models.SalesVoucherFieldStatus) == string(models.VoucherStatusRevoked) {
		return mutateRefusal(violationsOf(models.SalesVoucherFieldStatus, "promotion.voucher_revoked",
			"a revoked voucher cannot be changed"), nil)
	}
	if models.VoucherType(stringOf(stored, models.SalesVoucherFieldVoucherType)) == models.VoucherTypeSingleUse {
		delete(cmd, models.SalesVoucherFieldUsageLimit)
	}
	return this.CrudDomainService.Update(ctx, cmd, options...)
}

// assertVoucherAllowed: only discount_code, coupon and next_order_coupon take vouchers (BR-V01),
// and a next-order coupon is issued only by its source order's confirmation (BR §9.2).
func assertVoucherAllowed(promotion, cmd dmodel.DynamicFields) *ft.ClientErrors {
	programType := models.PromotionProgramType(stringOf(promotion, models.SalesPromotionFieldProgramType))
	if !programType.TakesVouchers() {
		return violationsOf(models.SalesVoucherFieldPromotionId, "promotion.vouchers_not_allowed",
			"an automatic or buy X get Y promotion takes no voucher")
	}
	unsourced := stringOf(cmd, models.SalesVoucherFieldSourceOrderId) == ""
	if programType == models.PromotionTypeNextOrderCoupon && unsourced {
		return violationsOf(models.SalesVoucherFieldPromotionId, "promotion.vouchers_not_allowed",
			"a next-order coupon is issued only when its source order is confirmed")
	}
	return nil
}

// Delete hard-deletes only a voucher never attached to an order, never applied and not issued by a
// source order (BR-L04); anything else is archived or revoked.
func (this *SalesVoucherDomainServiceImpl) Delete(
	ctx corectx.Context, cmd dynamicresource.DeleteCommand, options ...dynamicresource.DeleteOptions,
) (*dynamicresource.MutateResult, error) {
	opts := firstDeleteOptions(options)
	opts.ValidateExtra = dynamicresource.ChainValidateExtra(assertVoucherDeletable, opts.ValidateExtra)
	return this.CrudDomainService.Delete(ctx, cmd, opts)
}

func assertVoucherDeletable(
	ctx corectx.Context, _ *dynamicresource.DynamicEntity, found *dynamicresource.DynamicEntity,
	vErrs *ft.ClientErrors,
) error {
	if found == nil {
		return nil
	}
	stored := found.GetFieldData()
	used := stringOf(stored, models.SalesVoucherFieldSourceOrderId) != ""
	for _, schemaName := range []string{
		models.SalesOrderVoucherSchemaName, models.SalesPromotionApplicationSchemaName, models.SalesVoucherIssuanceSchemaName,
	} {
		if used {
			break
		}
		rows, err := searchBy(ctx, schemaName, models.SalesOrderVoucherFieldVoucherId, stringOf(stored, basemodel.FieldId))
		if err != nil {
			return err
		}
		used = len(rows) > 0
	}
	if used {
		vErrs.Append(*ft.NewAnonymousBusinessViolation(ft.ErrorKey("promotion.voucher_not_deletable", "sales"),
			"the voucher was used or issued by an order; archive or revoke it instead"))
	}
	return nil
}

// stampNewVoucher normalises the code and derives the type and limit from the promotion.
func stampNewVoucher(ctx corectx.Context, cmd dmodel.DynamicFields) (*ft.ClientErrors, error) {
	normalized := NormalizeVoucherCode(stringOf(cmd, models.SalesVoucherFieldCode))
	if !voucherCodePattern.MatchString(normalized) {
		return violationsOf(models.SalesVoucherFieldCode, "promotion.voucher_code_invalid",
			"a code is 6 to 32 characters of A-Z, 0-9 and '-'"), nil
	}
	promotionId := stringOf(cmd, models.SalesVoucherFieldPromotionId)
	promotion, err := loadRecord(ctx, models.SalesPromotionSchemaName, basemodel.FieldId, promotionId)
	if err != nil {
		return nil, err
	}
	if promotion == nil || stringOf(promotion, basemodel.FieldOrgId) != stringOf(cmd, basemodel.FieldOrgId) {
		return violationsOf(models.SalesVoucherFieldPromotionId, "promotion.not_found",
			"no promotion with id '"+promotionId+"'"), nil
	}
	if vErrs := assertVoucherAllowed(promotion, cmd); vErrs != nil {
		return vErrs, nil
	}
	if taken, err := FindVoucherByCode(ctx, stringOf(cmd, basemodel.FieldOrgId), normalized); err != nil || taken != nil {
		if err != nil {
			return nil, err
		}
		return violationsOf(models.SalesVoucherFieldCode, "promotion.voucher_code_taken",
			"this code is already used in the organization"), nil
	}
	programType := models.PromotionProgramType(stringOf(promotion, models.SalesPromotionFieldProgramType))
	cmd[models.SalesVoucherFieldCode] = strings.TrimSpace(stringOf(cmd, models.SalesVoucherFieldCode))
	cmd[models.SalesVoucherFieldNormalizedCode] = normalized
	cmd[models.SalesVoucherFieldStatus] = string(models.VoucherStatusActive)
	cmd[models.SalesVoucherFieldUsedCount] = int32(0)
	cmd[models.SalesVoucherFieldVoucherType] = string(models.VoucherTypeReusable)
	if programType != models.PromotionTypeDiscountCode {
		cmd[models.SalesVoucherFieldVoucherType] = string(models.VoucherTypeSingleUse)
		cmd[models.SalesVoucherFieldUsageLimit] = int32(1)
	}
	if _, set := cmd[models.SalesVoucherFieldIssuedAt]; !set {
		cmd[models.SalesVoucherFieldIssuedAt] = model.ModelDateTime(time.Now().UTC())
	}
	return nil, nil
}
