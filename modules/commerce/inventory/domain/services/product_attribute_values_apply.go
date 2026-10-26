package services

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
	itProduct "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/product"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// valuesParent is the attribute a delta writes under. Its id and org are stamped on every added
// value; a row cannot choose them.
type valuesParent struct {
	attributeId string
	orgId       string
}

// applyValuesDelta expects the caller's transaction. It writes through the value domain service,
// so every value rule applies, and stops at the first refusal with the errors pointed at the row.
// Removals run first and adds last, so a delta can free a name and reuse it. A name swap between
// two updated rows is still refused row by row: rename through a third name, or in two saves.
func applyValuesDelta(
	ctx corectx.Context, parent valuesParent, delta itProduct.ProductAttributeValuesDelta,
) (ft.ClientErrors, error) {
	valueSvc, err := domainServiceFor(models.ProductAttributeValueSchemaName)
	if err != nil {
		return nil, err
	}
	for i, id := range delta.Remove {
		res, err := valueSvc.Delete(ctx, dmodel.DynamicFields{basemodel.FieldId: string(id)})
		if cErrs, stop, err := rowOutcome(res, err, valuesDeltaRowPath(valuesDeltaFieldRemove, i)); stop {
			return cErrs, err
		}
	}
	for i, row := range delta.Update {
		cmd := pickDeltaRow(row, basemodel.FieldId, basemodel.FieldEtag)
		res, err := valueSvc.Update(ctx, cmd)
		if cErrs, stop, err := rowOutcome(res, err, valuesDeltaRowPath(valuesDeltaFieldUpdate, i)); stop {
			return cErrs, err
		}
	}
	for i, row := range delta.Add {
		cmd := pickDeltaRow(row)
		cmd[models.ProductAttributeValueFieldAttributeId] = parent.attributeId
		cmd[models.ProductAttributeValueFieldOrgId] = parent.orgId
		res, err := valueSvc.Create(ctx, cmd)
		if err != nil {
			return nil, err
		}
		if res.ClientErrors.Count() > 0 {
			return prefixRowErrors(res.ClientErrors, valuesDeltaRowPath(valuesDeltaFieldAdd, i)), nil
		}
	}
	return nil, nil
}

// rowOutcome turns one update or delete result into the delta's verdict. A row that matched
// nothing is reported against the row: the id was checked to belong to the attribute, so it
// vanished under a concurrent writer.
func rowOutcome(
	res *composable.MutateResult, err error, path string,
) (ft.ClientErrors, bool, error) {
	if err != nil {
		return nil, true, err
	}
	if res.ClientErrors.Count() > 0 {
		return prefixRowErrors(res.ClientErrors, path), true, nil
	}
	if !res.HasData {
		return ft.ClientErrors{*ft.NewBusinessViolation(path, ft.ErrorKey("err_not_found"),
			"the value no longer exists")}, true, nil
	}
	return nil, false, nil
}
