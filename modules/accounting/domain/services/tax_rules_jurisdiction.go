package services

import (
	"go.bryk.io/pkg/errors"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/accounting/domain/models"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// maxJurisdictionDepth bounds the parent walk so that a cycle already present in the database
// terminates the walk instead of hanging the request. A real hierarchy is four levels deep.
const maxJurisdictionDepth = 10

// validateJurisdictionCreate rejects a jurisdiction whose parent chain already forms a cycle. The
// new row has no id yet so cannot be in a cycle itself, but inheriting a broken parent chain would
// make every later walk through it non-terminating.
func validateJurisdictionCreate(repo models.TaxSearcher) composable.ValidateExtraFn {
	return func(
		ctx corectx.Context, inputModel *composable.DynamicEntity, _ *composable.DynamicEntity, vErrs *ft.ClientErrors,
	) error {
		params := inputModel.GetFieldData()
		juris := models.NewTaxJurisdictionFrom(params)
		return assertJurisdictionAcyclic(ctx, repo, juris.GetParentId(), nil, vErrs)
	}
}

// validateJurisdictionUpdate rejects a reparent that would put the record inside its own ancestry.
func validateJurisdictionUpdate(repo models.TaxSearcher) composable.ValidateExtraFn {
	return func(
		ctx corectx.Context, inputModel *composable.DynamicEntity, foundModel *composable.DynamicEntity,
		vErrs *ft.ClientErrors,
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
		return assertJurisdictionAcyclic(ctx, repo, juris.GetParentId(), selfId, vErrs)
	}
}

// assertJurisdictionAcyclic walks from parentId up to the root, failing if it meets selfId or
// exceeds the depth bound. It walks one row at a time because the search graph cannot express
// "follow parent_id until null"; the chains are short and the check runs only when a parent is
// submitted.
func assertJurisdictionAcyclic(
	ctx corectx.Context,
	repo models.TaxSearcher,
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
			ft.ErrorKey("tax.jurisdiction_self_parent", "accounting"),
			"a tax jurisdiction cannot be its own parent"))
		return nil
	}

	seen := map[string]bool{}
	current := *parentId
	for depth := 0; depth < maxJurisdictionDepth; depth++ {
		if selfId != nil && current == *selfId {
			vErrs.Append(*ft.NewBusinessViolation(models.TaxJurisdictionFieldParentId,
				ft.ErrorKey("tax.jurisdiction_cycle", "accounting"),
				"this parent is a descendant of the jurisdiction being updated"))
			return nil
		}
		if seen[current] {
			vErrs.Append(*ft.NewBusinessViolation(models.TaxJurisdictionFieldParentId,
				ft.ErrorKey("tax.jurisdiction_cycle", "accounting"),
				"the parent chain of this jurisdiction contains a cycle"))
			return nil
		}
		seen[current] = true

		parent, err := models.FindJurisdictionById(ctx, repo, current)
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
		ft.ErrorKey("tax.jurisdiction_too_deep", "accounting"),
		"the jurisdiction hierarchy is nested more deeply than is supported"))
	return nil
}
