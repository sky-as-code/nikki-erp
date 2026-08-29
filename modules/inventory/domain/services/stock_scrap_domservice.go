package services

import (
	"strings"

	"github.com/shopspring/decimal"
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"

	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// NewStockScrapDomainService derives the scrap service from the engine's default one, which it
// embeds so built-in CRUD keeps running unchanged.
func NewStockScrapDomainService(base drif.DynamicResourceService) *StockScrapDomainServiceImpl {
	return &StockScrapDomainServiceImpl{DynamicResourceService: base}
}

// StockScrapDomainServiceImpl adds the scrap document's own rules. A scrap is a document, not a
// balance: while draft it changes nothing, and completing it generates the movement. That is why
// create, update and delete stay open here, merely constrained, where the quant refuses them.
type StockScrapDomainServiceImpl struct {
	drif.DynamicResourceService
}

var _ drif.DynamicResourceService = (*StockScrapDomainServiceImpl)(nil)

// Create stamps the scrap number and forces the status to draft. A client-chosen number could
// collide with or impersonate another document's, and a scrap created `done` would be a completed
// movement with nothing behind it.
func (this *StockScrapDomainServiceImpl) Create(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[dmodel.DynamicFields], error) {
	if vErrs := assertScrapQuantityPositive(params); vErrs.Count() > 0 {
		return &dyn.OpResult[dmodel.DynamicFields]{ClientErrors: *vErrs}, nil
	}

	scrapNumber, err := generateScrapNumber()
	if err != nil {
		return nil, err
	}

	prepared := dmodel.DynamicFields{}
	for key, value := range params {
		prepared[key] = value
	}
	prepared[models.StockScrapFieldScrapNumber] = scrapNumber
	prepared[models.StockScrapFieldStatus] = models.StockScrapStatusDraft
	// Set by Do Scrap and nothing else, so a client cannot present an unexecuted document as
	// executed.
	delete(prepared, models.StockScrapFieldMoveId)
	delete(prepared, models.StockScrapFieldCompletedAt)
	applyEmptyStringDefault(prepared, models.StockScrapFieldLotRef)
	applyEmptyStringDefault(prepared, models.StockScrapFieldPackageRef)
	applyEmptyStringDefault(prepared, models.StockScrapFieldOwnerRef)

	return this.DynamicResourceService.Create(ctx, prepared)
}

// Update refuses to change a scrap that is already done: editing it would rewrite the description
// of a movement that already happened, leaving document and stock disagreeing.
func (this *StockScrapDomainServiceImpl) Update(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	scrap, vErrs, err := this.loadScrap(ctx, readStringParam(params, models.StockScrapFieldId))
	if err != nil {
		return nil, err
	}
	if vErrs.Count() > 0 {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}, nil
	}

	if derefString(scrap.GetStatus()) == models.StockScrapStatusDone {
		return &dyn.OpResult[dyn.MutateResultData]{
			ClientErrors: *scrapViolation("stock_scrap.already_done", "a completed scrap cannot be edited"),
		}, nil
	}
	if vErrs := assertScrapQuantityPositive(params); vErrs.Count() > 0 {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}, nil
	}

	// The execution fields belong to Do Scrap; an update that could set them would let a client mark
	// a document done without any stock moving.
	prepared := dmodel.DynamicFields{}
	for key, value := range params {
		prepared[key] = value
	}
	delete(prepared, models.StockScrapFieldStatus)
	delete(prepared, models.StockScrapFieldMoveId)
	delete(prepared, models.StockScrapFieldCompletedAt)
	delete(prepared, models.StockScrapFieldScrapNumber)

	return this.DynamicResourceService.Update(ctx, prepared)
}

// Delete refuses to remove a scrap that is already done: the movement it generated is permanent,
// and deleting the document would leave it unexplained. A scrap made in error is corrected by a
// reverse movement.
func (this *StockScrapDomainServiceImpl) Delete(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	scrap, vErrs, err := this.loadScrap(ctx, readStringParam(params, models.StockScrapFieldId))
	if err != nil {
		return nil, err
	}
	if vErrs.Count() > 0 {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}, nil
	}

	if derefString(scrap.GetStatus()) == models.StockScrapStatusDone {
		return &dyn.OpResult[dyn.MutateResultData]{
			ClientErrors: *scrapViolation(
				"stock_scrap.done_not_deletable",
				"a completed scrap cannot be deleted; reverse the movement instead"),
		}, nil
	}
	return this.DynamicResourceService.Delete(ctx, params)
}

