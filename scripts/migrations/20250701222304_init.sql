-- Modify "ident_organizations" table
ALTER TABLE "ident_organizations" ADD COLUMN "address" character varying NULL, ADD COLUMN "legal_name" character varying NULL, ADD COLUMN "phone_number" character varying NULL;
