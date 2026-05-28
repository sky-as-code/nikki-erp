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
