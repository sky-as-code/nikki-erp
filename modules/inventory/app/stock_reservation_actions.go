package app

import (
	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"

	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

// The reservation operations a client may call. Each authorizes first and then binds the request
// field map into the typed port the domain service takes. A client reserving over REST is stamped
// with the source module "inventory": the module name in a source is what the calling port
// vouches for, never something a request body may claim.

const (
	clientSourceModule      = "inventory"
	defaultClientSourceType = "manual"

	rsvParamWarehouseId    = "warehouse_id"
	rsvParamProductVariant = "product_variant_id"
	rsvParamLines          = "lines"
	rsvParamReservedUntil  = "reserved_until"
	rsvParamSourceType     = "source_type"
	rsvParamSourceId       = "source_id"
	rsvParamSourceRevision = "source_revision"
	rsvParamIdempotencyKey = "idempotency_key"
	rsvParamRequested      = "requested"
	rsvParamSourceLineId   = "source_line_id"
	rsvParamUomId          = "uom_id"
	rsvParamQuantity       = "quantity"

	rsvParamExecutionId     = "execution_id"
	rsvParamActualSources   = "actual_sources"
	rsvParamLocationId      = "location_id"
	rsvParamLotRef          = "lot_ref"
	rsvParamPackageRef      = "package_ref"
	rsvParamOwnerRef        = "owner_ref"
	rsvParamOperationTypeId = "operation_type_id"
	rsvParamDestinationId   = "destination_location_id"
	rsvParamOriginReference = "origin_reference"
)

// ConsumeReservation records goods that physically left stock against one reservation.
func (this *StockReservationApplicationServiceImpl) ConsumeReservation(
	ctx corectx.Context, cmd itStock.ConsumeReservationCommand,
) (*dyn.OpResult[any], error) {
	cErrs, err := assertRecordAction(this, ctx, PermissionConsumeReservation, cmd)
	if err != nil || cErrs != nil {
		return anyFailure(cErrs, err)
	}

	request, vErrs := bindConsumeRequest(cmd)
	if vErrs.Count() > 0 {
		return anyFailure(vErrs, nil)
	}
	result, err := this.reservationSvc.ConsumeReservation(ctx, request)
	if err != nil {
		return nil, err
	}
	if result.Refused() {
		return anyFailure(&result.ClientErrors, nil)
	}
	return anyResult(result), nil
}

// ReleaseReservation gives back the unconsumed remainder of one reservation.
func (this *StockReservationApplicationServiceImpl) ReleaseReservation(
	ctx corectx.Context, cmd itStock.ReleaseReservationCommand,
) (*dyn.OpResult[any], error) {
	cErrs, err := assertRecordAction(this, ctx, PermissionReleaseReservation, cmd)
	if err != nil || cErrs != nil {
		return anyFailure(cErrs, err)
	}
	result, err := this.reservationSvc.ReleaseReservation(ctx, itStock.ReleaseReservationRequest{
		OrgId:         model.Id(readStringField(cmd, "org_id")),
		ReservationId: model.Id(readStringField(cmd, paramRecordId)),
		Reason:        readStringField(cmd, "reason"),
	})
	if err != nil {
		return nil, err
	}
	if result.Refused() {
		return anyFailure(&result.ClientErrors, nil)
	}
	return anyResult(result), nil
}

func bindConsumeRequest(cmd dmodel.DynamicFields) (itStock.ConsumeReservationRequest, *ft.ClientErrors) {
	vErrs := ft.NewClientErrors()
	request := itStock.ConsumeReservationRequest{
		OrgId:                 model.Id(readStringField(cmd, "org_id")),
		ReservationId:         model.Id(readStringField(cmd, paramRecordId)),
		ExecutionId:           readStringField(cmd, rsvParamExecutionId),
		IdempotencyKey:        readStringField(cmd, rsvParamIdempotencyKey),
		OperationTypeId:       model.Id(readStringField(cmd, rsvParamOperationTypeId)),
		DestinationLocationId: model.Id(readStringField(cmd, rsvParamDestinationId)),
		OriginReference:       readStringField(cmd, rsvParamOriginReference),
	}
	rawSources, _ := cmd[rsvParamActualSources].([]any)
	for index, raw := range rawSources {
		fields, ok := raw.(map[string]any)
		if !ok {
			vErrs.Append(*ft.NewBusinessViolation(models.StockReservationSchemaName,
				models.StockReservationSchemaName+"."+rsvParamActualSources+"_malformed",
				"actual source "+itoa(index+1)+" must be an object"))
			continue
		}
		source := dmodel.DynamicFields(fields)
		quantity, sourceErrs := readDecimalField(source, models.StockReservationSchemaName, rsvParamQuantity)
		if sourceErrs.Count() > 0 {
			vErrs.ConcatPtr(sourceErrs)
			continue
		}
		request.ActualSources = append(request.ActualSources, itStock.ActualSource{
			LocationId: model.Id(readStringField(source, rsvParamLocationId)),
			LotRef:     readStringField(source, rsvParamLotRef),
			PackageRef: readStringField(source, rsvParamPackageRef),
			OwnerRef:   readStringField(source, rsvParamOwnerRef),
			Quantity:   quantity,
		})
	}
	return request, vErrs
}

// ReserveWarehouseStock holds every line of the request at a warehouse, or nothing.
func (this *StockReservationApplicationServiceImpl) ReserveWarehouseStock(
	ctx corectx.Context, cmd itStock.ReserveWarehouseStockCommand,
) (*dyn.OpResult[any], error) {
	orgId, cErrs := this.AssertAction(ctx, PermissionReserveWarehouseStock, cmd)
	if cErrs != nil {
		return anyFailure(cErrs, nil)
	}

	request, vErrs := bindReserveRequest(cmd, orgId)
	if vErrs.Count() > 0 {
		return anyFailure(vErrs, nil)
	}
	result, err := this.reservationSvc.ReserveWarehouseStock(ctx, request)
	if err != nil {
		return nil, err
	}
	if result.Refused() {
		return anyFailure(&result.ClientErrors, nil)
	}
	return anyResult(result), nil
}

// CheckWarehouseAvailability answers how much of a variant the warehouse can still commit.
func (this *StockReservationApplicationServiceImpl) CheckWarehouseAvailability(
	ctx corectx.Context, query itStock.CheckWarehouseAvailabilityCommand,
) (*dyn.OpResult[any], error) {
	orgId, cErrs := this.AssertAction(ctx, PermissionCheckWarehouseAvailability, query)
	if cErrs != nil {
		return anyFailure(cErrs, nil)
	}

	requested := decimal.Zero
	if _, present := query[rsvParamRequested]; present {
		parsed, vErrs := readDecimalField(query, models.StockReservationSchemaName, rsvParamRequested)
		if vErrs.Count() > 0 {
			return anyFailure(vErrs, nil)
		}
		requested = parsed
	}
	report, err := this.reservationSvc.CheckWarehouseAvailability(ctx, itStock.CheckWarehouseAvailabilityQuery{
		OrgId:            idOrEmpty(orgId),
		WarehouseId:      model.Id(readStringField(query, rsvParamWarehouseId)),
		ProductVariantId: model.Id(readStringField(query, rsvParamProductVariant)),
		Requested:        requested,
	})
	if err != nil {
		return nil, err
	}
	if report.ClientErrors.Count() > 0 {
		return anyFailure(&report.ClientErrors, nil)
	}
	return anyResult(report), nil
}

// bindReserveRequest reads the reserve request out of the bound field map. Lines arrive as a JSON
// array of objects; a malformed line is reported by position rather than silently dropped, since
// a dropped line would reserve less than the caller asked for and report success.
func bindReserveRequest(cmd dmodel.DynamicFields, orgId *model.Id) (itStock.ReserveWarehouseStockRequest, *ft.ClientErrors) {
	vErrs := ft.NewClientErrors()
	request := itStock.ReserveWarehouseStockRequest{
		OrgId:          idOrEmpty(orgId),
		WarehouseId:    model.Id(readStringField(cmd, rsvParamWarehouseId)),
		SourceModule:   clientSourceModule,
		SourceType:     readStringField(cmd, rsvParamSourceType),
		SourceId:       readStringField(cmd, rsvParamSourceId),
		IdempotencyKey: readStringField(cmd, rsvParamIdempotencyKey),
	}
	if request.SourceType == "" {
		request.SourceType = defaultClientSourceType
	}
	if revision, ok := cmd[rsvParamSourceRevision]; ok && revision != nil {
		parsed, ok := toDecimal(revision)
		if !ok || !parsed.IsInteger() || !parsed.IsPositive() {
			vErrs.Append(*ft.NewBusinessViolation(models.StockReservationSchemaName,
				models.StockReservationSchemaName+"."+rsvParamSourceRevision+"_malformed",
				"'"+rsvParamSourceRevision+"' must be a positive whole number"))
		} else {
			request.SourceRevision = int32(parsed.IntPart())
		}
	}
	if until := readStringField(cmd, rsvParamReservedUntil); until != "" {
		parsed, err := model.ParseModelDateTime(until)
		if err != nil {
			vErrs.Append(*ft.NewBusinessViolation(models.StockReservationSchemaName,
				models.StockReservationSchemaName+"."+rsvParamReservedUntil+"_malformed",
				"'"+rsvParamReservedUntil+"' must be an RFC 3339 UTC timestamp ending in Z"))
		} else {
			goTime := parsed.GoTime()
			request.ReservedUntil = &goTime
		}
	}

	rawLines, _ := cmd[rsvParamLines].([]any)
	for index, raw := range rawLines {
		fields, ok := raw.(map[string]any)
		if !ok {
			vErrs.Append(*ft.NewBusinessViolation(models.StockReservationSchemaName,
				models.StockReservationSchemaName+"."+rsvParamLines+"_malformed",
				"line "+itoa(index+1)+" must be an object"))
			continue
		}
		line := dmodel.DynamicFields(fields)
		quantity, lineErrs := readDecimalField(line, models.StockReservationSchemaName, rsvParamQuantity)
		if lineErrs.Count() > 0 {
			vErrs.ConcatPtr(lineErrs)
			continue
		}
		request.Lines = append(request.Lines, itStock.WarehouseReservationLine{
			SourceLineId:     readStringField(line, rsvParamSourceLineId),
			ProductVariantId: model.Id(readStringField(line, rsvParamProductVariant)),
			UomId:            model.Id(readStringField(line, rsvParamUomId)),
			Quantity:         quantity,
		})
	}
	return request, vErrs
}

func itoa(value int) string {
	return decimal.NewFromInt(int64(value)).String()
}
