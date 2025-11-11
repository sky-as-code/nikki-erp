package app

import (
	"github.com/sky-as-code/nikki-erp/common/defense"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/contacts/domain"
	itContactEnum "github.com/sky-as-code/nikki-erp/modules/contacts/interfaces/contactenum"
	itParty "github.com/sky-as-code/nikki-erp/modules/contacts/interfaces/party"
	pt "github.com/sky-as-code/nikki-erp/modules/contacts/interfaces/party"
	itRelationship "github.com/sky-as-code/nikki-erp/modules/contacts/interfaces/relationship"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

func NewRelationshipServiceImpl(
	relationshipRepo itRelationship.RelationshipRepository,
	contactsEnumServiceImpl itContactEnum.ContactsEnumService,
	partySvc itParty.PartyService,
	cqrsBus cqrs.CqrsBus,
) itRelationship.RelationshipService {
	return &RelationshipServiceImpl{
		relationshipRepo:        relationshipRepo,
		partySvc:                partySvc,
		cqrsBus:                 cqrsBus,
		contactsEnumServiceImpl: contactsEnumServiceImpl,
	}
}

type RelationshipServiceImpl struct {
	relationshipRepo        itRelationship.RelationshipRepository
	partySvc                itParty.PartyService
	cqrsBus                 cqrs.CqrsBus
	contactsEnumServiceImpl itContactEnum.ContactsEnumService
}

func (this *RelationshipServiceImpl) CreateRelationship(ctx crud.Context, cmd itRelationship.CreateRelationshipCommand) (*itRelationship.CreateRelationshipResult, error) {
	result, err := crud.Create(ctx, crud.CreateParam[*domain.Relationship, itRelationship.CreateRelationshipCommand, itRelationship.CreateRelationshipResult]{
		Action:              "create relationship",
		Command:             cmd,
		AssertBusinessRules: this.assertCreateRules,
		RepoCreate:          this.relationshipRepo.Create,
		SetDefault:          this.setRelationshipDefaults,
		Sanitize:            this.sanitizeRelationship,
		ToFailureResult: func(vErrs *ft.ValidationErrors) *itRelationship.CreateRelationshipResult {
			return &itRelationship.CreateRelationshipResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.Relationship) *itRelationship.CreateRelationshipResult {
			return &itRelationship.CreateRelationshipResult{
				HasData: true,
				Data:    model,
			}
		},
	})

	return result, err
}

func (this *RelationshipServiceImpl) assertCreateRules(ctx crud.Context, relationship *domain.Relationship, vErrs *ft.ValidationErrors) error {
	err := this.assertPartyExists(ctx, *relationship.PartyId, vErrs) // check party exists
	if err != nil {
		return err
	}

	err = this.assertPartyExists(ctx, *relationship.TargetPartyId, vErrs) // check party exists
	if err != nil {
		return err
	}

	err = this.assertEnumRelationship(ctx, "contacts_party_relationship_type", *relationship.Type, vErrs) // check relationship type exists
	if err != nil {
		return err
	}

	return nil
}

func (this *RelationshipServiceImpl) sanitizeRelationship(relationship *domain.Relationship) {
	if relationship.Note != nil {
		cleaned := defense.SanitizePlainText(*relationship.Note, true)
		relationship.Note = &cleaned
	}
}

func (this *RelationshipServiceImpl) assertPartyExists(ctx crud.Context, partyId model.Id, vErrs *ft.ValidationErrors) error {
	dbParty, err := this.partySvc.GetPartyById(ctx, pt.FindByIdParam{
		Id: partyId,
	})
	if err != nil {
		return err
	}

	if dbParty.Data == nil {
		vErrs.Append("partyId", "party or target party not found")
	}
	return nil
}

func (this *RelationshipServiceImpl) assertEnumRelationship(ctx crud.Context, typeEnum, valueEnum string, vErrs *ft.ValidationErrors) error {
	enum, err := this.contactsEnumServiceImpl.GetEnum(ctx, typeEnum, valueEnum, vErrs)
	if err != nil {
		return err
	}

	if enum.Data == nil {
		vErrs.Append("type", "type does not exist")
		return nil
	}
	return nil
}

func (this *RelationshipServiceImpl) setRelationshipDefaults(relationship *domain.Relationship) {
	relationship.SetDefaults()
}

// func (ps *PartyServiceImpl) assertTypeExists(ctx crud.Context, entityName, typeEnum string, valErrs *ft.ValidationErrors) error {

// 	existCmd := &eif.GetEnumQuery{
// 		EntityName: entityName + " title",
// 		Type:       &typeEnum,
// 		Value:      &valueEnum,
// 	}

// 	enum, err := ps.coreSvc.GetEnum(ctx, *existCmd)
// 	ft.PanicOnErr(err)
// 	if enum.ClientError != nil {
// 		valErrs.MergeClientError(enum.ClientError)
// 		return nil
// 	}

// 	if enum.Data == nil {
// 		valErrs.Append("title", "title not found")
// 	}

// 	return nil
// }
