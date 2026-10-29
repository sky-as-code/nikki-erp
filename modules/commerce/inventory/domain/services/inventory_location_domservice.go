package services

import (
	"strings"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/safe"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
	itStock "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/stock"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// NewInventoryLocationDomainService derives the location service from the engine's default one.
// usage tells it what Stock holds at a location, injected rather than looked up so the archive
// guard can be tested without a stock engine.
func NewInventoryLocationDomainService(
	base composable.CrudDomainService, usage itStock.LocationUsageReadService,
) *InventoryLocationDomainServiceImpl {
	return &InventoryLocationDomainServiceImpl{CrudDomainService: base, usage: usage}
}

// InventoryLocationDomainServiceImpl adds the tree rules and lifecycle guards to the location
// resource: keeping the tree coherent, and stopping a location from being retired while something
// depends on it. None of it changes a quantity.
type InventoryLocationDomainServiceImpl struct {
	composable.CrudDomainService

	usage itStock.LocationUsageReadService
}

var _ composable.CrudDomainService = (*InventoryLocationDomainServiceImpl)(nil)

// Create is the client's create: it applies the tree rules and derives what the location inherits
// from its parent. is_system_generated and the managed fields are stripped from whatever the client
// sent: a client able to set them could mint a location that then refuses to be archived or deleted.
func (this *InventoryLocationDomainServiceImpl) Create(
	ctx corectx.Context, params dmodel.DynamicFields, options ...composable.CreateOptions,
) (*dyn.OpResult[dmodel.DynamicFields], error) {
	prepared := copyFields(params)
	delete(prepared, basemodel.FieldManagedByType)
	delete(prepared, basemodel.FieldManagedById)
	return this.create(ctx, prepared, options...)
}

// CreateManaged is the create of a location something manages - a warehouse's root, an org's
// virtual locations - and keeps the managed_by_type and managed_by_id the caller set. No REST
// route reaches it.
func (this *InventoryLocationDomainServiceImpl) CreateManaged(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[dmodel.DynamicFields], error) {
	return this.create(ctx, copyFields(params))
}

func (this *InventoryLocationDomainServiceImpl) create(
	ctx corectx.Context, prepared dmodel.DynamicFields, options ...composable.CreateOptions,
) (*dyn.OpResult[dmodel.DynamicFields], error) {
	vErrs := ft.NewClientErrors()
	rejectDerivedLocationFields(prepared, vErrs)
	if vErrs.Count() > 0 {
		return &dyn.OpResult[dmodel.DynamicFields]{ClientErrors: *vErrs}, nil
	}
	delete(prepared, models.InventoryLocationFieldIsSystemGenerated)
	delete(prepared, models.InventoryLocationFieldHierarchyDepth)

	vErrs, err := this.assertPlacementValid(ctx, prepared)
	if err != nil {
		return nil, err
	}
	if vErrs.Count() > 0 {
		return &dyn.OpResult[dmodel.DynamicFields]{ClientErrors: *vErrs}, nil
	}
	opts := safe.GetOptional(options, composable.CreateOptions{})
	opts.ValidateExtra = composable.ChainValidateExtra(opts.ValidateExtra, this.deriveOnCreate)
	return this.CrudDomainService.Create(ctx, prepared, opts)
}

func (this *InventoryLocationDomainServiceImpl) CreateBulk(
	ctx corectx.Context, cmds []composable.CreateCommand,
) (*dyn.OpResult[[]dmodel.DynamicFields], error) {
	return this.createBulk(ctx, cmds, func(
		tranxCtx corectx.Context, cmd composable.CreateCommand,
	) (*dyn.OpResult[dmodel.DynamicFields], error) {
		return this.Create(tranxCtx, cmd)
	})
}

// CreateManagedBulk is CreateBulk for locations something manages - a kiosk's slots - keeping the
// managed_by_type and managed_by_id each command carries. All or nothing, like CreateBulk.
func (this *InventoryLocationDomainServiceImpl) CreateManagedBulk(
	ctx corectx.Context, cmds []composable.CreateCommand,
) (*dyn.OpResult[[]dmodel.DynamicFields], error) {
	return this.createBulk(ctx, cmds, this.CreateManaged)
}

func (this *InventoryLocationDomainServiceImpl) createBulk(
	ctx corectx.Context, cmds []composable.CreateCommand,
	createOne func(corectx.Context, composable.CreateCommand) (*dyn.OpResult[dmodel.DynamicFields], error),
) (*dyn.OpResult[[]dmodel.DynamicFields], error) {
	created := make([]dmodel.DynamicFields, 0, len(cmds))
	var refused ft.ClientErrors

	err := withLocationTransaction(ctx, func(tranxCtx corectx.Context) error {
		for _, cmd := range cmds {
			res, err := createOne(tranxCtx, cmd)
			if err != nil {
				return err
			}
			if res.ClientErrors.Count() > 0 {
				refused = res.ClientErrors
				return errLocationBulkRefused
			}
			created = append(created, res.Data)
		}
		return nil
	})
	if errors.Is(err, errLocationBulkRefused) {
		return &dyn.OpResult[[]dmodel.DynamicFields]{ClientErrors: refused}, nil
	}
	if err != nil {
		return nil, err
	}

	return &dyn.OpResult[[]dmodel.DynamicFields]{Data: created, HasData: true}, nil
}

var errLocationBulkRefused = errors.New("inventory location bulk create refused; rolling back")

// Update applies the placement rules, refuses to restructure a system-generated location, and
// rewrites what the subtree inherits when the name or parent changes, in one transaction. A
// warehouse's root, Stock, Input or Output location may be renamed and given a storage category,
// but re-parenting it or changing its type would break the warehouse that created it.
func (this *InventoryLocationDomainServiceImpl) Update(
	ctx corectx.Context, params dmodel.DynamicFields, options ...composable.UpdateOptions,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	locationId := readStringParam(params, models.InventoryLocationFieldId)
	current, vErrs, err := this.loadLocation(ctx, locationId)
	if err != nil {
		return nil, err
	}
	if vErrs.Count() > 0 {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}, nil
	}

	prepared := copyFields(params)
	// Structural fields belong to the move and lifecycle operations, which validate what a plain
	// update cannot: the subtree, the stock behind it and the names to be rewritten.
	delete(prepared, models.InventoryLocationFieldIsSystemGenerated)
	delete(prepared, models.InventoryLocationFieldStatus)
	delete(prepared, models.InventoryLocationFieldHierarchyDepth)

	if vErrs, err := this.assertUpdateAllowed(ctx, prepared, *current); err != nil {
		return nil, err
	} else if vErrs.Count() > 0 {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}, nil
	}

	opts := safe.GetOptional(options, composable.UpdateOptions{})
	opts.ValidateExtra = composable.ChainValidateExtra(opts.ValidateExtra, this.deriveOnUpdate)
	if !changesLocationTree(prepared) {
		return this.CrudDomainService.Update(ctx, prepared, opts)
	}
	return this.updateAndCascade(ctx, prepared, opts)
}

