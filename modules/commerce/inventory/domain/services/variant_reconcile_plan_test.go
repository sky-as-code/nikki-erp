package services

import (
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
)

const (
	attrColor    = "a_color"
	attrSize     = "a_size"
	attrMaterial = "a_material"
	attrGift     = "a_gift"
)

var (
	modeInstant = models.VariantCreationModeInstant
	modeDynamic = models.VariantCreationModeDynamic
	modeNever   = models.VariantCreationModeNever
	kindBase    = models.VariantKindBase
	kindFull    = models.VariantKindComplete
)

func planLine(attributeId string, mode models.VariantCreationMode, values ...string) PlanLine {
	return PlanLine{AttributeId: attributeId, Mode: mode, ValueIds: values}
}

func planConfig(lines ...PlanLine) PlanConfig {
	return PlanConfig{Lines: lines}
}

// planKey builds a key from attribute, value, attribute, value...
func planKey(pairs ...string) string {
	byAttribute := map[string]string{}
	for i := 0; i+1 < len(pairs); i += 2 {
		byAttribute[pairs[i]] = pairs[i+1]
	}
	return BuildKeyFromPairs(byAttribute)
}

func planActiveVariant(id string, kind models.VariantKind, key string) PlanVariant {
	return PlanVariant{Id: id, Kind: kind, Key: key}
}

func planArchivedVariant(id string, kind models.VariantKind, key string, source models.ArchiveSource) PlanVariant {
	return PlanVariant{Id: id, Kind: kind, Key: key, IsArchived: true, ArchiveSource: source}
}

func plannedIds(items []PlannedVariant) []string {
	ids := []string{}
	for _, item := range items {
		ids = append(ids, item.VariantId)
	}
	sort.Strings(ids)
	return ids
}

func plannedKeys(items []PlannedVariant) []string {
	keys := []string{}
	for _, item := range items {
		keys = append(keys, item.Key)
	}
	sort.Strings(keys)
	return keys
}

func sortedKeys(keys ...string) []string {
	sort.Strings(keys)
	return keys
}

func requireViolation(t *testing.T, plan *VariantPlan, reason string) {
	t.Helper()
	require.True(t, plan.HasViolations())
	require.Equal(t, "inventory:"+reason, plan.ClientErrors[0].Key)
}

func requirePossible(t *testing.T, plan *VariantPlan, expected int) {
	t.Helper()
	require.False(t, plan.PossibleOverThreshold)
	require.NotNil(t, plan.PossibleCompleteCount)
	require.Equal(t, expected, *plan.PossibleCompleteCount)
}

// AC04: no attribute, or never-only, yields exactly one empty complete; a repeat keeps it.
func TestPlan_AC04_DefaultVariant(t *testing.T) {
	cases := []struct {
		name     string
		config   PlanConfig
		existing []PlanVariant
		creates  int
		keeps    int
	}{
		{"no attribute", planConfig(), nil, 1, 0},
		{"never only", planConfig(planLine(attrGift, modeNever, "yes", "no")), nil, 1, 0},
		{"repeat is idempotent", planConfig(), []PlanVariant{planActiveVariant("v0", kindFull, "")}, 0, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			plan := PlanTemplateVariants(PlanInput{Next: tc.config, Existing: tc.existing})
			require.False(t, plan.HasViolations())
			require.Len(t, plan.Create, tc.creates)
			require.Len(t, plan.Keep, tc.keeps)
			require.Equal(t, 1, plan.PreCreateCount)
			requirePossible(t, plan, 1)
			if tc.creates == 1 {
				require.Equal(t, models.EmptyCombinationKey, plan.Create[0].Key)
				require.Equal(t, kindFull, plan.Create[0].Kind)
			}
		})
	}
}

