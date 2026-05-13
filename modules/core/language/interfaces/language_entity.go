package interfaces

import (
	"github.com/sky-as-code/nikki-erp/common/model"
)

type Language struct {
	Id                 *model.Id           `json:"id"`
	Name               *string             `json:"name"`
	IsoCode            *model.LanguageCode `json:"iso_code"`
	Direction          *string             `json:"direction"`
	DecimalSeparator   *string             `json:"decimal_separator"`
	ThousandsSeparator *string             `json:"thousands_separator"`
	DateFormat         *string             `json:"date_format"`
	TimeFormat         *string             `json:"time_format"`
	ShortTimeFormat    *string             `json:"short_time_format"`
	FirstDayOfWeek     *string             `json:"first_day_of_week"`
}
