package relationship

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/contacts/domain"
)

func (this CreateRelationshipCommand) ToDomainModel() *domain.Relationship {
	relationship := &domain.Relationship{}
	model.MustCopy(this, relationship)
	return relationship
}

func (this UpdateRelationshipCommand) ToDomainModel() *domain.Relationship {
	relationship := &domain.Relationship{}
	model.MustCopy(this, relationship)
	return relationship
}
