package essential

import (
	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/modules"
	itLang "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/language"
)

// registerModuleLanguages collects the packs of every loaded module that ships its own. It needs
// only the modules' data, not their state, so running early in the start-up order is harmless.
func registerModuleLanguages(langSvc itLang.LanguageAppService, loadedMods []modules.InCodeModule) error {
	for _, mod := range loadedMods {
		langMod, ok := mod.(modules.InCodeModuleRegisterLanguage)
		if !ok {
			continue
		}
		if err := langSvc.RegisterModuleLanguage(mod.Name(), langMod.RegisterLanguage()); err != nil {
			return errors.Wrapf(err, "module '%s'", mod.Name())
		}
	}
	return nil
}
