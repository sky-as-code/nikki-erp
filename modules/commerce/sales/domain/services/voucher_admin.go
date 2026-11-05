package services

import (
	"strconv"
	"time"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"

	"github.com/sky-as-code/nikki-erp/modules/commerce/sales/domain/models"
	itPromotion "github.com/sky-as-code/nikki-erp/modules/commerce/sales/interfaces/promotion"
)

// Voucher administration (BR §9.2): batch generation and revocation.

const voucherBatchMax = 1000

var _ itPromotion.SalesVoucherDomainService = (*SalesVoucherDomainServiceImpl)(nil)

// The fields of a generate_batch body copied onto each voucher.
var voucherBatchFields = []string{
	basemodel.FieldOrgId, models.SalesVoucherFieldPromotionId, models.SalesVoucherFieldValidFrom,
	models.SalesVoucherFieldExpiresAt, models.SalesVoucherFieldCustomerId, models.SalesVoucherFieldUsageLimit,
}

// GenerateBatch creates up to 1000 vouchers in one transaction: all or none (BR-V06). Each is its
// own resource with a fresh random code; the type and limit follow the promotion. The codes are
// answered once, to the caller, and never logged.
func (this *SalesVoucherDomainServiceImpl) GenerateBatch(
	ctx corectx.Context, cmd itPromotion.GenerateVouchersCommand,
) (*itPromotion.GenerateVouchersResult, error) {
	quantity := int(int32Of(cmd, itPromotion.FieldVoucherQuantity))
	prefix := stringOf(cmd, itPromotion.FieldVoucherPrefix)
	if vErrs := assertBatchRequest(quantity, prefix); vErrs != nil {
		return &itPromotion.GenerateVouchersResult{ClientErrors: *vErrs}, nil
	}
	issued := make([]IssuedVoucher, 0, quantity)
	refused, err := dynamicresource.InRefusableTransaction(ctx, this.Repository(),
		func(tranxCtx corectx.Context) (*ft.ClientErrors, error) {
			for index := 0; index < quantity; index++ {
				code, err := freshVoucherCode(tranxCtx, stringOf(cmd, basemodel.FieldOrgId), prefix)
				if err != nil {
					return nil, err
				}
				row := pickFields(cmd, voucherBatchFields)
				row[models.SalesVoucherFieldCode] = code
				voucher, vErrs, err := createVoucher(tranxCtx, row)
				if err != nil || vErrs != nil {
					return vErrs, err
				}
				issued = append(issued, *voucher)
			}
			return nil, nil
		})
	if err != nil {
		return nil, err
	}
	if refused != nil {
		return &itPromotion.GenerateVouchersResult{ClientErrors: *refused}, nil
	}
	return &itPromotion.GenerateVouchersResult{HasData: true, Data: map[string]any{
		"count": len(issued), "vouchers": issued,
	}}, nil
}

func assertBatchRequest(quantity int, prefix string) *ft.ClientErrors {
	if quantity < 1 || quantity > voucherBatchMax {
		return violationsOf(itPromotion.FieldVoucherQuantity, "promotion.batch_quantity_invalid",
			"quantity must be between 1 and "+strconv.Itoa(voucherBatchMax))
	}
	if !ValidVoucherPrefix(prefix) {
		return violationsOf(itPromotion.FieldVoucherPrefix, "promotion.voucher_code_invalid",
			"a prefix is at most 24 of A-Z, 0-9 and '-'")
	}
	return nil
}

// Revoke withdraws a code for good (BR-L03); a revoked voucher answers as it is.
func (this *SalesVoucherDomainServiceImpl) Revoke(
	ctx corectx.Context, voucherId, note string,
) (*itPromotion.RecordResult, error) {
	return revokeVoucher(ctx, voucherId, models.VoucherRevocationManual, note, actorIdOf(ctx))
}

// RevokeVoucherBySystem revokes a voucher for a system reason, inside ctx's transaction: the
// cancel of a source order revokes its unused coupon with source_order_cancelled (BR-V12).
func RevokeVoucherBySystem(ctx corectx.Context, voucherId, reason string) (*ft.ClientErrors, error) {
	result, err := revokeVoucher(ctx, voucherId, reason, "", nil)
	if err != nil || result.ClientErrors.Count() == 0 {
		return nil, err
	}
	return &result.ClientErrors, nil
}

func revokeVoucher(
	ctx corectx.Context, voucherId, reason, note string, actorId *string,
) (*itPromotion.RecordResult, error) {
	repo, err := repoFor(models.SalesVoucherSchemaName)
	if err != nil {
		return nil, err
	}
	err = dynamicresource.InTransaction(ctx, repo, func(tranxCtx corectx.Context) error {
		stored, err := loadRecord(tranxCtx, models.SalesVoucherSchemaName, basemodel.FieldId, voucherId)
		if err != nil || stored == nil {
			return err
		}
		if stringOf(stored, models.SalesVoucherFieldStatus) == string(models.VoucherStatusRevoked) {
			return nil
		}
		changes := dmodel.DynamicFields{
			models.SalesVoucherFieldStatus:           string(models.VoucherStatusRevoked),
			models.SalesVoucherFieldRevokedAt:        model.ModelDateTime(time.Now().UTC()),
			models.SalesVoucherFieldRevocationReason: reason,
		}
		if actorId != nil {
			changes[models.SalesVoucherFieldRevokedBy] = *actorId
		}
		if note != "" {
			changes[models.SalesVoucherFieldRevocationNote] = note
		}
		return writeChanges(tranxCtx, models.SalesVoucherSchemaName, stored, changes)
	})
	if err != nil {
		return nil, err
	}
	return recordResult(ctx, models.SalesVoucherSchemaName, voucherId)
}

func actorIdOf(ctx corectx.Context) *string {
	principal := ctx.GetPermissions().Principal
	if principal.IsZero() {
		return nil
	}
	id := string(principal.Id)
	return &id
}