// AC05: the default V0 takes on Color=Red in place.
func TestPlan_AC05_DefaultExtendedBySingleValue(t *testing.T) {
	plan := PlanTemplateVariants(PlanInput{
		Next:     planConfig(planLine(attrColor, modeInstant, "red")),
		Existing: []PlanVariant{planActiveVariant("v0", kindFull, "")},
	})
	require.False(t, plan.HasViolations())
	require.Len(t, plan.Extend, 1)
	require.Equal(t, "v0", plan.Extend[0].VariantId)
	require.Equal(t, planKey(attrColor, "red"), plan.Extend[0].NewKey)
	require.Empty(t, plan.Create)
	require.Empty(t, plan.Invalid)
}

// AC05 (multi-dimension): several new instant dimensions with one value each still reuse V0.
func TestPlan_AC05_DefaultExtendedBySeveralSingleValues(t *testing.T) {
	plan := PlanTemplateVariants(PlanInput{
		Next:     planConfig(planLine(attrColor, modeInstant, "red"), planLine(attrSize, modeInstant, "s")),
		Existing: []PlanVariant{planActiveVariant("v0", kindFull, "")},
	})
	require.Len(t, plan.Extend, 1)
	require.Equal(t, planKey(attrColor, "red", attrSize, "s"), plan.Extend[0].NewKey)
	require.Empty(t, plan.Create)
}

// AC06: after AC05, adding Blue keeps Red and creates only Blue.
func TestPlan_AC06_AddValueKeepsExisting(t *testing.T) {
	prev := planConfig(planLine(attrColor, modeInstant, "red"))
	plan := PlanTemplateVariants(PlanInput{
		Previous: &prev,
		Next:     planConfig(planLine(attrColor, modeInstant, "red", "blue")),
		Existing: []PlanVariant{planActiveVariant("v0", kindFull, planKey(attrColor, "red"))},
	})
	require.Equal(t, []string{"v0"}, plannedIds(plan.Keep))
	require.Equal(t, []string{planKey(attrColor, "blue")}, plannedKeys(plan.Create))
	require.Empty(t, plan.Extend)
}

// AC07: Red and Blue at once never pick V0; V0 becomes invalid.
func TestPlan_AC07_MultiValueDoesNotReuseDefault(t *testing.T) {
	plan := PlanTemplateVariants(PlanInput{
		Next:     planConfig(planLine(attrColor, modeInstant, "red", "blue")),
		Existing: []PlanVariant{planActiveVariant("v0", kindFull, "")},
	})
	require.Empty(t, plan.Extend)
	require.Equal(t, []string{"v0"}, plannedIds(plan.Invalid))
	require.Equal(t, sortedKeys(planKey(attrColor, "red"), planKey(attrColor, "blue")), plannedKeys(plan.Create))
}

// AC08: 2x2 instant yields four keys, independent of line order.
func TestPlan_AC08_KeysIndependentOfLineOrder(t *testing.T) {
	color := planLine(attrColor, modeInstant, "red", "blue")
	size := planLine(attrSize, modeInstant, "s", "m")
	first := PlanTemplateVariants(PlanInput{Next: planConfig(color, size)})
	second := PlanTemplateVariants(PlanInput{Next: planConfig(size, color)})
	require.Len(t, first.Create, 4)
	require.Equal(t, plannedKeys(first.Create), plannedKeys(second.Create))
}

// AC11: a single new instant dimension is appended to every current variant; a clash is refused.
func TestPlan_AC11_ExtendAllCurrentVariants(t *testing.T) {
	prev := planConfig(planLine(attrColor, modeInstant, "red", "blue"))
	next := planConfig(planLine(attrColor, modeInstant, "red", "blue"), planLine(attrMaterial, modeInstant, "cotton"))
	existing := []PlanVariant{
		planActiveVariant("v10", kindFull, planKey(attrColor, "red")),
		planActiveVariant("v11", kindFull, planKey(attrColor, "blue")),
		planArchivedVariant("v12", kindFull, planKey(attrColor, "green"), models.ArchiveSourceSystemSync),
	}
	plan := PlanTemplateVariants(PlanInput{Previous: &prev, Next: next, Existing: existing})
	require.False(t, plan.HasViolations())
	require.Equal(t, []string{"v10", "v11"}, plannedIds(plan.Extend))
	require.Equal(t, []string{"v12"}, plannedIds(plan.Untouched), "archived variants are never extended")
	require.Empty(t, plan.Create)

	existing = append(existing,
		planArchivedVariant("v13", kindFull, planKey(attrColor, "red", attrMaterial, "cotton"), models.ArchiveSourceUser))
	plan = PlanTemplateVariants(PlanInput{Previous: &prev, Next: next, Existing: existing})
	requireViolation(t, plan, ReasonPlanCombinationConflict)
	require.Equal(t, []string{"v10"}, plannedIds(plan.Conflict))
}

