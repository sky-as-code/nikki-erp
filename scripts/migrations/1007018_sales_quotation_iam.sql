-- IAM for quotations (SALES-038).
--
-- Full CRUD on the quotation, unlike the order: a quotation is a document an operator writes, and
-- creating or correcting one commits the business to nothing and moves no money. What is gated is
-- the STATUS, which is declared no_update and moves only through the actions.
--
-- `convert` is its own permission, not folded into `update`. Accepting a quotation CREATES A SALES
-- ORDER - it commits the business to a sale - and a role that may draft and correct an offer should
-- not thereby be able to turn one into a binding order. Sending and cancelling ride on `update`,
-- because both are ordinary handling of a document by whoever owns it.
--
-- The lines get CRUD too: editing what was offered is editing the quotation, and splitting the
-- permission would mean an operator who could write the header could not fill it in.

DO $$
BEGIN
	IF EXISTS (SELECT FROM information_schema.tables WHERE table_schema='public' AND table_name='iam_resources') THEN
		INSERT INTO "iam_resources" ("id","name","code","description","owner_type","max_scope","min_scope","created_at","etag") VALUES
		('01M3SALES00000000000000062', 'Sales Quotation', 'sales_quotation', 'An offer made to a customer, which may become a sales order', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000063', 'Sales Quotation Line', 'sales_quotation_line', 'One line of an offer', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (SELECT FROM information_schema.tables WHERE table_schema='public' AND table_name='iam_actions') THEN
		INSERT INTO "iam_actions" ("id","name","code","description","resource_id","etag") VALUES
		('01M3SALES00000000000000064', 'Create', 'create', NULL, '01M3SALES00000000000000062', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000065', 'Read', 'read', NULL, '01M3SALES00000000000000062', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000066', 'Update', 'update', 'Edit a quotation, send it, or withdraw it', '01M3SALES00000000000000062', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000067', 'Delete', 'delete', NULL, '01M3SALES00000000000000062', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000068', 'Convert', 'convert', 'Accept a quotation and create the sales order it becomes', '01M3SALES00000000000000062', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000069', 'Create', 'create', NULL, '01M3SALES00000000000000063', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000070', 'Read', 'read', NULL, '01M3SALES00000000000000063', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000071', 'Update', 'update', NULL, '01M3SALES00000000000000063', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000072', 'Delete', 'delete', NULL, '01M3SALES00000000000000063', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;
