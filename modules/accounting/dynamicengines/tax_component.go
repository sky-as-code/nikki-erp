package dynamicengines

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/accounting/domain/models"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
)

// maxComponentDepth bounds the transitive walk through group taxes.
//
// Real compositions are shallow — a total tax made of an excise and a VAT is two levels — so eight
// is generous. As with jurisdictions the bound is a termination guarantee rather than a business
// rule: a cycle written directly into the database must stop the walk rather than hang the request.
const maxComponentDepth = 8

func taxComponentEngineSpec() engineSpec {
	return engineSpec{
		SchemaName:    models.TaxComponentSchemaName,
		DefineActions: defineComponentActions,
	}
}

func defineComponentActions(engine drif.DynamicResourceEngine) error {
	err := engine.ModifyAction(drif.DynamicActionDelta{
		ActionName:    drif.ActionCreate,
		ValidateExtra: validateComponentCreate(engine),
	})
	if err != nil {
		return errors.Wrap(err, "failed to attach tax component create validation")
	}

	err = engine.ModifyAction(drif.DynamicActionDelta{
		ActionName:    drif.ActionUpdate,
		KeysToFetch:   componentKeysToFetch,
		ValidateExtra: validateComponentUpdate(engine),
	})
	return errors.Wrap(err, "failed to attach tax component update validation")
}

func componentKeysToFetch(params dmodel.DynamicFields) dmodel.DynamicFields {
	return dmodel.DynamicFields{
		models.TaxComponentFieldId: params[models.TaxComponentFieldId],
	}
}

func validateComponentCreate(engine drif.DynamicResourceEngine) drif.ActionValidateExtraFn {
	return func(
		ctx corectx.Context, inputModel *drif.DynamicEntity, _ *drif.DynamicEntity, vErrs *ft.ClientErrors,
	) error {
		params := inputModel.GetFieldData()
		component := models.NewTaxComponentFrom(params)
		return assertComponentAcyclic(ctx, engine, component, vErrs)
	}
}

func validateComponentUpdate(engine drif.DynamicResourceEngine) drif.ActionValidateExtraFn {
	return func(
		ctx corectx.Context, inputModel *drif.DynamicEntity, foundModel *drif.DynamicEntity, vErrs *ft.ClientErrors,
	) error {
		params := inputModel.GetFieldData()
		found := entityFields(foundModel)
		if err := assertParentVersionEditable(ctx, engine, found, vErrs); err != nil {
			return err
		}
		if vErrs.Count() > 0 {
			return nil
		}
		component := models.NewTaxComponentFrom(mergeFields(found, params))
		return assertComponentAcyclic(ctx, engine, component, vErrs)
	}
}

// assertParentVersionEditable refuses to change a component of a published definition version.
//
// A component has no lifecycle of its own — BR-TAX-ESS-SUP-005 is explicit that it inherits the
// parent's. Without this check the immutability of a published version would be trivially
// bypassable: the version's own fields would stay frozen while the set of taxes it is composed of
// changed underneath them, which alters the computed total just as effectively as editing a rate.
func assertParentVersionEditable(
	ctx corectx.Context,
	engine drif.DynamicResourceEngine,
	found *dmodel.DynamicFields,
	vErrs *ft.ClientErrors,
) error {
	if found == nil {
		return nil
	}
	parentId := models.NewTaxComponentFrom(*found).GetParentTaxDefinitionVersionId()
	if parentId == nil {
		return nil
	}

	parent, err := models.FindDefinitionVersionById(ctx, engine.ResourceRepository(), *parentId)
	if err != nil {
		return errors.Wrap(err, "assertParentVersionEditable")
	}
	if parent == nil {
		return nil
	}

	status := parent.GetLifecycleStatus()
	if status != nil && *status != string(models.LifecycleDraft) {
		vErrs.Append(*ft.NewBusinessViolation(
			models.TaxComponentFieldParentTaxDefinitionVersionId,
			"tax.component_parent_not_draft",
			"the tax definition version this component belongs to is no longer a draft; "+
				"create a new version to change what the tax is composed of"))
	}
	return nil
}

// assertComponentAcyclic rejects a composition in which a tax contains itself, directly or through
// any chain of group taxes.
//
// The walk is over taxes rather than components: a component names a child tax, and that tax's own
// published definition version may itself be a group with components of its own. Following only
// the direct children would catch "A contains A" and miss "A contains B contains A", which is the
// case that actually reaches production, because nobody writes the first one by hand.
func assertComponentAcyclic(
	ctx corectx.Context,
	engine drif.DynamicResourceEngine,
	component *models.TaxComponent,
	vErrs *ft.ClientErrors,
) error {
	parentVersionId := component.GetParentTaxDefinitionVersionId()
	childTaxId := component.GetComponentTaxId()
	if parentVersionId == nil || childTaxId == nil {
		return nil
	}

	parent, err := models.FindDefinitionVersionById(ctx, engine.ResourceRepository(), *parentVersionId)
	if err != nil {
		return errors.Wrap(err, "assertComponentAcyclic")
	}
	if parent == nil {
		return nil
	}
	ownerTaxId := parent.GetTaxId()
	if ownerTaxId == nil {
		return nil
	}

	if *childTaxId == *ownerTaxId {
		vErrs.Append(*ft.NewBusinessViolation(models.TaxComponentFieldComponentTaxId,
			"tax.component_self_reference",
			"a tax cannot be a component of itself"))
		return nil
	}

	reaches, err := taxReaches(ctx, engine, *childTaxId, *ownerTaxId)
	if err != nil {
		return err
	}
	if reaches {
		vErrs.Append(*ft.NewBusinessViolation(models.TaxComponentFieldComponentTaxId,
			"tax.component_cycle",
			"this component already contains the tax being composed, so the composition "+
				"would form a cycle"))
	}
	return nil
}

// taxReaches reports whether targetTaxId is reachable from startTaxId by following components.
func taxReaches(
	ctx corectx.Context, engine drif.DynamicResourceEngine, startTaxId string, targetTaxId string,
) (bool, error) {
	repo := engine.ResourceRepository()
	visited := map[string]bool{startTaxId: true}
	frontier := []string{startTaxId}

	for depth := 0; depth < maxComponentDepth && len(frontier) > 0; depth++ {
		next := []string{}
		for _, taxId := range frontier {
			versions, err := models.FindVersionsOfTaxAnyStatus(ctx, repo, taxId, 50)
			if err != nil {
				return false, errors.Wrap(err, "taxReaches")
			}
			for _, versionFields := range versions {
				version := models.NewTaxDefinitionVersionFrom(versionFields)
				versionId := version.GetId()
				if versionId == nil {
					continue
				}
				components, err := models.FindComponentsOfVersion(ctx, repo, *versionId, 100)
				if err != nil {
					return false, errors.Wrap(err, "taxReaches")
				}
				for _, componentFields := range components {
					childId := models.NewTaxComponentFrom(componentFields).GetComponentTaxId()
					if childId == nil {
						continue
					}
					if *childId == targetTaxId {
						return true, nil
					}
					if !visited[*childId] {
						visited[*childId] = true
						next = append(next, *childId)
					}
				}
			}
		}
		frontier = next
	}
	return false, nil
}
