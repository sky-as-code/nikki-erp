package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// The iam_service schema is a typed view over iam_users (shares_table: it emits no DDL): the rows
// of type service, with the human-only fields (email, avatar, org unit, owner flag) left out. Its
// domain service reads and writes through the user repository and forces type = service.
const (
	ServiceSchemaName = "iam_service"

	ServiceFieldId          = basemodel.FieldId
	ServiceFieldDisplayName = UserFieldDisplayName
	ServiceFieldStatus      = UserFieldStatus

	ServiceEdgeOrgs  = UserEdgeOrgs
	ServiceEdgeRoles = UserEdgeRoles
)

//go:embed service.json
var serviceSchemaJson string

func ServiceSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(serviceSchemaJson)
}
