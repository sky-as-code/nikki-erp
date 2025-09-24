package v1

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/contacts/domain"
	"github.com/sky-as-code/nikki-erp/modules/contacts/interfaces/relationship"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
)

type RelationshipDto struct {
	Id            string  `json:"id"`
	PartyId       string  `json:"partyId"`
	TargetPartyId string  `json:"targetPartyId"`
	Type          string  `json:"type"`
	Note          *string `json:"note,omitempty"`
}

func (this *RelationshipDto) FromRelationship(relationshipEntity domain.Relationship) {
	model.MustCopy(relationshipEntity.AuditableBase, this)
	model.MustCopy(relationshipEntity.ModelBase, this)
	model.MustCopy(relationshipEntity, this)
}

type CreateRelationshipRequest = relationship.CreateRelationshipCommand
type CreateRelationshipResponse = httpserver.RestCreateResponse
