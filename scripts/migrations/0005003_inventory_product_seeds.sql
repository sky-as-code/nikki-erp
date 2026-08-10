-- IAM resources and actions for the Inventory Products submodule (product BR §6, §12).
--
-- The dynamic resource engine asserts permissions using the schema name as the resource code,
-- so these codes must stay identical to the "inventory_product_template" /
-- "inventory_product_variant" / ... schema names.
--
-- Deliberately no iam_entitlements rows. Unit of Measure grants the system "User" role a
-- domain-wide read so that any user can pick a unit while filling in an unrelated form. Product
-- master data is not universally readable: access to it follows explicitly assigned roles, so
-- there is no blanket grant here.

DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_resources'
	) THEN
		INSERT INTO "iam_resources" (
			"id", "name", "code", "description", "owner_type", "max_scope", "min_scope", "created_at", "etag"
		) VALUES
		('01M0A1P4QM3P0Y3T60T7RH2KZC', 'Product Template', 'inventory_product_template', 'Catalog-level definition of a product line', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0A1P4QJ0MFYJY0EQH0NVJ2R', 'Product Variant', 'inventory_product_variant', 'Concrete, transactable product of a template', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0A1P4Q8S7V5FR3KKY1V5ZMF', 'Product Type', 'inventory_product_type', 'How the system processes a product: goods, service, combo', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0A1P4QGGABP6SVFCTD6MF8K', 'Product Category', 'inventory_product_category', 'Hierarchical classification of product templates', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0A1P4QMJTAEFCJY5WRTNHWZ', 'Product Attribute', 'inventory_product_attribute', 'Attributes whose values form product variants', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0A1P4QAH9DQW7VXG9K87HWP', 'Product Attribute Value', 'inventory_product_attribute_value', 'Allowed values of a product attribute', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0A1P4QDXYWAJ8D4ZAD50VW7', 'Brand', 'inventory_brand', 'Product brands', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text);
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_actions'
	) THEN
		INSERT INTO "iam_actions" ("id", "name", "code", "description", "resource_id", "etag") VALUES
		-- Product Template. generate_variants is separate from update: regenerating the
		-- variant set of a live product is a heavier capability than editing its description.
		('01M0A1P4Q4K8REGFBYKH82M8Z1', 'Create', 'create', NULL, '01M0A1P4QM3P0Y3T60T7RH2KZC', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0A1P4Q8WKR2A0RRQ21W6ZKT', 'Update', 'update', NULL, '01M0A1P4QM3P0Y3T60T7RH2KZC', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0A1P4QQ799ZX5PPHC4RRN9F', 'Delete', 'delete', NULL, '01M0A1P4QM3P0Y3T60T7RH2KZC', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0A1P4Q4XDRGWA4DDDNKD2DM', 'Read', 'read', NULL, '01M0A1P4QM3P0Y3T60T7RH2KZC', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0A1P4QS5PTBGTJSN1GS991H', 'Set archived status', 'set_archived', 'Archive a product template so it and its variants cannot be used by new transactions', '01M0A1P4QM3P0Y3T60T7RH2KZC', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0A1P4QBXY1DQFZPP4P4VD1T', 'Generate variants', 'generate_variants', 'Generate or synchronize the variants of a product template', '01M0A1P4QM3P0Y3T60T7RH2KZC', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),

		-- Product Variant
		('01M0A1P4Q9YQ7HAFRENHBMJ9TR', 'Create', 'create', NULL, '01M0A1P4QJ0MFYJY0EQH0NVJ2R', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0A1P4Q4P4J5K2NANT4K78VK', 'Update', 'update', NULL, '01M0A1P4QJ0MFYJY0EQH0NVJ2R', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0A1P4QB7BD6NB8M9SEK5Y20', 'Delete', 'delete', NULL, '01M0A1P4QJ0MFYJY0EQH0NVJ2R', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0A1P4QSQMEVAZ1FFTAVA317', 'Read', 'read', NULL, '01M0A1P4QJ0MFYJY0EQH0NVJ2R', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0A1P4QEVX39CJ2TJ2AN97FT', 'Set archived status', 'set_archived', 'Archive a product variant so it cannot be used by new transactions', '01M0A1P4QJ0MFYJY0EQH0NVJ2R', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),

		-- Product Type
		('01M0A1P4QCYXK4QQ5PP8K6EQ0Y', 'Create', 'create', NULL, '01M0A1P4Q8S7V5FR3KKY1V5ZMF', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0A1P4QR944RNJFF7716BF5C', 'Update', 'update', NULL, '01M0A1P4Q8S7V5FR3KKY1V5ZMF', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0A1P4Q2QJ2Y2TYW6XXX9BRN', 'Delete', 'delete', NULL, '01M0A1P4Q8S7V5FR3KKY1V5ZMF', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0A1P4QCVHQCN7ECG2TCKJ97', 'Read', 'read', NULL, '01M0A1P4Q8S7V5FR3KKY1V5ZMF', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0A1P4QPQSBNB8VK8F1H6XM3', 'Set archived status', 'set_archived', 'Archive a product type', '01M0A1P4Q8S7V5FR3KKY1V5ZMF', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),

		-- Product Category
		('01M0A1P4QB6H18345VP8SG849J', 'Create', 'create', NULL, '01M0A1P4QGGABP6SVFCTD6MF8K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0A1P4Q4QZBR3FPNJJ1V9FRJ', 'Update', 'update', NULL, '01M0A1P4QGGABP6SVFCTD6MF8K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0A1P4QMSFDXYBN3D3RS9K5E', 'Delete', 'delete', NULL, '01M0A1P4QGGABP6SVFCTD6MF8K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0A1P4QYG8G9QC4EWTWAV723', 'Read', 'read', NULL, '01M0A1P4QGGABP6SVFCTD6MF8K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0A1P4QWZHYXV89G9BT6XJBE', 'Set archived status', 'set_archived', 'Archive a product category', '01M0A1P4QGGABP6SVFCTD6MF8K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),

		-- Product Attribute
		('01M0A1P4Q94CATJA4MXP6SED7N', 'Create', 'create', NULL, '01M0A1P4QMJTAEFCJY5WRTNHWZ', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0A1P4QXAG5J1BV1JSP6W6G9', 'Update', 'update', NULL, '01M0A1P4QMJTAEFCJY5WRTNHWZ', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0A1P4Q95527SM8SJK2QRH4R', 'Delete', 'delete', NULL, '01M0A1P4QMJTAEFCJY5WRTNHWZ', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0A1P4QFDM0XSX8VXAHZ31EG', 'Read', 'read', NULL, '01M0A1P4QMJTAEFCJY5WRTNHWZ', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0A1P4QVNXPS112TGMJTTHTJ', 'Set archived status', 'set_archived', 'Archive a product attribute', '01M0A1P4QMJTAEFCJY5WRTNHWZ', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),

		-- Product Attribute Value
		('01M0A1P4QKK2BS2M7Q4ZJVX3ED', 'Create', 'create', NULL, '01M0A1P4QAH9DQW7VXG9K87HWP', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0A1P4QBJYDR8D8Y1EESRH4W', 'Update', 'update', NULL, '01M0A1P4QAH9DQW7VXG9K87HWP', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0A1P4Q71ATJEWZM1BAM7E3N', 'Delete', 'delete', NULL, '01M0A1P4QAH9DQW7VXG9K87HWP', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0A1P4Q4QF419T7W52WS0HMF', 'Read', 'read', NULL, '01M0A1P4QAH9DQW7VXG9K87HWP', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0A1P4QEYDQTRE7QH9NHGQA3', 'Set archived status', 'set_archived', 'Archive a product attribute value', '01M0A1P4QAH9DQW7VXG9K87HWP', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),

		-- Brand
		('01M0A1P4QKBPP966KQK6NDTY5P', 'Create', 'create', NULL, '01M0A1P4QDXYWAJ8D4ZAD50VW7', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0A1P4QXRE54C2AJPY554FD9', 'Update', 'update', NULL, '01M0A1P4QDXYWAJ8D4ZAD50VW7', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0A1P4QCM2KQ3VNHZVHGV45P', 'Delete', 'delete', NULL, '01M0A1P4QDXYWAJ8D4ZAD50VW7', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0A1P4QRS1VZA44V25595WBX', 'Read', 'read', NULL, '01M0A1P4QDXYWAJ8D4ZAD50VW7', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0A1P4Q67457GP4WXDW09M6F', 'Set archived status', 'set_archived', 'Archive a brand', '01M0A1P4QDXYWAJ8D4ZAD50VW7', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text);
	END IF;
END $$;
