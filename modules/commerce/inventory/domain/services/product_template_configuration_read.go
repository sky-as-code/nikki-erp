package services

import (
	"sort"

	itProduct "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/product"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// Reads of the configuration for a page of templates, behind the template's function-kind
// computed field. It reads the whole page in one search per configuration table.

// TemplateConfigurationView is the read shape of a template's current configuration: the same
// fields a save sends, plus what the system keeps on each line.
type TemplateConfigurationView struct {
	AttributeLines []TemplateLineView         `json:"attribute_lines"`
	Exclusions     []itProduct.ExclusionInput `json:"exclusions"`
}

// TemplateLineView is one enabled line with its selected catalogue values in display order.
type TemplateLineView struct {
	Id                      string   `json:"id"`
	AttributeId             string   `json:"attribute_id"`
	ValueIds                []string `json:"value_ids"`
	Sequence                *int32   `json:"sequence"`
	VariantCreationSnapshot string   `json:"variant_creation_snapshot"`
	ApplicationRevision     int32    `json:"application_revision"`
}

// ReadTemplateConfigurations returns the current configuration of each template.
func ReadTemplateConfigurations(ctx corectx.Context, templateIds []string) (map[string]TemplateConfigurationView, error) {
	states, err := loadTemplateConfigStates(ctx, templateIds)
	if err != nil {
		return nil, err
	}
	views := map[string]TemplateConfigurationView{}
	for templateId, state := range states {
		views[templateId] = configurationView(state)
	}
	return views, nil
}

func configurationView(state *templateConfigState) TemplateConfigurationView {
	view := TemplateConfigurationView{AttributeLines: []TemplateLineView{}, Exclusions: []itProduct.ExclusionInput{}}
	for _, line := range state.enabledLines() {
		view.AttributeLines = append(view.AttributeLines, TemplateLineView{
			Id: line.id, AttributeId: line.attributeId, ValueIds: selectedInSequence(line),
			Sequence: line.sequence, VariantCreationSnapshot: line.snapshot.String(), ApplicationRevision: line.revision,
		})
	}
	for _, exclusion := range state.planConfig().Exclusions {
		view.Exclusions = append(view.Exclusions,
			itProduct.ExclusionInput{LeftValueId: exclusion.LeftValueId, RightValueId: exclusion.RightValueId})
	}
	return view
}

func selectedInSequence(line *configLineRow) []string {
	values := make([]*configValueRow, 0, len(line.values))
	for _, value := range line.values {
		if value.selected {
			values = append(values, value)
		}
	}
	sortValueRows(values)
	out := make([]string, 0, len(values))
	for _, value := range values {
		out = append(out, value.valueId)
	}
	return out
}

func sortValueRows(values []*configValueRow) {
	sequenceOf := func(value *configValueRow) int32 {
		if value.sequence == nil {
			return 0
		}
		return *value.sequence
	}
	sort.SliceStable(values, func(i, j int) bool { return sequenceOf(values[i]) < sequenceOf(values[j]) })
}
