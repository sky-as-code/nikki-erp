-- IAM for the fulfillment method catalogue.
--
-- Five actions, the standard CRUD set plus set_archived, which is the permission BOTH the archive
-- and unarchive routes answer to: withdrawing a policy from new orders and restoring it are the same
-- power in reverse, and splitting them would let a role retire a method it could not put back.
--
-- There is no suspend action here, unlike sales_channel and sales_point. A method is either offered
-- to new orders or it is not, and the fulfillments that already snapshotted it keep running either
-- way, so a second "temporarily off" state would say nothing is_archived does not already say.
--
-- sales_channel_fulfillment_methods gets no resource row at all: the mapping is configured through
-- the channel that owns it and nothing routes to it, exactly as sales_channel_payment_rel is
-- handled. A resource row for it would advertise a permission over an endpoint that does not exist.

DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_resources'
	) THEN
		INSERT INTO "iam_resources" (
			"id", "name", "code", "description", "owner_type", "max_scope", "min_scope", "created_at", "etag"
		) VALUES
		('01M3SALES0000000000000009E', 'Sales Fulfillment Method', 'sales_fulfillment_method', 'The policy a sale is handed over under: how the goods reach the customer, and what happens when that fails', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_actions'
	) THEN
		INSERT INTO "iam_actions" ("id", "name", "code", "description", "resource_id", "etag") VALUES
		('01M3SALES0000000000000009F', 'Create', 'create', NULL, '01M3SALES0000000000000009E', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000009G', 'Update', 'update', NULL, '01M3SALES0000000000000009E', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000009H', 'Delete', 'delete', NULL, '01M3SALES0000000000000009E', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000009J', 'Read', 'read', NULL, '01M3SALES0000000000000009E', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000009K', 'Set archived status', 'set_archived', 'Withdraw a fulfillment method from new orders, or bring a withdrawn one back; running fulfillments are unaffected either way', '01M3SALES0000000000000009E', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;
