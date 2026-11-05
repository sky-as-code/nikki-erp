-- IAM resources and actions for the Sales module: the sales channel and the sales point.
-- Every Sales permission row lives here, in one file, so that a reviewer can see the module's
-- entire permission surface without opening several migrations.
--
-- The dynamic resource engine asserts permissions using the schema name as the resource code, so
-- these codes must stay byte-identical to the "sales_channel" / "sales_point" schema names. A code
-- that drifts from its schema denies every request, with nothing in the response pointing at the
-- seed as the cause.
--
-- Lifecycle operations are seeded as SEPARATE actions rather than folded into "update", because
-- they are materially different powers. Suspending a channel stops every sales point under it from
-- selling; archiving one retires it; unarchiving resurrects a point that was deliberately
-- decommissioned. A role that may correct a display name should not thereby be able to do any of
-- them.
--
-- Deliberate omissions, each of which would otherwise look like something forgotten:
--
--   * There is NO "register" action on sales_channel. RegisterSalesChannel is the idempotent
--     module-integration entry point, and it is create-or-return-existing: a caller that may not
--     create a channel must not be able to reach it, and one that may needs no second permission.
--     It rides on "create".
--   * There is NO "resolve" action. Resolving a channel code to its id is a read of one row by a
--     different key, and seeding it would let a role be granted "may resolve" while unable to read
--     what it resolved.
--   * sales_point has no "register" either, for the same reason as the channel.
--   * Archive and unarchive both ride on the built-in set_archived permission rather than taking
--     actions of their own. Unarchiving is the same power applied in reverse, so splitting them
--     would let a role archive sales points it could not bring back.
--
-- Deliberately NO iam_entitlements rows. Unit of Measure grants the system "User" role a
-- domain-wide read so that any user can pick a unit while filling in an unrelated form. Sales
-- configuration is not comparable: a channel names which selling surfaces exist and which payment
-- methods they accept, and a sales point identifies a physical machine or store. Access follows
-- explicitly assigned roles, which is the same choice Products, Contacts and Purchase made.

DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_resources'
	) THEN
		INSERT INTO "iam_resources" (
			"id", "name", "code", "description", "owner_type", "max_scope", "min_scope", "created_at", "etag"
		) VALUES
		('01M3SALES00000000000000001', jsonb_build_object('en-US', 'Sales - Channel', 'vi-VN', 'Bán hàng - Kênh'), 'sales_channel', 'Phân loại nơi diễn ra giao dịch bán và các phương thức thanh toán được chấp nhận', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000002', jsonb_build_object('en-US', 'Sales - Point', 'vi-VN', 'Bán hàng - Điểm'), 'sales_point', 'Một địa điểm bán hàng cụ thể thuộc một kênh: kiosk, cửa hàng hoặc gian hàng trực tuyến', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000000K', jsonb_build_object('en-US', 'Sales - Order', 'vi-VN', 'Bán hàng - Đơn hàng'), 'sales_order', 'Một giao dịch thương mại: bán gì, cho ai, qua kênh nào', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000000M', jsonb_build_object('en-US', 'Sales - Order Line', 'vi-VN', 'Bán hàng - Dòng đơn hàng'), 'sales_order_line', 'Một mặt hàng được bán trong đơn hàng, kèm số lượng và giá tại thời điểm bán', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000000Z', jsonb_build_object('en-US', 'Sales - Order Line Component', 'vi-VN', 'Bán hàng - Thành phần dòng đơn hàng'), 'sales_order_line_component', 'Một sản phẩm thực trong dòng combo, là thứ mà phân hệ Kho hàng thực sự xuất giao', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000010', jsonb_build_object('en-US', 'Sales - Order Adjustment', 'vi-VN', 'Bán hàng - Điều chỉnh đơn hàng'), 'sales_order_adjustment', 'Một bước trong quá trình tính giá, được lưu lại để giải thích và tính lại giá', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000011', jsonb_build_object('en-US', 'Sales - Order Event', 'vi-VN', 'Bán hàng - Sự kiện đơn hàng'), 'sales_order_event', 'Một sự kiện xảy ra với giao dịch bán: lịch sử kiểm toán của chứng từ', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000015', jsonb_build_object('en-US', 'Sales - Pricelist', 'vi-VN', 'Bán hàng - Bảng giá bán'), 'sales_pricelist', 'Bộ giá áp dụng cho một kênh hoặc một điểm bán hàng trong một khoảng thời gian', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000016', jsonb_build_object('en-US', 'Sales - Pricelist Item', 'vi-VN', 'Bán hàng - Mục bảng giá bán'), 'sales_pricelist_item', 'Một mức giá cho một biến thể sản phẩm, theo một đơn vị tính, áp dụng từ một số lượng trở lên', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000017', jsonb_build_object('en-US', 'Sales - Combo', 'vi-VN', 'Bán hàng - Combo'), 'sales_combo', 'Gói sản phẩm bán theo giá riêng, không phụ thuộc vào giá bán lẻ của từng thành phần', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000018', jsonb_build_object('en-US', 'Sales - Combo Component', 'vi-VN', 'Bán hàng - Thành phần combo'), 'sales_combo_component', 'Một sản phẩm trong định nghĩa gói sản phẩm', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000034', jsonb_build_object('en-US', 'Sales - Promotion', 'vi-VN', 'Bán hàng - Khuyến mãi'), 'sales_promotion', 'Cấu hình một chương trình khuyến mãi: điều kiện, ưu đãi và phạm vi kênh, điểm bán', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000035', jsonb_build_object('en-US', 'Sales - Promotion Rule', 'vi-VN', 'Bán hàng - Điều kiện khuyến mãi'), 'sales_promotion_rule', 'Một cách để đơn hàng đạt điều kiện của chương trình khuyến mãi', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000036', jsonb_build_object('en-US', 'Sales - Promotion Rule Product', 'vi-VN', 'Bán hàng - Sản phẩm của điều kiện khuyến mãi'), 'sales_promotion_rule_product', 'Sản phẩm được tính khi xét điều kiện khuyến mãi', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000037', jsonb_build_object('en-US', 'Sales - Promotion Rule Category', 'vi-VN', 'Bán hàng - Danh mục của điều kiện khuyến mãi'), 'sales_promotion_rule_category', 'Danh mục sản phẩm được tính khi xét điều kiện khuyến mãi', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000038', jsonb_build_object('en-US', 'Sales - Promotion Reward', 'vi-VN', 'Bán hàng - Ưu đãi khuyến mãi'), 'sales_promotion_reward', 'Một lợi ích khách hàng có thể nhận: giảm giá, tặng sản phẩm hoặc miễn phí vận chuyển', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000039', jsonb_build_object('en-US', 'Sales - Promotion Reward Product', 'vi-VN', 'Bán hàng - Sản phẩm của ưu đãi khuyến mãi'), 'sales_promotion_reward_product', 'Sản phẩm được giảm giá bởi ưu đãi khuyến mãi', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALESPRM0000000000000A', jsonb_build_object('en-US', 'Sales - Promotion Reward Category', 'vi-VN', 'Bán hàng - Danh mục của ưu đãi khuyến mãi'), 'sales_promotion_reward_category', 'Danh mục sản phẩm được giảm giá bởi ưu đãi khuyến mãi', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_actions'
	) THEN
		INSERT INTO "iam_actions" ("id", "name", "code", "description", "resource_id", "etag") VALUES
		-- Sales Channel
		('01M3SALES00000000000000003', jsonb_build_object('en-US', 'Create', 'vi-VN', 'Tạo'), 'create', NULL, '01M3SALES00000000000000001', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000004', jsonb_build_object('en-US', 'Update', 'vi-VN', 'Cập nhật'), 'update', NULL, '01M3SALES00000000000000001', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000005', jsonb_build_object('en-US', 'Delete', 'vi-VN', 'Xóa'), 'delete', NULL, '01M3SALES00000000000000001', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000006', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M3SALES00000000000000001', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000007', jsonb_build_object('en-US', 'Set archived status', 'vi-VN', 'Đặt trạng thái lưu trữ'), 'set_archived', 'Lưu trữ kênh bán hàng, hoặc khôi phục kênh đã lưu trữ', '01M3SALES00000000000000001', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000008', jsonb_build_object('en-US', 'Suspend', 'vi-VN', 'Tạm ngưng'), 'suspend', 'Ngừng nhận giao dịch mới trên kênh, vẫn giữ lịch sử, trả hàng và hoàn tiền hoạt động', '01M3SALES00000000000000001', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000009', jsonb_build_object('en-US', 'Activate', 'vi-VN', 'Kích hoạt'), 'activate', 'Đưa kênh đang tạm ngưng hoạt động trở lại', '01M3SALES00000000000000001', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000000A', jsonb_build_object('en-US', 'Enable payment method', 'vi-VN', 'Bật phương thức thanh toán'), 'enable_payment_method', 'Cho phép sử dụng một phương thức thanh toán trên kênh này', '01M3SALES00000000000000001', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000000B', jsonb_build_object('en-US', 'Disable payment method', 'vi-VN', 'Tắt phương thức thanh toán'), 'disable_payment_method', 'Ngừng thanh toán mới bằng một phương thức trên kênh này, giữ nguyên các thanh toán trước đây', '01M3SALES00000000000000001', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Sales Point
		('01M3SALES0000000000000000C', jsonb_build_object('en-US', 'Create', 'vi-VN', 'Tạo'), 'create', NULL, '01M3SALES00000000000000002', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000000D', jsonb_build_object('en-US', 'Update', 'vi-VN', 'Cập nhật'), 'update', NULL, '01M3SALES00000000000000002', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000000E', jsonb_build_object('en-US', 'Delete', 'vi-VN', 'Xóa'), 'delete', NULL, '01M3SALES00000000000000002', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000000F', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M3SALES00000000000000002', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000000G', jsonb_build_object('en-US', 'Set archived status', 'vi-VN', 'Đặt trạng thái lưu trữ'), 'set_archived', 'Lưu trữ điểm bán hàng, hoặc khôi phục điểm đã lưu trữ', '01M3SALES00000000000000002', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000000H', jsonb_build_object('en-US', 'Suspend', 'vi-VN', 'Tạm ngưng'), 'suspend', 'Ngừng nhận đơn hàng mới tại điểm bán hàng, vẫn cho phép trả hàng và hoàn tiền', '01M3SALES00000000000000002', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000000J', jsonb_build_object('en-US', 'Activate', 'vi-VN', 'Kích hoạt'), 'activate', 'Đưa điểm bán hàng đang tạm ngưng hoạt động trở lại', '01M3SALES00000000000000002', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALESPOINTPAY00000000A', jsonb_build_object('en-US', 'Enable payment method', 'vi-VN', 'Bật phương thức thanh toán'), 'enable_payment_method', 'Cho phép thu tiền bằng một phương thức thanh toán tại điểm bán hàng này và chỉ định tài khoản người bán nhận tiền', '01M3SALES00000000000000002', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALESPOINTPAY00000000B', jsonb_build_object('en-US', 'Disable payment method', 'vi-VN', 'Tắt phương thức thanh toán'), 'disable_payment_method', 'Ngừng thanh toán mới bằng một phương thức tại điểm bán hàng này, giữ nguyên các thanh toán trước đây', '01M3SALES00000000000000002', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Sales Order
		('01M3SALES0000000000000000N', jsonb_build_object('en-US', 'Create', 'vi-VN', 'Tạo'), 'create', NULL, '01M3SALES0000000000000000K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000000P', jsonb_build_object('en-US', 'Update', 'vi-VN', 'Cập nhật'), 'update', NULL, '01M3SALES0000000000000000K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000000Q', jsonb_build_object('en-US', 'Delete', 'vi-VN', 'Xóa'), 'delete', NULL, '01M3SALES0000000000000000K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000000R', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M3SALES0000000000000000K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000000S', jsonb_build_object('en-US', 'Set archived status', 'vi-VN', 'Đặt trạng thái lưu trữ'), 'set_archived', 'Lưu trữ đơn bán hàng, hoặc khôi phục đơn đã lưu trữ', '01M3SALES0000000000000000K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Sales Order Line
		('01M3SALES0000000000000000T', jsonb_build_object('en-US', 'Create', 'vi-VN', 'Tạo'), 'create', NULL, '01M3SALES0000000000000000M', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000000V', jsonb_build_object('en-US', 'Update', 'vi-VN', 'Cập nhật'), 'update', NULL, '01M3SALES0000000000000000M', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000000W', jsonb_build_object('en-US', 'Delete', 'vi-VN', 'Xóa'), 'delete', NULL, '01M3SALES0000000000000000M', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000000X', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M3SALES0000000000000000M', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000000Y', jsonb_build_object('en-US', 'Set archived status', 'vi-VN', 'Đặt trạng thái lưu trữ'), 'set_archived', 'Lưu trữ dòng đơn bán hàng, hoặc khôi phục dòng đã lưu trữ', '01M3SALES0000000000000000M', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Sales Order Line Component
		('01M3SALES00000000000000012', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M3SALES0000000000000000Z', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Sales Order Adjustment
		('01M3SALES00000000000000013', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M3SALES00000000000000010', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Sales Order Event
		('01M3SALES00000000000000014', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M3SALES00000000000000011', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Sales Pricelist
		('01M3SALES00000000000000019', jsonb_build_object('en-US', 'Create', 'vi-VN', 'Tạo'), 'create', NULL, '01M3SALES00000000000000015', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000001A', jsonb_build_object('en-US', 'Update', 'vi-VN', 'Cập nhật'), 'update', NULL, '01M3SALES00000000000000015', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000001B', jsonb_build_object('en-US', 'Delete', 'vi-VN', 'Xóa'), 'delete', NULL, '01M3SALES00000000000000015', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000001C', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M3SALES00000000000000015', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000001D', jsonb_build_object('en-US', 'Set archived status', 'vi-VN', 'Đặt trạng thái lưu trữ'), 'set_archived', 'Lưu trữ bảng giá, hoặc khôi phục bảng giá đã lưu trữ', '01M3SALES00000000000000015', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Sales Pricelist Item
		('01M3SALES0000000000000001E', jsonb_build_object('en-US', 'Create', 'vi-VN', 'Tạo'), 'create', NULL, '01M3SALES00000000000000016', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000001F', jsonb_build_object('en-US', 'Update', 'vi-VN', 'Cập nhật'), 'update', NULL, '01M3SALES00000000000000016', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000001G', jsonb_build_object('en-US', 'Delete', 'vi-VN', 'Xóa'), 'delete', NULL, '01M3SALES00000000000000016', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000001H', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M3SALES00000000000000016', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000001J', jsonb_build_object('en-US', 'Set archived status', 'vi-VN', 'Đặt trạng thái lưu trữ'), 'set_archived', 'Lưu trữ mục bảng giá, hoặc khôi phục mục đã lưu trữ', '01M3SALES00000000000000016', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Sales Combo
		('01M3SALES0000000000000001K', jsonb_build_object('en-US', 'Create', 'vi-VN', 'Tạo'), 'create', NULL, '01M3SALES00000000000000017', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000001M', jsonb_build_object('en-US', 'Update', 'vi-VN', 'Cập nhật'), 'update', NULL, '01M3SALES00000000000000017', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000001N', jsonb_build_object('en-US', 'Delete', 'vi-VN', 'Xóa'), 'delete', NULL, '01M3SALES00000000000000017', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000001P', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M3SALES00000000000000017', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000001Q', jsonb_build_object('en-US', 'Set archived status', 'vi-VN', 'Đặt trạng thái lưu trữ'), 'set_archived', 'Lưu trữ combo, hoặc khôi phục combo đã lưu trữ', '01M3SALES00000000000000017', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Sales Combo Component
		('01M3SALES0000000000000001R', jsonb_build_object('en-US', 'Create', 'vi-VN', 'Tạo'), 'create', NULL, '01M3SALES00000000000000018', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000001S', jsonb_build_object('en-US', 'Update', 'vi-VN', 'Cập nhật'), 'update', NULL, '01M3SALES00000000000000018', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000001T', jsonb_build_object('en-US', 'Delete', 'vi-VN', 'Xóa'), 'delete', NULL, '01M3SALES00000000000000018', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000001V', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M3SALES00000000000000018', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000001W', jsonb_build_object('en-US', 'Set archived status', 'vi-VN', 'Đặt trạng thái lưu trữ'), 'set_archived', 'Lưu trữ thành phần combo, hoặc khôi phục thành phần đã lưu trữ', '01M3SALES00000000000000018', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Sales Promotion Program
		('01M3SALESPRMA0340000000000', jsonb_build_object('en-US', 'Create', 'vi-VN', 'Tạo'), 'create', NULL, '01M3SALES00000000000000034', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALESPRMA0340000000001', jsonb_build_object('en-US', 'Update', 'vi-VN', 'Cập nhật'), 'update', NULL, '01M3SALES00000000000000034', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALESPRMA0340000000002', jsonb_build_object('en-US', 'Delete', 'vi-VN', 'Xóa'), 'delete', NULL, '01M3SALES00000000000000034', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALESPRMA0340000000003', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M3SALES00000000000000034', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALESPRMA0340000000004', jsonb_build_object('en-US', 'Set archived status', 'vi-VN', 'Đặt trạng thái lưu trữ'), 'set_archived', 'Lưu trữ hoặc khôi phục', '01M3SALES00000000000000034', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALESPRMA0340000000005', jsonb_build_object('en-US', 'Activate', 'vi-VN', 'Kích hoạt'), 'activate', 'Kiểm tra cấu hình và kích hoạt chương trình khuyến mãi', '01M3SALES00000000000000034', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALESPRMA0340000000006', jsonb_build_object('en-US', 'Suspend', 'vi-VN', 'Tạm ngưng'), 'suspend', 'Tạm ngưng chương trình khuyến mãi để sửa cấu hình', '01M3SALES00000000000000034', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALESPRMA0350000000000', jsonb_build_object('en-US', 'Create', 'vi-VN', 'Tạo'), 'create', NULL, '01M3SALES00000000000000035', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALESPRMA0350000000001', jsonb_build_object('en-US', 'Update', 'vi-VN', 'Cập nhật'), 'update', NULL, '01M3SALES00000000000000035', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALESPRMA0350000000002', jsonb_build_object('en-US', 'Delete', 'vi-VN', 'Xóa'), 'delete', NULL, '01M3SALES00000000000000035', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALESPRMA0350000000003', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M3SALES00000000000000035', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALESPRMA0360000000000', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M3SALES00000000000000036', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALESPRMA0370000000000', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M3SALES00000000000000037', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALESPRMA0380000000000', jsonb_build_object('en-US', 'Create', 'vi-VN', 'Tạo'), 'create', NULL, '01M3SALES00000000000000038', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALESPRMA0380000000001', jsonb_build_object('en-US', 'Update', 'vi-VN', 'Cập nhật'), 'update', NULL, '01M3SALES00000000000000038', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALESPRMA0380000000002', jsonb_build_object('en-US', 'Delete', 'vi-VN', 'Xóa'), 'delete', NULL, '01M3SALES00000000000000038', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALESPRMA0380000000003', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M3SALES00000000000000038', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALESPRMA0390000000000', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M3SALES00000000000000039', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALESPRMA00A0000000000', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M3SALESPRM0000000000000A', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		-- Sales Promotion Condition Group
		-- Sales Promotion Condition
		-- Sales Promotion Condition Target
		-- Sales Promotion Reward
		-- Sales Promotion Compatibility
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;


-- Pricing authorization, added by the product-pricing change request.
--
-- The resources and their CRUD actions were already seeded above; what was missing was roles, and
-- the entitlements binding them. Two roles rather than one, because reading what something sells
-- for and deciding it are different powers: a shop floor or a report needs the first and must not
-- have the second.
--
-- The reader holds `read` on BOTH the list and its rules. A rule is meaningless without the list it
-- belongs to -- the currency and the scope live there -- so granting one without the other would
-- produce a reader that could see prices it could not interpret.
--
-- There is deliberately NO separate permission for choosing the organization default. That
-- operation asserts `update` (see PermissionSetDefaultPricelist): choosing which list prices an
-- order that named none is the same class of power as editing what a list charges, and a role that
-- may do one may sensibly do the other.

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
		('01M3ROLE0000000000PRICING1', 'Xem giá bán', 'Xem bảng giá bán và quy tắc giá, không được thay đổi giá bán', false, '01JWNMZ36QHC7CQQ748H9NQ6J6', true, false, true, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ROLE0000000000PRICING2', 'Quản lý giá bán', 'Tạo và chỉnh sửa bảng giá bán và quy tắc giá, bao gồm chọn bảng giá mặc định của tổ chức', false, '01JWNMZ36QHC7CQQ748H9NQ6J6', true, false, true, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
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
		('01M3ENT00000000000PRICE001', 'Xem bảng giá', 'Xem bảng giá', 'read:sales_pricelist:org', '01M3SALES0000000000000001C', '01M3SALES00000000000000015', '01M3ROLE0000000000PRICING1', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT00000000000PRICE002', 'Xem các quy tắc trong bảng giá', 'Xem các quy tắc trong bảng giá', 'read:sales_pricelist_item:org', '01M3SALES0000000000000001H', '01M3SALES00000000000000016', '01M3ROLE0000000000PRICING1', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT00000000000PRICE003', 'Tạo bảng giá', 'Tạo bảng giá', 'create:sales_pricelist:org', '01M3SALES00000000000000019', '01M3SALES00000000000000015', '01M3ROLE0000000000PRICING2', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT00000000000PRICE004', 'Tạo quy tắc giá', 'Tạo quy tắc giá', 'create:sales_pricelist_item:org', '01M3SALES0000000000000001E', '01M3SALES00000000000000016', '01M3ROLE0000000000PRICING2', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT00000000000PRICE005', 'Cập nhật bảng giá', 'Cập nhật bảng giá', 'update:sales_pricelist:org', '01M3SALES0000000000000001A', '01M3SALES00000000000000015', '01M3ROLE0000000000PRICING2', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT00000000000PRICE006', 'Cập nhật quy tắc giá', 'Cập nhật quy tắc giá', 'update:sales_pricelist_item:org', '01M3SALES0000000000000001F', '01M3SALES00000000000000016', '01M3ROLE0000000000PRICING2', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT00000000000PRICE007', 'Xóa bảng giá', 'Xóa bảng giá', 'delete:sales_pricelist:org', '01M3SALES0000000000000001B', '01M3SALES00000000000000015', '01M3ROLE0000000000PRICING2', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT00000000000PRICE008', 'Xóa quy tắc giá', 'Xóa quy tắc giá', 'delete:sales_pricelist_item:org', '01M3SALES0000000000000001G', '01M3SALES00000000000000016', '01M3ROLE0000000000PRICING2', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT00000000000PRICE009', 'Xem bảng giá', 'Xem bảng giá', 'read:sales_pricelist:org', '01M3SALES0000000000000001C', '01M3SALES00000000000000015', '01M3ROLE0000000000PRICING2', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT00000000000PRICE010', 'Xem quy tắc giá', 'Xem quy tắc giá', 'read:sales_pricelist_item:org', '01M3SALES0000000000000001H', '01M3SALES00000000000000016', '01M3ROLE0000000000PRICING2', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT00000000000PRICE011', 'Lưu trữ bảng giá', 'Lưu trữ bảng giá', 'set_archived:sales_pricelist:org', '01M3SALES0000000000000001D', '01M3SALES00000000000000015', '01M3ROLE0000000000PRICING2', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT00000000000PRICE012', 'Lưu trữ quy tắc giá', 'Lưu trữ quy tắc giá', 'set_archived:sales_pricelist_item:org', '01M3SALES0000000000000001J', '01M3SALES00000000000000016', '01M3ROLE0000000000PRICING2', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;


-- ============================================================================
-- Merged from 1007006_sales_voucher_iam.sql
-- ============================================================================
-- IAM resources and actions for the voucher pair (SALES-022).
--
-- Two resources with deliberately different permission surfaces.
--
-- A voucher CODE is operator-managed master data and gets the full five actions: somebody runs
-- campaigns, and running one means creating codes, correcting a validity window, withdrawing a
-- mispriced offer and reading the result.
--
-- A voucher REDEMPTION gets `read` alone. It is a ledger the system writes as orders move, and every
-- non-read power over it is a way to corrupt what it records: a client able to create one could forge
-- a discount's provenance, one able to update one could release a hold another order is relying on,
-- and one able to delete one would break the usage count that is derived from these rows. The same
-- reasoning already applies to sales_order_events, sales_order_adjustments and
-- sales_order_line_components.
--
-- The codes must stay byte-identical to the schema names. A code that drifts denies every request to
-- that resource, with nothing in the 403 pointing at this file as the cause.

DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_resources'
	) THEN
		INSERT INTO "iam_resources" (
			"id", "name", "code", "description", "owner_type", "max_scope", "min_scope", "created_at", "etag"
		) VALUES
		('01M3SALES00000000000000048', jsonb_build_object('en-US', 'Sales - Voucher', 'vi-VN', 'Bán hàng - Mã khuyến mãi'), 'sales_voucher', 'Mã kích hoạt một chương trình khuyến mãi; không mang giá trị riêng', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000049', jsonb_build_object('en-US', 'Sales - Promotion Application', 'vi-VN', 'Bán hàng - Kết quả áp dụng khuyến mãi'), 'sales_promotion_application', 'Bằng chứng một chương trình đã cấp lợi ích cho một đơn đã xác nhận', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALESPRM0000000000000B', jsonb_build_object('en-US', 'Sales - Promotion Allocation', 'vi-VN', 'Bán hàng - Phân bổ khuyến mãi'), 'sales_promotion_allocation', 'Phần lợi ích khuyến mãi phân bổ xuống từng dòng đơn', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALESPRM0000000000000C', jsonb_build_object('en-US', 'Sales - Voucher Issuance', 'vi-VN', 'Bán hàng - Phát hành phiếu đơn sau'), 'sales_voucher_issuance', 'Dấu vết phiếu cho đơn hàng sau được cấp từ một đơn nguồn', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALESPRM0000000000000D', jsonb_build_object('en-US', 'Sales - Order Voucher', 'vi-VN', 'Bán hàng - Mã gắn trên đơn'), 'sales_order_voucher', 'Mã khuyến mãi đã nhập trên đơn nháp; không giữ lượt dùng', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALESPRM0000000000000E', jsonb_build_object('en-US', 'Sales - Order Promotion Selection', 'vi-VN', 'Bán hàng - Lựa chọn ưu đãi trên đơn'), 'sales_order_promotion_selection', 'Ưu đãi khách đã chọn hoặc từ chối trên đơn nháp', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_actions'
	) THEN
		INSERT INTO "iam_actions" ("id", "name", "code", "description", "resource_id", "etag") VALUES
		-- Sales Voucher Code: full CRUD, an operator manages campaigns.
		('01M3SALESPRMA0480000000000', jsonb_build_object('en-US', 'Create', 'vi-VN', 'Tạo'), 'create', NULL, '01M3SALES00000000000000048', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALESPRMA0480000000001', jsonb_build_object('en-US', 'Update', 'vi-VN', 'Cập nhật'), 'update', NULL, '01M3SALES00000000000000048', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALESPRMA0480000000002', jsonb_build_object('en-US', 'Delete', 'vi-VN', 'Xóa'), 'delete', NULL, '01M3SALES00000000000000048', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALESPRMA0480000000003', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M3SALES00000000000000048', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALESPRMA0480000000004', jsonb_build_object('en-US', 'Set archived status', 'vi-VN', 'Đặt trạng thái lưu trữ'), 'set_archived', 'Lưu trữ hoặc khôi phục', '01M3SALES00000000000000048', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALESPRMA0480000000005', jsonb_build_object('en-US', 'Generate vouchers', 'vi-VN', 'Sinh mã hàng loạt'), 'generate', 'Sinh hàng loạt mã khuyến mãi cho một chương trình', '01M3SALES00000000000000048', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALESPRMA0480000000006', jsonb_build_object('en-US', 'Revoke', 'vi-VN', 'Thu hồi'), 'revoke', 'Thu hồi mã khuyến mãi; không thể hoàn tác', '01M3SALES00000000000000048', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALESPRMA0490000000000', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M3SALES00000000000000049', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALESPRMA00B0000000000', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M3SALESPRM0000000000000B', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALESPRMA00C0000000000', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M3SALESPRM0000000000000C', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALESPRMA00D0000000000', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M3SALESPRM0000000000000D', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALESPRMA00E0000000000', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M3SALESPRM0000000000000E', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		-- Sales Voucher Redemption: read alone. See the note at the top of this file.
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_actions'
	) THEN
		INSERT INTO "iam_actions" ("id", "name", "code", "description", "resource_id", "etag") VALUES
		-- Applying a voucher is a power over the ORDER, not over the voucher code.
		--
		-- It reserves a use and changes what the basket costs; the code itself is untouched master
		-- data. Seating it on sales_order is what lets a till take a discount at the counter without
		-- also holding write permission over campaign configuration.
		('01M3SALES0000000000000004G', jsonb_build_object('en-US', 'Apply voucher', 'vi-VN', 'Áp dụng phiếu giảm giá'), 'apply_voucher', 'Áp dụng mã phiếu giảm giá cho đơn hàng nháp, giữ trước một lượt sử dụng', '01M3SALES0000000000000000K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- The three parties to a sale are three separate powers, because a business splits them.
		--
		-- Recording who bought is counter work; redirecting who gets invoiced decides who owes the
		-- money, and is the one a finance team keeps to itself. Granting them together would mean
		-- anyone who may serve a customer may also bill a different company for the sale.
		('01M3SALES0000000000000009A', jsonb_build_object('en-US', 'Assign parties', 'vi-VN', 'Chỉ định các bên'), 'assign_parties', 'Chỉ định cùng lúc bất kỳ tổ hợp bên mua hàng, bên nhận hóa đơn và bên thanh toán cho giao dịch bán', '01M3SALES0000000000000000K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000009B', jsonb_build_object('en-US', 'Assign sold-to party', 'vi-VN', 'Chỉ định bên mua hàng'), 'assign_sold_to_party', 'Chỉ định đối tác mua hàng của giao dịch bán', '01M3SALES0000000000000000K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000009C', jsonb_build_object('en-US', 'Assign bill-to party', 'vi-VN', 'Chỉ định bên nhận hóa đơn'), 'assign_bill_to_party', 'Chỉ định đối tác đứng tên trên hóa đơn; bị từ chối khi chỉ dẫn lập hóa đơn đang được phát hành', '01M3SALES0000000000000000K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000009D', jsonb_build_object('en-US', 'Assign payer party', 'vi-VN', 'Chỉ định bên thanh toán'), 'assign_payer_party', 'Chỉ định đối tác chịu trách nhiệm thanh toán cho giao dịch bán', '01M3SALES0000000000000000K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_actions'
	) THEN
		INSERT INTO "iam_actions" ("id", "name", "code", "description", "resource_id", "etag") VALUES
		-- Confirm and cancel are separate powers from update, deliberately.
		--
		-- Confirming commits the business to a price and redeems vouchers; cancelling unwinds a sale
		-- and hands voucher uses back. A role that may correct a line should not thereby be able to
		-- do either, which is what folding them into `update` would grant.
		('01M3SALES0000000000000004H', jsonb_build_object('en-US', 'Confirm', 'vi-VN', 'Xác nhận'), 'confirm', 'Xác nhận đơn bán hàng nháp: cố định giá và sử dụng các phiếu giảm giá', '01M3SALES0000000000000000K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000004J', jsonb_build_object('en-US', 'Cancel', 'vi-VN', 'Hủy'), 'cancel', 'Hủy đơn bán hàng chưa thanh toán hoặc chưa giao hàng', '01M3SALES0000000000000000K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;


-- ============================================================================
-- Merged from 1007008_sales_bill_iam.sql
-- ============================================================================
-- IAM resources and actions for billing (SALES-024).
--
-- Three resources with deliberately different surfaces.
--
-- A BILL is operator-managed and gets full CRUD: somebody opens, corrects and cancels settlement
-- units. Its ALLOCATIONS get read alone - they are computed by split, merge and the initial bill,
-- and a client able to write one could make an order's bills stop summing to the order (BR 36),
-- which is the one invariant the whole billing model rests on. The LINEAGE is read alone for the
-- same reason as the audit trail: a client able to POST one could fabricate a paper trail showing a
-- payment settled a bill it never touched.

DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_resources'
	) THEN
		INSERT INTO "iam_resources" (
			"id", "name", "code", "description", "owner_type", "max_scope", "min_scope", "created_at", "etag"
		) VALUES
		('01M3SALES0000000000000004K', jsonb_build_object('en-US', 'Sales - Bill', 'vi-VN', 'Bán hàng - Hóa đơn'), 'sales_bill', 'Đơn vị thanh toán của một giao dịch bán - không phải hóa đơn GTGT (BR 33)', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000004M', jsonb_build_object('en-US', 'Sales - Bill Line', 'vi-VN', 'Bán hàng - Dòng hóa đơn'), 'sales_bill_line', 'Phần được phân bổ của một dòng đơn hàng trong một hóa đơn', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000004N', jsonb_build_object('en-US', 'Sales - Bill Relation', 'vi-VN', 'Bán hàng - Liên kết hóa đơn'), 'sales_bill_relation', 'Dấu vết liên kết để lại sau khi tách hoặc gộp hóa đơn', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_actions'
	) THEN
		INSERT INTO "iam_actions" ("id", "name", "code", "description", "resource_id", "etag") VALUES
		-- Sales Bill: full CRUD, an operator manages settlement.
		('01M3SALES0000000000000004P', jsonb_build_object('en-US', 'Create', 'vi-VN', 'Tạo'), 'create', NULL, '01M3SALES0000000000000004K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000004Q', jsonb_build_object('en-US', 'Update', 'vi-VN', 'Cập nhật'), 'update', NULL, '01M3SALES0000000000000004K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000004R', jsonb_build_object('en-US', 'Delete', 'vi-VN', 'Xóa'), 'delete', NULL, '01M3SALES0000000000000004K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000004S', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M3SALES0000000000000004K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000004T', jsonb_build_object('en-US', 'Set archived status', 'vi-VN', 'Đặt trạng thái lưu trữ'), 'set_archived', 'Lưu trữ hóa đơn bán hàng, hoặc khôi phục hóa đơn đã lưu trữ', '01M3SALES0000000000000004K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Allocations and lineage: read alone. See the note at the top of this file.
		('01M3SALES0000000000000004V', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M3SALES0000000000000004M', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000004W', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M3SALES0000000000000004N', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_actions'
	) THEN
		INSERT INTO "iam_actions" ("id", "name", "code", "description", "resource_id", "etag") VALUES
		-- Split and merge are separate powers from update: they restructure settlement units and
		-- write lineage, and a role that may correct a bill's details need not be able to divide it.
		('01M3SALES0000000000000004X', jsonb_build_object('en-US', 'Split', 'vi-VN', 'Tách'), 'split', 'Tách một hóa đơn đang mở thành nhiều hóa đơn', '01M3SALES0000000000000004K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000004Y', jsonb_build_object('en-US', 'Merge', 'vi-VN', 'Gộp'), 'merge', 'Gộp nhiều hóa đơn đang mở của cùng một đơn hàng thành một hóa đơn', '01M3SALES0000000000000004K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;


-- ============================================================================
-- Merged from 1007010_sales_payment_iam.sql
-- ============================================================================
-- IAM for payments (SALES-027/028).
--
-- The resource gets `read` alone. Money is recorded through the record_payment operation, which
-- applies six gates a plain POST would bypass - including the two that ask paymentinvoice whether
-- the method may be used at all (CR 33). A writable payment resource would let a client record money
-- that no gateway ever took.
--
-- `pay` and `settle` hang off the BILL, because that is what they change.

DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_resources'
	) THEN
		INSERT INTO "iam_resources" (
			"id", "name", "code", "description", "owner_type", "max_scope", "min_scope", "created_at", "etag"
		) VALUES
		('01M3SALES0000000000000004Z', jsonb_build_object('en-US', 'Sales - Payment', 'vi-VN', 'Bán hàng - Thanh toán'), 'sales_payment', 'Một lần dịch chuyển tiền đối với hóa đơn', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_actions'
	) THEN
		INSERT INTO "iam_actions" ("id", "name", "code", "description", "resource_id", "etag") VALUES
		('01M3SALES00000000000000050', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M3SALES0000000000000004Z', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000051', jsonb_build_object('en-US', 'Record payment', 'vi-VN', 'Ghi nhận thanh toán'), 'pay', 'Thu tiền cho hóa đơn đang mở', '01M3SALES0000000000000004K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000052', jsonb_build_object('en-US', 'Settle', 'vi-VN', 'Tất toán'), 'settle', 'Đóng hóa đơn đã thu đủ tiền', '01M3SALES0000000000000004K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;


-- ============================================================================
-- Merged from 1007012_sales_fulfillment_iam.sql
-- ============================================================================
-- IAM for fulfilment requests (SALES-029).
--
-- Both resources get `read` alone. Requests are raised by confirm and by the return workflow, and a
-- client able to write one could tell Inventory to move goods that no sale asked for - which is the
-- one thing BR 44's separation exists to prevent.

DO $$
BEGIN
	IF EXISTS (SELECT FROM information_schema.tables WHERE table_schema='public' AND table_name='iam_resources') THEN
		INSERT INTO "iam_resources" ("id","name","code","description","owner_type","max_scope","min_scope","created_at","etag") VALUES
		('01M3SALES00000000000000053', jsonb_build_object('en-US', 'Sales - Fulfilment Request', 'vi-VN', 'Bán hàng - Yêu cầu giao hàng'), 'sales_fulfillment_request', 'Một yêu cầu phân hệ Bán hàng gửi phân hệ Kho hàng cho một giao dịch bán', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000054', jsonb_build_object('en-US', 'Sales - Fulfilment Request Line', 'vi-VN', 'Bán hàng - Dòng yêu cầu giao hàng'), 'sales_fulfillment_request_line', 'Một dòng đơn hàng thuộc một yêu cầu giao hàng', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (SELECT FROM information_schema.tables WHERE table_schema='public' AND table_name='iam_actions') THEN
		INSERT INTO "iam_actions" ("id","name","code","description","resource_id","etag") VALUES
		('01M3SALES00000000000000055', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M3SALES00000000000000053', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000056', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M3SALES00000000000000054', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;


-- ============================================================================
-- Merged from 1007014_sales_fiscal_iam.sql
-- ============================================================================
-- IAM for fiscal document requests (SALES-030, SALES-031).
--
-- `read` and `create`, and nothing else. There is no `update` and no `delete`, and that omission is
-- the point: a client able to PATCH one of these rows could mark an unissued request as issued -
-- telling a customer they hold a VAT invoice that does not exist, which is precisely the state
-- BR 77 exists to keep honest - and a client able to DELETE one could erase the record of a legal
-- document that the provider still holds.
--
-- `create` is the permission the request_invoice action demands, and it hangs off THIS resource
-- rather than off sales_bill. Asking for a VAT invoice is not a power over the settlement: whoever
-- serves a customer who paid at a till and afterwards wants an invoice for their company need not
-- also be able to split, merge or settle the bill they are invoicing.

DO $$
BEGIN
	IF EXISTS (SELECT FROM information_schema.tables WHERE table_schema='public' AND table_name='iam_resources') THEN
		INSERT INTO "iam_resources" ("id","name","code","description","owner_type","max_scope","min_scope","created_at","etag") VALUES
		('01M3SALES00000000000000057', jsonb_build_object('en-US', 'Sales - Fiscal Request', 'vi-VN', 'Bán hàng - Yêu cầu hóa đơn điện tử'), 'sales_fiscal_request', 'Nội dung phân hệ Bán hàng yêu cầu nhà cung cấp hóa đơn điện tử và kết quả nhận về', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (SELECT FROM information_schema.tables WHERE table_schema='public' AND table_name='iam_actions') THEN
		INSERT INTO "iam_actions" ("id","name","code","description","resource_id","etag") VALUES
		('01M3SALES00000000000000058', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M3SALES00000000000000057', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000059', jsonb_build_object('en-US', 'Create', 'vi-VN', 'Tạo'), 'create', 'Yêu cầu nhà cung cấp hóa đơn điện tử phát hành chứng từ thuế hợp lệ', '01M3SALES00000000000000057', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;


-- ============================================================================
-- Merged from 1007016_sales_outbox_iam.sql
-- ============================================================================
-- IAM for the integration event outbox (SALES-037).
--
-- `read` alone. Rows are written by the domain services that produce the events, inside their own
-- transactions, and drained by the background sweep.
--
-- The omissions are the point. A client able to POST one could announce a sale that never happened
-- to every downstream consumer - stock released, money reconciled, invoices raised, all against
-- nothing. A client able to PATCH one could set published_at on an event that never went, which
-- deletes it from the queue as surely as a DELETE would, and silently: the row still looks correct.
--
-- It is readable at all because an operator investigating a consumer that has fallen behind needs to
-- see what Sales believes it published, when, and how many times it has tried.

DO $$
BEGIN
	IF EXISTS (SELECT FROM information_schema.tables WHERE table_schema='public' AND table_name='iam_resources') THEN
		INSERT INTO "iam_resources" ("id","name","code","description","owner_type","max_scope","min_scope","created_at","etag") VALUES
		('01M3SALES00000000000000060', jsonb_build_object('en-US', 'Sales - Integration Outbox', 'vi-VN', 'Bán hàng - Hộp thư đi tích hợp'), 'sales_integration_outbox', 'Các sự kiện tích hợp mà phân hệ Bán hàng đã phát hoặc đang chờ phát', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (SELECT FROM information_schema.tables WHERE table_schema='public' AND table_name='iam_actions') THEN
		INSERT INTO "iam_actions" ("id","name","code","description","resource_id","etag") VALUES
		('01M3SALES00000000000000061', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M3SALES00000000000000060', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;


-- ============================================================================
-- Merged from 1007018_sales_quotation_iam.sql
-- ============================================================================
-- IAM for quotations (SALES-038).
--
-- Full CRUD on the quotation, unlike the order: a quotation is a document an operator writes, and
-- creating or correcting one commits the business to nothing and moves no money. What is gated is
-- the STATUS, which is declared no_update and moves only through the actions.
--
-- `convert` is its own permission, not folded into `update`. Accepting a quotation CREATES A SALES
-- ORDER - it commits the business to a sale - and a role that may draft and correct an offer should
-- not thereby be able to turn one into a binding order. Sending and cancelling ride on `update`,
-- because both are ordinary handling of a document by whoever owns it.
--
-- The lines get CRUD too: editing what was offered is editing the quotation, and splitting the
-- permission would mean an operator who could write the header could not fill it in.

DO $$
BEGIN
	IF EXISTS (SELECT FROM information_schema.tables WHERE table_schema='public' AND table_name='iam_resources') THEN
		INSERT INTO "iam_resources" ("id","name","code","description","owner_type","max_scope","min_scope","created_at","etag") VALUES
		('01M3SALES00000000000000062', jsonb_build_object('en-US', 'Sales - Quotation', 'vi-VN', 'Bán hàng - Báo giá'), 'sales_quotation', 'Đề nghị gửi khách hàng, có thể trở thành đơn bán hàng', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000063', jsonb_build_object('en-US', 'Sales - Quotation Line', 'vi-VN', 'Bán hàng - Dòng báo giá'), 'sales_quotation_line', 'Một dòng của báo giá', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (SELECT FROM information_schema.tables WHERE table_schema='public' AND table_name='iam_actions') THEN
		INSERT INTO "iam_actions" ("id","name","code","description","resource_id","etag") VALUES
		('01M3SALES00000000000000064', jsonb_build_object('en-US', 'Create', 'vi-VN', 'Tạo'), 'create', NULL, '01M3SALES00000000000000062', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000065', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M3SALES00000000000000062', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000066', jsonb_build_object('en-US', 'Update', 'vi-VN', 'Cập nhật'), 'update', 'Sửa, gửi hoặc rút lại báo giá', '01M3SALES00000000000000062', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000067', jsonb_build_object('en-US', 'Delete', 'vi-VN', 'Xóa'), 'delete', NULL, '01M3SALES00000000000000062', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000068', jsonb_build_object('en-US', 'Convert', 'vi-VN', 'Chuyển thành đơn hàng'), 'convert', 'Chấp nhận báo giá và tạo đơn bán hàng từ báo giá đó', '01M3SALES00000000000000062', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000069', jsonb_build_object('en-US', 'Create', 'vi-VN', 'Tạo'), 'create', NULL, '01M3SALES00000000000000063', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000070', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M3SALES00000000000000063', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000071', jsonb_build_object('en-US', 'Update', 'vi-VN', 'Cập nhật'), 'update', NULL, '01M3SALES00000000000000063', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000072', jsonb_build_object('en-US', 'Delete', 'vi-VN', 'Xóa'), 'delete', NULL, '01M3SALES00000000000000063', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;


-- ============================================================================
-- Merged from 1007020_sales_manual_discount_iam.sql
-- ============================================================================
-- IAM for manual price overrides (SALES-039).
--
-- Two things are seeded here, and they sit on different resources.
--
-- `manual_discount` on SALES_ORDER is the power BR 87.4 gates: changing what a customer pays for
-- reasons outside the price list. It is deliberately NOT `update` — a role that may correct a
-- quantity should not thereby be able to discount a sale — and revoking takes the same permission as
-- granting, since withdrawing a discount changes the price just as granting one does.
--
-- `read` on SALES_MANUAL_DISCOUNT is the whole surface of the record itself. No create, no update,
-- no delete: rows are written by the grant operation and by nothing else, because a plain POST would
-- bypass every gate that makes an override auditable — the mandatory reason, the draft-only check,
-- and the audit entry recording both the old and the new price.

DO $$
BEGIN
	IF EXISTS (SELECT FROM information_schema.tables WHERE table_schema='public' AND table_name='iam_resources') THEN
		INSERT INTO "iam_resources" ("id","name","code","description","owner_type","max_scope","min_scope","created_at","etag") VALUES
		('01M3SALES00000000000000073', jsonb_build_object('en-US', 'Sales - Manual Discount', 'vi-VN', 'Bán hàng - Giảm giá thủ công'), 'sales_manual_discount', 'Việc nhân viên ghi đè giá bán, kèm lý do', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (SELECT FROM information_schema.tables WHERE table_schema='public' AND table_name='iam_actions') THEN
		INSERT INTO "iam_actions" ("id","name","code","description","resource_id","etag") VALUES
		('01M3SALES00000000000000074', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M3SALES00000000000000073', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;

		-- The power itself, on the ORDER, since that is what gets discounted. The resource id is
		-- written literally rather than looked up by code: the seed-coverage test parses these
		-- files without a database, so a SELECT-based insert would be invisible to it and an
		-- action could go unseeded without anything failing.
		INSERT INTO "iam_actions" ("id","name","code","description","resource_id","etag") VALUES
		('01M3SALES00000000000000075', jsonb_build_object('en-US', 'Manual discount', 'vi-VN', 'Giảm giá thủ công'), 'manual_discount', 'Thay đổi số tiền khách hàng phải trả ngoài bảng giá, kèm lý do', '01M3SALES0000000000000000K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;


-- ============================================================================
-- Merged from 1007023_sales_return_iam.sql
-- ============================================================================
-- IAM for returns, return lines and refund legs.
--
-- `process_return` is its own action rather than `update` because it moves money out of the
-- business: a role that may correct a return's reason must not thereby be able to release its
-- refund.
--
-- sales_return_line and sales_refund_payment are read-only to clients. Lines are written by
-- create_return, which caps each quantity at what is still returnable and prices it from the
-- historical amounts; refund legs are written by the return workflow, which caps each at what its
-- original payment captured. A writable row either way is an outflow with no matching inflow.

DO $$
BEGIN
	IF EXISTS (SELECT FROM information_schema.tables WHERE table_schema='public' AND table_name='iam_resources') THEN
		INSERT INTO "iam_resources" ("id","name","code","description","owner_type","max_scope","min_scope","created_at","etag") VALUES
		('01M3SALES00000000000000076', jsonb_build_object('en-US', 'Sales - Return', 'vi-VN', 'Bán hàng - Phiếu trả hàng bán'), 'sales_return', 'Việc khách hàng trả lại hàng và khoản hoàn tiền tương ứng', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000080', jsonb_build_object('en-US', 'Sales - Return Line', 'vi-VN', 'Bán hàng - Dòng phiếu trả hàng bán'), 'sales_return_line', 'Một dòng đơn hàng được trả lại, toàn bộ hoặc một phần', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000082', jsonb_build_object('en-US', 'Sales - Refund Payment', 'vi-VN', 'Bán hàng - Khoản hoàn tiền'), 'sales_refund_payment', 'Một phần của khoản hoàn tiền, ứng với một khoản thanh toán gốc', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (SELECT FROM information_schema.tables WHERE table_schema='public' AND table_name='iam_actions') THEN
		-- The return document: readable, creatable, correctable, and archivable.
		INSERT INTO "iam_actions" ("id","name","code","description","resource_id","etag") VALUES
		('01M3SALES00000000000000077', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M3SALES00000000000000076', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000078', jsonb_build_object('en-US', 'Create', 'vi-VN', 'Tạo'), 'create', 'Ghi nhận khách hàng muốn trả lại hàng', '01M3SALES00000000000000076', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000079', jsonb_build_object('en-US', 'Update', 'vi-VN', 'Cập nhật'), 'update', 'Sửa phiếu trả hàng, hoặc hủy phiếu khi chưa có hàng nào được dịch chuyển', '01M3SALES00000000000000076', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),

		-- The power that moves money and goods. Separate from `update` on purpose: this is the one
		-- that cannot be undone once it starts.
		('01M3SALES0000000000000007A', jsonb_build_object('en-US', 'Process return', 'vi-VN', 'Xử lý trả hàng'), 'process_return', 'Nhập lại hàng trả, hoàn tiền cho khách hàng và điều chỉnh hóa đơn GTGT', '01M3SALES00000000000000076', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),

		-- Read only; see the header.
		('01M3SALES00000000000000081', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M3SALES00000000000000080', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000083', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M3SALES00000000000000082', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;


-- ============================================================================
-- Merged from 1007027_sales_billing_instruction_iam.sql
-- ============================================================================
-- IAM for billing instructions and their issuance attempts.
--
-- `mark_ready` is its own action rather than `update`, because the two are materially different
-- powers. Correcting a tax code is clerical; marking ready is what releases a legal document to be
-- issued in a company's name, and a role that may fix a typo must not thereby be able to bill
-- someone.
--
-- `cancel` is likewise separate: withdrawing a buyer's request to be invoiced is not an edit of it.
--
-- sales_billing_issuance_attempt is READ-ONLY to clients. Attempts are written only by the issuance
-- job, and they are the evidence of whether a document was created — including the indeterminate
-- case where nobody knows. A client able to write one could fabricate a record showing an invoice
-- was issued, or erase the trace of one that may exist.

DO $$
BEGIN
	IF EXISTS (SELECT FROM information_schema.tables WHERE table_schema='public' AND table_name='iam_resources') THEN
		INSERT INTO "iam_resources" ("id","name","code","description","owner_type","max_scope","min_scope","created_at","etag") VALUES
		('01M3SALES00000000000000090', jsonb_build_object('en-US', 'Sales - Billing Instruction', 'vi-VN', 'Bán hàng - Chỉ dẫn lập hóa đơn'), 'sales_billing_instruction', 'Giao dịch bán được xuất hóa đơn cho ai, theo pháp nhân nào', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000091', jsonb_build_object('en-US', 'Sales - Billing Issuance Attempt', 'vi-VN', 'Bán hàng - Lần phát hành hóa đơn'), 'sales_billing_issuance_attempt', 'Một lần thử phát hành hóa đơn điện tử, được lưu làm bằng chứng', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (SELECT FROM information_schema.tables WHERE table_schema='public' AND table_name='iam_actions') THEN
		INSERT INTO "iam_actions" ("id","name","code","description","resource_id","etag") VALUES
		-- The instruction: readable, creatable, correctable while it is still a draft.
		('01M3SALES00000000000000092', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M3SALES00000000000000090', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000093', jsonb_build_object('en-US', 'Create', 'vi-VN', 'Tạo'), 'create', 'Ghi nhận người mua muốn lấy hóa đơn và thông tin đứng tên', '01M3SALES00000000000000090', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000094', jsonb_build_object('en-US', 'Update', 'vi-VN', 'Cập nhật'), 'update', 'Sửa thông tin thuế trước khi chứng từ được phát hành', '01M3SALES00000000000000090', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),

		-- The power that releases a document to be issued. See the note at the top of this file.
		('01M3SALES00000000000000095', jsonb_build_object('en-US', 'Mark ready', 'vi-VN', 'Đánh dấu sẵn sàng'), 'mark_ready', 'Xác nhận thông tin thuế và chuyển chỉ dẫn đi phát hành', '01M3SALES00000000000000090', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000096', jsonb_build_object('en-US', 'Revert to draft', 'vi-VN', 'Chuyển về nháp'), 'revert_to_draft', 'Thu hồi chỉ dẫn đã chuyển đi để sửa tiếp', '01M3SALES00000000000000090', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000097', jsonb_build_object('en-US', 'Cancel', 'vi-VN', 'Hủy'), 'cancel', 'Rút lại yêu cầu lấy hóa đơn của người mua, trước khi có chứng từ nào', '01M3SALES00000000000000090', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- refresh_from_bill_to was seeded here with no route or assert behind it; removed so the seed
		-- advertises no power the API does not offer. Re-add it alongside the action that checks it.

		-- Attempts: read alone. See the note at the top of this file.
		('01M3SALES00000000000000099', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M3SALES00000000000000091', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;

-- ---------------------------------------------------------------------------
-- Merged from 1007005_sales_fulfillment_method_iam.sql
-- ---------------------------------------------------------------------------

-- IAM for the fulfillment method catalogue.
--
-- Five actions, the standard CRUD set plus set_archived, which is the permission BOTH the archive
-- and unarchive routes answer to: withdrawing a policy from new orders and restoring it are the same
-- power in reverse, and splitting them would let a role retire a method it could not put back.
--
-- There is no suspend action here, unlike sales_channel and sales_point. A method is either offered
-- to new orders or it is not, and the fulfillments that already snapshotted it keep running either
-- way, so a second "temporarily off" state would say nothing is_archived does not already say.
--
-- sales_channel_fulfillment_methods gets no resource row at all: the mapping is configured through
-- the channel that owns it and nothing routes to it, exactly as sales_channel_payment_rel is
-- handled. A resource row for it would advertise a permission over an endpoint that does not exist.

DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_resources'
	) THEN
		INSERT INTO "iam_resources" (
			"id", "name", "code", "description", "owner_type", "max_scope", "min_scope", "created_at", "etag"
		) VALUES
		('01M3SALES0000000000000009E', jsonb_build_object('en-US', 'Sales - Fulfillment Method', 'vi-VN', 'Bán hàng - Phương thức giao hàng'), 'sales_fulfillment_method', 'Chính sách giao hàng của giao dịch bán: hàng đến tay khách hàng bằng cách nào và xử lý ra sao khi thất bại', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_actions'
	) THEN
		INSERT INTO "iam_actions" ("id", "name", "code", "description", "resource_id", "etag") VALUES
		('01M3SALES0000000000000009F', jsonb_build_object('en-US', 'Create', 'vi-VN', 'Tạo'), 'create', NULL, '01M3SALES0000000000000009E', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000009G', jsonb_build_object('en-US', 'Update', 'vi-VN', 'Cập nhật'), 'update', NULL, '01M3SALES0000000000000009E', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000009H', jsonb_build_object('en-US', 'Delete', 'vi-VN', 'Xóa'), 'delete', NULL, '01M3SALES0000000000000009E', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000009J', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M3SALES0000000000000009E', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000009K', jsonb_build_object('en-US', 'Set archived status', 'vi-VN', 'Đặt trạng thái lưu trữ'), 'set_archived', 'Ngừng áp dụng phương thức giao hàng cho đơn hàng mới, hoặc áp dụng lại; các lượt giao hàng đang chạy không bị ảnh hưởng', '01M3SALES0000000000000009E', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;

-- ---------------------------------------------------------------------------
-- Order fulfillments (SFL-016, SFL-024)
-- ---------------------------------------------------------------------------

-- IAM for what a kiosk owes a customer.
--
-- Read alone, for both resources. A fulfillment is created by confirming an order and moved only by
-- the fulfillment services: every one of its columns is either a policy snapshot the customer bought
-- under, or a status that must not move without the stock moving with it. Granting create or update
-- would let a holder declare goods delivered that never left a machine, or rewrite the refund policy
-- a sale was made under after the fact.
--
-- The item resource is separate rather than folded into its parent because the two answer different
-- questions — "what is this delivery doing" against "how much of this product is still owed" — and a
-- report over quantities should not require the power to read every fulfillment's target and policy.

DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_resources'
	) THEN
		INSERT INTO "iam_resources" (
			"id", "name", "code", "description", "owner_type", "max_scope", "min_scope", "created_at", "etag"
		) VALUES
		('01M3SALES0000000000000009M', jsonb_build_object('en-US', 'Sales - Order Fulfillment', 'vi-VN', 'Bán hàng - Giao hàng đơn hàng'), 'sales_order_fulfillment', 'Một lượt giao hàng mà đơn bán hàng cần thực hiện: giao ở đâu, theo chính sách nào và đã tiến triển đến đâu', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000009N', jsonb_build_object('en-US', 'Sales - Order Fulfillment Item', 'vi-VN', 'Bán hàng - Mặt hàng giao của đơn hàng'), 'sales_order_fulfillment_item', 'Một sản phẩm trong lượt giao hàng: số lượng đã đặt, đã giao, đã hoàn tiền và còn nợ', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_actions'
	) THEN
		INSERT INTO "iam_actions" ("id", "name", "code", "description", "resource_id", "etag") VALUES
		('01M3SALES0000000000000009P', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M3SALES0000000000000009M', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Update is what the attempt and reassignment actions assert (see the attempts note below):
		-- commanding a dispense, recording its result and moving the target all change what this
		-- delivery owes. Without the row no role built from the seed could do any of them.
		('01M3SALES0000000000000009Y', jsonb_build_object('en-US', 'Update', 'vi-VN', 'Cập nhật'), 'update', 'Ra lệnh hoặc ghi nhận một lần giao hàng, hoặc đổi nơi thực hiện giao hàng', '01M3SALES0000000000000009M', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000009Q', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M3SALES0000000000000009N', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;

-- ---------------------------------------------------------------------------
-- Fulfillment attempts (SFL-026, SFL-031)
-- ---------------------------------------------------------------------------

-- IAM for the evidence of what a machine was asked to do and what it did.
--
-- The attempt records themselves are READ-only: they are evidence about a physical event, and a
-- holder able to write one could claim a machine dispensed goods it never did, which is the one lie
-- this whole feature is built to make impossible. Attempts are created and settled by the services
-- behind the two actions below, never by CRUD.
--
-- Those two actions live on sales_order_fulfillment and answer its `update` permission rather than
-- carrying their own: commanding a dispense and recording its result both change what a delivery
-- owes, which is the same power over the same sale. A separate action code would let a role report
-- results it could not cause, and the pairing is what keeps the two together.

DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_resources'
	) THEN
		INSERT INTO "iam_resources" (
			"id", "name", "code", "description", "owner_type", "max_scope", "min_scope", "created_at", "etag"
		) VALUES
		('01M3SALES0000000000000009R', jsonb_build_object('en-US', 'Sales - Fulfillment Attempt', 'vi-VN', 'Bán hàng - Lần thực hiện giao hàng'), 'sales_fulfillment_attempt', 'Một lần thử giao hàng của lượt giao hàng: máy nào được yêu cầu, khi nào và kết quả ra sao', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000009S', jsonb_build_object('en-US', 'Sales - Fulfillment Attempt Item', 'vi-VN', 'Bán hàng - Mặt hàng trong lần giao hàng'), 'sales_fulfillment_attempt_item', 'Kết quả một lần thử ghi nhận cho một sản phẩm: bao nhiêu đã ra, bao nhiêu chưa ra và lý do', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_actions'
	) THEN
		INSERT INTO "iam_actions" ("id", "name", "code", "description", "resource_id", "etag") VALUES
		('01M3SALES0000000000000009T', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M3SALES0000000000000009R', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000009V', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M3SALES0000000000000009S', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;

-- ---------------------------------------------------------------------------
-- Fulfillment target changes (SFL-041)
-- ---------------------------------------------------------------------------

-- IAM for the record of where a delivery has been promised from.
--
-- Read alone. The rows are an audit trail: they record something that already happened, written in
-- the same transaction that moved the target, and a row a holder could create or edit would be
-- evidence of nothing. The power to MOVE a target is `update` on sales_order_fulfillment, which is
-- where the reassignment action lives -- granting it here as well would let a role rewrite the
-- history of a move it was not allowed to make.

DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_resources'
	) THEN
		INSERT INTO "iam_resources" (
			"id", "name", "code", "description", "owner_type", "max_scope", "min_scope", "created_at", "etag"
		) VALUES
		('01M3SALES0000000000000009W', jsonb_build_object('en-US', 'Sales - Fulfillment Target Change', 'vi-VN', 'Bán hàng - Thay đổi đích giao hàng'), 'sales_fulfillment_target_change', 'Lịch sử nơi cam kết giao hàng: những điểm nào, vì sao thay đổi và ai thay đổi', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_actions'
	) THEN
		INSERT INTO "iam_actions" ("id", "name", "code", "description", "resource_id", "etag") VALUES
		('01M3SALES0000000000000009X', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M3SALES0000000000000009W', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;

-- Confirming a refund request is business approval, distinct from processing it: it records the
-- confirmer's note, locks the amount against the order's other pending requests and only then
-- dispatches. Its own action, because approving money out is not the same power as raising the
-- request (CR-INV-SALES-WH-RESERVATION §6.3).
DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_actions'
	) THEN
		INSERT INTO "iam_actions" ("id", "name", "code", "description", "resource_id", "etag") VALUES
		('01M3SALESRFCONFIRM00000001', jsonb_build_object('en-US', 'Confirm refund request', 'vi-VN', 'Xác nhận yêu cầu hoàn tiền'), 'confirm_return', 'Phê duyệt yêu cầu hoàn tiền, ghi chú và thực hiện hoàn tiền', '01M3SALES00000000000000076', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;

-- Promotion and voucher roles (PROMO-003).
-- A reader sees promotions, vouchers and what orders received; a manager configures, activates,
-- suspends, generates and revokes. The right to process orders grants neither (BR §9.4).
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
		('01M3ROLE000000000000PROMO1', 'Xem khuyến mãi', 'Xem chương trình khuyến mãi, mã khuyến mãi và kết quả áp dụng, không được thay đổi', false, '01JWNMZ36QHC7CQQ748H9NQ6J6', true, false, true, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ROLE000000000000PROMO2', 'Quản lý khuyến mãi', 'Tạo, sửa, kích hoạt, tạm ngưng chương trình khuyến mãi; sinh và thu hồi mã khuyến mãi', false, '01JWNMZ36QHC7CQQ748H9NQ6J6', true, false, true, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_entitlements'
	) THEN
		INSERT INTO "iam_entitlements" (
			"id", "name", "description", "expression", "action_id", "resource_id",
			"role_id", "scope", "org_id", "org_unit_id", "is_archived", "created_at", "etag"
		) VALUES
		('01M3ENT0000000000PROMO0001', 'Tạo sales_promotion', 'Tạo sales_promotion', 'create:sales_promotion:org', '01M3SALESPRMA0340000000000', '01M3SALES00000000000000034', '01M3ROLE000000000000PROMO2', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT0000000000PROMO0002', 'Cập nhật sales_promotion', 'Cập nhật sales_promotion', 'update:sales_promotion:org', '01M3SALESPRMA0340000000001', '01M3SALES00000000000000034', '01M3ROLE000000000000PROMO2', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT0000000000PROMO0003', 'Xóa sales_promotion', 'Xóa sales_promotion', 'delete:sales_promotion:org', '01M3SALESPRMA0340000000002', '01M3SALES00000000000000034', '01M3ROLE000000000000PROMO2', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT0000000000PROMO0004', 'Xem sales_promotion', 'Xem sales_promotion', 'read:sales_promotion:org', '01M3SALESPRMA0340000000003', '01M3SALES00000000000000034', '01M3ROLE000000000000PROMO2', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT0000000000PROMO0005', 'Xem sales_promotion', 'Xem sales_promotion', 'read:sales_promotion:org', '01M3SALESPRMA0340000000003', '01M3SALES00000000000000034', '01M3ROLE000000000000PROMO1', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT0000000000PROMO0006', 'Đặt trạng thái lưu trữ sales_promotion', 'Đặt trạng thái lưu trữ sales_promotion', 'set_archived:sales_promotion:org', '01M3SALESPRMA0340000000004', '01M3SALES00000000000000034', '01M3ROLE000000000000PROMO2', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT0000000000PROMO0007', 'Kích hoạt sales_promotion', 'Kích hoạt sales_promotion', 'activate:sales_promotion:org', '01M3SALESPRMA0340000000005', '01M3SALES00000000000000034', '01M3ROLE000000000000PROMO2', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT0000000000PROMO0008', 'Tạm ngưng sales_promotion', 'Tạm ngưng sales_promotion', 'suspend:sales_promotion:org', '01M3SALESPRMA0340000000006', '01M3SALES00000000000000034', '01M3ROLE000000000000PROMO2', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT0000000000PROMO0009', 'Tạo sales_promotion_rule', 'Tạo sales_promotion_rule', 'create:sales_promotion_rule:org', '01M3SALESPRMA0350000000000', '01M3SALES00000000000000035', '01M3ROLE000000000000PROMO2', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT0000000000PROMO0010', 'Cập nhật sales_promotion_rule', 'Cập nhật sales_promotion_rule', 'update:sales_promotion_rule:org', '01M3SALESPRMA0350000000001', '01M3SALES00000000000000035', '01M3ROLE000000000000PROMO2', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT0000000000PROMO0011', 'Xóa sales_promotion_rule', 'Xóa sales_promotion_rule', 'delete:sales_promotion_rule:org', '01M3SALESPRMA0350000000002', '01M3SALES00000000000000035', '01M3ROLE000000000000PROMO2', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT0000000000PROMO0012', 'Xem sales_promotion_rule', 'Xem sales_promotion_rule', 'read:sales_promotion_rule:org', '01M3SALESPRMA0350000000003', '01M3SALES00000000000000035', '01M3ROLE000000000000PROMO2', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT0000000000PROMO0013', 'Xem sales_promotion_rule', 'Xem sales_promotion_rule', 'read:sales_promotion_rule:org', '01M3SALESPRMA0350000000003', '01M3SALES00000000000000035', '01M3ROLE000000000000PROMO1', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT0000000000PROMO0014', 'Xem sales_promotion_rule_product', 'Xem sales_promotion_rule_product', 'read:sales_promotion_rule_product:org', '01M3SALESPRMA0360000000000', '01M3SALES00000000000000036', '01M3ROLE000000000000PROMO2', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT0000000000PROMO0015', 'Xem sales_promotion_rule_product', 'Xem sales_promotion_rule_product', 'read:sales_promotion_rule_product:org', '01M3SALESPRMA0360000000000', '01M3SALES00000000000000036', '01M3ROLE000000000000PROMO1', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT0000000000PROMO0016', 'Xem sales_promotion_rule_category', 'Xem sales_promotion_rule_category', 'read:sales_promotion_rule_category:org', '01M3SALESPRMA0370000000000', '01M3SALES00000000000000037', '01M3ROLE000000000000PROMO2', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT0000000000PROMO0017', 'Xem sales_promotion_rule_category', 'Xem sales_promotion_rule_category', 'read:sales_promotion_rule_category:org', '01M3SALESPRMA0370000000000', '01M3SALES00000000000000037', '01M3ROLE000000000000PROMO1', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT0000000000PROMO0018', 'Tạo sales_promotion_reward', 'Tạo sales_promotion_reward', 'create:sales_promotion_reward:org', '01M3SALESPRMA0380000000000', '01M3SALES00000000000000038', '01M3ROLE000000000000PROMO2', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT0000000000PROMO0019', 'Cập nhật sales_promotion_reward', 'Cập nhật sales_promotion_reward', 'update:sales_promotion_reward:org', '01M3SALESPRMA0380000000001', '01M3SALES00000000000000038', '01M3ROLE000000000000PROMO2', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT0000000000PROMO0020', 'Xóa sales_promotion_reward', 'Xóa sales_promotion_reward', 'delete:sales_promotion_reward:org', '01M3SALESPRMA0380000000002', '01M3SALES00000000000000038', '01M3ROLE000000000000PROMO2', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT0000000000PROMO0021', 'Xem sales_promotion_reward', 'Xem sales_promotion_reward', 'read:sales_promotion_reward:org', '01M3SALESPRMA0380000000003', '01M3SALES00000000000000038', '01M3ROLE000000000000PROMO2', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT0000000000PROMO0022', 'Xem sales_promotion_reward', 'Xem sales_promotion_reward', 'read:sales_promotion_reward:org', '01M3SALESPRMA0380000000003', '01M3SALES00000000000000038', '01M3ROLE000000000000PROMO1', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT0000000000PROMO0023', 'Xem sales_promotion_reward_product', 'Xem sales_promotion_reward_product', 'read:sales_promotion_reward_product:org', '01M3SALESPRMA0390000000000', '01M3SALES00000000000000039', '01M3ROLE000000000000PROMO2', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT0000000000PROMO0024', 'Xem sales_promotion_reward_product', 'Xem sales_promotion_reward_product', 'read:sales_promotion_reward_product:org', '01M3SALESPRMA0390000000000', '01M3SALES00000000000000039', '01M3ROLE000000000000PROMO1', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT0000000000PROMO0025', 'Xem sales_promotion_reward_category', 'Xem sales_promotion_reward_category', 'read:sales_promotion_reward_category:org', '01M3SALESPRMA00A0000000000', '01M3SALESPRM0000000000000A', '01M3ROLE000000000000PROMO2', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT0000000000PROMO0026', 'Xem sales_promotion_reward_category', 'Xem sales_promotion_reward_category', 'read:sales_promotion_reward_category:org', '01M3SALESPRMA00A0000000000', '01M3SALESPRM0000000000000A', '01M3ROLE000000000000PROMO1', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT0000000000PROMO0027', 'Tạo sales_voucher', 'Tạo sales_voucher', 'create:sales_voucher:org', '01M3SALESPRMA0480000000000', '01M3SALES00000000000000048', '01M3ROLE000000000000PROMO2', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT0000000000PROMO0028', 'Cập nhật sales_voucher', 'Cập nhật sales_voucher', 'update:sales_voucher:org', '01M3SALESPRMA0480000000001', '01M3SALES00000000000000048', '01M3ROLE000000000000PROMO2', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT0000000000PROMO0029', 'Xóa sales_voucher', 'Xóa sales_voucher', 'delete:sales_voucher:org', '01M3SALESPRMA0480000000002', '01M3SALES00000000000000048', '01M3ROLE000000000000PROMO2', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT0000000000PROMO0030', 'Xem sales_voucher', 'Xem sales_voucher', 'read:sales_voucher:org', '01M3SALESPRMA0480000000003', '01M3SALES00000000000000048', '01M3ROLE000000000000PROMO2', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT0000000000PROMO0031', 'Xem sales_voucher', 'Xem sales_voucher', 'read:sales_voucher:org', '01M3SALESPRMA0480000000003', '01M3SALES00000000000000048', '01M3ROLE000000000000PROMO1', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT0000000000PROMO0032', 'Đặt trạng thái lưu trữ sales_voucher', 'Đặt trạng thái lưu trữ sales_voucher', 'set_archived:sales_voucher:org', '01M3SALESPRMA0480000000004', '01M3SALES00000000000000048', '01M3ROLE000000000000PROMO2', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT0000000000PROMO0033', 'Sinh mã hàng loạt sales_voucher', 'Sinh mã hàng loạt sales_voucher', 'generate:sales_voucher:org', '01M3SALESPRMA0480000000005', '01M3SALES00000000000000048', '01M3ROLE000000000000PROMO2', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT0000000000PROMO0034', 'Thu hồi sales_voucher', 'Thu hồi sales_voucher', 'revoke:sales_voucher:org', '01M3SALESPRMA0480000000006', '01M3SALES00000000000000048', '01M3ROLE000000000000PROMO2', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT0000000000PROMO0035', 'Xem sales_promotion_application', 'Xem sales_promotion_application', 'read:sales_promotion_application:org', '01M3SALESPRMA0490000000000', '01M3SALES00000000000000049', '01M3ROLE000000000000PROMO2', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT0000000000PROMO0036', 'Xem sales_promotion_application', 'Xem sales_promotion_application', 'read:sales_promotion_application:org', '01M3SALESPRMA0490000000000', '01M3SALES00000000000000049', '01M3ROLE000000000000PROMO1', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT0000000000PROMO0037', 'Xem sales_promotion_allocation', 'Xem sales_promotion_allocation', 'read:sales_promotion_allocation:org', '01M3SALESPRMA00B0000000000', '01M3SALESPRM0000000000000B', '01M3ROLE000000000000PROMO2', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT0000000000PROMO0038', 'Xem sales_promotion_allocation', 'Xem sales_promotion_allocation', 'read:sales_promotion_allocation:org', '01M3SALESPRMA00B0000000000', '01M3SALESPRM0000000000000B', '01M3ROLE000000000000PROMO1', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT0000000000PROMO0039', 'Xem sales_voucher_issuance', 'Xem sales_voucher_issuance', 'read:sales_voucher_issuance:org', '01M3SALESPRMA00C0000000000', '01M3SALESPRM0000000000000C', '01M3ROLE000000000000PROMO2', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT0000000000PROMO0040', 'Xem sales_voucher_issuance', 'Xem sales_voucher_issuance', 'read:sales_voucher_issuance:org', '01M3SALESPRMA00C0000000000', '01M3SALESPRM0000000000000C', '01M3ROLE000000000000PROMO1', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT0000000000PROMO0041', 'Xem sales_order_voucher', 'Xem sales_order_voucher', 'read:sales_order_voucher:org', '01M3SALESPRMA00D0000000000', '01M3SALESPRM0000000000000D', '01M3ROLE000000000000PROMO2', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT0000000000PROMO0042', 'Xem sales_order_voucher', 'Xem sales_order_voucher', 'read:sales_order_voucher:org', '01M3SALESPRMA00D0000000000', '01M3SALESPRM0000000000000D', '01M3ROLE000000000000PROMO1', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT0000000000PROMO0043', 'Xem sales_order_promotion_selection', 'Xem sales_order_promotion_selection', 'read:sales_order_promotion_selection:org', '01M3SALESPRMA00E0000000000', '01M3SALESPRM0000000000000E', '01M3ROLE000000000000PROMO2', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3ENT0000000000PROMO0044', 'Xem sales_order_promotion_selection', 'Xem sales_order_promotion_selection', 'read:sales_order_promotion_selection:org', '01M3SALESPRMA00E0000000000', '01M3SALESPRM0000000000000E', '01M3ROLE000000000000PROMO1', 'org', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;