func (this *InventoryLocationDomainServiceImpl) assertUpdateAllowed(
	ctx corectx.Context, prepared dmodel.DynamicFields, current models.InventoryLocation,
) (*ft.ClientErrors, error) {
	vErrs := ft.NewClientErrors()
	rejectDerivedLocationFields(prepared, vErrs)
	if current.IsWarehouseRoot() && vErrs.Count() == 0 {
		vErrs.ConcatPtr(assertWarehouseRootUnchanged(prepared, current))
	}
	if isSystemGenerated(current) && vErrs.Count() == 0 {
		vErrs.ConcatPtr(assertSystemLocationUnchanged(prepared, current))
	}
	if vErrs.Count() > 0 {
		return vErrs, nil
	}
	if _, moved := prepared[models.InventoryLocationFieldParentLocationId]; !moved {
		return vErrs, nil
	}
	newParentId := readStringParam(prepared, models.InventoryLocationFieldParentLocationId)
	if vErrs, err := this.assertMoveTargetValid(ctx, current, newParentId); err != nil || vErrs.Count() > 0 {
		return vErrs, err
	}
	return this.assertMoveKeepsReservationsBacked(ctx, current, newParentId)
}

func (this *InventoryLocationDomainServiceImpl) updateAndCascade(
	ctx corectx.Context, prepared dmodel.DynamicFields, opts composable.UpdateOptions,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	var result *dyn.OpResult[dyn.MutateResultData]
	err := runInTransaction(ctx, this.Repository(), func(tranxCtx corectx.Context) error {
		res, err := this.CrudDomainService.Update(tranxCtx, prepared, opts)
		if err != nil {
			return err
		}
		result = res
		if res.ClientErrors.Count() > 0 {
			return composable.ClientErrorSignal{Errors: res.ClientErrors}
		}
		if !res.HasData {
			return nil
		}
		return this.cascadeSubtree(tranxCtx, readStringParam(prepared, models.InventoryLocationFieldId))
	})
	if signal, ok := composable.AsClientErrorSignal(err); ok {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: signal.Errors}, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "InventoryLocationDomainService.Update")
	}
	return result, nil
}

