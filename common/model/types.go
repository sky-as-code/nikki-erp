package model

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	"go.bryk.io/pkg/errors"
	"go.bryk.io/pkg/ulid"
	"golang.org/x/text/language"

	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/defense"
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

func EtagPtrValidateRule(field **Etag, isRequired bool) *val.FieldRules {
	return val.Field(field,
		val.NotNilWhen(isRequired),
		val.When(*field != nil,
			val.NotEmpty,
			val.Length(MODEL_RULE_ETAG_MIN_LENGTH, MODEL_RULE_ETAG_MAX_LENGTH),
		),
	)
}

func EtagValidateRule(field *Etag, isRequired bool) *val.FieldRules {
	return val.Field(field,
		val.NotEmpty,
		val.Length(MODEL_RULE_ETAG_MIN_LENGTH, MODEL_RULE_ETAG_MAX_LENGTH),
	)
}

type Slug = string

func SlugPtrValidateRule(field **Slug, isRequired bool) *val.FieldRules {
	return val.Field(field,
		val.NotNilWhen(isRequired),
		val.When(*field != nil,
			val.NotEmpty,
			val.Length(1, MODEL_RULE_SHORT_NAME_LENGTH),
			val.RegExp(regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)),
		),
	)
}

func SlugValidateRule(field *Slug, isRequired bool) *val.FieldRules {
	return val.Field(field,
		val.NotEmpty,
		val.Length(1, MODEL_RULE_SHORT_NAME_LENGTH),
		val.RegExp(regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)),
	)
}

type OpResult[TData any] struct {
	Data TData `json:"data"`

	// Indicates whether "Data" has value. If ClientError is nil but HasData is false,
	// it means the query is successfull but doesn't return any data.
	HasData     bool            `json:"hasData"`
	ClientError *ft.ClientError `json:"error,omitempty"`
}

func (this OpResult[TData]) GetClientError() *ft.ClientError {
	return this.ClientError
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

// LanguageCode is a BCP47-compliant language code with region part.
// It must be an alias of string to easily map from map[LanguageCode]string to LangJson
type LanguageCode = string

func IsBCP47LanguageCode(src string) bool {
	canonical, err := ToBCP47LanguageCode(src)
	if err != nil {
		return false
	}

	// Must have at least one hyphen (i.e. at least 2 parts)
	if strings.Count(canonical, "-") >= 1 {
		return true
	}

	return false
}

func ToBCP47LanguageCode(src string) (string, error) {
	src = strings.ReplaceAll(src, "_", "-")
	parsed, err := language.Parse(src)
	if err != nil {
		return "", err
	}

	// Canonical string from parsed tag (e.g. "en-US")
	canonical := parsed.String()
	return canonical, nil
}

const (
	LanguageCodeEnUS    = LanguageCode("en-US")
	DefaultLanguageCode = LanguageCodeEnUS
)

var langCodeRules = []val.Rule{
	val.NotEmpty,
	val.By(func(value any) error {
		code, _ := value.(string)
		if !IsBCP47LanguageCode(code) {
			return errors.New("must be a valid BCP47-compliant language code with region part")
		}
		return nil
	}),
}

func LanguageCodeValidateRule(field **LanguageCode, isRequired bool) *val.FieldRules {
	rules := []val.Rule{
		val.NotNilWhen(isRequired),
		val.When(*field != nil, langCodeRules...),
	}
	return val.Field(field, rules...)
}

type LangJson map[LanguageCode]string

func (this LangJson) SanitizeClone(whitelistLangs []LanguageCode, isRichText bool) (*LangJson, int, error) {
	sanitizedLabel := make(LangJson)
	fieldCount := 0
	for labelCode, labelStr := range this {
		stdLabelCode, err := ToBCP47LanguageCode(labelCode)
		if err != nil {
			return nil, 0, err
		}
		if len(labelStr) == 0 || !array.Contains(whitelistLangs, stdLabelCode) {
			continue
		}
		if isRichText {
			sanitizedLabel[stdLabelCode] = defense.SanitizeRichText(labelStr)
			fieldCount++
		} else {
			sanitizedLabel[stdLabelCode] = defense.SanitizePlainText(labelStr)
			fieldCount++
		}
	}
	return &sanitizedLabel, fieldCount, nil
}

// Transform creates a new LangJson with the same keys but
// with the values transformed by the given function.
func (this LangJson) Transform(fn func(key LanguageCode, value string) string) LangJson {
	transformed := make(LangJson)
	for key, value := range this {
		transformed[key] = fn(key, value)
	}
	return transformed
}

func LangJsonValidateRule(field **LangJson, isRequired bool, minLength int, maxLength int) *val.FieldRules {
	fieldValue := *field
	mapRules := []*val.KeyRules{}
	keyRules := []val.Rule{
		val.NotEmpty,
		val.Length(minLength, maxLength),
	}

	// This is a map instead of an array to workaround the validator limitation:
	// With array, the error message only includes the index. E.g: "0: must be a language code; 1: must be a valid language code.".
	// With map, the error message includes the key name. E.g: "en_lish: must be a language code".
	allKeys := make(map[LanguageCode]LanguageCode)

	if fieldValue != nil {
		for langCode := range *fieldValue {
			allKeys[langCode] = langCode
			mapRules = append(mapRules, val.Key(langCode, keyRules...))
		}
	}

	return val.Field(field,
		val.NotNilWhen(isRequired),
		val.When(*field != nil,
			val.NotEmpty,
			val.Map(mapRules...),
			val.By(func(_ any) error {
				return val.ApiBased.ValidateRaw(allKeys, val.Each(langCodeRules...))
			}),
		),
	)
}

func init() {
	AddConversion[LangJson, *LangJson](func(in reflect.Value) (reflect.Value, error) {
		if in.IsNil() {
			return reflect.ValueOf((*LangJson)(nil)), nil
		}

		result := in.Interface().(LangJson)
		return reflect.ValueOf(&result), nil
	})

	AddConversion[*LangJson, LangJson](func(in reflect.Value) (reflect.Value, error) {
		if in.IsNil() {
			return reflect.ValueOf((LangJson)(nil)), nil
		}

		result := *in.Interface().(*LangJson)
		return reflect.ValueOf(result), nil
	})

	AddConversion[Id, *Id](func(in reflect.Value) (reflect.Value, error) {
		result := in.Interface().(Id)
		return reflect.ValueOf(&result), nil
	})

	AddConversion[*Id, Id](func(in reflect.Value) (reflect.Value, error) {
		if in.IsNil() {
			return reflect.ValueOf(Id("")), nil
		}

		result := *in.Interface().(*Id)
		return reflect.ValueOf(result), nil
	})
}
