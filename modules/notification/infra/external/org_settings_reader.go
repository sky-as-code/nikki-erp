package external

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	modconstants "github.com/sky-as-code/nikki-erp/modules/notification/constants"
	itExt "github.com/sky-as-code/nikki-erp/modules/notification/interfaces/external"
)

// OrgSettingsReader reads Notification's own organization-level settings.
//
// It exists so that a channel asking "am I switched on here" writes one line instead of unpacking
// the settings envelope itself. Every channel would otherwise repeat the same lookup, and a
// difference between two copies of it would be a channel that silently disagrees about whether it
// is enabled.
type OrgSettingsReader struct {
	settings itExt.OrgSettingsReadExtService
	logger   logging.LoggerService
}

func NewOrgSettingsReader(
	settings itExt.OrgSettingsReadExtService, logger logging.LoggerService,
) *OrgSettingsReader {
	return &OrgSettingsReader{
		settings: settings,
		logger:   logger,
	}
}

// Bool reads one boolean setting, falling back to fallback when it cannot be read.
//
// A missing or unreadable setting is not an error to the caller. The settings module already fills
// in a schema default for a name with no row, so the fallback is reached only when the read itself
// fails -- and refusing to send a notification because its channel's configuration could not be
// loaded would turn a settings outage into a business one (BR 32).
func (this *OrgSettingsReader) Bool(ctx corectx.Context, name string, fallback bool) bool {
	item := this.item(ctx, name)
	if item == nil {
		return fallback
	}

	value, isBool := item.Value.(bool)
	if !isBool {
		this.warnf("notification: setting '%s' is not a boolean; using the default", name)
		return fallback
	}

	return value
}

func (this *OrgSettingsReader) item(ctx corectx.Context, name string) *itExt.SettingItem {
	if this.settings == nil || ctx == nil {
		return nil
	}

	result, err := this.settings.GetOrgSettings(ctx, itExt.GetSettingsQuery{
		ModuleKey: modconstants.NotificationModuleName,
	})
	if err != nil {
		this.warnf("notification: the organization's settings could not be read: %v", err)
		return nil
	}
	if result == nil || !result.HasData {
		return nil
	}

	for index := range result.Data.Items {
		if result.Data.Items[index].Name == name {
			return &result.Data.Items[index]
		}
	}

	return nil
}

func (this *OrgSettingsReader) warnf(format string, args ...any) {
	if this.logger != nil {
		this.logger.Warnf(format, args...)
	}
}
