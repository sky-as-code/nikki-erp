package app

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/services"
	itWarehouse "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/warehouse"
)

// provisionSystemLocations creates the root of a warehouse's tree and the stops its flows need.
// The root carries the warehouse code, so a path reads 'MAIN/Stock'. Every location created here is
// marked system-generated, which protects it from being restructured or archived while the
// warehouse still needs it.
func (this *WarehouseAppServiceImpl) provisionSystemLocations(
	ctx corectx.Context, warehouse models.Warehouse,
) error {
	warehouseId := derefString((*string)(warehouse.GetId()))
	orgId := derefString((*string)(warehouse.GetOrgId()))
	code := derefString(warehouse.GetCode())

	rootId, err := this.createSystemLocation(ctx, systemLocationRequest{
		WarehouseId: warehouseId,
		OrgId:       orgId,
		Code:        code,
		ParentId:    "",
		// The root is an organisational node, not somewhere goods sit: stock lives in the Stock
		// location beneath it.
		Usage:   models.InventoryLocationUsageVirtual,
		Purpose: "",
	})
	if err != nil {
		return err
	}

	for _, spec := range services.RequiredSystemLocations(
		derefString(warehouse.GetIncomingFlow()),
		derefString(warehouse.GetOutgoingFlow()),
	) {
		if _, err := this.createSystemLocation(ctx, systemLocationRequest{
			WarehouseId: warehouseId,
			OrgId:       orgId,
			Code:        spec.Code,
			ParentId:    rootId,
			Usage:       models.InventoryLocationUsageInternal,
			Purpose:     spec.Purpose,
		}); err != nil {
			return err
		}
	}
	return nil
}

// ensureFlowLocations creates whatever a newly chosen flow needs and does not already have. An
// existing but suspended location is resumed rather than duplicated, so switching a flow back and
// forth does not accumulate copies.
func (this *WarehouseAppServiceImpl) ensureFlowLocations(
	ctx corectx.Context, warehouse models.Warehouse, flow string, outgoing bool,
) error {
	warehouseId := derefString((*string)(warehouse.GetId()))
	rootId, err := this.findRootLocationId(ctx, warehouse)
	if err != nil {
		return err
	}

	specs := services.IncomingFlowLocations(flow)
	if outgoing {
		specs = services.OutgoingFlowLocations(flow)
	}

	for _, spec := range specs {
		existing, err := services.FindWarehouseLocationByCode(ctx, warehouseId, spec.Code)
		if err != nil {
			return err
		}
		if existing == nil {
			_, err := this.createSystemLocation(ctx, systemLocationRequest{
				WarehouseId: warehouseId,
				OrgId:       derefString((*string)(warehouse.GetOrgId())),
				Code:        spec.Code,
				ParentId:    rootId,
				Usage:       models.InventoryLocationUsageInternal,
				Purpose:     spec.Purpose,
			})
			if err != nil {
				return err
			}
			continue
		}

		if derefString(existing.GetStatus()) == models.InventoryLocationStatusSuspended {
			if _, err := this.locationSvc.Resume(ctx, derefString((*string)(existing.GetId()))); err != nil {
				return err
			}
		}
	}
	return nil
}

// suspendObsoleteLocations retires the stops a reduced flow no longer uses. Suspended, never
// deleted: past moves still name those locations and the history has to keep resolving. A location
// holding stock is suspended too, leaving the goods visible where they are.
func (this *WarehouseAppServiceImpl) suspendObsoleteLocations(
	ctx corectx.Context, warehouse models.Warehouse, previousFlow string, nextFlow string, outgoing bool,
) error {
	warehouseId := derefString((*string)(warehouse.GetId()))

	for _, spec := range services.ObsoleteSystemLocations(previousFlow, nextFlow, outgoing) {
		existing, err := services.FindWarehouseLocationByCode(ctx, warehouseId, spec.Code)
		if err != nil {
			return err
		}
		if existing == nil || derefString(existing.GetStatus()) == models.InventoryLocationStatusSuspended {
			continue
		}
		if err := services.WriteLocationStatus(ctx,
			derefString((*string)(existing.GetId())), models.InventoryLocationStatusSuspended); err != nil {
			return err
		}
	}
	return nil
}

