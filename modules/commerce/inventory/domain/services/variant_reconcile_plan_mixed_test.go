package services

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
)

// ADD11: always 2 colours x 3 sizes yields 6 completes and 0 bases.
func TestPlan_ADD11_AlwaysOnly(t *testing.T) {
	plan := PlanTemplateVariants(PlanInput{Next: planConfig(
		planLine(attrColor, modeInstant, "red", "blue"), planLine(attrSize, modeInstant, "s", "m", "l"))})
	require.Len(t, plan.Create, 6)
	for _, item := range plan.Create {
		require.Equal(t, kindFull, item.Kind)
	}
	require.Equal(t, kindFull, plan.PreCreateKind)
	require.Equal(t, 6, plan.PreCreateCount)
	requirePossible(t, plan, 6)
}

// ADD12 / ADD13: mixed pre-creates bases only; never values do not multiply anything.
func TestPlan_ADD12_ADD13_Mixed(t *testing.T) {
	lines := []PlanLine{planLine(attrColor, modeInstant, "red", "blue"), planLine(attrSize, modeDynamic, "s", "m", "l")}
	for _, withNever := range []bool{false, true} {
		config := planConfig(lines...)
		if withNever {
			config.Lines = append(config.Lines, planLine(attrGift, modeNever, "g1", "g2", "g3", "g4"))
		}
		plan := PlanTemplateVariants(PlanInput{Next: config})
		require.False(t, plan.HasViolations())
		require.Equal(t, sortedKeys(planKey(attrColor, "red"), planKey(attrColor, "blue")), plannedKeys(plan.Create))
		for _, item := range plan.Create {
			require.Equal(t, kindBase, item.Kind)
		}
		require.Equal(t, 2, plan.PreCreateCount)
		requirePossible(t, plan, 6)
	}
}

// ADD14 / ADD15: dynamic-only pre-creates nothing, even with a single value.
func TestPlan_ADD14_ADD15_DynamicOnly(t *testing.T) {
	for _, values := range [][]string{{"s", "m", "l"}, {"s"}} {
		plan := PlanTemplateVariants(PlanInput{Next: planConfig(planLine(attrSize, modeDynamic, values...))})
		require.False(t, plan.HasViolations())
		require.Empty(t, plan.Create)
		require.Equal(t, 0, plan.PreCreateCount)
		requirePossible(t, plan, len(values))
	}
}

// ADD22: Red-S excluded keeps base Red through Red-M; resolving Red-S is refused.
func TestPlan_ADD22_PartialExclusionKeepsBase(t *testing.T) {
	config := planConfig(planLine(attrColor, modeInstant, "red", "blue"), planLine(attrSize, modeDynamic, "s", "m"))
	config.Exclusions = []PlanExclusion{{LeftValueId: "red", RightValueId: "s"}}
	plan := PlanTemplateVariants(PlanInput{Next: config})
	require.Equal(t, sortedKeys(planKey(attrColor, "red"), planKey(attrColor, "blue")), plannedKeys(plan.Create))
	requirePossible(t, plan, 3)

	_, _, violations := ValidateSelection(config, map[string]string{attrColor: "red", attrSize: "s"})
	require.Equal(t, "inventory:"+ReasonSelectionExcluded, violations[0].Key)
	completeKey, baseKey, violations := ValidateSelection(config, map[string]string{attrColor: "red", attrSize: "m"})
	require.Empty(t, violations)
	require.Equal(t, planKey(attrColor, "red", attrSize, "m"), completeKey)
	require.Equal(t, planKey(attrColor, "red"), baseKey)
}

// ADD23: every completion of Red excluded drops base Red (and its completes); Blue stays.
func TestPlan_ADD23_BaseLosesEveryCompletion(t *testing.T) {
	prev := planConfig(planLine(attrColor, modeInstant, "red", "blue"), planLine(attrSize, modeDynamic, "s", "m"))
	next := prev
	next.Exclusions = []PlanExclusion{{LeftValueId: "red", RightValueId: "s"}, {LeftValueId: "m", RightValueId: "red"}}
	plan := PlanTemplateVariants(PlanInput{Previous: &prev, Next: next, Existing: []PlanVariant{
		planActiveVariant("base-red", kindBase, planKey(attrColor, "red")),
		planActiveVariant("base-blue", kindBase, planKey(attrColor, "blue")),
		planActiveVariant("red-m", kindFull, planKey(attrColor, "red", attrSize, "m")),
	}})
	require.False(t, plan.HasViolations())
	require.Equal(t, []string{"base-blue"}, plannedIds(plan.Keep))
	require.Equal(t, []string{"base-red", "red-m"}, plannedIds(plan.Invalid))
	require.Equal(t, kindBase, plan.Invalid[0].Kind, "bases are reconciled before completes")
	require.Empty(t, plan.Create)
}

