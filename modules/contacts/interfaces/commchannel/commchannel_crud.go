// Package commchannel declares the layers of the contacts_comm_channel resource: the ways a party
// can be reached (phone, email, address, …).
package commchannel

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

type CommChannelRepository interface {
	composable.CrudRepository
}

type CommChannelDomainService interface {
	composable.CrudDomainService
}

type CommChannelApplicationService interface {
	composable.CrudApplicationService
}

type (
	CreateCommChannelCommand      = composable.CreateCommand
	UpdateCommChannelCommand      = composable.UpdateCommand
	DeleteCommChannelCommand      = composable.DeleteCommand
	SetCommChannelArchivedCommand = composable.SetArchivedCommand
	GetCommChannelByIdQuery       = composable.GetByIdQuery
	SearchCommChannelsQuery       = composable.SearchQuery
	CommChannelExistsQuery        = composable.ExistsQuery
)

type (
	CreateCommChannelResult      = composable.CreateResult
	UpdateCommChannelResult      = composable.MutateResult
	DeleteCommChannelResult      = composable.MutateResult
	SetCommChannelArchivedResult = composable.MutateResult
	GetCommChannelByIdResult     = composable.GetOneResult
	SearchCommChannelsResult     = composable.SearchResult
	CommChannelExistsResult      = composable.ExistsResult
)
