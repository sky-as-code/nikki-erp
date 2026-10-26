package services

import (
	"fmt"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
	itProduct "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/product"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// desiredLine is one validated line of the desired configuration. row is its index in the
// payload, or -1 when the payload left the lines unchanged and the line was read from the
// template.
type desiredLine struct {
	attributeId string
	mode        models.VariantCreationMode
	valueIds    []string
	sequence    *int32
	row         int
}

// desiredConfig is the configuration a save asks for, validated against the catalogue.
type desiredConfig struct {
	lines      []desiredLine
	exclusions []PlanExclusion
}

func (this desiredConfig) planConfig() PlanConfig {
	config := PlanConfig{Exclusions: this.exclusions}
	for _, line := range this.lines {
		config.Lines = append(config.Lines, PlanLine{AttributeId: line.attributeId, Mode: line.mode, ValueIds: line.valueIds})
	}
	return config
}

func (this desiredConfig) rowOfAttribute(attributeId string) int {
	for _, line := range this.lines {
		if line.attributeId == attributeId {
			return line.row
		}
	}
	return -1
}

func linePath(row int, field string) string {
	if row < 0 {
		return itProduct.ProductTemplateFieldAttributeLines
	}
	return fmt.Sprintf("%s[%d].%s", itProduct.ProductTemplateFieldAttributeLines, row, field)
}

func exclusionPath(row int) string {
	return fmt.Sprintf("%s[%d]", itProduct.ProductTemplateFieldExclusions, row)
}

func configViolation(field string, key string, message string, vars ...map[string]any) ft.ClientErrorItem {
	return *ft.NewBusinessViolation(field, ft.ErrorKey(key, "inventory"), message, vars...)
}

// validateConfigurationShape checks what needs no read: sizes, required ids, an attribute named
// twice, a value named twice within a line, and empty lines (dropping every value of an attribute
// is done by removing its line).
func validateConfigurationShape(input itProduct.TemplateConfigurationInput) ft.ClientErrors {
	cErrs := ft.ClientErrors{}
	if input.AttributeLines != nil {
		validateLinesShape(*input.AttributeLines, &cErrs)
	}
	if input.Exclusions != nil && len(*input.Exclusions) > itProduct.ProductTemplateConfigurationMaxExclusions {
		cErrs.Append(configViolation(itProduct.ProductTemplateFieldExclusions, "product_template.too_many_exclusions",
			"at most {{max}} exclusions can be sent", map[string]any{"max": itProduct.ProductTemplateConfigurationMaxExclusions}))
	}
	return cErrs
}

func validateLinesShape(lines []itProduct.AttributeLineInput, cErrs *ft.ClientErrors) {
	if len(lines) > itProduct.ProductTemplateConfigurationMaxLines {
		cErrs.Append(configViolation(itProduct.ProductTemplateFieldAttributeLines, "product_template.too_many_lines",
			"at most {{max}} attribute lines can be sent", map[string]any{"max": itProduct.ProductTemplateConfigurationMaxLines}))
		return
	}
	seenAttributes := map[string]bool{}
	for row, line := range lines {
		if line.AttributeId == "" {
			cErrs.Append(*ft.NewValidationError(linePath(row, "attribute_id"),
				ft.ErrorKey("err_missing_required_field"), "attribute_id is required"))
		} else if seenAttributes[line.AttributeId] {
			cErrs.Append(configViolation(linePath(row, "attribute_id"), ReasonPlanDuplicateLine,
				"attribute {{attribute_id}} appears on more than one line", map[string]any{"attribute_id": line.AttributeId}))
		}
		seenAttributes[line.AttributeId] = true
		validateLineValuesShape(row, line.ValueIds, cErrs)
	}
}

func validateLineValuesShape(row int, valueIds []string, cErrs *ft.ClientErrors) {
	if len(valueIds) == 0 {
		cErrs.Append(configViolation(linePath(row, "value_ids"), "product_template.line_without_values",
			"select at least one value, or remove the attribute line"))
		return
	}
	seen := map[string]bool{}
	for _, valueId := range valueIds {
		if valueId == "" || seen[valueId] {
			cErrs.Append(configViolation(linePath(row, "value_ids"), "product_template.duplicate_line_value",
				"each value may be selected once and must not be empty"))
			return
		}
		seen[valueId] = true
	}
}

// catalogueRows are the attributes and values a payload names, archived ones included.
type catalogueRows struct {
	attributes map[string]dmodel.DynamicFields
	values     map[string]dmodel.DynamicFields
}

func loadCatalogueRows(ctx corectx.Context, lines []itProduct.AttributeLineInput) (*catalogueRows, error) {
	attributeIds, valueIds := []string{}, []string{}
	for _, line := range lines {
		attributeIds = append(attributeIds, line.AttributeId)
		valueIds = append(valueIds, line.ValueIds...)
	}
	rows := &catalogueRows{}
	var err error
	if rows.attributes, err = indexRowsById(ctx, models.ProductAttributeSchemaName, attributeIds); err != nil {
		return nil, err
	}
	rows.values, err = indexRowsById(ctx, models.ProductAttributeValueSchemaName, valueIds)
	return rows, err
}

func indexRowsById(ctx corectx.Context, schemaName string, ids []string) (map[string]dmodel.DynamicFields, error) {
	index := map[string]dmodel.DynamicFields{}
	if len(ids) == 0 {
		return index, nil
	}
	rows, err := searchConfigRows(ctx, schemaName, basemodel.FieldId, ids)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		index[readStringParam(row, basemodel.FieldId)] = row
	}
	return index, nil
}

