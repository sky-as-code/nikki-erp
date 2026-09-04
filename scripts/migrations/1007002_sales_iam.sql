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
		('01M3SALES00000000000000002', 'Sales Point', 'sales_point', 'A concrete selling place within a channel: a kiosk, a store, a storefront', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000000K', 'Sales Order', 'sales_order', 'One commercial transaction: what was sold, to whom, through which channel', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000000M', 'Sales Order Line', 'sales_order_line', 'One thing sold on one order, with its quantities and its price at the moment of sale', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000000Z', 'Sales Order Line Component', 'sales_order_line_component', 'One real product inside a combo line, which is what Inventory actually fulfils', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000010', 'Sales Order Adjustment', 'sales_order_adjustment', 'One step of the pricing calculation, kept so a price can be explained and replayed', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000011', 'Sales Order Event', 'sales_order_event', 'One thing that happened to a sale: the document audit trail', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000015', 'Sales Pricelist', 'sales_pricelist', 'A set of prices that applies to a channel or a selling place for a period', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000016', 'Sales Pricelist Item', 'sales_pricelist_item', 'One price, for one product variant, in one unit, from one quantity upward', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000017', 'Sales Combo', 'sales_combo', 'A bundle sold at a price of its own, independent of what its parts cost apart', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000018', 'Sales Combo Component', 'sales_combo_component', 'One product inside a bundle definition', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000034', 'Sales Promotion Program', 'sales_promotion_program', 'One campaign: what has to be true, and what the customer then gets', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000035', 'Sales Promotion Condition Group', 'sales_promotion_condition_group', 'A set of conditions ANDed together; groups are ORed with each other', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000036', 'Sales Promotion Condition', 'sales_promotion_condition', 'One test a promotion applies to decide whether it is eligible', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000037', 'Sales Promotion Condition Target', 'sales_promotion_condition_target', 'One member of a set-valued condition, such as an eligible product variant', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000038', 'Sales Promotion Reward', 'sales_promotion_reward', 'One thing a promotion gives: a discount, a fixed price, a free item', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000039', 'Sales Promotion Compatibility', 'sales_promotion_compatibility', 'Whether two promotions may apply to the same order', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
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
		('01M3SALES0000000000000000J', 'Activate', 'activate', 'Return a suspended sales point to service', '01M3SALES00000000000000002', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Sales Order
		('01M3SALES0000000000000000N', 'Create', 'create', NULL, '01M3SALES0000000000000000K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000000P', 'Update', 'update', NULL, '01M3SALES0000000000000000K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000000Q', 'Delete', 'delete', NULL, '01M3SALES0000000000000000K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000000R', 'Read', 'read', NULL, '01M3SALES0000000000000000K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000000S', 'Set archived status', 'set_archived', 'Archive a sales order, or bring an archived one back', '01M3SALES0000000000000000K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Sales Order Line
		('01M3SALES0000000000000000T', 'Create', 'create', NULL, '01M3SALES0000000000000000M', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000000V', 'Update', 'update', NULL, '01M3SALES0000000000000000M', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000000W', 'Delete', 'delete', NULL, '01M3SALES0000000000000000M', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000000X', 'Read', 'read', NULL, '01M3SALES0000000000000000M', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000000Y', 'Set archived status', 'set_archived', 'Archive a sales order line, or bring an archived one back', '01M3SALES0000000000000000M', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Sales Order Line Component
		('01M3SALES00000000000000012', 'Read', 'read', NULL, '01M3SALES0000000000000000Z', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Sales Order Adjustment
		('01M3SALES00000000000000013', 'Read', 'read', NULL, '01M3SALES00000000000000010', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Sales Order Event
		('01M3SALES00000000000000014', 'Read', 'read', NULL, '01M3SALES00000000000000011', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Sales Pricelist
		('01M3SALES00000000000000019', 'Create', 'create', NULL, '01M3SALES00000000000000015', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000001A', 'Update', 'update', NULL, '01M3SALES00000000000000015', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000001B', 'Delete', 'delete', NULL, '01M3SALES00000000000000015', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000001C', 'Read', 'read', NULL, '01M3SALES00000000000000015', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000001D', 'Set archived status', 'set_archived', 'Archive a pricelist, or bring an archived one back', '01M3SALES00000000000000015', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Sales Pricelist Item
		('01M3SALES0000000000000001E', 'Create', 'create', NULL, '01M3SALES00000000000000016', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000001F', 'Update', 'update', NULL, '01M3SALES00000000000000016', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000001G', 'Delete', 'delete', NULL, '01M3SALES00000000000000016', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000001H', 'Read', 'read', NULL, '01M3SALES00000000000000016', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000001J', 'Set archived status', 'set_archived', 'Archive a pricelist item, or bring an archived one back', '01M3SALES00000000000000016', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Sales Combo
		('01M3SALES0000000000000001K', 'Create', 'create', NULL, '01M3SALES00000000000000017', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000001M', 'Update', 'update', NULL, '01M3SALES00000000000000017', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000001N', 'Delete', 'delete', NULL, '01M3SALES00000000000000017', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000001P', 'Read', 'read', NULL, '01M3SALES00000000000000017', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000001Q', 'Set archived status', 'set_archived', 'Archive a combo, or bring an archived one back', '01M3SALES00000000000000017', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Sales Combo Component
		('01M3SALES0000000000000001R', 'Create', 'create', NULL, '01M3SALES00000000000000018', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000001S', 'Update', 'update', NULL, '01M3SALES00000000000000018', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000001T', 'Delete', 'delete', NULL, '01M3SALES00000000000000018', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000001V', 'Read', 'read', NULL, '01M3SALES00000000000000018', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000001W', 'Set archived status', 'set_archived', 'Archive a combo component, or bring an archived one back', '01M3SALES00000000000000018', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Sales Promotion Program
		('01M3SALES0000000000000003A', 'Create', 'create', NULL, '01M3SALES00000000000000034', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000003B', 'Update', 'update', NULL, '01M3SALES00000000000000034', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000003C', 'Delete', 'delete', NULL, '01M3SALES00000000000000034', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000003D', 'Read', 'read', NULL, '01M3SALES00000000000000034', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000003E', 'Set archived status', 'set_archived', 'Archive a sales promotion program, or bring an archived one back', '01M3SALES00000000000000034', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Sales Promotion Condition Group
		('01M3SALES0000000000000003F', 'Create', 'create', NULL, '01M3SALES00000000000000035', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000003G', 'Update', 'update', NULL, '01M3SALES00000000000000035', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000003H', 'Delete', 'delete', NULL, '01M3SALES00000000000000035', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000003J', 'Read', 'read', NULL, '01M3SALES00000000000000035', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000003K', 'Set archived status', 'set_archived', 'Archive a sales promotion condition group, or bring an archived one back', '01M3SALES00000000000000035', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Sales Promotion Condition
		('01M3SALES0000000000000003M', 'Create', 'create', NULL, '01M3SALES00000000000000036', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000003N', 'Update', 'update', NULL, '01M3SALES00000000000000036', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000003P', 'Delete', 'delete', NULL, '01M3SALES00000000000000036', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000003Q', 'Read', 'read', NULL, '01M3SALES00000000000000036', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000003R', 'Set archived status', 'set_archived', 'Archive a sales promotion condition, or bring an archived one back', '01M3SALES00000000000000036', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Sales Promotion Condition Target
		('01M3SALES0000000000000003S', 'Create', 'create', NULL, '01M3SALES00000000000000037', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000003T', 'Update', 'update', NULL, '01M3SALES00000000000000037', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000003V', 'Delete', 'delete', NULL, '01M3SALES00000000000000037', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000003W', 'Read', 'read', NULL, '01M3SALES00000000000000037', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000003X', 'Set archived status', 'set_archived', 'Archive a sales promotion condition target, or bring an archived one back', '01M3SALES00000000000000037', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Sales Promotion Reward
		('01M3SALES0000000000000003Y', 'Create', 'create', NULL, '01M3SALES00000000000000038', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000003Z', 'Update', 'update', NULL, '01M3SALES00000000000000038', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000040', 'Delete', 'delete', NULL, '01M3SALES00000000000000038', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000041', 'Read', 'read', NULL, '01M3SALES00000000000000038', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000042', 'Set archived status', 'set_archived', 'Archive a sales promotion reward, or bring an archived one back', '01M3SALES00000000000000038', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Sales Promotion Compatibility
		('01M3SALES00000000000000043', 'Create', 'create', NULL, '01M3SALES00000000000000039', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000044', 'Update', 'update', NULL, '01M3SALES00000000000000039', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000045', 'Delete', 'delete', NULL, '01M3SALES00000000000000039', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000046', 'Read', 'read', NULL, '01M3SALES00000000000000039', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000047', 'Set archived status', 'set_archived', 'Archive a sales promotion compatibility, or bring an archived one back', '01M3SALES00000000000000039', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;


-- Pricing authorization, added by the product-pricing change request.
--
-- The resources and their CRUD actions were already seeded above; what was missing was roles, and
-- the entitlements binding them. Two roles rather than one, because reading what something sells
-- for and deciding it are different powers: a shop floor or a report needs the first and must not
-- have the second.
--
-- The reader holds `read` on BOTH the list and its rules. A rule is meaningless without the list it
-- belongs to -- the currency and the scope live there -- so granting one without the other would
-- produce a reader that could see prices it could not interpret.
--
-- There is deliberately NO separate permission for choosing the organization default. That
-- operation asserts `update` (see PermissionSetDefaultPricelist): choosing which list prices an
-- order that named none is the same class of power as editing what a list charges, and a role that
-- may do one may sensibly do the other.

DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_roles'
	) THEN
		INSERT INTO "iam_roles" (
			"id", "name", "description", "is_private", "owner_user_id", "is_requestable",
			"is_required_attachment", "is_required_comment", "is_archived", "created_at", "etag"
		) VALUES
		('01M3ROLE0000000000PRICING1', 'Sales Pricing Readonly', 'Read sales price lists and their rules, without being able to change what anything sells for', false, '01JWNMZ36QHC7CQQ748H9NQ6J6', true, false, true, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ROLE0000000000PRICING2', 'Sales Pricing Manager', 'Create and edit sales price lists and rules, including choosing the organization default', false, '01JWNMZ36QHC7CQQ748H9NQ6J6', true, false, true, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;

DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_entitlements'
	) THEN
		INSERT INTO "iam_entitlements" (
			"id", "name", "description", "expression", "action_id", "resource_id",
			"role_id", "scope", "org_id", "org_unit_id", "is_archived", "created_at", "etag"
		) VALUES
		('01M3ENT00000000000PRICE001', 'Read price lists', 'Read price lists', 'read:sales_pricelist:org', '01M3SALES0000000000000001C', '01M3SALES00000000000000015', '01M3ROLE0000000000PRICING1', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT00000000000PRICE002', 'Read the rules inside a price list', 'Read the rules inside a price list', 'read:sales_pricelist_item:org', '01M3SALES0000000000000001H', '01M3SALES00000000000000016', '01M3ROLE0000000000PRICING1', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT00000000000PRICE003', 'Create price lists', 'Create price lists', 'create:sales_pricelist:org', '01M3SALES00000000000000019', '01M3SALES00000000000000015', '01M3ROLE0000000000PRICING2', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT00000000000PRICE004', 'Create pricing rules', 'Create pricing rules', 'create:sales_pricelist_item:org', '01M3SALES0000000000000001E', '01M3SALES00000000000000016', '01M3ROLE0000000000PRICING2', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT00000000000PRICE005', 'Update price lists', 'Update price lists', 'update:sales_pricelist:org', '01M3SALES0000000000000001A', '01M3SALES00000000000000015', '01M3ROLE0000000000PRICING2', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT00000000000PRICE006', 'Update pricing rules', 'Update pricing rules', 'update:sales_pricelist_item:org', '01M3SALES0000000000000001F', '01M3SALES00000000000000016', '01M3ROLE0000000000PRICING2', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT00000000000PRICE007', 'Delete price lists', 'Delete price lists', 'delete:sales_pricelist:org', '01M3SALES0000000000000001B', '01M3SALES00000000000000015', '01M3ROLE0000000000PRICING2', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT00000000000PRICE008', 'Delete pricing rules', 'Delete pricing rules', 'delete:sales_pricelist_item:org', '01M3SALES0000000000000001G', '01M3SALES00000000000000016', '01M3ROLE0000000000PRICING2', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT00000000000PRICE009', 'Read price lists', 'Read price lists', 'read:sales_pricelist:org', '01M3SALES0000000000000001C', '01M3SALES00000000000000015', '01M3ROLE0000000000PRICING2', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT00000000000PRICE010', 'Read pricing rules', 'Read pricing rules', 'read:sales_pricelist_item:org', '01M3SALES0000000000000001H', '01M3SALES00000000000000016', '01M3ROLE0000000000PRICING2', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT00000000000PRICE011', 'Set archived price lists', 'Set archived price lists', 'set_archived:sales_pricelist:org', '01M3SALES0000000000000001D', '01M3SALES00000000000000015', '01M3ROLE0000000000PRICING2', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT00000000000PRICE012', 'Set archived pricing rules', 'Set archived pricing rules', 'set_archived:sales_pricelist_item:org', '01M3SALES0000000000000001J', '01M3SALES00000000000000016', '01M3ROLE0000000000PRICING2', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;


-- ============================================================================
-- Merged from 1007006_sales_voucher_iam.sql
-- ============================================================================
-- IAM resources and actions for the voucher pair (SALES-022).
--
-- Two resources with deliberately different permission surfaces.
--
-- A voucher CODE is operator-managed master data and gets the full five actions: somebody runs
-- campaigns, and running one means creating codes, correcting a validity window, withdrawing a
-- mispriced offer and reading the result.
--
-- A voucher REDEMPTION gets `read` alone. It is a ledger the system writes as orders move, and every
-- non-read power over it is a way to corrupt what it records: a client able to create one could forge
-- a discount's provenance, one able to update one could release a hold another order is relying on,
-- and one able to delete one would break the usage count that is derived from these rows. The same
-- reasoning already applies to sales_order_events, sales_order_adjustments and
-- sales_order_line_components.
--
-- The codes must stay byte-identical to the schema names. A code that drifts denies every request to
-- that resource, with nothing in the 403 pointing at this file as the cause.

DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_resources'
	) THEN
		INSERT INTO "iam_resources" (
			"id", "name", "code", "description", "owner_type", "max_scope", "min_scope", "created_at", "etag"
		) VALUES
		('01M3SALES00000000000000048', 'Sales Voucher Code', 'sales_voucher_code', 'A credential a customer presents to activate a promotion program', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000049', 'Sales Voucher Redemption', 'sales_voucher_redemption', 'The ledger of which orders consumed which voucher codes', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_actions'
	) THEN
		INSERT INTO "iam_actions" ("id", "name", "code", "description", "resource_id", "etag") VALUES
		-- Sales Voucher Code: full CRUD, an operator manages campaigns.
		('01M3SALES0000000000000004A', 'Create', 'create', NULL, '01M3SALES00000000000000048', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000004B', 'Update', 'update', NULL, '01M3SALES00000000000000048', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000004C', 'Delete', 'delete', NULL, '01M3SALES00000000000000048', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000004D', 'Read', 'read', NULL, '01M3SALES00000000000000048', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000004E', 'Set archived status', 'set_archived', 'Withdraw a voucher code, or bring a withdrawn one back', '01M3SALES00000000000000048', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Sales Voucher Redemption: read alone. See the note at the top of this file.
		('01M3SALES0000000000000004F', 'Read', 'read', NULL, '01M3SALES00000000000000049', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_actions'
	) THEN
		INSERT INTO "iam_actions" ("id", "name", "code", "description", "resource_id", "etag") VALUES
		-- Applying a voucher is a power over the ORDER, not over the voucher code.
		--
		-- It reserves a use and changes what the basket costs; the code itself is untouched master
		-- data. Seating it on sales_order is what lets a till take a discount at the counter without
		-- also holding write permission over campaign configuration.
		('01M3SALES0000000000000004G', 'Apply voucher', 'apply_voucher', 'Apply a voucher code to a draft order, reserving one of its uses', '01M3SALES0000000000000000K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- The three parties to a sale are three separate powers, because a business splits them.
		--
		-- Recording who bought is counter work; redirecting who gets invoiced decides who owes the
		-- money, and is the one a finance team keeps to itself. Granting them together would mean
		-- anyone who may serve a customer may also bill a different company for the sale.
		('01M3SALES0000000000000009A', 'Assign parties', 'assign_parties', 'Set any combination of the sold-to, bill-to and payer parties on a sale in one call', '01M3SALES0000000000000000K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000009B', 'Assign sold-to party', 'assign_sold_to_party', 'Name the business partner buying on a sale', '01M3SALES0000000000000000K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000009C', 'Assign bill-to party', 'assign_bill_to_party', 'Name the business partner the invoice is made out to; refused once its billing instruction is being issued', '01M3SALES0000000000000000K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000009D', 'Assign payer party', 'assign_payer_party', 'Name the business partner responsible for paying a sale', '01M3SALES0000000000000000K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_actions'
	) THEN
		INSERT INTO "iam_actions" ("id", "name", "code", "description", "resource_id", "etag") VALUES
		-- Confirm and cancel are separate powers from update, deliberately.
		--
		-- Confirming commits the business to a price and redeems vouchers; cancelling unwinds a sale
		-- and hands voucher uses back. A role that may correct a line should not thereby be able to
		-- do either, which is what folding them into `update` would grant.
		('01M3SALES0000000000000004H', 'Confirm', 'confirm', 'Commit a draft sales order: freeze its prices and redeem its vouchers', '01M3SALES0000000000000000K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000004J', 'Cancel', 'cancel', 'Cancel a sales order that has not been paid or fulfilled', '01M3SALES0000000000000000K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;


-- ============================================================================
-- Merged from 1007008_sales_bill_iam.sql
-- ============================================================================
-- IAM resources and actions for billing (SALES-024).
--
-- Three resources with deliberately different surfaces.
--
-- A BILL is operator-managed and gets full CRUD: somebody opens, corrects and cancels settlement
-- units. Its ALLOCATIONS get read alone - they are computed by split, merge and the initial bill,
-- and a client able to write one could make an order's bills stop summing to the order (BR 36),
-- which is the one invariant the whole billing model rests on. The LINEAGE is read alone for the
-- same reason as the audit trail: a client able to POST one could fabricate a paper trail showing a
-- payment settled a bill it never touched.

DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_resources'
	) THEN
		INSERT INTO "iam_resources" (
			"id", "name", "code", "description", "owner_type", "max_scope", "min_scope", "created_at", "etag"
		) VALUES
		('01M3SALES0000000000000004K', 'Sales Bill', 'sales_bill', 'A settlement unit of a sale - never a VAT invoice (BR 33)', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000004M', 'Sales Bill Line', 'sales_bill_line', 'One order line''s allocated share of one bill', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000004N', 'Sales Bill Relation', 'sales_bill_relation', 'The lineage left behind by a bill split or merge', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_actions'
	) THEN
		INSERT INTO "iam_actions" ("id", "name", "code", "description", "resource_id", "etag") VALUES
		-- Sales Bill: full CRUD, an operator manages settlement.
		('01M3SALES0000000000000004P', 'Create', 'create', NULL, '01M3SALES0000000000000004K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000004Q', 'Update', 'update', NULL, '01M3SALES0000000000000004K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000004R', 'Delete', 'delete', NULL, '01M3SALES0000000000000004K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000004S', 'Read', 'read', NULL, '01M3SALES0000000000000004K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000004T', 'Set archived status', 'set_archived', 'Archive a sales bill, or bring an archived one back', '01M3SALES0000000000000004K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Allocations and lineage: read alone. See the note at the top of this file.
		('01M3SALES0000000000000004V', 'Read', 'read', NULL, '01M3SALES0000000000000004M', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000004W', 'Read', 'read', NULL, '01M3SALES0000000000000004N', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_actions'
	) THEN
		INSERT INTO "iam_actions" ("id", "name", "code", "description", "resource_id", "etag") VALUES
		-- Split and merge are separate powers from update: they restructure settlement units and
		-- write lineage, and a role that may correct a bill's details need not be able to divide it.
		('01M3SALES0000000000000004X', 'Split', 'split', 'Divide one open bill into several', '01M3SALES0000000000000004K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000004Y', 'Merge', 'merge', 'Combine several open bills of one order into a single bill', '01M3SALES0000000000000004K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;


-- ============================================================================
-- Merged from 1007010_sales_payment_iam.sql
-- ============================================================================
-- IAM for payments (SALES-027/028).
--
-- The resource gets `read` alone. Money is recorded through the record_payment operation, which
-- applies six gates a plain POST would bypass - including the two that ask paymentinvoice whether
-- the method may be used at all (CR 33). A writable payment resource would let a client record money
-- that no gateway ever took.
--
-- `pay` and `settle` hang off the BILL, because that is what they change.

DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_resources'
	) THEN
		INSERT INTO "iam_resources" (
			"id", "name", "code", "description", "owner_type", "max_scope", "min_scope", "created_at", "etag"
		) VALUES
		('01M3SALES0000000000000004Z', 'Sales Payment', 'sales_payment', 'One movement of money against a bill', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_actions'
	) THEN
		INSERT INTO "iam_actions" ("id", "name", "code", "description", "resource_id", "etag") VALUES
		('01M3SALES00000000000000050', 'Read', 'read', NULL, '01M3SALES0000000000000004Z', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000051', 'Record payment', 'pay', 'Take a payment against an open bill', '01M3SALES0000000000000004K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000052', 'Settle', 'settle', 'Close a bill whose money is fully in', '01M3SALES0000000000000004K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;


-- ============================================================================
-- Merged from 1007012_sales_fulfillment_iam.sql
-- ============================================================================
-- IAM for fulfilment requests (SALES-029).
--
-- Both resources get `read` alone. Requests are raised by confirm and by the return workflow, and a
-- client able to write one could tell Inventory to move goods that no sale asked for - which is the
-- one thing BR 44's separation exists to prevent.

DO $$
BEGIN
	IF EXISTS (SELECT FROM information_schema.tables WHERE table_schema='public' AND table_name='iam_resources') THEN
		INSERT INTO "iam_resources" ("id","name","code","description","owner_type","max_scope","min_scope","created_at","etag") VALUES
		('01M3SALES00000000000000053', 'Sales Fulfilment Request', 'sales_fulfillment_request', 'One thing Sales asked Inventory to do for a sale', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000054', 'Sales Fulfilment Request Line', 'sales_fulfillment_request_line', 'One order line covered by a fulfilment request', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (SELECT FROM information_schema.tables WHERE table_schema='public' AND table_name='iam_actions') THEN
		INSERT INTO "iam_actions" ("id","name","code","description","resource_id","etag") VALUES
		('01M3SALES00000000000000055', 'Read', 'read', NULL, '01M3SALES00000000000000053', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000056', 'Read', 'read', NULL, '01M3SALES00000000000000054', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;


-- ============================================================================
-- Merged from 1007014_sales_fiscal_iam.sql
-- ============================================================================
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


-- ============================================================================
-- Merged from 1007016_sales_outbox_iam.sql
-- ============================================================================
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


-- ============================================================================
-- Merged from 1007018_sales_quotation_iam.sql
-- ============================================================================
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


-- ============================================================================
-- Merged from 1007020_sales_manual_discount_iam.sql
-- ============================================================================
-- IAM for manual price overrides (SALES-039).
--
-- Two things are seeded here, and they sit on different resources.
--
-- `manual_discount` on SALES_ORDER is the power BR 87.4 gates: changing what a customer pays for
-- reasons outside the price list. It is deliberately NOT `update` — a role that may correct a
-- quantity should not thereby be able to discount a sale — and revoking takes the same permission as
-- granting, since withdrawing a discount changes the price just as granting one does.
--
-- `read` on SALES_MANUAL_DISCOUNT is the whole surface of the record itself. No create, no update,
-- no delete: rows are written by the grant operation and by nothing else, because a plain POST would
-- bypass every gate that makes an override auditable — the mandatory reason, the draft-only check,
-- and the audit entry recording both the old and the new price.

DO $$
BEGIN
	IF EXISTS (SELECT FROM information_schema.tables WHERE table_schema='public' AND table_name='iam_resources') THEN
		INSERT INTO "iam_resources" ("id","name","code","description","owner_type","max_scope","min_scope","created_at","etag") VALUES
		('01M3SALES00000000000000073', 'Sales Manual Discount', 'sales_manual_discount', 'An operator override of a sale price, with the reason it was granted', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (SELECT FROM information_schema.tables WHERE table_schema='public' AND table_name='iam_actions') THEN
		INSERT INTO "iam_actions" ("id","name","code","description","resource_id","etag") VALUES
		('01M3SALES00000000000000074', 'Read', 'read', NULL, '01M3SALES00000000000000073', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;

		-- The power itself, on the ORDER, since that is what gets discounted. The resource id is
		-- written literally rather than looked up by code: the seed-coverage test parses these
		-- files without a database, so a SELECT-based insert would be invisible to it and an
		-- action could go unseeded without anything failing.
		INSERT INTO "iam_actions" ("id","name","code","description","resource_id","etag") VALUES
		('01M3SALES00000000000000075', 'Manual discount', 'manual_discount', 'Change what a customer pays, for a stated reason outside the price list', '01M3SALES0000000000000000K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;


-- ============================================================================
-- Merged from 1007023_sales_return_iam.sql
-- ============================================================================
-- IAM for returns, return lines and refund legs.
--
-- `process_return` is its own action rather than `update` because it moves money out of the
-- business: a role that may correct a return's reason must not thereby be able to release its
-- refund.
--
-- sales_return_line and sales_refund_payment are read-only to clients. Lines are written by
-- create_return, which caps each quantity at what is still returnable and prices it from the
-- historical amounts; refund legs are written by the return workflow, which caps each at what its
-- original payment captured. A writable row either way is an outflow with no matching inflow.

DO $$
BEGIN
	IF EXISTS (SELECT FROM information_schema.tables WHERE table_schema='public' AND table_name='iam_resources') THEN
		INSERT INTO "iam_resources" ("id","name","code","description","owner_type","max_scope","min_scope","created_at","etag") VALUES
		('01M3SALES00000000000000076', 'Sales Return', 'sales_return', 'A customer sending goods back, and the refund that settles it', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000080', 'Sales Return Line', 'sales_return_line', 'One order line coming back, in whole or in part', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000082', 'Sales Refund Payment', 'sales_refund_payment', 'One leg of a refund, against one original payment', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (SELECT FROM information_schema.tables WHERE table_schema='public' AND table_name='iam_actions') THEN
		-- The return document: readable, creatable, correctable, and archivable.
		INSERT INTO "iam_actions" ("id","name","code","description","resource_id","etag") VALUES
		('01M3SALES00000000000000077', 'Read', 'read', NULL, '01M3SALES00000000000000076', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000078', 'Create', 'create', 'Record that a customer wants to send something back', '01M3SALES00000000000000076', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000079', 'Update', 'update', 'Correct a return, or cancel one before anything has moved', '01M3SALES00000000000000076', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),

		-- The power that moves money and goods. Separate from `update` on purpose: this is the one
		-- that cannot be undone once it starts.
		('01M3SALES0000000000000007A', 'Process return', 'process_return', 'Send the goods back, refund the customer and adjust the VAT invoice', '01M3SALES00000000000000076', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),

		-- Read only; see the header.
		('01M3SALES00000000000000081', 'Read', 'read', NULL, '01M3SALES00000000000000080', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000083', 'Read', 'read', NULL, '01M3SALES00000000000000082', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;


-- ============================================================================
-- Merged from 1007027_sales_billing_instruction_iam.sql
-- ============================================================================
-- IAM for billing instructions and their issuance attempts.
--
-- `mark_ready` is its own action rather than `update`, because the two are materially different
-- powers. Correcting a tax code is clerical; marking ready is what releases a legal document to be
-- issued in a company's name, and a role that may fix a typo must not thereby be able to bill
-- someone.
--
-- `cancel` is likewise separate: withdrawing a buyer's request to be invoiced is not an edit of it.
--
-- sales_billing_issuance_attempt is READ-ONLY to clients. Attempts are written only by the issuance
-- job, and they are the evidence of whether a document was created — including the indeterminate
-- case where nobody knows. A client able to write one could fabricate a record showing an invoice
-- was issued, or erase the trace of one that may exist.

DO $$
BEGIN
	IF EXISTS (SELECT FROM information_schema.tables WHERE table_schema='public' AND table_name='iam_resources') THEN
		INSERT INTO "iam_resources" ("id","name","code","description","owner_type","max_scope","min_scope","created_at","etag") VALUES
		('01M3SALES00000000000000090', 'Sales Billing Instruction', 'sales_billing_instruction', 'Who a sale is to be invoiced to, under which legal identity', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000091', 'Sales Billing Issuance Attempt', 'sales_billing_issuance_attempt', 'One try at issuing an electronic invoice, kept as evidence', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (SELECT FROM information_schema.tables WHERE table_schema='public' AND table_name='iam_actions') THEN
		INSERT INTO "iam_actions" ("id","name","code","description","resource_id","etag") VALUES
		-- The instruction: readable, creatable, correctable while it is still a draft.
		('01M3SALES00000000000000092', 'Read', 'read', NULL, '01M3SALES00000000000000090', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000093', 'Create', 'create', 'Record that a buyer wants to be invoiced, and under which identity', '01M3SALES00000000000000090', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000094', 'Update', 'update', 'Correct the fiscal details before the document is issued', '01M3SALES00000000000000090', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),

		-- The power that releases a document to be issued. See the note at the top of this file.
		('01M3SALES00000000000000095', 'Mark ready', 'mark_ready', 'Confirm the fiscal details and release the instruction for issuance', '01M3SALES00000000000000090', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000096', 'Revert to draft', 'revert_to_draft', 'Take a released instruction back for further correction', '01M3SALES00000000000000090', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000097', 'Cancel', 'cancel', 'Withdraw a buyer''s request to be invoiced, before any document exists', '01M3SALES00000000000000090', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000098', 'Refresh from bill-to', 'refresh_from_bill_to', 'Re-copy the fiscal details from the business partner record', '01M3SALES00000000000000090', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),

		-- Attempts: read alone. See the note at the top of this file.
		('01M3SALES00000000000000099', 'Read', 'read', NULL, '01M3SALES00000000000000091', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;
