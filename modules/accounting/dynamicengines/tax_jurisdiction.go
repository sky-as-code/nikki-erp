package dynamicengines

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/accounting/domain/models"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
)

// maxJurisdictionDepth bounds the parent walk so that a cycle already present in the database
// terminates the walk instead of hanging the request. A real hierarchy is four levels deep.
const maxJurisdictionDepth = 10

func taxJurisdictionEngineSpec() engineSpec {
	return engineSpec{
		SchemaName:    models.TaxJurisdictionSchemaName,
		DefineActions: defineJurisdictionActions,
	}
}

func defineJurisdictionActions(engine drif.DynamicResourceEngine) error {
	err := engine.ModifyAction(drif.DynamicActionDelta{
		ActionName:    drif.ActionCreate,
		ValidateExtra: validateJurisdictionCreate(engine),
	})
	if err != nil {
		return errors.Wrap(err, "failed to attach tax jurisdiction create validation")
	}

	err = engine.ModifyAction(drif.DynamicActionDelta{
		ActionName:    drif.ActionUpdate,
		KeysToFetch:   jurisdictionKeysToFetch,
		ValidateExtra: validateJurisdictionUpdate(engine),
	})
	return errors.Wrap(err, "failed to attach tax jurisdiction update validation")
}

func jurisdictionKeysToFetch(params dmodel.DynamicFields) dmodel.DynamicFields {
	return dmodel.DynamicFields{
		models.TaxJurisdictionFieldId: params[models.TaxJurisdictionFieldId],
	}
}

// validateJurisdictionCreate rejects a jurisdiction whose parent chain already forms a cycle. The
// new row has no id yet so cannot be in a cycle itself, but inheriting a broken parent chain would
// make every later walk through it non-terminating.
func validateJurisdictionCreate(engine drif.DynamicResourceEngine) drif.ActionValidateExtraFn {
	return func(
		ctx corectx.Context, inputModel *drif.DynamicEntity, _ *drif.DynamicEntity, vErrs *ft.ClientErrors,
	) error {
		params := inputModel.GetFieldData()
		juris := models.NewTaxJurisdictionFrom(params)
		return assertJurisdictionAcyclic(ctx, engine, juris.GetParentId(), nil, vErrs)
	}
}

// validateJurisdictionUpdate rejects a reparent that would put the record inside its own ancestry.
func validateJurisdictionUpdate(engine drif.DynamicResourceEngine) drif.ActionValidateExtraFn {
	return func(
		ctx corectx.Context, inputModel *drif.DynamicEntity, foundModel *drif.DynamicEntity, vErrs *ft.ClientErrors,
	) error {
		params := inputModel.GetFieldData()
		found := entityFields(foundModel)
		if _, submitted := params[models.TaxJurisdictionFieldParentId]; !submitted {
			return nil
		}
		juris := models.NewTaxJurisdictionFrom(params)

		var selfId *string
		if found != nil {
			existing := models.NewTaxJurisdictionFrom(*found)
			selfId = existing.GetId()
		}
		return assertJurisdictionAcyclic(ctx, engine, juris.GetParentId(), selfId, vErrs)
	}
}

// assertJurisdictionAcyclic walks from parentId up to the root, failing if it meets selfId or
// exceeds the depth bound. It walks one row at a time because the search graph cannot express
// "follow parent_id until null"; the chains are short and the check runs only when a parent is
// submitted.
func assertJurisdictionAcyclic(
	ctx corectx.Context,
	engine drif.DynamicResourceEngine,
	parentId *string,
	selfId *string,
	vErrs *ft.ClientErrors,
) error {
	if parentId == nil || *parentId == "" {
		return nil
	}

	// The one-step cycle gets its own check; the walk below would report it as a generic ancestry
	// violation.
	if selfId != nil && *parentId == *selfId {
		vErrs.Append(*ft.NewBusinessViolation(models.TaxJurisdictionFieldParentId,
			"tax.jurisdiction_self_parent",
			"a tax jurisdiction cannot be its own parent"))
		return nil
	}

	seen := map[string]bool{}
	current := *parentId
	for depth := 0; depth < maxJurisdictionDepth; depth++ {
		if selfId != nil && current == *selfId {
			vErrs.Append(*ft.NewBusinessViolation(models.TaxJurisdictionFieldParentId,
				"tax.jurisdiction_cycle",
				"this parent is a descendant of the jurisdiction being updated"))
			return nil
		}
		if seen[current] {
			vErrs.Append(*ft.NewBusinessViolation(models.TaxJurisdictionFieldParentId,
				"tax.jurisdiction_cycle",
				"the parent chain of this jurisdiction contains a cycle"))
			return nil
		}
		seen[current] = true

		parent, err := models.FindJurisdictionById(ctx, engine.ResourceRepository(), current)
		if err != nil {
			return errors.Wrap(err, "assertJurisdictionAcyclic")
		}
		// A missing parent is left to the engine's reference checking; this answers only cycles.
		if parent == nil {
			return nil
		}
		next := parent.GetParentId()
		if next == nil || *next == "" {
			return nil
		}
		current = *next
	}

	vErrs.Append(*ft.NewBusinessViolation(models.TaxJurisdictionFieldParentId,
		"tax.jurisdiction_too_deep",
		"the jurisdiction hierarchy is nested more deeply than is supported"))
	return nil
}
