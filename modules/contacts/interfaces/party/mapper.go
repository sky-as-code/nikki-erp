package party

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/contacts/domain"
)

func (this CreatePartyCommand) ToParty() *domain.Party {
	party := &domain.Party{}
	model.MustCopy(this, party)
	return party
}

func (this CreatePartyCommand) ToEntity() *domain.Party {
	return this.ToParty()
}

func (this UpdatePartyCommand) ToParty() *domain.Party {
	party := &domain.Party{}
	model.MustCopy(this, party)
	return party
}

func (this UpdatePartyCommand) ToEntity() *domain.Party {
	return this.ToParty()
}