// loadScrap reads one scrap, reporting a missing id as a client error.
func (this *StockScrapDomainServiceImpl) loadScrap(
	ctx corectx.Context, scrapId string,
) (*models.StockScrap, *ft.ClientErrors, error) {
	vErrs := ft.NewClientErrors()
	if scrapId == "" {
		vErrs.Append(*ft.NewBusinessViolation(
			models.StockScrapSchemaName, "stock_scrap.id_required", "a scrap must be identified"))
		return nil, vErrs, nil
	}

	engine, err := engineFor(models.StockScrapSchemaName)
	if err != nil {
		return nil, vErrs, err
	}
	found, err := engine.ResourceRepository().FindByKeys(ctx, dmodel.DynamicFields{
		models.StockScrapFieldId: scrapId,
	})
	if err != nil {
		return nil, vErrs, errors.Wrap(err, "loadScrap")
	}
	if found == nil || !found.HasData {
		vErrs.Append(*ft.NewBusinessViolation(
			models.StockScrapSchemaName, "stock_scrap.not_found",
			"no stock scrap with id '"+scrapId+"'"))
		return nil, vErrs, nil
	}
	return models.NewStockScrapFrom(found.Data), vErrs, nil
}

// assertScrapQuantityPositive rejects a zero quantity, which the column allows: a scrap of nothing
// claims a movement happened when none did.
func assertScrapQuantityPositive(params dmodel.DynamicFields) *ft.ClientErrors {
	vErrs := ft.NewClientErrors()
	raw, present := params[models.StockScrapFieldQuantity]
	if !present || raw == nil {
		return vErrs
	}

	quantity, err := decimalFromAny(raw)
	if err != nil {
		vErrs.Append(*ft.NewBusinessViolation(
			models.StockScrapSchemaName, "stock_scrap.quantity_malformed",
			"'quantity' must be a decimal number"))
		return vErrs
	}
	if quantity.LessThanOrEqual(decimal.Zero) {
		vErrs.Append(*ft.NewBusinessViolation(
			models.StockScrapSchemaName, "stock_scrap.quantity_not_positive",
			"a scrap must remove a quantity greater than zero"))
	}
	return vErrs
}

func decimalFromAny(value any) (decimal.Decimal, error) {
	switch typed := value.(type) {
	case string:
		return decimal.NewFromString(typed)
	case decimal.Decimal:
		return typed, nil
	case *decimal.Decimal:
		if typed == nil {
			return decimal.Zero, errors.New("nil decimal")
		}
		return *typed, nil
	case float64:
		return decimal.NewFromFloat(typed), nil
	case int:
		return decimal.NewFromInt(int64(typed)), nil
	case int64:
		return decimal.NewFromInt(typed), nil
	default:
		return decimal.Zero, errors.New("not a decimal")
	}
}

// applyEmptyStringDefault keeps the dimension fields as "" rather than NULL. The quant and the move
// line use the same convention and must agree: a scrap whose lot_ref is NULL would not line up with
// the balance it draws from. Revisit all three together if real Lot/Package master data lands.
func applyEmptyStringDefault(params dmodel.DynamicFields, field string) {
	if value, ok := params[field]; !ok || value == nil {
		params[field] = ""
	}
}

func scrapViolation(key, message string) *ft.ClientErrors {
	vErrs := ft.NewClientErrors()
	vErrs.Append(*ft.NewBusinessViolation(models.StockScrapSchemaName, key, message))
	return vErrs
}

// generateScrapNumber builds the document number from a ULID rather than a counter, which would
// need its own sequence table and serialise every create in an org. Uniqueness is enforced by the
// composite unique on (scrap_number, org_id).
func generateScrapNumber() (string, error) {
	id, err := model.NewId()
	if err != nil {
		return "", errors.Wrap(err, "generateScrapNumber")
	}
	return "SCR-" + strings.ToUpper(string(*id)), nil
}
