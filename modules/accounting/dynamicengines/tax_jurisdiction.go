package dynamicengines

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/accounting/domain/models"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
)

// maxJurisdictionDepth bounds the parent walk.
//
// A real hierarchy is country -> state -> county -> city, so four levels; ten is generous. The
// bound exists so that a cycle already present in the database — written before this check, or by
// a direct SQL edit — terminates the walk instead of hanging the request.
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

// validateJurisdictionCreate rejects a jurisdiction whose parent chain already forms a cycle.
//
// On create the new row has no id yet, so it cannot be part of a cycle itself; what it can do is
// point at a parent whose own chain is already broken, and inheriting that would make every later
// walk through this row non-terminating.
func validateJurisdictionCreate(engine drif.DynamicResourceEngine) drif.ActionValidateExtraFn {
	return func(
		ctx corectx.Context, params dmodel.DynamicFields, _ *dmodel.DynamicFields, vErrs *ft.ClientErrors,
	) error {
		juris := models.NewTaxJurisdictionFrom(params)
		return assertJurisdictionAcyclic(ctx, engine, juris.GetParentId(), nil, vErrs)
	}
}

// validateJurisdictionUpdate rejects a reparent that would put the record inside its own ancestry.
func validateJurisdictionUpdate(engine drif.DynamicResourceEngine) drif.ActionValidateExtraFn {
	return func(
		ctx corectx.Context, params dmodel.DynamicFields, found *dmodel.DynamicFields, vErrs *ft.ClientErrors,
	) error {
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
// exceeds the depth bound.
//
// It walks one row at a time because the depth is not known in advance and the search graph cannot
// express "follow parent_id until it is null". The chains are short and the check runs only when a
// parent is actually submitted, so the extra round trips are bounded and rare.
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

	// A record that is its own parent is the one-step cycle, and worth its own check because the
	// walk below would otherwise report it as a generic ancestry violation.
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
		// A parent that does not exist is left to the engine's own reference checking; this
		// function answers only the question about cycles.
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
