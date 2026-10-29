package services

import (
	"fmt"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// paramIsArchived is the flag the set_archived action carries.
const paramIsArchived = basemodel.FieldIsArchived

// loadLocation fetches one location, reporting a missing id as a client error: the id came from the
// caller.
func (this *InventoryLocationDomainServiceImpl) loadLocation(
	ctx corectx.Context, locationId string,
) (*models.InventoryLocation, *ft.ClientErrors, error) {
	vErrs := ft.NewClientErrors()
	if locationId == "" {
		appendLocationViolation(vErrs, "inventory_location.not_found", "no location id was given")
		return nil, vErrs, nil
	}

	engine, err := repoFor(models.InventoryLocationSchemaName)
	if err != nil {
		return nil, vErrs, err
	}
	found, err := engine.FindByKeys(ctx, dmodel.DynamicFields{
		models.InventoryLocationFieldId: locationId,
	})
	if err != nil {
		return nil, vErrs, errors.Wrap(err, "loadLocation")
	}
	if found == nil || !found.HasData {
		appendLocationViolation(vErrs, "inventory_location.not_found",
			"no location with id '"+locationId+"'")
		return nil, vErrs, nil
	}
	return models.NewInventoryLocationFrom(found.Data), vErrs, nil
}

// writeStatus sets a location's operational state through the repository, not the service, so the
// caller's already-checked lifecycle rules are not re-run and the write cannot recurse back into
// the overridden Update.
func (this *InventoryLocationDomainServiceImpl) writeStatus(
	ctx corectx.Context, locationId string, status string,
) error {
	return this.writeLocationFields(ctx, locationId, dmodel.DynamicFields{
		models.InventoryLocationFieldStatus: status,
	})
}

// writeParent re-points a location at a new parent. An empty id makes it a root.
func (this *InventoryLocationDomainServiceImpl) writeParent(
	ctx corectx.Context, locationId string, newParentId string,
) error {
	var parent any
	if newParentId != "" {
		parent = newParentId
	}
	return this.writeLocationFields(ctx, locationId, dmodel.DynamicFields{
		models.InventoryLocationFieldParentLocationId: parent,
	})
}

func (this *InventoryLocationDomainServiceImpl) writeLocationFields(
	ctx corectx.Context, locationId string, fields dmodel.DynamicFields,
) error {
	engine, err := repoFor(models.InventoryLocationSchemaName)
	if err != nil {
		return err
	}

	update := dmodel.DynamicFields{models.InventoryLocationFieldId: locationId}
	for key, value := range fields {
		update[key] = value
	}
	_, err = engine.Update(ctx, update)
	return errors.Wrap(err, "writeLocationFields")
}

// withLocationTransaction runs body inside one transaction on a cloned context. The transaction
// must go on the clone, never ctx itself, or a committed transaction stays visible to whatever runs
// next. When ctx already carries a transaction - another module provisioning locations inside its
// own create - body joins it, so a later failure there rolls these writes back too.
func withLocationTransaction(ctx corectx.Context, body func(tranxCtx corectx.Context) error) error {
	if ctx.GetDbTranx() != nil {
		return body(ctx)
	}
	engine, err := repoFor(models.InventoryLocationSchemaName)
	if err != nil {
		return err
	}

	tranx, err := engine.BeginTransaction(ctx)
	if err != nil {
		return errors.Wrap(err, "withLocationTransaction")
	}
	defer tranx.Rollback()

	tranxCtx := corectx.CloneRequestContext(ctx)
	tranxCtx.SetDbTranx(tranx)

	if err := body(tranxCtx); err != nil {
		return err
	}
	return errors.Wrap(tranx.Commit(), "withLocationTransaction")
}

func appendLocationViolation(vErrs *ft.ClientErrors, key string, message string) {
	vErrs.Append(*ft.NewBusinessViolation(models.InventoryLocationSchemaName, ft.ErrorKey(key, "inventory"), message))
}

func locationViolationResult(key string, message string) *dyn.OpResult[dyn.MutateResultData] {
	vErrs := ft.NewClientErrors()
	appendLocationViolation(vErrs, key, message)
	return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}
}

func locationMutateOk() *dyn.OpResult[dyn.MutateResultData] {
	return &dyn.OpResult[dyn.MutateResultData]{
		Data:    dyn.MutateResultData{AffectedCount: 1},
		HasData: true,
	}
}

func copyFields(params dmodel.DynamicFields) dmodel.DynamicFields {
	copied := dmodel.DynamicFields{}
	for key, value := range params {
		copied[key] = value
	}
	return copied
}

func readBoolParam(params dmodel.DynamicFields, field string) bool {
	val := params.GetBool(field)
	return val != nil && *val
}

func derefInt(v *int32) int {
	if v == nil {
		return 0
	}
	return int(*v)
}

func derefId(v *model.Id) string {
	if v == nil {
		return ""
	}
	return string(*v)
}

// toComparableString renders a param value for the equality check protecting a system location's
// structural fields. Only strings and ids are compared, which is all those fields hold.
func toComparableString(value any) string {
	if value == nil {
		return ""
	}
	// model.Id is a string type, so the string cases cover it too.
	switch typed := value.(type) {
	case string:
		return typed
	case *string:
		if typed == nil {
			return ""
		}
		return *typed
	default:
		return fmt.Sprint(value)
	}
}

// currentFieldString reads the stored value of one of the protected structural fields.
func currentFieldString(location models.InventoryLocation, field string) string {
	return toComparableString(location.GetFieldData().GetAny(field))
}
