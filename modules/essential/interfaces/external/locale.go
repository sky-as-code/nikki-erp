package external

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	itSettings "github.com/sky-as-code/nikki-erp/modules/settings/interfaces/settings"
)

// EffectiveSettingsExtService is Essential's port onto the Settings module's level-agnostic read.
//
// Narrowed to the single read, on the same reasoning as SettingsRegistrationExtService: Essential
// needs to know which language the acting user reads in, and nothing more. A port that could also
// write would hand that reach to everything holding it.
type EffectiveSettingsExtService interface {
	GetEffectiveSettings(
		ctx corectx.Context, query GetEffectiveSettingsQuery,
	) (*GetEffectiveSettingsResult, error)
}

type GetEffectiveSettingsQuery = itSettings.GetEffectiveSettingsQuery
type GetEffectiveSettingsResult = itSettings.GetEffectiveSettingsResult
