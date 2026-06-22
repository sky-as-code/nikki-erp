package domain

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

type RelationshipType string

const (
	RelationshipTypeEmployee   = RelationshipType("employee")
	RelationshipTypeSpouse     = RelationshipType("spouse")
	RelationshipTypeParent     = RelationshipType("parent")
	RelationshipTypeSibling    = RelationshipType("sibling")
	RelationshipTypeEmergency  = RelationshipType("emergency")
	RelationshipTypeSubsidiary = RelationshipType("subsidiary")
)

const (
	RelationshipResourceCode = "contacts_relationship"
	RelationshipAuthScope    = "org"

	RelationshipActionCreate      = "create"
	RelationshipActionDelete      = "delete"
	RelationshipActionUpdate      = "update"
	RelationshipActionView        = "view"
	RelationshipActionSetArchived = "set_archived"
)

const (
	RelationshipSchemaName = "contacts_relationship"

	RelationshipFieldId            = basemodel.FieldId
	RelationshipFieldEtag          = basemodel.FieldEtag
	RelationshipFieldIsArchived    = basemodel.FieldIsArchived
	RelationshipFieldPartyId       = "party_id"
	RelationshipFieldTargetPartyId = "target_party_id"
	RelationshipFieldType          = "type"
	RelationshipFieldNote          = "note"

	RelationshipEdgeSourceParty = "source_party"
	RelationshipEdgeTargetParty = "target_party"
)

func RelationshipSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(RelationshipSchemaName).
		Label(model.NewLangJsonRefSf("%s.label", RelationshipSchemaName)).
		TableName("contacts_relationships").
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Field(
			basemodel.DefineFieldId(RelationshipFieldPartyId).
				Label(model.NewLangJsonRefSf("fields.%s", RelationshipFieldPartyId)).
				RequiredForCreate(),
		).
		Field(
			basemodel.DefineFieldId(RelationshipFieldTargetPartyId).
				Label(model.NewLangJsonRefSf("fields.%s", RelationshipFieldTargetPartyId)).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(RelationshipFieldType).
				Label(model.NewLangJsonRefSf("fields.%s", RelationshipFieldType)).
				DataType(dmodel.FieldDataTypeEnumString([]string{
					string(RelationshipTypeEmployee),
					string(RelationshipTypeSpouse),
					string(RelationshipTypeParent),
					string(RelationshipTypeSibling),
					string(RelationshipTypeEmergency),
					string(RelationshipTypeSubsidiary),
				})).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(RelationshipFieldNote).
				Label(model.NewLangJsonRefSf("fields.%s", RelationshipFieldNote)).
				DataType(dmodel.FieldDataTypeString(0, model.MODEL_RULE_DESC_LENGTH)),
		).
		Extend(basemodel.ArchivableModelSchemaBuilder()).
		Extend(basemodel.VersionedModelSchemaBuilder()).
		Extend(basemodel.AuditableModelSchemaBuilder()).
		EdgeTo(
			dmodel.Edge(RelationshipEdgeSourceParty).
				Label(model.LangJson{model.LanguageCodeEnUs: "Source Party"}).
				ManyToOne(PartySchemaName, dmodel.DynamicFields{
					RelationshipFieldPartyId: basemodel.FieldId,
				}).
				OnDelete(dmodel.RelationCascadeCascade),
		).
		EdgeTo(
			dmodel.Edge(RelationshipEdgeTargetParty).
				Label(model.LangJson{model.LanguageCodeEnUs: "Target Party"}).
				ManyToOne(PartySchemaName, dmodel.DynamicFields{
					RelationshipFieldTargetPartyId: basemodel.FieldId,
				}).
				OnDelete(dmodel.RelationCascadeCascade),
		)
}

type Relationship struct {
	fields dmodel.DynamicFields
}

func NewRelationship() *Relationship {
	return &Relationship{fields: make(dmodel.DynamicFields)}
}

func NewRelationshipFrom(src dmodel.DynamicFields) *Relationship {
	return &Relationship{fields: src}
}

func (this Relationship) GetFieldData() dmodel.DynamicFields {
	return this.fields
}

func (this *Relationship) SetFieldData(data dmodel.DynamicFields) {
	this.fields = data
}

func (this Relationship) GetId() *model.Id {
	return this.fields.GetModelId(RelationshipFieldId)
}

func (this *Relationship) SetId(v *model.Id) {
	this.fields.SetModelId(RelationshipFieldId, v)
}

func (this Relationship) GetEtag() *model.Etag {
	return this.fields.GetEtag(RelationshipFieldEtag)
}

func (this *Relationship) SetEtag(v *model.Etag) {
	this.fields.SetEtag(RelationshipFieldEtag, v)
}

func (this Relationship) IsArchived() *bool {
	return this.fields.GetBool(RelationshipFieldIsArchived)
}

func (this *Relationship) SetIsArchived(v *bool) {
	this.fields.SetBool(RelationshipFieldIsArchived, v)
}

func (this Relationship) GetPartyId() *model.Id {
	return this.fields.GetModelId(RelationshipFieldPartyId)
}

func (this *Relationship) SetPartyId(v *model.Id) {
	this.fields.SetModelId(RelationshipFieldPartyId, v)
}

func (this Relationship) GetTargetPartyId() *model.Id {
	return this.fields.GetModelId(RelationshipFieldTargetPartyId)
}

func (this *Relationship) SetTargetPartyId(v *model.Id) {
	this.fields.SetModelId(RelationshipFieldTargetPartyId, v)
}

func (this Relationship) GetType() *string {
	return this.fields.GetString(RelationshipFieldType)
}

func (this *Relationship) SetType(v *string) {
	this.fields.SetString(RelationshipFieldType, v)
}

func (this Relationship) GetNote() *string {
	return this.fields.GetString(RelationshipFieldNote)
}

func (this *Relationship) SetNote(v *string) {
	this.fields.SetString(RelationshipFieldNote, v)
}
