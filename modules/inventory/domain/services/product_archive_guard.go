package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

// Whether a product may be withdrawn from the working set.
//
// Archiving is not a way to make stock go away. A variant still holding goods, still owing them to
// a reservation, or still named by work in flight cannot be archived: the stock must be dealt with
// through a stock operation first (CR §14, §14.4, PROD-INT-INV-013..016).
//
// What does *not* block it is history. A variant whose only remaining trace is completed movement
// archives fine, because those records keep resolving it afterwards — treating them as blockers
// would make a product that has ever been sold impossible to retire (CR §14.2, TS-PROD-11).
//
// Stock reports the numbers, Product decides what they mean. The two are separate so that Stock
// stays free to change how it stores any of it, and so this rule lives in one place rather than
// being restated at each call site.

// stockUsageReader resolves the service that answers what a variant holds.
//
// It is the quant engine's service, which implements the port. Resolving it here rather than
// injecting it keeps the guard callable from the archive path without threading a dependency
// through the resource service's constructor, which the engine builds.
//
// An error means the stock side is not wired in this deployment. Callers treat that as "cannot
// check" rather than "must fail": Product has to keep working when Stock is unavailable
// (CR §25, AC-PROD-INT-036).
func stockUsageReader() (itStock.StockProductUsageReader, error) {
	engine, err := engineFor(models.StockQuantSchemaName)
	if err != nil {
		return nil, err
	}

	reader, ok := engine.ResourceService().(itStock.StockProductUsageReader)
	if !ok {
		return nil, errors.New(
			"the stock quant engine is not running the derived quant service; " +
				"InventoryModule.Init must install it with SetResourceService")
	}
	return reader, nil
}

// AssertVariantArchivable adds a violation for each reason the variant cannot be archived.
//
// All four are reported rather than the first, so a user fixing them learns the whole story at
// once instead of discovering the next blocker after clearing each one.
func AssertVariantArchivable(usage itStock.ProductUsage, vErrs *ft.ClientErrors) {
	if !usage.OnHandQuantity.IsZero() {
		vErrs.Append(*ft.NewBusinessViolation(
			models.ProductVariantSchemaName,
			"product_variant.has_on_hand_stock",
			"the product still has stock on hand; move, sell or write it off before archiving",
		))
	}
	if !usage.ReservedQuantity.IsZero() {
		vErrs.Append(*ft.NewBusinessViolation(
			models.ProductVariantSchemaName,
			"product_variant.has_reserved_stock",
			"the product still has stock reserved against an outgoing operation",
		))
	}
	if usage.OpenMoveCount > 0 {
		vErrs.Append(*ft.NewBusinessViolation(
			models.ProductVariantSchemaName,
			"product_variant.has_open_moves",
			"the product is named by a stock move that has not completed",
		))
	}
	if usage.OpenTransferCount > 0 {
		vErrs.Append(*ft.NewBusinessViolation(
			models.ProductVariantSchemaName,
			"product_variant.has_open_transfers",
			"the product is named by a transfer that has not completed",
		))
	}
}

// AssertTemplateArchivable refuses a template archive when any of its variants is still in use.
//
// The whole set is judged before anything is written. Archiving a template cascades to its
// variants, so checking one variant at a time inside that cascade would leave the first few
// archived when a later one turns out to hold stock — a half-archived template that no single
// operation had asked for (CR §14.3, AC-PROD-INT-032, TS-PROD-12).
//
// The offending variant is named. "Some variant has stock" sends a user through a product line
// one row at a time looking for the one that does.
func AssertTemplateArchivable(
	usages map[string]itStock.ProductUsage, skuOf map[string]string, vErrs *ft.ClientErrors,
) {
	for variantId, usage := range usages {
		if usage.IsEmpty() {
			continue
		}

		label := skuOf[variantId]
		if label == "" {
			label = variantId
		}
		vErrs.Append(*ft.NewBusinessViolation(
			models.ProductTemplateSchemaName,
			"product_template.variant_has_stock",
			"variant '"+label+"' still has stock, a reservation or work in flight; "+
				"archiving the product line would strand it",
		))
	}
}

// GuardVariantArchive checks one variant before it is archived.
//
// Unarchiving is never guarded: restoring a product to the working set strands nothing, and
// requiring it to be stockless would make an archived-by-mistake variant unrecoverable.
func GuardVariantArchive(
	ctx corectx.Context, reader itStock.StockProductUsageReader,
	variantId string, archiving bool, vErrs *ft.ClientErrors,
) error {
	if !archiving || variantId == "" {
		return nil
	}

	result, err := reader.GetProductUsage(ctx, itStock.GetProductUsageQuery{VariantId: variantId})
	if err != nil {
		return errors.Wrap(err, "GuardVariantArchive")
	}
	AssertVariantArchivable(result.Data.Usage, vErrs)
	return nil
}

// guardVariantStockUsage refuses a variant archive that would strand stock.
//
// It returns a non-nil result when the operation is refused, which the caller returns as-is; nil
// means it may proceed. Unarchiving is waved through: it strands nothing.
func guardVariantStockUsage(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	variant := models.NewProductVariantFrom(params)
	if !derefBool(variant.IsArchived()) {
		return nil, nil
	}

	reader, err := stockUsageReader()
	if err != nil {
		// No reader means the stock side is not wired in this deployment. Product must keep
		// working without it rather than becoming unusable (CR §25, AC-PROD-INT-036).
		return nil, nil
	}

	vErrs := &ft.ClientErrors{}
	err = GuardVariantArchive(ctx, reader, derefId(variant.GetId()), true, vErrs)
	if err != nil {
		return nil, err
	}
	if vErrs.Count() == 0 {
		return nil, nil
	}
	return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}, nil
}

// GuardTemplateArchive checks every variant of a template before any of them is archived.
//
// Variants already archived are skipped. One of those is not being withdrawn by this operation —
// it is already out — so its stock, if any, is a pre-existing condition rather than something this
// archive would strand, and blocking on it would make the template permanently unarchivable.
func GuardTemplateArchive(
	ctx corectx.Context, reader itStock.StockProductUsageReader,
	templateId string, archiving bool, vErrs *ft.ClientErrors,
) error {
	if !archiving || templateId == "" {
		return nil
	}

	variantEngine, err := engineFor(models.ProductVariantSchemaName)
	if err != nil {
		return err
	}

	rows, err := models.FindTemplateVariants(
		ctx, variantEngine.ResourceRepository(), templateId, MaxCascadeVariants)
	if err != nil {
		return errors.Wrap(err, "GuardTemplateArchive")
	}

	variantIds := make([]string, 0, len(rows))
	skuOf := make(map[string]string, len(rows))
	for _, row := range rows {
		variant := models.NewProductVariantFrom(row)
		if derefBool(variant.IsArchived()) {
			continue
		}
		id := derefId(variant.GetId())
		if id == "" {
			continue
		}
		variantIds = append(variantIds, id)
		skuOf[id] = derefString(variant.GetSku())
	}
	if len(variantIds) == 0 {
		return nil
	}

	result, err := reader.GetProductUsageBatch(
		ctx, itStock.GetProductUsageBatchQuery{VariantIds: variantIds})
	if err != nil {
		return errors.Wrap(err, "GuardTemplateArchive")
	}
	AssertTemplateArchivable(result.Data.Usages, skuOf, vErrs)
	return nil
}
