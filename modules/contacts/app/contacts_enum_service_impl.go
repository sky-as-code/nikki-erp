package app

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/contacts/interfaces/contacts_enum"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	itEnum "github.com/sky-as-code/nikki-erp/modules/core/enum/interfaces"
)

type ContactsEnumServiceImpl struct {
	enumSvc itEnum.EnumService
}

func NewContactsEnumServiceImpl(enumSvc itEnum.EnumService) contacts_enum.ContactsEnumService {
	return &ContactsEnumServiceImpl{
		enumSvc: enumSvc,
	}
}

func (cts *ContactsEnumServiceImpl) GetEnum(ctx crud.Context, typeEnum, valueEnum string, valErrs *ft.ValidationErrors) (*itEnum.GetEnumResult, error) {
	query := itEnum.GetEnumQuery{
		EntityName: "party",
		Type:       &typeEnum,
		Value:      &valueEnum,
	}

	enum, err := cts.enumSvc.GetEnum(ctx, query)
	if err != nil {
		valErrs.Append("title", "failed to get title enum")
		return nil, err
	}

	return enum, nil
}

// func (cts *ContactsTitleServiceImpl) CreateTitleEnum(ctx crud.Context) (*itEnum.CreateEnumResult, error) {
// 	command := itEnum.CreateEnumCommand{
// 		EntityName: "party",
// 		Type:       "contacts_party_title",
// 		Label: model.LangJson{
// 			"vi-VN": "Ông",
// 			"en-US": "Mr.",
// 		},
// 	}

// 	return cts.enumSvc.CreateEnum(ctx, command)
// }
