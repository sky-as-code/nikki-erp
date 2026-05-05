-- Seeds for inventory: products, attributes, values, variants, categories
-- Uses same etag/ts pattern as identity seeds
DO $$
BEGIN
	-- Products (5 sample products)
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'inventory_products'
	) THEN
		INSERT INTO "inventory_products" ("id", "org_id", "name", "description", "thumbnail_url", "unit_id", "default_variant_id", "tag_ids", "is_archived", "etag", "created_at", "updated_at") VALUES
		('01INVPRD0001', '01JWNY20G23KD4RV5VWYABQYHD', jsonb_build_object('en-US','Basic T-Shirt','vi-VN','Áo thun cơ bản'), jsonb_build_object('en-US','Cotton t-shirt, unisex','vi-VN','Áo cotton unisex'), 'https://example.com/images/tshirt-basic.png', NULL, NULL, NULL, FALSE, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL),
		('01INVPRD0002', '01JWNY20G23KD4RV5VWYABQYHD', jsonb_build_object('en-US','Denim Jeans','vi-VN','Quần jean'), jsonb_build_object('en-US','Slim fit denim jeans','vi-VN','Quần jean ôm'), 'https://example.com/images/jeans.png', NULL, NULL, NULL, FALSE, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL),
		('01INVPRD0003', '01JWNY20G23KD4RV5VWYABQYHD', jsonb_build_object('en-US','Hoodie','vi-VN','Áo hoodie'), jsonb_build_object('en-US','Warm fleece hoodie','vi-VN','Áo hoodie nỉ ấm'), 'https://example.com/images/hoodie.png', NULL, NULL, NULL, FALSE, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL),
		('01INVPRD0004', '01JWNY20G23KD4RV5VWYABQYHD', jsonb_build_object('en-US','Sneakers','vi-VN','Giày thể thao'), jsonb_build_object('en-US','Comfort running sneakers','vi-VN','Giày chạy bộ thoải mái'), 'https://example.com/images/sneakers.png', NULL, NULL, NULL, FALSE, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL),
		('01INVPRD0005', '01JWNY20G23KD4RV5VWYABQYHD', jsonb_build_object('en-US','Baseball Cap','vi-VN','Mũ lưỡi trai'), jsonb_build_object('en-US','Adjustable cotton cap','vi-VN','Mũ cotton có thể điều chỉnh'), 'https://example.com/images/cap.png', NULL, NULL, NULL, FALSE, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL);
	END IF;

	-- Attributes: 5 attributes per product (total 25)
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'inventory_attributes'
	) THEN
		INSERT INTO "inventory_attributes" ("id", "code_name", "display_name", "sort_index", "data_type", "is_required", "is_enum", "enum_value_sort", "enum_value_text", "attribute_group_id", "product_id", "is_archived", "etag", "created_at", "updated_at") VALUES
		-- Product 1 attributes
		('01INVATTR0001','prd1_size', jsonb_build_object('en-US','Size','vi-VN','Kích thước'), 1, 'text', TRUE, TRUE, FALSE, ARRAY['"S"'::jsonb,'"M"'::jsonb,'"L"'::jsonb,'"XL"'::jsonb,'"XXL"'::jsonb]::jsonb[], NULL, '01INVPRD0001', FALSE, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL),
		('01INVATTR0002','prd1_color', jsonb_build_object('en-US','Color','vi-VN','Màu'), 2, 'text', FALSE, TRUE, FALSE, ARRAY['"Blue"'::jsonb,'"Black"'::jsonb,'"White"'::jsonb,'"Red"'::jsonb,'"Green"'::jsonb]::jsonb[], NULL, '01INVPRD0001', FALSE, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL),
		('01INVATTR0003','prd1_material', jsonb_build_object('en-US','Material','vi-VN','Chất liệu'), 3, 'text', FALSE, TRUE, FALSE, ARRAY['"Cotton"'::jsonb,'"Denim"'::jsonb,'"Polyester"'::jsonb,'"Leather"'::jsonb,'"Wool"'::jsonb]::jsonb[], NULL, '01INVPRD0001', FALSE, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL),
		('01INVATTR0004','prd1_brand', jsonb_build_object('en-US','Brand','vi-VN','Nhãn hiệu'), 4, 'text', FALSE, FALSE, FALSE, NULL, NULL, '01INVPRD0001', FALSE, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL),
		('01INVATTR0005','prd1_gender', jsonb_build_object('en-US','Gender','vi-VN','Giới tính'), 5, 'text', FALSE, TRUE, FALSE, ARRAY['"Unisex"'::jsonb,'"Men"'::jsonb,'"Women"'::jsonb,'"Kids"'::jsonb,'"All"'::jsonb]::jsonb[], NULL, '01INVPRD0001', FALSE, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL),
		-- Product 2 attributes
		('01INVATTR0006','prd2_size', jsonb_build_object('en-US','Size','vi-VN','Kích thước'), 1, 'text', TRUE, TRUE, FALSE, ARRAY['"28"'::jsonb,'"30"'::jsonb,'"32"'::jsonb,'"34"'::jsonb,'"36"'::jsonb]::jsonb[], NULL, '01INVPRD0002', FALSE, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL),
		('01INVATTR0007','prd2_color', jsonb_build_object('en-US','Color','vi-VN','Màu'), 2, 'text', FALSE, TRUE, FALSE, ARRAY['"Blue"'::jsonb,'"Black"'::jsonb,'"Grey"'::jsonb,'"White"'::jsonb,'"Indigo"'::jsonb]::jsonb[], NULL, '01INVPRD0002', FALSE, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL),
		('01INVATTR0008','prd2_material', jsonb_build_object('en-US','Material','vi-VN','Chất liệu'), 3, 'text', FALSE, TRUE, FALSE, ARRAY['"Denim"'::jsonb,'"Cotton"'::jsonb,'"Elastane"'::jsonb,'"Polyester"'::jsonb,'"Blend"'::jsonb]::jsonb[], NULL, '01INVPRD0002', FALSE, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL),
		('01INVATTR0009','prd2_brand', jsonb_build_object('en-US','Brand','vi-VN','Nhãn hiệu'), 4, 'text', FALSE, FALSE, FALSE, NULL, NULL, '01INVPRD0002', FALSE, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL),
		('01INVATTR0010','prd2_gender', jsonb_build_object('en-US','Gender','vi-VN','Giới tính'), 5, 'text', FALSE, TRUE, FALSE, ARRAY['"Men"'::jsonb,'"Women"'::jsonb,'"Unisex"'::jsonb,'"Boy"'::jsonb,'"Girl"'::jsonb]::jsonb[], NULL, '01INVPRD0002', FALSE, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL),
		-- Product 3 attributes
		('01INVATTR0011','prd3_size', jsonb_build_object('en-US','Size','vi-VN','Kích thước'), 1, 'text', TRUE, TRUE, FALSE, ARRAY['"S"'::jsonb,'"M"'::jsonb,'"L"'::jsonb,'"XL"'::jsonb,'"XXL"'::jsonb]::jsonb[], NULL, '01INVPRD0003', FALSE, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL),
		('01INVATTR0012','prd3_color', jsonb_build_object('en-US','Color','vi-VN','Màu'), 2, 'text', FALSE, TRUE, FALSE, ARRAY['"Black"'::jsonb,'"Grey"'::jsonb,'"Navy"'::jsonb,'"White"'::jsonb,'"Maroon"'::jsonb]::jsonb[], NULL, '01INVPRD0003', FALSE, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL),
		('01INVATTR0013','prd3_material', jsonb_build_object('en-US','Material','vi-VN','Chất liệu'), 3, 'text', FALSE, TRUE, FALSE, ARRAY['"Fleece"'::jsonb,'"Cotton"'::jsonb,'"Polyester"'::jsonb,'"Blend"'::jsonb,'"Wool"'::jsonb]::jsonb[], NULL, '01INVPRD0003', FALSE, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL),
		('01INVATTR0014','prd3_brand', jsonb_build_object('en-US','Brand','vi-VN','Nhãn hiệu'), 4, 'text', FALSE, FALSE, FALSE, NULL, NULL, '01INVPRD0003', FALSE, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL),
		('01INVATTR0015','prd3_gender', jsonb_build_object('en-US','Gender','vi-VN','Giới tính'), 5, 'text', FALSE, TRUE, FALSE, ARRAY['"Unisex"'::jsonb,'"Men"'::jsonb,'"Women"'::jsonb,'"Kids"'::jsonb,'"All"'::jsonb]::jsonb[], NULL, '01INVPRD0003', FALSE, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL),
		-- Product 4 attributes
		('01INVATTR0016','prd4_size', jsonb_build_object('en-US','Size','vi-VN','Kích thước'), 1, 'text', TRUE, TRUE, FALSE, ARRAY['"40"'::jsonb,'"41"'::jsonb,'"42"'::jsonb,'"43"'::jsonb,'"44"'::jsonb]::jsonb[], NULL, '01INVPRD0004', FALSE, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL),
		('01INVATTR0017','prd4_color', jsonb_build_object('en-US','Color','vi-VN','Màu'), 2, 'text', FALSE, TRUE, FALSE, ARRAY['"White"'::jsonb,'"Black"'::jsonb,'"Red"'::jsonb,'"Blue"'::jsonb,'"Yellow"'::jsonb]::jsonb[], NULL, '01INVPRD0004', FALSE, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL),
		('01INVATTR0018','prd4_material', jsonb_build_object('en-US','Material','vi-VN','Chất liệu'), 3, 'text', FALSE, TRUE, FALSE, ARRAY['"Mesh"'::jsonb,'"Leather"'::jsonb,'"Canvas"'::jsonb,'"Synthetic"'::jsonb,'"Rubber"'::jsonb]::jsonb[], NULL, '01INVPRD0004', FALSE, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL),
		('01INVATTR0019','prd4_brand', jsonb_build_object('en-US','Brand','vi-VN','Nhãn hiệu'), 4, 'text', FALSE, FALSE, FALSE, NULL, NULL, '01INVPRD0004', FALSE, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL),
		('01INVATTR0020','prd4_gender', jsonb_build_object('en-US','Gender','vi-VN','Giới tính'), 5, 'text', FALSE, TRUE, FALSE, ARRAY['"Unisex"'::jsonb,'"Men"'::jsonb,'"Women"'::jsonb,'"Kids"'::jsonb,'"All"'::jsonb]::jsonb[], NULL, '01INVPRD0004', FALSE, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL),
		-- Product 5 attributes
		('01INVATTR0021','prd5_size', jsonb_build_object('en-US','Size','vi-VN','Kích thước'), 1, 'text', TRUE, TRUE, FALSE, ARRAY['"OneSize"'::jsonb,'"Adjustable"'::jsonb,'"S"'::jsonb,'"M"'::jsonb,'"L"'::jsonb]::jsonb[], NULL, '01INVPRD0005', FALSE, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL),
		('01INVATTR0022','prd5_color', jsonb_build_object('en-US','Color','vi-VN','Màu'), 2, 'text', FALSE, TRUE, FALSE, ARRAY['"Black"'::jsonb,'"Navy"'::jsonb,'"Grey"'::jsonb,'"White"'::jsonb,'"Brown"'::jsonb]::jsonb[], NULL, '01INVPRD0005', FALSE, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL),
		('01INVATTR0023','prd5_material', jsonb_build_object('en-US','Material','vi-VN','Chất liệu'), 3, 'text', FALSE, TRUE, FALSE, ARRAY['"Cotton"'::jsonb,'"Polyester"'::jsonb,'"Wool"'::jsonb,'"Leather"'::jsonb,'"Blend"'::jsonb]::jsonb[], NULL, '01INVPRD0005', FALSE, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL),
		('01INVATTR0024','prd5_brand', jsonb_build_object('en-US','Brand','vi-VN','Nhãn hiệu'), 4, 'text', FALSE, FALSE, FALSE, NULL, NULL, '01INVPRD0005', FALSE, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL),
		('01INVATTR0025','prd5_gender', jsonb_build_object('en-US','Gender','vi-VN','Giới tính'), 5, 'text', FALSE, TRUE, FALSE, ARRAY['"Unisex"'::jsonb,'"Men"'::jsonb,'"Women"'::jsonb,'"Kids"'::jsonb,'"All"'::jsonb]::jsonb[], NULL, '01INVPRD0005', FALSE, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL);
	END IF;

	-- Attribute values: each attribute gets 5 possible values (25 values per product)
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'inventory_attribute_values'
	) THEN
		INSERT INTO "inventory_attribute_values" ("id", "attribute_id", "product_id", "value_text", "variant_count") VALUES
		-- Product 1 values (25)
		('01INVAL0001','01INVATTR0001','01INVPRD0001','"S"'::jsonb,0),
		('01INVAL0002','01INVATTR0001','01INVPRD0001','"M"'::jsonb,0),
		('01INVAL0003','01INVATTR0001','01INVPRD0001','"L"'::jsonb,0),
		('01INVAL0004','01INVATTR0001','01INVPRD0001','"XL"'::jsonb,0),
		('01INVAL0005','01INVATTR0001','01INVPRD0001','"XXL"'::jsonb,0),
		('01INVAL0006','01INVATTR0002','01INVPRD0001','"Blue"'::jsonb,0),
		('01INVAL0007','01INVATTR0002','01INVPRD0001','"Black"'::jsonb,0),
		('01INVAL0008','01INVATTR0002','01INVPRD0001','"White"'::jsonb,0),
		('01INVAL0009','01INVATTR0002','01INVPRD0001','"Red"'::jsonb,0),
		('01INVAL0010','01INVATTR0002','01INVPRD0001','"Green"'::jsonb,0),
		('01INVAL0011','01INVATTR0003','01INVPRD0001','"Cotton"'::jsonb,0),
		('01INVAL0012','01INVATTR0003','01INVPRD0001','"Denim"'::jsonb,0),
		('01INVAL0013','01INVATTR0003','01INVPRD0001','"Polyester"'::jsonb,0),
		('01INVAL0014','01INVATTR0003','01INVPRD0001','"Leather"'::jsonb,0),
		('01INVAL0015','01INVATTR0003','01INVPRD0001','"Wool"'::jsonb,0),
		('01INVAL0016','01INVATTR0004','01INVPRD0001','"BrandA"'::jsonb,0),
		('01INVAL0017','01INVATTR0004','01INVPRD0001','"BrandB"'::jsonb,0),
		('01INVAL0018','01INVATTR0004','01INVPRD0001','"BrandC"'::jsonb,0),
		('01INVAL0019','01INVATTR0004','01INVPRD0001','"BrandD"'::jsonb,0),
		('01INVAL0020','01INVATTR0004','01INVPRD0001','"BrandE"'::jsonb,0),
		('01INVAL0021','01INVATTR0005','01INVPRD0001','"Unisex"'::jsonb,0),
		('01INVAL0022','01INVATTR0005','01INVPRD0001','"Men"'::jsonb,0),
		('01INVAL0023','01INVATTR0005','01INVPRD0001','"Women"'::jsonb,0),
		('01INVAL0024','01INVATTR0005','01INVPRD0001','"Kids"'::jsonb,0),
		('01INVAL0025','01INVATTR0005','01INVPRD0001','"All"'::jsonb,0),
		-- Product 2 values (26..50)
		('01INVAL0026','01INVATTR0006','01INVPRD0002','"28"'::jsonb,0),
		('01INVAL0027','01INVATTR0006','01INVPRD0002','"30"'::jsonb,0),
		('01INVAL0028','01INVATTR0006','01INVPRD0002','"32"'::jsonb,0),
		('01INVAL0029','01INVATTR0006','01INVPRD0002','"34"'::jsonb,0),
		('01INVAL0030','01INVATTR0006','01INVPRD0002','"36"'::jsonb,0),
		('01INVAL0031','01INVATTR0007','01INVPRD0002','"Blue"'::jsonb,0),
		('01INVAL0032','01INVATTR0007','01INVPRD0002','"Black"'::jsonb,0),
		('01INVAL0033','01INVATTR0007','01INVPRD0002','"Grey"'::jsonb,0),
		('01INVAL0034','01INVATTR0007','01INVPRD0002','"White"'::jsonb,0),
		('01INVAL0035','01INVATTR0007','01INVPRD0002','"Indigo"'::jsonb,0),
		('01INVAL0036','01INVATTR0008','01INVPRD0002','"Denim"'::jsonb,0),
		('01INVAL0037','01INVATTR0008','01INVPRD0002','"Cotton"'::jsonb,0),
		('01INVAL0038','01INVATTR0008','01INVPRD0002','"Elastane"'::jsonb,0),
		('01INVAL0039','01INVATTR0008','01INVPRD0002','"Polyester"'::jsonb,0),
		('01INVAL0040','01INVATTR0008','01INVPRD0002','"Blend"'::jsonb,0),
		('01INVAL0041','01INVATTR0009','01INVPRD0002','"BrandF"'::jsonb,0),
		('01INVAL0042','01INVATTR0009','01INVPRD0002','"BrandG"'::jsonb,0),
		('01INVAL0043','01INVATTR0009','01INVPRD0002','"BrandH"'::jsonb,0),
		('01INVAL0044','01INVATTR0009','01INVPRD0002','"BrandI"'::jsonb,0),
		('01INVAL0045','01INVATTR0009','01INVPRD0002','"BrandJ"'::jsonb,0),
		('01INVAL0046','01INVATTR0010','01INVPRD0002','"Men"'::jsonb,0),
		('01INVAL0047','01INVATTR0010','01INVPRD0002','"Women"'::jsonb,0),
		('01INVAL0048','01INVATTR0010','01INVPRD0002','"Unisex"'::jsonb,0),
		('01INVAL0049','01INVATTR0010','01INVPRD0002','"Boy"'::jsonb,0),
		('01INVAL0050','01INVATTR0010','01INVPRD0002','"Girl"'::jsonb,0),
		-- Product 3 values (51..75)
		('01INVAL0051','01INVATTR0011','01INVPRD0003','"S"'::jsonb,0),
		('01INVAL0052','01INVATTR0011','01INVPRD0003','"M"'::jsonb,0),
		('01INVAL0053','01INVATTR0011','01INVPRD0003','"L"'::jsonb,0),
		('01INVAL0054','01INVATTR0011','01INVPRD0003','"XL"'::jsonb,0),
		('01INVAL0055','01INVATTR0011','01INVPRD0003','"XXL"'::jsonb,0),
		('01INVAL0056','01INVATTR0012','01INVPRD0003','"Black"'::jsonb,0),
		('01INVAL0057','01INVATTR0012','01INVPRD0003','"Grey"'::jsonb,0),
		('01INVAL0058','01INVATTR0012','01INVPRD0003','"Navy"'::jsonb,0),
		('01INVAL0059','01INVATTR0012','01INVPRD0003','"White"'::jsonb,0),
		('01INVAL0060','01INVATTR0012','01INVPRD0003','"Maroon"'::jsonb,0),
		('01INVAL0061','01INVATTR0013','01INVPRD0003','"Fleece"'::jsonb,0),
		('01INVAL0062','01INVATTR0013','01INVPRD0003','"Cotton"'::jsonb,0),
		('01INVAL0063','01INVATTR0013','01INVPRD0003','"Polyester"'::jsonb,0),
		('01INVAL0064','01INVATTR0013','01INVPRD0003','"Blend"'::jsonb,0),
		('01INVAL0065','01INVATTR0013','01INVPRD0003','"Wool"'::jsonb,0),
		('01INVAL0066','01INVATTR0014','01INVPRD0003','"BrandK"'::jsonb,0),
		('01INVAL0067','01INVATTR0014','01INVPRD0003','"BrandL"'::jsonb,0),
		('01INVAL0068','01INVATTR0014','01INVPRD0003','"BrandM"'::jsonb,0),
		('01INVAL0069','01INVATTR0014','01INVPRD0003','"BrandN"'::jsonb,0),
		('01INVAL0070','01INVATTR0014','01INVPRD0003','"BrandO"'::jsonb,0),
		('01INVAL0071','01INVATTR0015','01INVPRD0003','"Unisex"'::jsonb,0),
		('01INVAL0072','01INVATTR0015','01INVPRD0003','"Men"'::jsonb,0),
		('01INVAL0073','01INVATTR0015','01INVPRD0003','"Women"'::jsonb,0),
		('01INVAL0074','01INVATTR0015','01INVPRD0003','"Kids"'::jsonb,0),
		('01INVAL0075','01INVATTR0015','01INVPRD0003','"All"'::jsonb,0),
		-- Product 4 values (76..100)
		('01INVAL0076','01INVATTR0016','01INVPRD0004','"40"'::jsonb,0),
		('01INVAL0077','01INVATTR0016','01INVPRD0004','"41"'::jsonb,0),
		('01INVAL0078','01INVATTR0016','01INVPRD0004','"42"'::jsonb,0),
		('01INVAL0079','01INVATTR0016','01INVPRD0004','"43"'::jsonb,0),
		('01INVAL0080','01INVATTR0016','01INVPRD0004','"44"'::jsonb,0),
		('01INVAL0081','01INVATTR0017','01INVPRD0004','"White"'::jsonb,0),
		('01INVAL0082','01INVATTR0017','01INVPRD0004','"Black"'::jsonb,0),
		('01INVAL0083','01INVATTR0017','01INVPRD0004','"Red"'::jsonb,0),
		('01INVAL0084','01INVATTR0017','01INVPRD0004','"Blue"'::jsonb,0),
		('01INVAL0085','01INVATTR0017','01INVPRD0004','"Yellow"'::jsonb,0),
		('01INVAL0086','01INVATTR0018','01INVPRD0004','"Mesh"'::jsonb,0),
		('01INVAL0087','01INVATTR0018','01INVPRD0004','"Leather"'::jsonb,0),
		('01INVAL0088','01INVATTR0018','01INVPRD0004','"Canvas"'::jsonb,0),
		('01INVAL0089','01INVATTR0018','01INVPRD0004','"Synthetic"'::jsonb,0),
		('01INVAL0090','01INVATTR0018','01INVPRD0004','"Rubber"'::jsonb,0),
		('01INVAL0091','01INVATTR0019','01INVPRD0004','"BrandP"'::jsonb,0),
		('01INVAL0092','01INVATTR0019','01INVPRD0004','"BrandQ"'::jsonb,0),
		('01INVAL0093','01INVATTR0019','01INVPRD0004','"BrandR"'::jsonb,0),
		('01INVAL0094','01INVATTR0019','01INVPRD0004','"BrandS"'::jsonb,0),
		('01INVAL0095','01INVATTR0019','01INVPRD0004','"BrandT"'::jsonb,0),
		('01INVAL0096','01INVATTR0020','01INVPRD0004','"Unisex"'::jsonb,0),
		('01INVAL0097','01INVATTR0020','01INVPRD0004','"Men"'::jsonb,0),
		('01INVAL0098','01INVATTR0020','01INVPRD0004','"Women"'::jsonb,0),
		('01INVAL0099','01INVATTR0020','01INVPRD0004','"Kids"'::jsonb,0),
		('01INVAL0100','01INVATTR0020','01INVPRD0004','"All"'::jsonb,0),
		-- Product 5 values (101..125)
		('01INVAL0101','01INVATTR0021','01INVPRD0005','"OneSize"'::jsonb,0),
		('01INVAL0102','01INVATTR0021','01INVPRD0005','"Adjustable"'::jsonb,0),
		('01INVAL0103','01INVATTR0021','01INVPRD0005','"S"'::jsonb,0),
		('01INVAL0104','01INVATTR0021','01INVPRD0005','"M"'::jsonb,0),
		('01INVAL0105','01INVATTR0021','01INVPRD0005','"L"'::jsonb,0),
		('01INVAL0106','01INVATTR0022','01INVPRD0005','"Black"'::jsonb,0),
		('01INVAL0107','01INVATTR0022','01INVPRD0005','"Navy"'::jsonb,0),
		('01INVAL0108','01INVATTR0022','01INVPRD0005','"Grey"'::jsonb,0),
		('01INVAL0109','01INVATTR0022','01INVPRD0005','"White"'::jsonb,0),
		('01INVAL0110','01INVATTR0022','01INVPRD0005','"Brown"'::jsonb,0),
		('01INVAL0111','01INVATTR0023','01INVPRD0005','"Cotton"'::jsonb,0),
		('01INVAL0112','01INVATTR0023','01INVPRD0005','"Polyester"'::jsonb,0),
		('01INVAL0113','01INVATTR0023','01INVPRD0005','"Wool"'::jsonb,0),
		('01INVAL0114','01INVATTR0023','01INVPRD0005','"Leather"'::jsonb,0),
		('01INVAL0115','01INVATTR0023','01INVPRD0005','"Blend"'::jsonb,0),
		('01INVAL0116','01INVATTR0024','01INVPRD0005','"BrandU"'::jsonb,0),
		('01INVAL0117','01INVATTR0024','01INVPRD0005','"BrandV"'::jsonb,0),
		('01INVAL0118','01INVATTR0024','01INVPRD0005','"BrandW"'::jsonb,0),
		('01INVAL0119','01INVATTR0024','01INVPRD0005','"BrandX"'::jsonb,0),
		('01INVAL0120','01INVATTR0024','01INVPRD0005','"BrandY"'::jsonb,0),
		('01INVAL0121','01INVATTR0025','01INVPRD0005','"Unisex"'::jsonb,0),
		('01INVAL0122','01INVATTR0025','01INVPRD0005','"Men"'::jsonb,0),
		('01INVAL0123','01INVATTR0025','01INVPRD0005','"Women"'::jsonb,0),
		('01INVAL0124','01INVATTR0025','01INVPRD0005','"Kids"'::jsonb,0),
		('01INVAL0125','01INVATTR0025','01INVPRD0005','"All"'::jsonb,0);
	END IF;

	-- Variants: 5 variants per product (25 total)
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'inventory_variants'
	) THEN
		INSERT INTO "inventory_variants" ("id", "org_id", "product_id", "name", "sku", "barcode", "proposed_price", "status", "image_url", "etag", "created_at", "updated_at") VALUES
		-- Product 1 variants (1..5)
		('01INVVAR0001','01JWNY20G23KD4RV5VWYABQYHD','01INVPRD0001', jsonb_build_object('en-US','Basic T-Shirt - Variant 1'), 'PRD1-V1-SKU','1000000000001',199900,'active','https://example.com/images/tshirt-basic.png',(EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text,NOW(),NULL),
		('01INVVAR0002','01JWNY20G23KD4RV5VWYABQYHD','01INVPRD0001', jsonb_build_object('en-US','Basic T-Shirt - Variant 2'), 'PRD1-V2-SKU','1000000000002',199900,'active','https://example.com/images/tshirt-basic.png',(EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text,NOW(),NULL),
		('01INVVAR0003','01JWNY20G23KD4RV5VWYABQYHD','01INVPRD0001', jsonb_build_object('en-US','Basic T-Shirt - Variant 3'), 'PRD1-V3-SKU','1000000000003',199900,'active','https://example.com/images/tshirt-basic.png',(EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text,NOW(),NULL),
		('01INVVAR0004','01JWNY20G23KD4RV5VWYABQYHD','01INVPRD0001', jsonb_build_object('en-US','Basic T-Shirt - Variant 4'), 'PRD1-V4-SKU','1000000000004',199900,'active','https://example.com/images/tshirt-basic.png',(EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text,NOW(),NULL),
		('01INVVAR0005','01JWNY20G23KD4RV5VWYABQYHD','01INVPRD0001', jsonb_build_object('en-US','Basic T-Shirt - Variant 5'), 'PRD1-V5-SKU','1000000000005',199900,'active','https://example.com/images/tshirt-basic.png',(EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text,NOW(),NULL),
		-- Product 2 variants (6..10)
		('01INVVAR0006','01JWNY20G23KD4RV5VWYABQYHD','01INVPRD0002', jsonb_build_object('en-US','Denim Jeans - Variant 1'), 'PRD2-V1-SKU','1000000000011',219900,'active','https://example.com/images/jeans.png',(EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text,NOW(),NULL),
		('01INVVAR0007','01JWNY20G23KD4RV5VWYABQYHD','01INVPRD0002', jsonb_build_object('en-US','Denim Jeans - Variant 2'), 'PRD2-V2-SKU','1000000000012',219900,'active','https://example.com/images/jeans.png',(EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text,NOW(),NULL),
		('01INVVAR0008','01JWNY20G23KD4RV5VWYABQYHD','01INVPRD0002', jsonb_build_object('en-US','Denim Jeans - Variant 3'), 'PRD2-V3-SKU','1000000000013',219900,'active','https://example.com/images/jeans.png',(EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text,NOW(),NULL),
		('01INVVAR0009','01JWNY20G23KD4RV5VWYABQYHD','01INVPRD0002', jsonb_build_object('en-US','Denim Jeans - Variant 4'), 'PRD2-V4-SKU','1000000000014',219900,'active','https://example.com/images/jeans.png',(EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text,NOW(),NULL),
		('01INVVAR0010','01JWNY20G23KD4RV5VWYABQYHD','01INVPRD0002', jsonb_build_object('en-US','Denim Jeans - Variant 5'), 'PRD2-V5-SKU','1000000000015',219900,'active','https://example.com/images/jeans.png',(EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text,NOW(),NULL),
		-- Product 3 variants (11..15)
		('01INVVAR0011','01JWNY20G23KD4RV5VWYABQYHD','01INVPRD0003', jsonb_build_object('en-US','Hoodie - Variant 1'), 'PRD3-V1-SKU','1000000000021',249900,'active','https://example.com/images/hoodie.png',(EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text,NOW(),NULL),
		('01INVVAR0012','01JWNY20G23KD4RV5VWYABQYHD','01INVPRD0003', jsonb_build_object('en-US','Hoodie - Variant 2'), 'PRD3-V2-SKU','1000000000022',249900,'active','https://example.com/images/hoodie.png',(EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text,NOW(),NULL),
		('01INVVAR0013','01JWNY20G23KD4RV5VWYABQYHD','01INVPRD0003', jsonb_build_object('en-US','Hoodie - Variant 3'), 'PRD3-V3-SKU','1000000000023',249900,'active','https://example.com/images/hoodie.png',(EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text,NOW(),NULL),
		('01INVVAR0014','01JWNY20G23KD4RV5VWYABQYHD','01INVPRD0003', jsonb_build_object('en-US','Hoodie - Variant 4'), 'PRD3-V4-SKU','1000000000024',249900,'active','https://example.com/images/hoodie.png',(EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text,NOW(),NULL),
		('01INVVAR0015','01JWNY20G23KD4RV5VWYABQYHD','01INVPRD0003', jsonb_build_object('en-US','Hoodie - Variant 5'), 'PRD3-V5-SKU','1000000000025',249900,'active','https://example.com/images/hoodie.png',(EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text,NOW(),NULL),
		-- Product 4 variants (16..20)
		('01INVVAR0016','01JWNY20G23KD4RV5VWYABQYHD','01INVPRD0004', jsonb_build_object('en-US','Sneakers - Variant 1'), 'PRD4-V1-SKU','1000000000031',329900,'active','https://example.com/images/sneakers.png',(EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text,NOW(),NULL),
		('01INVVAR0017','01JWNY20G23KD4RV5VWYABQYHD','01INVPRD0004', jsonb_build_object('en-US','Sneakers - Variant 2'), 'PRD4-V2-SKU','1000000000032',329900,'active','https://example.com/images/sneakers.png',(EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text,NOW(),NULL),
		('01INVVAR0018','01JWNY20G23KD4RV5VWYABQYHD','01INVPRD0004', jsonb_build_object('en-US','Sneakers - Variant 3'), 'PRD4-V3-SKU','1000000000033',329900,'active','https://example.com/images/sneakers.png',(EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text,NOW(),NULL),
		('01INVVAR0019','01JWNY20G23KD4RV5VWYABQYHD','01INVPRD0004', jsonb_build_object('en-US','Sneakers - Variant 4'), 'PRD4-V4-SKU','1000000000034',329900,'active','https://example.com/images/sneakers.png',(EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text,NOW(),NULL),
		('01INVVAR0020','01JWNY20G23KD4RV5VWYABQYHD','01INVPRD0004', jsonb_build_object('en-US','Sneakers - Variant 5'), 'PRD4-V5-SKU','1000000000035',329900,'active','https://example.com/images/sneakers.png',(EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text,NOW(),NULL),
		-- Product 5 variants (21..25)
		('01INVVAR0021','01JWNY20G23KD4RV5VWYABQYHD','01INVPRD0005', jsonb_build_object('en-US','Baseball Cap - Variant 1'), 'PRD5-V1-SKU','1000000000041',79900,'active','https://example.com/images/cap.png',(EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text,NOW(),NULL),
		('01INVVAR0022','01JWNY20G23KD4RV5VWYABQYHD','01INVPRD0005', jsonb_build_object('en-US','Baseball Cap - Variant 2'), 'PRD5-V2-SKU','1000000000042',79900,'active','https://example.com/images/cap.png',(EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text,NOW(),NULL),
		('01INVVAR0023','01JWNY20G23KD4RV5VWYABQYHD','01INVPRD0005', jsonb_build_object('en-US','Baseball Cap - Variant 3'), 'PRD5-V3-SKU','1000000000043',79900,'active','https://example.com/images/cap.png',(EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text,NOW(),NULL),
		('01INVVAR0024','01JWNY20G23KD4RV5VWYABQYHD','01INVPRD0005', jsonb_build_object('en-US','Baseball Cap - Variant 4'), 'PRD5-V4-SKU','1000000000044',79900,'active','https://example.com/images/cap.png',(EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text,NOW(),NULL),
		('01INVVAR0025','01JWNY20G23KD4RV5VWYABQYHD','01INVPRD0005', jsonb_build_object('en-US','Baseball Cap - Variant 5'), 'PRD5-V5-SKU','1000000000045',79900,'active','https://example.com/images/cap.png',(EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text,NOW(),NULL);
	END IF;

	-- Link variant <-> attribute values (each variant takes the k-th value from each attribute)
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'inventory_variant_attr_val_rel'
	) THEN
		INSERT INTO "inventory_variant_attr_val_rel" ("variant_id", "attribute_value_id") VALUES
		-- Product 1: variants 1..5 (variant ids 1..5, values 1..25)
		('01INVVAR0001','01INVAL0001'),('01INVVAR0001','01INVAL0006'),('01INVVAR0001','01INVAL0011'),('01INVVAR0001','01INVAL0016'),('01INVVAR0001','01INVAL0021'),
		('01INVVAR0002','01INVAL0002'),('01INVVAR0002','01INVAL0007'),('01INVVAR0002','01INVAL0012'),('01INVVAR0002','01INVAL0017'),('01INVVAR0002','01INVAL0022'),
		('01INVVAR0003','01INVAL0003'),('01INVVAR0003','01INVAL0008'),('01INVVAR0003','01INVAL0013'),('01INVVAR0003','01INVAL0018'),('01INVVAR0003','01INVAL0023'),
		('01INVVAR0004','01INVAL0004'),('01INVVAR0004','01INVAL0009'),('01INVVAR0004','01INVAL0014'),('01INVVAR0004','01INVAL0019'),('01INVVAR0004','01INVAL0024'),
		('01INVVAR0005','01INVAL0005'),('01INVVAR0005','01INVAL0010'),('01INVVAR0005','01INVAL0015'),('01INVVAR0005','01INVAL0020'),('01INVVAR0005','01INVAL0025'),
		-- Product 2: variants 6..10 (values 26..50)
		('01INVVAR0006','01INVAL0026'),('01INVVAR0006','01INVAL0031'),('01INVVAR0006','01INVAL0036'),('01INVVAR0006','01INVAL0041'),('01INVVAR0006','01INVAL0046'),
		('01INVVAR0007','01INVAL0027'),('01INVVAR0007','01INVAL0032'),('01INVVAR0007','01INVAL0037'),('01INVVAR0007','01INVAL0042'),('01INVVAR0007','01INVAL0047'),
		('01INVVAR0008','01INVAL0028'),('01INVVAR0008','01INVAL0033'),('01INVVAR0008','01INVAL0038'),('01INVVAR0008','01INVAL0043'),('01INVVAR0008','01INVAL0048'),
		('01INVVAR0009','01INVAL0029'),('01INVVAR0009','01INVAL0034'),('01INVVAR0009','01INVAL0039'),('01INVVAR0009','01INVAL0044'),('01INVVAR0009','01INVAL0049'),
		('01INVVAR0010','01INVAL0030'),('01INVVAR0010','01INVAL0035'),('01INVVAR0010','01INVAL0040'),('01INVVAR0010','01INVAL0045'),('01INVVAR0010','01INVAL0050'),
		-- Product 3: variants 11..15 (values 51..75)
		('01INVVAR0011','01INVAL0051'),('01INVVAR0011','01INVAL0056'),('01INVVAR0011','01INVAL0061'),('01INVVAR0011','01INVAL0066'),('01INVVAR0011','01INVAL0071'),
		('01INVVAR0012','01INVAL0052'),('01INVVAR0012','01INVAL0057'),('01INVVAR0012','01INVAL0062'),('01INVVAR0012','01INVAL0067'),('01INVVAR0012','01INVAL0072'),
		('01INVVAR0013','01INVAL0053'),('01INVVAR0013','01INVAL0058'),('01INVVAR0013','01INVAL0063'),('01INVVAR0013','01INVAL0068'),('01INVVAR0013','01INVAL0073'),
		('01INVVAR0014','01INVAL0054'),('01INVVAR0014','01INVAL0059'),('01INVVAR0014','01INVAL0064'),('01INVVAR0014','01INVAL0069'),('01INVVAR0014','01INVAL0074'),
		('01INVVAR0015','01INVAL0055'),('01INVVAR0015','01INVAL0060'),('01INVVAR0015','01INVAL0065'),('01INVVAR0015','01INVAL0070'),('01INVVAR0015','01INVAL0075'),
		-- Product 4: variants 16..20 (values 76..100)
		('01INVVAR0016','01INVAL0076'),('01INVVAR0016','01INVAL0081'),('01INVVAR0016','01INVAL0086'),('01INVVAR0016','01INVAL0091'),('01INVVAR0016','01INVAL0096'),
		('01INVVAR0017','01INVAL0077'),('01INVVAR0017','01INVAL0082'),('01INVVAR0017','01INVAL0087'),('01INVVAR0017','01INVAL0092'),('01INVVAR0017','01INVAL0097'),
		('01INVVAR0018','01INVAL0078'),('01INVVAR0018','01INVAL0083'),('01INVVAR0018','01INVAL0088'),('01INVVAR0018','01INVAL0093'),('01INVVAR0018','01INVAL0098'),
		('01INVVAR0019','01INVAL0079'),('01INVVAR0019','01INVAL0084'),('01INVVAR0019','01INVAL0089'),('01INVVAR0019','01INVAL0094'),('01INVVAR0019','01INVAL0099'),
		('01INVVAR0020','01INVAL0080'),('01INVVAR0020','01INVAL0085'),('01INVVAR0020','01INVAL0090'),('01INVVAR0020','01INVAL0095'),('01INVVAR0020','01INVAL0100'),
		-- Product 5: variants 21..25 (values 101..125)
		('01INVVAR0021','01INVAL0101'),('01INVVAR0021','01INVAL0106'),('01INVVAR0021','01INVAL0111'),('01INVVAR0021','01INVAL0116'),('01INVVAR0021','01INVAL0121'),
		('01INVVAR0022','01INVAL0102'),('01INVVAR0022','01INVAL0107'),('01INVVAR0022','01INVAL0112'),('01INVVAR0022','01INVAL0117'),('01INVVAR0022','01INVAL0122'),
		('01INVVAR0023','01INVAL0103'),('01INVVAR0023','01INVAL0108'),('01INVVAR0023','01INVAL0113'),('01INVVAR0023','01INVAL0118'),('01INVVAR0023','01INVAL0123'),
		('01INVVAR0024','01INVAL0104'),('01INVVAR0024','01INVAL0109'),('01INVVAR0024','01INVAL0114'),('01INVVAR0024','01INVAL0119'),('01INVVAR0024','01INVAL0124'),
		('01INVVAR0025','01INVAL0105'),('01INVVAR0025','01INVAL0110'),('01INVVAR0025','01INVAL0115'),('01INVVAR0025','01INVAL0120'),('01INVVAR0025','01INVAL0125');
	END IF;

	-- Product categories and relations (one shared category)
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'inventory_product_categories'
	) THEN
		INSERT INTO "inventory_product_categories" ("id", "org_id", "name", "etag", "created_at", "updated_at") VALUES
		('01INVCAT0001','01JWNY20G23KD4RV5VWYABQYHD', jsonb_build_object('en-US','Clothing','vi-VN','Quần áo'), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL);
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'inventory_product_category_rel'
	) THEN
		INSERT INTO "inventory_product_category_rel" ("product_id", "product_category_id") VALUES
		('01INVPRD0001','01INVCAT0001'),('01INVPRD0002','01INVCAT0001'),('01INVPRD0003','01INVCAT0001'),('01INVPRD0004','01INVCAT0001'),('01INVPRD0005','01INVCAT0001');
	END IF;

	-- Set default variant for each product (first variant)
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'inventory_products'
	) THEN
		UPDATE "inventory_products" SET "default_variant_id" = '01INVVAR0001' WHERE "id" = '01INVPRD0001';
		UPDATE "inventory_products" SET "default_variant_id" = '01INVVAR0006' WHERE "id" = '01INVPRD0002';
		UPDATE "inventory_products" SET "default_variant_id" = '01INVVAR0011' WHERE "id" = '01INVPRD0003';
		UPDATE "inventory_products" SET "default_variant_id" = '01INVVAR0016' WHERE "id" = '01INVPRD0004';
		UPDATE "inventory_products" SET "default_variant_id" = '01INVVAR0021' WHERE "id" = '01INVPRD0005';
	END IF;
END $$;