// resolveDesiredLines validates every payload line against the catalogue and the template's
// current lines. A new selection must not name an archived attribute or value; a selection the
// template already holds stays allowed, so archiving the catalogue never forces a template edit.
func resolveDesiredLines(
	ctx corectx.Context, state *templateConfigState, orgId string, lines []itProduct.AttributeLineInput,
) ([]desiredLine, ft.ClientErrors, error) {
	catalogue, err := loadCatalogueRows(ctx, lines)
	if err != nil {
		return nil, nil, err
	}
	cErrs := ft.ClientErrors{}
	out := make([]desiredLine, 0, len(lines))
	for row, line := range lines {
		current := state.enabledLine(line.AttributeId)
		mode, ok := resolveLineAttribute(catalogue, current, orgId, row, line.AttributeId, &cErrs)
		if !ok {
			continue
		}
		resolveLineValues(catalogue, current, orgId, row, line, &cErrs)
		out = append(out, desiredLine{
			attributeId: line.AttributeId, mode: mode, valueIds: line.ValueIds, sequence: line.Sequence, row: row,
		})
	}
	return out, cErrs, nil
}

func resolveLineAttribute(
	catalogue *catalogueRows, current *configLineRow, orgId string, row int, attributeId string, cErrs *ft.ClientErrors,
) (models.VariantCreationMode, bool) {
	field := linePath(row, "attribute_id")
	found := catalogue.attributes[attributeId]
	if found == nil || readStringParam(found, models.ProductAttributeFieldOrgId) != orgId {
		cErrs.Append(*ft.NewNotFoundError(field))
		return "", false
	}
	attribute := models.NewProductAttributeFrom(found)
	if current == nil && isArchivedRow(found) {
		cErrs.Append(configViolation(field, "product_template.attribute_archived",
			"an archived attribute cannot be added to a template"))
		return "", false
	}
	mode := models.VariantCreationModeInstant
	if current != nil && current.snapshot != "" {
		mode = current.snapshot
	} else if stored := attribute.GetVariantCreationMode(); stored != nil {
		mode = *stored
	}
	dataType := attribute.GetDataType()
	if mode.CreatesVariants() && (dataType == nil || *dataType != models.AttributeDataTypeOption) {
		cErrs.Append(configViolation(field, "product_template.attribute_not_option",
			"only an attribute of data type option can create variants"))
		return "", false
	}
	return mode, true
}

func resolveLineValues(
	catalogue *catalogueRows, current *configLineRow, orgId string, row int,
	line itProduct.AttributeLineInput, cErrs *ft.ClientErrors,
) {
	field := linePath(row, "value_ids")
	for _, valueId := range line.ValueIds {
		found := catalogue.values[valueId]
		vars := map[string]any{"value_id": valueId}
		switch {
		case found == nil || readStringParam(found, models.ProductAttributeValueFieldOrgId) != orgId:
			cErrs.Append(configViolation(field, "product_template.value_not_found", "value {{value_id}} does not exist", vars))
		case readStringParam(found, models.ProductAttributeValueFieldAttributeId) != line.AttributeId:
			cErrs.Append(configViolation(field, "product_template.value_not_of_attribute",
				"value {{value_id}} does not belong to this attribute", vars))
		case isArchivedRow(found) && !isCurrentlySelected(current, valueId):
			cErrs.Append(configViolation(field, "product_template.value_archived",
				"archived value {{value_id}} cannot be newly selected", vars))
		}
	}
}

