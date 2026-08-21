-- Payment methods of the Payment & Invoice module.
--
-- These rows are what makes a method nameable at all: an order carries payment_method_id as a
-- required foreign key, so a deployment with an empty table can take no payment through any
-- gateway, whatever its credentials say.
--
-- "code" is the name every caller quotes and must stay byte-identical to the vending machine's
-- PaymentMethod enum, which is what the kiosks send as payment_type. "adapter_code" names the
-- implementation that carries the payment out and happens to match here; the two are separate
-- columns so that a second merchant account can later be offered under its own code while still
-- being served by the same adapter.
--
-- currency_id is filled even though the column is nullable. An order's currency is copied from
-- its method and the order's own column is NOT NULL, so a method that leaves it blank composes an
-- order its own schema refuses — which surfaces as a 500 rather than as anything about currency.
--
-- The amount bounds are the ones the standalone NestJS service held in configuration. The upper
-- bound is exclusive: assertAmountWithinMethodBounds refuses an amount greater than *or equal to*
-- max_amount, matching that service's `amount >= maximumAmount`.

DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'paymentinvoice_payment_methods'
	) THEN
		INSERT INTO "paymentinvoice_payment_methods" (
			"id", "code", "adapter_code", "name", "description", "currency_id",
			"min_amount", "max_amount", "is_active", "config", "is_archived",
			"created_at", "updated_at", "etag"
		) VALUES
		('01M0PAY1MTHD00000000000MM0', 'momo', 'momo', jsonb_build_object('en-US', 'MoMo', 'vi-VN', 'MoMo'), NULL, '01KZQC0000CURRENCY000ND300', 1000, 50000000, TRUE, NULL, FALSE, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0PAY1MTHD00000000000VQ0', 'vietqr', 'vietqr', jsonb_build_object('en-US', 'VietQR', 'vi-VN', 'VietQR'), NULL, '01KZQC0000CURRENCY000ND300', 2000, 20000000, TRUE, NULL, FALSE, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0PAY1MTHD00000000000MP0', 'mpos', 'mpos', jsonb_build_object('en-US', 'mPOS', 'vi-VN', 'mPOS'), NULL, '01KZQC0000CURRENCY000ND300', 3000, 20000000, TRUE, NULL, FALSE, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;
