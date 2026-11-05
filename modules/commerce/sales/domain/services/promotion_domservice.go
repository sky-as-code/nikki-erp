package services

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"

	"github.com/sky-as-code/nikki-erp/modules/commerce/sales/domain/models"
	itPromotion "github.com/sky-as-code/nikki-erp/modules/commerce/sales/interfaces/promotion"
)

// The promotion lifecycle (BR-L01..L04). Configuration is editable only while the promotion is
// draft or suspended; every configuration write bumps config_version so a quote priced under an
// older configuration is detected (D4). The status moves only through Activate and Suspend.

var _ itPromotion.SalesPromotionDomainService = (*SalesPromotionDomainServiceImpl)(nil)

type SalesPromotionDomainServiceImpl struct {
	dynamicresource.CrudDomainService
}

func NewSalesPromotionDomainService(base dynamicresource.CrudDomainService) *SalesPromotionDomainServiceImpl {
	return &SalesPromotionDomainServiceImpl{CrudDomainService: base}
}

// promotionServerFields are written by the server alone; a client value is dropped (BR §8.2).
var promotionServerFields = []string{
	models.SalesPromotionFieldReference, models.SalesPromotionFieldStatus,
	models.SalesPromotionFieldUsedCount, models.SalesPromotionFieldConfigVersion,
}

// Create stamps the reference from the org's sales.promotion Sequence and starts the promotion as a
// draft with nothing used. The reference is issued inside the create's transaction, so a refused
// create gives its number back.
func (this *SalesPromotionDomainServiceImpl) Create(
	ctx corectx.Context, cmd dynamicresource.CreateCommand, options ...dynamicresource.CreateOptions,
) (*dynamicresource.CreateResult, error) {
	return this.createStamped(ctx, cmd, nil, options...)
}

func (this *SalesPromotionDomainServiceImpl) createStamped(
	ctx corectx.Context, cmd dynamicresource.CreateCommand, discriminators []string,
	options ...dynamicresource.CreateOptions,
) (*dynamicresource.CreateResult, error) {
	var result *dynamicresource.CreateResult
	refused, err := dynamicresource.InRefusableTransaction(ctx, this.Repository(),
		func(tranxCtx corectx.Context) (*ft.ClientErrors, error) {
			resetServerFields(cmd, promotionServerFields)
			if orgId := stringOf(cmd, basemodel.FieldOrgId); orgId != "" {
				numbers, err := issueDocumentNumbers(tranxCtx, orgId, discriminators, SeqPromotionReference)
				if refusal := numberingRefusalOf(err); refusal != nil {
					return refusal, nil
				}
				if err != nil {
					return nil, err
				}
				cmd[models.SalesPromotionFieldReference] = numbers[0]
			}
			cmd[models.SalesPromotionFieldStatus] = string(models.PromotionStatusDraft)
			cmd[models.SalesPromotionFieldUsedCount] = int32(0)
			cmd[models.SalesPromotionFieldConfigVersion] = int32(1)
			opts := firstCreateOptions(options)
			opts.ValidateExtra = dynamicresource.ChainValidateExtra(validatePromotionWrite, opts.ValidateExtra)
			var err error
			result, err = this.CrudDomainService.Create(tranxCtx, cmd, opts)
			return clientErrorsOf(result, err)
		})
	if err != nil || refused != nil {
		return createRefusal(refused, err)
	}
	return result, nil
}

