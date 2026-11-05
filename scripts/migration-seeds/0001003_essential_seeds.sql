-- IAM resources, actions and default entitlements for Unit of Measure (BR-UOM-ESS §13).
--
-- The dynamic resource engine asserts permissions using the schema name as the resource code,
-- so these codes must stay identical to the "essential_uom" / "essential_uomcat" schema names.

DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_resources'
	) THEN
		INSERT INTO "iam_resources" (
			"id", "name", "code", "description", "owner_type", "max_scope", "min_scope", "created_at", "etag"
		) VALUES
		('01KZJ8Q3ZK7XN4WVBM2PDHR5T0', jsonb_build_object('en-US', 'Essential - Unit of Measure', 'vi-VN', 'Thiết yếu - Đơn vị tính'), 'essential_uom', 'Đơn vị tính và hệ số quy đổi của chúng', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZJ8Q3ZKB8F1TC6RJWNQ4XS2', jsonb_build_object('en-US', 'Essential - UoM Category', 'vi-VN', 'Thiết yếu - Nhóm đơn vị tính'), 'essential_uomcat', 'Nhóm giới hạn phạm vi quy đổi đơn vị tính', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text);
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_actions'
	) THEN
		INSERT INTO "iam_actions" ("id", "name", "code", "description", "resource_id", "etag") VALUES
		-- Unit of Measure
		('01KZJ8Q3ZKC2M9YHTV5KFAP7W3', jsonb_build_object('en-US', 'Create', 'vi-VN', 'Tạo'), 'create', NULL, '01KZJ8Q3ZK7XN4WVBM2PDHR5T0', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZJ8Q3ZKD5R0XQGN8JBSE1Y4', jsonb_build_object('en-US', 'Update', 'vi-VN', 'Cập nhật'), 'update', NULL, '01KZJ8Q3ZK7XN4WVBM2PDHR5T0', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZJ8Q3ZKE7S3ZPKW1MCTF6Z5', jsonb_build_object('en-US', 'Delete', 'vi-VN', 'Xóa'), 'delete', NULL, '01KZJ8Q3ZK7XN4WVBM2PDHR5T0', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZJ8Q3ZKF9T5A1NX3PDVG8B6', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01KZJ8Q3ZK7XN4WVBM2PDHR5T0', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZJ8Q3ZKG1V7B2QY4RFWH9C7', jsonb_build_object('en-US', 'Set archived status', 'vi-VN', 'Đặt trạng thái lưu trữ'), 'set_archived', 'Lưu trữ một đơn vị tính để giao dịch mới không thể dùng nó nữa', '01KZJ8Q3ZK7XN4WVBM2PDHR5T0', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- UoM Category
		('01KZJ8Q3ZKH3W9C4RZ6SGXJ0D8', jsonb_build_object('en-US', 'Create', 'vi-VN', 'Tạo'), 'create', NULL, '01KZJ8Q3ZKB8F1TC6RJWNQ4XS2', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZJ8Q3ZKJ5X1D6S08THYK2E9', jsonb_build_object('en-US', 'Update', 'vi-VN', 'Cập nhật'), 'update', NULL, '01KZJ8Q3ZKB8F1TC6RJWNQ4XS2', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZJ8Q3ZKK7Y3E8T1AVJZL4F0', jsonb_build_object('en-US', 'Delete', 'vi-VN', 'Xóa'), 'delete', NULL, '01KZJ8Q3ZKB8F1TC6RJWNQ4XS2', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZJ8Q3ZKM9Z5F0V3BWK1M6G1', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01KZJ8Q3ZKB8F1TC6RJWNQ4XS2', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text);
	END IF;
END $$;

-- ---------------------------------------------------------------------------
-- Merged from 0001004_essential_currency_seeds.sql
-- ---------------------------------------------------------------------------

-- The currencies the product trades in: US dollar and Vietnamese dong.
--
-- essential_currencies ships empty, so without this seed no module that records money can create a
-- record at all: a purchase order, an invoice and a payment all carry a currency_id that must
-- resolve. That is why this is a migration rather than an operator task.
--
-- Only the US and Vietnamese entries are seeded, by product decision: the product serves those two
-- markets, and an administrator adds any other currency through the Currency page. Each row is
-- complete -- symbol, where it is written, unit names -- so no amount renders as a bare code.
--
-- rounding is the smallest settled amount and decimal_places is derived from it (1 -> 0 for VND,
-- 0.01 -> 2 for USD); Essential's currency service rounds every computed total to it. Whether an
-- organization uses a currency is its essential_org_currencies row, written when the org is created.
--
-- Ids are deterministic, derived from the alphabetic code, so this file is safe to re-run and the
-- same currency carries the same id in every deployment. ON CONFLICT DO NOTHING makes re-running a
-- no-op rather than an error.
--
-- These seed values are meant to stay identical between this tree and nikkierp's, aside from the
-- "tenant_id" column that the hosted app's apptrait adds to every table.
DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'essential_currencies'
	) THEN
		INSERT INTO "essential_currencies" (
			"id", "code", "numeric_code", "name", "symbol", "symbol_position", "currency_unit_label", "currency_subunit_label", "rounding", "decimal_places", "is_archived", "created_at", "etag"
		) VALUES
		('01KZQC0000CURRENCY000MJ300', 'USD', '840', jsonb_build_object('en-US', 'US Dollar', 'vi-VN', 'Đô la Mỹ'), '$', 'before', jsonb_build_object('en-US', 'Dollar', 'vi-VN', 'Đô la'), jsonb_build_object('en-US', 'Cent', 'vi-VN', 'Xu'), 0.01, 2, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000ND300', 'VND', '704', jsonb_build_object('en-US', 'Vietnamese Dong', 'vi-VN', 'Đồng Việt Nam'), '₫', 'after', jsonb_build_object('en-US', 'Dong', 'vi-VN', 'Đồng'), jsonb_build_object('en-US', 'Xu', 'vi-VN', 'Xu'), 1, 0, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT DO NOTHING;
	END IF;
END $$;

-- ---------------------------------------------------------------------------
-- The languages the application ships translations for.
-- ---------------------------------------------------------------------------

-- essential_languages ships empty, which is why every consumer that needed a language's formatting
-- rules carried its own literal copy of them -- iam's user_rest.languageOf, chiefly. These rows are
-- what lets that copy go away: the user-context endpoint resolves the acting user's language
-- against this table instead of a switch statement that had to be edited in step with
-- SupportedLanguages, and silently disagreed with it.
--
-- The Vietnamese separators here are deliberately NOT the ones the old literal carried. It had
-- decimal '.' and thousands ',', which is the English convention; Vietnamese is the other way round
-- (1.234.567). VND is quoted in whole units, so the thousands separator is the only one most users
-- ever see, and the old value made every amount on a Vietnamese screen wrong.
--
-- Ids are fixed and readable rather than generated, matching the currency seed above, so the same
-- language carries the same id in every deployment. ON CONFLICT DO NOTHING makes a re-run a no-op.

DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'essential_languages'
	) THEN
		INSERT INTO "essential_languages" (
			"id", "name", "locale_code", "direction", "decimal_separator", "thousands_separator", "grouping", "date_format", "time_format", "short_time_format", "first_day_of_week", "public_use", "is_archived", "created_at", "etag"
		) VALUES
		('01KZQC0000LANGUAGE0000EN00', 'English (US)', 'en-US', 'ltr', '.', ',', '[3,0]', '%m/%d/%Y', '%H:%M:%S', '%H:%M', 'sunday', TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000LANGUAGE0000VI00', 'Vietnamese / Tiếng Việt', 'vi-VN', 'ltr', ',', '.', '[3,0]', '%d/%m/%Y', '%H:%M:%S', '%H:%M', 'monday', TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT DO NOTHING;
	END IF;
END $$;

-- ---------------------------------------------------------------------------
-- Rename the org setting system_locale -> system_language.
-- ---------------------------------------------------------------------------

-- The stored value lives in settings_records.name, so renaming the Go constant and the JSON schema
-- is not enough on its own: an organization that had already chosen a language would keep a row
-- under a name the schema no longer declares, and silently read as unset.
--
-- The DELETE guards the (module_key, name, owner_id) unique key. An owner holding both names can
-- only have got there by a partial run of this migration, and the new name is the one to keep --
-- without this, the UPDATE would abort the whole migration on a constraint violation.

DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'settings_records'
	) THEN
		DELETE FROM "settings_records" AS stale
		WHERE stale."module_key" = 'essential' AND stale."name" = 'system_locale'
			AND EXISTS (
				SELECT 1 FROM "settings_records" AS renamed
				WHERE renamed."module_key" = 'essential' AND renamed."name" = 'system_language'
					AND renamed."owner_id" = stale."owner_id"
			);

		UPDATE "settings_records"
		SET "name" = 'system_language'
		WHERE "module_key" = 'essential' AND "name" = 'system_locale';
	END IF;
