package app

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	coreSettings "github.com/sky-as-code/nikki-erp/modules/core/settings"
	c "github.com/sky-as-code/nikki-erp/modules/essential/constants"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/settings"
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
		c.EssentialModuleName, schemas, settingsSvc, enabledLanguageRule(languageSvc),
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

// enabledLanguageRule refuses a language a user could not read the interface in: only the
// essential_languages rows marked is_enabled ship translations.
func enabledLanguageRule(languageSvc itLanguage.LanguageDomainService) coreSettings.SettingsRule {
	return func(ctx corectx.Context, level string, items []coreSettings.SetSettingItem) (ft.ClientErrors, error) {
		vErrs := ft.ClientErrors{}
		for _, item := range items {
			code, isString := item.Value.(string)
			if !isString || !isLanguageSetting(item.Name) {
				continue
			}
			found, err := languageSvc.GetLanguageByIsoCode(ctx, itLanguage.GetLanguageByIsoCodeQuery{IsoCode: code})
			if err != nil {
				return nil, errors.Wrap(err, "enabledLanguageRule")
			}
			if !found.HasData || found.Data.GetIsEnabled() == nil || !*found.Data.GetIsEnabled() {
				vErrs.Append(*ft.NewValidationError(item.Name, "essential.language_not_enabled",
					"is not a language this application is enabled for"))
			}
		}
		return vErrs, nil
	}
}

func isLanguageSetting(name string) bool {
	return name == settings.UserSettingLanguage || name == settings.OrgSettingSystemLanguage
}
