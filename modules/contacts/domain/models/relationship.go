package models

import (
	_ "embed"

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

// Relationship is a directed link between two parties: A employs B, A is the parent of B.
//
// The direction is meaningful, so party_id and target_party_id are not interchangeable — "A employs
// B" and "B employs A" are different facts, and reversing them silently inverts an org chart.
//
// Unlike its two sibling resources this one carries **no org_id**, and that asymmetry is
// deliberately preserved through the conversion rather than tidied up. It is load-bearing: adding
// org_id would change how every relationship query filters, and a relationship spanning two
// organizations' parties has no single organization to be scoped to. Changing it is a decision with
// a migration behind it, not a consistency clean-up. The API tests pin it in both directions.
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

//go:embed relationship.json
var relationshipSchemaJson string

func RelationshipSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(relationshipSchemaJson)
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
