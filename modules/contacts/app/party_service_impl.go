package app

import (
	it "github.com/sky-as-code/nikki-erp/modules/contacts/interfaces/party"
)

func NewPartyServiceImpl(
	partyTagSvc it.PartyTagService,
) it.PartyService {
	return &PartyServiceImpl{
		PartyTagService: partyTagSvc,
	}
}

type PartyServiceImpl struct {
	it.PartyTagService
}
