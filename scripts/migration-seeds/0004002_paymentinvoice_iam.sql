-- IAM resources and actions for the Payment & Invoice module.
--
-- The dynamic resource engine asserts permissions using the schema name as the resource code, so
-- these codes must stay byte-identical to the "paymentinvoice_payment_method",
-- "paymentinvoice_payment_profile", "paymentinvoice_order", "paymentinvoice_transaction",
-- "paymentinvoice_invoice" and "paymentinvoice_invoice_line" schema names. A code that drifts from
-- its schema denies every request, with nothing in the response pointing at the seed.
--
-- Payment Order carries no create action. An order is not typed in: it comes into existence only
-- through create_payment, which generates its identifiers and asks the gateway for the payment
-- instrument in the same step. Seeding create would advertise a way to record an order that no
-- gateway knows about, which is a payment that can never be collected or reconciled.
--
-- create_payment and refund are separate actions rather than folded into one, and neither is
-- update. They are materially different powers: refund moves money back out of the business and
-- cannot be undone, while create_payment only asks for money. A role that may take payments should
-- not thereby be able to return them.
--
-- Transaction is read only. Its rows are the audit trail of what each gateway was asked and what
-- it answered, written by the payment flow and the callbacks; a row that a person could write or
-- edit would be worthless as evidence in a dispute, which is the only reason it is kept.
--
-- Invoice's issue is separate from update for the same reason as refund: issuing assigns a number
-- from a gap-free sequence and freezes the totals, and an issued invoice is an accounting document
-- rather than a draft someone may still correct.
--
-- Payment Profile is a resource of its own rather than a corner of Payment Method. A method says
-- what a payer may choose; a profile holds the merchant credentials the money settles into, and
-- reading one is not the same power as reading the other.
--
-- Deliberately no iam_entitlements rows. Payment and invoice data is not universally readable —
-- it contains customer identities, tax codes and amounts — so access follows explicitly assigned
-- roles rather than a blanket grant.

DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_resources'
	) THEN
		INSERT INTO "iam_resources" (
			"id", "name", "code", "description", "owner_type", "max_scope", "min_scope", "created_at", "etag"
		) VALUES
		('01M0PAY1SMY4Q2E7VC4DDZAPWF', jsonb_build_object('en-US', 'Payment & Invoice - Payment Order', 'vi-VN', 'Thanh toán & Hóa đơn - Lệnh thanh toán'), 'paymentinvoice_order', 'Yêu cầu thu tiền qua cổng thanh toán', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0PAY1V59JZP88YNHA9FEC4F', jsonb_build_object('en-US', 'Payment & Invoice - Payment Transaction', 'vi-VN', 'Thanh toán & Hóa đơn - Giao dịch thanh toán'), 'paymentinvoice_transaction', 'Một lần thanh toán hoặc hoàn tiền được ghi nhận cho đơn hàng', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0PAY16EZ1JQ1TS64Q21TPE7', jsonb_build_object('en-US', 'Payment & Invoice - Invoice', 'vi-VN', 'Thanh toán & Hóa đơn - Hóa đơn'), 'paymentinvoice_invoice', 'Chứng từ kế toán được phát hành cho một giao dịch bán', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0PAY13K1F11EGR7EET7DHAM', jsonb_build_object('en-US', 'Payment & Invoice - Invoice Line', 'vi-VN', 'Thanh toán & Hóa đơn - Dòng hóa đơn'), 'paymentinvoice_invoice_line', 'Một khoản tính tiền trong hóa đơn', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0PAY1PRF7Z8QK3M2N4T5V6W', jsonb_build_object('en-US', 'Payment & Invoice - Payment Profile', 'vi-VN', 'Thanh toán & Hóa đơn - Hồ sơ thanh toán'), 'paymentinvoice_payment_profile', 'Một tài khoản người bán tại cổng thanh toán: thông tin xác thực của tài khoản nhận tiền', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0PAY1GEVS6GFS22NP6V3ZRW', jsonb_build_object('en-US', 'Payment & Invoice - Payment Method', 'vi-VN', 'Thanh toán & Hóa đơn - Phương thức thanh toán'), 'paymentinvoice_payment_method', 'Một cách thanh toán đã được cấu hình cho doanh nghiệp, kèm bộ kết nối thực hiện thanh toán', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_actions'
	) THEN
		INSERT INTO "iam_actions" ("id", "name", "code", "description", "resource_id", "etag") VALUES
		-- Payment Method. Reference data: adding a gateway account or withdrawing one is a row
		-- rather than a release, which is the whole point of it not being an enum.
		('01M0PAY1Z535174MEV75RNHAVA', jsonb_build_object('en-US', 'Create', 'vi-VN', 'Tạo'), 'create', NULL, '01M0PAY1GEVS6GFS22NP6V3ZRW', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0PAY1PSDC0EYRH48TMMAA74', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M0PAY1GEVS6GFS22NP6V3ZRW', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0PAY1RSEQZJJHXZ2FMVG54N', jsonb_build_object('en-US', 'Update', 'vi-VN', 'Cập nhật'), 'update', NULL, '01M0PAY1GEVS6GFS22NP6V3ZRW', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0PAY1QHHPQ1HBD5ABESXBKS', jsonb_build_object('en-US', 'Delete', 'vi-VN', 'Xóa'), 'delete', NULL, '01M0PAY1GEVS6GFS22NP6V3ZRW', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0PAY11YGG1SE82N91KSX4MX', jsonb_build_object('en-US', 'Set archived status', 'vi-VN', 'Đặt trạng thái lưu trữ'), 'set_archived', 'Lưu trữ phương thức thanh toán để loại khỏi danh sách đang sử dụng', '01M0PAY1GEVS6GFS22NP6V3ZRW', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),

		-- Payment Profile. The credentials a payment settles into, so read is a materially larger
		-- grant here than on a payment method: reading one profile hands back the merchant secrets in
		-- the clear, and so does a search that names the config field.
		('01M0PAY1PRF7Z8QK3M2N4T5V70', jsonb_build_object('en-US', 'Create', 'vi-VN', 'Tạo'), 'create', NULL, '01M0PAY1PRF7Z8QK3M2N4T5V6W', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0PAY1PRF7Z8QK3M2N4T5V71', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', 'Xem hồ sơ, bao gồm thông tin xác thực cổng thanh toán', '01M0PAY1PRF7Z8QK3M2N4T5V6W', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0PAY1PRF7Z8QK3M2N4T5V72', jsonb_build_object('en-US', 'Update', 'vi-VN', 'Cập nhật'), 'update', NULL, '01M0PAY1PRF7Z8QK3M2N4T5V6W', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0PAY1PRF7Z8QK3M2N4T5V73', jsonb_build_object('en-US', 'Delete', 'vi-VN', 'Xóa'), 'delete', NULL, '01M0PAY1PRF7Z8QK3M2N4T5V6W', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0PAY1PRF7Z8QK3M2N4T5V74', jsonb_build_object('en-US', 'Set archived status', 'vi-VN', 'Đặt trạng thái lưu trữ'), 'set_archived', 'Ngừng sử dụng tài khoản người bán mà không mất các khoản thanh toán đã thu', '01M0PAY1PRF7Z8QK3M2N4T5V6W', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),

		-- Payment Order. No create: see the note at the top of this file.
		('01M0PAY10AA5Y9QZ20ZKN9CQ89', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M0PAY1SMY4Q2E7VC4DDZAPWF', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0PAY1ZFBX5PXV06Z0C12PQH', jsonb_build_object('en-US', 'Update', 'vi-VN', 'Cập nhật'), 'update', 'Sửa các trường mô tả của đơn hàng; không bao giờ sửa số tiền hay trạng thái', '01M0PAY1SMY4Q2E7VC4DDZAPWF', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0PAY1431QXXE8YQ1T0EMF6A', jsonb_build_object('en-US', 'Delete', 'vi-VN', 'Xóa'), 'delete', NULL, '01M0PAY1SMY4Q2E7VC4DDZAPWF', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0PAY1B6W0ZMV0FYA841TK27', jsonb_build_object('en-US', 'Take payment', 'vi-VN', 'Thu tiền'), 'create_payment', 'Mở lệnh thanh toán và yêu cầu cổng thanh toán cung cấp phương tiện thanh toán', '01M0PAY1SMY4Q2E7VC4DDZAPWF', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0PAY15MWQ2C0ZT4Y516XR83', jsonb_build_object('en-US', 'Refund', 'vi-VN', 'Hoàn tiền'), 'refund', 'Hoàn tiền cho đơn hàng đã thanh toán', '01M0PAY1SMY4Q2E7VC4DDZAPWF', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0PAY1HAVGWD8TKTK7KBZASV', jsonb_build_object('en-US', 'Clear terminal orders', 'vi-VN', 'Dọn đơn hàng trên máy POS'), 'remove_pos_orders', 'Xóa các đơn hàng đang chờ trên một máy POS thanh toán thẻ', '01M0PAY1SMY4Q2E7VC4DDZAPWF', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),

		-- Payment Transaction. Read only: see the note at the top of this file.
		('01M0PAY1SK0G6D93AX36PAN8G6', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M0PAY1V59JZP88YNHA9FEC4F', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),

		-- Invoice
		('01M0PAY1PSC5C510R71NK4S771', jsonb_build_object('en-US', 'Create', 'vi-VN', 'Tạo'), 'create', NULL, '01M0PAY16EZ1JQ1TS64Q21TPE7', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0PAY1FGC5MGT15EF15VYP82', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M0PAY16EZ1JQ1TS64Q21TPE7', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0PAY1QC6AXW0XZ3YVDASK7G', jsonb_build_object('en-US', 'Update', 'vi-VN', 'Cập nhật'), 'update', NULL, '01M0PAY16EZ1JQ1TS64Q21TPE7', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0PAY1938DEYJVVE8T3MZG6C', jsonb_build_object('en-US', 'Delete', 'vi-VN', 'Xóa'), 'delete', NULL, '01M0PAY16EZ1JQ1TS64Q21TPE7', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0PAY1AVSMSG28BHSAVAJG2R', jsonb_build_object('en-US', 'Set archived status', 'vi-VN', 'Đặt trạng thái lưu trữ'), 'set_archived', 'Lưu trữ hóa đơn để loại khỏi danh sách đang sử dụng', '01M0PAY16EZ1JQ1TS64Q21TPE7', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0PAY1DS35GY6SWYQ8YEEZV2', jsonb_build_object('en-US', 'Issue', 'vi-VN', 'Phát hành'), 'issue', 'Chốt bản nháp: cấp số, cố định tổng tiền và ghi ngày phát hành', '01M0PAY16EZ1JQ1TS64Q21TPE7', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),

		-- Invoice Line
		('01M0PAY11935H0BKW0TFNJNNHW', jsonb_build_object('en-US', 'Create', 'vi-VN', 'Tạo'), 'create', NULL, '01M0PAY13K1F11EGR7EET7DHAM', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0PAY18KMV8VEXQ78F5QBHAN', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M0PAY13K1F11EGR7EET7DHAM', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0PAY1PSKGBGKYZ9K7Q1BGJC', jsonb_build_object('en-US', 'Update', 'vi-VN', 'Cập nhật'), 'update', NULL, '01M0PAY13K1F11EGR7EET7DHAM', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0PAY1VFM713F2EEJTEAT2B0', jsonb_build_object('en-US', 'Delete', 'vi-VN', 'Xóa'), 'delete', NULL, '01M0PAY13K1F11EGR7EET7DHAM', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;
