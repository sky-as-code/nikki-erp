package services

import (
	"sort"

	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// previewLabeler names preview items the way a variant's computed display_name does (see
// models.FormatVariantDisplayName): the template name, then the value names in line order. An
// existing record is named by the lines as stored, which is what its display_name reads today; a
// created record and an extended key are named by the lines the save asks for.
type previewLabeler struct {
	templateName string
	valueNames   map[string]string
	storedRank   map[string]int
	desiredRank  map[string]int
}

// newPreviewLabeler reads every value name the plan mentions in one search, whatever the number of
// items.
func newPreviewLabeler(ctx corectx.Context, outcome *reconcileOutcome) (*previewLabeler, error) {
	labeler := &previewLabeler{
		templateName: langJsonToString(outcome.template.GetName()),
		valueNames:   map[string]string{},
		storedRank:   storedLineRanks(outcome.state),
	}
	labeler.desiredRank = desiredLineRanks(outcome.desired, outcome.state)
	valueIds := planValueIds(outcome.plan)
	if len(valueIds) == 0 {
		return labeler, nil
	}
	rows, err := searchConfigRows(ctx, models.ProductAttributeValueSchemaName, basemodel.FieldId, valueIds)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		value := models.NewProductAttributeValueFrom(row)
		labeler.valueNames[derefId(value.GetId())] = langJsonToString(value.GetName())
	}
	return labeler, nil
}

// storedLabel names an existing record; fallbackName is its SKU, which display_name shows when the
// template has no name.
func (this *previewLabeler) storedLabel(key string, fallbackName string) string {
	name := this.templateName
	if name == "" {
		name = fallbackName
	}
	return models.FormatVariantDisplayName(name, this.orderedValueNames(key, this.storedRank))
}

func (this *previewLabeler) desiredLabel(key string) string {
	return models.FormatVariantDisplayName(this.templateName, this.orderedValueNames(key, this.desiredRank))
}

// orderedValueNames sorts the key's pairs by line rank; an attribute with no line sorts last, by
// id. A value the catalogue no longer holds is skipped, as display_name skips it.
func (this *previewLabeler) orderedValueNames(key string, rank map[string]int) []string {
	pairs := ParseCombinationKey(key)
	sort.SliceStable(pairs, func(a, b int) bool {
		rankA, okA := rank[pairs[a].AttributeId]
		rankB, okB := rank[pairs[b].AttributeId]
		if okA != okB {
			return okA
		}
		if !okA || rankA == rankB {
			return pairs[a].AttributeId < pairs[b].AttributeId
		}
		return rankA < rankB
	})
	names := make([]string, 0, len(pairs))
	for _, pair := range pairs {
		if name, ok := this.valueNames[pair.ValueId]; ok {
			names = append(names, name)
		}
	}
	return names
}

// storedLineRanks orders the stored lines by sequence, as display_name does. An attribute's
// enabled line wins over a disabled one left behind by an earlier save.
func storedLineRanks(state *templateConfigState) map[string]int {
	ranks := map[string]int{}
	if state == nil {
		return ranks
	}
	lines := append([]*configLineRow{}, state.lines...)
	sort.SliceStable(lines, func(a, b int) bool {
		return sequenceOrZero(lines[a].sequence) < sequenceOrZero(lines[b].sequence)
	})
	for i, line := range lines {
		if _, seen := ranks[line.attributeId]; !seen || line.enabled {
			ranks[line.attributeId] = i
		}
	}
	return ranks
}

// desiredLineRanks orders the requested lines by the sequence they will be saved with: the one
// sent, else the one already stored (the writer keeps it), with the payload order breaking ties.
func desiredLineRanks(desired *desiredConfig, state *templateConfigState) map[string]int {
	ranks := map[string]int{}
	if desired == nil {
		return ranks
	}
	lines := append([]desiredLine{}, desired.lines...)
	effective := func(line desiredLine) int32 {
		if line.sequence != nil {
			return *line.sequence
		}
		if state != nil {
			if stored := state.enabledLine(line.attributeId); stored != nil {
				return sequenceOrZero(stored.sequence)
			}
		}
		return 0
	}
	sort.SliceStable(lines, func(a, b int) bool { return effective(lines[a]) < effective(lines[b]) })
	for i, line := range lines {
		ranks[line.attributeId] = i
	}
	return ranks
}

func planValueIds(plan *VariantPlan) []string {
	seen := map[string]bool{}
	out := []string{}
	lists := [][]PlannedVariant{plan.Keep, plan.Extend, plan.Restore, plan.Create, plan.Invalid,
		plan.Conflict, plan.Untouched}
	for _, list := range lists {
		for _, item := range list {
			for _, pair := range append(ParseCombinationKey(item.Key), ParseCombinationKey(item.NewKey)...) {
				if pair.ValueId != "" && !seen[pair.ValueId] {
					seen[pair.ValueId] = true
					out = append(out, pair.ValueId)
				}
			}
		}
	}
	return out
}

func sequenceOrZero(sequence *int32) int32 {
	if sequence == nil {
		return 0
	}
	return *sequence
}
