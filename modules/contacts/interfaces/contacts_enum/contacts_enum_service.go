package contacts_enum

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	itEnum "github.com/sky-as-code/nikki-erp/modules/core/enum/interfaces"
)

type ContactsEnumService interface {
	GetEnum(ctx crud.Context, typeEnum, valueEnum string, valErrs *ft.ValidationErrors) (*itEnum.GetEnumResult, error)
}
