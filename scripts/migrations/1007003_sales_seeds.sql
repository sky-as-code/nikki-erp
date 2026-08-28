-- Seed data for the Sales module: sales channels and the sales points under them.
--
-- Two of the channels are load-bearing rather than sample data, and are marked is_system so the
-- lifecycle operations refuse to archive or rename them:
--
--   * "vdmc" is the vending machine channel. Its code is a published integration contract: the
--     vending module names it by this string and lets Sales resolve the id, so that no database id
--     appears in a kiosk's configuration. Renaming it breaks every kiosk already registered.
--   * "bo" is the back office channel, and its single sales point is where an order with no
--     physical selling place is recorded. It exists so that sales_point_id can be NOT NULL rather
--     than carrying "unknown" as a null.
--
-- The remainder is demonstration data: regional variants of the commercial channels a retailer
-- would configure, and the stores, kiosks and counters under them.
--
-- Every insert is ON CONFLICT DO NOTHING and every id is fixed rather than generated, so running
-- this file twice changes nothing. That matters for "vdmc" in particular, whose seed the change
-- request requires to be idempotent.
--
-- Only the vending points carry an external_reference_id, because only they were registered by
-- another module on the retailer's behalf; the pair (channel, reference id) is what makes a
-- retried registration resolve to the point it already created. Points a human created carry none,
-- and there are many of them per channel — which is why those uniqueness rules are declared
-- partial_uniques_strict: the loose variant would additionally forbid a second reference-less
-- point per channel.


DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'sales_channels'
	) THEN
		INSERT INTO "sales_channels" (
			"id", "org_id", "code", "name", "description", "managed_by_module", "status", "is_system", "is_archived", "created_at", "updated_at", "etag"
		) VALUES
		('01M3SC00000000000000000001', '01JWNY20G23KD4RV5VWYABQYHD', 'vdmc', 'Vending Machine', 'Unattended kiosks selling directly to walk-up customers', 'vending_machine_new', 'active', true, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC00000000000000000002', '01JWNY20G23KD4RV5VWYABQYHD', 'bo', 'Back Office', 'Sales recorded by staff on behalf of a customer', NULL, 'active', true, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC00000000000000000003', '01JWNY20G23KD4RV5VWYABQYHD', 'pos01', 'Point of Sale - Ha Noi', 'Staffed checkout lanes in a physical store', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC00000000000000000004', '01JWNY20G23KD4RV5VWYABQYHD', 'pos02', 'Point of Sale - Ho Chi Minh', 'Staffed checkout lanes in a physical store', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC00000000000000000005', '01JWNY20G23KD4RV5VWYABQYHD', 'pos03', 'Point of Sale - Da Nang', 'Staffed checkout lanes in a physical store', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC00000000000000000006', '01JWNY20G23KD4RV5VWYABQYHD', 'pos04', 'Point of Sale - Hai Phong', 'Staffed checkout lanes in a physical store', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC00000000000000000007', '01JWNY20G23KD4RV5VWYABQYHD', 'pos05', 'Point of Sale - Can Tho', 'Staffed checkout lanes in a physical store', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC00000000000000000008', '01JWNY20G23KD4RV5VWYABQYHD', 'pos06', 'Point of Sale - Bien Hoa', 'Staffed checkout lanes in a physical store', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC00000000000000000009', '01JWNY20G23KD4RV5VWYABQYHD', 'pos07', 'Point of Sale - Nha Trang', 'Staffed checkout lanes in a physical store', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000000A', '01JWNY20G23KD4RV5VWYABQYHD', 'pos08', 'Point of Sale - Hue', 'Staffed checkout lanes in a physical store', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000000B', '01JWNY20G23KD4RV5VWYABQYHD', 'pos09', 'Point of Sale - Vung Tau', 'Staffed checkout lanes in a physical store', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000000C', '01JWNY20G23KD4RV5VWYABQYHD', 'pos10', 'Point of Sale - Buon Ma Thuot', 'Staffed checkout lanes in a physical store', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000000D', '01JWNY20G23KD4RV5VWYABQYHD', 'pos11', 'Point of Sale - Quy Nhon', 'Staffed checkout lanes in a physical store', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000000E', '01JWNY20G23KD4RV5VWYABQYHD', 'pos12', 'Point of Sale - Vinh', 'Staffed checkout lanes in a physical store', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000000F', '01JWNY20G23KD4RV5VWYABQYHD', 'ecom01', 'Online Store - Ha Noi', 'The retailer-operated web storefront', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000000G', '01JWNY20G23KD4RV5VWYABQYHD', 'ecom02', 'Online Store - Ho Chi Minh', 'The retailer-operated web storefront', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000000H', '01JWNY20G23KD4RV5VWYABQYHD', 'ecom03', 'Online Store - Da Nang', 'The retailer-operated web storefront', NULL, 'suspended', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000000J', '01JWNY20G23KD4RV5VWYABQYHD', 'ecom04', 'Online Store - Hai Phong', 'The retailer-operated web storefront', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000000K', '01JWNY20G23KD4RV5VWYABQYHD', 'ecom05', 'Online Store - Can Tho', 'The retailer-operated web storefront', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000000M', '01JWNY20G23KD4RV5VWYABQYHD', 'ecom06', 'Online Store - Bien Hoa', 'The retailer-operated web storefront', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000000N', '01JWNY20G23KD4RV5VWYABQYHD', 'ecom07', 'Online Store - Nha Trang', 'The retailer-operated web storefront', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000000P', '01JWNY20G23KD4RV5VWYABQYHD', 'ecom08', 'Online Store - Hue', 'The retailer-operated web storefront', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000000Q', '01JWNY20G23KD4RV5VWYABQYHD', 'ecom09', 'Online Store - Vung Tau', 'The retailer-operated web storefront', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000000R', '01JWNY20G23KD4RV5VWYABQYHD', 'ecom10', 'Online Store - Buon Ma Thuot', 'The retailer-operated web storefront', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000000S', '01JWNY20G23KD4RV5VWYABQYHD', 'ecom11', 'Online Store - Quy Nhon', 'The retailer-operated web storefront', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000000T', '01JWNY20G23KD4RV5VWYABQYHD', 'ecom12', 'Online Store - Vinh', 'The retailer-operated web storefront', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000000V', '01JWNY20G23KD4RV5VWYABQYHD', 'mobile01', 'Mobile App - Ha Noi', 'The customer-facing iOS and Android app', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000000W', '01JWNY20G23KD4RV5VWYABQYHD', 'mobile02', 'Mobile App - Ho Chi Minh', 'The customer-facing iOS and Android app', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000000X', '01JWNY20G23KD4RV5VWYABQYHD', 'mobile03', 'Mobile App - Da Nang', 'The customer-facing iOS and Android app', NULL, 'active', false, true, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000000Y', '01JWNY20G23KD4RV5VWYABQYHD', 'mobile04', 'Mobile App - Hai Phong', 'The customer-facing iOS and Android app', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000000Z', '01JWNY20G23KD4RV5VWYABQYHD', 'mobile05', 'Mobile App - Can Tho', 'The customer-facing iOS and Android app', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC00000000000000000010', '01JWNY20G23KD4RV5VWYABQYHD', 'mobile06', 'Mobile App - Bien Hoa', 'The customer-facing iOS and Android app', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC00000000000000000011', '01JWNY20G23KD4RV5VWYABQYHD', 'mobile07', 'Mobile App - Nha Trang', 'The customer-facing iOS and Android app', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC00000000000000000012', '01JWNY20G23KD4RV5VWYABQYHD', 'mobile08', 'Mobile App - Hue', 'The customer-facing iOS and Android app', NULL, 'suspended', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC00000000000000000013', '01JWNY20G23KD4RV5VWYABQYHD', 'mobile09', 'Mobile App - Vung Tau', 'The customer-facing iOS and Android app', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC00000000000000000014', '01JWNY20G23KD4RV5VWYABQYHD', 'mobile10', 'Mobile App - Buon Ma Thuot', 'The customer-facing iOS and Android app', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC00000000000000000015', '01JWNY20G23KD4RV5VWYABQYHD', 'mobile11', 'Mobile App - Quy Nhon', 'The customer-facing iOS and Android app', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC00000000000000000016', '01JWNY20G23KD4RV5VWYABQYHD', 'mobile12', 'Mobile App - Vinh', 'The customer-facing iOS and Android app', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC00000000000000000017', '01JWNY20G23KD4RV5VWYABQYHD', 'mktplace01', 'Marketplace - Ha Noi', 'Third-party marketplace storefronts', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC00000000000000000018', '01JWNY20G23KD4RV5VWYABQYHD', 'mktplace02', 'Marketplace - Ho Chi Minh', 'Third-party marketplace storefronts', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC00000000000000000019', '01JWNY20G23KD4RV5VWYABQYHD', 'mktplace03', 'Marketplace - Da Nang', 'Third-party marketplace storefronts', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000001A', '01JWNY20G23KD4RV5VWYABQYHD', 'mktplace04', 'Marketplace - Hai Phong', 'Third-party marketplace storefronts', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000001B', '01JWNY20G23KD4RV5VWYABQYHD', 'mktplace05', 'Marketplace - Can Tho', 'Third-party marketplace storefronts', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000001C', '01JWNY20G23KD4RV5VWYABQYHD', 'mktplace06', 'Marketplace - Bien Hoa', 'Third-party marketplace storefronts', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000001D', '01JWNY20G23KD4RV5VWYABQYHD', 'mktplace07', 'Marketplace - Nha Trang', 'Third-party marketplace storefronts', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000001E', '01JWNY20G23KD4RV5VWYABQYHD', 'mktplace08', 'Marketplace - Hue', 'Third-party marketplace storefronts', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000001F', '01JWNY20G23KD4RV5VWYABQYHD', 'mktplace09', 'Marketplace - Vung Tau', 'Third-party marketplace storefronts', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000001G', '01JWNY20G23KD4RV5VWYABQYHD', 'mktplace10', 'Marketplace - Buon Ma Thuot', 'Third-party marketplace storefronts', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000001H', '01JWNY20G23KD4RV5VWYABQYHD', 'mktplace11', 'Marketplace - Quy Nhon', 'Third-party marketplace storefronts', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000001J', '01JWNY20G23KD4RV5VWYABQYHD', 'mktplace12', 'Marketplace - Vinh', 'Third-party marketplace storefronts', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000001K', '01JWNY20G23KD4RV5VWYABQYHD', 'phone01', 'Telesales - Ha Noi', 'Orders taken over the telephone', NULL, 'suspended', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000001M', '01JWNY20G23KD4RV5VWYABQYHD', 'phone02', 'Telesales - Ho Chi Minh', 'Orders taken over the telephone', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000001N', '01JWNY20G23KD4RV5VWYABQYHD', 'phone03', 'Telesales - Da Nang', 'Orders taken over the telephone', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000001P', '01JWNY20G23KD4RV5VWYABQYHD', 'phone04', 'Telesales - Hai Phong', 'Orders taken over the telephone', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000001Q', '01JWNY20G23KD4RV5VWYABQYHD', 'phone05', 'Telesales - Can Tho', 'Orders taken over the telephone', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000001R', '01JWNY20G23KD4RV5VWYABQYHD', 'phone06', 'Telesales - Bien Hoa', 'Orders taken over the telephone', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000001S', '01JWNY20G23KD4RV5VWYABQYHD', 'phone07', 'Telesales - Nha Trang', 'Orders taken over the telephone', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000001T', '01JWNY20G23KD4RV5VWYABQYHD', 'phone08', 'Telesales - Hue', 'Orders taken over the telephone', NULL, 'active', false, true, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000001V', '01JWNY20G23KD4RV5VWYABQYHD', 'phone09', 'Telesales - Vung Tau', 'Orders taken over the telephone', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000001W', '01JWNY20G23KD4RV5VWYABQYHD', 'phone10', 'Telesales - Buon Ma Thuot', 'Orders taken over the telephone', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000001X', '01JWNY20G23KD4RV5VWYABQYHD', 'phone11', 'Telesales - Quy Nhon', 'Orders taken over the telephone', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000001Y', '01JWNY20G23KD4RV5VWYABQYHD', 'phone12', 'Telesales - Vinh', 'Orders taken over the telephone', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000001Z', '01JWNY20G23KD4RV5VWYABQYHD', 'wholesale01', 'Wholesale - Ha Noi', 'Bulk sales to trade customers', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC00000000000000000020', '01JWNY20G23KD4RV5VWYABQYHD', 'wholesale02', 'Wholesale - Ho Chi Minh', 'Bulk sales to trade customers', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC00000000000000000021', '01JWNY20G23KD4RV5VWYABQYHD', 'wholesale03', 'Wholesale - Da Nang', 'Bulk sales to trade customers', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC00000000000000000022', '01JWNY20G23KD4RV5VWYABQYHD', 'wholesale04', 'Wholesale - Hai Phong', 'Bulk sales to trade customers', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC00000000000000000023', '01JWNY20G23KD4RV5VWYABQYHD', 'wholesale05', 'Wholesale - Can Tho', 'Bulk sales to trade customers', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC00000000000000000024', '01JWNY20G23KD4RV5VWYABQYHD', 'wholesale06', 'Wholesale - Bien Hoa', 'Bulk sales to trade customers', NULL, 'suspended', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC00000000000000000025', '01JWNY20G23KD4RV5VWYABQYHD', 'wholesale07', 'Wholesale - Nha Trang', 'Bulk sales to trade customers', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC00000000000000000026', '01JWNY20G23KD4RV5VWYABQYHD', 'wholesale08', 'Wholesale - Hue', 'Bulk sales to trade customers', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC00000000000000000027', '01JWNY20G23KD4RV5VWYABQYHD', 'wholesale09', 'Wholesale - Vung Tau', 'Bulk sales to trade customers', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC00000000000000000028', '01JWNY20G23KD4RV5VWYABQYHD', 'wholesale10', 'Wholesale - Buon Ma Thuot', 'Bulk sales to trade customers', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC00000000000000000029', '01JWNY20G23KD4RV5VWYABQYHD', 'wholesale11', 'Wholesale - Quy Nhon', 'Bulk sales to trade customers', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000002A', '01JWNY20G23KD4RV5VWYABQYHD', 'wholesale12', 'Wholesale - Vinh', 'Bulk sales to trade customers', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000002B', '01JWNY20G23KD4RV5VWYABQYHD', 'corp01', 'Corporate Sales - Ha Noi', 'Negotiated contracts with corporate accounts', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000002C', '01JWNY20G23KD4RV5VWYABQYHD', 'corp02', 'Corporate Sales - Ho Chi Minh', 'Negotiated contracts with corporate accounts', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000002D', '01JWNY20G23KD4RV5VWYABQYHD', 'corp03', 'Corporate Sales - Da Nang', 'Negotiated contracts with corporate accounts', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000002E', '01JWNY20G23KD4RV5VWYABQYHD', 'corp04', 'Corporate Sales - Hai Phong', 'Negotiated contracts with corporate accounts', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000002F', '01JWNY20G23KD4RV5VWYABQYHD', 'corp05', 'Corporate Sales - Can Tho', 'Negotiated contracts with corporate accounts', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000002G', '01JWNY20G23KD4RV5VWYABQYHD', 'corp06', 'Corporate Sales - Bien Hoa', 'Negotiated contracts with corporate accounts', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000002H', '01JWNY20G23KD4RV5VWYABQYHD', 'corp07', 'Corporate Sales - Nha Trang', 'Negotiated contracts with corporate accounts', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000002J', '01JWNY20G23KD4RV5VWYABQYHD', 'corp08', 'Corporate Sales - Hue', 'Negotiated contracts with corporate accounts', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000002K', '01JWNY20G23KD4RV5VWYABQYHD', 'corp09', 'Corporate Sales - Vung Tau', 'Negotiated contracts with corporate accounts', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000002M', '01JWNY20G23KD4RV5VWYABQYHD', 'corp10', 'Corporate Sales - Buon Ma Thuot', 'Negotiated contracts with corporate accounts', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000002N', '01JWNY20G23KD4RV5VWYABQYHD', 'corp11', 'Corporate Sales - Quy Nhon', 'Negotiated contracts with corporate accounts', NULL, 'suspended', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000002P', '01JWNY20G23KD4RV5VWYABQYHD', 'corp12', 'Corporate Sales - Vinh', 'Negotiated contracts with corporate accounts', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000002Q', '01JWNY20G23KD4RV5VWYABQYHD', 'kiosk01', 'Self Checkout - Ha Noi', 'Attended-store self-checkout terminals', NULL, 'active', false, true, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000002R', '01JWNY20G23KD4RV5VWYABQYHD', 'kiosk02', 'Self Checkout - Ho Chi Minh', 'Attended-store self-checkout terminals', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000002S', '01JWNY20G23KD4RV5VWYABQYHD', 'kiosk03', 'Self Checkout - Da Nang', 'Attended-store self-checkout terminals', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000002T', '01JWNY20G23KD4RV5VWYABQYHD', 'kiosk04', 'Self Checkout - Hai Phong', 'Attended-store self-checkout terminals', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000002V', '01JWNY20G23KD4RV5VWYABQYHD', 'kiosk05', 'Self Checkout - Can Tho', 'Attended-store self-checkout terminals', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000002W', '01JWNY20G23KD4RV5VWYABQYHD', 'kiosk06', 'Self Checkout - Bien Hoa', 'Attended-store self-checkout terminals', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000002X', '01JWNY20G23KD4RV5VWYABQYHD', 'kiosk07', 'Self Checkout - Nha Trang', 'Attended-store self-checkout terminals', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000002Y', '01JWNY20G23KD4RV5VWYABQYHD', 'kiosk08', 'Self Checkout - Hue', 'Attended-store self-checkout terminals', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000002Z', '01JWNY20G23KD4RV5VWYABQYHD', 'kiosk09', 'Self Checkout - Vung Tau', 'Attended-store self-checkout terminals', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC00000000000000000030', '01JWNY20G23KD4RV5VWYABQYHD', 'kiosk10', 'Self Checkout - Buon Ma Thuot', 'Attended-store self-checkout terminals', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC00000000000000000031', '01JWNY20G23KD4RV5VWYABQYHD', 'kiosk11', 'Self Checkout - Quy Nhon', 'Attended-store self-checkout terminals', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC00000000000000000032', '01JWNY20G23KD4RV5VWYABQYHD', 'kiosk12', 'Self Checkout - Vinh', 'Attended-store self-checkout terminals', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC00000000000000000033', '01JWNY20G23KD4RV5VWYABQYHD', 'popup01', 'Pop-up Store - Ha Noi', 'Temporary retail locations and event stands', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC00000000000000000034', '01JWNY20G23KD4RV5VWYABQYHD', 'popup02', 'Pop-up Store - Ho Chi Minh', 'Temporary retail locations and event stands', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC00000000000000000035', '01JWNY20G23KD4RV5VWYABQYHD', 'popup03', 'Pop-up Store - Da Nang', 'Temporary retail locations and event stands', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC00000000000000000036', '01JWNY20G23KD4RV5VWYABQYHD', 'popup04', 'Pop-up Store - Hai Phong', 'Temporary retail locations and event stands', NULL, 'suspended', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC00000000000000000037', '01JWNY20G23KD4RV5VWYABQYHD', 'popup05', 'Pop-up Store - Can Tho', 'Temporary retail locations and event stands', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC00000000000000000038', '01JWNY20G23KD4RV5VWYABQYHD', 'popup06', 'Pop-up Store - Bien Hoa', 'Temporary retail locations and event stands', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC00000000000000000039', '01JWNY20G23KD4RV5VWYABQYHD', 'popup07', 'Pop-up Store - Nha Trang', 'Temporary retail locations and event stands', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000003A', '01JWNY20G23KD4RV5VWYABQYHD', 'popup08', 'Pop-up Store - Hue', 'Temporary retail locations and event stands', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000003B', '01JWNY20G23KD4RV5VWYABQYHD', 'popup09', 'Pop-up Store - Vung Tau', 'Temporary retail locations and event stands', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000003C', '01JWNY20G23KD4RV5VWYABQYHD', 'popup10', 'Pop-up Store - Buon Ma Thuot', 'Temporary retail locations and event stands', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000003D', '01JWNY20G23KD4RV5VWYABQYHD', 'popup11', 'Pop-up Store - Quy Nhon', 'Temporary retail locations and event stands', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000003E', '01JWNY20G23KD4RV5VWYABQYHD', 'popup12', 'Pop-up Store - Vinh', 'Temporary retail locations and event stands', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000003F', '01JWNY20G23KD4RV5VWYABQYHD', 'export01', 'Export - Ha Noi', 'Cross-border sales to overseas distributors', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000003G', '01JWNY20G23KD4RV5VWYABQYHD', 'export02', 'Export - Ho Chi Minh', 'Cross-border sales to overseas distributors', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000003H', '01JWNY20G23KD4RV5VWYABQYHD', 'export03', 'Export - Da Nang', 'Cross-border sales to overseas distributors', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000003J', '01JWNY20G23KD4RV5VWYABQYHD', 'export04', 'Export - Hai Phong', 'Cross-border sales to overseas distributors', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000003K', '01JWNY20G23KD4RV5VWYABQYHD', 'export05', 'Export - Can Tho', 'Cross-border sales to overseas distributors', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000003M', '01JWNY20G23KD4RV5VWYABQYHD', 'export06', 'Export - Bien Hoa', 'Cross-border sales to overseas distributors', NULL, 'active', false, true, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000003N', '01JWNY20G23KD4RV5VWYABQYHD', 'export07', 'Export - Nha Trang', 'Cross-border sales to overseas distributors', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000003P', '01JWNY20G23KD4RV5VWYABQYHD', 'export08', 'Export - Hue', 'Cross-border sales to overseas distributors', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000003Q', '01JWNY20G23KD4RV5VWYABQYHD', 'export09', 'Export - Vung Tau', 'Cross-border sales to overseas distributors', NULL, 'suspended', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000003R', '01JWNY20G23KD4RV5VWYABQYHD', 'export10', 'Export - Buon Ma Thuot', 'Cross-border sales to overseas distributors', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000003S', '01JWNY20G23KD4RV5VWYABQYHD', 'export11', 'Export - Quy Nhon', 'Cross-border sales to overseas distributors', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SC0000000000000000003T', '01JWNY20G23KD4RV5VWYABQYHD', 'export12', 'Export - Vinh', 'Cross-border sales to overseas distributors', NULL, 'active', false, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'sales_points'
	) THEN
		INSERT INTO "sales_points" (
			"id", "org_id", "sales_channel_id", "name", "code", "external_reference_id", "external_reference_type", "status", "is_archived", "created_at", "updated_at", "etag"
		) VALUES
		('01M3SP00000000000000000001', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000001', 'Kiosk VM001 - Ha Noi', 'VM001', '01M3KI00000000000000000001', 'vending_machine.kiosk', 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP00000000000000000002', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000001', 'Kiosk VM002 - Ho Chi Minh', 'VM002', '01M3KI00000000000000000002', 'vending_machine.kiosk', 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP00000000000000000003', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000001', 'Kiosk VM003 - Da Nang', 'VM003', '01M3KI00000000000000000003', 'vending_machine.kiosk', 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP00000000000000000004', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000001', 'Kiosk VM004 - Hai Phong', 'VM004', '01M3KI00000000000000000004', 'vending_machine.kiosk', 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP00000000000000000005', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000001', 'Kiosk VM005 - Can Tho', 'VM005', '01M3KI00000000000000000005', 'vending_machine.kiosk', 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP00000000000000000006', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000001', 'Kiosk VM006 - Bien Hoa', 'VM006', '01M3KI00000000000000000006', 'vending_machine.kiosk', 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP00000000000000000007', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000001', 'Kiosk VM007 - Nha Trang', 'VM007', '01M3KI00000000000000000007', 'vending_machine.kiosk', 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP00000000000000000008', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000001', 'Kiosk VM008 - Hue', 'VM008', '01M3KI00000000000000000008', 'vending_machine.kiosk', 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP00000000000000000009', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000001', 'Kiosk VM009 - Vung Tau', 'VM009', '01M3KI00000000000000000009', 'vending_machine.kiosk', 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000000A', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000001', 'Kiosk VM010 - Buon Ma Thuot', 'VM010', '01M3KI0000000000000000000A', 'vending_machine.kiosk', 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000000B', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000001', 'Kiosk VM011 - Quy Nhon', 'VM011', '01M3KI0000000000000000000B', 'vending_machine.kiosk', 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000000C', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000001', 'Kiosk VM012 - Vinh', 'VM012', '01M3KI0000000000000000000C', 'vending_machine.kiosk', 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000000D', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000001', 'Kiosk VM013 - Ha Noi', 'VM013', '01M3KI0000000000000000000D', 'vending_machine.kiosk', 'suspended', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000000E', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000001', 'Kiosk VM014 - Ho Chi Minh', 'VM014', '01M3KI0000000000000000000E', 'vending_machine.kiosk', 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000000F', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000001', 'Kiosk VM015 - Da Nang', 'VM015', '01M3KI0000000000000000000F', 'vending_machine.kiosk', 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000000G', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000001', 'Kiosk VM016 - Hai Phong', 'VM016', '01M3KI0000000000000000000G', 'vending_machine.kiosk', 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000000H', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000001', 'Kiosk VM017 - Can Tho', 'VM017', '01M3KI0000000000000000000H', 'vending_machine.kiosk', 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000000J', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000001', 'Kiosk VM018 - Bien Hoa', 'VM018', '01M3KI0000000000000000000J', 'vending_machine.kiosk', 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000000K', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000001', 'Kiosk VM019 - Nha Trang', 'VM019', '01M3KI0000000000000000000K', 'vending_machine.kiosk', 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000000M', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000001', 'Kiosk VM020 - Hue', 'VM020', '01M3KI0000000000000000000M', 'vending_machine.kiosk', 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000000N', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000001', 'Kiosk VM021 - Vung Tau', 'VM021', '01M3KI0000000000000000000N', 'vending_machine.kiosk', 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000000P', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000001', 'Kiosk VM022 - Buon Ma Thuot', 'VM022', '01M3KI0000000000000000000P', 'vending_machine.kiosk', 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000000Q', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000001', 'Kiosk VM023 - Quy Nhon', 'VM023', '01M3KI0000000000000000000Q', 'vending_machine.kiosk', 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000000R', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000001', 'Kiosk VM024 - Vinh', 'VM024', '01M3KI0000000000000000000R', 'vending_machine.kiosk', 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000000S', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000001', 'Kiosk VM025 - Ha Noi', 'VM025', '01M3KI0000000000000000000S', 'vending_machine.kiosk', 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000000T', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000001', 'Kiosk VM026 - Ho Chi Minh', 'VM026', '01M3KI0000000000000000000T', 'vending_machine.kiosk', 'suspended', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000000V', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000001', 'Kiosk VM027 - Da Nang', 'VM027', '01M3KI0000000000000000000V', 'vending_machine.kiosk', 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000000W', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000001', 'Kiosk VM028 - Hai Phong', 'VM028', '01M3KI0000000000000000000W', 'vending_machine.kiosk', 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000000X', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000001', 'Kiosk VM029 - Can Tho', 'VM029', '01M3KI0000000000000000000X', 'vending_machine.kiosk', 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000000Y', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000001', 'Kiosk VM030 - Bien Hoa', 'VM030', '01M3KI0000000000000000000Y', 'vending_machine.kiosk', 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000000Z', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000001', 'Kiosk VM031 - Nha Trang', 'VM031', '01M3KI0000000000000000000Z', 'vending_machine.kiosk', 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP00000000000000000010', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000001', 'Kiosk VM032 - Hue', 'VM032', '01M3KI00000000000000000010', 'vending_machine.kiosk', 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP00000000000000000011', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000001', 'Kiosk VM033 - Vung Tau', 'VM033', '01M3KI00000000000000000011', 'vending_machine.kiosk', 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP00000000000000000012', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000001', 'Kiosk VM034 - Buon Ma Thuot', 'VM034', '01M3KI00000000000000000012', 'vending_machine.kiosk', 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP00000000000000000013', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000001', 'Kiosk VM035 - Quy Nhon', 'VM035', '01M3KI00000000000000000013', 'vending_machine.kiosk', 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP00000000000000000014', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000001', 'Kiosk VM036 - Vinh', 'VM036', '01M3KI00000000000000000014', 'vending_machine.kiosk', 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP00000000000000000015', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000001', 'Kiosk VM037 - Ha Noi', 'VM037', '01M3KI00000000000000000015', 'vending_machine.kiosk', 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP00000000000000000016', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000001', 'Kiosk VM038 - Ho Chi Minh', 'VM038', '01M3KI00000000000000000016', 'vending_machine.kiosk', 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP00000000000000000017', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000001', 'Kiosk VM039 - Da Nang', 'VM039', '01M3KI00000000000000000017', 'vending_machine.kiosk', 'suspended', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP00000000000000000018', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000001', 'Kiosk VM040 - Hai Phong', 'VM040', '01M3KI00000000000000000018', 'vending_machine.kiosk', 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP00000000000000000019', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000001', 'Kiosk VM041 - Can Tho', 'VM041', '01M3KI00000000000000000019', 'vending_machine.kiosk', 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000001A', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000001', 'Kiosk VM042 - Bien Hoa', 'VM042', '01M3KI0000000000000000001A', 'vending_machine.kiosk', 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000001B', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000001', 'Kiosk VM043 - Nha Trang', 'VM043', '01M3KI0000000000000000001B', 'vending_machine.kiosk', 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000001C', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000001', 'Kiosk VM044 - Hue', 'VM044', '01M3KI0000000000000000001C', 'vending_machine.kiosk', 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000001D', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000001', 'Kiosk VM045 - Vung Tau', 'VM045', '01M3KI0000000000000000001D', 'vending_machine.kiosk', 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000001E', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000001', 'Kiosk VM046 - Buon Ma Thuot', 'VM046', '01M3KI0000000000000000001E', 'vending_machine.kiosk', 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000001F', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000001', 'Kiosk VM047 - Quy Nhon', 'VM047', '01M3KI0000000000000000001F', 'vending_machine.kiosk', 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000001G', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000001', 'Kiosk VM048 - Vinh', 'VM048', '01M3KI0000000000000000001G', 'vending_machine.kiosk', 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000001H', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000001', 'Kiosk VM049 - Ha Noi', 'VM049', '01M3KI0000000000000000001H', 'vending_machine.kiosk', 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000001J', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000001', 'Kiosk VM050 - Ho Chi Minh', 'VM050', '01M3KI0000000000000000001J', 'vending_machine.kiosk', 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000001K', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000001', 'Kiosk VM051 - Da Nang', 'VM051', '01M3KI0000000000000000001K', 'vending_machine.kiosk', 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000001M', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000001', 'Kiosk VM052 - Hai Phong', 'VM052', '01M3KI0000000000000000001M', 'vending_machine.kiosk', 'suspended', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000001N', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000001', 'Kiosk VM053 - Can Tho', 'VM053', '01M3KI0000000000000000001N', 'vending_machine.kiosk', 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000001P', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000001', 'Kiosk VM054 - Bien Hoa', 'VM054', '01M3KI0000000000000000001P', 'vending_machine.kiosk', 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000001Q', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000001', 'Kiosk VM055 - Nha Trang', 'VM055', '01M3KI0000000000000000001Q', 'vending_machine.kiosk', 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000001R', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000001', 'Kiosk VM056 - Hue', 'VM056', '01M3KI0000000000000000001R', 'vending_machine.kiosk', 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000001S', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000001', 'Kiosk VM057 - Vung Tau', 'VM057', '01M3KI0000000000000000001S', 'vending_machine.kiosk', 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000001T', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000001', 'Kiosk VM058 - Buon Ma Thuot', 'VM058', '01M3KI0000000000000000001T', 'vending_machine.kiosk', 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000001V', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000001', 'Kiosk VM059 - Quy Nhon', 'VM059', '01M3KI0000000000000000001V', 'vending_machine.kiosk', 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000001W', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000001', 'Kiosk VM060 - Vinh', 'VM060', '01M3KI0000000000000000001W', 'vending_machine.kiosk', 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000001X', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000002', 'Core Mart Back Office', 'BO001', NULL, NULL, 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000001Y', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000003', 'Ha Noi Store ST', 'ST001', NULL, NULL, 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000001Z', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000004', 'Ho Chi Minh Express EX', 'EX002', NULL, NULL, 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP00000000000000000020', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000005', 'Da Nang Depot DP', 'DP003', NULL, NULL, 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP00000000000000000021', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000006', 'Hai Phong Counter CT', 'CT004', NULL, NULL, 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP00000000000000000022', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000007', 'Can Tho Store ST', 'ST005', NULL, NULL, 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP00000000000000000023', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000008', 'Bien Hoa Express EX', 'EX006', NULL, NULL, 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP00000000000000000024', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000009', 'Nha Trang Depot DP', 'DP007', NULL, NULL, 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP00000000000000000025', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000A', 'Hue Counter CT', 'CT008', NULL, NULL, 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP00000000000000000026', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000B', 'Vung Tau Store ST', 'ST009', NULL, NULL, 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP00000000000000000027', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000C', 'Buon Ma Thuot Express EX', 'EX010', NULL, NULL, 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP00000000000000000028', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000D', 'Quy Nhon Depot DP', 'DP011', NULL, NULL, 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP00000000000000000029', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000E', 'Vinh Counter CT', 'CT012', NULL, NULL, 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000002A', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000F', 'Ha Noi Store ST', 'ST013', NULL, NULL, 'active', true, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000002B', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000G', 'Ho Chi Minh Express EX', 'EX014', NULL, NULL, 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000002C', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000H', 'Da Nang Depot DP', 'DP015', NULL, NULL, 'suspended', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000002D', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000J', 'Hai Phong Counter CT', 'CT016', NULL, NULL, 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000002E', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000K', 'Can Tho Store ST', 'ST017', NULL, NULL, 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000002F', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000M', 'Bien Hoa Express EX', 'EX018', NULL, NULL, 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000002G', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000N', 'Nha Trang Depot DP', 'DP019', NULL, NULL, 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000002H', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000P', 'Hue Counter CT', 'CT020', NULL, NULL, 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000002J', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000Q', 'Vung Tau Store ST', 'ST021', NULL, NULL, 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000002K', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000R', 'Buon Ma Thuot Express EX', 'EX022', NULL, NULL, 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000002M', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000S', 'Quy Nhon Depot DP', 'DP023', NULL, NULL, 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000002N', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000T', 'Vinh Counter CT', 'CT024', NULL, NULL, 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000002P', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000V', 'Ha Noi Store ST', 'ST025', NULL, NULL, 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000002Q', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000W', 'Ho Chi Minh Express EX', 'EX026', NULL, NULL, 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000002R', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000Y', 'Da Nang Depot DP', 'DP027', NULL, NULL, 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000002S', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000Z', 'Hai Phong Counter CT', 'CT028', NULL, NULL, 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000002T', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000010', 'Can Tho Store ST', 'ST029', NULL, NULL, 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000002V', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000011', 'Bien Hoa Express EX', 'EX030', NULL, NULL, 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000002W', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000012', 'Nha Trang Depot DP', 'DP031', NULL, NULL, 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000002X', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000013', 'Hue Counter CT', 'CT032', NULL, NULL, 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000002Y', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000014', 'Vung Tau Store ST', 'ST033', NULL, NULL, 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000002Z', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000015', 'Buon Ma Thuot Express EX', 'EX034', NULL, NULL, 'suspended', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP00000000000000000030', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000016', 'Quy Nhon Depot DP', 'DP035', NULL, NULL, 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP00000000000000000031', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000017', 'Vinh Counter CT', 'CT036', NULL, NULL, 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP00000000000000000032', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000018', 'Ha Noi Store ST', 'ST037', NULL, NULL, 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP00000000000000000033', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000019', 'Ho Chi Minh Express EX', 'EX038', NULL, NULL, 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP00000000000000000034', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000001A', 'Da Nang Depot DP', 'DP039', NULL, NULL, 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP00000000000000000035', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000001B', 'Hai Phong Counter CT', 'CT040', NULL, NULL, 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP00000000000000000036', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000001C', 'Can Tho Store ST', 'ST041', NULL, NULL, 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP00000000000000000037', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000001D', 'Bien Hoa Express EX', 'EX042', NULL, NULL, 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP00000000000000000038', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000001E', 'Nha Trang Depot DP', 'DP043', NULL, NULL, 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP00000000000000000039', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000001F', 'Hue Counter CT', 'CT044', NULL, NULL, 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000003A', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000001G', 'Vung Tau Store ST', 'ST045', NULL, NULL, 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000003B', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000001H', 'Buon Ma Thuot Express EX', 'EX046', NULL, NULL, 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000003C', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000001J', 'Quy Nhon Depot DP', 'DP047', NULL, NULL, 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000003D', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000001K', 'Vinh Counter CT', 'CT048', NULL, NULL, 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000003E', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000001M', 'Ha Noi Store ST', 'ST049', NULL, NULL, 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000003F', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000001N', 'Ho Chi Minh Express EX', 'EX050', NULL, NULL, 'active', true, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000003G', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000001P', 'Da Nang Depot DP', 'DP051', NULL, NULL, 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000003H', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000001Q', 'Hai Phong Counter CT', 'CT052', NULL, NULL, 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000003J', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000001R', 'Can Tho Store ST', 'ST053', NULL, NULL, 'suspended', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000003K', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000001S', 'Bien Hoa Express EX', 'EX054', NULL, NULL, 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000003M', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000001V', 'Nha Trang Depot DP', 'DP055', NULL, NULL, 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000003N', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000001W', 'Hue Counter CT', 'CT056', NULL, NULL, 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000003P', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000001X', 'Vung Tau Store ST', 'ST057', NULL, NULL, 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000003Q', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000001Y', 'Buon Ma Thuot Express EX', 'EX058', NULL, NULL, 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000003R', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000001Z', 'Quy Nhon Depot DP', 'DP059', NULL, NULL, 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SP0000000000000000003S', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000020', 'Vinh Counter CT', 'CT060', NULL, NULL, 'active', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;

-- Which payment methods each channel accepts.
--
-- The row IS the state (CR 27): there is no is_enabled column, so a channel accepts exactly the
-- methods it has a row for and nothing else. That is default-deny (CR 76), and it is why roughly
-- one channel in seven below is left with no rows at all - an unconfigured channel accepting
-- nothing is a real state the merged view must render, and a seed that configured every channel
-- would never show it.
--
-- payment_method_id names a row in paymentinvoice_payment_methods, seeded by
-- 0004003_paymentinvoice_seeds.sql. There is no foreign key between them (CR 25): the two tables
-- belong to different modules, and a constraint across that boundary would couple their
-- migrations. The mappings onto "mbbank" are the interesting ones - that method exists and is
-- active, but no adapter ships for it, so the merged view must report those as enabled and
-- unusable rather than as either working or absent.


DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'sales_channel_payment_rel'
	) THEN
		INSERT INTO "sales_channel_payment_rel" (
			"id", "org_id", "sales_channel_id", "payment_method_id", "created_at", "updated_at"
		) VALUES
		('01M3CPR0000000000000000001', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000001', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR0000000000000000002', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000001', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR0000000000000000003', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000001', '01M3PM00000000000000000003', NOW(), NULL),
		('01M3CPR0000000000000000004', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000001', '01M3PM00000000000000000004', NOW(), NULL),
		('01M3CPR0000000000000000005', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000002', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR0000000000000000006', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000002', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR0000000000000000007', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000002', '01M3PM00000000000000000004', NOW(), NULL),
		('01M3CPR0000000000000000008', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000003', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR0000000000000000009', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000004', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR000000000000000000A', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000004', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000000B', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000004', '01M3PM00000000000000000003', NOW(), NULL),
		('01M3CPR000000000000000000C', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000004', '01M3PM00000000000000000004', NOW(), NULL),
		('01M3CPR000000000000000000D', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000005', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR000000000000000000E', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000005', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000000F', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000005', '01M3PM00000000000000000004', NOW(), NULL),
		('01M3CPR000000000000000000G', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000006', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000000H', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000008', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR000000000000000000J', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000008', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000000K', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000008', '01M3PM00000000000000000004', NOW(), NULL),
		('01M3CPR000000000000000000M', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000009', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000000N', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000A', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR000000000000000000P', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000A', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000000Q', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000A', '01M3PM00000000000000000003', NOW(), NULL),
		('01M3CPR000000000000000000R', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000A', '01M3PM00000000000000000004', NOW(), NULL),
		('01M3CPR000000000000000000S', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000B', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR000000000000000000T', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000B', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000000V', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000B', '01M3PM00000000000000000004', NOW(), NULL),
		('01M3CPR000000000000000000W', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000C', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000000X', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000D', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR000000000000000000Y', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000D', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000000Z', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000D', '01M3PM00000000000000000003', NOW(), NULL),
		('01M3CPR0000000000000000010', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000D', '01M3PM00000000000000000004', NOW(), NULL),
		('01M3CPR0000000000000000011', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000F', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR0000000000000000012', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000G', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR0000000000000000013', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000G', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR0000000000000000014', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000G', '01M3PM00000000000000000003', NOW(), NULL),
		('01M3CPR0000000000000000015', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000G', '01M3PM00000000000000000004', NOW(), NULL),
		('01M3CPR0000000000000000016', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000H', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR0000000000000000017', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000H', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR0000000000000000018', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000H', '01M3PM00000000000000000004', NOW(), NULL),
		('01M3CPR0000000000000000019', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000J', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000001A', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000K', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR000000000000000001B', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000K', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000001C', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000K', '01M3PM00000000000000000003', NOW(), NULL),
		('01M3CPR000000000000000001D', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000K', '01M3PM00000000000000000004', NOW(), NULL),
		('01M3CPR000000000000000001E', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000M', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR000000000000000001F', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000M', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000001G', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000M', '01M3PM00000000000000000004', NOW(), NULL),
		('01M3CPR000000000000000001H', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000P', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR000000000000000001J', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000P', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000001K', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000P', '01M3PM00000000000000000003', NOW(), NULL),
		('01M3CPR000000000000000001M', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000P', '01M3PM00000000000000000004', NOW(), NULL),
		('01M3CPR000000000000000001N', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000Q', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR000000000000000001P', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000Q', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000001Q', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000Q', '01M3PM00000000000000000004', NOW(), NULL),
		('01M3CPR000000000000000001R', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000R', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000001S', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000S', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR000000000000000001T', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000S', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000001V', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000S', '01M3PM00000000000000000003', NOW(), NULL),
		('01M3CPR000000000000000001W', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000S', '01M3PM00000000000000000004', NOW(), NULL),
		('01M3CPR000000000000000001X', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000T', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR000000000000000001Y', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000T', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000001Z', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000T', '01M3PM00000000000000000004', NOW(), NULL),
		('01M3CPR0000000000000000020', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000V', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR0000000000000000021', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000X', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR0000000000000000022', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000X', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR0000000000000000023', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000X', '01M3PM00000000000000000004', NOW(), NULL),
		('01M3CPR0000000000000000024', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000Y', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR0000000000000000025', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000Z', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR0000000000000000026', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000Z', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR0000000000000000027', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000Z', '01M3PM00000000000000000003', NOW(), NULL),
		('01M3CPR0000000000000000028', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000000Z', '01M3PM00000000000000000004', NOW(), NULL),
		('01M3CPR0000000000000000029', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000010', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR000000000000000002A', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000010', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000002B', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000010', '01M3PM00000000000000000004', NOW(), NULL),
		('01M3CPR000000000000000002C', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000011', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000002D', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000012', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR000000000000000002E', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000012', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000002F', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000012', '01M3PM00000000000000000003', NOW(), NULL),
		('01M3CPR000000000000000002G', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000012', '01M3PM00000000000000000004', NOW(), NULL),
		('01M3CPR000000000000000002H', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000014', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000002J', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000015', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR000000000000000002K', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000015', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000002M', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000015', '01M3PM00000000000000000003', NOW(), NULL),
		('01M3CPR000000000000000002N', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000015', '01M3PM00000000000000000004', NOW(), NULL),
		('01M3CPR000000000000000002P', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000016', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR000000000000000002Q', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000016', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000002R', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000016', '01M3PM00000000000000000004', NOW(), NULL),
		('01M3CPR000000000000000002S', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000017', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000002T', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000018', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR000000000000000002V', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000018', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000002W', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000018', '01M3PM00000000000000000003', NOW(), NULL),
		('01M3CPR000000000000000002X', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000018', '01M3PM00000000000000000004', NOW(), NULL),
		('01M3CPR000000000000000002Y', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000019', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR000000000000000002Z', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000019', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR0000000000000000030', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000019', '01M3PM00000000000000000004', NOW(), NULL),
		('01M3CPR0000000000000000031', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000001B', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR0000000000000000032', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000001B', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR0000000000000000033', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000001B', '01M3PM00000000000000000003', NOW(), NULL),
		('01M3CPR0000000000000000034', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000001B', '01M3PM00000000000000000004', NOW(), NULL),
		('01M3CPR0000000000000000035', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000001C', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR0000000000000000036', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000001C', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR0000000000000000037', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000001C', '01M3PM00000000000000000004', NOW(), NULL),
		('01M3CPR0000000000000000038', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000001D', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR0000000000000000039', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000001E', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR000000000000000003A', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000001E', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000003B', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000001E', '01M3PM00000000000000000003', NOW(), NULL),
		('01M3CPR000000000000000003C', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000001E', '01M3PM00000000000000000004', NOW(), NULL),
		('01M3CPR000000000000000003D', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000001F', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR000000000000000003E', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000001F', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000003F', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000001F', '01M3PM00000000000000000004', NOW(), NULL),
		('01M3CPR000000000000000003G', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000001G', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000003H', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000001J', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR000000000000000003J', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000001J', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000003K', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000001J', '01M3PM00000000000000000004', NOW(), NULL),
		('01M3CPR000000000000000003M', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000001K', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000003N', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000001M', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR000000000000000003P', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000001M', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000003Q', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000001M', '01M3PM00000000000000000003', NOW(), NULL),
		('01M3CPR000000000000000003R', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000001M', '01M3PM00000000000000000004', NOW(), NULL),
		('01M3CPR000000000000000003S', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000001N', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR000000000000000003T', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000001N', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000003V', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000001N', '01M3PM00000000000000000004', NOW(), NULL),
		('01M3CPR000000000000000003W', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000001P', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000003X', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000001Q', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR000000000000000003Y', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000001Q', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000003Z', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000001Q', '01M3PM00000000000000000003', NOW(), NULL),
		('01M3CPR0000000000000000040', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000001Q', '01M3PM00000000000000000004', NOW(), NULL),
		('01M3CPR0000000000000000041', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000001S', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR0000000000000000042', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000001T', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR0000000000000000043', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000001T', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR0000000000000000044', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000001T', '01M3PM00000000000000000003', NOW(), NULL),
		('01M3CPR0000000000000000045', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000001T', '01M3PM00000000000000000004', NOW(), NULL),
		('01M3CPR0000000000000000046', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000001V', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR0000000000000000047', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000001V', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR0000000000000000048', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000001V', '01M3PM00000000000000000004', NOW(), NULL),
		('01M3CPR0000000000000000049', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000001W', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000004A', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000001X', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR000000000000000004B', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000001X', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000004C', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000001X', '01M3PM00000000000000000003', NOW(), NULL),
		('01M3CPR000000000000000004D', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000001X', '01M3PM00000000000000000004', NOW(), NULL),
		('01M3CPR000000000000000004E', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000001Y', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR000000000000000004F', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000001Y', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000004G', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000001Y', '01M3PM00000000000000000004', NOW(), NULL),
		('01M3CPR000000000000000004H', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000020', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR000000000000000004J', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000020', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000004K', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000020', '01M3PM00000000000000000003', NOW(), NULL),
		('01M3CPR000000000000000004M', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000020', '01M3PM00000000000000000004', NOW(), NULL),
		('01M3CPR000000000000000004N', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000021', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR000000000000000004P', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000021', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000004Q', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000021', '01M3PM00000000000000000004', NOW(), NULL),
		('01M3CPR000000000000000004R', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000022', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000004S', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000023', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR000000000000000004T', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000023', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000004V', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000023', '01M3PM00000000000000000003', NOW(), NULL),
		('01M3CPR000000000000000004W', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000023', '01M3PM00000000000000000004', NOW(), NULL),
		('01M3CPR000000000000000004X', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000024', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR000000000000000004Y', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000024', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000004Z', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000024', '01M3PM00000000000000000004', NOW(), NULL),
		('01M3CPR0000000000000000050', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000025', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR0000000000000000051', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000027', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR0000000000000000052', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000027', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR0000000000000000053', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000027', '01M3PM00000000000000000004', NOW(), NULL),
		('01M3CPR0000000000000000054', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000028', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR0000000000000000055', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000029', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR0000000000000000056', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000029', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR0000000000000000057', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000029', '01M3PM00000000000000000003', NOW(), NULL),
		('01M3CPR0000000000000000058', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000029', '01M3PM00000000000000000004', NOW(), NULL),
		('01M3CPR0000000000000000059', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000002A', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR000000000000000005A', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000002A', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000005B', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000002A', '01M3PM00000000000000000004', NOW(), NULL),
		('01M3CPR000000000000000005C', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000002B', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000005D', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000002C', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR000000000000000005E', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000002C', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000005F', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000002C', '01M3PM00000000000000000003', NOW(), NULL),
		('01M3CPR000000000000000005G', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000002C', '01M3PM00000000000000000004', NOW(), NULL),
		('01M3CPR000000000000000005H', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000002E', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000005J', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000002F', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR000000000000000005K', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000002F', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000005M', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000002F', '01M3PM00000000000000000003', NOW(), NULL),
		('01M3CPR000000000000000005N', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000002F', '01M3PM00000000000000000004', NOW(), NULL),
		('01M3CPR000000000000000005P', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000002G', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR000000000000000005Q', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000002G', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000005R', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000002G', '01M3PM00000000000000000004', NOW(), NULL),
		('01M3CPR000000000000000005S', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000002H', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000005T', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000002J', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR000000000000000005V', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000002J', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000005W', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000002J', '01M3PM00000000000000000003', NOW(), NULL),
		('01M3CPR000000000000000005X', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000002J', '01M3PM00000000000000000004', NOW(), NULL),
		('01M3CPR000000000000000005Y', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000002K', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR000000000000000005Z', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000002K', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR0000000000000000060', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000002K', '01M3PM00000000000000000004', NOW(), NULL),
		('01M3CPR0000000000000000061', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000002N', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR0000000000000000062', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000002N', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR0000000000000000063', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000002N', '01M3PM00000000000000000003', NOW(), NULL),
		('01M3CPR0000000000000000064', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000002N', '01M3PM00000000000000000004', NOW(), NULL),
		('01M3CPR0000000000000000065', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000002P', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR0000000000000000066', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000002P', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR0000000000000000067', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000002P', '01M3PM00000000000000000004', NOW(), NULL),
		('01M3CPR0000000000000000068', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000002Q', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR0000000000000000069', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000002R', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR000000000000000006A', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000002R', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000006B', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000002R', '01M3PM00000000000000000003', NOW(), NULL),
		('01M3CPR000000000000000006C', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000002R', '01M3PM00000000000000000004', NOW(), NULL),
		('01M3CPR000000000000000006D', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000002S', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR000000000000000006E', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000002S', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000006F', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000002S', '01M3PM00000000000000000004', NOW(), NULL),
		('01M3CPR000000000000000006G', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000002T', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000006H', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000002W', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR000000000000000006J', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000002W', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000006K', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000002W', '01M3PM00000000000000000004', NOW(), NULL),
		('01M3CPR000000000000000006M', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000002X', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000006N', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000002Y', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR000000000000000006P', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000002Y', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000006Q', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000002Y', '01M3PM00000000000000000003', NOW(), NULL),
		('01M3CPR000000000000000006R', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000002Y', '01M3PM00000000000000000004', NOW(), NULL),
		('01M3CPR000000000000000006S', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000002Z', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR000000000000000006T', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000002Z', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000006V', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000002Z', '01M3PM00000000000000000004', NOW(), NULL),
		('01M3CPR000000000000000006W', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000030', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000006X', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000031', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR000000000000000006Y', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000031', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000006Z', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000031', '01M3PM00000000000000000003', NOW(), NULL),
		('01M3CPR0000000000000000070', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000031', '01M3PM00000000000000000004', NOW(), NULL),
		('01M3CPR0000000000000000071', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000033', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR0000000000000000072', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000034', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR0000000000000000073', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000034', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR0000000000000000074', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000034', '01M3PM00000000000000000003', NOW(), NULL),
		('01M3CPR0000000000000000075', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000034', '01M3PM00000000000000000004', NOW(), NULL),
		('01M3CPR0000000000000000076', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000035', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR0000000000000000077', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000035', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR0000000000000000078', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000035', '01M3PM00000000000000000004', NOW(), NULL),
		('01M3CPR0000000000000000079', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000036', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000007A', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000037', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR000000000000000007B', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000037', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000007C', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000037', '01M3PM00000000000000000003', NOW(), NULL),
		('01M3CPR000000000000000007D', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000037', '01M3PM00000000000000000004', NOW(), NULL),
		('01M3CPR000000000000000007E', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000038', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR000000000000000007F', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000038', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000007G', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000038', '01M3PM00000000000000000004', NOW(), NULL),
		('01M3CPR000000000000000007H', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000003A', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR000000000000000007J', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000003A', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000007K', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000003A', '01M3PM00000000000000000003', NOW(), NULL),
		('01M3CPR000000000000000007M', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000003A', '01M3PM00000000000000000004', NOW(), NULL),
		('01M3CPR000000000000000007N', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000003B', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR000000000000000007P', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000003B', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000007Q', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000003B', '01M3PM00000000000000000004', NOW(), NULL),
		('01M3CPR000000000000000007R', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000003C', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000007S', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000003D', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR000000000000000007T', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000003D', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000007V', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000003D', '01M3PM00000000000000000003', NOW(), NULL),
		('01M3CPR000000000000000007W', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000003D', '01M3PM00000000000000000004', NOW(), NULL),
		('01M3CPR000000000000000007X', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000003E', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR000000000000000007Y', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000003E', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000007Z', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000003E', '01M3PM00000000000000000004', NOW(), NULL),
		('01M3CPR0000000000000000080', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000003F', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR0000000000000000081', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000003H', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR0000000000000000082', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000003H', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR0000000000000000083', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000003H', '01M3PM00000000000000000004', NOW(), NULL),
		('01M3CPR0000000000000000084', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000003J', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR0000000000000000085', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000003K', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR0000000000000000086', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000003K', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR0000000000000000087', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000003K', '01M3PM00000000000000000003', NOW(), NULL),
		('01M3CPR0000000000000000088', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000003K', '01M3PM00000000000000000004', NOW(), NULL),
		('01M3CPR0000000000000000089', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000003M', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR000000000000000008A', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000003M', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000008B', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000003M', '01M3PM00000000000000000004', NOW(), NULL),
		('01M3CPR000000000000000008C', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000003N', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000008D', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000003P', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR000000000000000008E', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000003P', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000008F', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000003P', '01M3PM00000000000000000003', NOW(), NULL),
		('01M3CPR000000000000000008G', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000003P', '01M3PM00000000000000000004', NOW(), NULL),
		('01M3CPR000000000000000008H', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000003R', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000008J', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000003S', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR000000000000000008K', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000003S', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000008M', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000003S', '01M3PM00000000000000000003', NOW(), NULL),
		('01M3CPR000000000000000008N', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000003S', '01M3PM00000000000000000004', NOW(), NULL),
		('01M3CPR000000000000000008P', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000003T', '01M3PM00000000000000000001', NOW(), NULL),
		('01M3CPR000000000000000008Q', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000003T', '01M3PM00000000000000000002', NOW(), NULL),
		('01M3CPR000000000000000008R', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC0000000000000000003T', '01M3PM00000000000000000004', NOW(), NULL)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;


-- Seed data for Sales pricing: the price lists a retailer would configure, and the rules inside
-- them.
--
-- Added by the product-pricing change request, which widened a pricelist rule from "one price for
-- one variant" to a targeted, dated, calculated thing. There were no pricelists at all before this,
-- so nothing exercised the resolution ladder outside its unit tests.
--
-- The set is chosen so that every step of that ladder has something to discriminate between:
--
--   * SCOPE -- a default list, a channel-scoped one (POS), and a point-scoped one (an airport
--     kiosk). The kiosk list marks everything UP by 15%, which must beat the POS list's specific
--     rules for the same products: pricelist specificity outranks target specificity, because the
--     list in force at this till is the one that applies (see rule_match.go).
--   * TARGET -- the retail list carries a rule at each of the four targets on overlapping
--     products: a variant rule, its template rule, its own category (energy_drinks) and that
--     category's parent (beverages). Resolution must prefer them in that order, and the nearest
--     ancestor category must beat the further one (PRICE-INV-017).
--   * QUANTITY -- breaks at 1/10/100 on the POS list and 1/10/50 on wholesale, so the highest
--     break a line actually reaches is observable.
--   * METHOD -- FIXED_PRICE throughout the retail list, DISCOUNT on the online list, and FORMULA
--     on the derived list using all three base sources: BASE_SALES_PRICE, COST, and
--     OTHER_PRICELIST pointing at the default list. That last is a legitimate CHAIN, and it is
--     seeded deliberately: the cycle check refuses loops, not depth, and a seed that avoided
--     chains entirely would never prove it.
--   * DATES -- an expired 2025 promotion, a future 2027 one, and a rule on a LIVE list whose own
--     window has closed. None resolves today; all remain readable, because an order priced under
--     one still has to explain itself.
--   * ARCHIVED -- an archived list, and an archived rule on a live list. Neither may price new
--     business, both stay resolvable for history (PRICE-INV-023, PRICE-INV-024).
--   * TIE-BREAK -- two rules alike in every ranked respect except sequence.
--   * ORG ISOLATION -- the second organization has its OWN default list, which proves
--     at-most-one-default is per organization rather than global, and whose rules must never
--     resolve for the first (PRICE-INV-026, PRICE-INV-027).
--
-- Every list carries a currency. The column is nullable in the DDL, for the benefit of deployments
-- that had lists before this change request, but the domain service requires one on create -- so a
-- seeded list without it would be a row the API itself would refuse to accept.
--
-- Every insert is ON CONFLICT DO NOTHING and every id is fixed rather than generated, so running
-- this file twice changes nothing.


DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'sales_pricelists'
	) THEN
		INSERT INTO "sales_pricelists" (
			"id", "org_id", "code", "name", "description", "currency_id", "is_default",
			"sales_channel_id", "sales_point_id", "valid_from", "valid_until", "priority",
			"is_archived", "created_at", "etag"
		) VALUES
		('01M3PL000000000000000001', '01JWNY20G23KD4RV5VWYABQYHD', 'RETAIL_VND', 'Retail price list', 'The organization default: what an order falls back to when it names no list', '01KZQC0000CURRENCY000ND300', TRUE, NULL, NULL, NULL, NULL, 0, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3PL000000000000000002', '01JWNY20G23KD4RV5VWYABQYHD', 'POS_VND', 'In-store POS price list', 'Channel-scoped, so it beats the default for POS orders', '01KZQC0000CURRENCY000ND300', FALSE, '01M3SC00000000000000000003', NULL, NULL, NULL, 10, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3PL000000000000000003', '01JWNY20G23KD4RV5VWYABQYHD', 'KIOSK_VND', 'Airport kiosk price list', 'Point-scoped: the narrowest scope, and it beats its own channel list', '01KZQC0000CURRENCY000ND300', FALSE, '01M3SC00000000000000000003', '01M3SP0000000000000000001X', NULL, NULL, 10, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3PL000000000000000004', '01JWNY20G23KD4RV5VWYABQYHD', 'ECOM_VND', 'Online price list', 'Online prices, typically below the shelf price', '01KZQC0000CURRENCY000ND300', FALSE, '01M3SC0000000000000000000F', NULL, NULL, NULL, 10, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3PL000000000000000005', '01JWNY20G23KD4RV5VWYABQYHD', 'VENDING_VND', 'Vending machine price list', 'Vending carries a convenience premium', '01KZQC0000CURRENCY000ND300', FALSE, '01M3SC00000000000000000001', NULL, NULL, NULL, 10, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3PL000000000000000006', '01JWNY20G23KD4RV5VWYABQYHD', 'WHOLESALE_VND', 'Wholesale price list', 'Trade prices, driven mostly by quantity breaks', '01KZQC0000CURRENCY000ND300', FALSE, '01M3SC0000000000000000001Z', NULL, NULL, NULL, 10, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3PL000000000000000007', '01JWNY20G23KD4RV5VWYABQYHD', 'EXPORT_USD', 'Export price list (USD)', 'Quoted in USD; there is no FX service, so the currency is carried not converted', '01KZQC0000CURRENCY000MJ300', FALSE, '01M3SC0000000000000000003F', NULL, NULL, NULL, 10, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3PL000000000000000008', '01JWNY20G23KD4RV5VWYABQYHD', 'PROMO_2025', 'Expired 2025 promotion', 'Window closed: must not resolve today, but still explains 2025 orders', '01KZQC0000CURRENCY000ND300', FALSE, NULL, NULL, '2025-01-01', '2025-12-31', 20, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3PL000000000000000009', '01JWNY20G23KD4RV5VWYABQYHD', 'PROMO_2027', 'Future 2027 promotion', 'Window not yet open', '01KZQC0000CURRENCY000ND300', FALSE, NULL, NULL, '2027-01-01', NULL, 20, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3PL000000000000000010', '01JWNY20G23KD4RV5VWYABQYHD', 'RETIRED_VND', 'Retired price list', 'Archived: excluded from new resolution, still readable for history', '01KZQC0000CURRENCY000ND300', FALSE, NULL, NULL, NULL, NULL, 5, TRUE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3PL000000000000000011', '01JWNY20G23KD4RV5VWYABQYHD', 'DERIVED_VND', 'Cost-plus derived price list', 'Every rule here derives from another list or from cost', '01KZQC0000CURRENCY000ND300', FALSE, NULL, NULL, NULL, NULL, 15, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3PL000000000000000012', '01K1H7M2K9VW3P5R7XQJY2C1Z9', 'OTHERORG_VND', 'Second organization list', 'Default of the OTHER org: proves at-most-one-default is per organization', '01KZQC0000CURRENCY000ND300', TRUE, NULL, NULL, NULL, NULL, 0, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;

DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'sales_pricelist_items'
	) THEN
		INSERT INTO "sales_pricelist_items" (
			"id", "org_id", "sales_pricelist_id", "applies_to",
			"product_variant_id", "product_template_id", "product_category_id",
			"uom_id", "price", "min_quantity", "calculation_method", "discount_percent",
			"base_price_source", "base_pricelist_id", "surcharge_amount",
			"rounding_increment", "valid_from", "valid_to", "sequence",
			"is_archived", "created_at", "etag"
		) VALUES
		('01M3PI00000000000000000001', '01JWNY20G23KD4RV5VWYABQYHD', '01M3PL000000000000000001', 'PRODUCT_VARIANT', '01K5INV00000000VARIANT0001', NULL, NULL, '01K5INV000000000000UOM0001', '14000', '0', 'FIXED_PRICE', NULL, NULL, NULL, NULL, NULL, NULL, NULL, 10, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3PI00000000000000000002', '01JWNY20G23KD4RV5VWYABQYHD', '01M3PL000000000000000001', 'PRODUCT_TEMPLATE', NULL, '01K5INV0000000TMPL00000001', NULL, '01K5INV000000000000UOM0001', '14500', '0', 'FIXED_PRICE', NULL, NULL, NULL, NULL, NULL, NULL, NULL, 10, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3PI00000000000000000003', '01JWNY20G23KD4RV5VWYABQYHD', '01M3PL000000000000000001', 'PRODUCT_CATEGORY', NULL, NULL, '01K5INV000000000000CAT0004', '01K5INV000000000000UOM0001', '14800', '0', 'FIXED_PRICE', NULL, NULL, NULL, NULL, NULL, NULL, NULL, 10, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3PI00000000000000000004', '01JWNY20G23KD4RV5VWYABQYHD', '01M3PL000000000000000001', 'PRODUCT_CATEGORY', NULL, NULL, '01K5INV000000000000CAT0001', '01K5INV000000000000UOM0001', '15200', '0', 'FIXED_PRICE', NULL, NULL, NULL, NULL, NULL, NULL, NULL, 10, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3PI00000000000000000005', '01JWNY20G23KD4RV5VWYABQYHD', '01M3PL000000000000000001', 'ALL_PRODUCTS', NULL, NULL, NULL, '01K5INV000000000000UOM0001', '99000', '0', 'FIXED_PRICE', NULL, NULL, NULL, NULL, NULL, NULL, NULL, 90, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3PI00000000000000000006', '01JWNY20G23KD4RV5VWYABQYHD', '01M3PL000000000000000001', 'PRODUCT_TEMPLATE', NULL, '01K5INV0000000TMPL00000002', NULL, '01K5INV000000000000UOM0001', '7600', '0', 'FIXED_PRICE', NULL, NULL, NULL, NULL, NULL, NULL, NULL, 10, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3PI00000000000000000007', '01JWNY20G23KD4RV5VWYABQYHD', '01M3PL000000000000000001', 'PRODUCT_TEMPLATE', NULL, '01K5INV0000000TMPL00000003', NULL, '01K5INV000000000000UOM0001', '11500', '0', 'FIXED_PRICE', NULL, NULL, NULL, NULL, NULL, NULL, NULL, 10, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3PI00000000000000000008', '01JWNY20G23KD4RV5VWYABQYHD', '01M3PL000000000000000001', 'PRODUCT_CATEGORY', NULL, NULL, '01K5INV000000000000CAT0006', '01K5INV000000000000UOM0001', '11800', '0', 'FIXED_PRICE', NULL, NULL, NULL, NULL, NULL, NULL, NULL, 10, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3PI00000000000000000009', '01JWNY20G23KD4RV5VWYABQYHD', '01M3PL000000000000000001', 'PRODUCT_CATEGORY', NULL, NULL, '01K5INV000000000000CAT0002', '01K5INV000000000000UOM0001', '12200', '0', 'FIXED_PRICE', NULL, NULL, NULL, NULL, NULL, NULL, NULL, 10, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3PI00000000000000000010', '01JWNY20G23KD4RV5VWYABQYHD', '01M3PL000000000000000001', 'PRODUCT_TEMPLATE', NULL, '01K5INV0000000TMPL00000004', NULL, '01K5INV000000000000UOM0004', '92000', '0', 'FIXED_PRICE', NULL, NULL, NULL, NULL, NULL, NULL, NULL, 10, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3PI00000000000000000011', '01JWNY20G23KD4RV5VWYABQYHD', '01M3PL000000000000000001', 'PRODUCT_TEMPLATE', NULL, '01K5INV0000000TMPL00000005', NULL, '01K5INV000000000000UOM0001', '9200', '0', 'FIXED_PRICE', NULL, NULL, NULL, NULL, NULL, NULL, NULL, 10, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3PI00000000000000000012', '01JWNY20G23KD4RV5VWYABQYHD', '01M3PL000000000000000001', 'PRODUCT_TEMPLATE', NULL, '01K5INV0000000TMPL00000010', NULL, '01K5INV000000000000UOM0001', '115000', '0', 'FIXED_PRICE', NULL, NULL, NULL, NULL, NULL, NULL, NULL, 10, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3PI00000000000000000013', '01JWNY20G23KD4RV5VWYABQYHD', '01M3PL000000000000000001', 'PRODUCT_VARIANT', '01K5INV00000000VARIANT0002', NULL, NULL, '01K5INV000000000000UOM0001', '14100', '0', 'FIXED_PRICE', NULL, NULL, NULL, NULL, NULL, NULL, NULL, 5, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3PI00000000000000000014', '01JWNY20G23KD4RV5VWYABQYHD', '01M3PL000000000000000001', 'PRODUCT_VARIANT', '01K5INV00000000VARIANT0002', NULL, NULL, '01K5INV000000000000UOM0001', '14400', '0', 'FIXED_PRICE', NULL, NULL, NULL, NULL, NULL, NULL, NULL, 50, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3PI00000000000000000015', '01JWNY20G23KD4RV5VWYABQYHD', '01M3PL000000000000000002', 'PRODUCT_VARIANT', '01K5INV00000000VARIANT0001', NULL, NULL, '01K5INV000000000000UOM0001', '14000', '1', 'FIXED_PRICE', NULL, NULL, NULL, NULL, NULL, NULL, NULL, 10, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3PI00000000000000000016', '01JWNY20G23KD4RV5VWYABQYHD', '01M3PL000000000000000002', 'PRODUCT_VARIANT', '01K5INV00000000VARIANT0001', NULL, NULL, '01K5INV000000000000UOM0001', '13500', '10', 'FIXED_PRICE', NULL, NULL, NULL, NULL, NULL, NULL, NULL, 10, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3PI00000000000000000017', '01JWNY20G23KD4RV5VWYABQYHD', '01M3PL000000000000000002', 'PRODUCT_VARIANT', '01K5INV00000000VARIANT0001', NULL, NULL, '01K5INV000000000000UOM0001', '12800', '100', 'FIXED_PRICE', NULL, NULL, NULL, NULL, NULL, NULL, NULL, 10, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3PI00000000000000000018', '01JWNY20G23KD4RV5VWYABQYHD', '01M3PL000000000000000002', 'PRODUCT_TEMPLATE', NULL, '01K5INV0000000TMPL00000003', NULL, '01K5INV000000000000UOM0001', '11400', '0', 'FIXED_PRICE', NULL, NULL, NULL, NULL, NULL, NULL, NULL, 10, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3PI00000000000000000019', '01JWNY20G23KD4RV5VWYABQYHD', '01M3PL000000000000000002', 'ALL_PRODUCTS', NULL, NULL, NULL, NULL, NULL, '0', 'DISCOUNT', '2', NULL, NULL, NULL, NULL, NULL, NULL, 80, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3PI00000000000000000020', '01JWNY20G23KD4RV5VWYABQYHD', '01M3PL000000000000000003', 'ALL_PRODUCTS', NULL, NULL, NULL, NULL, NULL, '0', 'DISCOUNT', '-15', NULL, NULL, NULL, NULL, NULL, NULL, 10, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3PI00000000000000000021', '01JWNY20G23KD4RV5VWYABQYHD', '01M3PL000000000000000003', 'PRODUCT_TEMPLATE', NULL, '01K5INV0000000TMPL00000002', NULL, '01K5INV000000000000UOM0001', '12000', '0', 'FIXED_PRICE', NULL, NULL, NULL, NULL, NULL, NULL, NULL, 10, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3PI00000000000000000022', '01JWNY20G23KD4RV5VWYABQYHD', '01M3PL000000000000000004', 'ALL_PRODUCTS', NULL, NULL, NULL, NULL, NULL, '0', 'DISCOUNT', '5', NULL, NULL, NULL, NULL, NULL, NULL, 50, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3PI00000000000000000023', '01JWNY20G23KD4RV5VWYABQYHD', '01M3PL000000000000000004', 'PRODUCT_CATEGORY', NULL, NULL, '01K5INV000000000000CAT0001', NULL, NULL, '0', 'DISCOUNT', '8', NULL, NULL, NULL, NULL, NULL, NULL, 20, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3PI00000000000000000024', '01JWNY20G23KD4RV5VWYABQYHD', '01M3PL000000000000000004', 'PRODUCT_TEMPLATE', NULL, '01K5INV0000000TMPL00000010', NULL, NULL, NULL, '0', 'DISCOUNT', '12', NULL, NULL, NULL, NULL, NULL, NULL, 10, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3PI00000000000000000025', '01JWNY20G23KD4RV5VWYABQYHD', '01M3PL000000000000000005', 'ALL_PRODUCTS', NULL, NULL, NULL, NULL, NULL, '0', 'FORMULA', '-20', 'BASE_SALES_PRICE', NULL, NULL, '1000', NULL, NULL, 10, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3PI00000000000000000026', '01JWNY20G23KD4RV5VWYABQYHD', '01M3PL000000000000000005', 'PRODUCT_TEMPLATE', NULL, '01K5INV0000000TMPL00000001', NULL, '01K5INV000000000000UOM0001', '18000', '0', 'FIXED_PRICE', NULL, NULL, NULL, NULL, NULL, NULL, NULL, 5, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3PI00000000000000000027', '01JWNY20G23KD4RV5VWYABQYHD', '01M3PL000000000000000006', 'PRODUCT_TEMPLATE', NULL, '01K5INV0000000TMPL00000001', NULL, '01K5INV000000000000UOM0003', '310000', '1', 'FIXED_PRICE', NULL, NULL, NULL, NULL, NULL, NULL, NULL, 10, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3PI00000000000000000028', '01JWNY20G23KD4RV5VWYABQYHD', '01M3PL000000000000000006', 'PRODUCT_TEMPLATE', NULL, '01K5INV0000000TMPL00000001', NULL, '01K5INV000000000000UOM0003', '295000', '10', 'FIXED_PRICE', NULL, NULL, NULL, NULL, NULL, NULL, NULL, 10, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3PI00000000000000000029', '01JWNY20G23KD4RV5VWYABQYHD', '01M3PL000000000000000006', 'PRODUCT_TEMPLATE', NULL, '01K5INV0000000TMPL00000001', NULL, '01K5INV000000000000UOM0003', '280000', '50', 'FIXED_PRICE', NULL, NULL, NULL, NULL, NULL, NULL, NULL, 10, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3PI00000000000000000030', '01JWNY20G23KD4RV5VWYABQYHD', '01M3PL000000000000000006', 'PRODUCT_TEMPLATE', NULL, '01K5INV0000000TMPL00000002', NULL, '01K5INV000000000000UOM0003', '168000', '1', 'FIXED_PRICE', NULL, NULL, NULL, NULL, NULL, NULL, NULL, 10, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3PI00000000000000000031', '01JWNY20G23KD4RV5VWYABQYHD', '01M3PL000000000000000006', 'PRODUCT_TEMPLATE', NULL, '01K5INV0000000TMPL00000002', NULL, '01K5INV000000000000UOM0003', '158000', '20', 'FIXED_PRICE', NULL, NULL, NULL, NULL, NULL, NULL, NULL, 10, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3PI00000000000000000032', '01JWNY20G23KD4RV5VWYABQYHD', '01M3PL000000000000000006', 'PRODUCT_TEMPLATE', NULL, '01K5INV0000000TMPL00000003', NULL, '01K5INV000000000000UOM0003', '252000', '1', 'FIXED_PRICE', NULL, NULL, NULL, NULL, NULL, NULL, NULL, 10, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3PI00000000000000000033', '01JWNY20G23KD4RV5VWYABQYHD', '01M3PL000000000000000006', 'PRODUCT_TEMPLATE', NULL, '01K5INV0000000TMPL00000004', NULL, '01K5INV000000000000UOM0004', '86000', '25', 'FIXED_PRICE', NULL, NULL, NULL, NULL, NULL, NULL, NULL, 10, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3PI00000000000000000034', '01JWNY20G23KD4RV5VWYABQYHD', '01M3PL000000000000000006', 'ALL_PRODUCTS', NULL, NULL, NULL, NULL, NULL, '0', 'DISCOUNT', '10', NULL, NULL, NULL, NULL, NULL, NULL, 90, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3PI00000000000000000035', '01JWNY20G23KD4RV5VWYABQYHD', '01M3PL000000000000000007', 'PRODUCT_TEMPLATE', NULL, '01K5INV0000000TMPL00000010', NULL, '01K5INV000000000000UOM0001', '5.40', '0', 'FIXED_PRICE', NULL, NULL, NULL, NULL, NULL, NULL, NULL, 10, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3PI00000000000000000036', '01JWNY20G23KD4RV5VWYABQYHD', '01M3PL000000000000000007', 'PRODUCT_TEMPLATE', NULL, '01K5INV0000000TMPL00000001', NULL, '01K5INV000000000000UOM0001', '0.72', '0', 'FIXED_PRICE', NULL, NULL, NULL, NULL, NULL, NULL, NULL, 10, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3PI00000000000000000037', '01JWNY20G23KD4RV5VWYABQYHD', '01M3PL000000000000000007', 'ALL_PRODUCTS', NULL, NULL, NULL, NULL, NULL, '0', 'DISCOUNT', '3', NULL, NULL, NULL, NULL, NULL, NULL, 80, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3PI00000000000000000038', '01JWNY20G23KD4RV5VWYABQYHD', '01M3PL000000000000000008', 'PRODUCT_TEMPLATE', NULL, '01K5INV0000000TMPL00000001', NULL, '01K5INV000000000000UOM0001', '11000', '0', 'FIXED_PRICE', NULL, NULL, NULL, NULL, NULL, '2025-06-01', '2025-08-31', 10, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3PI00000000000000000039', '01JWNY20G23KD4RV5VWYABQYHD', '01M3PL000000000000000008', 'PRODUCT_TEMPLATE', NULL, '01K5INV0000000TMPL00000003', NULL, '01K5INV000000000000UOM0001', '9000', '0', 'FIXED_PRICE', NULL, NULL, NULL, NULL, NULL, '2025-06-01', '2025-08-31', 10, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3PI00000000000000000040', '01JWNY20G23KD4RV5VWYABQYHD', '01M3PL000000000000000009', 'PRODUCT_TEMPLATE', NULL, '01K5INV0000000TMPL00000001', NULL, '01K5INV000000000000UOM0001', '16000', '0', 'FIXED_PRICE', NULL, NULL, NULL, NULL, NULL, '2027-01-01', NULL, 10, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3PI00000000000000000041', '01JWNY20G23KD4RV5VWYABQYHD', '01M3PL000000000000000001', 'PRODUCT_TEMPLATE', NULL, '01K5INV0000000TMPL00000005', NULL, '01K5INV000000000000UOM0001', '8000', '0', 'FIXED_PRICE', NULL, NULL, NULL, NULL, NULL, '2025-01-01', '2025-12-31', 1, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3PI00000000000000000042', '01JWNY20G23KD4RV5VWYABQYHD', '01M3PL000000000000000001', 'PRODUCT_TEMPLATE', NULL, '01K5INV0000000TMPL00000006', NULL, '01K5INV000000000000UOM0001', '3900', '0', 'FIXED_PRICE', NULL, NULL, NULL, NULL, NULL, NULL, NULL, 10, TRUE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3PI00000000000000000043', '01JWNY20G23KD4RV5VWYABQYHD', '01M3PL000000000000000010', 'PRODUCT_TEMPLATE', NULL, '01K5INV0000000TMPL00000001', NULL, '01K5INV000000000000UOM0001', '13000', '0', 'FIXED_PRICE', NULL, NULL, NULL, NULL, NULL, NULL, NULL, 10, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3PI00000000000000000044', '01JWNY20G23KD4RV5VWYABQYHD', '01M3PL000000000000000011', 'PRODUCT_TEMPLATE', NULL, '01K5INV0000000TMPL00000004', NULL, NULL, NULL, '0', 'FORMULA', '-40', 'COST', NULL, NULL, NULL, NULL, NULL, 10, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3PI00000000000000000045', '01JWNY20G23KD4RV5VWYABQYHD', '01M3PL000000000000000011', 'PRODUCT_TEMPLATE', NULL, '01K5INV0000000TMPL00000001', NULL, NULL, NULL, '0', 'FORMULA', '5', 'OTHER_PRICELIST', '01M3PL000000000000000001', NULL, NULL, NULL, NULL, 10, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3PI00000000000000000046', '01JWNY20G23KD4RV5VWYABQYHD', '01M3PL000000000000000011', 'PRODUCT_TEMPLATE', NULL, '01K5INV0000000TMPL00000003', NULL, NULL, NULL, '0', 'FORMULA', '-25', 'BASE_SALES_PRICE', NULL, '500', '500', NULL, NULL, 10, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3PI00000000000000000047', '01JWNY20G23KD4RV5VWYABQYHD', '01M3PL000000000000000011', 'ALL_PRODUCTS', NULL, NULL, NULL, NULL, NULL, '0', 'FORMULA', '-30', 'COST', NULL, NULL, '1000', NULL, NULL, 90, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3PI00000000000000000048', '01K1H7M2K9VW3P5R7XQJY2C1Z9', '01M3PL000000000000000012', 'PRODUCT_TEMPLATE', NULL, '01K5INV0000000TMPL00000001', NULL, '01K5INV000000000000UOM0001', '13300', '0', 'FIXED_PRICE', NULL, NULL, NULL, NULL, NULL, NULL, NULL, 10, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3PI00000000000000000049', '01K1H7M2K9VW3P5R7XQJY2C1Z9', '01M3PL000000000000000012', 'ALL_PRODUCTS', NULL, NULL, NULL, NULL, NULL, '0', 'DISCOUNT', '4', NULL, NULL, NULL, NULL, NULL, NULL, 80, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;
