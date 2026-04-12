-- Create "contacts_parties" table
CREATE TABLE "contacts_parties" (
  "id" character varying NOT NULL,
  "avatar_url" character varying NULL,
  "created_at" timestamptz NOT NULL,
  "deleted_at" timestamptz NULL,
  "deleted_by" character varying NULL,
  "display_name" character varying NOT NULL,
  "etag" character varying NOT NULL,
  "job_position" character varying NULL,
  "language_id" character varying NULL,
  "legal_address" character varying NULL,
  "legal_name" character varying NULL,
  "nationality_id" character varying NULL,
  "note" character varying NULL,
  "org_id" character varying NULL,
  "tax_id" character varying NULL,
  "title" character varying NULL,
  "type" character varying NOT NULL,
  "updated_at" timestamptz NULL,
  "website" character varying NULL,
  PRIMARY KEY ("id")
);
-- Create "contacts_comm_channels" table
CREATE TABLE "contacts_comm_channels" (
  "id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "deleted_by" character varying NULL,
  "deleted_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  "note" character varying NULL,
  "org_id" character varying NOT NULL,
  "type" character varying NOT NULL,
  "updated_at" timestamptz NULL,
  "value" character varying NULL,
  "value_json" jsonb NULL,
  "party_id" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "contacts_comm_channels_contacts_parties_party" FOREIGN KEY ("party_id") REFERENCES "contacts_parties" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create "contacts_relationships" table
CREATE TABLE "contacts_relationships" (
  "id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "deleted_at" timestamptz NULL,
  "deleted_by" character varying NULL,
  "etag" character varying NOT NULL,
  "note" character varying NULL,
  "type" character varying NOT NULL,
  "updated_at" timestamptz NULL,
  "party_id" character varying NOT NULL,
  "target_party_id" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "contacts_relationships_contacts_parties_source_party" FOREIGN KEY ("party_id") REFERENCES "contacts_parties" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "contacts_relationships_contacts_parties_target_party" FOREIGN KEY ("target_party_id") REFERENCES "contacts_parties" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