func isCurrentlySelected(line *configLineRow, valueId string) bool {
	if line == nil {
		return false
	}
	value := line.valueByCatalogueId(valueId)
	return value != nil && value.selected
}

func isArchivedRow(row dmodel.DynamicFields) bool {
	archived := row.GetBool(basemodel.FieldIsArchived)
	return archived != nil && *archived
}

// currentDesiredLines is the template's enabled configuration expressed as desired lines, for a
// save that leaves attribute_lines unchanged.
func currentDesiredLines(state *templateConfigState) []desiredLine {
	out := []desiredLine{}
	for _, line := range state.enabledLines() {
		out = append(out, desiredLine{
			attributeId: line.attributeId, mode: line.snapshot, valueIds: line.selectedValueIds(),
			sequence: line.sequence, row: -1,
		})
	}
	return out
}

// resolveDesiredExclusions validates the exclusion payload: two different values, both selected
// on the desired lines, of two different attributes. Pairs are normalized and deduplicated.
func resolveDesiredExclusions(lines []desiredLine, exclusions []itProduct.ExclusionInput) ([]PlanExclusion, ft.ClientErrors) {
	owner := map[string]string{}
	for _, line := range lines {
		for _, valueId := range line.valueIds {
			owner[valueId] = line.attributeId
		}
	}
	cErrs := ft.ClientErrors{}
	seen := map[string]bool{}
	out := []PlanExclusion{}
	for row, exclusion := range exclusions {
		left, right := NormalizeExclusionPair(exclusion.LeftValueId, exclusion.RightValueId)
		leftOwner, leftOk := owner[left]
		rightOwner, rightOk := owner[right]
		switch {
		case left == "" || left == right:
			cErrs.Append(configViolation(exclusionPath(row), "product_template_combination_exclusion.same_value",
				"an exclusion pairs two different values"))
		case !leftOk || !rightOk:
			cErrs.Append(configViolation(exclusionPath(row), "product_template.exclusion_value_not_selected",
				"both values of an exclusion must be selected on the template"))
		case leftOwner == rightOwner:
			cErrs.Append(configViolation(exclusionPath(row), "product_template_combination_exclusion.same_attribute",
				"an exclusion pairs values of two different attributes"))
		case !seen[left+"|"+right]:
			seen[left+"|"+right] = true
			out = append(out, PlanExclusion{LeftValueId: left, RightValueId: right})
		}
	}
	return out, cErrs
}

// resolveDesiredConfig turns the lifted payload into the validated desired configuration. A part
// the payload omits is read from the template's current configuration.
func resolveDesiredConfig(
	ctx corectx.Context, state *templateConfigState, orgId string, input itProduct.TemplateConfigurationInput,
) (*desiredConfig, ft.ClientErrors, error) {
	desired := &desiredConfig{lines: currentDesiredLines(state)}
	if input.AttributeLines != nil {
		lines, cErrs, err := resolveDesiredLines(ctx, state, orgId, *input.AttributeLines)
		if err != nil || cErrs.Count() > 0 {
			return nil, cErrs, err
		}
		desired.lines = lines
	}
	if input.Exclusions == nil {
		desired.exclusions = keepSelectedExclusions(state.planConfig().Exclusions, desired.lines)
		return desired, nil, nil
	}
	exclusions, cErrs := resolveDesiredExclusions(desired.lines, *input.Exclusions)
	if cErrs.Count() > 0 {
		return nil, cErrs, nil
	}
	desired.exclusions = exclusions
	return desired, nil, nil
}

// keepSelectedExclusions keeps the stored exclusions whose two values stay selected: an exclusion
// cannot outlive a value the lines no longer select, since it is stored against that selection.
func keepSelectedExclusions(exclusions []PlanExclusion, lines []desiredLine) []PlanExclusion {
	selected := map[string]bool{}
	for _, line := range lines {
		for _, valueId := range line.valueIds {
			selected[valueId] = true
		}
	}
	out := []PlanExclusion{}
	for _, exclusion := range exclusions {
		if selected[exclusion.LeftValueId] && selected[exclusion.RightValueId] {
			out = append(out, exclusion)
		}
	}
	return out
}
