-- Seed inventory sample data (product types, categories, attributes, templates, variants);
-- aligns with 0005001_inventory_schema.sql. Org IDs match 0002002_iam_identity_seeds.sql.
--
-- IAM resources and actions for the Inventory module live in 0005002_inventory_iam.sql, not here,
-- so a reviewer can see the module's entire permission surface in one file.

DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'inventory_product_types'
	) THEN
		-- Product types
		INSERT INTO "inventory_product_types" ("id", "code", "name", "description", "supports_stock", "supports_sale", "supports_purchase", "supports_manufacturing", "is_archived", "created_at", "updated_at", "etag") VALUES
		('01K5INV0000000000000TYPE01', 'beverage', jsonb_build_object('en-US', 'Beverage', 'vi-VN', 'Đồ uống'), jsonb_build_object('en-US', 'Beverage products', 'vi-VN', 'Sản phẩm đồ uống'), TRUE, TRUE, TRUE, FALSE, FALSE, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5INV0000000000000TYPE02', 'snack', jsonb_build_object('en-US', 'Snack', 'vi-VN', 'Đồ ăn vặt'), jsonb_build_object('en-US', 'Snack products', 'vi-VN', 'Sản phẩm đồ ăn vặt'), TRUE, TRUE, TRUE, FALSE, FALSE, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;

		-- Product categories
		INSERT INTO "inventory_product_categories" ("id", "code", "name", "parent_category_id", "sequence", "description", "org_id", "is_archived", "created_at", "updated_at", "etag") VALUES
		('01K5INV000000000000CAT0001', 'beverages', jsonb_build_object('en-US', 'Beverages', 'vi-VN', 'Đồ uống'), NULL, 1, NULL, '01JWNY20G23KD4RV5VWYABQYHD', FALSE, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5INV000000000000CAT0002', 'snacks', jsonb_build_object('en-US', 'Snacks', 'vi-VN', 'Đồ ăn vặt'), NULL, 2, NULL, '01JWNY20G23KD4RV5VWYABQYHD', FALSE, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;

		-- Product attributes
		INSERT INTO "inventory_product_attributes" ("id", "code", "name", "data_type", "variant_creation_mode", "display_type", "sequence", "org_id", "is_archived", "created_at", "updated_at", "etag") VALUES
		('01K5INV0000000000ATTR00001', 'flavor', jsonb_build_object('en-US', 'Flavor', 'vi-VN', 'Hương vị'), 'text', 'instant', 'dropdown', 1, '01JWNY20G23KD4RV5VWYABQYHD', FALSE, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5INV0000000000ATTR00002', 'size', jsonb_build_object('en-US', 'Size', 'vi-VN', 'Kích thước'), 'text', 'instant', 'dropdown', 2, '01JWNY20G23KD4RV5VWYABQYHD', FALSE, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;

		-- Attribute values
		INSERT INTO "inventory_product_attribute_values" ("id", "attribute_id", "code", "name", "sequence", "price_extra", "org_id", "is_archived", "created_at", "updated_at", "etag") VALUES
		('01K5INV00000000ATTRVAL0001', '01K5INV0000000000ATTR00001', 'berry', jsonb_build_object('en-US', 'Berry', 'vi-VN', 'Dâu'), 1, NULL, '01JWNY20G23KD4RV5VWYABQYHD', FALSE, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5INV00000000ATTRVAL0002', '01K5INV0000000000ATTR00001', 'citrus', jsonb_build_object('en-US', 'Citrus', 'vi-VN', 'Cam chanh'), 2, NULL, '01JWNY20G23KD4RV5VWYABQYHD', FALSE, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5INV00000000ATTRVAL0003', '01K5INV0000000000ATTR00002', 'size_500ml', jsonb_build_object('en-US', '500 ml', 'vi-VN', '500 ml'), 1, NULL, '01JWNY20G23KD4RV5VWYABQYHD', FALSE, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5INV00000000ATTRVAL0004', '01K5INV0000000000ATTR00002', 'size_90g', jsonb_build_object('en-US', '90g', 'vi-VN', '90g'), 2, NULL, '01JWNY20G23KD4RV5VWYABQYHD', FALSE, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;

		-- Product templates
		INSERT INTO "inventory_product_templates" ("id", "name", "short_name", "product_type_id", "category_id", "brand_id", "sale_ok", "purchase_ok", "description", "sales_description", "purchase_description", "default_image_id", "default_weight", "default_length", "default_width", "default_height", "status", "org_id", "is_archived", "created_at", "updated_at", "etag") VALUES
		('01K5INV0000000TMPL00000001', jsonb_build_object('en-US', 'Energy Drink', 'vi-VN', 'Nước tăng lực'), NULL, '01K5INV0000000000000TYPE01', '01K5INV000000000000CAT0001', NULL, TRUE, TRUE, jsonb_build_object('en-US', 'Caffeinated beverage', 'vi-VN', 'Đồ uống có caffeine'), NULL, NULL, NULL, NULL, NULL, NULL, NULL, 'active', '01JWNY20G23KD4RV5VWYABQYHD', FALSE, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5INV0000000TMPL00000002', jsonb_build_object('en-US', 'Mineral Water', 'vi-VN', 'Nước khoáng'), NULL, '01K5INV0000000000000TYPE01', '01K5INV000000000000CAT0001', NULL, TRUE, TRUE, jsonb_build_object('en-US', 'Still water', 'vi-VN', 'Nước không ga'), NULL, NULL, NULL, NULL, NULL, NULL, NULL, 'active', '01JWNY20G23KD4RV5VWYABQYHD', FALSE, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5INV0000000TMPL00000003', jsonb_build_object('en-US', 'Potato Chips', 'vi-VN', 'Khoai tây chiên'), NULL, '01K5INV0000000000000TYPE02', '01K5INV000000000000CAT0002', NULL, TRUE, TRUE, jsonb_build_object('en-US', 'Original flavor chips', 'vi-VN', 'Khoai tây chiên vị tự nhiên'), NULL, NULL, NULL, NULL, NULL, NULL, NULL, 'active', '01JWNY20G23KD4RV5VWYABQYHD', FALSE, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;

		-- Product template attributes
		INSERT INTO "inventory_product_template_attributes" ("id", "product_template_id", "attribute_id", "sequence", "created_at", "updated_at", "etag") VALUES
		('01K5INV0000000TMPLATTR0001', '01K5INV0000000TMPL00000001', '01K5INV0000000000ATTR00001', 1, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5INV0000000TMPLATTR0002', '01K5INV0000000TMPL00000002', '01K5INV0000000000ATTR00002', 1, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5INV0000000TMPLATTR0003', '01K5INV0000000TMPL00000003', '01K5INV0000000000ATTR00002', 1, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;

		-- Product template attribute values
		INSERT INTO "inventory_product_template_attribute_values" ("id", "template_attribute_id", "attribute_value_id", "sequence", "created_at", "updated_at", "etag") VALUES
		('01K5INV000000TMPLAVVAL0001', '01K5INV0000000TMPLATTR0001', '01K5INV00000000ATTRVAL0001', 1, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5INV000000TMPLAVVAL0002', '01K5INV0000000TMPLATTR0001', '01K5INV00000000ATTRVAL0002', 2, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5INV000000TMPLAVVAL0003', '01K5INV0000000TMPLATTR0002', '01K5INV00000000ATTRVAL0003', 1, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5INV000000TMPLAVVAL0004', '01K5INV0000000TMPLATTR0003', '01K5INV00000000ATTRVAL0004', 1, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;

		-- Product variants
		INSERT INTO "inventory_product_variants" ("id", "product_template_id", "combination_key", "sku", "primary_barcode", "is_materialized", "variant_image_id", "weight", "length", "width", "height", "status", "archive_source", "org_id", "is_archived", "created_at", "updated_at", "etag") VALUES
		('01K5INV00000000VARIANT0001', '01K5INV0000000TMPL00000001', 'flavor:berry', 'INV-SEED-ED-BERRY', NULL, FALSE, NULL, NULL, NULL, NULL, NULL, 'active', NULL, '01JWNY20G23KD4RV5VWYABQYHD', FALSE, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5INV00000000VARIANT0002', '01K5INV0000000TMPL00000001', 'flavor:citrus', 'INV-SEED-ED-CITRUS', NULL, FALSE, NULL, NULL, NULL, NULL, NULL, 'active', NULL, '01JWNY20G23KD4RV5VWYABQYHD', FALSE, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5INV00000000VARIANT0003', '01K5INV0000000TMPL00000002', 'size:size_500ml', 'INV-SEED-WATER-500', NULL, FALSE, NULL, NULL, NULL, NULL, NULL, 'active', NULL, '01JWNY20G23KD4RV5VWYABQYHD', FALSE, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5INV00000000VARIANT0004', '01K5INV0000000TMPL00000003', 'size:size_90g', 'INV-SEED-CHIPS-ORG', NULL, FALSE, NULL, NULL, NULL, NULL, NULL, 'active', NULL, '01JWNY20G23KD4RV5VWYABQYHD', FALSE, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;

		-- Product variant attribute values
		INSERT INTO "inventory_product_variant_attribute_values" ("id", "product_variant_id", "template_attribute_value_id", "created_at", "updated_at", "etag") VALUES
		('01K5INV0000000VARATVAL0001', '01K5INV00000000VARIANT0001', '01K5INV000000TMPLAVVAL0001', NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5INV0000000VARATVAL0002', '01K5INV00000000VARIANT0002', '01K5INV000000TMPLAVVAL0002', NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5INV0000000VARATVAL0003', '01K5INV00000000VARIANT0003', '01K5INV000000TMPLAVVAL0003', NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01K5INV0000000VARATVAL0004', '01K5INV00000000VARIANT0004', '01K5INV000000TMPLAVVAL0004', NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;
