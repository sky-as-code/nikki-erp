-- Default Essential org settings for every seeded org: system_language vi-VN (what Essential's InitOrg
-- writes for a new org) and system_timezone Asia/Ho_Chi_Minh.
--
-- These are Essential's org settings, but they live here rather than in an essential seed file
-- because settings_records is created by 0003001_settings_schema.sql; keeping each table's data with
-- its owning module's number mirrors the hosted app's scripts/migration-seeds/0003003_settings_defaults.sql.
DO $$
DECLARE
	essential_org_schema_id varchar;
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'settings_records'
	) THEN
		-- Seeds run before the first application boot, and settings_records.schema_id is NOT NULL.
		-- Snapshot of essential/domain/settings/org_settings.json in the ToModelJson shape; startup
		-- registration matches on (module_key, level), keeps this row's id and refreshes the document.
		INSERT INTO "settings_schemas" ("id", "module_key", "level", "schema", "created_at")
		VALUES ('01KQCMR00000000000SETSCH01', 'essential', 'org',
			$essential_org_schema${
  "name": "essential_org_settings",
  "label": {"$ref": "essential.settings.orgSettings"},
  "fields": [
    {"data_type": {"max": 10, "min": 5, "type": "string"}, "description": {"$ref": "settings_desc.system_language"}, "label": {"$ref": "essential.settings.systemLanguage"}, "metadata": {"allow_override": true}, "name": "system_language"},
    {"data_type": {"max": 64, "min": 1, "type": "string"}, "description": {"$ref": "settings_desc.system_timezone"}, "label": {"$ref": "essential.settings.systemTimezone"}, "metadata": {"allow_override": true}, "name": "system_timezone"}
  ]
}$essential_org_schema$::jsonb, NOW())
		ON CONFLICT ("module_key", "level") DO NOTHING;

		SELECT "id" INTO STRICT essential_org_schema_id
		FROM "settings_schemas" WHERE "module_key" = 'essential' AND "level" = 'org';

		INSERT INTO "settings_records" (
			"id", "schema_id", "module_key", "level", "owner_type", "owner_id",
			"name", "value", "allow_override", "created_at", "updated_at"
		) VALUES
		('01M3SETR000000000000001001', essential_org_schema_id, 'essential', 'org', 'org', '01JWNY20G23KD4RV5VWYABQYHD', 'system_language', jsonb_build_object('value', 'vi-VN'), NULL, NOW(), NULL),
		('01M3SETR000000000000001002', essential_org_schema_id, 'essential', 'org', 'org', '01JWNY20G23KD4RV5VWYABQYHD', 'system_timezone', jsonb_build_object('value', 'Asia/Ho_Chi_Minh'), NULL, NOW(), NULL),
		('01M3SETR000000000000002001', essential_org_schema_id, 'essential', 'org', 'org', '01K02G6J1CYAN9K8V4PAGSQ5Z8', 'system_language', jsonb_build_object('value', 'vi-VN'), NULL, NOW(), NULL),
		('01M3SETR000000000000002002', essential_org_schema_id, 'essential', 'org', 'org', '01K02G6J1CYAN9K8V4PAGSQ5Z8', 'system_timezone', jsonb_build_object('value', 'Asia/Ho_Chi_Minh'), NULL, NOW(), NULL),
		('01M3SETR000000000000003001', essential_org_schema_id, 'essential', 'org', 'org', '01K1H7M2K9VW3P5R7XQJY2C1Z9', 'system_language', jsonb_build_object('value', 'vi-VN'), NULL, NOW(), NULL),
		('01M3SETR000000000000003002', essential_org_schema_id, 'essential', 'org', 'org', '01K1H7M2K9VW3P5R7XQJY2C1Z9', 'system_timezone', jsonb_build_object('value', 'Asia/Ho_Chi_Minh'), NULL, NOW(), NULL)
		ON CONFLICT ("module_key", "name", "owner_id") DO NOTHING;
	END IF;
END $$;
