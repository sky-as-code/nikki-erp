package model

import (
	"fmt"
	"regexp"
	"time"

	"go.bryk.io/pkg/ulid"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	val "github.com/sky-as-code/nikki-erp/common/validator"
)

type Id = string

func NewId() (*Id, error) {
	newUlid, err := ulid.New()
	if err != nil {
		return nil, err
	}
	id := Id(newUlid.String())
	return &id, nil
}

func IdPtrValidateRule(field **Id, isRequired bool) *val.FieldRules {
	return val.Field(field,
		val.NotNilWhen(isRequired),
		val.When(*field != nil,
			val.NotEmpty,
			val.Length(MODEL_RULE_ULID_LENGTH, MODEL_RULE_ULID_LENGTH),
		),
	)
}

func IdValidateRule(field *Id, isRequired bool) *val.FieldRules {
	return val.Field(field,
		val.NotEmptyWhen(isRequired),
		val.Length(MODEL_RULE_ULID_LENGTH, MODEL_RULE_ULID_LENGTH),
	)
}

func IdValidateRuleMulti(field *[]Id, isRequired bool, minLength int, maxLength int) *val.FieldRules {
	return val.Field(field,
		val.NotNilWhen(isRequired),
		val.When(*field != nil,
			val.When(minLength > 0,
				val.NotEmpty,
			),
			val.Length(minLength, maxLength),
			val.Each(
				val.NotEmpty,
				val.Length(MODEL_RULE_ULID_LENGTH, MODEL_RULE_ULID_LENGTH),
			),
		),
	)
}

type Etag = string

func NewEtag() *Etag {
	etag := Etag(fmt.Sprintf("%d", time.Now().UnixNano()))
	return &etag
}

func EtagValidateRule(field any, isRequired bool) *val.FieldRules {
	return val.Field(field,
		val.NotNilWhen(isRequired),
		val.NotEmpty,
		val.Length(MODEL_RULE_ETAG_MIN_LENGTH, MODEL_RULE_ETAG_MAX_LENGTH),
	)
}

type Slug = string

func SlugValidateRule(field any, isRequired bool) *val.FieldRules {
	return val.Field(field,
		val.NotNilWhen(isRequired),
		val.NotEmpty,
		val.Length(1, MODEL_RULE_SHORT_NAME_LENGTH),
		val.RegExp(regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)),
	)
}

type OpResult[TData any] struct {
	Data        TData           `json:"data"`
	ClientError *ft.ClientError `json:"error,omitempty"`
}

func PageIndexValidateRule(field **int) *val.FieldRules {
	return val.Field(field,
		val.Min(MODEL_RULE_PAGE_INDEX_START),
		val.Max(MODEL_RULE_PAGE_INDEX_END),
	)
}

func PageSizeValidateRule(field **int) *val.FieldRules {
	return val.Field(field, val.When(*field != nil,
		val.NotEmpty,
		val.Min(MODEL_RULE_PAGE_MIN_SIZE),
		val.Max(MODEL_RULE_PAGE_MAX_SIZE),
	))
}
