package dynamicengines

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// maxCategoryDepth bounds the upward walk when checking for a cycle. It is a safety net rather
// than a business limit: a cycle already present in stored data would otherwise loop forever.
const maxCategoryDepth = 100

// derefId turns an optional id into a plain string, so that "absent" and "empty" collapse into
// the single case the callers below already handle.
func derefId(id *string) string {
	if id == nil {
		return ""
	}
	return *id
}

func productCategoryEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.ProductCategorySchemaName,
		DefaultFields: []string{
			models.ProductCategoryFieldCode,
			models.ProductCategoryFieldName,
			models.ProductCategoryFieldParentCategoryId,
			models.ProductCategoryFieldSequence,
		},
		DefineActions: defineProductCategoryActions,
	}
}

// defineProductCategoryActions attaches the category-tree invariant to create and update. The
// CRUD processing itself is entirely the engine's; only the rule the schema cannot express
// belongs here.
func defineProductCategoryActions(engine drif.DynamicResourceEngine) error {
	err := engine.ModifyAction(drif.DynamicActionDelta{
		ActionName:    drif.ActionCreate,
		ValidateExtra: validateCategoryCreate(engine),
	})
	if err != nil {
		return errors.Wrap(err, "failed to attach product category create validation")
	}

	err = engine.ModifyAction(drif.DynamicActionDelta{
		ActionName:    drif.ActionUpdate,
		KeysToFetch:   categoryKeysToFetch,
		ValidateExtra: validateCategoryUpdate(engine),
	})
	return errors.Wrap(err, "failed to attach product category update validation")
}

func categoryKeysToFetch(params dmodel.DynamicFields) dmodel.DynamicFields {
	return dmodel.DynamicFields{models.ProductCategoryFieldId: params[models.ProductCategoryFieldId]}
}

// validateCategoryCreate rejects a category that would be its own ancestor. On create the record
// has no id yet, so only the parent chain itself can be malformed.
func validateCategoryCreate(engine drif.DynamicResourceEngine) drif.ActionValidateExtraFn {
	return func(
		ctx corectx.Context, params dmodel.DynamicFields, _ *dmodel.DynamicFields, vErrs *ft.ClientErrors,
	) error {
		category := models.NewProductCategoryFrom(params)
		return assertNoCategoryCycle(ctx, engine, derefId(category.GetParentCategoryId()), "", vErrs)
	}
}

// validateCategoryUpdate additionally rejects re-parenting a category under one of its own
// descendants, which is the way a cycle is normally introduced. See BR §6.4.3.
func validateCategoryUpdate(engine drif.DynamicResourceEngine) drif.ActionValidateExtraFn {
	return func(
		ctx corectx.Context, params dmodel.DynamicFields, foundModel *dmodel.DynamicFields, vErrs *ft.ClientErrors,
	) error {
		if foundModel == nil {
			return nil
		}
		submitted := models.NewProductCategoryFrom(params)
		parentId := submitted.GetParentCategoryId()
		if parentId == nil {
			// The parent is not being changed, so the existing chain still holds.
			return nil
		}

		stored := models.NewProductCategoryFrom(*foundModel)
		selfId := derefId(stored.GetId())

		if derefId(parentId) == selfId {
			vErrs.Append(*ft.NewBusinessViolation(models.ProductCategoryFieldParentCategoryId,
				"product_category.self_parent",
				"a category cannot be its own parent"))
			return nil
		}
		return assertNoCategoryCycle(ctx, engine, derefId(parentId), selfId, vErrs)
	}
}

// assertNoCategoryCycle walks upwards from parentId. Reaching selfId means the proposed parent
// is a descendant of the category being edited, so the edit would close a loop.
func assertNoCategoryCycle(
	ctx corectx.Context,
	engine drif.DynamicResourceEngine,
	parentId string,
	selfId string,
	vErrs *ft.ClientErrors,
) error {
	if parentId == "" {
		return nil
	}

	seen := map[string]bool{}
	current := parentId
	for depth := 0; current != "" && depth < maxCategoryDepth; depth++ {
		if current == selfId {
			vErrs.Append(*ft.NewBusinessViolation(models.ProductCategoryFieldParentCategoryId,
				"product_category.cycle",
				"this parent is a descendant of the category, which would create a cycle"))
			return nil
		}
		if seen[current] {
			// A pre-existing cycle in stored data. Stop rather than loop, and let the edit
			// through: it is not the change under validation that introduced it.
			return nil
		}
		seen[current] = true

		found, err := engine.ResourceRepository().GetOne(ctx, dyn.RepoGetOneParam{
			Filter: dmodel.DynamicFields{models.ProductCategoryFieldId: current},
			Fields: []string{models.ProductCategoryFieldId, models.ProductCategoryFieldParentCategoryId},
		})
		if err != nil {
			return errors.Wrap(err, "assertNoCategoryCycle")
		}
		if !found.HasData {
			// A missing parent is the reference check's business, not ours.
			return nil
		}

		current = derefId(models.NewProductCategoryFrom(found.Data).GetParentCategoryId())
	}
	return nil
}
