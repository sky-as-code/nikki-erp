package escalationrule

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/helpdesk/domain"
)

func init() {
	var req cqrs.Request
	req = (*CreateEscalationRuleCommand)(nil)
	req = (*DeleteEscalationRuleCommand)(nil)
	req = (*GetEscalationRuleQuery)(nil)
	req = (*EscalationRuleExistsQuery)(nil)
	req = (*SearchEscalationRulesQuery)(nil)
	req = (*UpdateEscalationRuleCommand)(nil)
	util.Unused(req)
}

var createEscalationRuleCommandType = cqrs.RequestType{Module: "helpdesk", Submodule: "escalationrule", Action: "createEscalationRule"}

type CreateEscalationRuleCommand struct{ domain.EscalationRule }

func (CreateEscalationRuleCommand) CqrsRequestType() cqrs.RequestType {
	return createEscalationRuleCommandType
}
func (CreateEscalationRuleCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.EscalationRuleSchemaName)
}

type CreateEscalationRuleResult = dyn.OpResult[domain.EscalationRule]

var deleteEscalationRuleCommandType = cqrs.RequestType{Module: "helpdesk", Submodule: "escalationrule", Action: "deleteEscalationRule"}

type DeleteEscalationRuleCommand dyn.DeleteOneCommand

func (DeleteEscalationRuleCommand) CqrsRequestType() cqrs.RequestType {
	return deleteEscalationRuleCommandType
}

type DeleteEscalationRuleResult = dyn.OpResult[dyn.MutateResultData]

var getEscalationRuleQueryType = cqrs.RequestType{Module: "helpdesk", Submodule: "escalationrule", Action: "getEscalationRule"}

type GetEscalationRuleQuery dyn.GetOneQuery

func (GetEscalationRuleQuery) CqrsRequestType() cqrs.RequestType { return getEscalationRuleQueryType }

type GetEscalationRuleResult = dyn.OpResult[domain.EscalationRule]

var escalationRuleExistsQueryType = cqrs.RequestType{Module: "helpdesk", Submodule: "escalationrule", Action: "escalationRuleExists"}

type EscalationRuleExistsQuery dyn.ExistsQuery

func (EscalationRuleExistsQuery) CqrsRequestType() cqrs.RequestType {
	return escalationRuleExistsQueryType
}

type EscalationRuleExistsResult = dyn.OpResult[dyn.ExistsResultData]

var searchEscalationRulesQueryType = cqrs.RequestType{Module: "helpdesk", Submodule: "escalationrule", Action: "searchEscalationRules"}

type SearchEscalationRulesQuery dyn.SearchQuery

func (SearchEscalationRulesQuery) CqrsRequestType() cqrs.RequestType {
	return searchEscalationRulesQueryType
}

type SearchEscalationRulesResultData = dyn.PagedResultData[domain.EscalationRule]
type SearchEscalationRulesResult = dyn.OpResult[SearchEscalationRulesResultData]

var updateEscalationRuleCommandType = cqrs.RequestType{Module: "helpdesk", Submodule: "escalationrule", Action: "updateEscalationRule"}

type UpdateEscalationRuleCommand struct{ domain.EscalationRule }

func (UpdateEscalationRuleCommand) CqrsRequestType() cqrs.RequestType {
	return updateEscalationRuleCommandType
}
func (UpdateEscalationRuleCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.EscalationRuleSchemaName)
}

type UpdateEscalationRuleResult = dyn.OpResult[dyn.MutateResultData]
