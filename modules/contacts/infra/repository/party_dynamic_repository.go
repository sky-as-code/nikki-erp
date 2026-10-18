package repository

import (
	itParty "github.com/sky-as-code/nikki-erp/modules/contacts/interfaces/party"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// NewPartyRepository is handed the composable default by the onion. The type exists so a bespoke
// query has a home; it adds nothing today.
func NewPartyRepository(base composable.CrudRepository) itParty.PartyRepository {
	return &PartyRepositoryImpl{CrudRepository: base}
}

type PartyRepositoryImpl struct {
	composable.CrudRepository
}
