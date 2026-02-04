DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables 
		WHERE table_schema = 'public' AND table_name = 'authn_password_stores'
	) THEN
		INSERT INTO "authn_password_stores" (
			"id",
			"password",
			"password_expired_at",
			"password_updated_at",
			"passwordtmp",
			"passwordtmp_expired_at",
			"passwordotp",
			"passwordotp_expired_at",
			"passwordotp_recovery",
			"subject_type",
			"subject_ref",
			"subject_source_ref"
		) VALUES
		-- System (system@nikki.com) | password: Passwo0rd123 | temp: Passwo0rd123 | otp: JBSWY3DPEHPK3PXP
		(
			'01K0AUTH000000000000000001',
			'$argon2id$19$65536$3$2$/hr9fzDjiJjphmxADvaRRg$hWu5KMrK7d1W7bdPU6K9Gb12W2dLfsh7k6MGYwqFjEw',
			NOW() + INTERVAL '365 days',
			NOW() - INTERVAL '10 days',
			'$argon2id$19$65536$3$2$/hr9fzDjiJjphmxADvaRRg$hWu5KMrK7d1W7bdPU6K9Gb12W2dLfsh7k6MGYwqFjEw',
			NOW() + INTERVAL '365 days',
			'JBSWY3DPEHPK3PXP',
			NOW() + INTERVAL '365 days',
			'["A2BC-3DEF-4GHI-5JKL", "M2NP-3QRS-4TUV-5WXY", "Z2A3-4BCD-5EFG-6HJK", "L2MN-3PQR-4STU-5VWX", "Y2Z3-4ABC-5DEF-6GHJ", "K2LM-3NOP-4QRS-5TUV", "W2XY-3ZAB-4CDE-5FGH", "J2KL-3MNP-4QRS-5TUV", "V2WX-3YZA-4BCD-5EFG", "H2JK-3LMN-4PQR-5STU"]'::jsonb,
			'user',
			'01JWNNJGS70Y07MBEV3AQ0M526',
			'system@nikki.com'
		),
		-- Admin Owner (owner@nikki.com) | password: Passwo0rd123 | temp: Passwo0rd123 | otp: KBSWY3DPEHPK3PXQ
		(
			'01K0AUTH000000000000000002',
			'$argon2id$19$65536$3$2$/hr9fzDjiJjphmxADvaRRg$hWu5KMrK7d1W7bdPU6K9Gb12W2dLfsh7k6MGYwqFjEw',
			NOW() + INTERVAL '365 days',
			NOW() - INTERVAL '20 days',
			'$argon2id$19$65536$3$2$/hr9fzDjiJjphmxADvaRRg$hWu5KMrK7d1W7bdPU6K9Gb12W2dLfsh7k6MGYwqFjEw',
			NOW() + INTERVAL '365 days',
			'KBSWY3DPEHPK3PXQ',
			NOW() + INTERVAL '365 days',
			'["C2DE-3FGH-4JKL-5MNP", "Q2RS-3TUV-4WXY-5ZAB", "D2EF-3GHJ-4KLM-5NPQ", "R2ST-3UVW-4XYZ-5ABC", "E2FG-3HJK-4LMN-5PQR", "S2TU-3VWX-4YZA-5BCD", "F2GH-3JKL-4MNP-5QRS", "T2UV-3WXY-4ZAB-5CDE", "G2HJ-3KLM-4NPQ-5RST", "U2VW-3XYZ-4ABC-5DEF"]'::jsonb,
			'user',
			'01JWNMZ36QHC7CQQ748H9NQ6J6',
			'owner@nikki.com'
		),
		-- Nguyễn Văn An (nguyen.van.an@nikki.com) | password: Passwo0rd123 | temp: Passwo0rd123 | otp: LBSWY3DPEHPK3PXR
		(
			'01K0AUTH000000000000000003',
			'$argon2id$19$65536$3$2$/hr9fzDjiJjphmxADvaRRg$hWu5KMrK7d1W7bdPU6K9Gb12W2dLfsh7k6MGYwqFjEw',
			NOW() + INTERVAL '180 days',
			NOW() - INTERVAL '30 days',
			'$argon2id$19$65536$3$2$/hr9fzDjiJjphmxADvaRRg$hWu5KMrK7d1W7bdPU6K9Gb12W2dLfsh7k6MGYwqFjEw',
			NOW() + INTERVAL '365 days',
			'LBSWY3DPEHPK3PXR',
			NOW() + INTERVAL '180 days',
			'["H2JK-3LMN-4PQR-5STU", "V2WX-3YZA-4BCD-5EFG", "I2KL-3MNP-4QRS-5TUV", "W2XY-3ZAB-4CDE-5FGH", "J2LM-3NPQ-4RST-5UVW", "X2YZ-3ABC-4DEF-5GHJ", "K2MN-3PQR-4STU-5VWX", "Y2ZA-3BCD-4EFG-5HJK", "L2NP-3QRS-4TUV-5WXY", "Z2AB-3CDE-4FGH-5JKL"]'::jsonb,
			'user',
			'01JWNXT3EY7FG47VDJTEPTDC98',
			'nguyen.van.an@nikki.com'
		),
		-- Trần Thị Bình (tran.thi.binh@nikki.com) | password: Passwo0rd123 | temp: Passwo0rd123 | otp: MBSWY3DPEHPK3PXS
		(
			'01K0AUTH000000000000000004',
			'$argon2id$19$65536$3$2$/hr9fzDjiJjphmxADvaRRg$hWu5KMrK7d1W7bdPU6K9Gb12W2dLfsh7k6MGYwqFjEw',
			NOW() + INTERVAL '180 days',
			NOW() - INTERVAL '45 days',
			'$argon2id$19$65536$3$2$/hr9fzDjiJjphmxADvaRRg$hWu5KMrK7d1W7bdPU6K9Gb12W2dLfsh7k6MGYwqFjEw',
			NOW() + INTERVAL '365 days',
			'MBSWY3DPEHPK3PXS',
			NOW() + INTERVAL '180 days',
			'["M2NP-3QRS-4TUV-5WXY", "A2BC-3DEF-4GHI-5JKL", "N2PQ-3RST-4UVW-5XYZ", "B2CD-3EFG-4HIJ-5KLM", "P2QR-3STU-4VWX-5YZA", "C2DE-3FGH-4JKL-5MNP", "Q2RS-3TUV-4WXY-5ZAB", "D2EF-3GHJ-4KLM-5NPQ", "R2ST-3UVW-4XYZ-5ABC", "E2FG-3HJK-4LMN-5PQR"]'::jsonb,
			'user',
			'01JWNXXTF8958VVYAV33MVVMDN',
			'tran.thi.binh@nikki.com'
		),
		-- Lê Văn Cường (le.van.cuong@nikki.com) | password: Passwo0rd123 | temp: Passwo0rd123 | otp: NBSWY3DPEHPK3PXT
		(
			'01K0AUTH000000000000000005',
			'$argon2id$19$65536$3$2$/hr9fzDjiJjphmxADvaRRg$hWu5KMrK7d1W7bdPU6K9Gb12W2dLfsh7k6MGYwqFjEw',
			NOW() + INTERVAL '365 days',
			NOW() - INTERVAL '12 days',
			'$argon2id$19$65536$3$2$/hr9fzDjiJjphmxADvaRRg$hWu5KMrK7d1W7bdPU6K9Gb12W2dLfsh7k6MGYwqFjEw',
			NOW() + INTERVAL '365 days',
			'NBSWY3DPEHPK3PXT',
			NOW() + INTERVAL '365 days',
			'["F2GH-3JKL-4MNP-5QRS", "T2UV-3WXY-4ZAB-5CDE", "G2HJ-3KLM-4NPQ-5RST", "U2VW-3XYZ-4ABC-5DEF", "H2JK-3LMN-4PQR-5STU", "V2WX-3YZA-4BCD-5EFG", "I2KL-3MNP-4QRS-5TUV", "W2XY-3ZAB-4CDE-5FGH", "J2LM-3NPQ-4RST-5UVW", "X2YZ-3ABC-4DEF-5GHJ"]'::jsonb,
			'user',
			'01JZQFDH0N51Q3BFQFMFFGSCSV',
			'le.van.cuong@nikki.com'
		),
		-- Phạm Thị Dung (pham.thi.dung@nikki.com) | password: Passwo0rd123 | temp: Passwo0rd123 | otp: OBSWY3DPEHPK3PXU
		(
			'01K0AUTH000000000000000006',
			'$argon2id$19$65536$3$2$/hr9fzDjiJjphmxADvaRRg$hWu5KMrK7d1W7bdPU6K9Gb12W2dLfsh7k6MGYwqFjEw',
			NOW() + INTERVAL '180 days',
			NOW() - INTERVAL '90 days',
			'$argon2id$19$65536$3$2$/hr9fzDjiJjphmxADvaRRg$hWu5KMrK7d1W7bdPU6K9Gb12W2dLfsh7k6MGYwqFjEw',
			NOW() + INTERVAL '365 days',
			'OBSWY3DPEHPK3PXU',
			NOW() + INTERVAL '180 days',
			'["K2MN-3PQR-4STU-5VWX", "Y2ZA-3BCD-4EFG-5HJK", "L2NP-3QRS-4TUV-5WXY", "Z2AB-3CDE-4FGH-5JKL", "M2PQ-3RST-4UVW-5XYZ", "A2BC-3DEF-4GHI-5JKL", "N2QR-3STU-4VWX-5YZA", "B2CD-3EFG-4HIJ-5KLM", "P2RS-3TUV-4WXY-5ZAB", "C2DE-3FGH-4JKL-5MNP"]'::jsonb,
			'user',
			'01JZQFF9QEXH71P2CG9Y9MY8MM',
			'pham.thi.dung@nikki.com'
		),
		-- Hoàng Văn Em (hoang.van.em@nikki.com) | password: Passwo0rd123 | temp: Passwo0rd123 | otp: PBSWY3DPEHPK3PXV
		(
			'01K0AUTH000000000000000007',
			'$argon2id$19$65536$3$2$/hr9fzDjiJjphmxADvaRRg$hWu5KMrK7d1W7bdPU6K9Gb12W2dLfsh7k6MGYwqFjEw',
			NOW() + INTERVAL '365 days',
			NOW() - INTERVAL '5 days',
			'$argon2id$19$65536$3$2$/hr9fzDjiJjphmxADvaRRg$hWu5KMrK7d1W7bdPU6K9Gb12W2dLfsh7k6MGYwqFjEw',
			NOW() + INTERVAL '365 days',
			'PBSWY3DPEHPK3PXV',
			NOW() + INTERVAL '365 days',
			'["D2EF-3GHJ-4KLM-5NPQ", "R2ST-3UVW-4XYZ-5ABC", "E2FG-3HJK-4LMN-5PQR", "S2TU-3VWX-4YZA-5BCD", "F2GH-3JKL-4MNP-5QRS", "T2UV-3WXY-4ZAB-5CDE", "G2HJ-3KLM-4NPQ-5RST", "U2VW-3XYZ-4ABC-5DEF", "H2JK-3LMN-4PQR-5STU", "V2WX-3YZA-4BCD-5EFG"]'::jsonb,
			'user',
			'01JZQFFDKY8T4JB8R6NSY1331J',
			'hoang.van.em@nikki.com'
		),
		-- Đặng Thị Phương (dang.thi.phuong@nikki.com) | password: Passwo0rd123 | temp: Passwo0rd123 | otp: QBSWY3DPEHPK3PXW
		(
			'01K0AUTH000000000000000008',
			'$argon2id$19$65536$3$2$/hr9fzDjiJjphmxADvaRRg$hWu5KMrK7d1W7bdPU6K9Gb12W2dLfsh7k6MGYwqFjEw',
			NOW() + INTERVAL '365 days',
			NOW() - INTERVAL '7 days',
			'$argon2id$19$65536$3$2$/hr9fzDjiJjphmxADvaRRg$hWu5KMrK7d1W7bdPU6K9Gb12W2dLfsh7k6MGYwqFjEw',
			NOW() + INTERVAL '365 days',
			'QBSWY3DPEHPK3PXW',
			NOW() + INTERVAL '365 days',
			'["I2KL-3MNP-4QRS-5TUV", "W2XY-3ZAB-4CDE-5FGH", "J2LM-3NPQ-4RST-5UVW", "X2YZ-3ABC-4DEF-5GHJ", "K2MN-3PQR-4STU-5VWX", "Y2ZA-3BCD-4EFG-5HJK", "L2NP-3QRS-4TUV-5WXY", "Z2AB-3CDE-4FGH-5JKL", "M2PQ-3RST-4UVW-5XYZ", "A2BC-3DEF-4GHI-5JKL"]'::jsonb,
			'user',
			'01JZQFGVKZCTV7S310W0BDMWCS',
			'dang.thi.phuong@nikki.com'
		),
		-- Võ Văn Giang (vo.van.giang@nikki.com) | password: Passwo0rd123 | temp: Passwo0rd123 | otp: RBSWY3DPEHPK3PXX
		(
			'01K0AUTH000000000000000009',
			'$argon2id$19$65536$3$2$/hr9fzDjiJjphmxADvaRRg$hWu5KMrK7d1W7bdPU6K9Gb12W2dLfsh7k6MGYwqFjEw',
			NOW() + INTERVAL '180 days',
			NOW() - INTERVAL '45 days',
			'$argon2id$19$65536$3$2$/hr9fzDjiJjphmxADvaRRg$hWu5KMrK7d1W7bdPU6K9Gb12W2dLfsh7k6MGYwqFjEw',
			NOW() + INTERVAL '365 days',
			'RBSWY3DPEHPK3PXX',
			NOW() + INTERVAL '180 days',
			'["N2QR-3STU-4VWX-5YZA", "B2CD-3EFG-4HIJ-5KLM", "P2RS-3TUV-4WXY-5ZAB", "C2DE-3FGH-4JKL-5MNP", "Q2ST-3UVW-4XYZ-5ABC", "D2EF-3GHJ-4KLM-5NPQ", "R2TU-3VWX-4YZA-5BCD", "E2FG-3HJK-4LMN-5PQR", "S2UV-3WXY-4ZAB-5CDE", "F2GH-3JKL-4MNP-5QRS"]'::jsonb,
			'user',
			'01JZQFY6EXRG0959Z95Y2EM3AM',
			'vo.van.giang@nikki.com'
		),
		-- Bùi Thị Hoa (bui.thi.hoa@nikki.com) | password: Passwo0rd123 | temp: Passwo0rd123 | otp: SBSWY3DPEHPK3PXY
		(
			'01K0AUTH00000000000000000A',
			'$argon2id$19$65536$3$2$/hr9fzDjiJjphmxADvaRRg$hWu5KMrK7d1W7bdPU6K9Gb12W2dLfsh7k6MGYwqFjEw',
			NOW() + INTERVAL '365 days',
			NOW() - INTERVAL '3 days',
			'$argon2id$19$65536$3$2$/hr9fzDjiJjphmxADvaRRg$hWu5KMrK7d1W7bdPU6K9Gb12W2dLfsh7k6MGYwqFjEw',
			NOW() + INTERVAL '365 days',
			'SBSWY3DPEHPK3PXY',
			NOW() + INTERVAL '365 days',
			'["G2HJ-3KLM-4NPQ-5RST", "U2VW-3XYZ-4ABC-5DEF", "H2JK-3LMN-4PQR-5STU", "V2WX-3YZA-4BCD-5EFG", "I2KL-3MNP-4QRS-5TUV", "W2XY-3ZAB-4CDE-5FGH", "J2LM-3NPQ-4RST-5UVW", "X2YZ-3ABC-4DEF-5GHJ", "K2MN-3PQR-4STU-5VWX", "Y2ZA-3BCD-4EFG-5HJK"]'::jsonb,
			'user',
			'01JZQFZFK6GM2D5X6MYHWH6FND',
			'bui.thi.hoa@nikki.com'
		);
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables 
		WHERE table_schema = 'public' AND table_name = 'authn_method_settings'
	) THEN
		INSERT INTO "authn_method_settings" (
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
		('01K0AUTH000000000000000101', 'password', 1, 5, 1800, 'domain', '01JWNNJGS70Y07MBEV3AQ0M526', 'system'),
		('01K0AUTH000000000000000102', 'captcha', 2, 3, 900,  'domain', '01JWNNJGS70Y07MBEV3AQ0M526', 'system'),
		('01K0AUTH000000000000000103', 'otpCode',  3, 3, 1800, 'domain', '01JWNNJGS70Y07MBEV3AQ0M526', 'system');
	END IF;
END $$;
