-- Create "contacts_parties" table
CREATE TABLE "contacts_parties" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "avatar_url" character varying NULL,
  "display_name" character varying NOT NULL,
  "legal_name" character varying NULL,
  "legal_address" character varying NULL,
  "tax_id" character varying NULL,
  "job_position" character varying NULL,
  "title" character varying NULL,
  "type" character varying NOT NULL,
  "note" character varying NULL,
  "nationality_id" character varying NULL,
  "language_id" character varying NULL,
  "website" character varying NULL,
  "is_archived" boolean NOT NULL,
  "etag" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "contacts_parties_tid_org_id_idx" to table: "contacts_parties"
CREATE INDEX "contacts_parties_tid_org_id_idx" ON "contacts_parties" ("org_id");
-- Create index "contacts_parties_tid_tax_id_idx" to table: "contacts_parties"
CREATE INDEX "contacts_parties_tid_tax_id_idx" ON "contacts_parties" ("org_id", "tax_id");
-- Create index "contacts_parties_tid_website_idx" to table: "contacts_parties"
CREATE INDEX "contacts_parties_tid_website_idx" ON "contacts_parties" ("org_id", "website");
-- Create "contacts_comm_channels" table
CREATE TABLE "contacts_comm_channels" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "note" character varying NULL,
  "party_id" character varying NOT NULL,
  "type" character varying NOT NULL,
  "value" character varying NULL,
  "value_json" jsonb NULL,
  "is_archived" boolean NOT NULL,
  "etag" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "contacts_comm_channels_party_id_fkey" FOREIGN KEY ("party_id") REFERENCES "contacts_parties" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "contacts_comm_chans_tid_org_id_idx" to table: "contacts_comm_channels"
CREATE INDEX "contacts_comm_chans_tid_org_id_idx" ON "contacts_comm_channels" ("org_id");
-- Create index "contacts_comm_chans_tid_party_id_idx" to table: "contacts_comm_channels"
CREATE INDEX "contacts_comm_chans_tid_party_id_idx" ON "contacts_comm_channels" ("party_id");
-- Create "contacts_relationships" table
CREATE TABLE "contacts_relationships" (
  "id" character varying NOT NULL,
  "party_id" character varying NOT NULL,
  "target_party_id" character varying NOT NULL,
  "type" character varying NOT NULL,
  "note" character varying NULL,
  "is_archived" boolean NOT NULL,
  "etag" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "contacts_relationships_party_id_fkey" FOREIGN KEY ("party_id") REFERENCES "contacts_parties" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "contacts_relationships_target_party_id_fkey" FOREIGN KEY ("target_party_id") REFERENCES "contacts_parties" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "contacts_rels_tid_party_id_idx" to table: "contacts_relationships"
CREATE INDEX "contacts_rels_tid_party_id_idx" ON "contacts_relationships" ("party_id");
-- Create index "contacts_rels_tid_target_party_id_idx" to table: "contacts_relationships"
CREATE INDEX "contacts_rels_tid_target_party_id_idx" ON "contacts_relationships" ("target_party_id");
-- Create "contacts_vendor_profiles" table
CREATE TABLE "contacts_vendor_profiles" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "party_id" character varying NOT NULL,
  "status" character varying NOT NULL,
  "status_reason" character varying NULL,
  "default_currency_id" character varying NULL,
  "payment_terms" character varying NULL,
  "lead_time_days" integer NULL,
  "note" character varying NULL,
  "is_archived" boolean NOT NULL,
  "etag" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "contacts_vnd_profiles_tid_party_id_ukey" UNIQUE ("party_id", "org_id"),
  CONSTRAINT "contacts_vendor_profiles_party_id_fkey" FOREIGN KEY ("party_id") REFERENCES "contacts_parties" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "contacts_vnd_profiles_tid_org_id_idx" to table: "contacts_vendor_profiles"
CREATE INDEX "contacts_vnd_profiles_tid_org_id_idx" ON "contacts_vendor_profiles" ("org_id");
-- Create index "contacts_vnd_profiles_tid_status_idx" to table: "contacts_vendor_profiles"
CREATE INDEX "contacts_vnd_profiles_tid_status_idx" ON "contacts_vendor_profiles" ("status");
