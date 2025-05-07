package model

import (
	"fmt"
	"regexp"
	"time"

	"go.bryk.io/pkg/ulid"

	val "github.com/sky-as-code/nikki-erp/common/validator"
)

type Id string

func (this Id) String() string {
	return string(this)
}

func NewId() Id {
	return Id(MustNewULID())
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

func MustNewULID() string {
	newUlid, err := ulid.New()
	if err != nil {
		panic(err)
	}
	return newUlid.String()
}

func IdValidateRule(field any, isRequired bool) *val.FieldRules {
	return val.Field(field,
		val.RequiredWhen(isRequired),
		val.Length(36, 36),
	)
}

type Etag string

func (this Etag) String() string {
	return string(this)
}

func NewEtag() Etag {
	return Etag(fmt.Sprintf("%d", time.Now().UnixNano()))
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
