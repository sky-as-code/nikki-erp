package model

const (
	MODEL_RULE_DESC_LENGTH         = 3000
	MODEL_RULE_EMAIL_LENGTH        = 254
	MODEL_RULE_ETAG_MIN_LENGTH     = 7
	MODEL_RULE_ETAG_MAX_LENGTH     = 30
	MODEL_RULE_NON_NIKKI_ID_LENGTH = 50

	// Allow any Unicode character, but reject
	// !@#$%^&*()_+=
	// 1234
	// Symbols like -, ., ,, etc.
	// MODEL_RULE_NO_SPECIAL_CHAR = regexp.MustCompile(``) // No special character, but allow Unicode chars

	MODEL_RULE_LONG_NAME_LENGTH  = 200
	MODEL_RULE_SHORT_NAME_LENGTH = 100
	MODEL_RULE_TINY_NAME_LENGTH  = 50
	MODEL_RULE_ULID_LENGTH       = 26
)
