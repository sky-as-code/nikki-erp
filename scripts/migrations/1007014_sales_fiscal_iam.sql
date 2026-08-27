-- IAM for fiscal document requests (SALES-030, SALES-031).
--
-- `read` and `create`, and nothing else. There is no `update` and no `delete`, and that omission is
-- the point: a client able to PATCH one of these rows could mark an unissued request as issued -
-- telling a customer they hold a VAT invoice that does not exist, which is precisely the state
-- BR 77 exists to keep honest - and a client able to DELETE one could erase the record of a legal
-- document that the provider still holds.
--
-- `create` is the permission the request_invoice action demands, and it hangs off THIS resource
-- rather than off sales_bill. Asking for a VAT invoice is not a power over the settlement: whoever
-- serves a customer who paid at a till and afterwards wants an invoice for their company need not
-- also be able to split, merge or settle the bill they are invoicing.

DO $$
BEGIN
	IF EXISTS (SELECT FROM information_schema.tables WHERE table_schema='public' AND table_name='iam_resources') THEN
		INSERT INTO "iam_resources" ("id","name","code","description","owner_type","max_scope","min_scope","created_at","etag") VALUES
		('01M3SALES00000000000000057', 'Sales Fiscal Request', 'sales_fiscal_request', 'What Sales asked an eInvoice provider for, and what came back', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (SELECT FROM information_schema.tables WHERE table_schema='public' AND table_name='iam_actions') THEN
		INSERT INTO "iam_actions" ("id","name","code","description","resource_id","etag") VALUES
		('01M3SALES00000000000000058', 'Read', 'read', NULL, '01M3SALES00000000000000057', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000059', 'Create', 'create', 'Ask an eInvoice provider for a legal fiscal document', '01M3SALES00000000000000057', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;
