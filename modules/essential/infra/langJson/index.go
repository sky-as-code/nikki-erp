package langjson

import (
	"embed"
	"encoding/json"
	"path"
	"strings"

	"go.bryk.io/pkg/errors"
)

//go:embed en-US vi-VN
var langFiles embed.FS

// CommonModuleName is the module file stem (without ".json") whose result is enriched with
// "module.label.*" keys aggregated from every other module of the same locale.
const CommonModuleName = "common"

// ModuleLabelKeyPrefix selects keys to bubble up from non-common modules into the common module.
const ModuleLabelKeyPrefix = "module.label"

// ByLanguageCodeAndModule maps locale (e.g. en-US) to module file stem (e.g. common) to the JSON object at file root.
var ByLanguageCodeAndModule map[string]map[string]map[string]any

func init() {
	loaded, err := loadEmbeddedLangJson()
	if err != nil {
		panic(errors.Wrap(err, "langjson: load embedded language json"))
	}
	ByLanguageCodeAndModule = loaded
}

func loadEmbeddedLangJson() (map[string]map[string]map[string]any, error) {
	out := make(map[string]map[string]map[string]any)
	locales := []string{"en-US", "vi-VN"}
	for _, locale := range locales {
		entries, err := langFiles.ReadDir(locale)
		if err != nil {
			return nil, errors.Wrap(err, locale)
		}
		modules := make(map[string]map[string]any)
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
				continue
			}
			relPath := path.Join(locale, entry.Name())
			raw, err := langFiles.ReadFile(relPath)
			if err != nil {
				return nil, errors.Wrap(err, relPath)
			}
			var root map[string]any
			if err := json.Unmarshal(raw, &root); err != nil {
				return nil, errors.Wrap(err, relPath)
			}
			moduleKey := strings.TrimSuffix(entry.Name(), ".json")
			modules[moduleKey] = root
		}
		out[locale] = modules
	}
	return out, nil
}

// MergedCommonMessages returns "common" module's messages, merged with
// every "module.label.*" key found in the other modules of the same locale.
// Returns nil when the locale has no "common" module.
func MergedCommonMessages(byModule map[string]map[string]any) map[string]any {
	base, hasCommon := byModule[CommonModuleName]
	if !hasCommon {
		return nil
	}
	for moduleName, data := range byModule {
		if moduleName == CommonModuleName {
			continue
		}
		for key, value := range data {
			if strings.HasPrefix(key, ModuleLabelKeyPrefix) {
				base[key] = value
			}
		}
	}
	return base
}
