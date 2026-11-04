package services

import (
	"strings"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
	invsettings "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/settings"
	itExt "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/external"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// Discriminators of the transfers one request numbers besides the one it creates: each names the
// source document, so a retry keys the same issuance.
const (
	transferKeyBackorderOf = "backorder_of"
	transferKeyReturnOf    = "return_of"
	transferKeyCorrection  = "correction"
)

// transferNumberIssuer is what the operation type's domain service offers the transfer paths: the
// next number on one type's own counter.
type transferNumberIssuer interface {
	IssueTransferNumber(
		ctx corectx.Context, operationTypeId string, discriminators ...string,
	) (string, *ft.ClientErrors, error)
}

// buildTransferNumber numbers a new transfer of the operation type: {warehouse code}/{sequence
// prefix}/{padded number}. discriminators tell apart several transfers one request numbers (see
// itExt.SequenceOperationKey); a plain create passes none. Client errors (numbering not set up, an
// archived definition, an unknown type) must make the caller roll back: the number is reserved.
func buildTransferNumber(
	ctx corectx.Context, operationTypeId string, discriminators ...string,
) (string, *ft.ClientErrors, error) {
	domSvc, err := domainServiceFor(models.StockOperationTypeSchemaName)
	if err != nil {
		return "", nil, err
	}
	issuer, ok := domSvc.(transferNumberIssuer)
	if !ok {
		return "", nil, errors.New("the stock operation type service does not number transfers")
	}
	return issuer.IssueTransferNumber(ctx, operationTypeId, discriminators...)
}

// mustBuildTransferNumber is buildTransferNumber for a path whose only failure channel is an error,
// which also rolls back the transaction it runs in.
func mustBuildTransferNumber(ctx corectx.Context, operationTypeId string, discriminators ...string) (string, error) {
	number, cErrs, err := buildTransferNumber(ctx, operationTypeId, discriminators...)
	if err != nil {
		return "", err
	}
	if cErrs != nil && cErrs.Count() > 0 {
		return "", errors.Wrap(cErrs.ToError(), "number the transfer")
	}
	return number, nil
}

// IssueTransferNumber issues the next number of the operation type from the org's inventory.transfer
// definition, the type being the scope object of its own counter.
func (this *StockOperationTypeDomainServiceImpl) IssueTransferNumber(
	ctx corectx.Context, operationTypeId string, discriminators ...string,
) (string, *ft.ClientErrors, error) {
	operationType, cErrs, err := this.loadForNumbering(ctx, operationTypeId)
	if err != nil || cErrs != nil {
		return "", cErrs, err
	}
	orgId := model.Id(derefId(operationType.GetOrgId()))
	definitionId, cErrs, err := sequenceDefinitionOf(ctx, this.settings, orgId,
		invsettings.OrgSettingTransferSequenceDefinitionId)
	if err != nil || cErrs != nil {
		return "", cErrs, err
	}
	warehouseCode, err := warehouseCodeOf(ctx, derefId(operationType.GetWarehouseId()))
	if err != nil {
		return "", nil, err
	}
	scopeObjectId := model.Id(operationTypeId)
	padded, cErrs, err := nextSequenceValue(ctx, this.sequences, sequenceRequest{
		definitionId: definitionId, orgId: orgId, scopeObjectId: &scopeObjectId,
		operationKey: itExt.SequenceOperationKey(ctx, models.StockTransferFieldTransferNumber, discriminators...),
	})
	if err != nil || cErrs != nil {
		return "", cErrs, err
	}
	prefix := derefString(operationType.GetSequencePrefix())
	warehouseCode, prefix = numberingSegments(warehouseCode, prefix, operationTypeId)
	return FormatTransferNumber(warehouseCode, prefix, padded), nil, nil
}

// numberingSegments falls back, for a type created before it had a sequence prefix, to the type's
// own id in place of warehouse and prefix: every type counts from 1, so two such types in one org
// would otherwise both issue "00001" against the (transfer_number, org_id) unique.
func numberingSegments(warehouseCode string, prefix string, operationTypeId string) (string, string) {
	if prefix != "" {
		return warehouseCode, prefix
	}
	return "", operationTypeId
}

func (this *StockOperationTypeDomainServiceImpl) loadForNumbering(
	ctx corectx.Context, operationTypeId string,
) (*models.StockOperationType, *ft.ClientErrors, error) {
	found, err := this.Repository().FindByKeys(ctx, dmodel.DynamicFields{
		models.StockOperationTypeFieldId: operationTypeId,
	})
	if err != nil {
		return nil, nil, errors.Wrap(err, "load the operation type to number")
	}
	if found == nil || !found.HasData {
		cErrs := ft.ClientErrors{*ft.NewBusinessViolation(models.StockTransferFieldOperationTypeId,
			ft.ErrorKey("stock_transfer.operation_type_not_found", "inventory"),
			"no stock operation type with id '"+operationTypeId+"'")}
		return nil, &cErrs, nil
	}
	return models.NewStockOperationTypeFrom(found.Data), nil, nil
}

// warehouseCodeOf reads the code that opens a transfer number; empty when the type names none.
func warehouseCodeOf(ctx corectx.Context, warehouseId string) (string, error) {
	if warehouseId == "" {
		return "", nil
	}
	repo, err := repoFor(models.WarehouseSchemaName)
	if err != nil {
		return "", err
	}
	found, err := repo.FindByKeys(ctx, dmodel.DynamicFields{models.WarehouseFieldId: warehouseId})
	if err != nil {
		return "", errors.Wrap(err, "read the warehouse code")
	}
	if found == nil || !found.HasData {
		return "", errors.Errorf("warehouse %s of the operation type is missing", warehouseId)
	}
	return derefString(models.NewWarehouseFrom(found.Data).GetCode()), nil
}

// correctionDiscriminator tells apart the corrections one request generates (a count applying many
// quants, a reservation consumed piece by piece) with what the caller already keys them by, so a
// retry draws the same numbers. Without either key the movement itself is the identity.
func correctionDiscriminator(request CorrectionRequest) string {
	if request.IdempotencyKey != "" {
		return request.IdempotencyKey
	}
	if request.OriginReference != "" {
		return request.OriginReference
	}
	return strings.Join([]string{request.ProductVariantId, request.SourceLocationId,
		request.DestinationLocationId, request.LotRef, request.PackageRef, request.OwnerRef}, "|")
}
