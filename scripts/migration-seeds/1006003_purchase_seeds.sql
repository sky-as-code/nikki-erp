-- Seed data for the Purchase module: what each vendor currently offers, at what quantity, in what
-- unit and until when.
--
-- The module had no seed file before this; it is added by the product-pricing change request, which
-- introduced the resource these rows populate.
--
-- The set is chosen to cover every shape ResolvePurchasePrice discriminates between, so that the
-- resolver can be exercised against real data rather than only against its unit tests:
--
--   * QUANTITY BREAKS -- Energy Drink from Saigon Beverage at 1+/10+/100+ (250,000 / 240,000 /
--     220,000). This is the change request's own TS-PRICE-06 worked example: a line of 120 must
--     resolve to 220,000, taking the highest break it actually reaches rather than the lowest or
--     the nearest.
--   * VARIANT BEATS TEMPLATE -- the same vendor also quotes one variant of that product
--     specifically, which must win over its own template-wide rows (PRICE-INV-018).
--   * COMPETING VENDORS -- three suppliers quote Energy Drink. Resolution filters by vendor first,
--     so the cheapest is NOT automatically the answer; automatic vendor selection is explicitly out
--     of scope (section 44.5), and the cheapest row here deliberately carries a 21-day lead time to
--     show why.
--   * UNIT CONVERSION -- Mineral Water is quoted by the dozen, by the case and by the piece, all in
--     the Unit category. Ground Coffee is stocked by weight and quoted per kilogram and per gram.
--     Both exercise the conversion path; a quote in the wrong CATEGORY is what the validator
--     refuses, and is therefore deliberately absent.
--   * FOREIGN CURRENCY -- the USB-C cable is quoted in USD. There is no FX service, so a purchase
--     order in another currency cannot reconcile this automatically. That is precisely why the
--     currency is recorded per price rather than assumed (BR-PRICE-CUR-002).
--   * EFFECTIVE DATING -- an expired 2025 price and a future 2027 one, neither of which resolves
--     today, both of which remain readable so an order placed under one can still explain itself.
--   * ARCHIVED -- two rows from a retired supplier: excluded from new resolution, still resolvable
--     for history (PRICE-INV-023, PRICE-INV-024).
--   * UNORDERABLE VENDORS -- prices from blacklisted, proposed and suspended suppliers. The rows
--     are real records of what was quoted; whether an order may use one is the vendor's status,
--     which the validator asks Contacts about.
--   * TIE-BREAK -- two rows identical in every ranked respect except sequence.
--   * CROSS-ORG -- one row belonging to the second organization, which must never resolve for the
--     first however closely it matches (PRICE-INV-027).
--
-- Prices are in the vendor's own currency and are NOT the product's cost. A product costing 10,200
-- may be offered at 10,000 by one supplier and 9,500 by another; cost is Inventory's to calculate
-- from goods actually received (PRICE-INV-004, PRICE-INV-009).
--
-- Every insert is ON CONFLICT DO NOTHING and every id is fixed rather than generated, so running
-- this file twice changes nothing.


DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'purchase_vendor_product_prices'
	) THEN
		INSERT INTO "purchase_vendor_product_prices" (
			"id", "org_id", "vendor_id", "product_template_id", "product_variant_id",
			"purchase_uom_id", "currency_id", "min_quantity", "unit_price",
			"valid_from", "valid_to", "lead_time_days", "sequence",
			"vendor_product_code", "vendor_product_name",
			"is_archived", "created_at", "etag"
		) VALUES
		('01K6PUR000000000000VPP0001', '01JWNY20G23KD4RV5VWYABQYHD', '01K5CNT000000000VENDOR0001', '01K5INV0000000TMPL00000001', NULL, '01K5INV000000000000UOM0003', '01KZQC0000CURRENCY000ND300', 1, 250000, NULL, NULL, 3, 10, 'SBC-ED-C24', 'Thùng 24 lon nước tăng lực', FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K6PUR000000000000VPP0002', '01JWNY20G23KD4RV5VWYABQYHD', '01K5CNT000000000VENDOR0001', '01K5INV0000000TMPL00000001', NULL, '01K5INV000000000000UOM0003', '01KZQC0000CURRENCY000ND300', 10, 240000, NULL, NULL, 3, 10, 'SBC-ED-C24', 'Thùng 24 lon nước tăng lực', FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K6PUR000000000000VPP0003', '01JWNY20G23KD4RV5VWYABQYHD', '01K5CNT000000000VENDOR0001', '01K5INV0000000TMPL00000001', NULL, '01K5INV000000000000UOM0003', '01KZQC0000CURRENCY000ND300', 100, 220000, NULL, NULL, 3, 10, 'SBC-ED-C24', 'Thùng 24 lon nước tăng lực', FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K6PUR000000000000VPP0004', '01JWNY20G23KD4RV5VWYABQYHD', '01K5CNT000000000VENDOR0001', '01K5INV0000000TMPL00000001', '01K5INV00000000VARIANT0001', '01K5INV000000000000UOM0003', '01KZQC0000CURRENCY000ND300', 1, 235000, NULL, NULL, 3, 10, 'SBC-ED-BERRY', 'Thùng 24 lon nước tăng lực vị dâu', FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K6PUR000000000000VPP0005', '01JWNY20G23KD4RV5VWYABQYHD', '01K5CNT000000000VENDOR0009', '01K5INV0000000TMPL00000001', NULL, '01K5INV000000000000UOM0003', '01KZQC0000CURRENCY000ND300', 1, 255000, NULL, NULL, 4, 10, 'NSP-ED-24', 'Thùng nước tăng lực 24 lon', FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K6PUR000000000000VPP0006', '01JWNY20G23KD4RV5VWYABQYHD', '01K5CNT000000000VENDOR0009', '01K5INV0000000TMPL00000001', NULL, '01K5INV000000000000UOM0003', '01KZQC0000CURRENCY000ND300', 50, 242000, NULL, NULL, 4, 10, 'NSP-ED-24', 'Thùng nước tăng lực 24 lon', FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K6PUR000000000000VPP0007', '01JWNY20G23KD4RV5VWYABQYHD', '01K5CNT000000000VENDOR0008', '01K5INV0000000TMPL00000001', NULL, '01K5INV000000000000UOM0003', '01KZQC0000CURRENCY000ND300', 200, 198000, NULL, NULL, 21, 20, 'BW-ED-BULK', 'Giá pallet nước tăng lực số lượng lớn', FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K6PUR000000000000VPP0008', '01JWNY20G23KD4RV5VWYABQYHD', '01K5CNT000000000VENDOR0001', '01K5INV0000000TMPL00000002', NULL, '01K5INV000000000000UOM0002', '01KZQC0000CURRENCY000ND300', 1, 84000, NULL, NULL, 3, 10, 'SBC-MW-DZ', 'Nước khoáng lốc 12 chai', FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K6PUR000000000000VPP0009', '01JWNY20G23KD4RV5VWYABQYHD', '01K5CNT000000000VENDOR0001', '01K5INV0000000TMPL00000002', NULL, '01K5INV000000000000UOM0001', '01KZQC0000CURRENCY000ND300', 1, 7200, NULL, NULL, 3, 20, 'SBC-MW-PC', 'Nước khoáng chai lẻ', FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K6PUR000000000000VPP0010', '01JWNY20G23KD4RV5VWYABQYHD', '01K5CNT000000000VENDOR0009', '01K5INV0000000TMPL00000002', NULL, '01K5INV000000000000UOM0003', '01KZQC0000CURRENCY000ND300', 5, 160000, NULL, NULL, 4, 10, 'NSP-MW-C24', 'Thùng nước 24 chai', FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K6PUR000000000000VPP0011', '01JWNY20G23KD4RV5VWYABQYHD', '01K5CNT000000000VENDOR0002', '01K5INV0000000TMPL00000003', NULL, '01K5INV000000000000UOM0003', '01KZQC0000CURRENCY000ND300', 1, 260000, NULL, NULL, 5, 10, 'MKF-PC-24', 'Hộp trưng bày khoai tây chiên', FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K6PUR000000000000VPP0012', '01JWNY20G23KD4RV5VWYABQYHD', '01K5CNT000000000VENDOR0002', '01K5INV0000000TMPL00000003', NULL, '01K5INV000000000000UOM0003', '01KZQC0000CURRENCY000ND300', 20, 248000, NULL, NULL, 5, 10, 'MKF-PC-24', 'Hộp trưng bày khoai tây chiên', FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K6PUR000000000000VPP0013', '01JWNY20G23KD4RV5VWYABQYHD', '01K5CNT000000000VENDOR0002', '01K5INV0000000TMPL00000003', '01K5INV00000000VARIANT0004', '01K5INV000000000000UOM0001', '01KZQC0000CURRENCY000ND300', 1, 11000, NULL, NULL, 5, 10, 'MKF-PC-SALT', 'Khoai tây chiên muối gói lẻ', FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K6PUR000000000000VPP0014', '01JWNY20G23KD4RV5VWYABQYHD', '01K5CNT000000000VENDOR0003', '01K5INV0000000TMPL00000004', NULL, '01K5INV000000000000UOM0004', '01KZQC0000CURRENCY000ND300', 1, 92000, NULL, NULL, 7, 10, 'HCT-GC-KG', 'Cà phê xay theo kg', FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K6PUR000000000000VPP0015', '01JWNY20G23KD4RV5VWYABQYHD', '01K5CNT000000000VENDOR0003', '01K5INV0000000TMPL00000004', NULL, '01K5INV000000000000UOM0004', '01KZQC0000CURRENCY000ND300', 25, 88000, NULL, NULL, 7, 10, 'HCT-GC-KG', 'Cà phê xay theo kg', FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K6PUR000000000000VPP0016', '01JWNY20G23KD4RV5VWYABQYHD', '01K5CNT000000000VENDOR0003', '01K5INV0000000TMPL00000004', NULL, '01K5INV000000000000UOM0004', '01KZQC0000CURRENCY000ND300', 100, 84000, NULL, NULL, 14, 10, 'HCT-GC-BULK', 'Bao cà phê xay số lượng lớn', FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K6PUR000000000000VPP0017', '01JWNY20G23KD4RV5VWYABQYHD', '01K5CNT000000000VENDOR0003', '01K5INV0000000TMPL00000004', NULL, '01K5INV000000000000UOM0005', '01KZQC0000CURRENCY000ND300', 500, 96, NULL, NULL, 7, 30, 'HCT-GC-G', 'Cà phê xay theo gram', FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K6PUR000000000000VPP0018', '01JWNY20G23KD4RV5VWYABQYHD', '01K5CNT000000000VENDOR0004', '01K5INV0000000TMPL00000005', NULL, '01K5INV000000000000UOM0003', '01KZQC0000CURRENCY000ND300', 1, 215000, NULL, NULL, 1, 10, 'DD-FM-24', 'Két sữa tươi', FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K6PUR000000000000VPP0019', '01JWNY20G23KD4RV5VWYABQYHD', '01K5CNT000000000VENDOR0004', '01K5INV0000000TMPL00000005', NULL, '01K5INV000000000000UOM0003', '01KZQC0000CURRENCY000ND300', 10, 208000, NULL, NULL, 1, 10, 'DD-FM-24', 'Két sữa tươi', FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K6PUR000000000000VPP0020', '01JWNY20G23KD4RV5VWYABQYHD', '01K5CNT000000000VENDOR0005', '01K5INV0000000TMPL00000006', NULL, '01K5INV000000000000UOM0001', '01KZQC0000CURRENCY000ND300', 1, 4200, NULL, NULL, 10, 10, 'PR-BOX-STD', 'Thùng carton vận chuyển tiêu chuẩn', FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K6PUR000000000000VPP0021', '01JWNY20G23KD4RV5VWYABQYHD', '01K5CNT000000000VENDOR0005', '01K5INV0000000TMPL00000006', NULL, '01K5INV000000000000UOM0001', '01KZQC0000CURRENCY000ND300', 1000, 3400, NULL, NULL, 10, 10, 'PR-BOX-STD', 'Thùng carton vận chuyển tiêu chuẩn', FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K6PUR000000000000VPP0022', '01JWNY20G23KD4RV5VWYABQYHD', '01K5CNT000000000VENDOR0006', '01K5INV0000000TMPL00000010', NULL, '01K5INV000000000000UOM0001', '01KZQC0000CURRENCY000MJ300', 1, 3.20, NULL, NULL, 30, 10, 'SZC-USBC-1M', 'Cáp USB-C 1m', FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K6PUR000000000000VPP0023', '01JWNY20G23KD4RV5VWYABQYHD', '01K5CNT000000000VENDOR0006', '01K5INV0000000TMPL00000010', NULL, '01K5INV000000000000UOM0001', '01KZQC0000CURRENCY000MJ300', 500, 2.65, NULL, NULL, 30, 10, 'SZC-USBC-1M', 'Cáp USB-C 1m', FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K6PUR000000000000VPP0024', '01JWNY20G23KD4RV5VWYABQYHD', '01K5CNT000000000VENDOR0006', '01K5INV0000000TMPL00000010', NULL, '01K5INV000000000000UOM0001', '01KZQC0000CURRENCY000MJ300', 5000, 2.10, NULL, NULL, 45, 10, 'SZC-USBC-1M', 'Cáp USB-C 1m', FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K6PUR000000000000VPP0025', '01JWNY20G23KD4RV5VWYABQYHD', '01K5CNT000000000VENDOR0001', '01K5INV0000000TMPL00000001', NULL, '01K5INV000000000000UOM0003', '01KZQC0000CURRENCY000ND300', 1, 265000, '2025-01-01', '2025-12-31', 3, 10, 'SBC-ED-C24', 'Thùng 24 lon nước tăng lực (2025)', FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K6PUR000000000000VPP0026', '01JWNY20G23KD4RV5VWYABQYHD', '01K5CNT000000000VENDOR0001', '01K5INV0000000TMPL00000001', NULL, '01K5INV000000000000UOM0003', '01KZQC0000CURRENCY000ND300', 1, 272000, '2027-01-01', NULL, 3, 5, 'SBC-ED-C24', 'Thùng 24 lon nước tăng lực (2027)', FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K6PUR000000000000VPP0027', '01JWNY20G23KD4RV5VWYABQYHD', '01K5CNT000000000VENDOR0010', '01K5INV0000000TMPL00000001', NULL, '01K5INV000000000000UOM0003', '01KZQC0000CURRENCY000ND300', 1, 244000, NULL, NULL, 14, 10, 'RS-ED-24', 'Thùng nước tăng lực (nhà cung cấp đã ngưng)', TRUE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K6PUR000000000000VPP0028', '01JWNY20G23KD4RV5VWYABQYHD', '01K5CNT000000000VENDOR0010', '01K5INV0000000TMPL00000003', NULL, '01K5INV000000000000UOM0003', '01KZQC0000CURRENCY000ND300', 1, 251000, NULL, NULL, 14, 10, 'RS-PC-24', 'Hộp khoai tây chiên (nhà cung cấp đã ngưng)', TRUE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K6PUR000000000000VPP0029', '01JWNY20G23KD4RV5VWYABQYHD', '01K5CNT000000000VENDOR0011', '01K5INV0000000TMPL00000003', NULL, '01K5INV000000000000UOM0003', '01KZQC0000CURRENCY000ND300', 1, 232000, NULL, NULL, 14, 10, 'BT-PC-24', 'Hộp khoai tây chiên (nhà cung cấp bị cấm)', FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K6PUR000000000000VPP0030', '01JWNY20G23KD4RV5VWYABQYHD', '01K5CNT000000000VENDOR0012', '01K5INV0000000TMPL00000002', NULL, '01K5INV000000000000UOM0003', '01KZQC0000CURRENCY000ND300', 1, 158000, NULL, NULL, 14, 10, 'PN-MW-24', 'Thùng nước (nhà cung cấp đề xuất)', FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K6PUR000000000000VPP0031', '01JWNY20G23KD4RV5VWYABQYHD', '01K5CNT000000000VENDOR0015', '01K5INV0000000TMPL00000003', NULL, '01K5INV000000000000UOM0003', '01KZQC0000CURRENCY000ND300', 1, 245000, NULL, NULL, 9, 10, 'SF-PC-24', 'Hộp khoai tây chiên (nhà cung cấp tạm ngưng)', FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K6PUR000000000000VPP0032', '01JWNY20G23KD4RV5VWYABQYHD', '01K5CNT000000000VENDOR0014', '01K5INV0000000TMPL00000005', NULL, '01K5INV000000000000UOM0001', '01KZQC0000CURRENCY000ND300', 1, 9100, NULL, NULL, 2, 10, 'TVL-MILK', 'Chai sữa tươi', FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K6PUR000000000000VPP0033', '01JWNY20G23KD4RV5VWYABQYHD', '01K5CNT000000000VENDOR0014', '01K5INV0000000TMPL00000005', NULL, '01K5INV000000000000UOM0001', '01KZQC0000CURRENCY000ND300', 1, 9300, NULL, NULL, 2, 20, 'TVL-MILK-ALT', 'Chai sữa tươi (dòng thay thế)', FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K6PUR000000000000VPP0034', '01K1H7M2K9VW3P5R7XQJY2C1Z9', '01K5CNT000000000VENDOR0013', '01K5INV0000000TMPL00000001', NULL, '01K5INV000000000000UOM0003', '01KZQC0000CURRENCY000ND300', 1, 199000, NULL, NULL, 5, 10, 'TSS-ED-24', 'Thùng nước tăng lực (tổ chức khác)', FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;


-- Purchase code numbering for every seeded organization, mirroring the Purchase org lifecycle
-- (OrgLifecycleAppServiceImpl.InitOrg): one purchase.order (PO) and one purchase.agreement (PA)
-- definition per org, their ids held in the org's "purchase" settings. No counter rows: no purchase
-- order or agreement is seeded, so new codes start at PO00001 / PA00001.
DO $$
DECLARE
	purchase_settings_schema_id varchar;
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'essential_sequence_definitions'
	) AND EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'settings_records'
	) THEN
		INSERT INTO "essential_sequence_definitions" (
			"id", "configuration_scope", "org_id", "owner_module", "code", "name", "description", "sub_scope",
			"scope_type", "format_owner", "algorithm", "algorithm_params", "prefix_template", "suffix_template", "padding",
			"increment", "initial_value", "reset_period", "timezone", "issuance_retention_days", "first_issued_at",
			"is_archived", "created_at", "updated_at", "created_by", "etag"
		) VALUES
		('01M3SQPD000000000000000011', 'org', '01JWNY20G23KD4RV5VWYABQYHD', 'purchase', 'purchase.order', 'Mã đơn mua hàng', NULL, NULL, 'org', 'sequence', 'increment', NULL, 'PO', NULL, 5, 1, 1, 'never', 'Asia/Ho_Chi_Minh', 90, NULL, false, NOW(), NULL, NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SQPD000000000000000012', 'org', '01JWNY20G23KD4RV5VWYABQYHD', 'purchase', 'purchase.agreement', 'Mã thỏa thuận mua hàng', NULL, NULL, 'org', 'sequence', 'increment', NULL, 'PA', NULL, 5, 1, 1, 'never', 'Asia/Ho_Chi_Minh', 90, NULL, false, NOW(), NULL, NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SQPD000000000000000021', 'org', '01K02G6J1CYAN9K8V4PAGSQ5Z8', 'purchase', 'purchase.order', 'Mã đơn mua hàng', NULL, NULL, 'org', 'sequence', 'increment', NULL, 'PO', NULL, 5, 1, 1, 'never', 'Asia/Ho_Chi_Minh', 90, NULL, false, NOW(), NULL, NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SQPD000000000000000022', 'org', '01K02G6J1CYAN9K8V4PAGSQ5Z8', 'purchase', 'purchase.agreement', 'Mã thỏa thuận mua hàng', NULL, NULL, 'org', 'sequence', 'increment', NULL, 'PA', NULL, 5, 1, 1, 'never', 'Asia/Ho_Chi_Minh', 90, NULL, false, NOW(), NULL, NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SQPD000000000000000031', 'org', '01K1H7M2K9VW3P5R7XQJY2C1Z9', 'purchase', 'purchase.order', 'Mã đơn mua hàng', NULL, NULL, 'org', 'sequence', 'increment', NULL, 'PO', NULL, 5, 1, 1, 'never', 'Asia/Ho_Chi_Minh', 90, NULL, false, NOW(), NULL, NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SQPD000000000000000032', 'org', '01K1H7M2K9VW3P5R7XQJY2C1Z9', 'purchase', 'purchase.agreement', 'Mã thỏa thuận mua hàng', NULL, NULL, 'org', 'sequence', 'increment', NULL, 'PA', NULL, 5, 1, 1, 'never', 'Asia/Ho_Chi_Minh', 90, NULL, false, NOW(), NULL, NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT DO NOTHING;

		-- Seeds run before the first boot registers the module's org settings, and a settings record
		-- must name its declaration. Snapshot of purchase/domain/settings.OrgSettingsSchemaBuilder()
		-- .Build().ToModelJson(); startup registration matches on (module_key, level), keeps this id
		-- and refreshes the document.
		INSERT INTO "settings_schemas" ("id", "module_key", "level", "schema", "created_at")
		VALUES ('01M3SQPS000000000000000001', 'purchase', 'org',
			$purchase_org_schema${"fields":[{"data_type":{"max":30,"min":0,"type":"string"},"description":{"$ref":"settings_desc.order_code_sequence_id"},"label":{"$ref":"purchase.settings.orderCodeSequenceId"},"metadata":{"allow_override":true},"name":"order_code_sequence_id"},{"data_type":{"max":30,"min":0,"type":"string"},"description":{"$ref":"settings_desc.agreement_code_sequence_id"},"label":{"$ref":"purchase.settings.agreementCodeSequenceId"},"metadata":{"allow_override":true},"name":"agreement_code_sequence_id"}],"label":{"$ref":"purchase.settings.orgSettings"},"name":"purchase_org_settings"}$purchase_org_schema$::jsonb, NOW())
		ON CONFLICT DO NOTHING;

		SELECT "id" INTO STRICT purchase_settings_schema_id
		FROM "settings_schemas" WHERE "module_key" = 'purchase' AND "level" = 'org';

		INSERT INTO "settings_records" (
			"id", "schema_id", "module_key", "level", "owner_type", "owner_id",
			"name", "value", "allow_override", "created_at", "updated_at"
		) VALUES
		('01M3SQPR000000000000000011', purchase_settings_schema_id, 'purchase', 'org', 'org', '01JWNY20G23KD4RV5VWYABQYHD', 'order_code_sequence_id', jsonb_build_object('value', '01M3SQPD000000000000000011'), NULL, NOW(), NULL),
		('01M3SQPR000000000000000012', purchase_settings_schema_id, 'purchase', 'org', 'org', '01JWNY20G23KD4RV5VWYABQYHD', 'agreement_code_sequence_id', jsonb_build_object('value', '01M3SQPD000000000000000012'), NULL, NOW(), NULL),
		('01M3SQPR000000000000000021', purchase_settings_schema_id, 'purchase', 'org', 'org', '01K02G6J1CYAN9K8V4PAGSQ5Z8', 'order_code_sequence_id', jsonb_build_object('value', '01M3SQPD000000000000000021'), NULL, NOW(), NULL),
		('01M3SQPR000000000000000022', purchase_settings_schema_id, 'purchase', 'org', 'org', '01K02G6J1CYAN9K8V4PAGSQ5Z8', 'agreement_code_sequence_id', jsonb_build_object('value', '01M3SQPD000000000000000022'), NULL, NOW(), NULL),
		('01M3SQPR000000000000000031', purchase_settings_schema_id, 'purchase', 'org', 'org', '01K1H7M2K9VW3P5R7XQJY2C1Z9', 'order_code_sequence_id', jsonb_build_object('value', '01M3SQPD000000000000000031'), NULL, NOW(), NULL),
		('01M3SQPR000000000000000032', purchase_settings_schema_id, 'purchase', 'org', 'org', '01K1H7M2K9VW3P5R7XQJY2C1Z9', 'agreement_code_sequence_id', jsonb_build_object('value', '01M3SQPD000000000000000032'), NULL, NOW(), NULL)
		ON CONFLICT DO NOTHING;
	END IF;
END $$;
