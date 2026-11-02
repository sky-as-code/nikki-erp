package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

type LanguageDirection string
type WeekDay string

const (
	LanguageDirectionLtr = LanguageDirection("ltr")
	LanguageDirectionRtl = LanguageDirection("rtl")
)

const (
	WeekDayMonday    = WeekDay("monday")
	WeekDayTuesday   = WeekDay("tuesday")
	WeekDayWednesday = WeekDay("wednesday")
	WeekDayThursday  = WeekDay("thursday")
	WeekDayFriday    = WeekDay("friday")
	WeekDaySaturday  = WeekDay("saturday")
	WeekDaySunday    = WeekDay("sunday")
)

const (
	LanguageSchemaName = "essential_language"

	LanguageFieldId                 = basemodel.FieldId
	LanguageFieldName               = "name"
	LanguageFieldLocaleCode         = "locale_code"
	LanguageFieldDirection          = "direction"
	LanguageFieldDecimalSeparator   = "decimal_separator"
	LanguageFieldThousandsSeparator = "thousands_separator"
	LanguageFieldGrouping           = "grouping"
	LanguageFieldDateFormat         = "date_format"
	LanguageFieldTimeFormat         = "time_format"
	LanguageFieldShortTimeFormat    = "short_time_format"
	LanguageFieldFirstDayOfWeek     = "first_day_of_week"
	LanguageFieldIsActive           = "is_active"
	LanguageFieldPublicUse          = "public_use"
)

//go:embed language.json
var languageSchemaJson string

func LanguageSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(languageSchemaJson)
}

type Language struct {
	basemodel.DynamicModelBase
}

func NewLanguage() *Language {
	return &Language{basemodel.NewDynamicModel()}
}

func NewLanguageFrom(src dmodel.DynamicFields) *Language {
	return &Language{basemodel.NewDynamicModel(src)}
}

// Every field of this model is stored as a string, locale_code included: GetFieldData has no
// LanguageCode accessor, and a getter that silently cast to one would hide the conversion the
// transport layer has to do anyway.

func (this Language) GetName() *string {
	return this.GetFieldData().GetString(LanguageFieldName)
}

func (this *Language) SetName(v *string) {
	this.GetFieldData().SetString(LanguageFieldName, v)
}

func (this Language) GetLocaleCode() *string {
	return this.GetFieldData().GetString(LanguageFieldLocaleCode)
}

func (this *Language) SetLocaleCode(v *string) {
	this.GetFieldData().SetString(LanguageFieldLocaleCode, v)
}

func (this Language) GetDirection() *string {
	return this.GetFieldData().GetString(LanguageFieldDirection)
}

func (this *Language) SetDirection(v *string) {
	this.GetFieldData().SetString(LanguageFieldDirection, v)
}

func (this Language) GetDecimalSeparator() *string {
	return this.GetFieldData().GetString(LanguageFieldDecimalSeparator)
}

func (this *Language) SetDecimalSeparator(v *string) {
	this.GetFieldData().SetString(LanguageFieldDecimalSeparator, v)
}

func (this Language) GetThousandsSeparator() *string {
	return this.GetFieldData().GetString(LanguageFieldThousandsSeparator)
}

func (this *Language) SetThousandsSeparator(v *string) {
	this.GetFieldData().SetString(LanguageFieldThousandsSeparator, v)
}

// GetGrouping returns the digit grouping as its stored JSON int array, e.g. "[3,0]".
func (this Language) GetGrouping() *string {
	return this.GetFieldData().GetString(LanguageFieldGrouping)
}

func (this *Language) SetGrouping(v *string) {
	this.GetFieldData().SetString(LanguageFieldGrouping, v)
}

func (this Language) GetDateFormat() *string {
	return this.GetFieldData().GetString(LanguageFieldDateFormat)
}

func (this *Language) SetDateFormat(v *string) {
	this.GetFieldData().SetString(LanguageFieldDateFormat, v)
}

func (this Language) GetTimeFormat() *string {
	return this.GetFieldData().GetString(LanguageFieldTimeFormat)
}

func (this *Language) SetTimeFormat(v *string) {
	this.GetFieldData().SetString(LanguageFieldTimeFormat, v)
}

func (this Language) GetShortTimeFormat() *string {
	return this.GetFieldData().GetString(LanguageFieldShortTimeFormat)
}

func (this *Language) SetShortTimeFormat(v *string) {
	this.GetFieldData().SetString(LanguageFieldShortTimeFormat, v)
}

func (this Language) GetFirstDayOfWeek() *string {
	return this.GetFieldData().GetString(LanguageFieldFirstDayOfWeek)
}

func (this *Language) SetFirstDayOfWeek(v *string) {
	this.GetFieldData().SetString(LanguageFieldFirstDayOfWeek, v)
}

// GetIsActive reads the computed per-organization activation; nil when the read had no acting org.
func (this Language) GetIsActive() *bool {
	return this.GetFieldData().GetBool(LanguageFieldIsActive)
}

func (this Language) GetPublicUse() *bool {
	return this.GetFieldData().GetBool(LanguageFieldPublicUse)
}

func (this *Language) SetPublicUse(v *bool) {
	this.GetFieldData().SetBool(LanguageFieldPublicUse, v)
}
