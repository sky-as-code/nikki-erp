package services

import (
	"sort"
	"strconv"

	"github.com/shopspring/decimal"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"

	"github.com/sky-as-code/nikki-erp/modules/accounting/domain/models"
	taxsvc "github.com/sky-as-code/nikki-erp/modules/accounting/domain/services/tax"
	itExt "github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/external"
	it "github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/tax"
)

// TaxCalculationDomainServiceImpl runs the determination and calculation pipeline. It is the only
// place that both reads configuration and drives the pure functions in services/tax, and holds no
// state between calls: a calculation is a function of its request plus the configuration in force
// on its tax date.
type TaxCalculationDomainServiceImpl struct {
	uomSvc itExt.UomExtService

	// now supplies the calculated-at stamp; a field so a test can freeze it, nil in production. This
	// is NOT the tax date, which is always the caller's and never defaulted from the clock.
	now func() string
}

func NewTaxCalculationDomainServiceImpl(uomSvc itExt.UomExtService) it.TaxCalculationDomainService {
	return &TaxCalculationDomainServiceImpl{uomSvc: uomSvc}
}

func (this *TaxCalculationDomainServiceImpl) Calculate(
	ctx corectx.Context, request it.CalculationRequest,
) (*it.CalculateResult, error) {
	result, _, err := this.run(ctx, request, false)
	if err != nil {
		return nil, err
	}
	return &it.CalculateResult{HasData: true, Data: *result}, nil
}

func (this *TaxCalculationDomainServiceImpl) Simulate(
	ctx corectx.Context, request it.CalculationRequest,
) (*it.SimulateResult, error) {
	result, trace, err := this.run(ctx, request, true)
	if err != nil {
		return nil, err
	}
	return &it.SimulateResult{
		HasData: true,
		Data:    it.SimulationResult{Calculation: *result, Trace: trace},
	}, nil
}

// run is the pipeline both Calculate and Simulate execute, differing only in whether the trace is
// assembled, so the simulator cannot drift from the engine it explains.
func (this *TaxCalculationDomainServiceImpl) run(
	ctx corectx.Context, request it.CalculationRequest, withTrace bool,
) (*it.CalculationResult, []it.TraceStep, error) {
	repos, err := NewTaxRepos()
	if err != nil {
		return nil, nil, err
	}

	var trace []it.TraceStep
	addTrace := func(stage string, detail string, taxIds []string, ruleIds []string) {
		if withTrace {
			trace = append(trace, it.TraceStep{
				Stage: stage, Detail: detail, TaxIds: taxIds, RuleIds: ruleIds,
			})
		}
	}

	policy, policyProblem, err := this.resolveRoundingPolicy(ctx, repos.RoundingPolicy, request)
	if err != nil {
		return nil, nil, err
	}
	if policyProblem != "" {
		// Without a policy there is no rounding scale, and guessing one is a silent pricing
		// decision since currencies round differently. The whole document is unresolved.
		return unresolvedDocument(request, policyProblem), trace, nil
	}
	addTrace("rounding_policy", "resolved rounding policy", nil, nil)

	rules, err := LoadEffectiveRules(ctx, repos, request.TaxDate)
	if err != nil {
		return nil, nil, err
	}
	mappings, err := LoadMappingsForRules(ctx, repos, rules)
	if err != nil {
		return nil, nil, err
	}
	addTrace("rule_load", "loaded "+strconv.Itoa(len(rules))+" published rules in force", nil, nil)

	result := it.CalculationResult{
		Status:        models.DeterminationResolved,
		TotalExcluded: decimal.Zero,
		TotalTax:      decimal.Zero,
		TotalIncluded: decimal.Zero,
	}

	appliedRules := map[string]bool{}
	appliedMappings := map[string]bool{}

	// allocations accumulate across lines so a document-scoped policy can round the document as a
	// whole. A line-scoped policy rounds each line as it goes and leaves this empty.
	var allocations []taxsvc.AllocationInput
	// specByKey attributes each rounded figure back to its version ids without resolving twice.
	specByKey := map[string]ResolvedTax{}

	for _, line := range request.Lines {
		lineResult, lineTrace, err := this.calculateLine(
			ctx, repos, request, line, rules, mappings, policy, appliedRules, appliedMappings,
			&allocations, specByKey,
		)
		if err != nil {
			return nil, nil, err
		}
		if withTrace {
			trace = append(trace, lineTrace...)
		}
		result.Lines = append(result.Lines, *lineResult)
	}

	if policy.Scope == models.RoundingScopeDocument {
		applyDocumentRounding(&result, allocations, policy)
		addTrace("rounding", "allocated document-scope rounding across components", nil, nil)
	}

	totalDocument(&result)
	result.AppliedRuleIds = sortedKeys(appliedRules)
	result.AppliedMappingIds = sortedKeys(appliedMappings)
	result.Snapshot = this.buildSnapshot(request, result, policy)

	return &result, trace, nil
}

