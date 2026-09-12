// Package settings holds Notification's settings schemas. Unlike a model schema, a settings schema
// owns no table: its values are stored as settings_records rows by the settings module.
package settings

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

// The setting names Notification declares.
const (
	// OrgSettingWebEnabled turns the web channel on or off for an organization. Disabling it does
	// not stop notifications being recorded — the inbox is the source of truth and stays readable
	// (BR-FS 22) — it only marks the web delivery skipped.
	OrgSettingWebEnabled = "web_enabled"

	// OrgSettingStreamHeartbeatIntervalSeconds is how often an idle stream emits a heartbeat, so
	// that a dead connection is noticed and no proxy closes a quiet one as idle (BR-FS 9). The
	// gateway's own idle timeout must stay comfortably above it.
	OrgSettingStreamHeartbeatIntervalSeconds = "stream_heartbeat_interval_seconds"

	// OrgSettingStreamMaxBufferedEvents caps how far one slow client may fall behind before its
	// stream is closed (BR-FS 19). Closing it is deliberate and safe: the client reconnects with
	// after_seq and replays what it missed, which costs far less than an unbounded per-client
	// buffer holding memory for a browser that may be suspended.
	OrgSettingStreamMaxBufferedEvents = "stream_max_buffered_events"
)

// Fallbacks used when an organization has set nothing. They mirror the JSON default_value, which
// the settings module applies on read; they exist for the paths that resolve a setting before the
// organization has any record at all.
const (
	DefaultWebEnabled                     = true
	DefaultStreamHeartbeatIntervalSeconds = int32(20)
	DefaultStreamMaxBufferedEvents        = int32(100)
)

// OrgSettingsSchemaName is the name Notification registers its org-level settings under. Not a
// table: the values live in settings_records.
const OrgSettingsSchemaName = "notification_org_settings"

//go:embed org_settings.json
var orgSettingsSchemaJson string

// OrgSettingsSchemaBuilder declares the notification policy an organization sets for itself. The
// document declares no table_name or should_build_db and extends no base model: the basemodel
// mixins would inject tenant_id and audit columns onto a schema with no table to put them in.
func OrgSettingsSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(orgSettingsSchemaJson)
}
