package app

import (
	itCommChannel "github.com/sky-as-code/nikki-erp/modules/contacts/interfaces/commchannel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// NewCommChannelApplicationService is handed the composable default by the onion: the authorized CRUD,
// nothing resource-specific yet.
func NewCommChannelApplicationService(
	base dynamicresource.CrudApplicationService,
) itCommChannel.CommChannelApplicationService {
	return &CommChannelApplicationServiceImpl{CrudApplicationService: base}
}

type CommChannelApplicationServiceImpl struct {
	dynamicresource.CrudApplicationService
}
