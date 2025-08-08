package relationship

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/contacts/domain"
)

func (this CreateRelationshipCommand) ToRelationship() *domain.Relationship {
	relationship := &domain.Relationship{}
	model.MustCopy(this, relationship)
	return relationship
}

func (this CreateRelationshipCommand) ToEntity() *domain.Relationship {
	return this.ToRelationship()
}

func (this UpdateRelationshipCommand) ToRelationship() *domain.Relationship {
	relationship := &domain.Relationship{}
	model.MustCopy(this, relationship)
	return relationship
}

func (this UpdateRelationshipCommand) ToEntity() *domain.Relationship {
	return this.ToRelationship()
}