// AC12: removing Red keeps Blue and invalidates Red.
func TestPlan_AC12_RemoveValue(t *testing.T) {
	prev := planConfig(planLine(attrColor, modeInstant, "red", "blue"))
	plan := PlanTemplateVariants(PlanInput{
		Previous: &prev,
		Next:     planConfig(planLine(attrColor, modeInstant, "blue")),
		Existing: []PlanVariant{
			planActiveVariant("red", kindFull, planKey(attrColor, "red")),
			planActiveVariant("blue", kindFull, planKey(attrColor, "blue")),
		},
	})
	require.Equal(t, []string{"blue"}, plannedIds(plan.Keep))
	require.Equal(t, []string{"red"}, plannedIds(plan.Invalid))
	require.Empty(t, plan.Create)
}

// AC13: removing Size invalidates Red-S and Red-M and creates Red; nothing is merged.
func TestPlan_AC13_RemoveDimension(t *testing.T) {
	prev := planConfig(planLine(attrColor, modeInstant, "red"), planLine(attrSize, modeInstant, "s", "m"))
	plan := PlanTemplateVariants(PlanInput{
		Previous: &prev,
		Next:     planConfig(planLine(attrColor, modeInstant, "red")),
		Existing: []PlanVariant{
			planActiveVariant("rs", kindFull, planKey(attrColor, "red", attrSize, "s")),
			planActiveVariant("rm", kindFull, planKey(attrColor, "red", attrSize, "m")),
		},
	})
	require.Equal(t, []string{"rm", "rs"}, plannedIds(plan.Invalid))
	require.Equal(t, []string{planKey(attrColor, "red")}, plannedKeys(plan.Create))
	require.Empty(t, plan.Extend)
}

// AC14: removing every dimension finds or creates the empty complete; a configured variant is not reused.
func TestPlan_AC14_RemoveAllDimensions(t *testing.T) {
	prev := planConfig(planLine(attrColor, modeInstant, "red"))
	existing := []PlanVariant{planActiveVariant("red", kindFull, planKey(attrColor, "red"))}
	plan := PlanTemplateVariants(PlanInput{Previous: &prev, Next: planConfig(), Existing: existing})
	require.Equal(t, []string{"red"}, plannedIds(plan.Invalid))
	require.Equal(t, []string{""}, plannedKeys(plan.Create))

	existing = append(existing, planArchivedVariant("v0", kindFull, "", models.ArchiveSourceSystemSync))
	plan = PlanTemplateVariants(PlanInput{Previous: &prev, Next: planConfig(), Existing: existing})
	require.Equal(t, []string{"v0"}, plannedIds(plan.Restore))
	require.Empty(t, plan.Create)
}

// AC15 / AC16 / D3: only a system_sync archive of an active template is restored; nothing is duplicated.
func TestPlan_AC15_RestoreBySource(t *testing.T) {
	next := planConfig(planLine(attrColor, modeInstant, "red", "blue"))
	cases := []struct {
		name             string
		source           models.ArchiveSource
		templateArchived bool
		outcome          PlanOutcome
	}{
		{"system sync restored", models.ArchiveSourceSystemSync, false, PlanOutcomeRestore},
		{"user archive kept (AC16)", models.ArchiveSourceUser, false, PlanOutcomeUntouched},
		{"template cascade left to template unarchive", models.ArchiveSourceTemplateCascade, false, PlanOutcomeUntouched},
		{"archived template never restores", models.ArchiveSourceSystemSync, true, PlanOutcomeUntouched},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			plan := PlanTemplateVariants(PlanInput{
				Next: next, TemplateArchived: tc.templateArchived,
				Existing: []PlanVariant{planArchivedVariant("red", kindFull, planKey(attrColor, "red"), tc.source)},
			})
			all := append(append(plan.Restore, plan.Untouched...), plan.Keep...)
			require.Len(t, all, 1)
			require.Equal(t, tc.outcome, all[0].Outcome)
			require.Equal(t, []string{planKey(attrColor, "blue")}, plannedKeys(plan.Create), "red is never duplicated")
		})
	}
}

