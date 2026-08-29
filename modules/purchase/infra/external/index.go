// Package external binds Purchase's local ports to the services other modules publish. It is the
// only package in Purchase that may import another module; everything else depends on
// interfaces/external, so a split into separate processes changes this file and nothing else.
package external

import (
	stdErr "errors"

	"go.bryk.io/pkg/errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	itVendor "github.com/sky-as-code/nikki-erp/modules/contacts/interfaces/vendor"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	itCurrency "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/currency"
	itUom "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/uom"
	invModels "github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
	itExt "github.com/sky-as-code/nikki-erp/modules/purchase/interfaces/external"
)

// InitExternal binds every port Purchase consumes, and registers what Purchase offers back.
func InitExternal() error {
	// Purchase is the first module to hold UoM references, so this is what turns Essential's
	// unit-in-use guard from an assumption into an enforced rule.
	RegisterUomUsageProbe()

	return stdErr.Join(
		deps.Register(func(uomSvc itUom.UomConversionAppService) itExt.UomExtService {
			// The upstream service already has exactly the two methods the port declares, so this is
			// a hand-over rather than an adapter.
			return uomSvc
		}),
		deps.Register(func(variantSvc itProduct.ProductVariantDomainService) itExt.ProductExtService {
			return &productAdapter{variants: variantSvc}
		}),
		deps.Register(func(vendorSvc itVendor.VendorAppService) itExt.VendorExtService {
			return vendorSvc
		}),
		deps.Register(func(currencySvc itCurrency.CurrencyAppService) itExt.CurrencyExtService {
			return currencySvc
		}),
	)
}

// productAdapter narrows Inventory's variant service to the two questions a purchase line asks. It
// is a real adapter because no single Inventory service answers both: purchasability comes from the
// variant's template field, the inventory unit from stock_product_config, which has no port.
type productAdapter struct {
	variants itProduct.ProductVariantDomainService
}

var _ itExt.ProductExtService = (*productAdapter)(nil)

func (this *productAdapter) GetPurchasableProduct(
	ctx corectx.Context, query itExt.GetPurchasableProductQuery,
) (*itExt.GetPurchasableProductResult, error) {
	if query.VariantId == "" {
		return &itExt.GetPurchasableProductResult{}, nil
	}

	found, err := this.variants.GetProductVariant(ctx, dyn.GetOneQuery{
		Id: query.VariantId,
		Fields: []string{
			invModels.ProductVariantFieldId,
			invModels.ProductVariantFieldProductTemplateId,
			invModels.ProductVariantFieldTemplatePurchaseOk,
			basemodel.FieldIsArchived,
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "GetPurchasableProduct")
	}
	if found == nil || !found.HasData {
		return &itExt.GetPurchasableProductResult{}, nil
	}

	fields := found.Data.GetFieldData()
	templateId := derefId(fields.GetModelId(invModels.ProductVariantFieldProductTemplateId))

	inventoryUomId, err := this.inventoryUomOf(ctx, string(templateId))
	if err != nil {
		return nil, err
	}

	return &itExt.GetPurchasableProductResult{
		HasData: true,
		Data: itExt.GetPurchasableProductResultData{
			VariantId:      query.VariantId,
			TemplateId:     templateId,
			Purchasable:    derefBool(fields.GetBool(invModels.ProductVariantFieldTemplatePurchaseOk)),
			InventoryUomId: inventoryUomId,
			Archived:       derefBool(fields.GetBool(basemodel.FieldIsArchived)),
		},
	}, nil
}

// inventoryUomOf reads the unit a template's stock is counted in. It searches
// inventory_stock_product_config directly because that resource has no external port. A template
// with no configuration row returns "", meaning no inventory unit to reconcile against.
func (this *productAdapter) inventoryUomOf(ctx corectx.Context, templateId string) (model.Id, error) {
	if templateId == "" {
		return "", nil
	}

	engine, ok := engineFor(invModels.StockProductConfigSchemaName)
	if !ok {
		// A deployment without the stock feature is not an error: nothing can be reconciled.
		return "", nil
	}

	graph := &dmodel.SearchGraph{}
	graph.And(*dmodel.NewSearchNode().NewCondition(
		invModels.StockProductConfigFieldProductTemplateId, dmodel.Equals, templateId))

	found, err := engine.ResourceRepository().Search(ctx, dyn.RepoSearchParam{
		Graph: graph,
		Page:  0,
		Size:  1,
	})
	if err != nil {
		return "", errors.Wrap(err, "inventoryUomOf")
	}
	if found == nil || !found.HasData || len(found.Data.Items) == 0 {
		return "", nil
	}
	return derefId(found.Data.Items[0].GetModelId(
		invModels.StockProductConfigFieldInventoryUomId)), nil
}

// engineFor resolves another module's resource engine from the shared registry — how this file
// reads a resource that publishes no port. Confined to this package, like every cross-module import.
func engineFor(schemaName string) (drif.DynamicResourceEngine, bool) {
	return dynamicresource.Registry().GetEngine(schemaName)
}

func derefId(value *model.Id) model.Id {
	if value == nil {
		return ""
	}
	return *value
}

func derefBool(value *bool) bool {
	return value != nil && *value
}
