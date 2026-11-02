package app

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	coreSettings "github.com/sky-as-code/nikki-erp/modules/core/settings"
	c "github.com/sky-as-code/nikki-erp/modules/essential/constants"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/settings"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/timezone"
	itLanguage "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/language"
	itSettings "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/settings"
)

// NewSettingsApplicationServiceImpl validates Essential's settings against the schemas it
// registers and against essential_languages, then persists them through core's settings port.
//
// Authorization is the writer's first step: update on settings_record for tenant and org, the
// actor's own row for user -- the same check the Settings module's application services run.
func NewSettingsApplicationServiceImpl(
	settingsSvc coreSettings.SettingsService, languageSvc itLanguage.LanguageDomainService,
) itSettings.SettingsAppService {
	schemas := map[string]*dmodel.ModelSchema{
		coreSettings.LevelUser: settings.UserSettingsSchemaBuilder().Build(),
		coreSettings.LevelOrg:  settings.OrgSettingsSchemaBuilder().Build(),
	}
	writer := coreSettings.NewModuleSettingsWriter(
		c.EssentialModuleName, schemas, settingsSvc, enabledLanguageRule(languageSvc), knownTimezoneRule,
	)
	return &SettingsApplicationServiceImpl{writer: writer}
}

type SettingsApplicationServiceImpl struct {
	writer *coreSettings.ModuleSettingsWriter
}

func (this *SettingsApplicationServiceImpl) UpdateSettings(
	ctx corectx.Context, cmd coreSettings.UpdateSettingsCommand,
) (*coreSettings.SetSettingsResult, error) {
	return this.writer.Update(ctx, cmd)
}

// knownTimezoneRule refuses a timezone outside the hardcoded catalogue the Settings page picks from.
// Clearing one is left to the schema's length rule.
func knownTimezoneRule(_ corectx.Context, _ string, items []coreSettings.SetSettingItem) (ft.ClientErrors, error) {
	vErrs := ft.ClientErrors{}
	for _, item := range items {
		zone, isString := item.Value.(string)
		if !isString || zone == "" || !isTimezoneSetting(item.Name) || timezone.IsKnown(zone) {
			continue
		}
		vErrs.Append(*ft.NewValidationError(item.Name, ft.ErrorKey("essential.timezone_unknown", "essential"),
			"is not a known timezone"))
	}
	return vErrs, nil
}

func isTimezoneSetting(name string) bool {
	return name == settings.UserSettingTimezone || name == settings.OrgSettingSystemTimezone
}

// enabledLanguageRule refuses a language a user could not read the interface in: only the languages
// the caller's organization has active (else the public ones) are offered, so only those are taken.
func enabledLanguageRule(languageSvc itLanguage.LanguageDomainService) coreSettings.SettingsRule {
	return func(ctx corectx.Context, level string, items []coreSettings.SetSettingItem) (ft.ClientErrors, error) {
		vErrs := ft.ClientErrors{}
		var enabled map[string]bool
		for _, item := range items {
			code, isString := item.Value.(string)
			if !isString || !isLanguageSetting(item.Name) {
				continue
			}
			if enabled == nil {
				codes, err := enabledLocaleCodes(ctx, languageSvc)
				if err != nil {
					return nil, errors.Wrap(err, "enabledLanguageRule")
				}
				enabled = codes
			}
			if !enabled[code] {
				vErrs.Append(*ft.NewValidationError(item.Name, ft.ErrorKey("essential.language_not_enabled", "essential"),
					"is not a language this application is enabled for"))
			}
		}
		return vErrs, nil
	}
}

func enabledLocaleCodes(ctx corectx.Context, languageSvc itLanguage.LanguageDomainService) (map[string]bool, error) {
	listed, err := languageSvc.ListEnabledLanguages(ctx, itLanguage.ListEnabledLanguagesQuery{})
	if err != nil {
		return nil, err
	}
	codes := make(map[string]bool, len(listed.Data))
	for _, language := range listed.Data {
		if code := language.GetLocaleCode(); code != nil {
			codes[*code] = true
		}
	}
	return codes, nil
}

func isLanguageSetting(name string) bool {
	return name == settings.UserSettingLanguage || name == settings.OrgSettingSystemLanguage
}
