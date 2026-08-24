-- IAM resources and actions for the Sales module: the sales channel and the sales point.
-- Every Sales permission row lives here, in one file, so that a reviewer can see the module's
-- entire permission surface without opening several migrations.
--
-- The dynamic resource engine asserts permissions using the schema name as the resource code, so
-- these codes must stay byte-identical to the "sales_channel" / "sales_point" schema names. A code
-- that drifts from its schema denies every request, with nothing in the response pointing at the
-- seed as the cause.
--
-- Lifecycle operations are seeded as SEPARATE actions rather than folded into "update", because
-- they are materially different powers. Suspending a channel stops every sales point under it from
-- selling; archiving one retires it; unarchiving resurrects a point that was deliberately
-- decommissioned. A role that may correct a display name should not thereby be able to do any of
-- them.
--
-- Deliberate omissions, each of which would otherwise look like something forgotten:
--
--   * There is NO "register" action on sales_channel. RegisterSalesChannel is the idempotent
--     module-integration entry point, and it is create-or-return-existing: a caller that may not
--     create a channel must not be able to reach it, and one that may needs no second permission.
--     It rides on "create".
--   * There is NO "resolve" action. Resolving a channel code to its id is a read of one row by a
--     different key, and seeding it would let a role be granted "may resolve" while unable to read
--     what it resolved.
--   * sales_point has no "register" either, for the same reason as the channel.
--   * Archive and unarchive both ride on the built-in set_archived permission rather than taking
--     actions of their own. Unarchiving is the same power applied in reverse, so splitting them
--     would let a role archive sales points it could not bring back.
--
-- Deliberately NO iam_entitlements rows. Unit of Measure grants the system "User" role a
-- domain-wide read so that any user can pick a unit while filling in an unrelated form. Sales
-- configuration is not comparable: a channel names which selling surfaces exist and which payment
-- methods they accept, and a sales point identifies a physical machine or store. Access follows
-- explicitly assigned roles, which is the same choice Products, Contacts and Purchase made.

DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_resources'
	) THEN
		INSERT INTO "iam_resources" (
			"id", "name", "code", "description", "owner_type", "max_scope", "min_scope", "created_at", "etag"
		) VALUES
		('01M3SALES00000000000000001', 'Sales Channel', 'sales_channel', 'Classification of where a sale happens, and the payment methods it accepts', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000002', 'Sales Point', 'sales_point', 'A concrete selling place within a channel: a kiosk, a store, a storefront', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_actions'
	) THEN
		INSERT INTO "iam_actions" ("id", "name", "code", "description", "resource_id", "etag") VALUES
		-- Sales Channel
		('01M3SALES00000000000000003', 'Create', 'create', NULL, '01M3SALES00000000000000001', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000004', 'Update', 'update', NULL, '01M3SALES00000000000000001', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000005', 'Delete', 'delete', NULL, '01M3SALES00000000000000001', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000006', 'Read', 'read', NULL, '01M3SALES00000000000000001', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000007', 'Set archived status', 'set_archived', 'Archive a sales channel, or bring an archived one back', '01M3SALES00000000000000001', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000008', 'Suspend', 'suspend', 'Stop a channel taking new business, leaving history, returns and refunds working', '01M3SALES00000000000000001', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000009', 'Activate', 'activate', 'Return a suspended channel to service', '01M3SALES00000000000000001', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000000A', 'Enable payment method', 'enable_payment_method', 'Allow a payment method to be used on this channel', '01M3SALES00000000000000001', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000000B', 'Disable payment method', 'disable_payment_method', 'Stop new payments using a method on this channel, leaving historical ones untouched', '01M3SALES00000000000000001', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Sales Point
		('01M3SALES0000000000000000C', 'Create', 'create', NULL, '01M3SALES00000000000000002', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000000D', 'Update', 'update', NULL, '01M3SALES00000000000000002', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000000E', 'Delete', 'delete', NULL, '01M3SALES00000000000000002', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000000F', 'Read', 'read', NULL, '01M3SALES00000000000000002', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000000G', 'Set archived status', 'set_archived', 'Archive a sales point, or bring an archived one back', '01M3SALES00000000000000002', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000000H', 'Suspend', 'suspend', 'Stop a sales point taking new orders, leaving returns and refunds working', '01M3SALES00000000000000002', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000000J', 'Activate', 'activate', 'Return a suspended sales point to service', '01M3SALES00000000000000002', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;