// calculateLine determines then computes one line. A line that cannot be resolved does not abort
// the document: it comes back unresolved with its error code, and the document status is downgraded
// when the totals are assembled.
func (this *TaxCalculationDomainServiceImpl) calculateLine(
	ctx corectx.Context,
	repos *TaxRepos,
	request it.CalculationRequest,
	line it.CalculationLine,
	rules []taxsvc.Rule,
	mappings map[string]taxsvc.Mapping,
	policy taxsvc.RoundingPolicy,
	appliedRules map[string]bool,
	appliedMappings map[string]bool,
	allocations *[]taxsvc.AllocationInput,
	specByKey map[string]ResolvedTax,
) (*it.LineResult, []it.TraceStep, error) {
	var trace []it.TraceStep

	outcome := taxsvc.Determine(taxsvc.DeterminationInput{
		Context:         buildContext(request, line),
		CurrencyCode:    request.CurrencyCode,
		CandidateTaxIds: line.CandidateTaxIds,
		Rules:           rules,
		MappingsById:    mappings,
		OverrideTaxIds:  line.OverrideTaxIds,
	})

	for _, ruleId := range outcome.AppliedRuleIds {
		appliedRules[ruleId] = true
	}
	if outcome.AppliedMappingId != "" {
		appliedMappings[outcome.AppliedMappingId] = true
	}
	trace = append(trace, it.TraceStep{
		Stage:   "determination",
		Detail:  "determined tax set for line " + line.LineReference,
		TaxIds:  outcome.TaxIds,
		RuleIds: outcome.AppliedRuleIds,
	})

	base := it.LineResult{
		LineReference: line.LineReference,
		Status:        outcome.Status,
		ErrorCode:     outcome.ErrorCode,
		Treatment:     outcome.Treatment,
		BaseAmount:    line.CommercialBaseAmount,
		TotalExcluded: line.CommercialBaseAmount,
		TotalTax:      decimal.Zero,
		TotalIncluded: line.CommercialBaseAmount,
	}
	if outcome.Status != models.DeterminationResolved {
		// Unresolved and no_tax_applicable both stop here with the same amounts, but mean
		// different things: a failure to fix versus a lawful zero. The status carries that.
		return &base, trace, nil
	}

	resolvedTaxes := make([]ResolvedTax, 0, len(outcome.TaxIds))
	for _, taxId := range outcome.TaxIds {
		resolved, problem, err := ResolveTax(ctx, repos, taxId, request.TaxDate)
		if err != nil {
			return nil, nil, err
		}
		if problem != nil {
			base.Status = models.DeterminationUnresolved
			base.ErrorCode = problem.ErrorCode
			return &base, trace, nil
		}
		resolvedTaxes = append(resolvedTaxes, *resolved)
	}

	quantities, problem, err := this.convertQuantities(ctx, line, request.CurrencyCode, resolvedTaxes)
	if err != nil {
		return nil, nil, err
	}
	if problem != "" {
		base.Status = models.DeterminationUnresolved
		base.ErrorCode = problem
		return &base, trace, nil
	}

	specs := make([]taxsvc.ComponentSpec, 0, len(resolvedTaxes))
	leaves := make([]ResolvedTax, 0, len(resolvedTaxes))
	for _, resolved := range resolvedTaxes {
		specs = append(specs, ToComponentSpecs(resolved, quantities)...)
		leaves = append(leaves, FlattenResolved(resolved)...)
	}
	// A compound chain is order-dependent, so sorting by sequence here is what makes the answer
	// reproducible rather than a function of determination order.
	sort.SliceStable(specs, func(i, j int) bool { return specs[i].Sequence < specs[j].Sequence })

	leafByTaxId := map[string]ResolvedTax{}
	for _, leaf := range leaves {
		leafByTaxId[leaf.TaxId] = leaf
	}

	amounts := taxsvc.CalculateLine(taxsvc.LineInput{
		LineReference:  line.LineReference,
		CommercialBase: line.CommercialBaseAmount,
		PriceMode:      request.PriceMode,
		Components:     specs,
	})

	base.TotalExcluded = amounts.TotalExcluded
	base.TotalTax = amounts.TotalTax
	base.TotalIncluded = amounts.TotalIncluded

	for _, amount := range amounts.Components {
		leaf := leafByTaxId[amount.TaxId]
		component := it.ComponentResult{
			TaxId:                  amount.TaxId,
			TaxCode:                amount.TaxCode,
			TaxName:                leaf.TaxName,
			TaxDefinitionVersionId: leaf.DefinitionVersionId,
			TaxDefinitionVersionNo: leaf.DefinitionVersionNo,
			TaxRateVersionId:       leaf.RateVersionId,
			TaxRateVersionNo:       leaf.RateVersionNo,
			Rate:                   leaf.Rate,
			FixedAmount:            leaf.FixedAmount,
			TaxGroupId:             leaf.TaxGroupId,
			Treatment:              amount.Treatment,
			JurisdictionId:         leaf.JurisdictionId,
			CalculationType:        leaf.CalculationType,
			PriceInclusionMode:     leaf.InclusionMode,
			Sequence:               amount.Sequence,
			TaxableBase:            amount.TaxableBase,
			UnroundedTaxAmount:     amount.Amount,
			LegalReference:         leaf.LegalReference,
		}

		switch policy.Scope {
		case models.RoundingScopeLine:
			// Round-per-line: round now, and the line total is the sum of the rounded components,
			// not the rounded sum, which can differ by a cent the invoice cannot show.
			component.TaxAmount = policy.Round(amount.Amount)
			component.RoundingAdjustment = component.TaxAmount.Sub(amount.Amount)

		default:
			// Round-per-document: this component's rounding depends on the other lines, so the
			// unrounded figure stands in until the allocation pass.
			component.TaxAmount = amount.Amount
			key := allocationKey(line.LineReference, amount.Sequence)
			specByKey[key] = leaf
			*allocations = append(*allocations, taxsvc.AllocationInput{
				LineReference:     line.LineReference,
				ComponentSequence: amount.Sequence,
				GroupKey:          groupKeyOf(leaf),
				Unrounded:         amount.Amount,
			})
		}

		base.Components = append(base.Components, component)
	}

	if policy.Scope == models.RoundingScopeLine {
		retotalLine(&base)
	}
	return &base, trace, nil
}

