package domain

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	ContactSchemaName = "essential.contact"

	ContactFieldId            = basemodel.FieldId
	ContactFieldAvatarUrl     = "avatar_url"
	ContactFieldDisplayName   = "display_name"
	ContactFieldLegalName     = "legal_name"
	ContactFieldLegalAddress  = "legal_address"
	ContactFieldTaxId         = "tax_id"
	ContactFieldJobPosition   = "job_position"
	ContactFieldTitle         = "title"
	ContactFieldType          = "type"
	ContactFieldNote          = "note"
	ContactFieldNationalityId = "nationality_id"
	ContactFieldOrgId         = "org_id"
	ContactFieldLanguageId    = "language_id"
	ContactFieldWebsite       = "website"

	ContactTypeIndividual = "individual"
	ContactTypeCompany    = "company"
)

const (
	ContactChannelSchemaName = "essential.contact_channel"

	ContactChannelFieldId        = basemodel.FieldId
	ContactChannelFieldContactId = "contact_id"
	ContactChannelFieldOrgId     = "org_id"
	ContactChannelFieldType      = "type"
	ContactChannelFieldValue     = "value"
	ContactChannelFieldValueJson = "value_json"
	ContactChannelFieldNote      = "note"

	ContactChannelTypePhone    = "phone"
	ContactChannelTypeZalo     = "zalo"
	ContactChannelTypeFacebook = "facebook"
	ContactChannelTypeEmail    = "email"
	ContactChannelTypePost     = "post"
)

const (
	ContactRelationshipSchemaName = "essential.contact_relationship"

	ContactRelationshipFieldId              = basemodel.FieldId
	ContactRelationshipFieldContactId       = "contact_id"
	ContactRelationshipFieldTargetContactId = "target_contact_id"
	ContactRelationshipFieldType            = "type"
	ContactRelationshipFieldNote            = "note"
)

const (
	ContactEdgeCommChannels          = "comm_channels"
	ContactEdgeRelationshipsAsSource = "relationships_as_source"
	ContactEdgeRelationshipsAsTarget = "relationships_as_target"
	ContactChannelEdgeContact        = "contact"
	ContactRelationshipEdgeSource    = "source_contact"
	ContactRelationshipEdgeTarget    = "target_contact"
)

func ContactSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(ContactSchemaName).
		Label(model.LangJson{model.LanguageCodeEnUs: "Contact"}).
		TableName("essential_contacts").
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Field(
			dmodel.DefineField().
				Name(ContactFieldAvatarUrl).
				Label(model.LangJson{model.LanguageCodeEnUs: "Avatar URL"}).
				DataType(dmodel.FieldDataTypeUrl()),
		).
		Field(
			dmodel.DefineField().
				Name(ContactFieldDisplayName).
				Label(model.LangJson{model.LanguageCodeEnUs: "Display name"}).
				DataType(dmodel.FieldDataTypeString(1, 50)).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(ContactFieldLegalName).
				Label(model.LangJson{model.LanguageCodeEnUs: "Legal name"}).
				DataType(dmodel.FieldDataTypeString(0, 100)),
		).
		Field(
			dmodel.DefineField().
				Name(ContactFieldLegalAddress).
				Label(model.LangJson{model.LanguageCodeEnUs: "Legal address"}).
				DataType(dmodel.FieldDataTypeString(0, model.MODEL_RULE_DESC_LENGTH)),
		).
		Field(
			dmodel.DefineField().
				Name(ContactFieldTaxId).
				Label(model.LangJson{model.LanguageCodeEnUs: "Tax ID"}).
				DataType(dmodel.FieldDataTypeString(0, model.MODEL_RULE_TINY_NAME_LENGTH)),
		).
		Field(
			dmodel.DefineField().
				Name(ContactFieldJobPosition).
				Label(model.LangJson{model.LanguageCodeEnUs: "Job position"}).
				DataType(dmodel.FieldDataTypeString(0, model.MODEL_RULE_TINY_NAME_LENGTH)),
		).
		Field(
			dmodel.DefineField().
				Name(ContactFieldTitle).
				Label(model.LangJson{model.LanguageCodeEnUs: "Title"}).
				DataType(dmodel.FieldDataTypeString(0, model.MODEL_RULE_TINY_NAME_LENGTH)),
		).
		Field(
			dmodel.DefineField().
				Name(ContactFieldType).
				Label(model.LangJson{model.LanguageCodeEnUs: "Type"}).
				DataType(dmodel.FieldDataTypeEnumString([]string{ContactTypeIndividual, ContactTypeCompany})).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(ContactFieldNote).
				Label(model.LangJson{model.LanguageCodeEnUs: "Note"}).
				DataType(dmodel.FieldDataTypeString(0, model.MODEL_RULE_DESC_LENGTH)),
		).
		Field(
			basemodel.DefineFieldId(ContactFieldNationalityId).
				Label(model.LangJson{model.LanguageCodeEnUs: "Nationality"}),
		).
		Field(
			basemodel.DefineFieldId(ContactFieldOrgId).
				Label(model.LangJson{model.LanguageCodeEnUs: "Organization ID"}).
				RequiredForCreate(),
		).
		Field(
			basemodel.DefineFieldId(ContactFieldLanguageId).
				Label(model.LangJson{model.LanguageCodeEnUs: "Language"}),
		).
		Field(
			dmodel.DefineField().
				Name(ContactFieldWebsite).
				Label(model.LangJson{model.LanguageCodeEnUs: "Website"}).
				DataType(dmodel.FieldDataTypeUrl()),
		).
		Extend(basemodel.ArchivableModelSchemaBuilder()).
		Extend(basemodel.VersionedModelSchemaBuilder()).
		Extend(basemodel.AuditableModelSchemaBuilder()).
		EdgeTo(
			dmodel.Edge(ContactEdgeCommChannels).
				OneToMany(ContactChannelSchemaName, dmodel.DynamicFields{
					ContactChannelFieldContactId: ContactFieldId,
				}).
				OnDelete(dmodel.RelationCascadeCascade),
		).
		EdgeTo(
			dmodel.Edge(ContactEdgeRelationshipsAsSource).
				OneToMany(ContactRelationshipSchemaName, dmodel.DynamicFields{
					ContactRelationshipFieldContactId: ContactFieldId,
				}).
				OnDelete(dmodel.RelationCascadeCascade),
		).
		EdgeTo(
			dmodel.Edge(ContactEdgeRelationshipsAsTarget).
				OneToMany(ContactRelationshipSchemaName, dmodel.DynamicFields{
					ContactRelationshipFieldTargetContactId: ContactFieldId,
				}).
				OnDelete(dmodel.RelationCascadeCascade),
		)
}

func ContactChannelSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(ContactChannelSchemaName).
		Label(model.LangJson{model.LanguageCodeEnUs: "Contact communication channel"}).
		TableName("essential_contact_channels").
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Field(
			basemodel.DefineFieldId(ContactChannelFieldOrgId).
				RequiredForCreate(),
		).
		Field(
			basemodel.DefineFieldId(ContactChannelFieldContactId).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(ContactChannelFieldType).
				Label(model.LangJson{model.LanguageCodeEnUs: "Channel type"}).
				DataType(
					dmodel.FieldDataTypeEnumString([]string{
						ContactChannelTypePhone,
						ContactChannelTypeZalo,
						ContactChannelTypeFacebook,
						ContactChannelTypeEmail,
						ContactChannelTypePost,
					}),
				).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(ContactChannelFieldValue).
				Label(model.LangJson{model.LanguageCodeEnUs: "Value"}).
				DataType(dmodel.FieldDataTypeString(0, 255)),
		).
		Field(
			dmodel.DefineField().
				Name(ContactChannelFieldValueJson).
				Label(model.LangJson{model.LanguageCodeEnUs: "Localized value"}).
				DataType(dmodel.FieldDataTypeLangJson(0, model.MODEL_RULE_DESC_LENGTH)),
		).
		Field(
			dmodel.DefineField().
				Name(ContactChannelFieldNote).
				Label(model.LangJson{model.LanguageCodeEnUs: "Note"}).
				DataType(dmodel.FieldDataTypeString(0, model.MODEL_RULE_DESC_LENGTH)),
		).
		Extend(basemodel.ArchivableModelSchemaBuilder()).
		Extend(basemodel.VersionedModelSchemaBuilder()).
		Extend(basemodel.AuditableModelSchemaBuilder()).
		EdgeTo(
			dmodel.Edge(ContactChannelEdgeContact).
				ManyToOne(ContactSchemaName, dmodel.DynamicFields{
					ContactChannelFieldContactId: ContactFieldId,
				}).
				OnDelete(dmodel.RelationCascadeCascade),
		)
}

func ContactRelationshipSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(ContactRelationshipSchemaName).
		Label(model.LangJson{model.LanguageCodeEnUs: "Contact relationship"}).
		TableName("essential_contact_relationships").
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Field(
			basemodel.DefineFieldId(ContactRelationshipFieldContactId).
				RequiredForCreate(),
		).
		Field(
			basemodel.DefineFieldId(ContactRelationshipFieldTargetContactId).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(ContactRelationshipFieldType).
				Label(model.LangJson{model.LanguageCodeEnUs: "Relationship type"}).
				DataType(
					dmodel.FieldDataTypeEnumString([]string{
						"employee",
						"spouse",
						"parent",
						"sibling",
						"emergency",
						"subsidiary",
					}),
				).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(ContactRelationshipFieldNote).
				Label(model.LangJson{model.LanguageCodeEnUs: "Note"}).
				DataType(dmodel.FieldDataTypeString(0, model.MODEL_RULE_DESC_LENGTH)),
		).
		Extend(basemodel.ArchivableModelSchemaBuilder()).
		Extend(basemodel.VersionedModelSchemaBuilder()).
		Extend(basemodel.AuditableModelSchemaBuilder()).
		EdgeTo(
			dmodel.Edge(ContactRelationshipEdgeSource).
				ManyToOne(ContactSchemaName, dmodel.DynamicFields{
					ContactRelationshipFieldContactId: ContactFieldId,
				}),
		).
		EdgeTo(
			dmodel.Edge(ContactRelationshipEdgeTarget).
				ManyToOne(ContactSchemaName, dmodel.DynamicFields{
					ContactRelationshipFieldTargetContactId: ContactFieldId,
				}),
		)
}

type Contact struct {
	basemodel.DynamicModelBase
}

func NewContact() *Contact {
	return &Contact{basemodel.NewDynamicModel()}
}

func NewContactFrom(src dmodel.DynamicFields) *Contact {
	return &Contact{basemodel.NewDynamicModel(src)}
}

type ContactChannel struct {
	basemodel.DynamicModelBase
}

func NewContactChannel() *ContactChannel {
	return &ContactChannel{basemodel.NewDynamicModel()}
}

func NewContactChannelFrom(src dmodel.DynamicFields) *ContactChannel {
	return &ContactChannel{basemodel.NewDynamicModel(src)}
}

type ContactRelationship struct {
	basemodel.DynamicModelBase
}

func NewContactRelationship() *ContactRelationship {
	return &ContactRelationship{basemodel.NewDynamicModel()}
}

func NewContactRelationshipFrom(src dmodel.DynamicFields) *ContactRelationship {
	return &ContactRelationship{basemodel.NewDynamicModel(src)}
}
