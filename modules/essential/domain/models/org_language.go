package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	OrgLanguageSchemaName = "essential_org_language"

	OrgLanguageFieldOrgId      = basemodel.FieldOrgId
	OrgLanguageFieldLanguageId = "language_id"
	OrgLanguageFieldIsActive   = "is_active"
)

//go:embed org_language.json
var orgLanguageSchemaJson string

func OrgLanguageSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(orgLanguageSchemaJson)
}

// OrgLanguage is one organization's activation of a tenant-wide language. The catalogue entry is shared by
// every organization of the tenant; whether an organization uses it is this row, so one
// organization switching a language off leaves the others untouched.
type OrgLanguage struct {
	basemodel.DynamicModelBase
}

func NewOrgLanguageFrom(src dmodel.DynamicFields) *OrgLanguage {
	return &OrgLanguage{basemodel.NewDynamicModel(src)}
}
