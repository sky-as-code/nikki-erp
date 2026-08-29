package services

import (
	"sort"

	"github.com/shopspring/decimal"

	"github.com/sky-as-code/nikki-erp/modules/accounting/domain/models"
	taxsvc "github.com/sky-as-code/nikki-erp/modules/accounting/domain/services/tax"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// ResolvedTax is one tax's configuration as it stood on a given date: everything the calculator
// needs, already chosen, so it cannot reach back into the database and choose differently. The
// version ids travel with it because the snapshot must record which configuration produced an
// amount.
type ResolvedTax struct {
	TaxId   string
	TaxCode string
	TaxName string

	DefinitionVersionId string
	DefinitionVersionNo int32
	RateVersionId       string
	RateVersionNo       int32

	Usage           string
	CalculationType models.CalculationType
	Treatment       models.TaxTreatment
	InclusionMode   models.PriceInclusionMode
	Sequence        int32

	JurisdictionId string
	TaxGroupId     string
	LegalReference string

	Rate        decimal.Decimal
	FixedAmount decimal.Decimal

	// RateUomId is the unit a fixed rate is quoted in. Empty means the rate is not per-unit, which
	// is every non-fixed tax.
	RateUomId string

	// RateCurrencyCode is the currency a fixed amount is denominated in. A percentage rate has none;
	// a percentage of the transaction amount is already in the transaction's currency.
	RateCurrencyCode string

	AffectSubsequentBase   bool
	BaseAffectedByPrevious bool

	// Components are the child taxes of a group tax, already resolved themselves. A non-group tax
	// has none.
	Components []ResolvedTax
}

// ResolutionProblem explains why a tax could not be resolved on a date. It is a value rather than
// an error because one bad line must not deny the rest of the document its answer.
type ResolutionProblem struct {
	TaxId     string
	ErrorCode string
	Detail    string
}

// ResolveTax loads one tax's effective configuration for a date. Exactly one of the resolved tax
// and the problem is always set.
func ResolveTax(
	ctx corectx.Context, repos *TaxRepos, taxId string, taxDate string,
) (*ResolvedTax, *ResolutionProblem, error) {
	return resolveTaxDepth(ctx, repos, taxId, taxDate, 0)
}

// maxComponentDepth bounds the group-tax recursion. Configuration validation already rejects
// cycles, so reaching this depth means a row was written around that guard; refusing beats
// recursing until the process dies.
const maxComponentDepth = 8

// maxComponentsPerGroup bounds one group tax's component fetch. A group with more parts than this
// is a configuration error rather than a business case.
const maxComponentsPerGroup = 50

func resolveTaxDepth(
	ctx corectx.Context, repos *TaxRepos, taxId string, taxDate string, depth int,
) (*ResolvedTax, *ResolutionProblem, error) {
	if depth > maxComponentDepth {
		return nil, &ResolutionProblem{
			TaxId:     taxId,
			ErrorCode: models.ErrCodeTaxConfigurationInvalid,
			Detail:    "tax component nesting exceeds the supported depth",
		}, nil
	}

	tax, err := models.FindTaxById(ctx, repos.Tax, taxId)
	if err != nil {
		return nil, nil, err
	}
	if tax == nil {
		return nil, &ResolutionProblem{
			TaxId:     taxId,
			ErrorCode: models.ErrCodeTaxNotFound,
			Detail:    "no tax with id '" + taxId + "'",
		}, nil
	}

	definition, count, err := models.FindEffectiveDefinitionVersion(ctx, repos.DefinitionVersion, taxId, taxDate)
	if err != nil {
		return nil, nil, err
	}
	// Both absence and multiplicity are configuration faults the caller must see. A tie is never
	// broken by taking the newest: two published versions covering one date means the data is
	// wrong, and picking one would charge under a configuration nobody chose.
	if count == 0 {
		return nil, &ResolutionProblem{
			TaxId:     taxId,
			ErrorCode: models.ErrCodeTaxDefinitionMissing,
			Detail:    "tax has no published definition version effective on " + taxDate,
		}, nil
	}
	if count > 1 {
		return nil, &ResolutionProblem{
			TaxId:     taxId,
			ErrorCode: models.ErrCodeTaxDefinitionAmbiguous,
			Detail:    "tax has more than one published definition version effective on " + taxDate,
		}, nil
	}

	resolved := ResolvedTax{
		TaxId:                  taxId,
		TaxCode:                derefString(tax.GetCode()),
		TaxName:                langText(tax),
		DefinitionVersionId:    idString(definition.GetId()),
		DefinitionVersionNo:    derefInt32(definition.GetVersionNo()),
		Usage:                  derefString(definition.GetUsage()),
		CalculationType:        models.CalculationType(derefString(definition.GetCalculationType())),
		Treatment:              models.TaxTreatment(derefString(definition.GetTaxTreatment())),
		InclusionMode:          models.PriceInclusionMode(derefString(definition.GetPriceInclusionMode())),
		Sequence:               derefInt32(definition.GetSequence()),
		JurisdictionId:         idString(definition.GetJurisdictionId()),
		TaxGroupId:             idString(definition.GetTaxGroupId()),
		LegalReference:         derefString(definition.GetLegalReference()),
		AffectSubsequentBase:   derefBool(definition.GetAffectSubsequentBase()),
		BaseAffectedByPrevious: derefBool(definition.GetBaseAffectedByPrevious()),
	}

	switch resolved.CalculationType {
	case models.CalculationGroup:
		components, problem, err := resolveComponents(ctx, repos, resolved.DefinitionVersionId, taxDate, depth)
		if err != nil || problem != nil {
			return nil, problem, err
		}
		resolved.Components = components
		return &resolved, nil, nil

	case models.CalculationNone:
		// An exempt or out-of-scope tax is calculated as zero, not skipped: the line still needs a
		// component recording its legal treatment. It has no rate, so nothing further to resolve.
		return &resolved, nil, nil
	}

	rate, rateCount, err := models.FindEffectiveRateVersion(ctx, repos.RateVersion, taxId, taxDate)
	if err != nil {
		return nil, nil, err
	}
	if rateCount == 0 {
		return nil, &ResolutionProblem{
			TaxId:     taxId,
			ErrorCode: models.ErrCodeTaxRateMissing,
			Detail:    "tax has no published rate version effective on " + taxDate,
		}, nil
	}
	if rateCount > 1 {
		return nil, &ResolutionProblem{
			TaxId:     taxId,
			ErrorCode: models.ErrCodeTaxRateAmbiguous,
			Detail:    "tax has more than one published rate version effective on " + taxDate,
		}, nil
	}

	resolved.RateVersionId = idString(rate.GetId())
	resolved.RateVersionNo = derefInt32(rate.GetVersionNo())
	resolved.Rate = derefDecimal(rate.GetRate())
	resolved.FixedAmount = derefDecimal(rate.GetFixedAmount())
	resolved.RateUomId = idString(rate.GetRateUomId())
	resolved.RateCurrencyCode = derefString(rate.GetCurrencyCode())

	return &resolved, nil, nil
}

// resolveComponents resolves the child taxes of a group tax, in declared sequence.
func resolveComponents(
	ctx corectx.Context, repos *TaxRepos, definitionVersionId string, taxDate string, depth int,
) ([]ResolvedTax, *ResolutionProblem, error) {
	rows, err := models.FindComponentsOfVersion(ctx, repos.Component, definitionVersionId, maxComponentsPerGroup)
	if err != nil {
		return nil, nil, err
	}
	if len(rows) == 0 {
		return nil, &ResolutionProblem{
			ErrorCode: models.ErrCodeTaxConfigurationInvalid,
			Detail:    "group tax definition version '" + definitionVersionId + "' declares no components",
		}, nil
	}

	parsed := make([]*models.TaxComponent, 0, len(rows))
	for _, row := range rows {
		parsed = append(parsed, models.NewTaxComponentFrom(row))
	}
	// A compound chain is order-dependent: the second component's base includes the first's amount.
	// The database returns rows in no guaranteed order, so sorting here makes the answer
	// reproducible.
	sort.SliceStable(parsed, func(i, j int) bool {
		return derefInt32(parsed[i].GetSequence()) < derefInt32(parsed[j].GetSequence())
	})

	components := make([]ResolvedTax, 0, len(parsed))
	for _, row := range parsed {
		componentTaxId := idString(row.GetComponentTaxId())
		child, problem, err := resolveTaxDepth(ctx, repos, componentTaxId, taxDate, depth+1)
		if err != nil || problem != nil {
			return nil, problem, err
		}

		// The component's own sequence overrides the one its definition carries for standalone use:
		// the same tax may sit second in one group and fourth in another.
		child.Sequence = derefInt32(row.GetSequence())

		// A component may override compounding for this group alone; absent an override the child's
		// definition decides, hence the nullable field rather than a plain bool.
		if override := row.GetAffectSubsequentBaseOverride(); override != nil {
			child.AffectSubsequentBase = *override
		}
		components = append(components, *child)
	}
	return components, nil, nil
}

// ToComponentSpecs flattens a resolved tax into the calculator's inputs. A group tax contributes
// its children and never itself, or the line would be taxed twice. quantity is the line's quantity
// already converted into each fixed rate's own unit by the caller, keyed by tax id.
func ToComponentSpecs(
	resolved ResolvedTax, quantityByTaxId map[string]decimal.Decimal,
) []taxsvc.ComponentSpec {
	if resolved.CalculationType == models.CalculationGroup {
		specs := make([]taxsvc.ComponentSpec, 0, len(resolved.Components))
		for _, component := range resolved.Components {
			specs = append(specs, ToComponentSpecs(component, quantityByTaxId)...)
		}
		return specs
	}

	return []taxsvc.ComponentSpec{{
		TaxId:                  resolved.TaxId,
		TaxCode:                resolved.TaxCode,
		Sequence:               resolved.Sequence,
		CalculationType:        resolved.CalculationType,
		Treatment:              resolved.Treatment,
		InclusionMode:          resolved.InclusionMode,
		Rate:                   resolved.Rate,
		FixedAmount:            resolved.FixedAmount,
		Quantity:               quantityByTaxId[resolved.TaxId],
		AffectSubsequentBase:   resolved.AffectSubsequentBase,
		BaseAffectedByPrevious: resolved.BaseAffectedByPrevious,
	}}
}

// FlattenResolved returns every leaf tax of a resolved tree, in calculation order. The calculator
// returns amounts keyed by tax id and sequence, and only the resolver knows which version ids
// those came from.
func FlattenResolved(resolved ResolvedTax) []ResolvedTax {
	if resolved.CalculationType != models.CalculationGroup {
		return []ResolvedTax{resolved}
	}
	leaves := make([]ResolvedTax, 0, len(resolved.Components))
	for _, component := range resolved.Components {
		leaves = append(leaves, FlattenResolved(component)...)
	}
	return leaves
}
