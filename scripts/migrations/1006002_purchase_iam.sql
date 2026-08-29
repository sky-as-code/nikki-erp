-- IAM resources and actions for the whole Purchase module: configuration, sourcing group,
-- agreement and its lines, purchase order and its lines, and the audit trail. Every purchase
-- permission row lives here, in one file, so that a reviewer can see the module's entire permission
-- surface without opening several migrations.
--
-- The dynamic resource engine asserts permissions using the schema name as the resource code, so
-- these codes must stay byte-identical to the "purchase_order" / "purchase_agreement" / ... schema
-- names. A code that drifts from its schema denies every request, with nothing in the response
-- pointing at the seed as the cause. constants/resources.go aliases the model constants for the
-- same reason.
--
-- Lifecycle operations are seeded as SEPARATE actions rather than folded into "update", because
-- they are materially different powers. Confirming an order commits the business to a purchase and
-- cannot be undone by an edit; approving one is the control that spending policy rests on;
-- unlocking reopens a document that was deliberately closed. A role that may correct a typo in a
-- description should not thereby be able to do any of them.
--
-- Deliberate omissions, each of which would otherwise look like something forgotten:
--
--   * print, duplicate, create_alternative, compare_alternatives and create_rfq get NO action rows.
--     They reuse read or create, because they grant no power the caller does not already have:
--     printing renders a document the caller can already read, and duplicating or creating an
--     alternative is a create. Seeding them would let a role be granted "may print" while being
--     unable to read what it prints.
--   * purchase_audit_event gets read ONLY. The trail is written by the system inside the same
--     transaction as the transition it records; its engine refuses create, update and delete, so
--     seeding those verbs would advertise a capability the engine rejects.
--   * purchase_sourcing_group gets read ONLY, for the same reason: it is created by adding an
--     alternative to an order and reaped when fewer than two remain.
--   * purchase_order has no set_archived. An order is not archivable at all — its lifecycle is its
--     status, and an archived-but-open order would be a document withdrawn from view while still
--     committing the business to a purchase. The agreement, which IS archivable, does get one.
--   * The agreement's archive and restore operations use the built-in set_archived permission
--     rather than an "archive" action of their own; restore is the same power applied in reverse,
--     so splitting them would let a role archive agreements it could not bring back.
--
-- Deliberately NO iam_entitlements rows. Unit of Measure grants the system "User" role a
-- domain-wide read so that any user can pick a unit while filling in an unrelated form. Purchase
-- data is not comparable: an order carries what the business pays, to whom, on what terms, and the
-- audit trail carries who approved it. Access follows explicitly assigned roles, which is the same
-- choice Products and Contacts made. A blanket grant would expose every order, price and approval
-- in the system, and nothing in the test tree would notice.

DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_resources'
	) THEN
		INSERT INTO "iam_resources" (
			"id", "name", "code", "description", "owner_type", "max_scope", "min_scope", "created_at", "etag"
		) VALUES
		('01M2PVRCH00000000000000001', 'Purchase Configuration', 'purchase_configuration', 'Per-organization approval mode, threshold and modification policy', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M2PVRCH00000000000000002', 'Purchase Sourcing Group', 'purchase_sourcing_group', 'Technical grouping of purchase orders being compared as alternatives', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M2PVRCH00000000000000003', 'Purchase Agreement', 'purchase_agreement', 'Blanket orders and purchase templates agreed with a vendor', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M2PVRCH00000000000000004', 'Purchase Agreement Line', 'purchase_agreement_line', 'Committed product, quantity and price on a purchase agreement', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M2PVRCH00000000000000005', 'Purchase Order', 'purchase_order', 'Requests for quotation and the purchase orders they become', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M2PVRCH00000000000000006', 'Purchase Order Line', 'purchase_order_line', 'Product, quantity and price on a purchase order', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M2PVRCH00000000000000007', 'Purchase Audit Event', 'purchase_audit_event', 'Immutable record of every purchase order and agreement state change', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_actions'
	) THEN
		INSERT INTO "iam_actions" ("id", "name", "code", "description", "resource_id", "etag") VALUES
		-- Purchase Configuration
		('01M2PVRCH00000000000000008', 'Create', 'create', NULL, '01M2PVRCH00000000000000001', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M2PVRCH00000000000000009', 'Update', 'update', NULL, '01M2PVRCH00000000000000001', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M2PVRCH0000000000000000A', 'Delete', 'delete', NULL, '01M2PVRCH00000000000000001', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M2PVRCH0000000000000000B', 'Read', 'read', NULL, '01M2PVRCH00000000000000001', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Purchase Sourcing Group
		('01M2PVRCH0000000000000000C', 'Read', 'read', NULL, '01M2PVRCH00000000000000002', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Purchase Agreement
		('01M2PVRCH0000000000000000D', 'Create', 'create', NULL, '01M2PVRCH00000000000000003', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M2PVRCH0000000000000000E', 'Update', 'update', NULL, '01M2PVRCH00000000000000003', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M2PVRCH0000000000000000F', 'Delete', 'delete', NULL, '01M2PVRCH00000000000000003', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M2PVRCH0000000000000000G', 'Read', 'read', NULL, '01M2PVRCH00000000000000003', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M2PVRCH0000000000000000H', 'Set archived status', 'set_archived', 'Archive a purchase agreement so it is hidden from the working set', '01M2PVRCH00000000000000003', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M2PVRCH0000000000000000J', 'Confirm', 'confirm', 'Commit the document, which may route it for approval first', '01M2PVRCH00000000000000003', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M2PVRCH0000000000000000K', 'Close', 'close', 'Close a confirmed agreement to further orders', '01M2PVRCH00000000000000003', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M2PVRCH0000000000000000M', 'Cancel', 'cancel', 'Cancel the document, leaving it and its audit trail in place', '01M2PVRCH00000000000000003', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Purchase Agreement Line
		('01M2PVRCH0000000000000000N', 'Create', 'create', NULL, '01M2PVRCH00000000000000004', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M2PVRCH0000000000000000P', 'Update', 'update', NULL, '01M2PVRCH00000000000000004', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M2PVRCH0000000000000000Q', 'Delete', 'delete', NULL, '01M2PVRCH00000000000000004', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M2PVRCH0000000000000000R', 'Read', 'read', NULL, '01M2PVRCH00000000000000004', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Purchase Order
		('01M2PVRCH0000000000000000S', 'Create', 'create', NULL, '01M2PVRCH00000000000000005', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M2PVRCH0000000000000000T', 'Update', 'update', NULL, '01M2PVRCH00000000000000005', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M2PVRCH0000000000000000V', 'Delete', 'delete', NULL, '01M2PVRCH00000000000000005', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M2PVRCH0000000000000000W', 'Read', 'read', NULL, '01M2PVRCH00000000000000005', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M2PVRCH0000000000000000X', 'Send', 'send', 'Send the request for quotation to the vendor', '01M2PVRCH00000000000000005', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M2PVRCH0000000000000000Y', 'Confirm', 'confirm', 'Commit the document, which may route it for approval first', '01M2PVRCH00000000000000005', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M2PVRCH0000000000000000Z', 'Approve', 'approve', 'Approve a confirmed order that needed approval', '01M2PVRCH00000000000000005', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M2PVRCH00000000000000010', 'Lock', 'lock', 'Close an order to further editing', '01M2PVRCH00000000000000005', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M2PVRCH00000000000000011', 'Unlock', 'unlock', 'Reopen a locked order for editing, with a stated reason', '01M2PVRCH00000000000000005', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M2PVRCH00000000000000012', 'Acknowledge', 'acknowledge', 'Record that the vendor confirmed receipt of the order', '01M2PVRCH00000000000000005', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M2PVRCH00000000000000013', 'Cancel', 'cancel', 'Cancel the document, leaving it and its audit trail in place', '01M2PVRCH00000000000000005', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M2PVRCH00000000000000014', 'Merge', 'merge', 'Merge several draft orders for the same vendor into one', '01M2PVRCH00000000000000005', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M2PVRCH00000000000000020', 'Reprice', 'reprice', 'Re-resolve a draft order''s line prices from the vendor''s current price list', '01M2PVRCH00000000000000005', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Purchase Order Line
		('01M2PVRCH00000000000000015', 'Create', 'create', NULL, '01M2PVRCH00000000000000006', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M2PVRCH00000000000000016', 'Update', 'update', NULL, '01M2PVRCH00000000000000006', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M2PVRCH00000000000000017', 'Delete', 'delete', NULL, '01M2PVRCH00000000000000006', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M2PVRCH00000000000000018', 'Read', 'read', NULL, '01M2PVRCH00000000000000006', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Purchase Audit Event
		('01M2PVRCH00000000000000019', 'Read', 'read', NULL, '01M2PVRCH00000000000000007', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;


-- The vendor price resource and its authorization, added by the product-pricing change request.
--
-- The resource had no IAM row at all: it is new, and the dynamic resource engine derives the code
-- it asserts against from the schema name, so `purchase_vendor_product_price` must appear here
-- byte-identically or every request to it is denied with nothing in the 403 pointing at this file.
--
-- Two roles, for the same reason the sales pricing roles are two: reading what a supplier charges
-- and recording it are different powers. A buyer comparing quotes needs the first; only whoever
-- maintains supplier terms needs the second.
--
-- The resource has no custom action, so the five built-in codes are its whole surface. `delete` is
-- included in the manager role because a vendor price has no dependents -- a purchase order line
-- records the price it resolved rather than a reference that would dangle -- so a row created by
-- mistake should be removable, while `set_archived` is the ordinary way to retire a real one.

DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_resources'
	) THEN
		INSERT INTO "iam_resources" (
			"id", "name", "code", "description", "owner_type", "max_scope",
			"min_scope", "created_at", "etag"
		) VALUES
		('01M2PVRCH00000000000000100', 'Vendor Product Price', 'purchase_vendor_product_price', 'What a vendor currently offers a product at, by quantity, unit and validity', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;

DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_actions'
	) THEN
		INSERT INTO "iam_actions" ("id", "name", "code", "description", "resource_id", "etag") VALUES
		('01M2PVRCH00000000000000101', 'Create', 'create', NULL, '01M2PVRCH00000000000000100', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M2PVRCH00000000000000102', 'Update', 'update', NULL, '01M2PVRCH00000000000000100', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M2PVRCH00000000000000103', 'Delete', 'delete', NULL, '01M2PVRCH00000000000000100', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M2PVRCH00000000000000104', 'Read', 'read', NULL, '01M2PVRCH00000000000000100', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M2PVRCH00000000000000105', 'Set archived status', 'set_archived', 'Retire a vendor price from new resolution while keeping it readable for existing orders', '01M2PVRCH00000000000000100', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;

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
		('01M3ROLE0000000000PRICING3', 'Purchase Vendor Price Readonly', 'Read what vendors are offering, without being able to change a recorded quote', false, '01JWNMZ36QHC7CQQ748H9NQ6J6', true, false, true, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ROLE0000000000PRICING4', 'Purchase Vendor Price Manager', 'Record and maintain vendor price quotes', false, '01JWNMZ36QHC7CQQ748H9NQ6J6', true, false, true, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
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
		('01M3ENT00000000000VNDPR001', 'Read vendor price quotes', 'Read vendor price quotes', 'read:purchase_vendor_product_price:org', '01M2PVRCH00000000000000104', '01M2PVRCH00000000000000100', '01M3ROLE0000000000PRICING3', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT00000000000VNDPR002', 'Create vendor price quotes', 'Create vendor price quotes', 'create:purchase_vendor_product_price:org', '01M2PVRCH00000000000000101', '01M2PVRCH00000000000000100', '01M3ROLE0000000000PRICING4', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT00000000000VNDPR003', 'Update vendor price quotes', 'Update vendor price quotes', 'update:purchase_vendor_product_price:org', '01M2PVRCH00000000000000102', '01M2PVRCH00000000000000100', '01M3ROLE0000000000PRICING4', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT00000000000VNDPR004', 'Delete vendor price quotes', 'Delete vendor price quotes', 'delete:purchase_vendor_product_price:org', '01M2PVRCH00000000000000103', '01M2PVRCH00000000000000100', '01M3ROLE0000000000PRICING4', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT00000000000VNDPR005', 'Read vendor price quotes', 'Read vendor price quotes', 'read:purchase_vendor_product_price:org', '01M2PVRCH00000000000000104', '01M2PVRCH00000000000000100', '01M3ROLE0000000000PRICING4', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT00000000000VNDPR006', 'Set archived vendor price quotes', 'Set archived vendor price quotes', 'set_archived:purchase_vendor_product_price:org', '01M2PVRCH00000000000000105', '01M2PVRCH00000000000000100', '01M3ROLE0000000000PRICING4', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;