// convertQuantities turns the line quantity into each fixed tax's own unit, delegating the
// conversion rather than reimplementing it. Only fixed taxes need a quantity, so the UoM port is
// called only for those and an outage cannot stop a percentage VAT being calculated.
func (this *TaxCalculationDomainServiceImpl) convertQuantities(
	ctx corectx.Context, line it.CalculationLine, currencyCode string, resolvedTaxes []ResolvedTax,
) (map[string]decimal.Decimal, string, error) {
	quantities := map[string]decimal.Decimal{}

	var walk func(resolved ResolvedTax) (string, error)
	walk = func(resolved ResolvedTax) (string, error) {
		for _, component := range resolved.Components {
			if problem, err := walk(component); problem != "" || err != nil {
				return problem, err
			}
		}
		if resolved.CalculationType != models.CalculationFixed {
			return "", nil
		}

		// A fixed tax is a money amount, meaningful only in the transaction's currency. There is
		// no FX capability, so a mismatch fails loudly rather than charging 4,000 JPY as 4,000 VND.
		if resolved.RateCurrencyCode != "" && resolved.RateCurrencyCode != currencyCode {
			return models.ErrCodeFixedTaxCurrency, nil
		}

		// Same unit, or none quoted: the line quantity is already what the rate is per.
		if resolved.RateUomId == "" || resolved.RateUomId == line.UomId {
			quantities[resolved.TaxId] = line.Quantity
			return "", nil
		}
		if line.UomId == "" {
			return models.ErrCodeUomConversion, nil
		}

		converted, err := this.uomSvc.Convert(ctx, itExt.ConvertQuantityQuery{
			SourceUomId: line.UomId,
			TargetUomId: resolved.RateUomId,
			Quantity:    line.Quantity,
		})
		if err != nil {
			return "", err
		}
		// An impossible conversion (different dimensions, missing factor) is a per-line business
		// outcome, not a request failure: this line is unresolved, the rest still answer.
		if converted == nil || converted.ClientErrors.Count() > 0 || !converted.HasData {
			return models.ErrCodeUomConversion, nil
		}
		// ExactQuantity, not Quantity: the tax rate is about to multiply this intermediate, and
		// rounding to the unit's display precision first would shed fractions the single final
		// rounding must account for.
		quantities[resolved.TaxId] = converted.Data.ExactQuantity
		return "", nil
	}

	for _, resolved := range resolvedTaxes {
		problem, err := walk(resolved)
		if problem != "" || err != nil {
			return nil, problem, err
		}
	}
	return quantities, "", nil
}

