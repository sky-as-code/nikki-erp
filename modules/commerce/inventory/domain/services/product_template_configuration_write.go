package services

import (
	"fmt"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

// Writing a template's configuration rows. Lines and values are never deleted: a removed attribute
// disables its line (keeping its mode snapshot), an unselected value keeps its row with
// is_selected=false, and re-adding a removed attribute opens a new revision. Exclusions are
// archived and unarchived. Every write runs in the template's transaction.

// configWriter writes the delta between a template's stored configuration and the desired one.
type configWriter struct {
	templateId string
	orgId      string
	state      *templateConfigState
}

// writeConfiguration applies the desired lines, then reloads the state so the exclusions and the
// variant join rows can name the template value rows just written.
func writeConfiguration(
	ctx corectx.Context, writer configWriter, desired *desiredConfig,
) (*templateConfigState, ft.ClientErrors, error) {
	if cErrs, err := writer.writeLines(ctx, desired.lines); err != nil || cErrs.Count() > 0 {
		return nil, cErrs, err
	}
	state, err := loadTemplateConfigState(ctx, writer.templateId)
	if err != nil {
		return nil, nil, err
	}
	writer.state = state
	cErrs, err := writer.writeExclusions(ctx, desired.exclusions)
	return state, cErrs, err
}

func (this configWriter) writeLines(ctx corectx.Context, lines []desiredLine) (ft.ClientErrors, error) {
	wanted := map[string]bool{}
	for _, line := range lines {
		wanted[line.attributeId] = true
	}
	for _, line := range this.state.lines {
		if line.enabled && !wanted[line.attributeId] {
			if err := updateConfigRow(ctx, models.ProductTemplateAttributeSchemaName, dmodel.DynamicFields{
				basemodel.FieldId: line.id, models.ProductTemplateAttributeFieldIsEnabled: false,
			}); err != nil {
				return nil, err
			}
		}
	}
	for _, line := range lines {
		current := this.state.enabledLine(line.attributeId)
		var cErrs ft.ClientErrors
		var err error
		if current == nil {
			cErrs, err = this.createLine(ctx, line)
		} else {
			cErrs, err = this.updateLine(ctx, current, line)
		}
		if err != nil || cErrs.Count() > 0 {
			return prefixLineErrors(cErrs, line.row), err
		}
	}
	return nil, nil
}

// createLine applies an attribute that has no enabled line: revision max+1 and a fresh snapshot
// of the attribute's mode, never a revived history line (CR02 §8).
func (this configWriter) createLine(ctx corectx.Context, line desiredLine) (ft.ClientErrors, error) {
	lineSvc, err := domainServiceFor(models.ProductTemplateAttributeSchemaName)
	if err != nil {
		return nil, err
	}
	fields := dmodel.DynamicFields{
		models.ProductTemplateAttributeFieldProductTemplateId: this.templateId,
		models.ProductTemplateAttributeFieldAttributeId:       line.attributeId,
		models.ProductTemplateAttributeFieldIsEnabled:         true,
		models.ProductTemplateAttributeFieldRevision:          this.state.maxRevision(line.attributeId) + 1,
		models.ProductTemplateAttributeFieldModeSnapshot:      line.mode.String(),
		models.ProductTemplateAttributeFieldOrgId:             this.orgId,
	}
	if line.sequence != nil {
		fields[models.ProductTemplateAttributeFieldSequence] = *line.sequence
	}
	created, err := lineSvc.Create(ctx, fields)
	if err != nil {
		return nil, errors.Wrap(err, "createLine")
	}
	if created.ClientErrors.Count() > 0 || !created.HasData {
		return created.ClientErrors, nil
	}
	lineId := readStringParam(created.Data, basemodel.FieldId)
	for index, valueId := range line.valueIds {
		if cErrs, err := this.createLineValue(ctx, lineId, valueId, index); err != nil || cErrs.Count() > 0 {
			return cErrs, err
		}
	}
	return nil, nil
}

// updateLine keeps the line and moves only what changed: its sequence and which value rows are
// selected. A value never selected on this line before gets a new row.
func (this configWriter) updateLine(ctx corectx.Context, current *configLineRow, line desiredLine) (ft.ClientErrors, error) {
	if line.sequence != nil && (current.sequence == nil || *current.sequence != *line.sequence) {
		if err := updateConfigRow(ctx, models.ProductTemplateAttributeSchemaName, dmodel.DynamicFields{
			basemodel.FieldId: current.id, models.ProductTemplateAttributeFieldSequence: *line.sequence,
		}); err != nil {
			return nil, err
		}
	}
	wanted := map[string]bool{}
	for _, valueId := range line.valueIds {
		wanted[valueId] = true
	}
	for _, value := range current.values {
		if value.selected != wanted[value.valueId] {
			if err := updateConfigRow(ctx, models.ProductTemplateAttributeValueSchemaName, dmodel.DynamicFields{
				basemodel.FieldId: value.id, models.ProductTemplateAttributeValueFieldIsSelected: wanted[value.valueId],
			}); err != nil {
				return nil, err
			}
		}
	}
	for index, valueId := range line.valueIds {
		if current.valueByCatalogueId(valueId) != nil {
			continue
		}
		if cErrs, err := this.createLineValue(ctx, current.id, valueId, index+len(current.values)); err != nil || cErrs.Count() > 0 {
			return cErrs, err
		}
	}
	return nil, nil
}

func (this configWriter) createLineValue(
	ctx corectx.Context, lineId string, valueId string, index int,
) (ft.ClientErrors, error) {
	valueSvc, err := domainServiceFor(models.ProductTemplateAttributeValueSchemaName)
	if err != nil {
		return nil, err
	}
	created, err := valueSvc.Create(ctx, dmodel.DynamicFields{
		models.ProductTemplateAttributeValueFieldTemplateAttributeId: lineId,
		models.ProductTemplateAttributeValueFieldAttributeValueId:    valueId,
		models.ProductTemplateAttributeValueFieldIsSelected:          true,
		models.ProductTemplateAttributeValueFieldSequence:            int32(index + 1),
		models.ProductTemplateAttributeValueFieldOrgId:               this.orgId,
	})
	if err != nil {
		return nil, errors.Wrap(err, "createLineValue")
	}
	return created.ClientErrors, nil
}

// writeExclusions makes the active exclusions exactly the desired pairs. A stored pair is
// unarchived rather than recorded twice; the unique index spans archived rows.
func (this configWriter) writeExclusions(ctx corectx.Context, exclusions []PlanExclusion) (ft.ClientErrors, error) {
	refs := this.state.valueRefs()
	wanted := map[string]PlanExclusion{}
	for _, exclusion := range exclusions {
		left, right := NormalizeExclusionPair(refs[exclusion.LeftValueId].TemplateValueId, refs[exclusion.RightValueId].TemplateValueId)
		wanted[left+"|"+right] = PlanExclusion{LeftValueId: left, RightValueId: right}
	}
	for _, stored := range this.state.exclusions {
		key := stored.leftId + "|" + stored.rightId
		_, keep := wanted[key]
		delete(wanted, key)
		if keep == !stored.archived {
			continue
		}
		if cErrs, err := setExclusionArchived(ctx, stored, !keep); err != nil || cErrs.Count() > 0 {
			return cErrs, err
		}
	}
	for _, pair := range wanted {
		if cErrs, err := this.createExclusion(ctx, pair); err != nil || cErrs.Count() > 0 {
			return cErrs, err
		}
	}
	return nil, nil
}

func setExclusionArchived(ctx corectx.Context, stored *configExclusionRow, archived bool) (ft.ClientErrors, error) {
	exclusionSvc, err := domainServiceFor(models.ProductTemplateCombinationExclusionSchemaName)
	if err != nil {
		return nil, err
	}
	result, err := exclusionSvc.SetArchived(ctx, dmodel.DynamicFields{
		basemodel.FieldId: stored.id, basemodel.FieldEtag: stored.etag, basemodel.FieldIsArchived: archived,
	})
	if err != nil {
		return nil, errors.Wrap(err, "setExclusionArchived")
	}
	return prefixRowErrors(result.ClientErrors, itProduct.ProductTemplateFieldExclusions), nil
}

func (this configWriter) createExclusion(ctx corectx.Context, pair PlanExclusion) (ft.ClientErrors, error) {
	exclusionSvc, err := domainServiceFor(models.ProductTemplateCombinationExclusionSchemaName)
	if err != nil {
		return nil, err
	}
	created, err := exclusionSvc.Create(ctx, dmodel.DynamicFields{
		models.ProductTemplateCombinationExclusionFieldProductTemplateId: this.templateId,
		models.ProductTemplateCombinationExclusionFieldLeftValueId:       pair.LeftValueId,
		models.ProductTemplateCombinationExclusionFieldRightValueId:      pair.RightValueId,
		models.ProductTemplateCombinationExclusionFieldOrgId:             this.orgId,
	})
	if err != nil {
		return nil, errors.Wrap(err, "createExclusion")
	}
	return prefixRowErrors(created.ClientErrors, itProduct.ProductTemplateFieldExclusions), nil
}

// updateConfigRow writes a system-owned field of a configuration row through the repository: these
// rows have no client write path, so there is no etag to honour.
func updateConfigRow(ctx corectx.Context, schemaName string, fields dmodel.DynamicFields) error {
	repo, err := repoFor(schemaName)
	if err != nil {
		return err
	}
	result, err := repo.Update(ctx, fields)
	if err != nil {
		return errors.Wrap(err, "update "+schemaName)
	}
	if result.ClientErrors.Count() > 0 {
		return errors.Wrap(result.ClientErrors.ToError(), "update "+schemaName)
	}
	return nil
}

func prefixLineErrors(cErrs ft.ClientErrors, row int) ft.ClientErrors {
	if row < 0 {
		return prefixRowErrors(cErrs, itProduct.ProductTemplateFieldAttributeLines)
	}
	return prefixRowErrors(cErrs, fmt.Sprintf("%s[%d]", itProduct.ProductTemplateFieldAttributeLines, row))
}
