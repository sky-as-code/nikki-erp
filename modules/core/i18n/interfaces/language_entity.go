package interfaces

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	enum "github.com/sky-as-code/nikki-erp/modules/core/enum/interfaces"
)

type Language struct {
	Id                 *model.Id           `json:"id"`
	Name               *string             `json:"name"`
	Code               *model.LanguageCode `json:"code"`
	Direction          *enum.Enum          `json:"direction"`
	DecimalSeparator   *string             `json:"decimalSeparator"`
	ThousandsSeparator *string             `json:"thousandsSeparator"`
	DateFormat         *string             `json:"dateFormat"`
	TimeFormat         *string             `json:"timeFormat"`
	ShortTimeFormat    *string             `json:"shortTimeFormat"`
	FirstDayOfWeek     *enum.Enum          `json:"firstDayOfWeek"`
}
