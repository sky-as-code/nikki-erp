// Package commchannel declares the layers of the contacts_comm_channel resource: the ways a party
// can be reached (phone, email, address, …).
package commchannel

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

type CommChannelRepository interface {
	dynamicresource.CrudRepository
}

type CommChannelDomainService interface {
	dynamicresource.CrudDomainService
}

type CommChannelApplicationService interface {
	dynamicresource.CrudApplicationService
}

type (
	CreateCommChannelCommand      = dynamicresource.CreateCommand
	UpdateCommChannelCommand      = dynamicresource.UpdateCommand
	DeleteCommChannelCommand      = dynamicresource.DeleteCommand
	SetCommChannelArchivedCommand = dynamicresource.SetArchivedCommand
	GetCommChannelByIdQuery       = dynamicresource.GetByIdQuery
	SearchCommChannelsQuery       = dynamicresource.SearchQuery
	CommChannelExistsQuery        = dynamicresource.ExistsQuery
)

type (
	CreateCommChannelResult      = dynamicresource.CreateResult
	UpdateCommChannelResult      = dynamicresource.MutateResult
	DeleteCommChannelResult      = dynamicresource.MutateResult
	SetCommChannelArchivedResult = dynamicresource.MutateResult
	GetCommChannelByIdResult     = dynamicresource.GetOneResult
	SearchCommChannelsResult     = dynamicresource.SearchResult
	CommChannelExistsResult      = dynamicresource.ExistsResult
)
