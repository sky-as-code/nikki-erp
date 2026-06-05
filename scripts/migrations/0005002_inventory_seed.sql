-- Seed inventory (product, category, attributes, variants) — aligns with
-- nikkierp/modules/inventory/product/domain and 0005001_inventory _schema.sql
-- Tenant/org IDs match coremart/scripts/migrations/0002002_identity_seeds.sql

DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'inventory_products'
	) THEN
		INSERT INTO "inventory_product_categories" (
			"id",  "org_id", "name", "etag", "created_at", "updated_at"
		) VALUES
		(
			'01K5INV0000000000000000101',
			'01JWNY20G23KD4RV5VWYABQYHD',
			jsonb_build_object('en-US', 'Beverages', 'vi-VN', 'Đồ uống'),
			(EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text,
			NOW(),
			NULL
		),
		(
			'01K5INV0000000000000000011',  -- ✅ 25→26
			'01JWNY20G23KD4RV5VWYABQYHD',
			jsonb_build_object('en-US', 'Snacks', 'vi-VN', 'Đồ ăn vặt'),
			(EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text,
			NOW(),
			NULL
		);

		INSERT INTO "inventory_products" (
			"id",  "org_id", "name", "description", "unit_id", "default_variant_id",
			"tag_ids", "is_archived", "etag", "created_at", "updated_at"
		) VALUES
		(
			'01K5INV0000000000000000011',
			'01JWNY20G23KD4RV5VWYABQYHD',
			jsonb_build_object('en-US', 'Energy Drink', 'vi-VN', 'Nước tăng lực'),
			jsonb_build_object('en-US', 'Caffeinated beverage', 'vi-VN', 'Đồ uống có caffeine'),
			NULL,
			NULL,
			NULL,
			FALSE,
			(EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text,
			NOW(),
			NULL
		),
		(
			'01K5INV0000000000000000021',
			'01JWNY20G23KD4RV5VWYABQYHD',
			jsonb_build_object('en-US', 'Mineral Water', 'vi-VN', 'Nước khoáng'),
			jsonb_build_object('en-US', 'Still water 500ml', 'vi-VN', 'Nước không ga 500ml'),
			NULL,
			NULL,
			NULL,
			FALSE,
			(EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text,
			NOW(),
			NULL
		),
		(
			'01K5INV0000000000000000031',
			'01JWNY20G23KD4RV5VWYABQYHD',
			jsonb_build_object('en-US', 'Potato Chips', 'vi-VN', 'Khoai tây chiên'),
			jsonb_build_object('en-US', 'Original flavor', 'vi-VN', 'Vị tự nhiên'),
			NULL,
			NULL,
			NULL,
			FALSE,
			(EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text,
			NOW(),
			NULL
		);

		INSERT INTO "inventory_product_category_rel" (
			 "product_id", "product_category_id"
		) VALUES
		('01K5INV0000000000000000011', '01K5INV0000000000000000101'),
		('01K5INV0000000000000000021', '01K5INV0000000000000000101'),
		('01K5INV0000000000000000031', '01K5INV0000000000000000011');  -- ✅ 25→26

		INSERT INTO "inventory_attribute_groups" (
			"id",  "name", "index", "product_id", "etag", "created_at", "updated_at"
		) VALUES
		(
			'01K5INV0000000000000000020',  -- ✅ 25→26
			'Main specifications',
			0,
			'01K5INV0000000000000000011',
			(EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text,
			NOW(),
			NULL
		);

		INSERT INTO "inventory_attributes" (
			"id", "etag",  "code_name", "display_name", "sort_index", "data_type",
			"is_required", "is_enum", "enum_value_sort", "attribute_group_id", "product_id", "is_archived", "created_at"
		) VALUES
		(
			'01K5INV0000000000000000030',  -- ✅ 25→26
			(EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text,
			'flavor',
			jsonb_build_object('en-US', 'Flavor', 'vi-VN', 'Hương vị'),
			0,
			'text',
			TRUE,
			TRUE,
			FALSE,
			'01K5INV0000000000000000020',  -- ✅ 25→26
			'01K5INV0000000000000000011',
			FALSE,
			NOW()
		);

		INSERT INTO "inventory_attribute_values" (
			"id",  "attribute_id", "value_text", "value_decimal", "value_integer", "value_bool", "value_ref", "variant_count", "product_id"
		) VALUES
		(
			'01K5INV0000000000000000040',  -- ✅ 25→26
			'01K5INV0000000000000000030',  -- ✅ 25→26
			jsonb_build_object('en-US', 'Berry', 'vi-VN', 'Dâu'),
			NULL,
			NULL,
			NULL,
			NULL,
			1,
			'01K5INV0000000000000000011'
		),
		(
			'01K5INV0000000000000000041',  -- ✅ 25→26
			'01K5INV0000000000000000030',  -- ✅ 25→26
			jsonb_build_object('en-US', 'Citrus', 'vi-VN', 'Cam chanh'),
			NULL,
			NULL,
			NULL,
			NULL,
			1,
			'01K5INV0000000000000000011'
		);

		INSERT INTO "inventory_variants" (
			"id",  "org_id", "product_id", "name", "sku", "barcode", "proposed_price", "status",
			"etag", "created_at", "updated_at"
		) VALUES
		(
			'01K5INV0000000000000000500',
			'01JWNY20G23KD4RV5VWYABQYHD',
			'01K5INV0000000000000000011',
			jsonb_build_object('en-US', 'Berry', 'vi-VN', 'Dâu'),
			'INV-SEED-ED-BERRY',
			NULL,
			15000,
			'active',
			(EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text,
			NOW(),
			NULL
		),
		(
			'01K5INV0000000000000000510',
			'01JWNY20G23KD4RV5VWYABQYHD',
			'01K5INV0000000000000000011',
			jsonb_build_object('en-US', 'Citrus', 'vi-VN', 'Cam chanh'),
			'INV-SEED-ED-CITRUS',
			NULL,
			15000,
			'active',
			(EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text,
			NOW(),
			NULL
		),
		(
			'01K5INV0000000000000000520',
			'01JWNY20G23KD4RV5VWYABQYHD',
			'01K5INV0000000000000000021',
			jsonb_build_object('en-US', '500 ml bottle', 'vi-VN', 'Chai 500 ml'),
			'INV-SEED-WATER-500',
			NULL,
			5000,
			'active',
			(EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text,
			NOW(),
			NULL
		),
		(
			'01K5INV0000000000000000530',
			'01JWNY20G23KD4RV5VWYABQYHD',
			'01K5INV0000000000000000031',
			jsonb_build_object('en-US', 'Original 90g', 'vi-VN', 'Tự nhiên 90g'),
			'INV-SEED-CHIPS-ORG',
			NULL,
			12000,
			'active',
			(EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text,
			NOW(),
			NULL
		);

		UPDATE "inventory_products"
		SET "default_variant_id" = v."id", "updated_at" = NOW()
		FROM (
			VALUES
				('01K5INV0000000000000000011'::varchar, '01K5INV0000000000000000500'::varchar),  -- ✅ 25→26
				('01K5INV0000000000000000021'::varchar, '01K5INV0000000000000000520'::varchar),  -- ✅ 25→26
				('01K5INV0000000000000000031'::varchar, '01K5INV0000000000000000530'::varchar)   -- ✅ 25→26
		) AS v(prod_id, id)
		WHERE "inventory_products"."id" = v.prod_id;

		INSERT INTO "inventory_variant_attr_val_rel" (
			 "variant_id", "attribute_value_id"
		) VALUES
		('01K5INV0000000000000000500', '01K5INV0000000000000000040'),  -- ✅ 25→26
		('01K5INV0000000000000000510', '01K5INV0000000000000000041');  -- ✅ 25→26
	END IF;
END $$;
