-- Create "essential_enums" table
CREATE TABLE "essential_enums" (
  "id" character varying NOT NULL,
  "label" jsonb NULL,
  "value" character varying NULL,
  "type" character varying NULL,
  "etag" character varying NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "created_by" character varying NULL,
  PRIMARY KEY ("id")
);
-- Create "essential_model_metadata" table
CREATE TABLE "essential_model_metadata" (
  "id" character varying NOT NULL,
  "name" jsonb NULL,
  "code" character varying NULL,
  "code_prefix" character varying NULL,
  "code_last_seq" integer NULL,
  "padding" integer NULL,
  "is_archived" boolean NULL,
  "etag" character varying NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "created_by" character varying NULL,
  PRIMARY KEY ("id")
);
-- Create "essential_modules" table
CREATE TABLE "essential_modules" (
  "id" character varying NOT NULL,
  "name" character varying NULL,
  "label" jsonb NULL,
  "is_orphaned" boolean NULL,
  "is_internal" boolean NULL,
  "version" character varying NULL,
  "has_ui" boolean NULL,
  "model_prefixes" character varying[] NULL,
  "category" character varying NULL,
  "is_enabled" boolean NULL,
  "dependencies" character varying[] NULL,
  "dep_level" integer NULL,
  "etag" character varying NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "created_by" character varying NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "essential_modules_name_ukey" UNIQUE ("name")
);
-- Create "essential_tags" table
CREATE TABLE "essential_tags" (
  "id" character varying NOT NULL,
  "label" jsonb NULL,
  "type" character varying NULL,
  "etag" character varying NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "created_by" character varying NULL,
  PRIMARY KEY ("id")
);
-- Create "essential_field_metadata" table
CREATE TABLE "essential_field_metadata" (
  "id" character varying NOT NULL,
  "model_metadata_id" character varying NULL,
  "name" jsonb NULL,
  "code" character varying NULL,
  "data_type" character varying NULL,
  "is_required" boolean NULL,
  "display_order" integer NULL,
  "is_archived" boolean NULL,
  "etag" character varying NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "created_by" character varying NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "essential_field_metadata_model_metadata_id_code_ukey" UNIQUE ("model_metadata_id", "code"),
  CONSTRAINT "essential_field_metadata_model_metadata_id_fkey" FOREIGN KEY ("model_metadata_id") REFERENCES "essential_model_metadata" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create "essential_currencies" table
CREATE TABLE "essential_currencies" (
  "id" character varying NOT NULL,
  "code" character varying NULL,
  "numeric_code" character varying NULL,
  "name" jsonb NULL,
  "symbol" character varying NULL,
  "symbol_position" character varying NULL,
  "currency_unit_label" jsonb NULL,
  "currency_subunit_label" jsonb NULL,
  "rounding" numeric NULL,
  "decimal_places" integer NULL,
  "is_archived" boolean NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "created_by" character varying NULL,
  "etag" character varying NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "essent_currencies_code_ukey" UNIQUE ("code")
);
-- Create "essential_org_currencies" table
CREATE TABLE "essential_org_currencies" (
  "id" character varying NOT NULL,
  "org_id" character varying NULL,
  "currency_id" character varying NULL,
  "is_active" boolean NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "created_by" character varying NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "essent_org_cur_org_currency_ukey" UNIQUE ("org_id", "currency_id"),
  CONSTRAINT "essential_org_currencies_currency_id_fkey" FOREIGN KEY ("currency_id") REFERENCES "essential_currencies" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create "essential_languages" table
CREATE TABLE "essential_languages" (
  "id" character varying NOT NULL,
  "name" character varying NULL,
  "locale_code" character varying NULL,
  "direction" character varying NULL,
  "decimal_separator" character varying NULL,
  "thousands_separator" character varying NULL,
  "grouping" character varying NULL,
  "date_format" character varying NULL,
  "time_format" character varying NULL,
  "short_time_format" character varying NULL,
  "first_day_of_week" character varying NULL,
  "public_use" boolean NULL,
  "is_archived" boolean NULL,
  "etag" character varying NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "created_by" character varying NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "essential_languages_locale_code_ukey" UNIQUE ("locale_code"),
  CONSTRAINT "essential_languages_name_ukey" UNIQUE ("name")
);
-- Create "essential_org_languages" table
CREATE TABLE "essential_org_languages" (
  "id" character varying NOT NULL,
  "org_id" character varying NULL,
  "language_id" character varying NULL,
  "is_active" boolean NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "created_by" character varying NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "essent_org_lang_org_language_ukey" UNIQUE ("org_id", "language_id"),
  CONSTRAINT "essential_org_languages_language_id_fkey" FOREIGN KEY ("language_id") REFERENCES "essential_languages" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create "essential_sequence_definitions" table
CREATE TABLE "essential_sequence_definitions" (
  "id" character varying NOT NULL,
  "configuration_scope" character varying NULL,
  "org_id" character varying NULL,
  "owner_module" character varying NULL,
  "code" character varying NULL,
  "name" character varying NULL,
  "description" character varying NULL,
  "sub_scope" character varying NULL,
  "scope_type" character varying NULL,
  "format_owner" character varying NULL,
  "algorithm" character varying NULL,
  "algorithm_params" jsonb NULL,
  "prefix_template" character varying NULL,
  "suffix_template" character varying NULL,
  "padding" integer NULL,
  "increment" bigint NULL,
  "initial_value" bigint NULL,
  "reset_period" character varying NULL,
  "timezone" character varying NULL,
  "issuance_retention_days" integer NULL,
  "first_issued_at" timestamptz NULL,
  "is_archived" boolean NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "created_by" character varying NULL,
  "etag" character varying NULL,
  PRIMARY KEY ("id")
);
-- Create index "essential_seqdef_tid_code_org_ukey_notnull" to table: "essential_sequence_definitions"
CREATE UNIQUE INDEX "essential_seqdef_tid_code_org_ukey_notnull" ON "essential_sequence_definitions" ("code", "org_id") WHERE (org_id IS NOT NULL);
-- Create index "essential_seqdef_tid_code_org_ukey_null" to table: "essential_sequence_definitions"
CREATE UNIQUE INDEX "essential_seqdef_tid_code_org_ukey_null" ON "essential_sequence_definitions" ("code") WHERE (org_id IS NULL);
-- Create "essential_sequence_counters" table
CREATE TABLE "essential_sequence_counters" (
  "id" character varying NOT NULL,
  "definition_id" character varying NULL,
  "org_id" character varying NULL,
  "scope_object_id" character varying NULL,
  "scope_key" character varying NULL,
  "period_key" character varying NULL,
  "last_number" bigint NULL,
  "algorithm_params" jsonb NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "created_by" character varying NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "essential_seqctr_tid_def_scope_period_ukey" UNIQUE ("definition_id", "scope_key", "period_key"),
  CONSTRAINT "essential_sequence_counters_definition_id_fkey" FOREIGN KEY ("definition_id") REFERENCES "essential_sequence_definitions" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT
);
-- Create "essential_sequence_issuances" table
CREATE TABLE "essential_sequence_issuances" (
  "id" character varying NOT NULL,
  "definition_id" character varying NULL,
  "counter_id" character varying NULL,
  "org_id" character varying NULL,
  "idempotency_id" uuid NULL,
  "operation_key" character varying NULL,
  "request_fingerprint" character varying NULL,
  "caller_module" character varying NULL,
  "caller_principal_id" character varying NULL,
  "accepted_at" timestamptz NULL,
  "effective_at" timestamptz NULL,
  "allocated_value" bigint NULL,
  "formatted_value" character varying NULL,
  "issued_at" timestamptz NULL,
  "expires_at" timestamptz NULL,
  "created_at" timestamptz NULL,
  "created_by" character varying NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "essential_seqiss_tid_counter_val_ukey" UNIQUE ("counter_id", "allocated_value"),
  CONSTRAINT "essential_seqiss_tid_def_idem_op_ukey" UNIQUE ("definition_id", "idempotency_id", "operation_key"),
  CONSTRAINT "essential_sequence_issuances_counter_id_fkey" FOREIGN KEY ("counter_id") REFERENCES "essential_sequence_counters" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT "essential_sequence_issuances_definition_id_fkey" FOREIGN KEY ("definition_id") REFERENCES "essential_sequence_definitions" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT
);
-- Create index "essential_seqiss_counter_issued_idx" to table: "essential_sequence_issuances"
CREATE INDEX "essential_seqiss_counter_issued_idx" ON "essential_sequence_issuances" ("counter_id", "issued_at", "id");
-- Create index "essential_seqiss_expires_at_idx" to table: "essential_sequence_issuances"
CREATE INDEX "essential_seqiss_expires_at_idx" ON "essential_sequence_issuances" ("expires_at");
-- Create "essential_uomcats" table
CREATE TABLE "essential_uomcats" (
  "id" character varying NOT NULL,
  "name" jsonb NULL,
  "reference_uom_id" character varying NULL,
  "org_id" character varying NULL,
  "is_archived" boolean NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "created_by" character varying NULL,
  "etag" character varying NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "essential_uomcats_name_org_id_ukey" UNIQUE ("name", "org_id")
);
-- Create "essential_uoms" table
CREATE TABLE "essential_uoms" (
  "id" character varying NOT NULL,
  "name" jsonb NULL,
  "symbol" character varying NULL,
  "category_id" character varying NULL,
  "uom_type" character varying NULL,
  "factor" numeric NULL,
  "rounding" numeric NULL,
  "org_id" character varying NULL,
  "is_archived" boolean NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "created_by" character varying NULL,
  "etag" character varying NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "essential_uoms_symbol_org_id_ukey" UNIQUE ("symbol", "org_id"),
  CONSTRAINT "essential_uoms_category_id_fkey" FOREIGN KEY ("category_id") REFERENCES "essential_uomcats" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
