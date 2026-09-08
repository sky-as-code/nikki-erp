-- The three kiosk fulfillment methods, and the vending channel's permission to use them.
--
-- These are configuration, not code: the codes are what the vending flows resolve a policy by, but
-- an operator may add methods of their own beside them, and may edit these. Ids are fixed and every
-- insert is ON CONFLICT DO NOTHING, so re-running the migration is a no-op.
--
-- The three differ in exactly two decisions, and the pair is the whole design:
--
--   AUTO_REFUND       one attempt, then refund what did not come out. For a walk-up customer
--                     nobody can identify — there is no one to ask what they would prefer, and no
--                     way to reach them later, so the money goes back immediately.
--   AUTHENTICATED     unlimited attempts, refund nothing automatically. The customer is known, so
--                     the entitlement is kept: they may try another machine or ask for the money.
--                     Refunding on their behalf would take away goods they may still want.
--   PICKUP_SELECTED   the same, but the target cannot be inferred from where the order was raised,
--                     because it was not raised at a kiosk at all.
--
-- Only the vending channel is mapped to them. Default-deny is the rule for these mappings, so the
-- point-of-sale and back-office channels permit no kiosk method until somebody says otherwise —
-- which is correct, since neither has a machine to dispense from.
--
-- Neither the channel nor any point is given a default here. A default is an operational choice
-- about which policy a given kiosk sells under, and guessing it in a migration would silently pick
-- the refund behaviour for every existing machine.

DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'sales_fulfillment_methods'
	) THEN
		INSERT INTO "sales_fulfillment_methods" (
			"id", "org_id", "code", "name", "description", "fulfillment_type", "initial_target_selection", "max_attempts", "failure_action", "allow_target_change", "allow_partial_fulfillment", "requires_authenticated_customer", "reservation_ttl_minutes", "is_archived", "created_at", "updated_at", "etag"
		) VALUES
		('01M3SFM0000000000000000001', '01JWNY20G23KD4RV5VWYABQYHD', 'KIOSK_DIRECT_AUTO_REFUND', 'Kiosk Direct - Auto Refund', 'A walk-up sale at the machine itself. One dispense attempt; whatever fails to come out is refunded without asking, because there is no identified customer to ask.', 'kiosk_dispense', 'current_sales_outlet', 1, 'auto_refund', false, true, false, NULL, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SFM0000000000000000002', '01JWNY20G23KD4RV5VWYABQYHD', 'KIOSK_DIRECT_AUTHENTICATED', 'Kiosk Direct - Customer Controlled', 'A sale at the machine by an identified customer. Nothing is refunded automatically: the entitlement is kept so the customer may try another machine or ask for their money back.', 'kiosk_dispense', 'current_sales_outlet', NULL, 'customer_action_required', true, true, true, 1440, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SFM0000000000000000003', '01JWNY20G23KD4RV5VWYABQYHD', 'KIOSK_PICKUP_SELECTED', 'Kiosk Pickup - Customer Selected', 'An order raised away from the machine and collected at a kiosk the customer chose. The target cannot be inferred from where the order was created, so one must be named before it is confirmed.', 'kiosk_dispense', 'customer_selected_outlet', NULL, 'customer_action_required', true, true, true, 1440, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'sales_channel_fulfillment_methods'
	) THEN
		INSERT INTO "sales_channel_fulfillment_methods" (
			"id", "org_id", "sales_channel_id", "fulfillment_method_id", "created_at", "updated_at", "etag"
		) VALUES
		('01M3SCF0000000000000000001', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000001', '01M3SFM0000000000000000001', NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SCF0000000000000000002', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000001', '01M3SFM0000000000000000002', NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SCF0000000000000000003', '01JWNY20G23KD4RV5VWYABQYHD', '01M3SC00000000000000000001', '01M3SFM0000000000000000003', NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;
