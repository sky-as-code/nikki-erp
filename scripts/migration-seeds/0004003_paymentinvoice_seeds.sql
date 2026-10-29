-- Seed data for the Payment Invoice module: the payment methods a deployment can be paid through.
--
-- These four mirror the methods coremart/scripts/migrations/2001003_vendingmachine_seed.sql
-- configures on its kiosks. That file writes them into vdmc_new_payments, which is the vending
-- module's own copy; this file writes the master records they are a copy OF, so that Sales and
-- anything else can name a payment method without reaching into the vending module's table.
--
-- Three of the four have an adapter in the gateway registry. "MBBANK" deliberately does not, and
-- is seeded active anyway: usability is not a property of the row, it also depends on which
-- adapters the running build ships, and a deployment with no mbbank adapter must report the method
-- as present-but-unusable rather than hide it. Seeding one such method is what makes that path
-- exercisable without editing data.
--
-- max_amount is EXCLUSIVE (paymentinvoice/domain/services/order_domservice.go:341) - an amount
-- equal to it is refused. The bounds below are therefore one VND above the intended ceiling.
--
-- Ids are fixed rather than generated and every insert is ON CONFLICT DO NOTHING, so running this
-- file twice changes nothing.


DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'paymentinvoice_payment_methods'
	) THEN
		INSERT INTO "paymentinvoice_payment_methods" (
			"id", "code", "adapter_code", "name", "description", "currency_id", "min_amount", "max_amount", "is_active", "config", "is_archived", "created_at", "updated_at", "etag"
		) VALUES
		('01M3PM00000000000000000001', 'MOMO', 'momo', '{"en-US": "MoMo Wallet", "vi-VN": "Vi MoMo"}'::jsonb, 'E-wallet payment, scanned from the customer phone', NULL, 1000, 20000001, true, NULL, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3PM00000000000000000002', 'VIETQR', 'vietqr', '{"en-US": "VietQR Bank Transfer", "vi-VN": "Chuyen khoan VietQR"}'::jsonb, 'Bank transfer against a generated QR code', NULL, 1000, 500000001, true, NULL, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3PM00000000000000000003', 'MPOS', 'mpos', '{"en-US": "mPOS Card Terminal", "vi-VN": "May POS the"}'::jsonb, 'Card payment prompted on a physical terminal', NULL, 10000, 50000001, true, NULL, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3PM00000000000000000004', 'MBBANK', 'mbbank', '{"en-US": "MB Bank Transfer", "vi-VN": "Chuyen khoan MB Bank"}'::jsonb, 'Bank transfer through MB Bank. Seeded active on purpose although no mbbank adapter ships: it is the deployment-dependent usability gate made visible, and the merged view must report it unusable rather than absent', NULL, 1000, 500000001, true, NULL, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;
