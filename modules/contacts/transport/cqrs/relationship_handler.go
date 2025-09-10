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

func (this *RelationshipHandler) UpdateRelationship(ctx context.Context, packet *cqrs.RequestPacket[relationship.UpdateRelationshipCommand]) (*cqrs.Reply[relationship.UpdateRelationshipResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.RelationshipSvc.UpdateRelationship)
}

func (this *RelationshipHandler) DeleteRelationship(ctx context.Context, packet *cqrs.RequestPacket[relationship.DeleteRelationshipCommand]) (*cqrs.Reply[relationship.DeleteRelationshipResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.RelationshipSvc.DeleteRelationship)
}

func (this *RelationshipHandler) GetRelationshipById(ctx context.Context, packet *cqrs.RequestPacket[relationship.GetRelationshipByIdQuery]) (*cqrs.Reply[relationship.GetRelationshipByIdResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.RelationshipSvc.GetRelationshipById)
}

func (this *RelationshipHandler) SearchRelationships(ctx context.Context, packet *cqrs.RequestPacket[relationship.SearchRelationshipsQuery]) (*cqrs.Reply[relationship.SearchRelationshipsResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.RelationshipSvc.SearchRelationships)
}
