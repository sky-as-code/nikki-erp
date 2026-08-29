// Package app holds the Inventory operations that span more than one domain service. Most
// resources need no application layer because the dynamic resource engine handles them. Warehouse
// creation and flow reconfiguration each write a warehouse and its locations, and must leave
// neither half applied, so they are orchestrated here above the two domain services.
package app

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/services"
	itWarehouse "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/warehouse"
)

func NewWarehouseAppService(
	warehouseSvc *services.WarehouseDomainServiceImpl,
	locationSvc *services.InventoryLocationDomainServiceImpl,
) itWarehouse.WarehouseAppService {
	return &WarehouseAppServiceImpl{warehouseSvc: warehouseSvc, locationSvc: locationSvc}
}

type WarehouseAppServiceImpl struct {
	warehouseSvc *services.WarehouseDomainServiceImpl
	locationSvc  *services.InventoryLocationDomainServiceImpl
}

var _ itWarehouse.WarehouseAppService = (*WarehouseAppServiceImpl)(nil)

// CreateWarehouse creates a warehouse together with the locations it needs to function. Both
// happen in one transaction: if any required location fails, the warehouse goes with it.
func (this *WarehouseAppServiceImpl) CreateWarehouse(
	ctx corectx.Context, cmd itWarehouse.CreateWarehouseCommand,
) (*itWarehouse.CreateWarehouseResult, error) {
	if vErrs := assertFlowsKnown(cmd.Fields); vErrs.Count() > 0 {
		return &itWarehouse.CreateWarehouseResult{ClientErrors: *vErrs}, nil
	}

	var created dmodel.DynamicFields
	err := services.WithWarehouseTransaction(ctx, func(tranxCtx corectx.Context) error {
		result, err := this.warehouseSvc.Create(tranxCtx, cmd.Fields)
		if err != nil {
			return err
		}
		if result == nil || result.ClientErrors.Count() > 0 {
			return newClientErrorSignal(result)
		}
		created = result.Data

		warehouse := models.NewWarehouseFrom(created)
		return this.provisionSystemLocations(tranxCtx, *warehouse)
	})

	if signal, ok := asClientErrorSignal(err); ok {
		return &itWarehouse.CreateWarehouseResult{ClientErrors: signal.errors}, nil
	}
	if err != nil {
		return nil, err
	}
	return &itWarehouse.CreateWarehouseResult{
		Data:    itWarehouse.CreateWarehouseResultData{Warehouse: created},
		HasData: true,
	}, nil
}

// ConfigureIncomingFlow changes how many stops goods make on the way in. It creates no stock move:
// a flow applies only to transactions made from now on, and a receipt already under way keeps the
// shape it was created with.
func (this *WarehouseAppServiceImpl) ConfigureIncomingFlow(
	ctx corectx.Context, cmd itWarehouse.ConfigureFlowCommand,
) (*itWarehouse.ConfigureFlowResult, error) {
	return this.configureFlow(ctx, cmd, false)
}

// ConfigureOutgoingFlow changes how many stops goods make on the way out. Same rules as incoming.
func (this *WarehouseAppServiceImpl) ConfigureOutgoingFlow(
	ctx corectx.Context, cmd itWarehouse.ConfigureFlowCommand,
) (*itWarehouse.ConfigureFlowResult, error) {
	return this.configureFlow(ctx, cmd, true)
}

func (this *WarehouseAppServiceImpl) configureFlow(
	ctx corectx.Context, cmd itWarehouse.ConfigureFlowCommand, outgoing bool,
) (*itWarehouse.ConfigureFlowResult, error) {
	if !services.IsKnownFlow(cmd.Flow) {
		return flowViolation("warehouse.flow_invalid",
			"'"+cmd.Flow+"' is not a known flow"), nil
	}

	warehouse, vErrs, err := services.LoadWarehouse(ctx, cmd.WarehouseId)
	if err != nil {
		return nil, err
	}
	if vErrs.Count() > 0 {
		return &itWarehouse.ConfigureFlowResult{ClientErrors: *vErrs}, nil
	}

	field := models.WarehouseFieldIncomingFlow
	previous := derefString(warehouse.GetIncomingFlow())
	if outgoing {
		field = models.WarehouseFieldOutgoingFlow
		previous = derefString(warehouse.GetOutgoingFlow())
	}
	if previous == cmd.Flow {
		return &itWarehouse.ConfigureFlowResult{
			Data:    itWarehouse.ConfigureFlowResultData{AffectedCount: 0},
			HasData: true,
		}, nil
	}

	err = services.WithWarehouseTransaction(ctx, func(tranxCtx corectx.Context) error {
		if err := services.WriteWarehouseFlow(tranxCtx, cmd.WarehouseId, field, cmd.Flow); err != nil {
			return err
		}
		if err := this.ensureFlowLocations(tranxCtx, *warehouse, cmd.Flow, outgoing); err != nil {
			return err
		}
		return this.suspendObsoleteLocations(tranxCtx, *warehouse, previous, cmd.Flow, outgoing)
	})
	if err != nil {
		return nil, err
	}

	return &itWarehouse.ConfigureFlowResult{
		Data:    itWarehouse.ConfigureFlowResultData{AffectedCount: 1},
		HasData: true,
	}, nil
}

