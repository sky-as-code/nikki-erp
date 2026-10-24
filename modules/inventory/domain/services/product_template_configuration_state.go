package services

import (
	"sort"

	"github.com/shopspring/decimal"
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// configStateScanLimit bounds every read of one template's configuration rows. Lines, values and
// exclusions of a single template stay far below it; history (disabled lines, unselected values,
// archived exclusions) is included.
const configStateScanLimit = 5000

// configValueRow is one template value (product_template_attribute_value) of a line.
type configValueRow struct {
	id              string
	valueId         string
	selected        bool
	sequence        *int32
	salesPriceExtra *decimal.Decimal
}

// configLineRow is one template line with all its value rows, selected or not.
type configLineRow struct {
	id          string
	attributeId string
	enabled     bool
	revision    int32
	snapshot    models.VariantCreationMode
	sequence    *int32
	values      []*configValueRow
}

func (this *configLineRow) valueByCatalogueId(valueId string) *configValueRow {
	for _, value := range this.values {
		if value.valueId == valueId {
			return value
		}
	}
	return nil
}

func (this *configLineRow) selectedValueIds() []string {
	out := make([]string, 0, len(this.values))
	for _, value := range this.values {
		if value.selected {
			out = append(out, value.valueId)
		}
	}
	return out
}

// configExclusionRow is a stored exclusion; its ends are template value ids.
type configExclusionRow struct {
	id       string
	leftId   string
	rightId  string
	archived bool
	etag     string
}

// templateConfigState is everything one template's attribute configuration has ever stored: lines
// of every revision, every value row and every exclusion, archived ones included. Keys and the
// configuration payload speak catalogue ids; the rows speak template value ids, and tavIndex maps
// one to the other.
type templateConfigState struct {
	templateId string
	lines      []*configLineRow
	exclusions []*configExclusionRow
	tavIndex   map[string]tavInfo
}

// tavInfo resolves a template value id to its catalogue value and line.
type tavInfo struct {
	valueId string
	line    *configLineRow
}

// enabledLine is the line currently applying attributeId, or nil.
func (this *templateConfigState) enabledLine(attributeId string) *configLineRow {
	for _, line := range this.lines {
		if line.enabled && line.attributeId == attributeId {
			return line
		}
	}
	return nil
}

func (this *templateConfigState) maxRevision(attributeId string) int32 {
	maxRev := int32(0)
	for _, line := range this.lines {
		if line.attributeId == attributeId && line.revision > maxRev {
			maxRev = line.revision
		}
	}
	return maxRev
}

func (this *templateConfigState) enabledLines() []*configLineRow {
	out := make([]*configLineRow, 0, len(this.lines))
	for _, line := range this.lines {
		if line.enabled {
			out = append(out, line)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return lineSequence(out[i]) < lineSequence(out[j]) })
	return out
}

func lineSequence(line *configLineRow) int32 {
	if line.sequence == nil {
		return 0
	}
	return *line.sequence
}

// planConfig is the configuration the existing variants were reconciled against: enabled lines
// with their snapshot mode and selected values, and the active exclusions in catalogue ids.
func (this *templateConfigState) planConfig() PlanConfig {
	config := PlanConfig{}
	for _, line := range this.enabledLines() {
		config.Lines = append(config.Lines, PlanLine{
			AttributeId: line.attributeId, Mode: line.snapshot, ValueIds: line.selectedValueIds(),
		})
	}
	for _, exclusion := range this.exclusions {
		if exclusion.archived {
			continue
		}
		left, leftOk := this.tavIndex[exclusion.leftId]
		right, rightOk := this.tavIndex[exclusion.rightId]
		if leftOk && rightOk {
			config.Exclusions = append(config.Exclusions,
				PlanExclusion{LeftValueId: left.valueId, RightValueId: right.valueId})
		}
	}
	return config
}

// previousPlanConfig is planConfig, or nil when no enabled line generates variants: the planner
// reads nil as "nothing generated yet", the state a fresh or default-only template is in.
func (this *templateConfigState) previousPlanConfig() *PlanConfig {
	config := this.planConfig()
	for _, line := range config.Lines {
		if line.Mode.CreatesVariants() {
			return &config
		}
	}
	return nil
}

// valueRefs maps each catalogue value selected on an enabled line to its template value row.
func (this *templateConfigState) valueRefs() TemplateValueRefs {
	refs := TemplateValueRefs{}
	for _, line := range this.lines {
		if !line.enabled {
			continue
		}
		for _, value := range line.values {
			if value.selected {
				refs[value.valueId] = TemplateValueRef{
					TemplateValueId: value.id, AttributeId: line.attributeId, SalesPriceExtra: value.salesPriceExtra,
				}
			}
		}
	}
	return refs
}

// loadTemplateConfigState reads the whole configuration history of one template in three
// bounded searches.
func loadTemplateConfigState(ctx corectx.Context, templateId string) (*templateConfigState, error) {
	states, err := loadTemplateConfigStates(ctx, []string{templateId})
	if err != nil {
		return nil, err
	}
	return states[templateId], nil
}

// loadTemplateConfigStates reads the configuration of several templates together, for a page of
// computed fields: one search per table whatever the page size.
func loadTemplateConfigStates(ctx corectx.Context, templateIds []string) (map[string]*templateConfigState, error) {
	states := map[string]*templateConfigState{}
	for _, templateId := range templateIds {
		states[templateId] = &templateConfigState{templateId: templateId, tavIndex: map[string]tavInfo{}}
	}
	if len(templateIds) == 0 {
		return states, nil
	}
	lineById, err := loadConfigLines(ctx, states, templateIds)
	if err != nil {
		return nil, err
	}
	if err := loadConfigValues(ctx, states, lineById); err != nil {
		return nil, err
	}
	return states, loadConfigExclusions(ctx, states, templateIds)
}

func loadConfigLines(
	ctx corectx.Context, states map[string]*templateConfigState, templateIds []string,
) (map[string]*configLineRow, error) {
	rows, err := searchConfigRows(ctx, models.ProductTemplateAttributeSchemaName,
		models.ProductTemplateAttributeFieldProductTemplateId, templateIds)
	if err != nil {
		return nil, err
	}
	lineById := map[string]*configLineRow{}
	for _, row := range rows {
		typed := models.NewProductTemplateAttributeFrom(row)
		line := &configLineRow{
			id:          derefId(typed.GetId()),
			attributeId: derefId(typed.GetAttributeId()),
			enabled:     typed.GetIsEnabled() == nil || *typed.GetIsEnabled(),
			revision:    1,
			sequence:    typed.GetSequence(),
		}
		if revision := typed.GetApplicationRevision(); revision != nil {
			line.revision = *revision
		}
		if snapshot := typed.GetVariantCreationSnapshot(); snapshot != nil {
			line.snapshot = *snapshot
		}
		state := states[derefId(typed.GetProductTemplateId())]
		if state == nil {
			continue
		}
		state.lines = append(state.lines, line)
		lineById[line.id] = line
	}
	return lineById, nil
}

func loadConfigValues(
	ctx corectx.Context, states map[string]*templateConfigState, lineById map[string]*configLineRow,
) error {
	if len(lineById) == 0 {
		return nil
	}
	lineIds := make([]string, 0, len(lineById))
	for id := range lineById {
		lineIds = append(lineIds, id)
	}
	rows, err := searchConfigRows(ctx, models.ProductTemplateAttributeValueSchemaName,
		models.ProductTemplateAttributeValueFieldTemplateAttributeId, lineIds)
	if err != nil {
		return err
	}
	lineTemplate := lineTemplateIndex(states)
	for _, row := range rows {
		typed := models.NewProductTemplateAttributeValueFrom(row)
		line := lineById[derefId(typed.GetTemplateAttributeId())]
		if line == nil {
			continue
		}
		value := &configValueRow{
			id:              derefId(typed.GetId()),
			valueId:         derefId(typed.GetAttributeValueId()),
			selected:        typed.GetIsSelected() == nil || *typed.GetIsSelected(),
			sequence:        typed.GetSequence(),
			salesPriceExtra: typed.GetSalesPriceExtra(),
		}
		line.values = append(line.values, value)
		lineTemplate[line.id].tavIndex[value.id] = tavInfo{valueId: value.valueId, line: line}
	}
	return nil
}

func lineTemplateIndex(states map[string]*templateConfigState) map[string]*templateConfigState {
	index := map[string]*templateConfigState{}
	for _, state := range states {
		for _, line := range state.lines {
			index[line.id] = state
		}
	}
	return index
}

func loadConfigExclusions(ctx corectx.Context, states map[string]*templateConfigState, templateIds []string) error {
	rows, err := searchConfigRows(ctx, models.ProductTemplateCombinationExclusionSchemaName,
		models.ProductTemplateCombinationExclusionFieldProductTemplateId, templateIds)
	if err != nil {
		return err
	}
	for _, row := range rows {
		typed := models.NewProductTemplateCombinationExclusionFrom(row)
		state := states[derefId(typed.GetProductTemplateId())]
		if state == nil {
			continue
		}
		archived := row.GetBool(basemodel.FieldIsArchived)
		state.exclusions = append(state.exclusions, &configExclusionRow{
			id:       derefId(typed.GetId()),
			leftId:   derefId(typed.GetLeftValueId()),
			rightId:  derefId(typed.GetRightValueId()),
			archived: archived != nil && *archived,
			etag:     readStringParam(row, basemodel.FieldEtag),
		})
	}
	return nil
}

// searchConfigRows reads every row of schemaName whose field is one of ids, archived included.
func searchConfigRows(
	ctx corectx.Context, schemaName string, field string, ids []string,
) ([]dmodel.DynamicFields, error) {
	repo, err := repoFor(schemaName)
	if err != nil {
		return nil, err
	}
	values := make([]any, 0, len(ids))
	for _, id := range ids {
		values = append(values, id)
	}
	graph := dmodel.NewSearchGraph().And(*dmodel.NewSearchNode().NewCondition(field, dmodel.In, values...))
	found, err := repo.Search(ctx, dyn.RepoSearchParam{
		Graph: graph, Page: 0, Size: configStateScanLimit, IncludeArchived: util.ToPtr(true),
	})
	if err != nil {
		return nil, errors.Wrap(err, "search "+schemaName)
	}
	if found.ClientErrors.Count() > 0 {
		return nil, errors.Wrap(found.ClientErrors.ToError(), "search "+schemaName)
	}
	if !found.HasData {
		return nil, nil
	}
	return found.Data.Items, nil
}
