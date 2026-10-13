package langjson

import (
	"embed"
	"encoding/json"
	"maps"
	"path"
	"strings"

	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/model"
)

//go:embed en-US vi-VN
var langFiles embed.FS

// CommonModuleName is the module file stem (without ".json") whose result is enriched with
// "module.label.*" keys aggregated from every other module of the same locale.
const CommonModuleName = "common"

// ModuleLabelKeyPrefix selects keys to bubble up from non-common modules into the common module.
const ModuleLabelKeyPrefix = "module.label"

// EmbeddedPack holds the translations essential embeds, keyed by module file stem (e.g. common),
// then by locale (e.g. en-US). Modules that ship their own register them with the language
// application service instead.
var EmbeddedPack model.LanguagePack

func init() {
	loaded, err := loadEmbeddedPack()
	if err != nil {
		panic(errors.Wrap(err, "langjson: load embedded language json"))
	}
	EmbeddedPack = loaded
}

func loadEmbeddedPack() (model.LanguagePack, error) {
	out := make(model.LanguagePack)
	for _, locale := range []model.LanguageCode{model.LanguageCodeEnUs, model.LanguageCodeViVn} {
		entries, err := langFiles.ReadDir(locale)
		if err != nil {
			return nil, errors.Wrap(err, locale)
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
				continue
			}
			relPath := path.Join(locale, entry.Name())
			raw, err := langFiles.ReadFile(relPath)
			if err != nil {
				return nil, errors.Wrap(err, relPath)
			}
			var item model.LanguageTranslationItem
			if err := json.Unmarshal(raw, &item); err != nil {
				return nil, errors.Wrap(err, relPath)
			}
			moduleName := strings.TrimSuffix(entry.Name(), ".json")
			if out[moduleName] == nil {
				out[moduleName] = make(model.LanguageTranslation)
			}
			out[moduleName][locale] = item
		}
	}
	return out, nil
}

// HasLanguage reports whether any module in pack has a translation for langCode.
func HasLanguage(pack model.LanguagePack, langCode model.LanguageCode) bool {
	for _, translation := range pack {
		if _, ok := translation[langCode]; ok {
			return true
		}
	}
	return false
}

// MergedCommonMessages returns a copy of the "common" item, overlaid with every "module.label.*" key
// found in the langCode item of each other module in packs. It copies because common is the shared
// embedded map: merging into it would leak one request's overlay into every later one.
func MergedCommonMessages(
	common model.LanguageTranslationItem, langCode model.LanguageCode, packs ...model.LanguagePack,
) model.LanguageTranslationItem {
	merged := make(model.LanguageTranslationItem, len(common))
	maps.Copy(merged, common)
	for _, pack := range packs {
		for moduleName, translation := range pack {
			if moduleName == CommonModuleName {
				continue
			}
			for key, value := range translation[langCode] {
				if strings.HasPrefix(key, ModuleLabelKeyPrefix) {
					merged[key] = value
				}
			}
		}
	}
	return merged
}
