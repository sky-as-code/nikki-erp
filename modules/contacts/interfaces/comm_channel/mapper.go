package comm_channel

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/contacts/domain"
)

func (this CreateCommChannelCommand) ToDomainModel() *domain.CommChannel {
	commChannel := &domain.CommChannel{}
	model.MustCopy(this, commChannel)
	return commChannel
}

func (this UpdateCommChannelCommand) ToDomainModel() *domain.CommChannel {
	commChannel := &domain.CommChannel{}
	model.MustCopy(this, commChannel)
	return commChannel
}