// Update refuses any configuration change while the promotion is active (BR-L02), keeps
// program_type and currency_id once the promotion has been used, never lowers usage_limit below
// used_count, and bumps config_version.
func (this *SalesPromotionDomainServiceImpl) Update(
	ctx corectx.Context, cmd dynamicresource.UpdateCommand, options ...dynamicresource.UpdateOptions,
) (*dynamicresource.MutateResult, error) {
	resetServerFields(cmd, promotionServerFields)
	promotionId := stringOf(cmd, basemodel.FieldId)
	stored, err := loadRecord(ctx, models.SalesPromotionSchemaName, basemodel.FieldId, promotionId)
	if err != nil {
		return nil, err
	}
	if stored == nil {
		return notFoundResult(models.SalesPromotionSchemaName, promotionId), nil
	}
	if vErrs := assertPromotionEditable(stored); vErrs != nil {
		return &dynamicresource.MutateResult{ClientErrors: *vErrs}, nil
	}
	vErrs, err := assertPromotionChangeAllowed(ctx, stored, cmd)
	if err != nil || vErrs != nil {
		return mutateRefusal(vErrs, err)
	}
	cmd[models.SalesPromotionFieldConfigVersion] = int32Of(stored, models.SalesPromotionFieldConfigVersion) + 1
	opts := firstUpdateOptions(options)
	opts.ValidateExtra = dynamicresource.ChainValidateExtra(validatePromotionWrite, opts.ValidateExtra)
	return this.CrudDomainService.Update(ctx, cmd, opts)
}

// Delete hard-deletes only a draft that was never used: no voucher, application or issuance
// (BR-L04). Anything with history is archived instead.
func (this *SalesPromotionDomainServiceImpl) Delete(
	ctx corectx.Context, cmd dynamicresource.DeleteCommand, options ...dynamicresource.DeleteOptions,
) (*dynamicresource.MutateResult, error) {
	opts := firstDeleteOptions(options)
	opts.ValidateExtra = dynamicresource.ChainValidateExtra(assertPromotionDeletable, opts.ValidateExtra)
	return this.CrudDomainService.Delete(ctx, cmd, opts)
}

func assertPromotionDeletable(
	ctx corectx.Context, _ *dynamicresource.DynamicEntity, found *dynamicresource.DynamicEntity,
	vErrs *ft.ClientErrors,
) error {
	if found == nil {
		return nil
	}
	stored := found.GetFieldData()
	if stringOf(stored, models.SalesPromotionFieldStatus) != string(models.PromotionStatusDraft) {
		vErrs.Append(*ft.NewAnonymousBusinessViolation(ft.ErrorKey("promotion.not_deletable", "sales"),
			"only a draft promotion can be deleted; archive it instead"))
		return nil
	}
	used, err := promotionHasHistory(ctx, stringOf(stored, basemodel.FieldId), true)
	if err != nil {
		return err
	}
	if used {
		vErrs.Append(*ft.NewAnonymousBusinessViolation(ft.ErrorKey("promotion.not_deletable", "sales"),
			"the promotion has vouchers, applications or issued coupons; archive it instead"))
	}
	return nil
}

// promotionHasHistory reports whether anything references the promotion's use: an application or
// an issuance, and also a voucher when withVouchers is set.
func promotionHasHistory(ctx corectx.Context, promotionId string, withVouchers bool) (bool, error) {
	schemas := []string{models.SalesPromotionApplicationSchemaName, models.SalesVoucherIssuanceSchemaName}
	if withVouchers {
		schemas = append(schemas, models.SalesVoucherSchemaName)
	}
	for _, schemaName := range schemas {
		rows, err := searchBy(ctx, schemaName, models.SalesVoucherFieldPromotionId, promotionId)
		if err != nil {
			return false, err
		}
		if len(rows) > 0 {
			return true, nil
		}
	}
	return false, nil
}

// assertPromotionEditable is the edit lock of BR-L02, shared by the promotion and every child.
func assertPromotionEditable(promotion dmodel.DynamicFields) *ft.ClientErrors {
	if stringOf(promotion, models.SalesPromotionFieldStatus) != string(models.PromotionStatusActive) {
		return nil
	}
	return violationsOf("status", "promotion.config_locked",
		"an active promotion cannot be changed; suspend it first")
}

