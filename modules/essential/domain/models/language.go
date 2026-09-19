package models

import (
	"regexp"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
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
	LanguageFieldIsoCode            = "iso_code"
	LanguageFieldDirection          = "direction"
	LanguageFieldDecimalSeparator   = "decimal_separator"
	LanguageFieldThousandsSeparator = "thousands_separator"
	LanguageFieldDateFormat         = "date_format"
	LanguageFieldTimeFormat         = "time_format"
	LanguageFieldShortTimeFormat    = "short_time_format"
	LanguageFieldFirstDayOfWeek     = "first_day_of_week"
)

func LanguageSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(LanguageSchemaName).
		Label(model.NewLangJsonRefSf("%s.label", LanguageSchemaName)).
		TableName("essential_languages").
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Field(
			dmodel.DefineField().
				Name(LanguageFieldName).
				DataType(dmodel.FieldDataTypeString(1, model.MODEL_RULE_LONG_NAME_LENGTH)).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(LanguageFieldIsoCode).
				// Hyphenated BCP47, not the underscore form: every consumer of an iso code speaks
				// hyphen -- SupportedLanguages, the stored `language` setting, the LangJson document
				// keys, ToBCP47LanguageCode, i18next on the frontend. An underscore code could not
				// have been matched by any of them.
				DataType(dmodel.FieldDataTypeString(5, 10, dmodel.FieldDataTypeStringOpts{
					Regex: regexp.MustCompile(`^[a-z]{2}-[A-Z]{2}$`),
				})).
				RequiredForCreate().
				Unique(),
		).
		Field(
			dmodel.DefineField().
				Name(LanguageFieldDirection).
				DataType(dmodel.FieldDataTypeEnumString([]string{
					string(LanguageDirectionLtr),
					string(LanguageDirectionRtl),
				})).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(LanguageFieldDecimalSeparator).
				DataType(dmodel.FieldDataTypeString(1, 5)).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(LanguageFieldThousandsSeparator).
				DataType(dmodel.FieldDataTypeString(1, 5)).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(LanguageFieldDateFormat).
				DataType(dmodel.FieldDataTypeString(1, 30)).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(LanguageFieldTimeFormat).
				DataType(dmodel.FieldDataTypeString(1, 30)).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(LanguageFieldShortTimeFormat).
				DataType(dmodel.FieldDataTypeString(1, 30)).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(LanguageFieldFirstDayOfWeek).
				DataType(dmodel.FieldDataTypeEnumString([]string{
					string(WeekDayMonday),
					string(WeekDayTuesday),
					string(WeekDayWednesday),
					string(WeekDayThursday),
					string(WeekDayFriday),
					string(WeekDaySaturday),
					string(WeekDaySunday),
				})).
				RequiredForCreate(),
		).
		Extend(basemodel.VersionedModelSchemaBuilder()).
		Extend(basemodel.AuditableModelSchemaBuilder())
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

// Every field of this model is stored as a string, iso_code included: GetFieldData has no
// LanguageCode accessor, and a getter that silently cast to one would hide the conversion the
// transport layer has to do anyway.

func (this Language) GetName() *string {
	return this.GetFieldData().GetString(LanguageFieldName)
}

func (this *Language) SetName(v *string) {
	this.GetFieldData().SetString(LanguageFieldName, v)
}

func (this Language) GetIsoCode() *string {
	return this.GetFieldData().GetString(LanguageFieldIsoCode)
}

func (this *Language) SetIsoCode(v *string) {
	this.GetFieldData().SetString(LanguageFieldIsoCode, v)
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
