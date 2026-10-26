-- Seed data for the Contacts module: the parties the business trades with, and the vendor profiles
-- that make some of them suppliers.
--
-- Added for the product-pricing change request, which needs vendors to hang vendor prices off.
-- Before this the module had no seed data at all, so purchase pricing had nothing to reference.
--
-- The set is chosen to cover every vendor STATE that downstream code branches on, not merely to
-- provide a list of names. A vendor price validator asks Contacts whether a party may be ordered
-- from, and that question has four different answers here:
--
--   * active    -- the ordinary case, eleven of them
--   * proposed  -- exists but not yet approved, so an order must be refused
--   * suspended -- temporarily out, and distinct from archived: the profile is still current
--   * blacklisted -- permanently out, with the reason recorded
--
-- plus one archived profile, which is a different axis again: archival hides the record, while
-- status describes the commercial relationship. A consumer that collapses the two gets one wrong.
--
-- One party belongs to the second organization, so org scoping has something to exclude, and one
-- is an individual rather than a company, so the type column is not uniformly one value.
--
-- Every insert is ON CONFLICT DO NOTHING and every id is fixed rather than generated, so running
-- this file twice changes nothing.


DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'contacts_parties'
	) THEN
		INSERT INTO "contacts_parties" (
			"id", "org_id", "display_name", "legal_name", "tax_id", "type",
			"website", "note",
			"is_archived", "created_at", "etag"
		) VALUES
		('01K5CNT00000000000VENDOR0001', '01JWNY20G23KD4RV5VWYABQYHD', 'Saigon Beverage Co.', 'Saigon Beverage Company Limited', '0301234567', 'company', 'https://vendor01.example.com', 'Primary soft drinks and water supplier', FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5CNT00000000000VENDOR0002', '01JWNY20G23KD4RV5VWYABQYHD', 'Mekong Foods JSC', 'Mekong Foods Joint Stock Company', '0302345678', 'company', 'https://vendor02.example.com', 'Snacks and confectionery', FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5CNT00000000000VENDOR0003', '01JWNY20G23KD4RV5VWYABQYHD', 'Highland Coffee Traders', 'Highland Coffee Trading Co., Ltd', '0303456789', 'company', 'https://vendor03.example.com', 'Green and roasted coffee, sold by weight', FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5CNT00000000000VENDOR0004', '01JWNY20G23KD4RV5VWYABQYHD', 'Dalat Dairy', 'Dalat Dairy Corporation', '0304567890', 'company', 'https://vendor04.example.com', 'Fresh milk and chilled dairy', FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5CNT00000000000VENDOR0005', '01JWNY20G23KD4RV5VWYABQYHD', 'PackRight Industries', 'PackRight Industries Co., Ltd', '0305678901', 'company', 'https://vendor05.example.com', 'Cartons and packaging materials', FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5CNT00000000000VENDOR0006', '01JWNY20G23KD4RV5VWYABQYHD', 'Shenzhen Cable Works', 'Shenzhen Cable Works Ltd', '9100000006', 'company', 'https://vendor06.example.com', 'Imported components, quotes in USD', FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5CNT00000000000VENDOR0007', '01JWNY20G23KD4RV5VWYABQYHD', 'Quick Logistics', 'Quick Logistics Services JSC', '0306789012', 'company', 'https://vendor07.example.com', 'Third-party delivery, fastest lead time', FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5CNT00000000000VENDOR0008', '01JWNY20G23KD4RV5VWYABQYHD', 'Budget Wholesale', 'Budget Wholesale Trading Co.', '0307890123', 'company', 'https://vendor08.example.com', 'Cheapest bulk prices, longest lead time', FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5CNT00000000000VENDOR0009', '01JWNY20G23KD4RV5VWYABQYHD', 'Northern Supply Partners', 'Northern Supply Partners Ltd', '0308901234', 'company', 'https://vendor09.example.com', 'Secondary source for beverages', FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5CNT00000000000VENDOR0010', '01JWNY20G23KD4RV5VWYABQYHD', 'Retired Supplier Co.', 'Retired Supplier Company Limited', '0309012345', 'company', 'https://vendor10.example.com', 'No longer traded with; archived', TRUE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5CNT00000000000VENDOR0011', '01JWNY20G23KD4RV5VWYABQYHD', 'Blocked Trading Co.', 'Blocked Trading Company', '0310123456', 'company', 'https://vendor11.example.com', 'Blacklisted after a quality incident', FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5CNT00000000000VENDOR0012', '01JWNY20G23KD4RV5VWYABQYHD', 'Pending Newco', 'Pending Newco Limited', '0313456789', 'company', 'https://vendor12.example.com', 'Proposed, not yet approved for ordering', FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5CNT00000000000VENDOR0013', '01K1H7M2K9VW3P5R7XQJY2C1Z9', 'Tech Solutions Supplier', 'Tech Solutions Supplier Ltd', '0311234567', 'company', 'https://vendor13.example.com', 'Supplier of the second organization', FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5CNT00000000000VENDOR0014', '01JWNY20G23KD4RV5VWYABQYHD', 'Mr Tran Van Long', NULL, NULL, 'individual', 'https://vendor14.example.com', 'Sole trader, occasional produce', FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5CNT00000000000VENDOR0015', '01JWNY20G23KD4RV5VWYABQYHD', 'Suspended Foods Ltd', 'Suspended Foods Limited', '0312345678', 'company', 'https://vendor15.example.com', 'Suspended pending contract renewal', FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;

DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'contacts_vendor_profiles'
	) THEN
		INSERT INTO "contacts_vendor_profiles" (
			"id", "org_id", "party_id", "status", "status_reason",
			"default_currency_id", "payment_terms", "lead_time_days", "note",
			"is_archived", "created_at", "etag"
		) VALUES
		('01K5CNT00000000000VNDPROF001', '01JWNY20G23KD4RV5VWYABQYHD', '01K5CNT00000000000VENDOR0001', 'active', NULL, '01KZQC0000CURRENCY000ND300', 'NET30', 3, NULL, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5CNT00000000000VNDPROF002', '01JWNY20G23KD4RV5VWYABQYHD', '01K5CNT00000000000VENDOR0002', 'active', NULL, '01KZQC0000CURRENCY000ND300', 'NET30', 5, NULL, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5CNT00000000000VNDPROF003', '01JWNY20G23KD4RV5VWYABQYHD', '01K5CNT00000000000VENDOR0003', 'active', NULL, '01KZQC0000CURRENCY000ND300', 'NET15', 7, NULL, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5CNT00000000000VNDPROF004', '01JWNY20G23KD4RV5VWYABQYHD', '01K5CNT00000000000VENDOR0004', 'active', NULL, '01KZQC0000CURRENCY000ND300', 'NET07', 1, NULL, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5CNT00000000000VNDPROF005', '01JWNY20G23KD4RV5VWYABQYHD', '01K5CNT00000000000VENDOR0005', 'active', NULL, '01KZQC0000CURRENCY000ND300', 'NET45', 10, NULL, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5CNT00000000000VNDPROF006', '01JWNY20G23KD4RV5VWYABQYHD', '01K5CNT00000000000VENDOR0006', 'active', NULL, '01KZQC0000CURRENCY000MJ300', 'NET60', 30, NULL, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5CNT00000000000VNDPROF007', '01JWNY20G23KD4RV5VWYABQYHD', '01K5CNT00000000000VENDOR0007', 'active', NULL, '01KZQC0000CURRENCY000ND300', 'NET15', 1, NULL, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5CNT00000000000VNDPROF008', '01JWNY20G23KD4RV5VWYABQYHD', '01K5CNT00000000000VENDOR0008', 'active', NULL, '01KZQC0000CURRENCY000ND300', 'NET60', 21, NULL, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5CNT00000000000VNDPROF009', '01JWNY20G23KD4RV5VWYABQYHD', '01K5CNT00000000000VENDOR0009', 'active', NULL, '01KZQC0000CURRENCY000ND300', 'NET30', 4, NULL, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5CNT00000000000VNDPROF010', '01JWNY20G23KD4RV5VWYABQYHD', '01K5CNT00000000000VENDOR0010', 'active', NULL, '01KZQC0000CURRENCY000ND300', 'NET30', 14, NULL, TRUE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5CNT00000000000VNDPROF011', '01JWNY20G23KD4RV5VWYABQYHD', '01K5CNT00000000000VENDOR0011', 'blacklisted', 'Repeated quality failures in 2025', '01KZQC0000CURRENCY000ND300', 'NET30', 14, NULL, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5CNT00000000000VNDPROF012', '01JWNY20G23KD4RV5VWYABQYHD', '01K5CNT00000000000VENDOR0012', 'proposed', 'Awaiting credit check', '01KZQC0000CURRENCY000ND300', NULL, 14, NULL, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5CNT00000000000VNDPROF013', '01K1H7M2K9VW3P5R7XQJY2C1Z9', '01K5CNT00000000000VENDOR0013', 'active', NULL, '01KZQC0000CURRENCY000ND300', 'NET30', 5, NULL, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5CNT00000000000VNDPROF014', '01JWNY20G23KD4RV5VWYABQYHD', '01K5CNT00000000000VENDOR0014', 'active', NULL, '01KZQC0000CURRENCY000ND300', 'COD', 2, NULL, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5CNT00000000000VNDPROF015', '01JWNY20G23KD4RV5VWYABQYHD', '01K5CNT00000000000VENDOR0015', 'suspended', 'Contract under renegotiation', '01KZQC0000CURRENCY000ND300', 'NET30', 9, NULL, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;
