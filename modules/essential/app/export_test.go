package app

import it "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/language"

// NewLanguageApplicationServiceForTest builds the language service without an onion: the
// translation-bundle methods under test never reach the CRUD or the domain service.
func NewLanguageApplicationServiceForTest() it.LanguageAppService {
	return newLanguageApplicationService(nil, nil)
}
