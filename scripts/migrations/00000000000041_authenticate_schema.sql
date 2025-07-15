-- Create "authn_attempts" table
CREATE TABLE "authn_attempts" ("id" character varying NOT NULL, "created_at" timestamptz NOT NULL, "methods" jsonb NOT NULL, "current_method" character varying NULL, "device_ip" character varying NULL, "device_name" character varying NULL, "device_location" character varying NULL, "expired_at" timestamptz NOT NULL, "is_genuine" boolean NOT NULL, "subject_type" character varying NOT NULL, "subject_ref" character varying NOT NULL, "subject_source_ref" character varying NULL, "status" character varying NOT NULL, "updated_at" timestamptz NULL, PRIMARY KEY ("id"));
-- Create index "loginattempt_subject_type_subject_ref" to table: "authn_attempts"
CREATE INDEX "loginattempt_subject_type_subject_ref" ON "authn_attempts" ("subject_type", "subject_ref");
-- Create "authn_method_settings" table
CREATE TABLE "authn_method_settings" ("id" character varying NOT NULL, "method" character varying NOT NULL, "order" bigint NOT NULL, "max_failures" bigint NOT NULL, "lock_duration_secs" bigint NULL, "subject_type" character varying NOT NULL, "subject_ref" character varying NULL, "subject_source_ref" character varying NULL, PRIMARY KEY ("id"));
-- Create index "methodsetting_subject_type_subject_ref" to table: "authn_method_settings"
CREATE INDEX "methodsetting_subject_type_subject_ref" ON "authn_method_settings" ("subject_type", "subject_ref");
-- Create "authn_password_stores" table
CREATE TABLE "authn_password_stores" ("id" character varying NOT NULL, "password" character varying NOT NULL, "password_expired_at" timestamptz NULL, "password_updated_at" timestamptz NOT NULL, "passwordtmp" character varying NULL, "passwordtmp_expired_at" timestamptz NULL, "passwordotp" character varying NULL, "passwordotp_expired_at" timestamptz NULL, "subject_type" character varying NOT NULL, "subject_ref" character varying NOT NULL, "subject_source_ref" character varying NULL, PRIMARY KEY ("id"));
-- Create index "passwordstore_subject_type_subject_ref" to table: "authn_password_stores"
CREATE INDEX "passwordstore_subject_type_subject_ref" ON "authn_password_stores" ("subject_type", "subject_ref");
