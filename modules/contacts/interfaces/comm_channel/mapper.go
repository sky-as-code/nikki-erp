package comm_channel

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/contacts/domain"
)

func (this CreateCommChannelCommand) ToCommChannel() *domain.CommChannel {
	commChannel := &domain.CommChannel{}
	model.MustCopy(this, commChannel)
	return commChannel
}

func (this CreateCommChannelCommand) ToEntity() *domain.CommChannel {
	return this.ToCommChannel()
}

func (this UpdateCommChannelCommand) ToCommChannel() *domain.CommChannel {
	commChannel := &domain.CommChannel{}
	model.MustCopy(this, commChannel)
	return commChannel
}

func (this UpdateCommChannelCommand) ToEntity() *domain.CommChannel {
	return this.ToCommChannel()
}
