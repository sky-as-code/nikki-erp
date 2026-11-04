package language

import (
	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	coreSettings "github.com/sky-as-code/nikki-erp/modules/core/settings"
	c "github.com/sky-as-code/nikki-erp/modules/essential/constants"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/settings"
)

// OrgLanguage is the language an organization's generated names are written in: its stored
// essential.system_language, or DefaultLanguageCode when unset or not a language code. It reads
// the org named explicitly, never the one guessed from the request (settings-003).
func OrgLanguage(
	ctx corectx.Context, reader coreSettings.OrgSettingsReader, orgId model.Id,
) (model.LanguageCode, error) {
	values, err := reader.GetOrgSettingValues(ctx, orgId, c.EssentialModuleName)
	if err != nil {
		return "", errors.Wrap(err, "read organization system language")
	}
	stored, _ := values[settings.OrgSettingSystemLanguage].(string)
	if stored == "" {
		return model.DefaultLanguageCode, nil
	}
	canonical, err := model.ToBCP47LanguageCode(stored)
	if err != nil || !model.IsBCP47LanguageCode(canonical) {
		return model.DefaultLanguageCode, nil
	}
	return canonical, nil
}