END $$;

-- Unit of Measure catalogue.
--
-- Ids 01K5INV...UOM0001-0005 are referenced by the inventory, purchase and sales seeds and must keep
-- their meaning (piece, dozen, carton, kg, g). The carton keeps factor 24 because the purchase and
-- sales seed lines are priced per carton of 24; changing the factor would silently change their cost.
-- Factors are exact at the 12-digit scale of essential_uoms.factor (1/60 for a minute, the
-- international yard-and-pound definitions for imperial units).
DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'essential_uomcats'
	) THEN
		INSERT INTO "essential_uomcats" ("id", "name", "reference_uom_id", "org_id", "is_archived", "created_at", "updated_at", "etag") VALUES
		('01K5INV00000000000UOMCAT01', jsonb_build_object('en-US', 'Unit', 'vi-VN', 'Đơn vị'), '01K5INV000000000000UOM0001', '01JWNY20G23KD4RV5VWYABQYHD', FALSE, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5INV00000000000UOMCAT02', jsonb_build_object('en-US', 'Weight', 'vi-VN', 'Khối lượng'), '01K5INV000000000000UOM0004', '01JWNY20G23KD4RV5VWYABQYHD', FALSE, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5ESS00000000000UOMCAT03', jsonb_build_object('en-US', 'Length', 'vi-VN', 'Chiều dài'), '01K5ESS000000000000UOM0009', '01JWNY20G23KD4RV5VWYABQYHD', FALSE, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5ESS00000000000UOMCAT04', jsonb_build_object('en-US', 'Volume', 'vi-VN', 'Thể tích'), '01K5ESS000000000000UOM000F', '01JWNY20G23KD4RV5VWYABQYHD', FALSE, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5ESS00000000000UOMCAT05', jsonb_build_object('en-US', 'Time', 'vi-VN', 'Thời gian'), '01K5ESS000000000000UOM000K', '01JWNY20G23KD4RV5VWYABQYHD', FALSE, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5ESS00000000000UOMCAT06', jsonb_build_object('en-US', 'Surface', 'vi-VN', 'Diện tích'), '01K5ESS000000000000UOM000Z', '01JWNY20G23KD4RV5VWYABQYHD', FALSE, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'essential_uoms'
	) THEN
		INSERT INTO "essential_uoms" ("id", "name", "symbol", "category_id", "uom_type", "factor", "rounding", "org_id", "is_archived", "created_at", "updated_at", "etag") VALUES
		-- Unit, reference unit piece; counted units are whole numbers only
		('01K5INV000000000000UOM0001', jsonb_build_object('en-US', 'Piece', 'vi-VN', 'Cái'), 'pc', '01K5INV00000000000UOMCAT01', 'reference', 1, 1, '01JWNY20G23KD4RV5VWYABQYHD', FALSE, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5ESS000000000000UOM000P', jsonb_build_object('en-US', 'Pair', 'vi-VN', 'Đôi'), 'pr', '01K5INV00000000000UOMCAT01', 'bigger_equal', 2, 1, '01JWNY20G23KD4RV5VWYABQYHD', FALSE, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5INV000000000000UOM0002', jsonb_build_object('en-US', 'Dozen', 'vi-VN', 'Tá'), 'dz', '01K5INV00000000000UOMCAT01', 'bigger_equal', 12, 1, '01JWNY20G23KD4RV5VWYABQYHD', FALSE, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5INV000000000000UOM0003', jsonb_build_object('en-US', 'Carton', 'vi-VN', 'Thùng'), 'ctn', '01K5INV00000000000UOMCAT01', 'bigger_equal', 24, 1, '01JWNY20G23KD4RV5VWYABQYHD', FALSE, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5ESS000000000000UOM000Q', jsonb_build_object('en-US', 'Gross', 'vi-VN', 'Gross'), 'gross', '01K5INV00000000000UOMCAT01', 'bigger_equal', 144, 1, '01JWNY20G23KD4RV5VWYABQYHD', FALSE, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Weight, reference unit kg
		('01K5INV000000000000UOM0004', jsonb_build_object('en-US', 'Kilogram', 'vi-VN', 'Ki-lô-gam'), 'kg', '01K5INV00000000000UOMCAT02', 'reference', 1, 0.001, '01JWNY20G23KD4RV5VWYABQYHD', FALSE, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5INV000000000000UOM0005', jsonb_build_object('en-US', 'Gram', 'vi-VN', 'Gam'), 'g', '01K5INV00000000000UOMCAT02', 'smaller', 0.001, 0.01, '01JWNY20G23KD4RV5VWYABQYHD', FALSE, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5ESS000000000000UOM000R', jsonb_build_object('en-US', 'Milligram', 'vi-VN', 'Mi-li-gam'), 'mg', '01K5INV00000000000UOMCAT02', 'smaller', 0.000001, 1, '01JWNY20G23KD4RV5VWYABQYHD', FALSE, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5ESS000000000000UOM0006', jsonb_build_object('en-US', 'Tonne', 'vi-VN', 'Tấn'), 't', '01K5INV00000000000UOMCAT02', 'bigger_equal', 1000, 0.001, '01JWNY20G23KD4RV5VWYABQYHD', FALSE, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5ESS000000000000UOM0007', jsonb_build_object('en-US', 'Ounce', 'vi-VN', 'Ao-xơ'), 'oz', '01K5INV00000000000UOMCAT02', 'smaller', 0.028349523125, 0.01, '01JWNY20G23KD4RV5VWYABQYHD', FALSE, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5ESS000000000000UOM0008', jsonb_build_object('en-US', 'Pound', 'vi-VN', 'Pao'), 'lb', '01K5INV00000000000UOMCAT02', 'smaller', 0.45359237, 0.01, '01JWNY20G23KD4RV5VWYABQYHD', FALSE, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Length, reference unit m
		('01K5ESS000000000000UOM0009', jsonb_build_object('en-US', 'Meter', 'vi-VN', 'Mét'), 'm', '01K5ESS00000000000UOMCAT03', 'reference', 1, 0.01, '01JWNY20G23KD4RV5VWYABQYHD', FALSE, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5ESS000000000000UOM000C', jsonb_build_object('en-US', 'Millimeter', 'vi-VN', 'Mi-li-mét'), 'mm', '01K5ESS00000000000UOMCAT03', 'smaller', 0.001, 1, '01JWNY20G23KD4RV5VWYABQYHD', FALSE, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5ESS000000000000UOM000B', jsonb_build_object('en-US', 'Centimeter', 'vi-VN', 'Xăng-ti-mét'), 'cm', '01K5ESS00000000000UOMCAT03', 'smaller', 0.01, 0.1, '01JWNY20G23KD4RV5VWYABQYHD', FALSE, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5ESS000000000000UOM000A', jsonb_build_object('en-US', 'Kilometer', 'vi-VN', 'Ki-lô-mét'), 'km', '01K5ESS00000000000UOMCAT03', 'bigger_equal', 1000, 0.001, '01JWNY20G23KD4RV5VWYABQYHD', FALSE, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5ESS000000000000UOM000D', jsonb_build_object('en-US', 'Inch', 'vi-VN', 'Inch'), 'in', '01K5ESS00000000000UOMCAT03', 'smaller', 0.0254, 0.01, '01JWNY20G23KD4RV5VWYABQYHD', FALSE, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5ESS000000000000UOM000E', jsonb_build_object('en-US', 'Foot', 'vi-VN', 'Foot'), 'ft', '01K5ESS00000000000UOMCAT03', 'smaller', 0.3048, 0.01, '01JWNY20G23KD4RV5VWYABQYHD', FALSE, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5ESS000000000000UOM000S', jsonb_build_object('en-US', 'Yard', 'vi-VN', 'Yard'), 'yd', '01K5ESS00000000000UOMCAT03', 'smaller', 0.9144, 0.01, '01JWNY20G23KD4RV5VWYABQYHD', FALSE, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5ESS000000000000UOM000T', jsonb_build_object('en-US', 'Mile', 'vi-VN', 'Dặm'), 'mi', '01K5ESS00000000000UOMCAT03', 'bigger_equal', 1609.344, 0.001, '01JWNY20G23KD4RV5VWYABQYHD', FALSE, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Volume, reference unit L
		('01K5ESS000000000000UOM000F', jsonb_build_object('en-US', 'Liter', 'vi-VN', 'Lít'), 'L', '01K5ESS00000000000UOMCAT04', 'reference', 1, 0.001, '01JWNY20G23KD4RV5VWYABQYHD', FALSE, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5ESS000000000000UOM000G', jsonb_build_object('en-US', 'Milliliter', 'vi-VN', 'Mi-li-lít'), 'mL', '01K5ESS00000000000UOMCAT04', 'smaller', 0.001, 1, '01JWNY20G23KD4RV5VWYABQYHD', FALSE, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5ESS000000000000UOM000V', jsonb_build_object('en-US', 'Centiliter', 'vi-VN', 'Xen-ti-lít'), 'cL', '01K5ESS00000000000UOMCAT04', 'smaller', 0.01, 1, '01JWNY20G23KD4RV5VWYABQYHD', FALSE, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5ESS000000000000UOM000H', jsonb_build_object('en-US', 'Cubic Meter', 'vi-VN', 'Mét khối'), 'm³', '01K5ESS00000000000UOMCAT04', 'bigger_equal', 1000, 0.001, '01JWNY20G23KD4RV5VWYABQYHD', FALSE, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5ESS000000000000UOM000X', jsonb_build_object('en-US', 'US Fluid Ounce', 'vi-VN', 'Ao-xơ chất lỏng Mỹ'), 'fl oz', '01K5ESS00000000000UOMCAT04', 'smaller', 0.029573529563, 0.01, '01JWNY20G23KD4RV5VWYABQYHD', FALSE, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5ESS000000000000UOM000J', jsonb_build_object('en-US', 'US Gallon', 'vi-VN', 'Ga-lông Mỹ'), 'gal', '01K5ESS00000000000UOMCAT04', 'bigger_equal', 3.785411784, 0.01, '01JWNY20G23KD4RV5VWYABQYHD', FALSE, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5ESS000000000000UOM000W', jsonb_build_object('en-US', 'Imperial Gallon', 'vi-VN', 'Ga-lông Anh'), 'imp gal', '01K5ESS00000000000UOMCAT04', 'bigger_equal', 4.54609, 0.01, '01JWNY20G23KD4RV5VWYABQYHD', FALSE, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Time, reference unit hour; a day is a calendar day, not a working day
		('01K5ESS000000000000UOM000K', jsonb_build_object('en-US', 'Hour', 'vi-VN', 'Giờ'), 'h', '01K5ESS00000000000UOMCAT05', 'reference', 1, 0.01, '01JWNY20G23KD4RV5VWYABQYHD', FALSE, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5ESS000000000000UOM000N', jsonb_build_object('en-US', 'Minute', 'vi-VN', 'Phút'), 'min', '01K5ESS00000000000UOMCAT05', 'smaller', 0.016666666667, 1, '01JWNY20G23KD4RV5VWYABQYHD', FALSE, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5ESS000000000000UOM000M', jsonb_build_object('en-US', 'Day', 'vi-VN', 'Ngày'), 'day', '01K5ESS00000000000UOMCAT05', 'bigger_equal', 24, 0.01, '01JWNY20G23KD4RV5VWYABQYHD', FALSE, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5ESS000000000000UOM000Y', jsonb_build_object('en-US', 'Week', 'vi-VN', 'Tuần'), 'wk', '01K5ESS00000000000UOMCAT05', 'bigger_equal', 168, 0.01, '01JWNY20G23KD4RV5VWYABQYHD', FALSE, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Surface, reference unit m²
		('01K5ESS000000000000UOM000Z', jsonb_build_object('en-US', 'Square Meter', 'vi-VN', 'Mét vuông'), 'm²', '01K5ESS00000000000UOMCAT06', 'reference', 1, 0.01, '01JWNY20G23KD4RV5VWYABQYHD', FALSE, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5ESS000000000000UOM0010', jsonb_build_object('en-US', 'Square Centimeter', 'vi-VN', 'Xăng-ti-mét vuông'), 'cm²', '01K5ESS00000000000UOMCAT06', 'smaller', 0.0001, 1, '01JWNY20G23KD4RV5VWYABQYHD', FALSE, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5ESS000000000000UOM0011', jsonb_build_object('en-US', 'Hectare', 'vi-VN', 'Héc-ta'), 'ha', '01K5ESS00000000000UOMCAT06', 'bigger_equal', 10000, 0.0001, '01JWNY20G23KD4RV5VWYABQYHD', FALSE, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;
