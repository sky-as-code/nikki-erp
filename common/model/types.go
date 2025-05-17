package model

import (
	"fmt"
	"regexp"
	"time"

	"go.bryk.io/pkg/ulid"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
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

func IdValidateRule(field any, isRequired bool) *val.FieldRules {
	return val.Field(field,
		val.RequiredWhen(isRequired),
		val.Length(26, 26),
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
		val.RequiredWhen(isRequired),
		val.Length(19, 30),
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
		val.RequiredWhen(isRequired),
		val.Length(1, 50),
		val.RegExp(regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)),
	)
}

type OpResult[TData any] struct {
	Data        TData           `json:"data"`
	ClientError *ft.ClientError `json:"error,omitempty"`
}
