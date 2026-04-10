-- Create "authn_attempts" table
CREATE TABLE "authn_attempts" (
  "id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "methods" character varying[] NOT NULL,
  "current_method" character varying NOT NULL,
  "device_ip" character varying NULL,
  "device_name" character varying NULL,
  "device_location" character varying NULL,
  "expires_at" timestamptz NOT NULL,
  "principal_type" character varying NOT NULL,
  "status" character varying NOT NULL,
  "username" character varying NOT NULL,
  PRIMARY KEY ("id")
);
-- Create "authn_method_settings" table
CREATE TABLE "authn_method_settings" (
  "id" character varying NOT NULL,
  "method" character varying NOT NULL,
  "order" integer NOT NULL,
  "max_failures" integer NOT NULL,
  "lock_duration_secs" integer NULL,
  "subject_type" character varying NOT NULL,
  "subject_ref" character varying NULL,
  "subject_source_ref" character varying NULL,
  PRIMARY KEY ("id")
);
-- Create "authn_password_stores" table
CREATE TABLE "authn_password_stores" (
  "id" character varying NOT NULL,
  "principal_type" character varying NOT NULL,
  "principal_id" character varying NOT NULL,
  "password" character varying NULL,
  "password_expires_at" timestamptz NULL,
  "password_updated_at" timestamptz NULL,
  "passwordtmp" character varying NULL,
  "passwordtmp_expires_at" timestamptz NULL,
  "passwordotp" character varying NULL,
  "passwordotp_expires_at" timestamptz NULL,
  "passwordotp_recovery" character varying[] NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "authn_password_stores_principal_type_principal_id_ukey" UNIQUE ("principal_type", "principal_id")
);
