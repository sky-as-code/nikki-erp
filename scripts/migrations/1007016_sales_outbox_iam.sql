-- IAM for the integration event outbox (SALES-037).
--
-- `read` alone. Rows are written by the domain services that produce the events, inside their own
-- transactions, and drained by the background sweep.
--
-- The omissions are the point. A client able to POST one could announce a sale that never happened
-- to every downstream consumer - stock released, money reconciled, invoices raised, all against
-- nothing. A client able to PATCH one could set published_at on an event that never went, which
-- deletes it from the queue as surely as a DELETE would, and silently: the row still looks correct.
--
-- It is readable at all because an operator investigating a consumer that has fallen behind needs to
-- see what Sales believes it published, when, and how many times it has tried.

DO $$
BEGIN
	IF EXISTS (SELECT FROM information_schema.tables WHERE table_schema='public' AND table_name='iam_resources') THEN
		INSERT INTO "iam_resources" ("id","name","code","description","owner_type","max_scope","min_scope","created_at","etag") VALUES
		('01M3SALES00000000000000060', 'Sales Integration Outbox', 'sales_integration_outbox', 'Integration events Sales has published, or is waiting to publish', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (SELECT FROM information_schema.tables WHERE table_schema='public' AND table_name='iam_actions') THEN
		INSERT INTO "iam_actions" ("id","name","code","description","resource_id","etag") VALUES
		('01M3SALES00000000000000061', 'Read', 'read', NULL, '01M3SALES00000000000000060', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;