// ADD25: a new dynamic value creates nothing; existing completes stay and point at their base.
func TestPlan_ADD25_AddDynamicValue(t *testing.T) {
	prev := planConfig(planLine(attrColor, modeInstant, "red"), planLine(attrSize, modeDynamic, "s"))
	plan := PlanTemplateVariants(PlanInput{
		Previous: &prev,
		Next:     planConfig(planLine(attrColor, modeInstant, "red"), planLine(attrSize, modeDynamic, "s", "m")),
		Existing: []PlanVariant{
			planActiveVariant("base-red", kindBase, planKey(attrColor, "red")),
			planActiveVariant("red-s", kindFull, planKey(attrColor, "red", attrSize, "s")),
		},
	})
	require.Empty(t, plan.Create)
	require.Equal(t, []string{"base-red", "red-s"}, plannedIds(plan.Keep))
	require.Equal(t, planKey(attrColor, "red"), plan.Keep[1].BaseKey)
}

// ADD26: adding a dynamic dimension never extends; the complete becomes invalid and a base is created.
func TestPlan_ADD26_AddDynamicDimension(t *testing.T) {
	prev := planConfig(planLine(attrColor, modeInstant, "red"))
	cases := []struct {
		name   string
		values []string
	}{{"several values", []string{"s", "m"}}, {"single value", []string{"s"}}}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			plan := PlanTemplateVariants(PlanInput{
				Previous: &prev,
				Next:     planConfig(planLine(attrColor, modeInstant, "red"), planLine(attrSize, modeDynamic, tc.values...)),
				Existing: []PlanVariant{planActiveVariant("v10", kindFull, planKey(attrColor, "red"))},
			})
			require.Empty(t, plan.Extend)
			require.Equal(t, []string{"v10"}, plannedIds(plan.Invalid))
			require.Len(t, plan.Create, 1)
			require.Equal(t, kindBase, plan.Create[0].Kind)
		})
	}
}

// ADD27: removing the last dynamic dimension invalidates bases and A+D completes and creates A completes.
func TestPlan_ADD27_RemoveLastDynamic(t *testing.T) {
	prev := planConfig(planLine(attrColor, modeInstant, "red"), planLine(attrSize, modeDynamic, "s"))
	plan := PlanTemplateVariants(PlanInput{
		Previous: &prev,
		Next:     planConfig(planLine(attrColor, modeInstant, "red")),
		Existing: []PlanVariant{
			planActiveVariant("base-red", kindBase, planKey(attrColor, "red")),
			planActiveVariant("red-s", kindFull, planKey(attrColor, "red", attrSize, "s")),
		},
	})
	require.Equal(t, []string{"base-red", "red-s"}, plannedIds(plan.Invalid))
	require.Len(t, plan.Create, 1)
	require.Equal(t, kindFull, plan.Create[0].Kind)
	require.Equal(t, planKey(attrColor, "red"), plan.Create[0].Key)
}

// ADD28: a single-value instant dimension extends bases and completes of the same kind.
func TestPlan_ADD28_ExtendMixedSameKind(t *testing.T) {
	prev := planConfig(planLine(attrColor, modeInstant, "red"), planLine(attrSize, modeDynamic, "s", "m"))
	next := planConfig(planLine(attrColor, modeInstant, "red"), planLine(attrSize, modeDynamic, "s", "m"),
		planLine(attrMaterial, modeInstant, "cotton"))
	plan := PlanTemplateVariants(PlanInput{Previous: &prev, Next: next, Existing: []PlanVariant{
		planActiveVariant("red-s", kindFull, planKey(attrColor, "red", attrSize, "s")),
		planActiveVariant("base-red", kindBase, planKey(attrColor, "red")),
	}})
	require.False(t, plan.HasViolations())
	require.Len(t, plan.Extend, 2)
	require.Equal(t, "base-red", plan.Extend[0].VariantId, "bases first")
	require.Equal(t, planKey(attrColor, "red", attrMaterial, "cotton"), plan.Extend[0].NewKey)
	require.Equal(t, planKey(attrColor, "red", attrMaterial, "cotton", attrSize, "s"), plan.Extend[1].NewKey)
	require.Equal(t, planKey(attrColor, "red", attrMaterial, "cotton"), plan.Extend[1].BaseKey)
	require.Empty(t, plan.Create)
}

