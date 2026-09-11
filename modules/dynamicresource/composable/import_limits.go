package composable

import (
	"strconv"
	"strings"

	"github.com/sky-as-code/nikki-erp/modules/core/config"
	"github.com/sky-as-code/nikki-erp/modules/core/constants"
)

const (
	defaultImportMaxFileSizeBytes int64 = 10 << 20
	defaultImportMaxRows          int   = 10_000
)

var defaultImportAllowedExtensions = []string{"xlsx", "csv"}

// ImportLimits is what the Configuration Service allows one import request to carry.
type ImportLimits struct {
	MaxFileSizeBytes  int64
	AllowedExtensions []string
	MaxRows           int
}

// readImportLimits never fails: a missing key falls back to the compiled default so an
// environment without the CORE.IMPORT section still imports with sane bounds.
func readImportLimits(cfg config.ConfigService) ImportLimits {
	if cfg == nil {
		return ImportLimits{
			MaxFileSizeBytes:  defaultImportMaxFileSizeBytes,
			AllowedExtensions: defaultImportAllowedExtensions,
			MaxRows:           defaultImportMaxRows,
		}
	}
	// ConfigService casts the fallback to string before parsing, hence the formatted defaults.
	return ImportLimits{
		MaxFileSizeBytes:  cfg.GetInt64(constants.ImportMaxFileSizeBytes, strconv.FormatInt(defaultImportMaxFileSizeBytes, 10)),
		AllowedExtensions: normalizeExtensions(cfg.GetStrArr(constants.ImportAllowedExtensions, strings.Join(defaultImportAllowedExtensions, ","))),
		MaxRows:           cfg.GetInt(constants.ImportMaxRows, strconv.Itoa(defaultImportMaxRows)),
	}
}

// normalizeExtensions lowercases, trims and strips a leading dot so ".XLSX" and "xlsx" match.
func normalizeExtensions(raw []string) []string {
	result := make([]string, 0, len(raw))
	for _, ext := range raw {
		ext = strings.ToLower(strings.TrimPrefix(strings.TrimSpace(ext), "."))
		if ext != "" {
			result = append(result, ext)
		}
	}
	return result
}

// IsAllowed reports whether a filename extension (with or without the dot) is importable.
func (this ImportLimits) IsAllowed(ext string) bool {
	ext = strings.ToLower(strings.TrimPrefix(strings.TrimSpace(ext), "."))
	for _, allowed := range this.AllowedExtensions {
		if allowed == ext {
			return true
		}
	}
	return false
}