// D3: a template_cascade archive that configuration invalidates is reported invalid (becomes system_sync).
func TestPlan_TemplateCascadeInvalidated(t *testing.T) {
	plan := PlanTemplateVariants(PlanInput{
		Next: planConfig(planLine(attrColor, modeInstant, "blue")),
		Existing: []PlanVariant{
			planArchivedVariant("red", kindFull, planKey(attrColor, "red"), models.ArchiveSourceTemplateCascade),
			planArchivedVariant("green", kindFull, planKey(attrColor, "green"), models.ArchiveSourceUser),
		},
	})
	require.Equal(t, []string{"red"}, plannedIds(plan.Invalid))
	require.Equal(t, []string{"green"}, plannedIds(plan.Untouched))
}

// AC23 / ADD37: pre-creating more than the limit is refused as a whole, with the count.
func TestPlan_AC23_ADD37_PreCreateLimit(t *testing.T) {
	plan := PlanTemplateVariants(PlanInput{Next: PlanConfig{
		Lines: []PlanLine{
			planLine(attrColor, modeInstant, numberedValues("c", 40)...),
			planLine(attrSize, modeInstant, numberedValues("s", 30)...),
		},
	}})
	requireViolation(t, plan, ReasonPlanTooManyVariants)
	require.Equal(t, 1200, plan.PreCreateCount)
	require.Equal(t, 1200, plan.ClientErrors[0].Vars["count"])
	require.Empty(t, plan.Create)

	plan = PlanTemplateVariants(PlanInput{Next: PlanConfig{
		Lines:         []PlanLine{planLine(attrColor, modeInstant, "red", "blue", "green")},
		MaxPreCreated: 2,
	}})
	requireViolation(t, plan, ReasonPlanTooManyVariants)
}

// AC23 (bounded): a huge product is counted only up to the cap, never materialized.
func TestPlan_AC23_BoundedCounting(t *testing.T) {
	lines := []PlanLine{}
	for i := 0; i < 20; i++ {
		values := numberedValues(fmt.Sprintf("v%02d_", i), 10)
		lines = append(lines, planLine(fmt.Sprintf("a%02d", i), modeInstant, values...))
	}
	plan := PlanTemplateVariants(PlanInput{Next: PlanConfig{Lines: lines}})
	requireViolation(t, plan, ReasonPlanTooManyVariants)
	require.True(t, plan.PossibleOverThreshold)
	require.Nil(t, plan.PossibleCompleteCount)
	require.Equal(t, DefaultPossibleCountCap+1, plan.PreCreateCount)
}

// AC24 / ADD24: exclusions leaving no complete refuse the configuration.
func TestPlan_AC24_ADD24_EverythingExcluded(t *testing.T) {
	cases := []struct {
		name string
		mode models.VariantCreationMode
	}{{"instant", modeInstant}, {"mixed", modeDynamic}}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			config := planConfig(planLine(attrColor, modeInstant, "red"), planLine(attrSize, tc.mode, "s"))
			config.Exclusions = []PlanExclusion{{LeftValueId: "s", RightValueId: "red"}}
			plan := PlanTemplateVariants(PlanInput{
				Next: config, Existing: []PlanVariant{planActiveVariant("v0", kindFull, "")},
			})
			requireViolation(t, plan, ReasonPlanNoValidCombination)
			require.Empty(t, plan.Invalid, "nothing is archived when the configuration is refused")
		})
	}
}