// SetArchived guards archiving: the location must be empty of stock and work in flight, have no
// unarchived children, and not be a system location its warehouse still needs. Unarchiving always
// lands on suspended rather than active, because the topology may have changed while it was away.
// A location its warehouse manages is archived and unarchived by the warehouse only
// (SetArchivedManaged).
func (this *InventoryLocationDomainServiceImpl) SetArchived(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	location, vErrs, err := this.loadLocation(ctx, readStringParam(params, models.InventoryLocationFieldId))
	if err != nil {
		return nil, err
	}
	if vErrs.Count() > 0 {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}, nil
	}
	if location.IsManagedByWarehouse() {
		return locationViolationResult("", "inventory_location.archive_with_warehouse",
			"a location its warehouse manages is archived and unarchived by the warehouse"), nil
	}
	return this.setArchived(ctx, params, *location, false)
}

// SetArchivedManaged is the manager's archive of a location it manages, in the manager's own
// transaction. The archive guard still applies, so a warehouse whose Stock still holds stock or
// active sub-locations cannot be archived; only the system-location check is the manager's to
// make. A location already in the requested state is left alone. Unarchiving keeps the stored
// status: the manager decides when its records return to service. No REST route reaches it.
func (this *InventoryLocationDomainServiceImpl) SetArchivedManaged(
	ctx corectx.Context, locationId string, isArchived bool,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	location, vErrs, err := this.loadLocation(ctx, locationId)
	if err != nil {
		return nil, err
	}
	if vErrs.Count() > 0 {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}, nil
	}
	if !basemodel.IsManaged(location.GetFieldData()) {
		return locationViolationResult("", "inventory_location.not_managed",
			"only a managed location is archived this way"), nil
	}
	if location.GetIsArchived() != nil && *location.GetIsArchived() == isArchived {
		return locationMutateOk(), nil
	}
	params := dmodel.DynamicFields{
		models.InventoryLocationFieldId: locationId,
		basemodel.FieldEtag:             location.GetFieldData()[basemodel.FieldEtag],
		paramIsArchived:                 isArchived,
	}
	return this.setArchived(ctx, params, *location, true)
}

// setArchived is the archive shared by the client's and the manager's path. byManager skips the
// system-location check, which only guards against a client, and keeps the stored status on return.
func (this *InventoryLocationDomainServiceImpl) setArchived(
	ctx corectx.Context, params dmodel.DynamicFields, location models.InventoryLocation, byManager bool,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	if readBoolParam(params, paramIsArchived) {
		if vErrs, err := this.assertArchivable(ctx, location, byManager); err != nil {
			return nil, err
		} else if vErrs.Count() > 0 {
			return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}, nil
		}
		return this.CrudDomainService.SetArchived(ctx, params)
	}

	result, err := this.CrudDomainService.SetArchived(ctx, params)
	if err != nil || result == nil || result.ClientErrors.Count() > 0 || byManager {
		return result, err
	}
	return result, this.writeStatus(ctx, derefString(location.GetId()), models.InventoryLocationStatusSuspended)
}

// Move re-parents a location and rewrites what everything beneath it inherits: canonical_name,
// warehouse_id and depth. The whole subtree is rewritten in one transaction, or descendants would
// claim a place in the tree they no longer have.
func (this *InventoryLocationDomainServiceImpl) Move(
	ctx corectx.Context, locationId string, newParentId string,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	location, vErrs, err := this.loadLocation(ctx, locationId)
	if err != nil {
		return nil, err
	}
	if vErrs.Count() > 0 {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}, nil
	}

	if isSystemGenerated(*location) {
		return locationViolationResult(
			models.InventoryLocationFieldParentLocationId,
			"inventory_location.system_protected",
			"a location the warehouse created for itself cannot be moved",
		), nil
	}
	if vErrs, err := this.assertMoveTargetValid(ctx, *location, newParentId); err != nil {
		return nil, err
	} else if vErrs.Count() > 0 {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}, nil
	}
	if vErrs, err := this.assertMoveKeepsReservationsBacked(ctx, *location, newParentId); err != nil {
		return nil, err
	} else if vErrs.Count() > 0 {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}, nil
	}

	err = withLocationTransaction(ctx, func(tranxCtx corectx.Context) error {
		if err := this.writeParent(tranxCtx, locationId, newParentId); err != nil {
			return err
		}
		return this.rederiveSubtree(tranxCtx, locationId)
	})
	if err != nil {
		return nil, err
	}
	return locationMutateOk(), nil
}

