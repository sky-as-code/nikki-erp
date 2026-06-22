package domain

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

type CommChannelType string

const (
	CommChannelTypePhone    = CommChannelType("phone")
	CommChannelTypeZalo     = CommChannelType("zalo")
	CommChannelTypeFacebook = CommChannelType("facebook")
	CommChannelTypeEmail    = CommChannelType("email")
	CommChannelTypePost     = CommChannelType("post")
)

const (
	CommChannelResourceCode = "contacts_comm_channel"
	CommChannelAuthScope    = "org"

	CommChannelActionCreate      = "create"
	CommChannelActionDelete      = "delete"
	CommChannelActionUpdate      = "update"
	CommChannelActionView        = "view"
	CommChannelActionSetArchived = "set_archived"
)

const (
	CommChannelSchemaName = "contacts_comm_channel"

	CommChannelFieldId         = basemodel.FieldId
	CommChannelFieldEtag       = basemodel.FieldEtag
	CommChannelFieldOrgId      = basemodel.FieldOrgId
	CommChannelFieldIsArchived = basemodel.FieldIsArchived
	CommChannelFieldNote       = "note"
	CommChannelFieldPartyId    = "party_id"
	CommChannelFieldType       = "type"
	CommChannelFieldValue      = "value"
	CommChannelFieldValueJson  = "value_json"

	CommChannelEdgeParty = "party"
)

func CommChannelSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(CommChannelSchemaName).
		Label(model.NewLangJsonRefSf("%s.label", CommChannelSchemaName)).
		TableName("contacts_comm_channels").
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Extend(basemodel.OrgIdModelSchemaBuilder()).
		Field(
			dmodel.DefineField().
				Name(CommChannelFieldNote).
				Label(model.NewLangJsonRefSf("fields.%s", CommChannelFieldNote)).
				DataType(dmodel.FieldDataTypeString(0, model.MODEL_RULE_DESC_LENGTH)),
		).
		Field(
			basemodel.DefineFieldId(CommChannelFieldPartyId).
				Label(model.NewLangJsonRefSf("fields.%s", CommChannelFieldPartyId)).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(CommChannelFieldType).
				Label(model.NewLangJsonRefSf("fields.%s", CommChannelFieldType)).
				DataType(dmodel.FieldDataTypeEnumString([]string{
					string(CommChannelTypePhone),
					string(CommChannelTypeZalo),
					string(CommChannelTypeFacebook),
					string(CommChannelTypeEmail),
					string(CommChannelTypePost),
				})).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(CommChannelFieldValue).
				Label(model.NewLangJsonRefSf("fields.%s", CommChannelFieldValue)).
				DataType(dmodel.FieldDataTypeString(0, 255)),
		).
		Field(
			dmodel.DefineField().
				Name(CommChannelFieldValueJson).
				Label(model.NewLangJsonRefSf("fields.%s", CommChannelFieldValueJson)).
				DataType(dmodel.FieldDataTypeJsonMap()),
		).
		Extend(basemodel.ArchivableModelSchemaBuilder()).
		Extend(basemodel.VersionedModelSchemaBuilder()).
		Extend(basemodel.AuditableModelSchemaBuilder()).
		EdgeTo(
			dmodel.Edge(CommChannelEdgeParty).
				Label(model.LangJson{model.LanguageCodeEnUs: "Party"}).
				ManyToOne(PartySchemaName, dmodel.DynamicFields{
					CommChannelFieldPartyId: basemodel.FieldId,
				}).
				OnDelete(dmodel.RelationCascadeCascade),
		)
}

type CommChannel struct {
	fields dmodel.DynamicFields
}

func NewCommChannel() *CommChannel {
	return &CommChannel{fields: make(dmodel.DynamicFields)}
}

func NewCommChannelFrom(src dmodel.DynamicFields) *CommChannel {
	return &CommChannel{fields: src}
}

func (this CommChannel) GetFieldData() dmodel.DynamicFields {
	return this.fields
}

func (this *CommChannel) SetFieldData(data dmodel.DynamicFields) {
	this.fields = data
}

func (this CommChannel) GetId() *model.Id {
	return this.fields.GetModelId(CommChannelFieldId)
}

func (this *CommChannel) SetId(v *model.Id) {
	this.fields.SetModelId(CommChannelFieldId, v)
}

func (this CommChannel) GetEtag() *model.Etag {
	return this.fields.GetEtag(CommChannelFieldEtag)
}

func (this *CommChannel) SetEtag(v *model.Etag) {
	this.fields.SetEtag(CommChannelFieldEtag, v)
}

func (this CommChannel) GetOrgId() *model.Id {
	return this.fields.GetModelId(CommChannelFieldOrgId)
}

func (this *CommChannel) SetOrgId(v *model.Id) {
	this.fields.SetModelId(CommChannelFieldOrgId, v)
}

func (this CommChannel) IsArchived() *bool {
	return this.fields.GetBool(CommChannelFieldIsArchived)
}

func (this *CommChannel) SetIsArchived(v *bool) {
	this.fields.SetBool(CommChannelFieldIsArchived, v)
}

func (this CommChannel) GetNote() *string {
	return this.fields.GetString(CommChannelFieldNote)
}

func (this *CommChannel) SetNote(v *string) {
	this.fields.SetString(CommChannelFieldNote, v)
}

func (this CommChannel) GetPartyId() *model.Id {
	return this.fields.GetModelId(CommChannelFieldPartyId)
}

func (this *CommChannel) SetPartyId(v *model.Id) {
	this.fields.SetModelId(CommChannelFieldPartyId, v)
}

func (this CommChannel) GetType() *string {
	return this.fields.GetString(CommChannelFieldType)
}

func (this *CommChannel) SetType(v *string) {
	this.fields.SetString(CommChannelFieldType, v)
}

func (this CommChannel) GetValue() *string {
	return this.fields.GetString(CommChannelFieldValue)
}

func (this *CommChannel) SetValue(v *string) {
	this.fields.SetString(CommChannelFieldValue, v)
}

func (this CommChannel) GetValueJson() any {
	return this.fields.GetAny(CommChannelFieldValueJson)
}

func (this *CommChannel) SetValueJson(v any) {
	this.fields.SetAny(CommChannelFieldValueJson, v)
}
