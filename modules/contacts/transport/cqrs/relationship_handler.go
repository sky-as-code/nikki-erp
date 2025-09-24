package cqrs

import (
	"context"

	"github.com/sky-as-code/nikki-erp/modules/contacts/interfaces/relationship"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
)

func NewRelationshipHandler(relationshipSvc relationship.RelationshipService) *RelationshipHandler {
	return &RelationshipHandler{
		RelationshipSvc: relationshipSvc,
	}
}

type RelationshipHandler struct {
	RelationshipSvc relationship.RelationshipService
}

func (this *RelationshipHandler) CreateRelationship(ctx context.Context, packet *cqrs.RequestPacket[relationship.CreateRelationshipCommand]) (*cqrs.Reply[relationship.CreateRelationshipResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.RelationshipSvc.CreateRelationship)
}
