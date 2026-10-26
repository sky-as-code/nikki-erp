package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
	itProduct "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/product"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// NewProductAttributeDomainService is handed the composable default by the product_attribute onion.
func NewProductAttributeDomainService(base composable.CrudDomainService) itProduct.ProductAttributeDomainService {
	return &ProductAttributeDomainServiceImpl{CrudDomainService: base}
}

// ProductAttributeDomainServiceImpl holds the attribute rules and saves the inline values delta
// in the attribute's own transaction.
type ProductAttributeDomainServiceImpl struct {
	composable.CrudDomainService
}

// Create writes the attribute and, when the body carries `values.add`, its first values in one
// transaction: a refused value rolls the attribute back.
func (this *ProductAttributeDomainServiceImpl) Create(
	ctx corectx.Context, cmd composable.CreateCommand, options ...composable.CreateOptions,
) (*composable.CreateResult, error) {
	delta, malformed := itProduct.LiftProductAttributeValuesDelta(cmd)
	if malformed {
		return &composable.CreateResult{ClientErrors: malformedValuesDelta()}, nil
	}
	opts := attrFirstOption(options)
	opts.ValidateExtra = composable.ChainValidateExtra(assertAttributeCreateAllowed, opts.ValidateExtra)
	if delta == nil || delta.IsEmpty() {
		return this.CrudDomainService.Create(ctx, cmd, opts)
	}
	if cErrs := validateValuesDeltaShape(*delta, true); cErrs.Count() > 0 {
		return &composable.CreateResult{ClientErrors: cErrs}, nil
	}
	if cErrs := validateValuesDeltaFinalState(nil, *delta); cErrs.Count() > 0 {
		return &composable.CreateResult{ClientErrors: cErrs}, nil
	}

	var result *composable.CreateResult
	refused, err := composable.InRefusableTransaction(ctx, this.Repository(),
		func(tranxCtx corectx.Context) (*ft.ClientErrors, error) {
			var err error
			if result, err = this.CrudDomainService.Create(tranxCtx, cmd, opts); err != nil {
				return nil, err
			}
			if result.ClientErrors.Count() > 0 || !result.HasData {
				return &result.ClientErrors, nil
			}
			parent := valuesParent{
				attributeId: composable.ParamString(result.Data, basemodel.FieldId),
				orgId:       composable.ParamString(result.Data, models.ProductAttributeFieldOrgId),
			}
			cErrs, err := applyValuesDelta(tranxCtx, parent, *delta)
			return &cErrs, err
		})
	if err != nil {
		return nil, errors.Wrap(err, "create product attribute")
	}
	if refused != nil {
		return &composable.CreateResult{ClientErrors: *refused}, nil
	}
	return result, nil
}

// Update applies the attribute's own columns and, when the body carries `values`, the values
// delta, in one transaction. The attribute row is written first so its etag guards the delta
// too, and a values-only save still moves the etag.
func (this *ProductAttributeDomainServiceImpl) Update(
	ctx corectx.Context, cmd composable.UpdateCommand, options ...composable.UpdateOptions,
) (*composable.MutateResult, error) {
	delta, malformed := itProduct.LiftProductAttributeValuesDelta(cmd)
	if malformed {
		return &composable.MutateResult{ClientErrors: malformedValuesDelta()}, nil
	}
	opts := attrFirstOption(options)
	opts.ValidateExtra = composable.ChainValidateExtra(assertAttributeUpdateAllowed, opts.ValidateExtra)
	if delta == nil || delta.IsEmpty() {
		return this.CrudDomainService.Update(ctx, cmd, opts)
	}
	if cErrs := validateValuesDeltaShape(*delta, false); cErrs.Count() > 0 {
		return &composable.MutateResult{ClientErrors: cErrs}, nil
	}

	var result *composable.MutateResult
	refused, err := composable.InRefusableTransaction(ctx, this.Repository(),
		func(tranxCtx corectx.Context) (*ft.ClientErrors, error) {
			var err error
			if result, err = this.CrudDomainService.Update(tranxCtx, cmd, opts); err != nil {
				return nil, err
			}
			if result.ClientErrors.Count() > 0 || !result.HasData {
				return &result.ClientErrors, nil
			}
			cErrs, err := this.applyUpdateDelta(tranxCtx, string(composable.ParamId(cmd)), *delta)
			return &cErrs, err
		})
	if err != nil {
		return nil, errors.Wrap(err, "update product attribute")
	}
	if refused != nil {
		return &composable.MutateResult{ClientErrors: *refused}, nil
	}
	return result, nil
}

