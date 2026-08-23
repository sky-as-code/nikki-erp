package services

import (
	"testing"

	"github.com/stretchr/testify/assert"

	c "github.com/sky-as-code/nikki-erp/modules/settings/constants"
)

// Essential carries language and locale, which the platform reads on paths that have no idea which
// modules a feature needs. A caller that asked only for its own module must still get them, or the
// one caller that forgot would silently run unlocalized.
func TestWithEssentialKey_AlwaysIncludesEssential(t *testing.T) {
	assert.Equal(t, []string{c.EssentialModuleKey}, withEssentialKey(nil))
	assert.Equal(t, []string{c.EssentialModuleKey}, withEssentialKey([]string{}))
	assert.Equal(t,
		[]string{c.EssentialModuleKey, "inventory"},
		withEssentialKey([]string{"inventory"}))
}

// Asking for essential explicitly must not read it twice.
func TestWithEssentialKey_DoesNotDuplicate(t *testing.T) {
	assert.Equal(t,
		[]string{c.EssentialModuleKey, "iam"},
		withEssentialKey([]string{c.EssentialModuleKey, "iam", c.EssentialModuleKey}))
}

// An empty string is not a module, and reading one would search for a schema that cannot exist.
func TestWithEssentialKey_DropsEmptyKeys(t *testing.T) {
	assert.Equal(t,
		[]string{c.EssentialModuleKey, "iam"},
		withEssentialKey([]string{"", "iam", ""}))
}

// The caller's order is kept so the fold is deterministic, which is what makes a repeated read
// return the same map rather than one that depends on map iteration order.
func TestWithEssentialKey_PreservesCallerOrder(t *testing.T) {
	assert.Equal(t,
		[]string{c.EssentialModuleKey, "purchase", "iam", "inventory"},
		withEssentialKey([]string{"purchase", "iam", "inventory"}))
}

// The three levels are folded most-specific-last, so an owner's own value would win if a name ever
// did appear at two levels.
func TestEffectiveLevels_AreOrderedMostSpecificLast(t *testing.T) {
	assert.Equal(t,
		[]string{c.LevelTenant, c.LevelOrg, c.LevelUser},
		[]string{effectiveLevels[0].level, effectiveLevels[1].level, effectiveLevels[2].level})
}
