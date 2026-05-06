package domain

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

type PartyType string

const (
	PartyTypeIndividual = PartyType("individual")
	PartyTypeCompany    = PartyType("company")
)

const (
	PartyResourceCode = "contacts_party"
	PartyAuthScope    = "org"

	PartyActionCreate      = "create"
	PartyActionDelete      = "delete"
	PartyActionUpdate      = "update"
	PartyActionView        = "view"
	PartyActionSetArchived = "set_archived"
)

const (
	PartySchemaName = "contacts.party"

	PartyFieldId            = basemodel.FieldId
	PartyFieldEtag          = basemodel.FieldEtag
	PartyFieldOrgId         = basemodel.FieldOrgId
	PartyFieldIsArchived    = basemodel.FieldIsArchived
	PartyFieldAvatarUrl     = "avatar_url"
	PartyFieldDisplayName   = "display_name"
	PartyFieldLegalName     = "legal_name"
	PartyFieldLegalAddress  = "legal_address"
	PartyFieldTaxId         = "tax_id"
	PartyFieldJobPosition   = "job_position"
	PartyFieldTitle         = "title"
	PartyFieldType          = "type"
	PartyFieldNote          = "note"
	PartyFieldNationalityId = "nationality_id"
	PartyFieldLanguageId    = "language_id"
	PartyFieldWebsite       = "website"

	PartyEdgeCommChannels          = "comm_channels"
	PartyEdgeRelationshipsAsSource = "relationships_as_source"
	PartyEdgeRelationshipsAsTarget = "relationships_as_target"
)

func PartySchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(PartySchemaName).
		Label(model.LangJson{model.LanguageCodeEnUs: "Party"}).
		TableName("contacts_parties").
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Extend(basemodel.OrgIdModelSchemaBuilder()).
		Field(
			dmodel.DefineField().
				Name(PartyFieldAvatarUrl).
				Label(model.LangJson{model.LanguageCodeEnUs: "Avatar URL"}).
				DataType(dmodel.FieldDataTypeUrl()),
		).
		Field(
			dmodel.DefineField().
				Name(PartyFieldDisplayName).
				Label(model.LangJson{model.LanguageCodeEnUs: "Display Name"}).
				DataType(dmodel.FieldDataTypeString(1, 50)).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(PartyFieldLegalName).
				Label(model.LangJson{model.LanguageCodeEnUs: "Legal Name"}).
				DataType(dmodel.FieldDataTypeString(0, 100)),
		).
		Field(
			dmodel.DefineField().
				Name(PartyFieldLegalAddress).
				Label(model.LangJson{model.LanguageCodeEnUs: "Legal Address"}).
				DataType(dmodel.FieldDataTypeString(0, model.MODEL_RULE_LONG_NAME_LENGTH)),
		).
		Field(
			dmodel.DefineField().
				Name(PartyFieldTaxId).
				Label(model.LangJson{model.LanguageCodeEnUs: "Tax ID"}).
				DataType(dmodel.FieldDataTypeString(0, model.MODEL_RULE_SHORT_NAME_LENGTH)).
				Unique(),
		).
		Field(
			dmodel.DefineField().
				Name(PartyFieldJobPosition).
				Label(model.LangJson{model.LanguageCodeEnUs: "Job Position"}).
				DataType(dmodel.FieldDataTypeString(0, model.MODEL_RULE_SHORT_NAME_LENGTH)),
		).
		Field(
			dmodel.DefineField().
				Name(PartyFieldTitle).
				Label(model.LangJson{model.LanguageCodeEnUs: "Title"}).
				DataType(dmodel.FieldDataTypeString(0, model.MODEL_RULE_TINY_NAME_LENGTH)),
		).
		Field(
			dmodel.DefineField().
				Name(PartyFieldType).
				Label(model.LangJson{model.LanguageCodeEnUs: "Type"}).
				DataType(dmodel.FieldDataTypeEnumString([]string{
					string(PartyTypeIndividual),
					string(PartyTypeCompany),
				})).
				Default(string(PartyTypeIndividual)).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(PartyFieldNote).
				Label(model.LangJson{model.LanguageCodeEnUs: "Note"}).
				DataType(dmodel.FieldDataTypeString(0, model.MODEL_RULE_DESC_LENGTH)),
		).
		Field(
			basemodel.DefineFieldId(PartyFieldNationalityId).
				Label(model.LangJson{model.LanguageCodeEnUs: "Nationality"}),
		).
		Field(
			basemodel.DefineFieldId(PartyFieldLanguageId).
				Label(model.LangJson{model.LanguageCodeEnUs: "Language"}),
		).
		Field(
			dmodel.DefineField().
				Name(PartyFieldWebsite).
				Label(model.LangJson{model.LanguageCodeEnUs: "Website"}).
				DataType(dmodel.FieldDataTypeUrl()).
				Unique(),
		).
		Extend(basemodel.ArchivableModelSchemaBuilder()).
		Extend(basemodel.VersionedModelSchemaBuilder()).
		Extend(basemodel.AuditableModelSchemaBuilder()).
		EdgeFrom(
			dmodel.Edge(PartyEdgeCommChannels).
				Label(model.LangJson{model.LanguageCodeEnUs: "Communication Channels"}).
				Existing(CommChannelSchemaName, CommChannelEdgeParty),
		).
		EdgeFrom(
			dmodel.Edge(PartyEdgeRelationshipsAsSource).
				Label(model.LangJson{model.LanguageCodeEnUs: "Relationships (as source)"}).
				Existing(RelationshipSchemaName, RelationshipEdgeSourceParty),
		).
		EdgeFrom(
			dmodel.Edge(PartyEdgeRelationshipsAsTarget).
				Label(model.LangJson{model.LanguageCodeEnUs: "Relationships (as target)"}).
				Existing(RelationshipSchemaName, RelationshipEdgeTargetParty),
		)
}

