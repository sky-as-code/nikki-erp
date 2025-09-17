package party

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/contacts/domain"
)

func (this CreatePartyCommand) ToDomainModel() *domain.Party {
	party := &domain.Party{}
	model.MustCopy(this, party)
	return party
}

func (this UpdatePartyCommand) ToDomainModel() *domain.Party {
	party := &domain.Party{}
	model.MustCopy(this, party)
	return party
}
