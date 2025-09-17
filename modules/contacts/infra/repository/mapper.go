package repository

import (
	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/contacts/domain"
	"github.com/sky-as-code/nikki-erp/modules/contacts/infra/ent"
)

// entToCommChannel converts ENT CommChannel entity to domain CommChannel
func entToCommChannel(entCommChannel *ent.CommChannel) *domain.CommChannel {
	commChannel := &domain.CommChannel{}
	model.MustCopy(entCommChannel, commChannel)

	// Handle Party relation if loaded
	// if entCommChannel.Edges.Party != nil {
	// 	commChannel.Party = entToParty(entCommChannel.Edges.Party)
	// }
	commChannel.ValueJson = &entCommChannel.ValueJSON

	return commChannel
}

// entToCommChannels converts slice of ENT CommChannel entities to domain CommChannels
func entToCommChannels(entCommChannels []*ent.CommChannel) []*domain.CommChannel {
	if entCommChannels == nil {
		return nil
	}
	return array.Map(entCommChannels, func(entCommChannel *ent.CommChannel) *domain.CommChannel {
		return entToCommChannel(entCommChannel)
	})
}

// entToCommChannelsNonPtr converts slice of ENT CommChannel entities to non-pointer domain CommChannels
func entToCommChannelsNonPtr(entCommChannels []*ent.CommChannel) []domain.CommChannel {
	if entCommChannels == nil {
		return nil
	}
	return array.Map(entCommChannels, func(entCommChannel *ent.CommChannel) domain.CommChannel {
		return *entToCommChannel(entCommChannel)
	})
}

// entToParty converts ENT Party entity to domain Party
func entToParty(entParty *ent.Party) *domain.Party {
	party := &domain.Party{}
	model.MustCopy(entParty, party)

	// Handle CommChannels relation if loaded
	if entParty.Edges.CommChannels != nil {
		party.CommChannels = array.Map(entParty.Edges.CommChannels, func(entCommChannel *ent.CommChannel) domain.CommChannel {
			return *entToCommChannel(entCommChannel)
		})
	}

	// Handle Relationships relation if loaded
	if entParty.Edges.RelationshipsAsSource != nil {
		party.Relationships = array.Map(entParty.Edges.RelationshipsAsSource, func(entRelationship *ent.Relationship) domain.Relationship {
			return *entToRelationship(entRelationship)
		})
	}

	return party
}

// entToParties converts slice of ENT Party entities to domain Parties
func entToParties(entParties []*ent.Party) []*domain.Party {
	if entParties == nil {
		return nil
	}
	return array.Map(entParties, func(entParty *ent.Party) *domain.Party {
		return entToParty(entParty)
	})
}

// entToPartiesNonPtr converts slice of ENT Party entities to non-pointer domain Parties
func entToPartiesNonPtr(entParties []*ent.Party) []domain.Party {
	if entParties == nil {
		return nil
	}
	return array.Map(entParties, func(entParty *ent.Party) domain.Party {
		return *entToParty(entParty)
	})
}

// entToRelationship converts ENT Relationship entity to domain Relationship
func entToRelationship(entRelationship *ent.Relationship) *domain.Relationship {
	relationship := &domain.Relationship{}
	model.MustCopy(entRelationship, relationship)

	// Note: Relationship entity doesn't have a Party field in domain
	// The target party reference is handled by TargetPartyId field

	return relationship
}

// entToRelationships converts slice of ENT Relationship entities to domain Relationships
func entToRelationships(entRelationships []*ent.Relationship) []*domain.Relationship {
	if entRelationships == nil {
		return nil
	}
	return array.Map(entRelationships, func(entRelationship *ent.Relationship) *domain.Relationship {
		return entToRelationship(entRelationship)
	})
}

// entToRelationshipsNonPtr converts slice of ENT Relationship entities to non-pointer domain Relationships
func entToRelationshipsNonPtr(entRelationships []*ent.Relationship) []domain.Relationship {
	if entRelationships == nil {
		return nil
	}
	return array.Map(entRelationships, func(entRelationship *ent.Relationship) domain.Relationship {
		return *entToRelationship(entRelationship)
	})
}
