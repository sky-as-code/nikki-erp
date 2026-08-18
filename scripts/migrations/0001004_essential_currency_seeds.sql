-- The ISO 4217 active currency list.
--
-- essential_currencies ships empty, so without this seed no module that records money can create a
-- record at all: a purchase order, an invoice and a payment all carry a currency_id that must
-- resolve. That is why this is a migration rather than an operator task.
--
-- The whole active list is seeded rather than a working subset. A partial list becomes a support
-- ticket the first time someone trades in a currency nobody thought to add, and the rows are inert
-- until one is selected.
--
-- decimal_places is the ISO minor-unit digit count, and it is load-bearing rather than cosmetic:
-- Essential's currency service rounds every computed total to it. VND and JPY are quoted in whole
-- units (0), most currencies in hundredths (2), and the Gulf dinars in thousandths (3). Getting one
-- wrong silently mis-rounds every amount denominated in it.
--
-- Deliberately excluded: the fund codes (XBA..XBD, XDR, XUA), the precious-metal codes (XAU, XAG,
-- XPT, XPD) and the test code XTS. None of them denominates a payable amount, and seeding them
-- would put unusable options in front of every user picking a currency.
--
-- symbol is left NULL. A currency's symbol is a presentation choice that varies by locale and by
-- house style, and a wrong symbol on money is worse than none; the UI falls back to the code.
--
-- Ids are deterministic, derived from the alphabetic code, so this file is safe to re-run and the
-- same currency carries the same id in every deployment. ON CONFLICT DO NOTHING makes re-running a
-- no-op rather than an error, which is what lets a deployment that already seeded some currencies
-- pick up the rest.

DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'essential_currencies'
	) THEN
		INSERT INTO "essential_currencies" (
			"id", "code", "numeric_code", "name", "symbol", "decimal_places", "is_active", "is_archived", "created_at", "etag"
		) VALUES
		('01KZQC0000CURRENCY00004300', 'AED', '784', jsonb_build_object('en-US', 'UAE Dirham', 'vi-VN', 'Dirham UAE', 'zh-TW', '阿聯酋迪拉姆'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY00005D00', 'AFN', '971', jsonb_build_object('en-US', 'Afghani', 'vi-VN', 'Afghani Afghanistan', 'zh-TW', '阿富汗尼'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY0000BB00', 'ALL', '008', jsonb_build_object('en-US', 'Lek', 'vi-VN', 'Lek Albania', 'zh-TW', '阿爾巴尼亞列克'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY0000C300', 'AMD', '051', jsonb_build_object('en-US', 'Armenian Dram', 'vi-VN', 'Dram Armenia', 'zh-TW', '亞美尼亞德拉姆'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY0000D600', 'ANG', '532', jsonb_build_object('en-US', 'Netherlands Antillean Guilder', 'vi-VN', 'Guilder Antilles Hà Lan', 'zh-TW', '荷屬安的列斯盾'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY0000E000', 'AOA', '973', jsonb_build_object('en-US', 'Kwanza', 'vi-VN', 'Kwanza Angola', 'zh-TW', '安哥拉寬扎'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY0000HJ00', 'ARS', '032', jsonb_build_object('en-US', 'Argentine Peso', 'vi-VN', 'Peso Argentina', 'zh-TW', '阿根廷比索'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY0000M300', 'AUD', '036', jsonb_build_object('en-US', 'Australian Dollar', 'vi-VN', 'Đô la Úc', 'zh-TW', '澳幣'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY0000P600', 'AWG', '533', jsonb_build_object('en-US', 'Aruban Florin', 'vi-VN', 'Florin Aruba', 'zh-TW', '阿魯巴弗羅林'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY0000SD00', 'AZN', '944', jsonb_build_object('en-US', 'Azerbaijan Manat', 'vi-VN', 'Manat Azerbaijan', 'zh-TW', '亞塞拜然馬納特'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY00010C00', 'BAM', '977', jsonb_build_object('en-US', 'Convertible Mark', 'vi-VN', 'Mark chuyển đổi Bosnia', 'zh-TW', '波士尼亞可兌換馬克'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY00011300', 'BBD', '052', jsonb_build_object('en-US', 'Barbados Dollar', 'vi-VN', 'Đô la Barbados', 'zh-TW', '巴貝多元'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY00013K00', 'BDT', '050', jsonb_build_object('en-US', 'Taka', 'vi-VN', 'Taka Bangladesh', 'zh-TW', '孟加拉塔卡'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY00016D00', 'BGN', '975', jsonb_build_object('en-US', 'Bulgarian Lev', 'vi-VN', 'Lev Bulgaria', 'zh-TW', '保加利亞列弗'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY00017300', 'BHD', '048', jsonb_build_object('en-US', 'Bahraini Dinar', 'vi-VN', 'Dinar Bahrain', 'zh-TW', '巴林第納爾'), NULL, 3, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY00018500', 'BIF', '108', jsonb_build_object('en-US', 'Burundi Franc', 'vi-VN', 'Franc Burundi', 'zh-TW', '蒲隆地法郎'), NULL, 0, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY0001C300', 'BMD', '060', jsonb_build_object('en-US', 'Bermudian Dollar', 'vi-VN', 'Đô la Bermuda', 'zh-TW', '百慕達元'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY0001D300', 'BND', '096', jsonb_build_object('en-US', 'Brunei Dollar', 'vi-VN', 'Đô la Brunei', 'zh-TW', '汶萊元'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY0001E100', 'BOB', '068', jsonb_build_object('en-US', 'Boliviano', 'vi-VN', 'Boliviano Bolivia', 'zh-TW', '玻利維亞諾'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY0001HB00', 'BRL', '986', jsonb_build_object('en-US', 'Brazilian Real', 'vi-VN', 'Real Brazil', 'zh-TW', '巴西雷亞爾'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY0001J300', 'BSD', '044', jsonb_build_object('en-US', 'Bahamian Dollar', 'vi-VN', 'Đô la Bahamas', 'zh-TW', '巴哈馬元'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY0001KD00', 'BTN', '064', jsonb_build_object('en-US', 'Ngultrum', 'vi-VN', 'Ngultrum Bhutan', 'zh-TW', '不丹努爾特魯姆'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY0001PF00', 'BWP', '072', jsonb_build_object('en-US', 'Pula', 'vi-VN', 'Pula Botswana', 'zh-TW', '波札那普拉'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY0001RD00', 'BYN', '933', jsonb_build_object('en-US', 'Belarusian Ruble', 'vi-VN', 'Rúp Belarus', 'zh-TW', '白俄羅斯盧布'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY0001S300', 'BZD', '084', jsonb_build_object('en-US', 'Belize Dollar', 'vi-VN', 'Đô la Belize', 'zh-TW', '貝里斯元'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY00020300', 'CAD', '124', jsonb_build_object('en-US', 'Canadian Dollar', 'vi-VN', 'Đô la Canada', 'zh-TW', '加拿大元'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY00023500', 'CDF', '976', jsonb_build_object('en-US', 'Congolese Franc', 'vi-VN', 'Franc Congo', 'zh-TW', '剛果法郎'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY00027500', 'CHF', '756', jsonb_build_object('en-US', 'Swiss Franc', 'vi-VN', 'Franc Thụy Sĩ', 'zh-TW', '瑞士法郎'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY0002BF00', 'CLP', '152', jsonb_build_object('en-US', 'Chilean Peso', 'vi-VN', 'Peso Chile', 'zh-TW', '智利比索'), NULL, 0, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY0002DR00', 'CNY', '156', jsonb_build_object('en-US', 'Yuan Renminbi', 'vi-VN', 'Nhân dân tệ', 'zh-TW', '人民幣'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY0002EF00', 'COP', '170', jsonb_build_object('en-US', 'Colombian Peso', 'vi-VN', 'Peso Colombia', 'zh-TW', '哥倫比亞比索'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY0002H200', 'CRC', '188', jsonb_build_object('en-US', 'Costa Rican Colon', 'vi-VN', 'Colon Costa Rica', 'zh-TW', '哥斯大黎加科朗'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY0002MF00', 'CUP', '192', jsonb_build_object('en-US', 'Cuban Peso', 'vi-VN', 'Peso Cuba', 'zh-TW', '古巴比索'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY0002N400', 'CVE', '132', jsonb_build_object('en-US', 'Cabo Verde Escudo', 'vi-VN', 'Escudo Cabo Verde', 'zh-TW', '維德角埃斯庫多'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY0002SA00', 'CZK', '203', jsonb_build_object('en-US', 'Czech Koruna', 'vi-VN', 'Koruna Séc', 'zh-TW', '捷克克朗'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY00039500', 'DJF', '262', jsonb_build_object('en-US', 'Djibouti Franc', 'vi-VN', 'Franc Djibouti', 'zh-TW', '吉布地法郎'), NULL, 0, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY0003AA00', 'DKK', '208', jsonb_build_object('en-US', 'Danish Krone', 'vi-VN', 'Krone Đan Mạch', 'zh-TW', '丹麥克朗'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY0003EF00', 'DOP', '214', jsonb_build_object('en-US', 'Dominican Peso', 'vi-VN', 'Peso Dominica', 'zh-TW', '多明尼加比索'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY0003S300', 'DZD', '012', jsonb_build_object('en-US', 'Algerian Dinar', 'vi-VN', 'Dinar Algeria', 'zh-TW', '阿爾及利亞第納爾'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY00046F00', 'EGP', '818', jsonb_build_object('en-US', 'Egyptian Pound', 'vi-VN', 'Bảng Ai Cập', 'zh-TW', '埃及鎊'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY0004HD00', 'ERN', '232', jsonb_build_object('en-US', 'Nakfa', 'vi-VN', 'Nakfa Eritrea', 'zh-TW', '厄利垂亞納克法'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY0004K100', 'ETB', '230', jsonb_build_object('en-US', 'Ethiopian Birr', 'vi-VN', 'Birr Ethiopia', 'zh-TW', '衣索比亞比爾'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY0004MH00', 'EUR', '978', jsonb_build_object('en-US', 'Euro', 'vi-VN', 'Euro', 'zh-TW', '歐元'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY00059300', 'FJD', '242', jsonb_build_object('en-US', 'Fiji Dollar', 'vi-VN', 'Đô la Fiji', 'zh-TW', '斐濟元'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY0005AF00', 'FKP', '238', jsonb_build_object('en-US', 'Falkland Islands Pound', 'vi-VN', 'Bảng Falkland', 'zh-TW', '福克蘭群島鎊'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY00061F00', 'GBP', '826', jsonb_build_object('en-US', 'Pound Sterling', 'vi-VN', 'Bảng Anh', 'zh-TW', '英鎊'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY00064B00', 'GEL', '981', jsonb_build_object('en-US', 'Lari', 'vi-VN', 'Lari Georgia', 'zh-TW', '喬治亞拉里'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY00067J00', 'GHS', '936', jsonb_build_object('en-US', 'Ghana Cedi', 'vi-VN', 'Cedi Ghana', 'zh-TW', '迦納塞地'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY00068F00', 'GIP', '292', jsonb_build_object('en-US', 'Gibraltar Pound', 'vi-VN', 'Bảng Gibraltar', 'zh-TW', '直布羅陀鎊'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY0006C300', 'GMD', '270', jsonb_build_object('en-US', 'Dalasi', 'vi-VN', 'Dalasi Gambia', 'zh-TW', '甘比亞達拉西'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY0006D500', 'GNF', '324', jsonb_build_object('en-US', 'Guinean Franc', 'vi-VN', 'Franc Guinea', 'zh-TW', '幾內亞法郎'), NULL, 0, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY0006KG00', 'GTQ', '320', jsonb_build_object('en-US', 'Quetzal', 'vi-VN', 'Quetzal Guatemala', 'zh-TW', '瓜地馬拉格查爾'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY0006R300', 'GYD', '328', jsonb_build_object('en-US', 'Guyana Dollar', 'vi-VN', 'Đô la Guyana', 'zh-TW', '蓋亞那元'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY0007A300', 'HKD', '344', jsonb_build_object('en-US', 'Hong Kong Dollar', 'vi-VN', 'Đô la Hồng Kông', 'zh-TW', '港幣'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY0007DB00', 'HNL', '340', jsonb_build_object('en-US', 'Lempira', 'vi-VN', 'Lempira Honduras', 'zh-TW', '宏都拉斯倫皮拉'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY0007K600', 'HTG', '332', jsonb_build_object('en-US', 'Gourde', 'vi-VN', 'Gourde Haiti', 'zh-TW', '海地古德'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY0007M500', 'HUF', '348', jsonb_build_object('en-US', 'Forint', 'vi-VN', 'Forint Hungary', 'zh-TW', '匈牙利福林'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY00083H00', 'IDR', '360', jsonb_build_object('en-US', 'Rupiah', 'vi-VN', 'Rupiah Indonesia', 'zh-TW', '印尼盾'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY0008BJ00', 'ILS', '376', jsonb_build_object('en-US', 'New Israeli Sheqel', 'vi-VN', 'Sheqel Israel', 'zh-TW', '以色列新謝克爾'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY0008DH00', 'INR', '356', jsonb_build_object('en-US', 'Indian Rupee', 'vi-VN', 'Rupee Ấn Độ', 'zh-TW', '印度盧比'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY0008G300', 'IQD', '368', jsonb_build_object('en-US', 'Iraqi Dinar', 'vi-VN', 'Dinar Iraq', 'zh-TW', '伊拉克第納爾'), NULL, 3, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY0008HH00', 'IRR', '364', jsonb_build_object('en-US', 'Iranian Rial', 'vi-VN', 'Rial Iran', 'zh-TW', '伊朗里亞爾'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY0008JA00', 'ISK', '352', jsonb_build_object('en-US', 'Iceland Krona', 'vi-VN', 'Krona Iceland', 'zh-TW', '冰島克朗'), NULL, 0, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY0009C300', 'JMD', '388', jsonb_build_object('en-US', 'Jamaican Dollar', 'vi-VN', 'Đô la Jamaica', 'zh-TW', '牙買加元'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY0009E300', 'JOD', '400', jsonb_build_object('en-US', 'Jordanian Dinar', 'vi-VN', 'Dinar Jordan', 'zh-TW', '約旦第納爾'), NULL, 3, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY0009FR00', 'JPY', '392', jsonb_build_object('en-US', 'Yen', 'vi-VN', 'Yên Nhật', 'zh-TW', '日圓'), NULL, 0, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000A4J00', 'KES', '404', jsonb_build_object('en-US', 'Kenyan Shilling', 'vi-VN', 'Shilling Kenya', 'zh-TW', '肯亞先令'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000A6J00', 'KGS', '417', jsonb_build_object('en-US', 'Som', 'vi-VN', 'Som Kyrgyzstan', 'zh-TW', '吉爾吉斯索姆'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000A7H00', 'KHR', '116', jsonb_build_object('en-US', 'Riel', 'vi-VN', 'Riel Campuchia', 'zh-TW', '柬埔寨瑞爾'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000AC500', 'KMF', '174', jsonb_build_object('en-US', 'Comorian Franc', 'vi-VN', 'Franc Comoros', 'zh-TW', '葛摩法郎'), NULL, 0, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000AFP00', 'KPW', '408', jsonb_build_object('en-US', 'North Korean Won', 'vi-VN', 'Won Triều Tiên', 'zh-TW', '北韓圜'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000AHP00', 'KRW', '410', jsonb_build_object('en-US', 'Won', 'vi-VN', 'Won Hàn Quốc', 'zh-TW', '韓圜'), NULL, 0, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000AP300', 'KWD', '414', jsonb_build_object('en-US', 'Kuwaiti Dinar', 'vi-VN', 'Dinar Kuwait', 'zh-TW', '科威特第納爾'), NULL, 3, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000AR300', 'KYD', '136', jsonb_build_object('en-US', 'Cayman Islands Dollar', 'vi-VN', 'Đô la Cayman', 'zh-TW', '開曼群島元'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000ASK00', 'KZT', '398', jsonb_build_object('en-US', 'Tenge', 'vi-VN', 'Tenge Kazakhstan', 'zh-TW', '哈薩克堅戈'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000B0A00', 'LAK', '418', jsonb_build_object('en-US', 'Lao Kip', 'vi-VN', 'Kip Lào', 'zh-TW', '寮國基普'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000B1F00', 'LBP', '422', jsonb_build_object('en-US', 'Lebanese Pound', 'vi-VN', 'Bảng Liban', 'zh-TW', '黎巴嫩鎊'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000BAH00', 'LKR', '144', jsonb_build_object('en-US', 'Sri Lanka Rupee', 'vi-VN', 'Rupee Sri Lanka', 'zh-TW', '斯里蘭卡盧比'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000BH300', 'LRD', '430', jsonb_build_object('en-US', 'Liberian Dollar', 'vi-VN', 'Đô la Liberia', 'zh-TW', '賴比瑞亞元'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000BJB00', 'LSL', '426', jsonb_build_object('en-US', 'Loti', 'vi-VN', 'Loti Lesotho', 'zh-TW', '賴索托洛蒂'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000BR300', 'LYD', '434', jsonb_build_object('en-US', 'Libyan Dinar', 'vi-VN', 'Dinar Libya', 'zh-TW', '利比亞第納爾'), NULL, 3, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000C0300', 'MAD', '504', jsonb_build_object('en-US', 'Moroccan Dirham', 'vi-VN', 'Dirham Maroc', 'zh-TW', '摩洛哥迪拉姆'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000C3B00', 'MDL', '498', jsonb_build_object('en-US', 'Moldovan Leu', 'vi-VN', 'Leu Moldova', 'zh-TW', '摩爾多瓦列伊'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000C6000', 'MGA', '969', jsonb_build_object('en-US', 'Malagasy Ariary', 'vi-VN', 'Ariary Madagascar', 'zh-TW', '馬達加斯加阿里亞里'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000CA300', 'MKD', '807', jsonb_build_object('en-US', 'Denar', 'vi-VN', 'Denar Bắc Macedonia', 'zh-TW', '北馬其頓代納爾'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000CCA00', 'MMK', '104', jsonb_build_object('en-US', 'Kyat', 'vi-VN', 'Kyat Myanmar', 'zh-TW', '緬甸元'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000CDK00', 'MNT', '496', jsonb_build_object('en-US', 'Tugrik', 'vi-VN', 'Tugrik Mông Cổ', 'zh-TW', '蒙古圖格里克'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000CEF00', 'MOP', '446', jsonb_build_object('en-US', 'Pataca', 'vi-VN', 'Pataca Macao', 'zh-TW', '澳門元'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000CHM00', 'MRU', '929', jsonb_build_object('en-US', 'Ouguiya', 'vi-VN', 'Ouguiya Mauritania', 'zh-TW', '茅利塔尼亞烏吉亞'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000CMH00', 'MUR', '480', jsonb_build_object('en-US', 'Mauritius Rupee', 'vi-VN', 'Rupee Mauritius', 'zh-TW', '模里西斯盧比'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000CNH00', 'MVR', '462', jsonb_build_object('en-US', 'Rufiyaa', 'vi-VN', 'Rufiyaa Maldives', 'zh-TW', '馬爾地夫拉菲亞'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000CPA00', 'MWK', '454', jsonb_build_object('en-US', 'Malawi Kwacha', 'vi-VN', 'Kwacha Malawi', 'zh-TW', '馬拉威克瓦查'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000CQD00', 'MXN', '484', jsonb_build_object('en-US', 'Mexican Peso', 'vi-VN', 'Peso Mexico', 'zh-TW', '墨西哥比索'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000CRH00', 'MYR', '458', jsonb_build_object('en-US', 'Malaysian Ringgit', 'vi-VN', 'Ringgit Malaysia', 'zh-TW', '馬來西亞令吉'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000CSD00', 'MZN', '943', jsonb_build_object('en-US', 'Mozambique Metical', 'vi-VN', 'Metical Mozambique', 'zh-TW', '莫三比克梅蒂卡爾'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000D0300', 'NAD', '516', jsonb_build_object('en-US', 'Namibia Dollar', 'vi-VN', 'Đô la Namibia', 'zh-TW', '納米比亞元'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000D6D00', 'NGN', '566', jsonb_build_object('en-US', 'Naira', 'vi-VN', 'Naira Nigeria', 'zh-TW', '奈及利亞奈拉'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000D8E00', 'NIO', '558', jsonb_build_object('en-US', 'Cordoba Oro', 'vi-VN', 'Cordoba Nicaragua', 'zh-TW', '尼加拉瓜科多巴'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000DEA00', 'NOK', '578', jsonb_build_object('en-US', 'Norwegian Krone', 'vi-VN', 'Krone Na Uy', 'zh-TW', '挪威克朗'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000DFH00', 'NPR', '524', jsonb_build_object('en-US', 'Nepalese Rupee', 'vi-VN', 'Rupee Nepal', 'zh-TW', '尼泊爾盧比'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000DS300', 'NZD', '554', jsonb_build_object('en-US', 'New Zealand Dollar', 'vi-VN', 'Đô la New Zealand', 'zh-TW', '紐西蘭元'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000ECH00', 'OMR', '512', jsonb_build_object('en-US', 'Rial Omani', 'vi-VN', 'Rial Oman', 'zh-TW', '阿曼里亞爾'), NULL, 3, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000F0100', 'PAB', '590', jsonb_build_object('en-US', 'Balboa', 'vi-VN', 'Balboa Panama', 'zh-TW', '巴拿馬巴波亞'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000F4D00', 'PEN', '604', jsonb_build_object('en-US', 'Sol', 'vi-VN', 'Sol Peru', 'zh-TW', '秘魯索爾'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000F6A00', 'PGK', '598', jsonb_build_object('en-US', 'Kina', 'vi-VN', 'Kina Papua New Guinea', 'zh-TW', '巴布亞紐幾內亞基那'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000F7F00', 'PHP', '608', jsonb_build_object('en-US', 'Philippine Peso', 'vi-VN', 'Peso Philippines', 'zh-TW', '菲律賓比索'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000FAH00', 'PKR', '586', jsonb_build_object('en-US', 'Pakistan Rupee', 'vi-VN', 'Rupee Pakistan', 'zh-TW', '巴基斯坦盧比'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000FBD00', 'PLN', '985', jsonb_build_object('en-US', 'Zloty', 'vi-VN', 'Zloty Ba Lan', 'zh-TW', '波蘭茲羅提'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000FR600', 'PYG', '600', jsonb_build_object('en-US', 'Guarani', 'vi-VN', 'Guarani Paraguay', 'zh-TW', '巴拉圭瓜拉尼'), NULL, 0, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000G0H00', 'QAR', '634', jsonb_build_object('en-US', 'Qatari Rial', 'vi-VN', 'Rial Qatar', 'zh-TW', '卡達里亞爾'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000HED00', 'RON', '946', jsonb_build_object('en-US', 'Romanian Leu', 'vi-VN', 'Leu Romania', 'zh-TW', '羅馬尼亞列伊'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000HJ300', 'RSD', '941', jsonb_build_object('en-US', 'Serbian Dinar', 'vi-VN', 'Dinar Serbia', 'zh-TW', '塞爾維亞第納爾'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000HM100', 'RUB', '643', jsonb_build_object('en-US', 'Russian Ruble', 'vi-VN', 'Rúp Nga', 'zh-TW', '俄羅斯盧布'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000HP500', 'RWF', '646', jsonb_build_object('en-US', 'Rwanda Franc', 'vi-VN', 'Franc Rwanda', 'zh-TW', '盧安達法郎'), NULL, 0, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000J0H00', 'SAR', '682', jsonb_build_object('en-US', 'Saudi Riyal', 'vi-VN', 'Riyal Ả Rập Xê Út', 'zh-TW', '沙烏地里亞爾'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000J1300', 'SBD', '090', jsonb_build_object('en-US', 'Solomon Islands Dollar', 'vi-VN', 'Đô la Solomon', 'zh-TW', '索羅門群島元'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000J2H00', 'SCR', '690', jsonb_build_object('en-US', 'Seychelles Rupee', 'vi-VN', 'Rupee Seychelles', 'zh-TW', '塞席爾盧比'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000J3600', 'SDG', '938', jsonb_build_object('en-US', 'Sudanese Pound', 'vi-VN', 'Bảng Sudan', 'zh-TW', '蘇丹鎊'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000J4A00', 'SEK', '752', jsonb_build_object('en-US', 'Swedish Krona', 'vi-VN', 'Krona Thụy Điển', 'zh-TW', '瑞典克朗'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000J6300', 'SGD', '702', jsonb_build_object('en-US', 'Singapore Dollar', 'vi-VN', 'Đô la Singapore', 'zh-TW', '新加坡元'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000J7F00', 'SHP', '654', jsonb_build_object('en-US', 'Saint Helena Pound', 'vi-VN', 'Bảng Saint Helena', 'zh-TW', '聖赫勒拿鎊'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000JB400', 'SLE', '925', jsonb_build_object('en-US', 'Leone', 'vi-VN', 'Leone Sierra Leone', 'zh-TW', '獅子山利昂'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000JEJ00', 'SOS', '706', jsonb_build_object('en-US', 'Somali Shilling', 'vi-VN', 'Shilling Somalia', 'zh-TW', '索馬利亞先令'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000JH300', 'SRD', '968', jsonb_build_object('en-US', 'Surinam Dollar', 'vi-VN', 'Đô la Suriname', 'zh-TW', '蘇利南元'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000JJF00', 'SSP', '728', jsonb_build_object('en-US', 'South Sudanese Pound', 'vi-VN', 'Bảng Nam Sudan', 'zh-TW', '南蘇丹鎊'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000JKD00', 'STN', '930', jsonb_build_object('en-US', 'Dobra', 'vi-VN', 'Dobra Sao Tome', 'zh-TW', '聖多美多布拉'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000JN200', 'SVC', '222', jsonb_build_object('en-US', 'El Salvador Colon', 'vi-VN', 'Colon El Salvador', 'zh-TW', '薩爾瓦多科朗'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000JRF00', 'SYP', '760', jsonb_build_object('en-US', 'Syrian Pound', 'vi-VN', 'Bảng Syria', 'zh-TW', '敘利亞鎊'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000JSB00', 'SZL', '748', jsonb_build_object('en-US', 'Lilangeni', 'vi-VN', 'Lilangeni Eswatini', 'zh-TW', '史瓦帝尼里蘭吉尼'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000K7100', 'THB', '764', jsonb_build_object('en-US', 'Baht', 'vi-VN', 'Baht Thái Lan', 'zh-TW', '泰銖'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000K9J00', 'TJS', '972', jsonb_build_object('en-US', 'Somoni', 'vi-VN', 'Somoni Tajikistan', 'zh-TW', '塔吉克索莫尼'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000KCK00', 'TMT', '934', jsonb_build_object('en-US', 'Turkmenistan New Manat', 'vi-VN', 'Manat Turkmenistan', 'zh-TW', '土庫曼馬納特'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000KD300', 'TND', '788', jsonb_build_object('en-US', 'Tunisian Dinar', 'vi-VN', 'Dinar Tunisia', 'zh-TW', '突尼西亞第納爾'), NULL, 3, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000KEF00', 'TOP', '776', jsonb_build_object('en-US', 'Pa''anga', 'vi-VN', 'Pa''anga Tonga', 'zh-TW', '東加潘加'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000KHR00', 'TRY', '949', jsonb_build_object('en-US', 'Turkish Lira', 'vi-VN', 'Lira Thổ Nhĩ Kỳ', 'zh-TW', '土耳其里拉'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000KK300', 'TTD', '780', jsonb_build_object('en-US', 'Trinidad and Tobago Dollar', 'vi-VN', 'Đô la Trinidad và Tobago', 'zh-TW', '千里達及托巴哥元'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000KP300', 'TWD', '901', jsonb_build_object('en-US', 'New Taiwan Dollar', 'vi-VN', 'Đô la Đài Loan', 'zh-TW', '新臺幣'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000KSJ00', 'TZS', '834', jsonb_build_object('en-US', 'Tanzanian Shilling', 'vi-VN', 'Shilling Tanzania', 'zh-TW', '坦尚尼亞先令'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000M0700', 'UAH', '980', jsonb_build_object('en-US', 'Hryvnia', 'vi-VN', 'Hryvnia Ukraine', 'zh-TW', '烏克蘭格里夫納'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000M6Q00', 'UGX', '800', jsonb_build_object('en-US', 'Uganda Shilling', 'vi-VN', 'Shilling Uganda', 'zh-TW', '烏干達先令'), NULL, 0, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000MJ300', 'USD', '840', jsonb_build_object('en-US', 'US Dollar', 'vi-VN', 'Đô la Mỹ', 'zh-TW', '美元'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000MRM00', 'UYU', '858', jsonb_build_object('en-US', 'Peso Uruguayo', 'vi-VN', 'Peso Uruguay', 'zh-TW', '烏拉圭比索'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000MSJ00', 'UZS', '860', jsonb_build_object('en-US', 'Uzbekistan Sum', 'vi-VN', 'Sum Uzbekistan', 'zh-TW', '烏茲別克蘇姆'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000N4300', 'VED', '926', jsonb_build_object('en-US', 'Bolivar Soberano', 'vi-VN', 'Bolivar Venezuela', 'zh-TW', '委內瑞拉玻利瓦'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000ND300', 'VND', '704', jsonb_build_object('en-US', 'Dong', 'vi-VN', 'Đồng Việt Nam', 'zh-TW', '越南盾'), NULL, 0, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000NMN00', 'VUV', '548', jsonb_build_object('en-US', 'Vatu', 'vi-VN', 'Vatu Vanuatu', 'zh-TW', '萬那杜瓦圖'), NULL, 0, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000PJK00', 'WST', '882', jsonb_build_object('en-US', 'Tala', 'vi-VN', 'Tala Samoa', 'zh-TW', '薩摩亞塔拉'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000Q0500', 'XAF', '950', jsonb_build_object('en-US', 'CFA Franc BEAC', 'vi-VN', 'Franc CFA BEAC', 'zh-TW', '中非法郎'), NULL, 0, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000Q2300', 'XCD', '951', jsonb_build_object('en-US', 'East Caribbean Dollar', 'vi-VN', 'Đô la Đông Caribe', 'zh-TW', '東加勒比元'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000QE500', 'XOF', '952', jsonb_build_object('en-US', 'CFA Franc BCEAO', 'vi-VN', 'Franc CFA BCEAO', 'zh-TW', '西非法郎'), NULL, 0, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000QF500', 'XPF', '953', jsonb_build_object('en-US', 'CFP Franc', 'vi-VN', 'Franc CFP', 'zh-TW', '太平洋法郎'), NULL, 0, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000R4H00', 'YER', '886', jsonb_build_object('en-US', 'Yemeni Rial', 'vi-VN', 'Rial Yemen', 'zh-TW', '葉門里亞爾'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000S0H00', 'ZAR', '710', jsonb_build_object('en-US', 'Rand', 'vi-VN', 'Rand Nam Phi', 'zh-TW', '南非蘭特'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000SCP00', 'ZMW', '967', jsonb_build_object('en-US', 'Zambian Kwacha', 'vi-VN', 'Kwacha Zambia', 'zh-TW', '尚比亞克瓦查'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZQC0000CURRENCY000SP600', 'ZWG', '924', jsonb_build_object('en-US', 'Zimbabwe Gold', 'vi-VN', 'Zimbabwe Gold', 'zh-TW', '辛巴威黃金'), NULL, 2, TRUE, FALSE, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT DO NOTHING;
	END IF;
END $$;
