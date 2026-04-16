package domain

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
	LanguageSchemaName = "essential.language"

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
		Label(model.LangJson{model.LanguageCodeEnUs: "Language"}).
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
				DataType(dmodel.FieldDataTypeString(5, 10, dmodel.FieldDataTypeStringOpts{
					Regex: regexp.MustCompile(`^[a-z]{2}_[A-Z]{2}$`),
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
