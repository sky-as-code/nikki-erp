package services

import (
	"sort"

	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
)

// planDimension is one attribute line with its selected values deduplicated.
type planDimension struct {
	attributeId string
	mode        models.VariantCreationMode
	valueIds    []string
	valueSet    map[string]bool
}

// planState is a PlanConfig indexed for validity checks and bounded combination walks. instant and
// dynamic are the generating dimensions (A and D), sorted by attribute id.
type planState struct {
	instant    []planDimension
	dynamic    []planDimension
	lines      map[string]*planDimension
	valueOwner map[string]string
	exclusions map[string]bool
	inactive   []PlanExclusion
	duplicates []string
}

func newPlanState(config PlanConfig) *planState {
	state := &planState{
		lines:      map[string]*planDimension{},
		valueOwner: map[string]string{},
		exclusions: map[string]bool{},
	}
	for _, line := range config.Lines {
		state.addLine(line)
	}
	sortDimensions(state.instant)
	sortDimensions(state.dynamic)
	for _, exclusion := range config.Exclusions {
		state.addExclusion(exclusion)
	}
	return state
}

func (this *planState) addLine(line PlanLine) {
	if line.AttributeId == "" {
		return
	}
	if _, seen := this.lines[line.AttributeId]; seen {
		this.duplicates = append(this.duplicates, line.AttributeId)
		return
	}
	dim := planDimension{attributeId: line.AttributeId, mode: line.Mode, valueSet: map[string]bool{}}
	for _, valueId := range line.ValueIds {
		if valueId == "" || dim.valueSet[valueId] {
			continue
		}
		dim.valueSet[valueId] = true
		dim.valueIds = append(dim.valueIds, valueId)
	}
	this.lines[line.AttributeId] = &dim
	if !line.Mode.CreatesVariants() {
		return
	}
	for _, valueId := range dim.valueIds {
		this.valueOwner[valueId] = line.AttributeId
	}
	if line.Mode == models.VariantCreationModeInstant {
		this.instant = append(this.instant, dim)
	} else {
		this.dynamic = append(this.dynamic, dim)
	}
}

// addExclusion keeps an exclusion only between selected values of two different generating
// attributes; anything else (never values, unselected values, same attribute) is inactive.
func (this *planState) addExclusion(exclusion PlanExclusion) {
	leftOwner, leftFound := this.valueOwner[exclusion.LeftValueId]
	rightOwner, rightFound := this.valueOwner[exclusion.RightValueId]
	if !leftFound || !rightFound || leftOwner == rightOwner {
		this.inactive = append(this.inactive, exclusion)
		return
	}
	this.exclusions[exclusionPairKey(exclusion.LeftValueId, exclusion.RightValueId)] = true
}

func exclusionPairKey(left string, right string) string {
	if left > right {
		left, right = right, left
	}
	return left + combinationSeparator + right
}

func sortDimensions(dims []planDimension) {
	sort.Slice(dims, func(i, j int) bool { return dims[i].attributeId < dims[j].attributeId })
}

func (this *planState) isMixed() bool {
	return len(this.instant) > 0 && len(this.dynamic) > 0
}

func (this *planState) generating() []planDimension {
	dims := make([]planDimension, 0, len(this.instant)+len(this.dynamic))
	dims = append(dims, this.instant...)
	return append(dims, this.dynamic...)
}

// compatible reports whether valueId may join the already chosen values without an exclusion.
func (this *planState) compatible(valueId string, chosen []string) bool {
	for _, other := range chosen {
		if this.exclusions[exclusionPairKey(valueId, other)] {
			return false
		}
	}
	return true
}

// countLeaves counts the valid ways to pick one value per dimension on top of chosen, pruning on
// exclusions, and stops as soon as the count passes limit. limit 0 is an existence check.
func (this *planState) countLeaves(dims []planDimension, chosen []string, limit int) int {
	if len(dims) == 0 {
		return 1
	}
	total := 0
	for _, valueId := range dims[0].valueIds {
		if !this.compatible(valueId, chosen) {
			continue
		}
		total += this.countLeaves(dims[1:], append(chosen, valueId), limit-total)
		if total > limit {
			return total
		}
	}
	return total
}

// walkLeaves visits every valid pick of one value per dimension; visit returns false to stop.
func (this *planState) walkLeaves(dims []planDimension, chosen []string, visit func([]string) bool) bool {
	if len(dims) == 0 {
		return visit(chosen)
	}
	for _, valueId := range dims[0].valueIds {
		if !this.compatible(valueId, chosen) {
			continue
		}
		if !this.walkLeaves(dims[1:], append(chosen, valueId), visit) {
			return false
		}
	}
	return true
}

func (this *planState) keyOfValues(valueIds []string) string {
	pairs := make(map[string]string, len(valueIds))
	for _, valueId := range valueIds {
		pairs[this.valueOwner[valueId]] = valueId
	}
	return BuildKeyFromPairs(pairs)
}

// keyValues returns the key's values when the key holds exactly one selected value for each of
// dims, nothing else, and no excluded pair.
func (this *planState) keyValues(key string, dims []planDimension) ([]string, bool) {
	pairs := ParseKeyToPairs(key)
	if len(pairs) != len(dims) {
		return nil, false
	}
	values := make([]string, 0, len(dims))
	for _, dim := range dims {
		valueId, found := pairs[dim.attributeId]
		if !found || !dim.valueSet[valueId] || !this.compatible(valueId, values) {
			return nil, false
		}
		values = append(values, valueId)
	}
	return values, true
}

func (this *planState) isValidComplete(key string) bool {
	_, valid := this.keyValues(key, this.generating())
	return valid
}

// isValidBase holds only in a mixed configuration, for an instant combination that still has at
// least one valid dynamic completion.
func (this *planState) isValidBase(key string) bool {
	if !this.isMixed() {
		return false
	}
	values, valid := this.keyValues(key, this.instant)
	return valid && this.countLeaves(this.dynamic, values, 0) > 0
}

func (this *planState) isValid(kind models.VariantKind, key string) bool {
	if kind == models.VariantKindBase {
		return this.isValidBase(key)
	}
	return this.isValidComplete(key)
}

// instantSubKey is the base key a complete of a mixed template hangs under.
func (this *planState) instantSubKey(key string) string {
	pairs := ParseKeyToPairs(key)
	sub := map[string]string{}
	for _, dim := range this.instant {
		sub[dim.attributeId] = pairs[dim.attributeId]
	}
	return BuildKeyFromPairs(sub)
}