// ResolveIncomingFlow reports the ordered path goods take into a warehouse. A pure read for the
// Stock movement engine to plan against: it creates nothing and moves nothing.
func (this *WarehouseAppServiceImpl) ResolveIncomingFlow(
	ctx corectx.Context, query itWarehouse.ResolveFlowQuery,
) (*itWarehouse.ResolveFlowResult, error) {
	return this.resolveFlow(ctx, query, false)
}

func (this *WarehouseAppServiceImpl) ResolveOutgoingFlow(
	ctx corectx.Context, query itWarehouse.ResolveFlowQuery,
) (*itWarehouse.ResolveFlowResult, error) {
	return this.resolveFlow(ctx, query, true)
}

func (this *WarehouseAppServiceImpl) resolveFlow(
	ctx corectx.Context, query itWarehouse.ResolveFlowQuery, outgoing bool,
) (*itWarehouse.ResolveFlowResult, error) {
	warehouse, vErrs, err := services.LoadWarehouse(ctx, query.WarehouseId)
	if err != nil {
		return nil, err
	}
	if vErrs.Count() > 0 {
		return &itWarehouse.ResolveFlowResult{ClientErrors: *vErrs}, nil
	}

	legs := services.ResolveIncomingFlow(derefString(warehouse.GetIncomingFlow()))
	if outgoing {
		legs = services.ResolveOutgoingFlow(derefString(warehouse.GetOutgoingFlow()))
	}

	resolved, err := this.resolveLegLocations(ctx, *warehouse, legs)
	if err != nil {
		return nil, err
	}
	return &itWarehouse.ResolveFlowResult{
		Data:    itWarehouse.ResolveFlowResultData{Legs: resolved},
		HasData: true,
	}, nil
}

func assertFlowsKnown(fields dmodel.DynamicFields) *ft.ClientErrors {
	vErrs := ft.NewClientErrors()
	for _, field := range []string{
		models.WarehouseFieldIncomingFlow,
		models.WarehouseFieldOutgoingFlow,
	} {
		flow := readStringField(fields, field)
		if flow != "" && !services.IsKnownFlow(flow) {
			vErrs.Append(*ft.NewBusinessViolation(models.WarehouseSchemaName,
				"warehouse.flow_invalid", "'"+flow+"' is not a known flow"))
		}
	}
	return vErrs
}

func flowViolation(key string, message string) *itWarehouse.ConfigureFlowResult {
	vErrs := ft.NewClientErrors()
	vErrs.Append(*ft.NewBusinessViolation(models.WarehouseSchemaName, key, message))
	return &itWarehouse.ConfigureFlowResult{ClientErrors: *vErrs}
}

// clientErrorSignal carries a rule violation out through the transaction body so a rejected
// operation rolls back rather than committing half of itself. The caller unwraps it back into
// client errors, which are a 400 rather than a 500.
type clientErrorSignal struct {
	errors ft.ClientErrors
}

func (this clientErrorSignal) Error() string {
	return "the operation was rejected by a business rule"
}

func newClientErrorSignal(result *dyn.OpResult[dmodel.DynamicFields]) error {
	if result == nil {
		return errors.New("the resource service returned no result")
	}
	return clientErrorSignal{errors: result.ClientErrors}
}

func asClientErrorSignal(err error) (clientErrorSignal, bool) {
	signal, ok := err.(clientErrorSignal)
	return signal, ok
}

// clientErrorsOf reads the violations off a result, tolerating a nil one.
func clientErrorsOf(result *dyn.OpResult[dmodel.DynamicFields]) ft.ClientErrors {
	if result == nil {
		return *ft.NewClientErrors()
	}
	return result.ClientErrors
}

func readStringField(fields dmodel.DynamicFields, key string) string {
	val := fields.GetString(key)
	if val == nil {
		return ""
	}
	return *val
}

func derefString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}