func (this *TaxCalculationDomainServiceImpl) resolveRoundingPolicy(
	ctx corectx.Context, repo models.TaxSearcher, request it.CalculationRequest,
) (taxsvc.RoundingPolicy, string, error) {
	if request.RoundingPolicyCode == "" {
		return taxsvc.RoundingPolicy{}, models.ErrCodeRoundingPolicyMissing, nil
	}

	policy, count, err := models.FindEffectiveRoundingPolicy(
		ctx, repo, request.RoundingPolicyCode, request.TaxDate)
	if err != nil {
		return taxsvc.RoundingPolicy{}, "", err
	}
	if count != 1 {
		// Zero means nothing in force on the date, more than one means contradictory config.
		// Neither may be resolved by picking one.
		return taxsvc.RoundingPolicy{}, models.ErrCodeRoundingPolicyMissing, nil
	}

	return taxsvc.RoundingPolicy{
		Scope:     models.RoundingScope(derefString(policy.GetRoundingScope())),
		Method:    models.RoundingMethod(derefString(policy.GetRoundingMethod())),
		Increment: derefDecimal(policy.GetRoundingIncrement()),
	}, "", nil
}

func applyDocumentRounding(
	result *it.CalculationResult, allocations []taxsvc.AllocationInput, policy taxsvc.RoundingPolicy,
) {
	if len(allocations) == 0 {
		return
	}

	allocated := taxsvc.AllocateDocumentRounding(allocations, policy)
	byKey := make(map[string]taxsvc.AllocationResult, len(allocated))
	for _, entry := range allocated {
		byKey[allocationKey(entry.LineReference, entry.ComponentSequence)] = entry
	}

	for lineIndex := range result.Lines {
		line := &result.Lines[lineIndex]
		for componentIndex := range line.Components {
			component := &line.Components[componentIndex]
			entry, found := byKey[allocationKey(line.LineReference, component.Sequence)]
			if !found {
				continue
			}
			component.TaxAmount = entry.Rounded
			component.RoundingAdjustment = entry.Adjustment
		}
		retotalLine(line)
	}
}

// retotalLine recomputes a line's totals as the sum of the rounded components, never the rounding
// of the summed ones, so the printed components add up to the printed total. It must run after
// rounding.
func retotalLine(line *it.LineResult) {
	total := decimal.Zero
	for _, component := range line.Components {
		total = total.Add(component.TaxAmount)
	}
	line.TotalTax = total
	line.TotalIncluded = line.TotalExcluded.Add(total)
}

