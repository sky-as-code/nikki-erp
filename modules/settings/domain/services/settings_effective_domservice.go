package services

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	c "github.com/sky-as-code/nikki-erp/modules/settings/constants"
	it "github.com/sky-as-code/nikki-erp/modules/settings/interfaces/settings"
)

// effectiveLevels is the order the levels are folded in.
//
// Most specific last, so that if a name ever did appear at two levels the owner's own value would
// be the one left standing. In practice it cannot: a setting name is unique within a module across
// all of its levels, which is the invariant that lets the result be a flat map at all. The order is
// a safety net, not a merge strategy.
var effectiveLevels = []struct {
	level     string
	ownerType string
}{
	{c.LevelTenant, c.OwnerTypeTenant},
	{c.LevelOrg, c.OwnerTypeOrg},
	{c.LevelUser, c.OwnerTypeUser},
}

// GetEffectiveSettings answers "what settings apply to me", across every level at once.
//
// The three per-level reads it is built from are unchanged: each already resolves schema defaults
// and unwraps the stored {"value": ...} envelope, so this adds the flattening and nothing else.
func (this *SettingsDomainServiceImpl) GetEffectiveSettings(
	ctx corectx.Context, query it.GetEffectiveSettingsQuery,
) (*it.GetEffectiveSettingsResult, error) {
	values := map[string]any{}

	for _, moduleKey := range withEssentialKey(query.ModuleKeys) {
		for _, scope := range effectiveLevels {
			result, err := this.GetSettings(ctx, scope.level, scope.ownerType, it.GetSettingsQuery{
				ModuleKey: moduleKey,
			})
			// A level with no owner behind it -- no tenant in the nikkierp binary, no single
			// organization for a user who belongs to none or to several -- contributes nothing.
			// Failing the whole read instead would make every caller of this depend on the acting
			// user's org membership, which is not what "what applies to me" should hinge on.
			if err != nil || result == nil || !result.HasData {
				continue
			}
			for _, item := range result.Data.Items {
				if item.Value == nil {
					continue
				}
				values[moduleKey+"."+item.Name] = item.Value
			}
		}
	}

	return &it.GetEffectiveSettingsResult{
		Data:    it.GetEffectiveSettingsResultData{Values: values},
		HasData: true,
	}, nil
}

// withEssentialKey folds Essential's key into the requested set, de-duplicated and order-preserving.
//
// Essential carries language and locale, which the platform itself reads on paths that have no idea
// which modules a feature needs -- the search layer localizing a jsonb column, chiefly. Making every
// such caller remember to ask for it would mean the one caller that forgot silently got no locale.
func withEssentialKey(moduleKeys []string) []string {
	keys := make([]string, 0, len(moduleKeys)+1)
	seen := map[string]bool{}
	for _, key := range append([]string{c.EssentialModuleKey}, moduleKeys...) {
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		keys = append(keys, key)
	}
	return keys
}
