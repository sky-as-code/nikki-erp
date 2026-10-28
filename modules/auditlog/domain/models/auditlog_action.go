package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	c "github.com/sky-as-code/nikki-erp/modules/auditlog/constants"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	ActionSchemaName = c.ResourceAuditLogAction

	ActionFieldId               = basemodel.FieldId
	ActionFieldOrgId            = basemodel.FieldOrgId
	ActionFieldCreatedAt        = basemodel.FieldCreatedAt
	ActionFieldAction           = "action"
	ActionFieldResourceCode     = "resource_code"
	ActionFieldResourceId       = "resource_id"
	ActionFieldActorType        = "actor_type"
	ActionFieldActorId          = "actor_id"
	ActionFieldActorDisplayName = "actor_display_name"
	ActionFieldBefore           = "before"
	ActionFieldAfter            = "after"
	ActionFieldRequestId        = "request_id"
	ActionFieldSource           = "source"

	ActionEdgeActorUser = "actor_user"
)

//go:embed auditlog_action.json
var actionSchemaJson string

func ActionSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(actionSchemaJson)
}

// Action is one audited action. It is append-only: written by the AuditSink port, never updated.
type Action struct {
	basemodel.DynamicModelBase
}

func NewActionFrom(src dmodel.DynamicFields) *Action {
	return &Action{DynamicModelBase: basemodel.NewDynamicModel(src)}
}