type Party struct {
	fields dmodel.DynamicFields
}

func NewParty() *Party {
	return &Party{fields: make(dmodel.DynamicFields)}
}

func NewPartyFrom(src dmodel.DynamicFields) *Party {
	return &Party{fields: src}
}

func (this Party) GetFieldData() dmodel.DynamicFields {
	return this.fields
}

func (this *Party) SetFieldData(data dmodel.DynamicFields) {
	this.fields = data
}

func (this Party) GetId() *model.Id {
	return this.fields.GetModelId(PartyFieldId)
}

func (this *Party) SetId(v *model.Id) {
	this.fields.SetModelId(PartyFieldId, v)
}

func (this Party) GetEtag() *model.Etag {
	return this.fields.GetEtag(PartyFieldEtag)
}

func (this *Party) SetEtag(v *model.Etag) {
	this.fields.SetEtag(PartyFieldEtag, v)
}

func (this Party) GetOrgId() *model.Id {
	return this.fields.GetModelId(PartyFieldOrgId)
}

func (this *Party) SetOrgId(v *model.Id) {
	this.fields.SetModelId(PartyFieldOrgId, v)
}

func (this Party) IsArchived() *bool {
	return this.fields.GetBool(PartyFieldIsArchived)
}

func (this *Party) SetIsArchived(v *bool) {
	this.fields.SetBool(PartyFieldIsArchived, v)
}

func (this Party) GetAvatarUrl() *string {
	return this.fields.GetString(PartyFieldAvatarUrl)
}

func (this *Party) SetAvatarUrl(v *string) {
	this.fields.SetString(PartyFieldAvatarUrl, v)
}

func (this Party) GetDisplayName() *string {
	return this.fields.GetString(PartyFieldDisplayName)
}

func (this *Party) SetDisplayName(v *string) {
	this.fields.SetString(PartyFieldDisplayName, v)
}

func (this Party) GetLegalName() *string {
	return this.fields.GetString(PartyFieldLegalName)
}

func (this *Party) SetLegalName(v *string) {
	this.fields.SetString(PartyFieldLegalName, v)
}

func (this Party) GetLegalAddress() *string {
	return this.fields.GetString(PartyFieldLegalAddress)
}

func (this *Party) SetLegalAddress(v *string) {
	this.fields.SetString(PartyFieldLegalAddress, v)
}

func (this Party) GetTaxId() *string {
	return this.fields.GetString(PartyFieldTaxId)
}

func (this *Party) SetTaxId(v *string) {
	this.fields.SetString(PartyFieldTaxId, v)
}

func (this Party) GetJobPosition() *string {
	return this.fields.GetString(PartyFieldJobPosition)
}

func (this *Party) SetJobPosition(v *string) {
	this.fields.SetString(PartyFieldJobPosition, v)
}

func (this Party) GetTitle() *string {
	return this.fields.GetString(PartyFieldTitle)
}

func (this *Party) SetTitle(v *string) {
	this.fields.SetString(PartyFieldTitle, v)
}

func (this Party) GetType() *string {
	return this.fields.GetString(PartyFieldType)
}

func (this *Party) SetType(v *string) {
	this.fields.SetString(PartyFieldType, v)
}

func (this Party) GetNote() *string {
	return this.fields.GetString(PartyFieldNote)
}

func (this *Party) SetNote(v *string) {
	this.fields.SetString(PartyFieldNote, v)
}

func (this Party) GetNationalityId() *model.Id {
	return this.fields.GetModelId(PartyFieldNationalityId)
}

func (this *Party) SetNationalityId(v *model.Id) {
	this.fields.SetModelId(PartyFieldNationalityId, v)
}

func (this Party) GetLanguageId() *model.Id {
	return this.fields.GetModelId(PartyFieldLanguageId)
}

func (this *Party) SetLanguageId(v *model.Id) {
	this.fields.SetModelId(PartyFieldLanguageId, v)
}

func (this Party) GetWebsite() *string {
	return this.fields.GetString(PartyFieldWebsite)
}

func (this *Party) SetWebsite(v *string) {
	this.fields.SetString(PartyFieldWebsite, v)
}
