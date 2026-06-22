DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_organizations'
	) THEN
		INSERT INTO "iam_organizations" ("id", "address", "display_name", "legal_name", "phone_number", "slug", "etag", "is_archived", "created_at", "updated_at") VALUES
		('01JWNY20G23KD4RV5VWYABQYHD', NULL, 'My Company', NULL, NULL, 'my-company', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, FALSE, NOW(), NULL),
		('01K02G6J1CYAN9K8V4PAGSQ5Z8', NULL, 'Old Company', NULL, NULL, 'old-company', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, FALSE, NOW(), NULL),
		('01K1H7M2K9VW3P5R7XQJY2C1Z9', NULL, 'Tech Solutions Ltd', NULL, NULL, 'tech-solutions', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, FALSE, NOW(), NULL);
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_org_units'
	) THEN
		INSERT INTO "iam_org_units" ("id", "name", "description", "path", "parent_id", "org_id", "etag", "created_at", "updated_at") VALUES
		('01K1H8N3L0WX4Q6S8YRKT3D2A2', 'Executive Office', NULL, ARRAY['01JWNY20G23KD4RV5VWYABQYHD', '01K1H8N3L0WX4Q6S8YRKT3D2A2']::varchar[], NULL, '01JWNY20G23KD4RV5VWYABQYHD', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL),
		('01K1H8N3L0WX4Q6S8YRKT3D2A3', 'Operations Division', NULL, ARRAY['01JWNY20G23KD4RV5VWYABQYHD', '01K1H8N3L0WX4Q6S8YRKT3D2A2', '01K1H8N3L0WX4Q6S8YRKT3D2A3']::varchar[], '01K1H8N3L0WX4Q6S8YRKT3D2A2', '01JWNY20G23KD4RV5VWYABQYHD', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL),
		('01K1H8N3L0WX4Q6S8YRKT3D2A4', 'Product Engineering Department', NULL, ARRAY['01JWNY20G23KD4RV5VWYABQYHD', '01K1H8N3L0WX4Q6S8YRKT3D2A2', '01K1H8N3L0WX4Q6S8YRKT3D2A3', '01K1H8N3L0WX4Q6S8YRKT3D2A4']::varchar[], '01K1H8N3L0WX4Q6S8YRKT3D2A3', '01JWNY20G23KD4RV5VWYABQYHD', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL),
		('01K1H8N3L0WX4Q6S8YRKT3D2A5', 'Engineering Division', NULL, ARRAY['01K1H7M2K9VW3P5R7XQJY2C1Z9', '01K1H8N3L0WX4Q6S8YRKT3D2A5']::varchar[], NULL, '01K1H7M2K9VW3P5R7XQJY2C1Z9', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL),
		('01K1H8N3L0WX4Q6S8YRKT3D2A6', 'Platform Engineering Department', NULL, ARRAY['01K1H7M2K9VW3P5R7XQJY2C1Z9', '01K1H8N3L0WX4Q6S8YRKT3D2A5', '01K1H8N3L0WX4Q6S8YRKT3D2A6']::varchar[], '01K1H8N3L0WX4Q6S8YRKT3D2A5', '01K1H7M2K9VW3P5R7XQJY2C1Z9', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL),
		('01K1H8N3L0WX4Q6S8YRKT3D2C0', 'Sales and Marketing Division', NULL, ARRAY['01K02G6J1CYAN9K8V4PAGSQ5Z8', '01K1H8N3L0WX4Q6S8YRKT3D2C0']::varchar[], NULL, '01K02G6J1CYAN9K8V4PAGSQ5Z8', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL),
		('01K1H8N3L0WX4Q6S8YRKT3D2C1', 'Sales Department', NULL, ARRAY['01K02G6J1CYAN9K8V4PAGSQ5Z8', '01K1H8N3L0WX4Q6S8YRKT3D2C0', '01K1H8N3L0WX4Q6S8YRKT3D2C1']::varchar[], '01K1H8N3L0WX4Q6S8YRKT3D2C0', '01K02G6J1CYAN9K8V4PAGSQ5Z8', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL),
		('01K1H8N3L0WX4Q6S8YRKT3D2C2', 'Marketing Department', NULL, ARRAY['01K02G6J1CYAN9K8V4PAGSQ5Z8', '01K1H8N3L0WX4Q6S8YRKT3D2C0', '01K1H8N3L0WX4Q6S8YRKT3D2C2']::varchar[], '01K1H8N3L0WX4Q6S8YRKT3D2C0', '01K02G6J1CYAN9K8V4PAGSQ5Z8', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL);
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_users'
	) THEN
		INSERT INTO "iam_users" (
			"id", "avatar_url", "display_name", "email", "status", "is_owner", "org_unit_id",
			"password", "password_expires_at", "password_updated_at", "passwordtmp", "passwordtmp_expires_at",
			"passwordotp", "passwordotp_expires_at", "passwordotp_recovery",
			"is_archived", "etag", "created_at", "updated_at"
		) VALUES
		-- System (system@nikki.com) | password: Passwo0rd123 | temp: Passwo0rd123 | otp: JBSWY3DPEHPK3PXP
		(
			'01JWNNJGS70Y07MBEV3AQ0M526', NULL, 'System', 'system@nikki.com', 'active', NULL, NULL,
			'$argon2id$19$65536$3$2$/hr9fzDjiJjphmxADvaRRg$hWu5KMrK7d1W7bdPU6K9Gb12W2dLfsh7k6MGYwqFjEw',
			NOW() + INTERVAL '365 days', NOW() - INTERVAL '10 days',
			'$argon2id$19$65536$3$2$/hr9fzDjiJjphmxADvaRRg$hWu5KMrK7d1W7bdPU6K9Gb12W2dLfsh7k6MGYwqFjEw',
			NOW() + INTERVAL '365 days', 'JBSWY3DPEHPK3PXP', NOW() + INTERVAL '365 days',
			'{"A2BC-3DEF-4GHI-5JKL", "M2NP-3QRS-4TUV-5WXY", "Z2A3-4BCD-5EFG-6HJK", "L2MN-3PQR-4STU-5VWX", "Y2Z3-4ABC-5DEF-6GHJ", "K2LM-3NOP-4QRS-5TUV", "W2XY-3ZAB-4CDE-5FGH", "J2KL-3MNP-4QRS-5TUV", "V2WX-3YZA-4BCD-5EFG", "H2JK-3LMN-4PQR-5STU"}'::character varying[],
			FALSE, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL
		),
		-- Admin Owner (owner@nikki.com) | password: Passwo0rd123 | temp: Passwo0rd123 | otp: KBSWY3DPEHPK3PXQ
		(
			'01JWNMZ36QHC7CQQ748H9NQ6J6', NULL, 'Admin Owner', 'owner@nikki.com', 'active', TRUE, NULL,
			'$argon2id$19$65536$3$2$/hr9fzDjiJjphmxADvaRRg$hWu5KMrK7d1W7bdPU6K9Gb12W2dLfsh7k6MGYwqFjEw',
			NOW() + INTERVAL '365 days', NOW() - INTERVAL '20 days',
			'$argon2id$19$65536$3$2$/hr9fzDjiJjphmxADvaRRg$hWu5KMrK7d1W7bdPU6K9Gb12W2dLfsh7k6MGYwqFjEw',
			NOW() + INTERVAL '365 days', 'KBSWY3DPEHPK3PXQ', NOW() + INTERVAL '365 days',
			'{"C2DE-3FGH-4JKL-5MNP", "Q2RS-3TUV-4WXY-5ZAB", "D2EF-3GHJ-4KLM-5NPQ", "R2ST-3UVW-4XYZ-5ABC", "E2FG-3HJK-4LMN-5PQR", "S2TU-3VWX-4YZA-5BCD", "F2GH-3JKL-4MNP-5QRS", "T2UV-3WXY-4ZAB-5CDE", "G2HJ-3KLM-4NPQ-5RST", "U2VW-3XYZ-4ABC-5DEF"}'::character varying[],
			FALSE, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL
		),
		-- Nguyễn Văn An (nguyen.van.an@nikki.com) | password: Passwo0rd123 | temp: Passwo0rd123 | otp: LBSWY3DPEHPK3PXR
		(
			'01JWNXT3EY7FG47VDJTEPTDC98', NULL, 'Nguyễn Văn An Domain Admin', 'nguyen.van.an@nikki.com', 'active', NULL, '01K1H8N3L0WX4Q6S8YRKT3D2A2',
			'$argon2id$19$65536$3$2$/hr9fzDjiJjphmxADvaRRg$hWu5KMrK7d1W7bdPU6K9Gb12W2dLfsh7k6MGYwqFjEw',
			NOW() + INTERVAL '180 days', NOW() - INTERVAL '30 days',
			'$argon2id$19$65536$3$2$/hr9fzDjiJjphmxADvaRRg$hWu5KMrK7d1W7bdPU6K9Gb12W2dLfsh7k6MGYwqFjEw',
			NOW() + INTERVAL '365 days', 'LBSWY3DPEHPK3PXR', NOW() + INTERVAL '180 days',
			'{"H2JK-3LMN-4PQR-5STU", "V2WX-3YZA-4BCD-5EFG", "I2KL-3MNP-4QRS-5TUV", "W2XY-3ZAB-4CDE-5FGH", "J2LM-3NPQ-4RST-5UVW", "X2YZ-3ABC-4DEF-5GHJ", "K2MN-3PQR-4STU-5VWX", "Y2ZA-3BCD-4EFG-5HJK", "L2NP-3QRS-4TUV-5WXY", "Z2AB-3CDE-4FGH-5JKL"}'::character varying[],
			FALSE, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL
		),
		-- Trần Thị Bình (tran.thi.binh@nikki.com) | password: Passwo0rd123 | temp: Passwo0rd123 | otp: MBSWY3DPEHPK3PXS
		(
			'01JWNXXTF8958VVYAV33MVVMDN', NULL, 'Trần Thị Bình Ident Readonly', 'tran.thi.binh@nikki.com', 'active', NULL, '01K1H8N3L0WX4Q6S8YRKT3D2A3',
			'$argon2id$19$65536$3$2$/hr9fzDjiJjphmxADvaRRg$hWu5KMrK7d1W7bdPU6K9Gb12W2dLfsh7k6MGYwqFjEw',
			NOW() + INTERVAL '180 days', NOW() - INTERVAL '45 days',
			'$argon2id$19$65536$3$2$/hr9fzDjiJjphmxADvaRRg$hWu5KMrK7d1W7bdPU6K9Gb12W2dLfsh7k6MGYwqFjEw',
			NOW() + INTERVAL '365 days', 'MBSWY3DPEHPK3PXS', NOW() + INTERVAL '180 days',
			'{"M2NP-3QRS-4TUV-5WXY", "A2BC-3DEF-4GHI-5JKL", "N2PQ-3RST-4UVW-5XYZ", "B2CD-3EFG-4HIJ-5KLM", "P2QR-3STU-4VWX-5YZA", "C2DE-3FGH-4JKL-5MNP", "Q2RS-3TUV-4WXY-5ZAB", "D2EF-3GHJ-4KLM-5NPQ", "R2ST-3UVW-4XYZ-5ABC", "E2FG-3HJK-4LMN-5PQR"}'::character varying[],
			FALSE, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL
		),
		-- Lê Văn Cường (le.van.cuong@nikki.com) | password: Passwo0rd123 | temp: Passwo0rd123 | otp: NBSWY3DPEHPK3PXT
		(
			'01JZQFDH0N51Q3BFQFMFFGSCSV', NULL, 'Lê Văn Cường Authz Admin', 'le.van.cuong@nikki.com', 'active', NULL, '01K1H8N3L0WX4Q6S8YRKT3D2A4',
			'$argon2id$19$65536$3$2$/hr9fzDjiJjphmxADvaRRg$hWu5KMrK7d1W7bdPU6K9Gb12W2dLfsh7k6MGYwqFjEw',
			NOW() + INTERVAL '365 days', NOW() - INTERVAL '12 days',
			'$argon2id$19$65536$3$2$/hr9fzDjiJjphmxADvaRRg$hWu5KMrK7d1W7bdPU6K9Gb12W2dLfsh7k6MGYwqFjEw',
			NOW() + INTERVAL '365 days', 'NBSWY3DPEHPK3PXT', NOW() + INTERVAL '365 days',
			'{"F2GH-3JKL-4MNP-5QRS", "T2UV-3WXY-4ZAB-5CDE", "G2HJ-3KLM-4NPQ-5RST", "U2VW-3XYZ-4ABC-5DEF", "H2JK-3LMN-4PQR-5STU", "V2WX-3YZA-4BCD-5EFG", "I2KL-3MNP-4QRS-5TUV", "W2XY-3ZAB-4CDE-5FGH", "J2LM-3NPQ-4RST-5UVW", "X2YZ-3ABC-4DEF-5GHJ"}'::character varying[],
			FALSE, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL
		),
		-- Phạm Thị Dung (pham.thi.dung@nikki.com) | password: Passwo0rd123 | temp: Passwo0rd123 | otp: OBSWY3DPEHPK3PXU
		(
			'01JZQFF9QEXH71P2CG9Y9MY8MM', NULL, 'Phạm Thị Dung Authz Mod', 'pham.thi.dung@nikki.com', 'locked', NULL, '01K1H8N3L0WX4Q6S8YRKT3D2A5',
			'$argon2id$19$65536$3$2$/hr9fzDjiJjphmxADvaRRg$hWu5KMrK7d1W7bdPU6K9Gb12W2dLfsh7k6MGYwqFjEw',
			NOW() + INTERVAL '180 days', NOW() - INTERVAL '90 days',
			'$argon2id$19$65536$3$2$/hr9fzDjiJjphmxADvaRRg$hWu5KMrK7d1W7bdPU6K9Gb12W2dLfsh7k6MGYwqFjEw',
			NOW() + INTERVAL '365 days', 'OBSWY3DPEHPK3PXU', NOW() + INTERVAL '180 days',
			'{"K2MN-3PQR-4STU-5VWX", "Y2ZA-3BCD-4EFG-5HJK", "L2NP-3QRS-4TUV-5WXY", "Z2AB-3CDE-4FGH-5JKL", "M2PQ-3RST-4UVW-5XYZ", "A2BC-3DEF-4GHI-5JKL", "N2QR-3STU-4VWX-5YZA", "B2CD-3EFG-4HIJ-5KLM", "P2RS-3TUV-4WXY-5ZAB", "C2DE-3FGH-4JKL-5MNP"}'::character varying[],
			FALSE, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL
		),
		-- Hoàng Văn Em (hoang.van.em@nikki.com) | password: Passwo0rd123 | temp: Passwo0rd123 | otp: PBSWY3DPEHPK3PXV
		(
			'01JZQFFDKY8T4JB8R6NSY1331J', NULL, 'Hoàng Văn Em My Company Org Admin', 'hoang.van.em@nikki.com', 'active', NULL, '01K1H8N3L0WX4Q6S8YRKT3D2A6',
			'$argon2id$19$65536$3$2$/hr9fzDjiJjphmxADvaRRg$hWu5KMrK7d1W7bdPU6K9Gb12W2dLfsh7k6MGYwqFjEw',
			NOW() + INTERVAL '365 days', NOW() - INTERVAL '5 days',
			'$argon2id$19$65536$3$2$/hr9fzDjiJjphmxADvaRRg$hWu5KMrK7d1W7bdPU6K9Gb12W2dLfsh7k6MGYwqFjEw',
			NOW() + INTERVAL '365 days', 'PBSWY3DPEHPK3PXV', NOW() + INTERVAL '365 days',
			'{"D2EF-3GHJ-4KLM-5NPQ", "R2ST-3UVW-4XYZ-5ABC", "E2FG-3HJK-4LMN-5PQR", "S2TU-3VWX-4YZA-5BCD", "F2GH-3JKL-4MNP-5QRS", "T2UV-3WXY-4ZAB-5CDE", "G2HJ-3KLM-4NPQ-5RST", "U2VW-3XYZ-4ABC-5DEF", "H2JK-3LMN-4PQR-5STU", "V2WX-3YZA-4BCD-5EFG"}'::character varying[],
			FALSE, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL
		),
		-- Đặng Thị Phương (dang.thi.phuong@nikki.com) | password: Passwo0rd123 | temp: Passwo0rd123 | otp: QBSWY3DPEHPK3PXW
		(
			'01JZQFGVKZCTV7S310W0BDMWCS', NULL, 'Đặng Thị Phương Membership Company Org Admin', 'dang.thi.phuong@nikki.com', 'active', NULL, '01K1H8N3L0WX4Q6S8YRKT3D2C0',
			'$argon2id$19$65536$3$2$/hr9fzDjiJjphmxADvaRRg$hWu5KMrK7d1W7bdPU6K9Gb12W2dLfsh7k6MGYwqFjEw',
			NOW() + INTERVAL '365 days', NOW() - INTERVAL '7 days',
			'$argon2id$19$65536$3$2$/hr9fzDjiJjphmxADvaRRg$hWu5KMrK7d1W7bdPU6K9Gb12W2dLfsh7k6MGYwqFjEw',
			NOW() + INTERVAL '365 days', 'QBSWY3DPEHPK3PXW', NOW() + INTERVAL '365 days',
			'{"I2KL-3MNP-4QRS-5TUV", "W2XY-3ZAB-4CDE-5FGH", "J2LM-3NPQ-4RST-5UVW", "X2YZ-3ABC-4DEF-5GHJ", "K2MN-3PQR-4STU-5VWX", "Y2ZA-3BCD-4EFG-5HJK", "L2NP-3QRS-4TUV-5WXY", "Z2AB-3CDE-4FGH-5JKL", "M2PQ-3RST-4UVW-5XYZ", "A2BC-3DEF-4GHI-5JKL"}'::character varying[],
			FALSE, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL
		),
		-- Võ Văn Giang (vo.van.giang@nikki.com) | password: Passwo0rd123 | temp: Passwo0rd123 | otp: RBSWY3DPEHPK3PXX
		(
			'01JZQFY6EXRG0959Z95Y2EM3AM', NULL, 'Võ Văn Giang', 'vo.van.giang@nikki.com', 'archived', NULL, '01K1H8N3L0WX4Q6S8YRKT3D2C1',
			'$argon2id$19$65536$3$2$/hr9fzDjiJjphmxADvaRRg$hWu5KMrK7d1W7bdPU6K9Gb12W2dLfsh7k6MGYwqFjEw',
			NOW() + INTERVAL '180 days', NOW() - INTERVAL '45 days',
			'$argon2id$19$65536$3$2$/hr9fzDjiJjphmxADvaRRg$hWu5KMrK7d1W7bdPU6K9Gb12W2dLfsh7k6MGYwqFjEw',
			NOW() + INTERVAL '365 days', 'RBSWY3DPEHPK3PXX', NOW() + INTERVAL '180 days',
			'{"N2QR-3STU-4VWX-5YZA", "B2CD-3EFG-4HIJ-5KLM", "P2RS-3TUV-4WXY-5ZAB", "C2DE-3FGH-4JKL-5MNP", "Q2ST-3UVW-4XYZ-5ABC", "D2EF-3GHJ-4KLM-5NPQ", "R2TU-3VWX-4YZA-5BCD", "E2FG-3HJK-4LMN-5PQR", "S2UV-3WXY-4ZAB-5CDE", "F2GH-3JKL-4MNP-5QRS"}'::character varying[],
			FALSE, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL
		),
		-- Bùi Thị Hoa (bui.thi.hoa@nikki.com) | password: Passwo0rd123 | temp: Passwo0rd123 | otp: SBSWY3DPEHPK3PXY
		(
			'01JZQFZFK6GM2D5X6MYHWH6FND', NULL, 'Bùi Thị Hoa', 'bui.thi.hoa@nikki.com', 'active', NULL, '01K1H8N3L0WX4Q6S8YRKT3D2C2',
			'$argon2id$19$65536$3$2$/hr9fzDjiJjphmxADvaRRg$hWu5KMrK7d1W7bdPU6K9Gb12W2dLfsh7k6MGYwqFjEw',
			NOW() + INTERVAL '365 days', NOW() - INTERVAL '3 days',
			'$argon2id$19$65536$3$2$/hr9fzDjiJjphmxADvaRRg$hWu5KMrK7d1W7bdPU6K9Gb12W2dLfsh7k6MGYwqFjEw',
			NOW() + INTERVAL '365 days', 'SBSWY3DPEHPK3PXY', NOW() + INTERVAL '365 days',
			'{"G2HJ-3KLM-4NPQ-5RST", "U2VW-3XYZ-4ABC-5DEF", "H2JK-3LMN-4PQR-5STU", "V2WX-3YZA-4BCD-5EFG", "I2KL-3MNP-4QRS-5TUV", "W2XY-3ZAB-4CDE-5FGH", "J2LM-3NPQ-4RST-5UVW", "X2YZ-3ABC-4DEF-5GHJ", "K2MN-3PQR-4STU-5VWX", "Y2ZA-3BCD-4EFG-5HJK"}'::character varying[],
			FALSE, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL
		);
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_groups'
	) THEN
		INSERT INTO "iam_groups" ("id", "name", "description", "owner_id", "is_archived", "etag", "created_at", "updated_at") VALUES
		('01JWNXBR5QJBH7PE9PQ9FW746V', jsonb_build_object('en-US', 'Domain Users', 'vi-VN', 'Người dùng domain', 'zh-TW', '網域使用者'), jsonb_build_object('en-US', 'Default group for all domain users', 'vi-VN', 'Nhóm mặc định cho tất cả người dùng domain', 'zh-TW', '所有網域使用者的預設群組'), '01JWNMZ36QHC7CQQ748H9NQ6J6', FALSE, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL),
		('01K1H8N3L0WX4Q6S8YRKT3D2A0', jsonb_build_object('en-US', 'Administrators', 'vi-VN', 'Quản trị viên', 'zh-TW', '管理員'), jsonb_build_object('en-US', 'System administrators group', 'vi-VN', 'Nhóm quản trị viên hệ thống', 'zh-TW', '系統管理員群組'), '01JWNMZ36QHC7CQQ748H9NQ6J6', FALSE, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL),
		('01K1H8N3L0WX4Q6S8YRKT3D2A1', jsonb_build_object('en-US', 'Project Managers', 'vi-VN', 'Quản lý dự án', 'zh-TW', '專案經理'), jsonb_build_object('en-US', 'Project management team', 'vi-VN', 'Đội quản lý dự án', 'zh-TW', '專案管理團隊'), '01JWNMZ36QHC7CQQ748H9NQ6J6', FALSE, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL),
		('01K1H8N3L0WX4Q6S8YRKT3D2B0', jsonb_build_object('en-US', 'Legacy Support', 'vi-VN', 'Hỗ trợ hệ thống cũ', 'zh-TW', '舊版支援'), jsonb_build_object('en-US', 'Legacy system support team', 'vi-VN', 'Đội hỗ trợ hệ thống kế thừa', 'zh-TW', '舊系統支援團隊'), '01JWNMZ36QHC7CQQ748H9NQ6J6', FALSE, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL),
		('01K1H8N3L0WX4Q6S8YRKT3D2B1', jsonb_build_object('en-US', 'Archives Team', 'vi-VN', 'Đội lưu trữ', 'zh-TW', '檔案團隊'), jsonb_build_object('en-US', 'Archived content management', 'vi-VN', 'Quản lý nội dung đã lưu trữ', 'zh-TW', '封存內容管理'), '01JWNMZ36QHC7CQQ748H9NQ6J6', FALSE, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL),
		('01K1H8N3L0WX4Q6S8YRKT3D2B2', jsonb_build_object('en-US', 'Maintenance Group', 'vi-VN', 'Nhóm bảo trì', 'zh-TW', '維護群組'), jsonb_build_object('en-US', 'System maintenance personnel', 'vi-VN', 'Nhân sự bảo trì hệ thống', 'zh-TW', '系統維護人員'), '01JWNMZ36QHC7CQQ748H9NQ6J6', FALSE, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL);
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_group_user_rel'
	) THEN
		INSERT INTO "iam_group_user_rel" ("id", "group_id", "user_id") VALUES
		('01KNHHJY18Q5EX6008FKEWMK8K', '01JWNXBR5QJBH7PE9PQ9FW746V', '01JWNXT3EY7FG47VDJTEPTDC98'),
		('01KNHHJY18DVCHSQNNAP800208', '01JWNXBR5QJBH7PE9PQ9FW746V', '01JWNXXTF8958VVYAV33MVVMDN'),
		('01KNHHJY19W0Q0ZXPBPKCTPRYP', '01K1H8N3L0WX4Q6S8YRKT3D2A0', '01JZQFDH0N51Q3BFQFMFFGSCSV'),
		('01KNHHJY191TZPXRX3FH77NG3C', '01K1H8N3L0WX4Q6S8YRKT3D2A1', '01JZQFF9QEXH71P2CG9Y9MY8MM'),
		('01KNHHJY19T3DAJJ6NR5CNHS8N', '01K1H8N3L0WX4Q6S8YRKT3D2A0', '01JZQFFDKY8T4JB8R6NSY1331J'),
		('01KNHHJY19BQXEP69G7WKBBN20', '01K1H8N3L0WX4Q6S8YRKT3D2B0', '01JZQFGVKZCTV7S310W0BDMWCS'),
		('01KNHHJY195N1T928PW599MRA2', '01K1H8N3L0WX4Q6S8YRKT3D2B1', '01JZQFY6EXRG0959Z95Y2EM3AM'),
		('01KNHHJY1900QTVWVK05SS1CMZ', '01K1H8N3L0WX4Q6S8YRKT3D2B2', '01JZQFZFK6GM2D5X6MYHWH6FND');
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_org_user_rel'
	) THEN
		INSERT INTO "iam_org_user_rel" ("id", "org_id", "user_id") VALUES
		('01KNHHJY19SYS1USRBASEORG01', '01JWNY20G23KD4RV5VWYABQYHD', '01JWNNJGS70Y07MBEV3AQ0M526'),
		('01KNHHJY19GHJEG95NBH6YFD0Z', '01JWNY20G23KD4RV5VWYABQYHD', '01JWNMZ36QHC7CQQ748H9NQ6J6'),
		('01KNHHJY19A3A7PNRQ5ENDC83Y', '01K02G6J1CYAN9K8V4PAGSQ5Z8', '01JWNMZ36QHC7CQQ748H9NQ6J6'),
		('01KNHHJY19CMA8GJH2TRPKS4YY', '01K1H7M2K9VW3P5R7XQJY2C1Z9', '01JWNMZ36QHC7CQQ748H9NQ6J6'),
		('01KNHHJY193QK0T2S3CA7YVHKQ', '01JWNY20G23KD4RV5VWYABQYHD', '01JWNXT3EY7FG47VDJTEPTDC98'),
		('01KNHHJY19ZN374TPEKH3SAQXE', '01JWNY20G23KD4RV5VWYABQYHD', '01JWNXXTF8958VVYAV33MVVMDN'),
		('01KNHHJY19T5V138945CYYJQ36', '01K02G6J1CYAN9K8V4PAGSQ5Z8', '01JZQFGVKZCTV7S310W0BDMWCS'),
		('01KNHHJY19GY7Y20QWSNZGJ6XE', '01K02G6J1CYAN9K8V4PAGSQ5Z8', '01JZQFY6EXRG0959Z95Y2EM3AM'),
		('01KNHHJY1929TPTFYCTY35FF84', '01K02G6J1CYAN9K8V4PAGSQ5Z8', '01JZQFZFK6GM2D5X6MYHWH6FND'),
		('01KNHHJY19TWA5THE0C61SMXS4', '01JWNY20G23KD4RV5VWYABQYHD', '01JZQFDH0N51Q3BFQFMFFGSCSV'),
		('01KNHHJY19DUNGBASEORG00001', '01JWNY20G23KD4RV5VWYABQYHD', '01JZQFF9QEXH71P2CG9Y9MY8MM'),
		('01KNHHJY19W1TEBX2SV5VYGGSZ', '01K1H7M2K9VW3P5R7XQJY2C1Z9', '01JZQFF9QEXH71P2CG9Y9MY8MM'),
		('01KNHHJY19PN9BZM3RQR7SE68C', '01K1H7M2K9VW3P5R7XQJY2C1Z9', '01JZQFFDKY8T4JB8R6NSY1331J'),
		('01KNHHJY19ORGMC01JZQFFDKYT', '01JWNY20G23KD4RV5VWYABQYHD', '01JZQFFDKY8T4JB8R6NSY1331J'),
		('01KNHHJY19PHUONGBASEORG001', '01JWNY20G23KD4RV5VWYABQYHD', '01JZQFGVKZCTV7S310W0BDMWCS'),
		('01KNHHJY19GIANGBASEORG0001', '01JWNY20G23KD4RV5VWYABQYHD', '01JZQFY6EXRG0959Z95Y2EM3AM'),
		('01KNHHJY19HOABASEORG000001', '01JWNY20G23KD4RV5VWYABQYHD', '01JZQFZFK6GM2D5X6MYHWH6FND'),
		('01KNHHJY19ORGTS01JZQFGVKZC', '01K1H7M2K9VW3P5R7XQJY2C1Z9', '01JZQFGVKZCTV7S310W0BDMWCS');
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_method_settings'
	) THEN
		INSERT INTO "iam_method_settings" (
			"id",
			"method",
			"order",
			"max_failures",
			"lock_duration_secs",
			"subject_type",
			"subject_ref",
			"subject_source_ref"
		) VALUES
		-- Domain-level settings (bound to system subject for consistency)
		('01K0AUTH000000000000000101', 'password', 1, 5, 1800, 'domain', '01JWNMZ36QHC7CQQ748H9NQ6J6', 'system'),
		('01K0AUTH000000000000000102', 'captcha', 2, 3, 900,  'domain', '01JWNMZ36QHC7CQQ748H9NQ6J6', 'system'),
		('01K0AUTH000000000000000103', 'otpCode',  3, 3, 1800, 'domain', '01JWNMZ36QHC7CQQ748H9NQ6J6', 'system');
	END IF;
END $$;
