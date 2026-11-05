package promotion

import (
	"slices"
	"sort"
)

// IsEmpty reports a selector that restricts nothing: every purchase line matches (BR-C03).
func (this Selector) IsEmpty() bool {
	return len(this.ProductTemplateIds) == 0 && len(this.ProductVariantIds) == 0 && len(this.CategoryIds) == 0
}

// Matches is OR inside a dimension and AND across dimensions. Products is one dimension: a variant
// matches directly or through its template. A category matches the line's own category or any
// ancestor the caller supplied, which is how descendants are included.
func (this Selector) Matches(line Line) bool {
	return this.matchesProducts(line) && this.matchesCategories(line)
}

func (this Selector) matchesProducts(line Line) bool {
	if len(this.ProductTemplateIds) == 0 && len(this.ProductVariantIds) == 0 {
		return true
	}
	if line.ProductVariantId != "" && slices.Contains(this.ProductVariantIds, line.ProductVariantId) {
		return true
	}
	return line.ProductTemplateId != "" && slices.Contains(this.ProductTemplateIds, line.ProductTemplateId)
}

func (this Selector) matchesCategories(line Line) bool {
	if len(this.CategoryIds) == 0 {
		return true
	}
	for _, categoryId := range line.CategoryIds {
		if slices.Contains(this.CategoryIds, categoryId) {
			return true
		}
	}
	return false
}

// isPurchaseLine keeps only what the customer pays for with a positive quantity (BR-C05): gift lines
// never qualify, so a gift can never raise its own condition (AC-16).
func isPurchaseLine(line Line) bool {
	if line.LineType != LineTypeProduct && line.LineType != LineTypeCombo {
		return false
	}
	return line.Quantity.IsPositive()
}

// purchaseLines returns the purchase lines in (LineNumber, Key) order, the tie-break order of every
// rule in this package.
func purchaseLines(input *Input) []Line {
	lines := make([]Line, 0, len(input.Lines))
	for _, line := range input.Lines {
		if isPurchaseLine(line) {
			lines = append(lines, line)
		}
	}
	sort.SliceStable(lines, func(i, j int) bool { return lineLess(lines[i], lines[j]) })
	return lines
}

func lineLess(left Line, right Line) bool {
	if left.LineNumber != right.LineNumber {
		return left.LineNumber < right.LineNumber
	}
	return left.Key < right.Key
}

func pricelistAllows(program Program, line Line) bool {
	return len(program.PricelistIds) == 0 || slices.Contains(program.PricelistIds, line.PricelistId)
}

// selectLines returns the purchase lines a selector picks for a program, plus the reason when none
// qualify: pricelist_mismatch when the selector matched lines that no linked pricelist priced (D2).
func selectLines(input *Input, program Program, selector Selector) ([]Line, string) {
	matched := []Line{}
	pricelistExcluded := false
	for _, line := range purchaseLines(input) {
		if !selector.Matches(line) {
			continue
		}
		if !pricelistAllows(program, line) {
			pricelistExcluded = true
			continue
		}
		matched = append(matched, line)
	}
	if len(matched) > 0 {
		return matched, ""
	}
	if pricelistExcluded {
		return matched, ReasonPricelistMismatch
	}
	return matched, ReasonNoEligibleLine
}

func lineKeys(lines []Line) []string {
	keys := make([]string, len(lines))
	for i, line := range lines {
		keys[i] = line.Key
	}
	return keys
}
