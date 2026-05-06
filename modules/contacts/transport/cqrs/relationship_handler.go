package cqrs

import (
	"context"

	c "github.com/sky-as-code/nikki-erp/modules/contacts/constants"
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
	return cqrs.ServePacket(ctx, string(c.ContactsModuleName), packet, this.RelationshipSvc.CreateRelationship)
}

func (this *RelationshipHandler) UpdateRelationship(ctx context.Context, packet *cqrs.RequestPacket[relationship.UpdateRelationshipCommand]) (*cqrs.Reply[relationship.UpdateRelationshipResult], error) {
	return cqrs.ServePacket(ctx, string(c.ContactsModuleName), packet, this.RelationshipSvc.UpdateRelationship)
}

func (this *RelationshipHandler) DeleteRelationship(ctx context.Context, packet *cqrs.RequestPacket[relationship.DeleteRelationshipCommand]) (*cqrs.Reply[relationship.DeleteRelationshipResult], error) {
	return cqrs.ServePacket(ctx, string(c.ContactsModuleName), packet, this.RelationshipSvc.DeleteRelationship)
}

func (this *RelationshipHandler) GetRelationship(ctx context.Context, packet *cqrs.RequestPacket[relationship.GetRelationshipQuery]) (*cqrs.Reply[relationship.GetRelationshipResult], error) {
	return cqrs.ServePacket(ctx, string(c.ContactsModuleName), packet, this.RelationshipSvc.GetRelationship)
}

func (this *RelationshipHandler) SearchRelationships(ctx context.Context, packet *cqrs.RequestPacket[relationship.SearchRelationshipsQuery]) (*cqrs.Reply[relationship.SearchRelationshipsResult], error) {
	return cqrs.ServePacket(ctx, string(c.ContactsModuleName), packet, this.RelationshipSvc.SearchRelationships)
}

func (this *RelationshipHandler) RelationshipExists(ctx context.Context, packet *cqrs.RequestPacket[relationship.RelationshipExistsQuery]) (*cqrs.Reply[relationship.RelationshipExistsResult], error) {
	return cqrs.ServePacket(ctx, string(c.ContactsModuleName), packet, this.RelationshipSvc.RelationshipExists)
}