// ADD36: same key but a different historical role is a conflict, never a restore.
func TestPlan_ADD36_RoleConflict(t *testing.T) {
	variant := planArchivedVariant("v1", kindFull, planKey(attrColor, "red"), models.ArchiveSourceSystemSync)
	variant.Roles = map[string]models.VariantCreationMode{attrColor: modeInstant}
	next := planConfig(planLine(attrColor, modeDynamic, "red"))
	plan := PlanTemplateVariants(PlanInput{Next: next, Existing: []PlanVariant{variant}})
	requireViolation(t, plan, ReasonPlanRoleConflict)
	require.Equal(t, []string{"v1"}, plannedIds(plan.Conflict))

	variant.Roles[attrColor] = modeDynamic
	plan = PlanTemplateVariants(PlanInput{Next: next, Existing: []PlanVariant{variant}})
	require.False(t, plan.HasViolations())
	require.Equal(t, []string{"v1"}, plannedIds(plan.Restore))
}

// ADD38: few bases but over 1000 possible completes is not refused and pre-creates only bases.
func TestPlan_ADD38_PossibleOverLimitAllowed(t *testing.T) {
	plan := PlanTemplateVariants(PlanInput{Next: planConfig(
		planLine(attrColor, modeInstant, "red", "blue"),
		planLine(attrSize, modeDynamic, numberedValues("s", 40)...),
		planLine(attrMaterial, modeDynamic, numberedValues("m", 40)...),
	)})
	require.False(t, plan.HasViolations())
	require.Len(t, plan.Create, 2)
	requirePossible(t, plan, 3200)

	plan = PlanTemplateVariants(PlanInput{Next: planConfig(
		planLine(attrColor, modeInstant, "red", "blue"),
		planLine(attrSize, modeDynamic, numberedValues("s", 40)...),
		planLine(attrMaterial, modeDynamic, numberedValues("m", 40)...),
	), CountCap: 1000})
	require.False(t, plan.HasViolations())
	require.True(t, plan.PossibleOverThreshold)
	require.Nil(t, plan.PossibleCompleteCount)
}

// CR02 §4.2: exclusions on never values, unselected values or within one attribute are inactive.
func TestPlan_InactiveExclusions(t *testing.T) {
	config := planConfig(planLine(attrColor, modeInstant, "red", "blue"), planLine(attrGift, modeNever, "yes"))
	config.Exclusions = []PlanExclusion{
		{LeftValueId: "red", RightValueId: "yes"},
		{LeftValueId: "red", RightValueId: "green"},
		{LeftValueId: "red", RightValueId: "blue"},
	}
	plan := PlanTemplateVariants(PlanInput{Next: config})
	require.Len(t, plan.InactiveExclusions, 3)
	require.Len(t, plan.Create, 2)
}

func TestPlan_DuplicateLineRefused(t *testing.T) {
	plan := PlanTemplateVariants(PlanInput{Next: planConfig(
		planLine(attrColor, modeInstant, "red"), planLine(attrColor, modeDynamic, "blue"))})
	requireViolation(t, plan, ReasonPlanDuplicateLine)
}

// ADD20 / AC27 / D4: resolve selection validation.
func TestValidateSelection(t *testing.T) {
	config := planConfig(planLine(attrColor, modeInstant, "red", "blue"), planLine(attrSize, modeDynamic, "s", "m"),
		planLine(attrGift, modeNever, "yes", "no"))
	cases := []struct {
		name      string
		selection map[string]string
		reason    string
	}{
		{"missing dynamic (ADD20)", map[string]string{attrColor: "red"}, ReasonSelectionMissingAttr},
		{"unknown attribute", map[string]string{attrColor: "red", attrSize: "s", "a_other": "x"},
			ReasonSelectionUnknownAttr},
		{"value outside line (AC27)", map[string]string{attrColor: "green", attrSize: "s"}, ReasonSelectionValueNotAllow},
		{"wrong never value", map[string]string{attrColor: "red", attrSize: "s", attrGift: "maybe"},
			ReasonSelectionValueNotAllow},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, violations := ValidateSelection(config, tc.selection)
			require.NotEmpty(t, violations)
			require.Equal(t, "inventory:"+tc.reason, violations[0].Key)
		})
	}

	withGift, base, violations := ValidateSelection(config,
		map[string]string{attrColor: "red", attrSize: "s", attrGift: "yes"})
	require.Empty(t, violations)
	withoutGift, _, _ := ValidateSelection(config, map[string]string{attrColor: "red", attrSize: "s"})
	require.Equal(t, withoutGift, withGift, "never values stay out of the key (ADD18)")
	require.Equal(t, planKey(attrColor, "red"), base)

	dynamicOnly := planConfig(planLine(attrSize, modeDynamic, "s"))
	completeKey, baseKey, _ := ValidateSelection(dynamicOnly, map[string]string{attrSize: "s"})
	require.Equal(t, planKey(attrSize, "s"), completeKey)
	require.Empty(t, baseKey, "dynamic-only templates have no base")
}

func numberedValues(prefix string, count int) []string {
	values := make([]string, 0, count)
	for i := 0; i < count; i++ {
		values = append(values, fmt.Sprintf("%s%03d", prefix, i))
	}
	return values
}
