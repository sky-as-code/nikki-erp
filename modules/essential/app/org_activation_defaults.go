package app

import (
	"encoding/json"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	c "github.com/sky-as-code/nikki-erp/modules/essential/constants"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/settings"
	orgDefaults "github.com/sky-as-code/nikki-erp/modules/essential/infra/defaults"
	itCurrency "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/currency"
	itExt "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/external"
)

const activationDefaultsFile = "essential_org_activation.json"

// publicLanguagesLimit bounds the read of the languages a new org starts with.
const publicLanguagesLimit = 500

// activationDefaults names, by code, what a new org starts with besides the public languages.
type activationDefaults struct {
	CurrencyCodes  []string `json:"currency_codes"`
	SystemLanguage string   `json:"system_language"`
}

func loadActivationDefaults() (*activationDefaults, error) {
	raw, err := orgDefaults.Files.ReadFile(activationDefaultsFile)
	if err != nil {
		return nil, errors.Wrap(err, "loadActivationDefaults")
	}
	out := &activationDefaults{}
	return out, errors.Wrap(json.Unmarshal(raw, out), "loadActivationDefaults")
}

// initActivations activates the public languages and the default currencies for a new org, and
// sets its system language. A default currency the catalogue lacks is an error: the defaults file
// and the seeds disagree.
func (this *OrgLifecycleAppServiceImpl) initActivations(ctx corectx.Context, orgId model.Id) error {
	defaults, err := loadActivationDefaults()
	if err != nil {
		return err
	}
	languageIds, err := this.publicLanguageIds(ctx)
	if err != nil {
		return err
	}
	for _, id := range languageIds {
		if err := this.param.Languages.SetActive(ctx, orgId, id, true); err != nil {
			return err
		}
	}
	for _, code := range defaults.CurrencyCodes {
		found, err := this.param.CurrencySvc.GetCurrencyByCode(ctx, itCurrency.GetCurrencyByCodeQuery{Code: code})
		if err != nil {
			return err
		}
		if !found.HasData {
			return errors.Errorf("default currency %s is not in the catalogue", code)
		}
		if err := this.param.Currencies.SetActive(ctx, orgId, found.Data.Id, true); err != nil {
			return err
		}
	}
	return this.seedSystemLanguage(ctx, defaults.SystemLanguage)
}

func (this *OrgLifecycleAppServiceImpl) publicLanguageIds(ctx corectx.Context) ([]model.Id, error) {
	found, err := this.param.LanguageSvc.Repository().Search(ctx, dyn.RepoSearchParam{
		Fields: []string{models.LanguageFieldId},
		Filter: []dmodel.DynamicFields{{models.LanguageFieldPublicUse: true}},
		Size:   publicLanguagesLimit,
	})
	if err != nil {
		return nil, errors.Wrap(err, "publicLanguageIds")
	}
	ids := make([]model.Id, 0, len(found.Data.Items))
	for _, row := range found.Data.Items {
		if id, ok := row[models.LanguageFieldId].(string); ok {
			ids = append(ids, model.Id(id))
		}
	}
	return ids, nil
}

func (this *OrgLifecycleAppServiceImpl) seedSystemLanguage(ctx corectx.Context, localeCode string) error {
	if localeCode == "" {
		return nil
	}
	result, err := this.param.SettingsSeed.SeedOrgSettings(ctx, itExt.SetSettingsCommand{
		ModuleKey: c.EssentialModuleName,
		Items:     []itExt.SetSettingItem{{Name: settings.OrgSettingSystemLanguage, Value: localeCode}},
	})
	if err != nil {
		return err
	}
	if result.ClientErrors.Count() > 0 {
		return errors.Wrap(result.ClientErrors.ToError(), "seed the system language")
	}
	return nil
}
