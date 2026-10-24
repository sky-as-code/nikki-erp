-- IAM resources and actions for the whole Inventory module: Products (product BR §6, §12), Stock
-- (stock BR §4.2.1, §4.2.2), Stock movement (§4.2.3 to §4.2.5) and the Phase 3 corrections
-- (§4.2.7 to §4.2.10). Every inventory permission row lives here, in one file, so that a reviewer
-- can see the module's entire permission surface without opening four migrations.
--
-- The dynamic resource engine asserts permissions using the schema name as the resource code,
-- so these codes must stay identical to the "inventory_product_template" /
-- "inventory_product_variant" / "inventory_stock_location" / "inventory_stock_transfer" / ...
-- schema names. A code that drifts from its schema denies every request, with nothing in the
-- response pointing at the seed.
--
-- Read-only resources: Stock Quant and Stock Move Line get no create/update/delete actions. Their
-- rows are derived state — a balance is the running total of completed movements, an allocation is
-- written by the reservation engine — and their engines refuse those verbs, so seeding a write
-- action would advertise a capability the engine rejects. The counting actions further down do not
-- contradict this: none of them writes on_hand_quantity directly.
--
-- Movement and correction operations are seeded as SEPARATE actions rather than folded into
-- "update", because they are materially different powers. Validating a transfer moves real goods
-- and cannot be undone by an edit — the correction is a reverse transfer — while updating one
-- changes a note. A role that may do the second should not thereby be able to do the first.
--
-- Deliberately no iam_entitlements rows. Unit of Measure grants the system "User" role a
-- domain-wide read so that any user can pick a unit while filling in an unrelated form. Product
-- and stock data are not universally readable: access follows explicitly assigned roles, so a
-- blanket grant would silently expose every product, balance, transfer and scrap in the system.
--
-- Note there is no seed for the inventory-loss location, the scrap location, or the
-- INV_CORRECTION operation type that adjustments and scraps generate their movements through.
-- Those are per-org business data — locations and operation types are created through their own
-- APIs, as Phase 1 established by seeding none — so a global migration is the wrong place for
-- them. The services resolve them per org and report a business violation naming what is missing
-- when an org has not configured them yet.
--
-- Warehouse Management resources and actions are seeded here too, in the same file as the rest of
-- the module's permission surface. Two resources carry lifecycle actions beyond CRUD and two do
-- not, which is the same split their schemas make. Warehouse and Inventory Location have an
-- operational state that can be paused and resumed, so they get suspend and resume. Storage
-- Category, Supply Relation and Putaway Rule have no state independent of archiving, so
-- set_archived is the whole of their lifecycle and seeding an activate/deactivate pair would
-- advertise operations that do not exist. Inventory Location's own resource row is seeded above as
-- Stock Location and later renamed; only its warehouse-specific actions are added here.

DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_resources'
	) THEN
		INSERT INTO "iam_resources" (
			"id", "name", "code", "description", "owner_type", "max_scope", "min_scope", "created_at", "etag"
		) VALUES
		('01M0A1P4QM3P0Y3T60T7RH2KZC', jsonb_build_object('en-US', 'Inventory - Product', 'vi-VN', 'Kho hàng - Sản phẩm'), 'inventory_product_template', 'Catalog-level definition of a product line', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0A1P4QJ0MFYJY0EQH0NVJ2R', jsonb_build_object('en-US', 'Inventory - Product Variant', 'vi-VN', 'Kho hàng - Biến thể sản phẩm'), 'inventory_product_variant', 'Concrete, transactable product of a template', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0A1P4QGGABP6SVFCTD6MF8K', jsonb_build_object('en-US', 'Inventory - Product Category', 'vi-VN', 'Kho hàng - Danh mục sản phẩm'), 'inventory_product_category', 'Hierarchical classification of product templates', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0A1P4QMJTAEFCJY5WRTNHWZ', jsonb_build_object('en-US', 'Inventory - Product Attribute', 'vi-VN', 'Kho hàng - Thuộc tính sản phẩm'), 'inventory_product_attribute', 'Attributes whose values form product variants', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0A1P4QAH9DQW7VXG9K87HWP', jsonb_build_object('en-US', 'Inventory - Product Attribute Value', 'vi-VN', 'Kho hàng - Giá trị thuộc tính sản phẩm'), 'inventory_product_attribute_value', 'Allowed values of a product attribute', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0A1P4QDXYWAJ8D4ZAD50VW7', jsonb_build_object('en-US', 'Inventory - Brand', 'vi-VN', 'Kho hàng - Thương hiệu'), 'inventory_brand', 'Product brands', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_actions'
	) THEN
		INSERT INTO "iam_actions" ("id", "name", "code", "description", "resource_id", "etag") VALUES
		-- Product Template. generate_variants is separate from update: regenerating the
		-- variant set of a live product is a heavier capability than editing its description.
		('01M0A1P4Q4K8REGFBYKH82M8Z1', jsonb_build_object('en-US', 'Create', 'vi-VN', 'Tạo'), 'create', NULL, '01M0A1P4QM3P0Y3T60T7RH2KZC', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0A1P4Q8WKR2A0RRQ21W6ZKT', jsonb_build_object('en-US', 'Update', 'vi-VN', 'Cập nhật'), 'update', NULL, '01M0A1P4QM3P0Y3T60T7RH2KZC', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0A1P4QQ799ZX5PPHC4RRN9F', jsonb_build_object('en-US', 'Delete', 'vi-VN', 'Xóa'), 'delete', NULL, '01M0A1P4QM3P0Y3T60T7RH2KZC', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0A1P4Q4XDRGWA4DDDNKD2DM', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M0A1P4QM3P0Y3T60T7RH2KZC', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0A1P4QS5PTBGTJSN1GS991H', jsonb_build_object('en-US', 'Set archived status', 'vi-VN', 'Đặt trạng thái lưu trữ'), 'set_archived', 'Archive a product template so it and its variants cannot be used by new transactions', '01M0A1P4QM3P0Y3T60T7RH2KZC', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0A1P4QBXY1DQFZPP4P4VD1T', jsonb_build_object('en-US', 'Generate variants', 'vi-VN', 'Tạo biến thể'), 'generate_variants', 'Generate or synchronize the variants of a product template', '01M0A1P4QM3P0Y3T60T7RH2KZC', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0A1P4QPRVWATTRCNFGPREVW', jsonb_build_object('en-US', 'Preview attribute configuration', 'vi-VN', 'Xem trước cấu hình thuộc tính'), 'preview_attribute_configuration', 'Dry-run an attribute configuration of a product template and see which variants it would keep, create, archive or delete', '01M0A1P4QM3P0Y3T60T7RH2KZC', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0A1P4QCPYPRDTMPXC0PY001', jsonb_build_object('en-US', 'Copy', 'vi-VN', 'Sao chép'), 'copy', 'Duplicate a product template with its attribute configuration', '01M0A1P4QM3P0Y3T60T7RH2KZC', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),

		-- Product Variant
		('01M0A1P4Q9YQ7HAFRENHBMJ9TR', jsonb_build_object('en-US', 'Create', 'vi-VN', 'Tạo'), 'create', NULL, '01M0A1P4QJ0MFYJY0EQH0NVJ2R', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0A1P4Q4P4J5K2NANT4K78VK', jsonb_build_object('en-US', 'Update', 'vi-VN', 'Cập nhật'), 'update', NULL, '01M0A1P4QJ0MFYJY0EQH0NVJ2R', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0A1P4QB7BD6NB8M9SEK5Y20', jsonb_build_object('en-US', 'Delete', 'vi-VN', 'Xóa'), 'delete', NULL, '01M0A1P4QJ0MFYJY0EQH0NVJ2R', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0A1P4QSQMEVAZ1FFTAVA317', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M0A1P4QJ0MFYJY0EQH0NVJ2R', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0A1P4QEVX39CJ2TJ2AN97FT', jsonb_build_object('en-US', 'Set archived status', 'vi-VN', 'Đặt trạng thái lưu trữ'), 'set_archived', 'Archive a product variant so it cannot be used by new transactions', '01M0A1P4QJ0MFYJY0EQH0NVJ2R', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),

		-- Product Category
		('01M0A1P4QB6H18345VP8SG849J', jsonb_build_object('en-US', 'Create', 'vi-VN', 'Tạo'), 'create', NULL, '01M0A1P4QGGABP6SVFCTD6MF8K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0A1P4Q4QZBR3FPNJJ1V9FRJ', jsonb_build_object('en-US', 'Update', 'vi-VN', 'Cập nhật'), 'update', NULL, '01M0A1P4QGGABP6SVFCTD6MF8K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0A1P4QMSFDXYBN3D3RS9K5E', jsonb_build_object('en-US', 'Delete', 'vi-VN', 'Xóa'), 'delete', NULL, '01M0A1P4QGGABP6SVFCTD6MF8K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0A1P4QYG8G9QC4EWTWAV723', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M0A1P4QGGABP6SVFCTD6MF8K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0A1P4QWZHYXV89G9BT6XJBE', jsonb_build_object('en-US', 'Set archived status', 'vi-VN', 'Đặt trạng thái lưu trữ'), 'set_archived', 'Archive a product category', '01M0A1P4QGGABP6SVFCTD6MF8K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),

		-- Product Attribute
		('01M0A1P4Q94CATJA4MXP6SED7N', jsonb_build_object('en-US', 'Create', 'vi-VN', 'Tạo'), 'create', NULL, '01M0A1P4QMJTAEFCJY5WRTNHWZ', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0A1P4QXAG5J1BV1JSP6W6G9', jsonb_build_object('en-US', 'Update', 'vi-VN', 'Cập nhật'), 'update', NULL, '01M0A1P4QMJTAEFCJY5WRTNHWZ', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0A1P4Q95527SM8SJK2QRH4R', jsonb_build_object('en-US', 'Delete', 'vi-VN', 'Xóa'), 'delete', NULL, '01M0A1P4QMJTAEFCJY5WRTNHWZ', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0A1P4QFDM0XSX8VXAHZ31EG', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M0A1P4QMJTAEFCJY5WRTNHWZ', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0A1P4QVNXPS112TGMJTTHTJ', jsonb_build_object('en-US', 'Set archived status', 'vi-VN', 'Đặt trạng thái lưu trữ'), 'set_archived', 'Archive a product attribute', '01M0A1P4QMJTAEFCJY5WRTNHWZ', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0A1P4QW9DVA1EPR1CEXTRA0', jsonb_build_object('en-US', 'Update value', 'vi-VN', 'Cập nhật giá trị'), 'update_value', 'Change the extra price a product template charges for one of its attribute values', '01M0A1P4QMJTAEFCJY5WRTNHWZ', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),

		-- Product Attribute Value
		('01M0A1P4QKK2BS2M7Q4ZJVX3ED', jsonb_build_object('en-US', 'Create', 'vi-VN', 'Tạo'), 'create', NULL, '01M0A1P4QAH9DQW7VXG9K87HWP', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0A1P4QBJYDR8D8Y1EESRH4W', jsonb_build_object('en-US', 'Update', 'vi-VN', 'Cập nhật'), 'update', NULL, '01M0A1P4QAH9DQW7VXG9K87HWP', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0A1P4Q71ATJEWZM1BAM7E3N', jsonb_build_object('en-US', 'Delete', 'vi-VN', 'Xóa'), 'delete', NULL, '01M0A1P4QAH9DQW7VXG9K87HWP', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0A1P4Q4QF419T7W52WS0HMF', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M0A1P4QAH9DQW7VXG9K87HWP', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0A1P4QEYDQTRE7QH9NHGQA3', jsonb_build_object('en-US', 'Set archived status', 'vi-VN', 'Đặt trạng thái lưu trữ'), 'set_archived', 'Archive a product attribute value', '01M0A1P4QAH9DQW7VXG9K87HWP', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),

		-- Brand
		('01M0A1P4QKBPP966KQK6NDTY5P', jsonb_build_object('en-US', 'Create', 'vi-VN', 'Tạo'), 'create', NULL, '01M0A1P4QDXYWAJ8D4ZAD50VW7', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0A1P4QXRE54C2AJPY554FD9', jsonb_build_object('en-US', 'Update', 'vi-VN', 'Cập nhật'), 'update', NULL, '01M0A1P4QDXYWAJ8D4ZAD50VW7', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0A1P4QCM2KQ3VNHZVHGV45P', jsonb_build_object('en-US', 'Delete', 'vi-VN', 'Xóa'), 'delete', NULL, '01M0A1P4QDXYWAJ8D4ZAD50VW7', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0A1P4QRS1VZA44V25595WBX', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M0A1P4QDXYWAJ8D4ZAD50VW7', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0A1P4Q67457GP4WXDW09M6F', jsonb_build_object('en-US', 'Set archived status', 'vi-VN', 'Đặt trạng thái lưu trữ'), 'set_archived', 'Archive a brand', '01M0A1P4QDXYWAJ8D4ZAD50VW7', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)

		ON CONFLICT ("id") DO NOTHING;
	END IF;

	-- ---------------------------------------------------------------------------------------
	-- Stock: locations, operation types and quants (stock BR §4.2.1, §4.2.2).
	-- ---------------------------------------------------------------------------------------
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_resources'
	) THEN
		INSERT INTO "iam_resources" (
			"id", "name", "code", "description", "owner_type", "max_scope", "min_scope", "created_at", "etag"
		) VALUES
		('01M0B2Q5R34KTKF2YBXPBBC2JH', jsonb_build_object('en-US', 'Inventory - Stock Location', 'vi-VN', 'Kho hàng - Vị trí kho'), 'inventory_stock_location', 'Where stock is held, and the counterparties movements run against', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0B2Q5R9XBX4YENXT9PN8J70', jsonb_build_object('en-US', 'Inventory - Stock Operation Type', 'vi-VN', 'Kho hàng - Loại nghiệp vụ kho'), 'inventory_stock_operation_type', 'Policy applied when processing a stock transfer', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0B2Q5R3BQX4QWMQCNPQV5H6', jsonb_build_object('en-US', 'Inventory - Stock Quant', 'vi-VN', 'Kho hàng - Tồn kho theo vị trí'), 'inventory_stock_quant', 'Current stock balance at one product, location, lot, package and owner', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_actions'
	) THEN
		INSERT INTO "iam_actions" ("id", "name", "code", "description", "resource_id", "etag") VALUES
		-- Stock Location
		('01M0B2Q5RVRPCQHTGPJBZWQZ6P', jsonb_build_object('en-US', 'Create', 'vi-VN', 'Tạo'), 'create', NULL, '01M0B2Q5R34KTKF2YBXPBBC2JH', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0B2Q5RBAYTGAPKWX3GJE67S', jsonb_build_object('en-US', 'Update', 'vi-VN', 'Cập nhật'), 'update', NULL, '01M0B2Q5R34KTKF2YBXPBBC2JH', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0B2Q5R9ZY40RV3Z3E87R2PY', jsonb_build_object('en-US', 'Delete', 'vi-VN', 'Xóa'), 'delete', NULL, '01M0B2Q5R34KTKF2YBXPBBC2JH', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0B2Q5RXJ7JESHQXT054DFVK', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M0B2Q5R34KTKF2YBXPBBC2JH', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0B2Q5RAG9ZMPANTTK8156YF', jsonb_build_object('en-US', 'Set archived status', 'vi-VN', 'Đặt trạng thái lưu trữ'), 'set_archived', 'Archive a stock location so new transfers cannot use it', '01M0B2Q5R34KTKF2YBXPBBC2JH', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),

		-- Stock Operation Type
		('01M0B2Q5RKQM849HKX7JRZ7E18', jsonb_build_object('en-US', 'Create', 'vi-VN', 'Tạo'), 'create', NULL, '01M0B2Q5R9XBX4YENXT9PN8J70', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0B2Q5RCCWQAGNZ9QV2KWX6J', jsonb_build_object('en-US', 'Update', 'vi-VN', 'Cập nhật'), 'update', NULL, '01M0B2Q5R9XBX4YENXT9PN8J70', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0B2Q5R42M238624GZGXXJN7', jsonb_build_object('en-US', 'Delete', 'vi-VN', 'Xóa'), 'delete', NULL, '01M0B2Q5R9XBX4YENXT9PN8J70', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0B2Q5R450DGZB5QD4YJXPZR', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M0B2Q5R9XBX4YENXT9PN8J70', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0B2Q5R2MZ1210Q7Q59YVGST', jsonb_build_object('en-US', 'Set archived status', 'vi-VN', 'Đặt trạng thái lưu trữ'), 'set_archived', 'Archive an operation type so new transfers cannot use it, while historical transfers still resolve it', '01M0B2Q5R9XBX4YENXT9PN8J70', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),

		-- Stock Quant. Read only: balances change through an adjustment, transfer or scrap,
		-- each of which records the movement that justifies the new number (AC-STOCK-002).
		('01M0B2Q5RPK5A179SW4KMBCQT5', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M0B2Q5R3BQX4QWMQCNPQV5H6', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	-- ---------------------------------------------------------------------------------------
	-- Stock movement: transfers, moves, move lines and dependencies (stock BR §4.2.3 to §4.2.5).
	-- ---------------------------------------------------------------------------------------
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_resources'
	) THEN
		INSERT INTO "iam_resources" (
			"id", "name", "code", "description", "owner_type", "max_scope", "min_scope", "created_at", "etag"
		) VALUES
		('01M0B434KTKF2YBXPBBC2JH9XB', jsonb_build_object('en-US', 'Inventory - Stock Transfer', 'vi-VN', 'Kho hàng - Phiếu chuyển kho'), 'inventory_stock_transfer', 'The header of a stock transaction: what moves, between which locations, in what state', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0B4X4YENXT9PN8J703BQX4Q', jsonb_build_object('en-US', 'Inventory - Stock Move', 'vi-VN', 'Kho hàng - Dịch chuyển kho'), 'inventory_stock_move', 'One line of demand within a transfer: this much of this variant, from here to there', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0B4WMQCNPQV5H6VRPCQHTGP', jsonb_build_object('en-US', 'Inventory - Stock Move Line', 'vi-VN', 'Kho hàng - Dòng dịch chuyển kho'), 'inventory_stock_move_line', 'The execution detail of a move: what was actually taken, and from which balance', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0B4JBZWQZ6PBAYTGAPKWX3G', jsonb_build_object('en-US', 'Inventory - Stock Move Dependency', 'vi-VN', 'Kho hàng - Phụ thuộc dịch chuyển kho'), 'inventory_stock_move_dependency', 'Ordering between the steps of a multi-step flow', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_actions'
	) THEN
		INSERT INTO "iam_actions" ("id", "name", "code", "description", "resource_id", "etag") VALUES
		-- Stock Transfer: CRUD, then the six movement operations.
		('01M0B4JE67S9ZY40RV3Z3E87R2', jsonb_build_object('en-US', 'Create', 'vi-VN', 'Tạo'), 'create', NULL, '01M0B434KTKF2YBXPBBC2JH9XB', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0B4PYXJ7JESHQXT054DFVKA', jsonb_build_object('en-US', 'Update', 'vi-VN', 'Cập nhật'), 'update', NULL, '01M0B434KTKF2YBXPBBC2JH9XB', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0B4G9ZMPANTTK8156YFKQM8', jsonb_build_object('en-US', 'Delete', 'vi-VN', 'Xóa'), 'delete', NULL, '01M0B434KTKF2YBXPBBC2JH9XB', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0B449HKX7JRZ7E18CCWQAGN', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M0B434KTKF2YBXPBBC2JH9XB', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0B4Z9QV2KWX6J42M238624G', jsonb_build_object('en-US', 'Confirm', 'vi-VN', 'Xác nhận'), 'confirm', 'Commit a draft transfer to its demand, reserving stock when the operation type says so', '01M0B434KTKF2YBXPBBC2JH9XB', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0B4ZGXXJN7450DGZB5QD4YJ', jsonb_build_object('en-US', 'Reserve', 'vi-VN', 'Giữ chỗ'), 'reserve', 'Claim stock for the transfer''s moves without moving any of it', '01M0B434KTKF2YBXPBBC2JH9XB', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0B4XPZR2MZ1210Q7Q59YVGS', jsonb_build_object('en-US', 'Unreserve', 'vi-VN', 'Hủy giữ chỗ'), 'unreserve', 'Release the transfer''s claims, leaving on-hand quantities untouched', '01M0B434KTKF2YBXPBBC2JH9XB', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0B4TPK5A179SW4KMBCQT5HY', jsonb_build_object('en-US', 'Validate', 'vi-VN', 'Thực hiện chuyển kho'), 'validate', 'Execute the transfer: the only operation that changes an on-hand balance, and one no edit can undo', '01M0B434KTKF2YBXPBBC2JH9XB', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0B4JHJFQXQF7T7WT9PJ6A5Z', jsonb_build_object('en-US', 'Cancel', 'vi-VN', 'Hủy'), 'cancel', 'Abandon an unfinished transfer, releasing whatever it holds', '01M0B434KTKF2YBXPBBC2JH9XB', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),

		-- Stock Move
		('01M0B4SPVX86PAWSQHJ62WARZ4', jsonb_build_object('en-US', 'Create', 'vi-VN', 'Tạo'), 'create', NULL, '01M0B4X4YENXT9PN8J703BQX4Q', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0B46CHAF22AG5P8MSV3APSZ', jsonb_build_object('en-US', 'Update', 'vi-VN', 'Cập nhật'), 'update', NULL, '01M0B4X4YENXT9PN8J703BQX4Q', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0B4WV9PR3JJWM5V1RSFBTK2', jsonb_build_object('en-US', 'Delete', 'vi-VN', 'Xóa'), 'delete', NULL, '01M0B4X4YENXT9PN8J703BQX4Q', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0B4SBHVJWKG3EXFYP6CDZHY', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M0B4X4YENXT9PN8J703BQX4Q', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),

		-- Stock Move Line. Read only: its rows are written by the reservation engine, and a
		-- client-written allocation would be a claim the balance itself knows nothing about.
		('01M0B4KB8JA48CQRTMF6J1B9WY', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M0B4WMQCNPQV5H6VRPCQHTGP', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),

		-- Stock Move Dependency
		('01M0B4GJM613QQCEKG6NVT8675', jsonb_build_object('en-US', 'Create', 'vi-VN', 'Tạo'), 'create', NULL, '01M0B4JBZWQZ6PBAYTGAPKWX3G', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0B44FMEYRB12NBBJPDHH40T', jsonb_build_object('en-US', 'Delete', 'vi-VN', 'Xóa'), 'delete', NULL, '01M0B4JBZWQZ6PBAYTGAPKWX3G', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0B407H0YTHMAN6ZHPANDZBA', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M0B4JBZWQZ6PBAYTGAPKWX3G', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	-- ---------------------------------------------------------------------------------------
	-- Phase 3 corrections: scraps, physical counting, returns (stock BR §4.2.7 to §4.2.10).
	--
	-- Stock Scrap carries write actions, unlike the quant, because a scrap is a document a user
	-- raises and may abandon while it is draft. What the derived service constrains is when: a
	-- done scrap can be neither edited nor deleted.
	-- ---------------------------------------------------------------------------------------
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_resources'
	) THEN
		INSERT INTO "iam_resources" (
			"id", "name", "code", "description", "owner_type", "max_scope", "min_scope", "created_at", "etag"
		) VALUES
		('01M0C3SQRTAAAAAAAAAAAAAAAB', jsonb_build_object('en-US', 'Inventory - Stock Scrap', 'vi-VN', 'Kho hàng - Hủy hàng tồn kho'), 'inventory_stock_scrap', 'A document that removes goods from usable stock by moving them to a scrap location', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_actions'
	) THEN
		INSERT INTO "iam_actions" ("id", "name", "code", "description", "resource_id", "etag") VALUES
		-- Stock Scrap: CRUD, then the operation that executes the document.
		('01M0C3SQRTAAAAAAAAAAAAAAAC', jsonb_build_object('en-US', 'Create', 'vi-VN', 'Tạo'), 'create', NULL, '01M0C3SQRTAAAAAAAAAAAAAAAB', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0C3SQRTAAAAAAAAAAAAAAAD', jsonb_build_object('en-US', 'Update', 'vi-VN', 'Cập nhật'), 'update', NULL, '01M0C3SQRTAAAAAAAAAAAAAAAB', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0C3SQRTAAAAAAAAAAAAAAAE', jsonb_build_object('en-US', 'Delete', 'vi-VN', 'Xóa'), 'delete', NULL, '01M0C3SQRTAAAAAAAAAAAAAAAB', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0C3SQRTAAAAAAAAAAAAAAAF', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M0C3SQRTAAAAAAAAAAAAAAAB', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0C3SQRTAAAAAAAAAAAAAAAG', jsonb_build_object('en-US', 'Do Scrap', 'vi-VN', 'Thực hiện hủy hàng'), 'do_scrap', 'Execute the scrap: generate the movement that writes the goods off. Cannot be undone by an edit.', '01M0C3SQRTAAAAAAAAAAAAAAAB', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),

		-- Physical inventory and cycle counting, on the Stock Quant resource seeded above. None of
		-- these writes on_hand_quantity: enter and reset write count metadata, and apply changes
		-- the balance only by generating a movement.
		('01M0C3SQRTAAAAAAAAAAAAAAAH', jsonb_build_object('en-US', 'Enter Count', 'vi-VN', 'Nhập số lượng kiểm đếm'), 'enter_count', 'Record what a physical count found. Does not change the balance.', '01M0B2Q5R3BQX4QWMQCNPQV5H6', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0C3SQRTAAAAAAAAAAAAAAAJ', jsonb_build_object('en-US', 'Reset Count', 'vi-VN', 'Đặt lại số kiểm đếm'), 'reset_count', 'Abandon a pending count. Does not change the balance.', '01M0B2Q5R3BQX4QWMQCNPQV5H6', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0C3SQRTAAAAAAAAAAAAAAAK', jsonb_build_object('en-US', 'Apply Adjustment', 'vi-VN', 'Áp dụng điều chỉnh'), 'apply_adjustment', 'Turn a pending count into a real balance change by generating an adjustment movement.', '01M0B2Q5R3BQX4QWMQCNPQV5H6', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0C3SQRTAAAAAAAAAAAAABA', jsonb_build_object('en-US', 'Schedule Count', 'vi-VN', 'Lên lịch kiểm đếm'), 'schedule_count', 'Set when a balance is next due to be counted.', '01M0B2Q5R3BQX4QWMQCNPQV5H6', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0C3SQRTAAAAAAAAAAAAABB', jsonb_build_object('en-US', 'Assign Counter', 'vi-VN', 'Phân công người kiểm đếm'), 'assign_counter', 'Name who is responsible for counting a balance.', '01M0B2Q5R3BQX4QWMQCNPQV5H6', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),

		-- Create Reverse Transfer, on the Stock Transfer resource seeded above.
		('01M0C3SQRTAAAAAAAAAAAAABC', jsonb_build_object('en-US', 'Create Return', 'vi-VN', 'Tạo phiếu trả hàng'), 'create_return', 'Raise a reverse transfer for goods coming back. Never alters the original transfer.', '01M0B434KTKF2YBXPBBC2JH9XB', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	-- ---------------------------------------------------------------------------------------
	-- Warehouse Management: warehouses, storage categories, supply relations, putaway rules.
	-- ---------------------------------------------------------------------------------------
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_resources'
	) THEN
		INSERT INTO "iam_resources" ("id", "name", "code", "description", "owner_type", "max_scope", "min_scope", "created_at", "etag") VALUES
		('01M1N2Q3V769G2C81EBDXR3QH8', jsonb_build_object('en-US', 'Inventory - Warehouse', 'vi-VN', 'Kho hàng - Kho hàng'), 'inventory_warehouse', 'A site that can receive, hold and dispatch goods', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M14PV5Q32776611SWMJMG04X', jsonb_build_object('en-US', 'Inventory - Storage Category', 'vi-VN', 'Kho hàng - Loại lưu trữ'), 'inventory_storage_category', 'Capacity and mixing policy a location may carry', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M1E4451SXVZ5KBSM1HEPQWAZ', jsonb_build_object('en-US', 'Inventory - Warehouse Supply Relation', 'vi-VN', 'Kho hàng - Quan hệ cung ứng kho'), 'inventory_warehouse_supply_relation', 'Which warehouse may resupply which', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M1E4XEPDJ24JDTC2R8H18X78', jsonb_build_object('en-US', 'Inventory - Putaway Rule', 'vi-VN', 'Kho hàng - Quy tắc cất hàng'), 'inventory_putaway_rule', 'Where arriving goods should be put', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_actions'
	) THEN
		INSERT INTO "iam_actions" ("id", "name", "code", "description", "resource_id", "etag") VALUES
		-- Warehouse
		('01M11G6HTXQC836YT846V981EN', jsonb_build_object('en-US', 'Create', 'vi-VN', 'Tạo'), 'create', NULL, '01M1N2Q3V769G2C81EBDXR3QH8', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M1G8C02SBBVGGFDCPEPE1EM2', jsonb_build_object('en-US', 'Update', 'vi-VN', 'Cập nhật'), 'update', NULL, '01M1N2Q3V769G2C81EBDXR3QH8', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M167H9RH93W0BACXVP7K49AY', jsonb_build_object('en-US', 'Delete', 'vi-VN', 'Xóa'), 'delete', NULL, '01M1N2Q3V769G2C81EBDXR3QH8', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M1PD46HPM6N2Z5ZN8Y4MW8P8', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M1N2Q3V769G2C81EBDXR3QH8', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M1BTGBM50Y0A8CB3EBC4XPRD', jsonb_build_object('en-US', 'Set archived status', 'vi-VN', 'Đặt trạng thái lưu trữ'), 'set_archived', 'Archive a warehouse so it is no longer offered for new work, while historical movements still resolve it', '01M1N2Q3V769G2C81EBDXR3QH8', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M16PK8AJTWVK2CN1G9Q26TW9', jsonb_build_object('en-US', 'Suspend', 'vi-VN', 'Tạm ngưng'), 'suspend', 'Temporarily close a warehouse for repairs, a count or an incident, without archiving it', '01M1N2Q3V769G2C81EBDXR3QH8', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M1TNQ0YZ60Q6M983RYMENHT9', jsonb_build_object('en-US', 'Resume', 'vi-VN', 'Tiếp tục sử dụng'), 'resume', 'Return a suspended warehouse to service once its configuration is verified', '01M1N2Q3V769G2C81EBDXR3QH8', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M1QM9N8PJXZRZ7SFT7A1TYE7', jsonb_build_object('en-US', 'Configure incoming flow', 'vi-VN', 'Cấu hình luồng nhập'), 'configure_incoming_flow', 'Change how many stops goods make on the way in, provisioning the locations the new flow needs', '01M1N2Q3V769G2C81EBDXR3QH8', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M1S7KGMGPV9P521Q90M9XS6G', jsonb_build_object('en-US', 'Configure outgoing flow', 'vi-VN', 'Cấu hình luồng xuất'), 'configure_outgoing_flow', 'Change how many stops goods make on the way out, provisioning the locations the new flow needs', '01M1N2Q3V769G2C81EBDXR3QH8', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),

		-- Storage Category
		('01M1R2YXJCBHS46FFG1S741XSB', jsonb_build_object('en-US', 'Create', 'vi-VN', 'Tạo'), 'create', NULL, '01M14PV5Q32776611SWMJMG04X', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M1YWJX8MCN9FHS7TT5QBTJD2', jsonb_build_object('en-US', 'Update', 'vi-VN', 'Cập nhật'), 'update', NULL, '01M14PV5Q32776611SWMJMG04X', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M1NSGN47CJFD0YJ0ZAZ72BDS', jsonb_build_object('en-US', 'Delete', 'vi-VN', 'Xóa'), 'delete', NULL, '01M14PV5Q32776611SWMJMG04X', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M1PGC9D35JZXEJQYD02NT6VA', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M14PV5Q32776611SWMJMG04X', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M1JRTC2M87NX1J47HCS96BJG', jsonb_build_object('en-US', 'Set archived status', 'vi-VN', 'Đặt trạng thái lưu trữ'), 'set_archived', 'Archive a storage category so it can no longer be assigned to a location', '01M14PV5Q32776611SWMJMG04X', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),

		-- Warehouse Supply Relation
		('01M1XKQ3Q4JB467QQH3V2ZTS5H', jsonb_build_object('en-US', 'Create', 'vi-VN', 'Tạo'), 'create', NULL, '01M1E4451SXVZ5KBSM1HEPQWAZ', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M1Y4K0P5B82YYA7HMV104YDC', jsonb_build_object('en-US', 'Update', 'vi-VN', 'Cập nhật'), 'update', NULL, '01M1E4451SXVZ5KBSM1HEPQWAZ', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M18GYESY6RFGGBTSWJV2G8BV', jsonb_build_object('en-US', 'Delete', 'vi-VN', 'Xóa'), 'delete', NULL, '01M1E4451SXVZ5KBSM1HEPQWAZ', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M1Q433TWYAQCQZQYRCD3TEZG', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M1E4451SXVZ5KBSM1HEPQWAZ', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M1FYVTS2JJ20WF0TDFFT1MVY', jsonb_build_object('en-US', 'Set archived status', 'vi-VN', 'Đặt trạng thái lưu trữ'), 'set_archived', 'Archive a supply relation so it is no longer used for resupply planning', '01M1E4451SXVZ5KBSM1HEPQWAZ', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),

		-- Putaway Rule
		('01M1R8PK2PYCKKW6N92Z87HEZ9', jsonb_build_object('en-US', 'Create', 'vi-VN', 'Tạo'), 'create', NULL, '01M1E4XEPDJ24JDTC2R8H18X78', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M1PF44FEPSM3KX9MX66G7VVJ', jsonb_build_object('en-US', 'Update', 'vi-VN', 'Cập nhật'), 'update', NULL, '01M1E4XEPDJ24JDTC2R8H18X78', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M1T7FFNF5CABX8M810TBBPVR', jsonb_build_object('en-US', 'Delete', 'vi-VN', 'Xóa'), 'delete', NULL, '01M1E4XEPDJ24JDTC2R8H18X78', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M1ZC6FM2WHP5FHPVH2GY5VZ2', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M1E4XEPDJ24JDTC2R8H18X78', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M1XC5WGZDP897TE11RA0GEAX', jsonb_build_object('en-US', 'Set archived status', 'vi-VN', 'Đặt trạng thái lưu trữ'), 'set_archived', 'Archive a putaway rule so it is no longer evaluated', '01M1E4XEPDJ24JDTC2R8H18X78', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M1322KA4ECQBCC1QA126YC6P', jsonb_build_object('en-US', 'Suggest location', 'vi-VN', 'Gợi ý vị trí'), 'suggest_location', 'Ask where arriving goods should be put. Returns a suggestion and changes nothing', '01M1E4XEPDJ24JDTC2R8H18X78', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	-- Inventory Location's new lifecycle actions. The resource row already exists, seeded as Stock
	-- Location above and renamed to Inventory Location, so only the actions are added.
	--
	-- Suspending a location is allowed while it still holds stock — locking a damaged rack that
	-- holds goods is exactly the point — whereas archiving one that does is refused. The two are
	-- separate permissions for the same reason they are separate operations.
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_actions'
	) AND EXISTS (
		SELECT FROM "iam_resources" WHERE "code" = 'inventory_location'
	) THEN
		INSERT INTO "iam_actions" ("id", "name", "code", "description", "resource_id", "etag")
		SELECT '01M10ZMX7HPFP8P614NF0520ET', jsonb_build_object('en-US', 'Suspend', 'vi-VN', 'Tạm ngưng'), 'suspend', 'Temporarily take a location out of use, keeping whatever it holds', "id", (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text
		FROM "iam_resources" WHERE "code" = 'inventory_location'
		UNION ALL
		SELECT '01M1M8VKXZ79440FXRX307GZPQ', jsonb_build_object('en-US', 'Resume', 'vi-VN', 'Tiếp tục sử dụng'), 'resume', 'Return a suspended location to use', "id", (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text
		FROM "iam_resources" WHERE "code" = 'inventory_location'
		UNION ALL
		SELECT '01M1ZHTE67SWYQV8P2J24P9C6F', jsonb_build_object('en-US', 'Move', 'vi-VN', 'Di chuyển'), 'move', 'Re-parent a location, rewriting the cached path of everything beneath it', "id", (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text
		FROM "iam_resources" WHERE "code" = 'inventory_location'
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;

-- ---------------------------------------------------------------------------
-- Merged from 1005006_inventory_product_stock_iam.sql
-- ---------------------------------------------------------------------------

-- Permissions for the product-stock integration.
--
-- Two things are added. The new configuration resource, which holds a product line's inventory
-- unit; and one read action on the existing Stock Quant resource, covering the summaries a
-- product page shows.
--
-- The reads are one permission rather than six. They are the same power — seeing how much stock a
-- product has — sliced only by which product is being looked at, so separating them would let an
-- administrator grant a nonsensical combination such as "may see a template's total but not the
-- variants behind it". Granting it lets a user look; it lets them change nothing, because every
-- action behind it is a read.

DO $$
BEGIN
	-- ---------------------------------------------------------------------------------------
	-- Stock's settings for a product line (CR §11.4).
	-- ---------------------------------------------------------------------------------------
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_resources'
	) THEN
		INSERT INTO "iam_resources" (
			"id", "name", "code", "description", "owner_type", "max_scope", "min_scope", "created_at", "etag"
		) VALUES
		('01M2A7QK3P5NXCW9VBDT4RGZH2', jsonb_build_object('en-US', 'Inventory - Stock Product Configuration', 'vi-VN', 'Kho hàng - Cấu hình kho của sản phẩm'), 'inventory_stock_product_config', 'Stock settings of a product line, currently the unit its balances are counted in', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_actions'
	) THEN
		INSERT INTO "iam_actions" ("id", "name", "code", "description", "resource_id", "etag") VALUES
		('01M2A7QK3PB8YFHV5JQW6E2SND', jsonb_build_object('en-US', 'Create', 'vi-VN', 'Tạo'), 'create', 'Set the unit a product line''s stock is counted in', '01M2A7QK3P5NXCW9VBDT4RGZH2', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M2A7QK3PC4TZR7KDXA9HM3VE', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M2A7QK3P5NXCW9VBDT4RGZH2', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Changing the unit after stock has moved would reinterpret every quantity ever recorded,
		-- so the engine refuses it outright. This permission covers the case where it is still
		-- allowed: a product that has never been used.
		('01M2A7QK3PDGX2W8N4VJQ5F7BK', jsonb_build_object('en-US', 'Update', 'vi-VN', 'Cập nhật'), 'update', 'Change the unit, which the engine permits only while the product has no stock or stock history', '01M2A7QK3P5NXCW9VBDT4RGZH2', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M2A7QK3PEH6RQY3TZB8KCW9M', jsonb_build_object('en-US', 'Delete', 'vi-VN', 'Xóa'), 'delete', NULL, '01M2A7QK3P5NXCW9VBDT4RGZH2', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	-- ---------------------------------------------------------------------------------------
	-- The product-facing stock reads, on the existing Stock Quant resource.
	-- ---------------------------------------------------------------------------------------
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_actions'
	) AND EXISTS (
		SELECT FROM "iam_resources" WHERE "code" = 'inventory_stock_quant'
	) THEN
		INSERT INTO "iam_actions" ("id", "name", "code", "description", "resource_id", "etag")
		SELECT '01M2A7QK3PF9J5MXV2NQ8YT4HR', jsonb_build_object('en-US', 'Read Product Stock', 'vi-VN', 'Xem tồn kho sản phẩm'), 'read_product_stock', 'See how much stock a product has, where it sits and what would be stranded by archiving it. Reads only.', "id", (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text
		FROM "iam_resources" WHERE "code" = 'inventory_stock_quant'
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;

-- ---------------------------------------------------------------------------
-- Merged from 1005008_inventory_fulfillment_result_iam.sql
-- ---------------------------------------------------------------------------

-- The permission for recording what an executor physically handed over.
--
-- Only one new action is seeded. The other three demand-addressed routes — reserve_for_source,
-- release_for_source and reallocate_reservation — deliberately answer to the existing `reserve` and
-- `unreserve` permissions: they are the same powers over the same stock, differing only in whether
-- the caller names the transfer or the demand it was raised for, and giving them separate codes
-- would let a role hold one without the other by accident.
--
-- apply_fulfillment_result is separate from `validate` because the callers are different in kind.
-- Validate is an operator saying "ship this document"; this is a service relaying what a machine
-- managed to dispense. Granting a vending service `validate` would let it complete any transfer in
-- the organization, which is far more than reporting its own result requires.

DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_actions'
	) THEN
		INSERT INTO "iam_actions" ("id", "name", "code", "description", "resource_id", "etag") VALUES
		('01M3NVFRES0000000000000001', jsonb_build_object('en-US', 'Apply fulfillment result', 'vi-VN', 'Áp dụng kết quả giao hàng'), 'apply_fulfillment_result', 'Record what an executor physically handed over against a reservation: consume what left, release the hold on what did not', '01M0B434KTKF2YBXPBBC2JH9XB', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;

-- Warehouse-level stock reservations (CR-INV-SALES-WH-RESERVATION). Read-only as a resource:
-- rows are written only by the reservation operations, which are seeded as their own actions.
-- Reserve and check are collection-level; release and consume act on one reservation. There is
-- no action for protecting a paid reservation, because only the paying module's port may do it.
DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_resources'
	) THEN
		INSERT INTO "iam_resources" (
			"id", "name", "code", "description", "owner_type", "max_scope", "min_scope", "created_at", "etag"
		) VALUES
		('01M3WHRSV00000000000000001', jsonb_build_object('en-US', 'Inventory - Stock Reservation', 'vi-VN', 'Kho hàng - Giữ chỗ tồn kho'), 'inventory_stock_reservation', 'A quantity of one variant committed at one warehouse to one demand line, with no location chosen', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_actions'
	) THEN
		INSERT INTO "iam_actions" ("id", "name", "code", "description", "resource_id", "etag") VALUES
		('01M3WHRSVACT00000000000001', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M3WHRSV00000000000000001', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3WHRSVACT00000000000002', jsonb_build_object('en-US', 'Reserve warehouse stock', 'vi-VN', 'Giữ chỗ tồn kho'), 'reserve_warehouse_stock', 'Commit quantities of variants at a warehouse to a demand, all lines or none', '01M3WHRSV00000000000000001', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3WHRSVACT00000000000003', jsonb_build_object('en-US', 'Check warehouse availability', 'vi-VN', 'Kiểm tra tồn kho khả dụng'), 'check_warehouse_availability', 'Read how much of a variant a warehouse can still commit, without promising it', '01M3WHRSV00000000000000001', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3WHRSVACT00000000000004', jsonb_build_object('en-US', 'Release reservation', 'vi-VN', 'Giải phóng hàng giữ chỗ'), 'release_reservation', 'Give back the unconsumed remainder of a reservation. Moves no stock', '01M3WHRSV00000000000000001', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3WHRSVACT00000000000005', jsonb_build_object('en-US', 'Consume reservation', 'vi-VN', 'Sử dụng hàng đã giữ chỗ'), 'consume_reservation', 'Record goods that physically left stock against a reservation, from the locations actually used', '01M3WHRSV00000000000000001', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;
