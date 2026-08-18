package models

import (
	_ "embed"

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

// CommChannel is one way of reaching a party: a phone number, an email address, a postal address.
//
// A party has many, and a channel has no meaning apart from its party — which is why the edge
// cascades on delete. Removing a contact removes the ways of reaching it; leaving the channels
// behind would strand a phone number pointing at nobody.
//
// The address is held in either value or value_json depending on the type. A postal address has a
// street, a city and a country and does not fit one string, while a phone number does; a single
// typed column cannot be all five channel types at once.
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

//go:embed comm_channel.json
var commChannelSchemaJson string

func CommChannelSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(commChannelSchemaJson)
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
