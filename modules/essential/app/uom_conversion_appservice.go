package app

import (
	"github.com/shopspring/decimal"
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itUom "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/uom"
)

const (
	fieldQuantity    = "quantity"
	fieldSourceUomId = "source_uom_id"
	fieldTargetUomId = "target_uom_id"
)

func NewUomConversionApplicationServiceImpl(
	conversionSvc itUom.UomConversionDomainService, uomApp itUom.UomApplicationService,
) itUom.UomConversionApplicationService {
	return &UomConversionApplicationServiceImpl{conversionSvc: conversionSvc, uomApp: uomApp}
}

// UomConversionApplicationServiceImpl is the capability boundary other modules bind to, plus the
// two REST actions on essential_uom.
//
// The cross-module methods stay a thin delegation and assert nothing: the caller is another module
// acting on its own authority. The actions authorize like every UoM read: the caller must hold
// read on essential_uom in the org the source unit belongs to.
type UomConversionApplicationServiceImpl struct {
	conversionSvc itUom.UomConversionDomainService
	uomApp        itUom.UomApplicationService
}

func (this *UomConversionApplicationServiceImpl) Convert(
	ctx corectx.Context, query itUom.ConvertQuantityQuery,
) (*itUom.ConvertQuantityResult, error) {
	return this.conversionSvc.Convert(ctx, query)
}

func (this *UomConversionApplicationServiceImpl) GetUom(
	ctx corectx.Context, query itUom.GetUomQuery,
) (*itUom.GetUomResult, error) {
	return this.conversionSvc.GetUom(ctx, query)
}

func (this *UomConversionApplicationServiceImpl) ToReference(
	ctx corectx.Context, query itUom.ToReferenceQuery,
) (*itUom.ToReferenceResult, error) {
	return this.conversionSvc.ToReference(ctx, query)
}

// ConvertAction serves POST essential_uom/convert {quantity, source_uom_id, target_uom_id}.
func (this *UomConversionApplicationServiceImpl) ConvertAction(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*itUom.ConvertQuantityResult, error) {
	quantity, cErrs := readQuantity(params)
	if cErrs != nil {
		return &itUom.ConvertQuantityResult{ClientErrors: *cErrs}, nil
	}
	cErrs, err := this.assertSourceReadable(ctx, params)
	if err != nil || cErrs != nil {
		return &itUom.ConvertQuantityResult{ClientErrors: derefErrors(cErrs)}, err
	}
	return this.conversionSvc.Convert(ctx, itUom.ConvertQuantityQuery{
		Quantity:    quantity,
		SourceUomId: readId(params, fieldSourceUomId),
		TargetUomId: readId(params, fieldTargetUomId),
	})
}

// ToReferenceAction serves POST essential_uom/to_reference {quantity, source_uom_id}.
func (this *UomConversionApplicationServiceImpl) ToReferenceAction(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*itUom.ToReferenceResult, error) {
	quantity, cErrs := readQuantity(params)
	if cErrs != nil {
		return &itUom.ToReferenceResult{ClientErrors: *cErrs}, nil
	}
	cErrs, err := this.assertSourceReadable(ctx, params)
	if err != nil || cErrs != nil {
		return &itUom.ToReferenceResult{ClientErrors: derefErrors(cErrs)}, err
	}
	return this.conversionSvc.ToReference(ctx, itUom.ToReferenceQuery{
		Quantity:    quantity,
		SourceUomId: readId(params, fieldSourceUomId),
	})
}

// assertSourceReadable confines the action to the org of the source unit. A source that does not
// exist is left to the conversion service, which reports it under its own error key; there is no
// data behind it to protect.
func (this *UomConversionApplicationServiceImpl) assertSourceReadable(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*ft.ClientErrors, error) {
	sourceId := readId(params, fieldSourceUomId)
	if sourceId == "" {
		return nil, nil
	}
	found, err := this.uomApp.DomainService().Repository().FindByKeys(ctx, dmodel.DynamicFields{
		basemodel.FieldId: string(sourceId),
	})
	if err != nil {
		return nil, errors.Wrap(err, "assertSourceReadable")
	}
	if !found.HasData || found.Data == nil {
		return nil, nil
	}
	_, cErrs := this.uomApp.AssertAction(ctx, composable.PermissionRead, dmodel.DynamicFields{
		basemodel.FieldOrgId: found.Data[basemodel.FieldOrgId],
	})
	return cErrs, nil
}

// readQuantity accepts the quantity as a JSON string (the documented form, which keeps precision)
// or a number.
func readQuantity(params dmodel.DynamicFields) (decimal.Decimal, *ft.ClientErrors) {
	var quantity decimal.Decimal
	var err error
	switch raw := params[fieldQuantity].(type) {
	case string:
		quantity, err = decimal.NewFromString(raw)
	case float64:
		quantity = decimal.NewFromFloat(raw)
	case nil:
		return quantity, nil
	default:
		err = errors.Errorf("unsupported quantity type %T", raw)
	}
	if err != nil {
		cErrs := ft.ClientErrors{}
		cErrs.Append(*ft.NewValidationError(fieldQuantity, string(ft.ErrorKey("err_invalid_value")),
			"quantity must be a decimal number"))
		return quantity, &cErrs
	}
	return quantity, nil
}

func readId(params dmodel.DynamicFields, field string) model.Id {
	raw, _ := params[field].(string)
	return model.Id(raw)
}

func derefErrors(cErrs *ft.ClientErrors) ft.ClientErrors {
	if cErrs == nil {
		return ft.ClientErrors{}
	}
	return *cErrs
}