func totalDocument(result *it.CalculationResult) {
	result.TotalExcluded = decimal.Zero
	result.TotalTax = decimal.Zero
	result.TotalIncluded = decimal.Zero
	result.RoundingAdjustment = decimal.Zero

	for _, line := range result.Lines {
		result.TotalExcluded = result.TotalExcluded.Add(line.TotalExcluded)
		result.TotalTax = result.TotalTax.Add(line.TotalTax)
		result.TotalIncluded = result.TotalIncluded.Add(line.TotalIncluded)
		for _, component := range line.Components {
			result.RoundingAdjustment = result.RoundingAdjustment.Add(component.RoundingAdjustment)
		}
		// A document is only as resolved as its least resolved line; reporting "resolved" while a
		// line contributed no tax stores a quietly wrong total.
		if line.Status == models.DeterminationUnresolved {
			result.Status = models.DeterminationUnresolved
		}
	}
}

// unresolvedDocument answers a request that could not be calculated at all. Every line is echoed
// back carrying the same error code so the caller can still map results onto its own document.
func unresolvedDocument(request it.CalculationRequest, errorCode string) *it.CalculationResult {
	result := it.CalculationResult{
		Status:        models.DeterminationUnresolved,
		TotalExcluded: decimal.Zero,
		TotalTax:      decimal.Zero,
		TotalIncluded: decimal.Zero,
	}
	for _, line := range request.Lines {
		result.Lines = append(result.Lines, it.LineResult{
			LineReference: line.LineReference,
			Status:        models.DeterminationUnresolved,
			ErrorCode:     errorCode,
			BaseAmount:    line.CommercialBaseAmount,
			TotalExcluded: line.CommercialBaseAmount,
			TotalTax:      decimal.Zero,
			TotalIncluded: line.CommercialBaseAmount,
		})
		result.TotalExcluded = result.TotalExcluded.Add(line.CommercialBaseAmount)
		result.TotalIncluded = result.TotalIncluded.Add(line.CommercialBaseAmount)
	}
	return &result
}

// buildContext assembles the facts a rule condition may test. The whitelist is closed, so a
// condition cannot come to depend on a field another module is free to change.
func buildContext(request it.CalculationRequest, line it.CalculationLine) map[string]string {
	context := map[string]string{
		models.CtxOperationType:            string(request.OperationType),
		models.CtxTaxDate:                  request.TaxDate,
		models.CtxCurrencyCode:             request.CurrencyCode,
		models.CtxProductTaxClassification: line.ProductTaxClassification,
		models.CtxPartyTaxClassification:   request.Buyer.PartyTaxClassification,
		models.CtxSellerJurisdictionId:     request.Seller.PrimaryJurisdictionId,
		models.CtxBuyerJurisdictionId:      request.Buyer.PrimaryJurisdictionId,
		models.CtxShipFromJurisdictionId:   request.ShipFromJurisdictionId,
		models.CtxShipToJurisdictionId:     request.ShipToJurisdictionId,
		models.CtxBuyerIsTaxRegistered:     boolText(isRegisteredAnywhere(request.Buyer)),
		models.CtxSellerIsTaxRegistered:    boolText(isRegisteredAnywhere(request.Seller)),
		models.CtxCommercialBaseAmount:     line.CommercialBaseAmount.String(),
		models.CtxBusinessChannelCode:      request.BusinessChannelCode,
		models.CtxProductReference:         line.ProductReference,
	}
	// A condition tests one value, so a rule only ever sees the first candidate; a rule needing
	// more should test the classification instead.
	if len(line.CandidateTaxIds) > 0 {
		context[models.CtxCandidateTaxId] = line.CandidateTaxIds[0]
	}
	return context
}

func isRegisteredAnywhere(party it.TaxPartyContext) bool {
	for _, registration := range party.TaxRegistrations {
		if registration.IsRegistered {
			return true
		}
	}
	return false
}

func boolText(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

// groupKeyOf is the bucket document-scoped rounding sums within: tax, rate version, treatment and
// jurisdiction. Different taxes must never pool — the per-tax total is what a VAT return reports.
func groupKeyOf(leaf ResolvedTax) string {
	return leaf.TaxId + "|" + leaf.RateVersionId + "|" + string(leaf.Treatment) + "|" + leaf.JurisdictionId
}

func allocationKey(lineReference string, sequence int32) string {
	return lineReference + "#" + strconv.FormatInt(int64(sequence), 10)
}

func sortedKeys(set map[string]bool) []string {
	if len(set) == 0 {
		return nil
	}
	keys := make([]string, 0, len(set))
	for key := range set {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