// resolveLegLocations turns a plan's location codes into ids. Endpoints outside the warehouse stay
// unresolved: which vendor or customer location applies depends on the transaction, so the movement
// engine fills those in.
func (this *WarehouseAppServiceImpl) resolveLegLocations(
	ctx corectx.Context, warehouse models.Warehouse, legs []services.FlowLeg,
) ([]itWarehouse.ResolvedLeg, error) {
	warehouseId := derefString((*string)(warehouse.GetId()))
	resolved := make([]itWarehouse.ResolvedLeg, 0, len(legs))

	for _, leg := range legs {
		fromId, err := this.locationIdForCode(ctx, warehouseId, leg.FromCode)
		if err != nil {
			return nil, err
		}
		toId, err := this.locationIdForCode(ctx, warehouseId, leg.ToCode)
		if err != nil {
			return nil, err
		}
		resolved = append(resolved, itWarehouse.ResolvedLeg{
			FromLocationId: fromId,
			ToLocationId:   toId,
			FromCode:       leg.FromCode,
			ToCode:         leg.ToCode,
		})
	}
	return resolved, nil
}

func (this *WarehouseAppServiceImpl) locationIdForCode(
	ctx corectx.Context, warehouseId string, code string,
) (string, error) {
	if code == services.FlowEndpointVendor || code == services.FlowEndpointCustomer {
		return "", nil
	}
	found, err := services.FindWarehouseLocationByCode(ctx, warehouseId, code)
	if err != nil || found == nil {
		return "", err
	}
	return derefString((*string)(found.GetId())), nil
}

// findRootLocationId returns the warehouse's root location, the one named after its code.
func (this *WarehouseAppServiceImpl) findRootLocationId(
	ctx corectx.Context, warehouse models.Warehouse,
) (string, error) {
	warehouseId := derefString((*string)(warehouse.GetId()))
	root, err := services.FindWarehouseLocationByCode(ctx, warehouseId, derefString(warehouse.GetCode()))
	if err != nil {
		return "", err
	}
	if root == nil {
		return "", errors.New("warehouse '" + warehouseId + "' has no root location")
	}
	return derefString((*string)(root.GetId())), nil
}

type systemLocationRequest struct {
	WarehouseId string
	OrgId       string
	Code        string
	ParentId    string
	Usage       string
	Purpose     string
}

// createSystemLocation writes one location the warehouse owns. It goes through the location domain
// service so the tree rules and derived path apply, then sets is_system_generated directly, because
// the service strips that flag from anything a caller sends.
func (this *WarehouseAppServiceImpl) createSystemLocation(
	ctx corectx.Context, request systemLocationRequest,
) (string, error) {
	fields := dmodel.DynamicFields{
		models.InventoryLocationFieldCode:          request.Code,
		models.InventoryLocationFieldName:          langJsonFor(request.Code),
		models.InventoryLocationFieldLocationUsage: request.Usage,
		models.InventoryLocationFieldWarehouseId:   request.WarehouseId,
		models.InventoryLocationFieldOrgId:         request.OrgId,
		models.InventoryLocationFieldStatus:        models.InventoryLocationStatusActive,
	}
	if request.ParentId != "" {
		fields[models.InventoryLocationFieldParentLocationId] = request.ParentId
	}
	if request.Purpose != "" {
		fields[models.InventoryLocationFieldPurpose] = request.Purpose
	}

	result, err := this.locationSvc.Create(ctx, fields)
	if err != nil {
		return "", err
	}
	if result == nil || result.ClientErrors.Count() > 0 {
		return "", clientErrorSignal{errors: clientErrorsOf(result)}
	}

	locationId := readStringField(result.Data, models.InventoryLocationFieldId)
	return locationId, services.MarkLocationSystemGenerated(ctx, locationId)
}

// langJsonFor gives a system location a name in the shape the schema expects. The code doubles as
// the name: these are structural locations whose label is the same in every language.
func langJsonFor(code string) map[string]any {
	return map[string]any{"en-US": code}
}