func (this *ProductAttributeDomainServiceImpl) applyUpdateDelta(
	ctx corectx.Context, attributeId string, delta itProduct.ProductAttributeValuesDelta,
) (ft.ClientErrors, error) {
	existing, err := loadAttributeValueNames(ctx, attributeId)
	if err != nil {
		return nil, err
	}
	if cErrs := validateValuesDeltaFinalState(existing, delta); cErrs.Count() > 0 {
		return cErrs, nil
	}
	rows, err := searchAttributeRows(ctx, models.ProductAttributeSchemaName,
		attrAndGraph(attrCondition(basemodel.FieldId, dmodel.Equals, attributeId)),
		[]string{basemodel.FieldId, models.ProductAttributeFieldOrgId}, 1)
	if err != nil || len(rows) == 0 {
		return nil, errors.Wrap(err, "applyUpdateDelta: reload attribute")
	}
	parent := valuesParent{
		attributeId: attributeId,
		orgId:       composable.ParamString(rows[0], models.ProductAttributeFieldOrgId),
	}
	return applyValuesDelta(ctx, parent, delta)
}

// SetArchived refuses to archive an attribute a template still applies. Unarchiving touches
// nothing else: the values keep whatever archive state they had.
func (this *ProductAttributeDomainServiceImpl) SetArchived(
	ctx corectx.Context, cmd composable.SetArchivedCommand,
) (*composable.MutateResult, error) {
	if archive, _ := composable.ParamBool(cmd, basemodel.FieldIsArchived); archive {
		cErrs, err := assertAttributeNotInUse(ctx, string(composable.ParamId(cmd)))
		if err != nil {
			return nil, errors.Wrap(err, "archive product attribute")
		}
		if cErrs.Count() > 0 {
			return &composable.MutateResult{ClientErrors: cErrs}, nil
		}
	}
	return this.CrudDomainService.SetArchived(ctx, cmd)
}

// Delete refuses an attribute any template line ever referenced. Otherwise it deletes the
// attribute's values first, each through the value rules, and then the attribute, all in one
// transaction: a value that cannot be deleted refuses the whole delete.
func (this *ProductAttributeDomainServiceImpl) Delete(
	ctx corectx.Context, cmd composable.DeleteCommand, options ...composable.DeleteOptions,
) (*composable.MutateResult, error) {
	attributeId := string(composable.ParamId(cmd))
	var result *composable.MutateResult
	refused, err := composable.InRefusableTransaction(ctx, this.Repository(),
		func(tranxCtx corectx.Context) (*ft.ClientErrors, error) {
			cErrs, err := assertAttributeDeletable(tranxCtx, attributeId)
			if err != nil || cErrs.Count() > 0 {
				return &cErrs, err
			}
			if cErrs, err = deleteAttributeValues(tranxCtx, attributeId); err != nil || cErrs.Count() > 0 {
				return &cErrs, err
			}
			if result, err = this.CrudDomainService.Delete(tranxCtx, cmd, options...); err != nil {
				return nil, err
			}
			return &result.ClientErrors, nil
		})
	if err != nil {
		return nil, errors.Wrap(err, "delete product attribute")
	}
	if refused != nil {
		return &composable.MutateResult{ClientErrors: *refused}, nil
	}
	return result, nil
}

func assertAttributeDeletable(ctx corectx.Context, attributeId string) (ft.ClientErrors, error) {
	lineRepo, err := repoFor(models.ProductTemplateAttributeSchemaName)
	if err != nil {
		return nil, err
	}
	vErrs := ft.NewClientErrors()
	if err := AssertProductAttributeDeletable(ctx, lineRepo, attributeId, vErrs); err != nil {
		return nil, err
	}
	return *vErrs, nil
}

func deleteAttributeValues(ctx corectx.Context, attributeId string) (ft.ClientErrors, error) {
	if attributeId == "" {
		return nil, nil
	}
	values, err := loadAttributeValueNames(ctx, attributeId)
	if err != nil || len(values) == 0 {
		return nil, err
	}
	valueSvc, err := domainServiceFor(models.ProductAttributeValueSchemaName)
	if err != nil {
		return nil, err
	}
	for _, value := range values {
		res, err := valueSvc.Delete(ctx, dmodel.DynamicFields{basemodel.FieldId: string(value.id)})
		if err != nil {
			return nil, err
		}
		if res.ClientErrors.Count() > 0 {
			return ft.ClientErrors{*ft.NewBusinessViolation(itProduct.ProductAttributeFieldValues,
				ft.ErrorKey("product_attribute.value_not_deletable", "inventory"),
				"value {{id}} of this attribute cannot be deleted, so neither can the attribute; archive it instead",
				map[string]any{"id": string(value.id)})}, nil
		}
	}
	return nil, nil
}

func malformedValuesDelta() ft.ClientErrors {
	return ft.ClientErrors{*ft.NewValidationError(itProduct.ProductAttributeFieldValues,
		ft.ErrorKey("err_malformed_request"),
		"values must be {add: [{code, name, sequence, color_hex}], update: [{id, etag, ...}], remove: [ids]}")}
}
