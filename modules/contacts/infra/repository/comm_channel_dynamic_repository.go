package repository

import (
	itCommChannel "github.com/sky-as-code/nikki-erp/modules/contacts/interfaces/commchannel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// NewCommChannelRepository is handed the composable default by the onion. The type exists so a bespoke
// query has a home; it adds nothing today.
func NewCommChannelRepository(base dynamicresource.CrudRepository) itCommChannel.CommChannelRepository {
	return &CommChannelRepositoryImpl{CrudRepository: base}
}

type CommChannelRepositoryImpl struct {
	dynamicresource.CrudRepository
}