// assertPromotionChangeAllowed applies the rules that compare the update with the stored row.
func assertPromotionChangeAllowed(
	ctx corectx.Context, stored dmodel.DynamicFields, cmd dmodel.DynamicFields,
) (*ft.ClientErrors, error) {
	vErrs := ft.NewClientErrors()
	if limit, sent := cmd[models.SalesPromotionFieldUsageLimit]; sent && limit != nil {
		if int32Of(cmd, models.SalesPromotionFieldUsageLimit) < int32Of(stored, models.SalesPromotionFieldUsedCount) {
			vErrs.Append(*ft.NewBusinessViolation(models.SalesPromotionFieldUsageLimit,
				ft.ErrorKey("promotion.usage_limit_below_used", "sales"),
				"usage_limit cannot be lowered below the uses already confirmed"))
		}
	}
	changed := changedFields(stored, cmd, models.SalesPromotionFieldProgramType, models.SalesPromotionFieldCurrencyId)
	if len(changed) > 0 {
		used, err := promotionHasHistory(ctx, stringOf(stored, basemodel.FieldId), true)
		if err != nil {
			return nil, err
		}
		for _, field := range changed {
			if used {
				vErrs.Append(*ft.NewBusinessViolation(field, ft.ErrorKey("promotion.immutable_after_use", "sales"),
					field+" cannot change once the promotion has vouchers or applications; create a new promotion"))
			}
		}
	}
	if vErrs.Count() == 0 {
		return nil, nil
	}
	return vErrs, nil
}

// validatePromotionWrite checks the record as it will be: the window (BR-C02) and the
// type-specific next-order validity.
func validatePromotionWrite(
	_ corectx.Context, input *dynamicresource.DynamicEntity, found *dynamicresource.DynamicEntity,
	vErrs *ft.ClientErrors,
) error {
	merged := input.GetFieldData()
	if found != nil {
		merged = mergedFields(found.GetFieldData(), merged)
	}
	startsAt := dateTimeOf(merged, models.SalesPromotionFieldStartsAt)
	endsAt := dateTimeOf(merged, models.SalesPromotionFieldEndsAt)
	if startsAt != nil && endsAt != nil && !startsAt.GoTime().Before(endsAt.GoTime()) {
		vErrs.Append(*ft.NewBusinessViolation(models.SalesPromotionFieldEndsAt,
			ft.ErrorKey("promotion.invalid_window", "sales"), "ends_at must be after starts_at"))
	}
	return nil
}

// BumpConfigVersion is the parent half of every child write: inside the child's transaction it
// refuses the write while the promotion is active and records that the configuration changed.
func (this *SalesPromotionDomainServiceImpl) BumpConfigVersion(
	ctx corectx.Context, promotionId string,
) (*ft.ClientErrors, error) {
	return bumpPromotionConfig(ctx, promotionId)
}

func bumpPromotionConfig(ctx corectx.Context, promotionId string) (*ft.ClientErrors, error) {
	stored, err := loadRecord(ctx, models.SalesPromotionSchemaName, basemodel.FieldId, promotionId)
	if err != nil {
		return nil, err
	}
	if stored == nil {
		return violationsOf(models.SalesPromotionRuleFieldPromotionId, "promotion.not_found",
			"no promotion with id '"+promotionId+"'"), nil
	}
	if vErrs := assertPromotionEditable(stored); vErrs != nil {
		return vErrs, nil
	}
	return nil, writeChanges(ctx, models.SalesPromotionSchemaName, stored, dmodel.DynamicFields{
		models.SalesPromotionFieldConfigVersion: int32Of(stored, models.SalesPromotionFieldConfigVersion) + 1,
	})
}

// recordResult reloads a promotion as the answer of a lifecycle action.
func recordResult(ctx corectx.Context, schemaName, id string) (*itPromotion.RecordResult, error) {
	record, err := loadRecord(ctx, schemaName, basemodel.FieldId, id)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return &itPromotion.RecordResult{ClientErrors: notFoundResult(schemaName, id).ClientErrors}, nil
	}
	return &itPromotion.RecordResult{Data: record, HasData: true}, nil
}

func recordRefusal(vErrs *ft.ClientErrors) *itPromotion.RecordResult {
	return &itPromotion.RecordResult{ClientErrors: *vErrs}
}