// assertArchivable is the full archive guard: stock, work in flight, children and system status.
// Historical movements are deliberately not consulted — a location referenced only by completed
// moves archives cleanly, since those records keep resolving it. The manager of a location decides
// itself when its system location is no longer needed (byManager).
func (this *InventoryLocationDomainServiceImpl) assertArchivable(
	ctx corectx.Context, location models.InventoryLocation, byManager bool,
) (*ft.ClientErrors, error) {
	vErrs := ft.NewClientErrors()
	locationId := derefString(location.GetId())

	usage, err := this.readUsage(ctx, locationId)
	if err != nil {
		return vErrs, err
	}
	if !usage.OnHandQuantity.IsZero() {
		appendLocationViolation(vErrs, "", "inventory_location.has_on_hand_stock",
			"the location still holds stock; move it out before archiving")
	}
	if !usage.ReservedQuantity.IsZero() {
		appendLocationViolation(vErrs, "", "inventory_location.has_reserved_stock",
			"the location has stock reserved against it")
	}
	if usage.OpenMoveCount > 0 {
		appendLocationViolation(vErrs, "", "inventory_location.has_open_moves",
			"the location is part of a move that has not finished")
	}
	if usage.OpenTransferCount > 0 {
		appendLocationViolation(vErrs, "", "inventory_location.has_open_transfers",
			"the location is part of a transfer that has not finished")
	}

	children, err := this.countUnarchivedChildren(ctx, locationId)
	if err != nil {
		return vErrs, err
	}
	if children > 0 {
		appendLocationViolation(vErrs, "", "inventory_location.has_children",
			"archive or move the locations underneath this one first")
	}

	backing, err := assertLocationNotBackingReservations(ctx, location)
	if err != nil {
		return vErrs, err
	}
	vErrs.ConcatPtr(backing)

	if isSystemGenerated(location) && !byManager {
		warehouseUsable, err := this.isOwningWarehouseUsable(ctx, location)
		if err != nil {
			return vErrs, err
		}
		if warehouseUsable {
			appendLocationViolation(vErrs, "", "inventory_location.system_protected",
				"the warehouse still needs this location; archive the warehouse instead")
		}
	}

	operationTypes, err := this.countOperationTypesUsing(ctx, locationId)
	if err != nil {
		return vErrs, err
	}
	if operationTypes > 0 {
		appendLocationViolation(vErrs, "", "inventory_location.used_by_operation_type",
			"an operation type still uses this location as a default; change it first")
	}
	return vErrs, nil
}

func isSystemGenerated(location models.InventoryLocation) bool {
	flag := location.GetIsSystemGenerated()
	return flag != nil && *flag
}

// assertSystemLocationUnchanged refuses the edits that would break the flow a system location
// belongs to. Renaming one, or giving it a storage category, is fine.
func assertSystemLocationUnchanged(
	params dmodel.DynamicFields, current models.InventoryLocation,
) *ft.ClientErrors {
	vErrs := ft.NewClientErrors()
	protected := []string{
		models.InventoryLocationFieldParentLocationId,
		models.InventoryLocationFieldLocationType,
	}

	for _, field := range protected {
		incoming, present := params[field]
		if !present {
			continue
		}
		if !strings.EqualFold(toComparableString(incoming), currentFieldString(current, field)) {
			appendLocationViolation(vErrs, field, "inventory_location.system_protected",
				"'"+field+"' belongs to the warehouse flow that created this location")
			return vErrs
		}
	}
	return vErrs
}

// assertWarehouseRootUnchanged refuses a change of type on a warehouse's root: the root is the
// group every location of the warehouse hangs under, and holds no stock. Its name stays editable.
func assertWarehouseRootUnchanged(
	params dmodel.DynamicFields, current models.InventoryLocation,
) *ft.ClientErrors {
	vErrs := ft.NewClientErrors()
	incoming, present := params[models.InventoryLocationFieldLocationType]
	if present && toComparableString(incoming) != derefString(current.GetLocationType()) {
		vErrs.Append(*ft.NewBusinessViolation(models.InventoryLocationFieldLocationType,
			ft.ErrorKey("inventory_location.root_type_immutable", "inventory"),
			"the type of a warehouse's root location cannot be changed"))
	}
	return vErrs
}

// assertMoveKeepsReservationsBacked refuses to re-parent a location into another warehouse while
// the warehouse it leaves needs the stock it holds. A move within the same warehouse changes no
// eligibility and is never refused for this.
func (this *InventoryLocationDomainServiceImpl) assertMoveKeepsReservationsBacked(
	ctx corectx.Context, location models.InventoryLocation, newParentId string,
) (*ft.ClientErrors, error) {
	if newParentId == "" {
		return ft.NewClientErrors(), nil
	}
	parent, parentErrs, err := this.loadLocation(ctx, newParentId)
	if err != nil || parentErrs.Count() > 0 {
		return ft.NewClientErrors(), err
	}
	if derefId(parent.GetWarehouseId()) == derefId(location.GetWarehouseId()) {
		return ft.NewClientErrors(), nil
	}
	return assertLocationNotBackingReservations(ctx, location)
}
