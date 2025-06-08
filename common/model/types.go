package model

import (
	"fmt"
	"regexp"
	"time"

	"go.bryk.io/pkg/ulid"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	util "github.com/sky-as-code/nikki-erp/common/util"
	val "github.com/sky-as-code/nikki-erp/common/validator"
)

type Id string

func (this Id) String() string {
	return string(this)
}

func NewId() (*Id, error) {
	newUlid, err := ulid.New()
	if err != nil {
		return nil, err
	}
	id := Id(newUlid.String())
	return &id, nil
}

func WrapId(s string) *Id {
	id := Id(s)
	return &id
}

func WrapNillableId(s *string) *Id {
	if s == nil {
		return nil
	}
	id := Id(*s)
	return &id
}

func IdToNillableStr(id *Id) *string {
	if id == nil {
		return nil
	}
	return util.ToPtr(id.String())
}

func IdValidateRule(field any, isRequired bool) *val.FieldRules {
	return val.Field(field,
		val.NotNilWhen(isRequired),
		val.NotEmpty,
		val.Length(MODEL_RULE_ULID_LENGTH, MODEL_RULE_ULID_LENGTH),
	)
}

type Etag string

func (this Etag) String() string {
	return string(this)
}

func NewEtag() *Etag {
	etag := Etag(fmt.Sprintf("%d", time.Now().UnixNano()))
	return &etag
}

func WrapEtag(s string) *Etag {
	etag := Etag(s)
	return &etag
}

func EtagValidateRule(field any, isRequired bool) *val.FieldRules {
	return val.Field(field,
		val.NotNilWhen(isRequired),
		val.NotEmpty,
		val.Length(MODEL_RULE_ETAG_MIN_LENGTH, MODEL_RULE_ETAG_MAX_LENGTH),
	)
}

type Slug string

func (this Slug) String() string {
	return string(this)
}

func WrapSlug(s string) *Slug {
	slug := Slug(s)
	return &slug
}

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
