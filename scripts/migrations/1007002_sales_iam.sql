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
