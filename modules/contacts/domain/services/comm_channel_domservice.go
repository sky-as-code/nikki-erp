package services

import (
	itCommChannel "github.com/sky-as-code/nikki-erp/modules/contacts/interfaces/commchannel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// NewCommChannelDomainService is handed the composable default by the onion. Plain CRUD: the type
// exists so a rule has a home the day one is needed.
func NewCommChannelDomainService(base composable.CrudDomainService) itCommChannel.CommChannelDomainService {
	return &CommChannelDomainServiceImpl{CrudDomainService: base}
}

type CommChannelDomainServiceImpl struct {
	composable.CrudDomainService
}
