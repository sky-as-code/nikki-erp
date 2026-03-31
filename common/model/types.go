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
	"github.com/sky-as-code/nikki-erp/common/safe"
	val "github.com/sky-as-code/nikki-erp/common/validator"
)

var ZeroTime = time.Time{}

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
	if src == LabelRefLanguageCode {
		return src, nil
	}
	src = strings.ReplaceAll(src, "_", "-")
	parsed, err := language.Parse(src)
	if err != nil {
		return "", err
	}

	// Canonical string from parsed tag (e.g. "en-US")
	canonical := parsed.String()
	return canonical, nil
}

type TranslationKey = string

var translationKeyRules val.Rule = val.RegExp(regexp.MustCompile(`^[a-zA-Z0-9_\.]+$`))

const (
	LanguageCodeEnUS    = LanguageCode("en-US")
	DefaultLanguageCode = LanguageCodeEnUS
	// When this Label Reference is specified in a LangJson, all other keys are ignored.
	//
	LabelRefLanguageCode = LanguageCode("$ref")
)

var langCodeRules = []val.Rule{
	val.NotEmpty,
	val.By(func(value any) error {
		code, ok := value.(string)
		if !ok {
			codePtr, _ := value.(*string)
			code = safe.GetVal(codePtr, "")
		}

		if code != LabelRefLanguageCode && !IsBCP47LanguageCode(code) {
			return errors.New("must be a valid BCP47-compliant language code with region part")
		}
		return nil
	}),
}

func LanguageCodePtrValidateRule(field **LanguageCode, isRequired bool) *val.FieldRules {
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
		if len(labelStr) == 0 || (stdLabelCode != LabelRefLanguageCode && !array.Contains(whitelistLangs, stdLabelCode)) {
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

func (this LangJson) IsEqual(target LangJson) bool {
	return reflect.DeepEqual(this, target)
}

func (this LangJson) TranslationKey() TranslationKey {
	key, hasKey := this[LabelRefLanguageCode]
	if !hasKey {
		return ""
	}
	return key
}

// SetTranslationKey references this LangJson to another translation object.
//
// WARNING: It clears all other keys in this LangJson.
func (this LangJson) SetTranslationKey(key TranslationKey) {
	this[LabelRefLanguageCode] = key
	// Delete all other keys
	for k := range this {
		if k != LabelRefLanguageCode {
			delete(this, k)
		}
	}
}

func (this LangJson) RemoveTranslationKey() {
	delete(this, LabelRefLanguageCode)
}

func LangJsonPtrValidateRule(field **LangJson, isRequired bool, minLength int, maxLength int) *val.FieldRules {
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

			if langCode == LabelRefLanguageCode {
				mapRules = append(mapRules, val.Key(langCode, translationKeyRules))
			} else {
				mapRules = append(mapRules, val.Key(langCode, keyRules...))
			}

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

func NewModelDateTime() ModelDateTime {
	return ModelDateTime(time.Now().UTC())
}

func ParseModelDateTime(timestamp string) (ModelDateTime, error) {
	if !strings.HasSuffix(timestamp, "Z") {
		return ModelDateTime{}, errors.New("timestamp must be in RFC3339 format and in UTC (end with 'Z' only)")
	}

	parsed, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return ModelDateTime{}, err
	}
	return ModelDateTime(parsed), nil
}

// A time.Time wrapper to represent a date-time. Use this to consistently handle date-time thoughout this application.
type ModelDateTime time.Time

func (this ModelDateTime) GoTime() time.Time {
	return time.Time(this)
}

func (this ModelDateTime) String() string {
	return time.Time(this).Format(time.RFC3339)
}

func (this ModelDateTime) MarshalText() ([]byte, error) {
	return []byte(this.String()), nil
}

func (this *ModelDateTime) UnmarshalText(data []byte) error {
	parsed, err := ParseModelDateTime(string(data))
	if err != nil {
		return err
	}
	*this = parsed
	return nil
}

func NewModelDate() ModelDate {
	now := time.Now().UTC()
	y, m, d := now.Date()
	dateOnly := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	return ModelDate(dateOnly)
}

func ParseModelDate(timestamp string) (ModelDate, error) {
	parsed, err := time.Parse(time.DateOnly, timestamp)
	if err != nil {
		return ModelDate{}, err
	}
	return ModelDate(parsed), nil
}

// A time.Time wrapper to represent a date without time. Use this to consistently handle date thoughout this application.
type ModelDate time.Time

func (this ModelDate) GoTime() time.Time {
	return time.Time(this)
}

func (this ModelDate) String() string {
	return time.Time(this).Format(time.DateOnly)
}

func (this ModelDate) MarshalText() ([]byte, error) {
	return []byte(this.String()), nil
}

func (this *ModelDate) UnmarshalText(data []byte) error {
	parsed, err := ParseModelDate(string(data))
	if err != nil {
		return err
	}
	*this = parsed
	return nil
}

func NewModelTime() ModelTime {
	now := time.Now().UTC()
	onlyTime := time.Date(0, 1, 1, now.Hour(), now.Minute(), now.Second(), now.Nanosecond(), time.UTC)
	return ModelTime(onlyTime)
}

func ParseModelTime(timestamp string) (ModelTime, error) {
	parsed, err := time.Parse(time.TimeOnly, timestamp)
	if err != nil {
		return ModelTime{}, err
	}
	return ModelTime(parsed), nil
}

// A time.Time wrapper to represent a time without date. Use this to consistently handle time this application.
type ModelTime time.Time

func (this ModelTime) GoTime() time.Time {
	return time.Time(this)
}

func (this ModelTime) String() string {
	return time.Time(this).Format(time.TimeOnly)
}

func (this ModelTime) MarshalText() ([]byte, error) {
	return []byte(this.String()), nil
}

func (this *ModelTime) UnmarshalText(data []byte) error {
	parsed, err := ParseModelTime(string(data))
	if err != nil {
		return err
	}
	*this = parsed
	return nil
}
