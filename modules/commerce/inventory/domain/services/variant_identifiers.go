package services

import (
	"strings"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// variantIdentifier is a variant field that must be unique within the org when present: what a
// person types or scans to name exactly one variant.
type variantIdentifier struct {
	field    string
	errorKey string
	message  string
}

var variantIdentifiers = []variantIdentifier{
	{
		field:    models.ProductVariantFieldSku,
		errorKey: "product_variant.sku_duplicate",
		message:  "another variant in this organization already uses this SKU",
	},
	{
		field:    models.ProductVariantFieldPrimaryBarcode,
		errorKey: "product_variant.primary_barcode_duplicate",
		message:  "another variant in this organization already uses this barcode",
	},
}

// normalizeVariantIdentifiers trims the SKU and barcode and turns a blank one into NULL. The
// database's partial uniques only skip NULL, so an empty string left as it is would collide with
// every other variant whose identifier was cleared the same way.
func normalizeVariantIdentifiers() composable.BeforeValidationFn {
	return func(
		_ corectx.Context, entity *composable.DynamicEntity, _ *ft.ClientErrors,
	) (*composable.DynamicEntity, error) {
		data := entity.GetFieldData()
		for _, identifier := range variantIdentifiers {
			raw, ok := data[identifier.field].(string)
			if !ok {
				continue
			}
			if trimmed := strings.TrimSpace(raw); trimmed != "" {
				data[identifier.field] = trimmed
			} else {
				data[identifier.field] = nil
			}
		}
		return entity, nil
	}
}

// validateCreateIdentifiers reports a SKU or barcode clash on its own field before the partial
// unique index refuses the insert with an error naming no field.
func validateCreateIdentifiers(repo models.ProductSearcher) composable.ValidateExtraFn {
	return func(
		ctx corectx.Context, inputModel *composable.DynamicEntity, _ *composable.DynamicEntity, vErrs *ft.ClientErrors,
	) error {
		data := inputModel.GetFieldData()
		orgId, _ := data[models.ProductVariantFieldOrgId].(string)
		return checkUniqueIdentifiers(ctx, repo, orgId, data, "", vErrs)
	}
}

// validateUpdateIdentifiers checks only the identifiers the update submits: an untouched one
// cannot have started colliding through this write. The org comes from the stored row, because an
// update payload need not carry it.
func validateUpdateIdentifiers(repo models.ProductSearcher) composable.ValidateExtraFn {
	return func(
		ctx corectx.Context, inputModel *composable.DynamicEntity, foundModel *composable.DynamicEntity, vErrs *ft.ClientErrors,
	) error {
		if foundModel == nil {
			return nil
		}
		stored := models.NewProductVariantFrom(foundModel.GetFieldData())
		orgId, _ := foundModel.GetFieldData()[models.ProductVariantFieldOrgId].(string)
		return checkUniqueIdentifiers(ctx, repo, orgId, inputModel.GetFieldData(), derefString(stored.GetId()), vErrs)
	}
}

// checkUniqueIdentifiers searches the org for another variant holding each submitted identifier.
// Archived variants count: an archived variant can be restored, and two live variants answering
// to one scan is exactly what the constraint prevents.
func checkUniqueIdentifiers(
	ctx corectx.Context, repo models.ProductSearcher, orgId string,
	data dmodel.DynamicFields, selfId string, vErrs *ft.ClientErrors,
) error {
	if orgId == "" {
		return nil
	}
	for _, identifier := range variantIdentifiers {
		value, ok := data[identifier.field].(string)
		if !ok || value == "" {
			continue
		}
		taken, err := identifierTakenByOther(ctx, repo, orgId, identifier.field, value, selfId)
		if err != nil {
			return err
		}
		if taken {
			vErrs.Append(*ft.NewBusinessViolation(identifier.field,
				ft.ErrorKey(identifier.errorKey, "inventory"), identifier.message))
		}
	}
	return nil
}

// identifierTakenByOther reads at most two rows: the record being updated may be one of them,
// which is no conflict with itself.
func identifierTakenByOther(
	ctx corectx.Context, repo models.ProductSearcher, orgId string, field string, value string, selfId string,
) (bool, error) {
	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(models.ProductVariantFieldOrgId, dmodel.Equals, orgId),
		*dmodel.NewSearchNode().NewCondition(field, dmodel.Equals, value),
	)
	found, err := repo.Search(ctx, dyn.RepoSearchParam{
		Graph:  graph,
		Fields: []string{models.ProductVariantFieldId},
		Page:   0,
		Size:   2,
	})
	if err != nil {
		return false, errors.Wrap(err, "identifierTakenByOther")
	}
	if found.ClientErrors.Count() > 0 {
		return false, errors.Wrap(found.ClientErrors.ToError(), "identifierTakenByOther")
	}
	for _, row := range found.Data.Items {
		if derefString(models.NewProductVariantFrom(row).GetId()) != selfId {
			return true, nil
		}
	}
	return false, nil
}
