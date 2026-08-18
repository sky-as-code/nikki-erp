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
-- Create index "contacts_parties_tid_tax_id_ukey_notnull" to table: "contacts_parties"
CREATE UNIQUE INDEX "contacts_parties_tid_tax_id_ukey_notnull" ON "contacts_parties" ("org_id", "tax_id") WHERE (tax_id IS NOT NULL);
-- Create index "contacts_parties_tid_tax_id_ukey_null" to table: "contacts_parties"
CREATE UNIQUE INDEX "contacts_parties_tid_tax_id_ukey_null" ON "contacts_parties" ("org_id") WHERE (tax_id IS NULL);
-- Create index "contacts_parties_tid_website_ukey_notnull" to table: "contacts_parties"
CREATE UNIQUE INDEX "contacts_parties_tid_website_ukey_notnull" ON "contacts_parties" ("org_id", "website") WHERE (website IS NOT NULL);
-- Create index "contacts_parties_tid_website_ukey_null" to table: "contacts_parties"
CREATE UNIQUE INDEX "contacts_parties_tid_website_ukey_null" ON "contacts_parties" ("org_id") WHERE (website IS NULL);
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
