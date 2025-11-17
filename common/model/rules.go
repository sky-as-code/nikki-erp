package model

import (
	"math"
	"regexp"

	val "github.com/sky-as-code/nikki-erp/common/validator"
)

const (
	MODEL_RULE_BCP47_LANGUAGE_CODE_LENGTH = 64
	MODEL_RULE_ID_ARR_MAX                 = 100
	MODEL_RULE_DESC_LENGTH                = 3000
	MODEL_RULE_ETAG_MIN_LENGTH            = 7
	MODEL_RULE_ETAG_MAX_LENGTH            = 30
	MODEL_RULE_MAX_INT16                  = math.MaxInt16
	MODEL_RULE_MAX_INT64                  = math.MaxInt64
	MODEL_RULE_NON_NIKKI_ID_LENGTH        = 50
	MODEL_RULE_PAGE_INDEX_START           = 0
	MODEL_RULE_PAGE_INDEX_END             = math.MaxInt16
	MODEL_RULE_PAGE_DEFAULT_SIZE          = 50
	MODEL_RULE_PAGE_MAX_SIZE              = 500
	MODEL_RULE_PAGE_MIN_SIZE              = 1
	MODEL_RULE_PASSWORD_MIN_LENGTH        = 8
	MODEL_RULE_PASSWORD_MAX_LENGTH        = 100
	MODEL_RULE_LONG_NAME_LENGTH           = 200
	MODEL_RULE_SHORT_NAME_LENGTH          = 100
	MODEL_RULE_TINY_NAME_LENGTH           = 50
	MODEL_RULE_ULID_LENGTH                = 26
	MODEL_RULE_URL_LENGTH                 = 2000
	MODEL_RULE_USERNAME_LENGTH            = 254
)

var ModelRuleCodeName = val.RegExp(regexp.MustCompile(`^[a-zA-Z0-9_]+$`))
