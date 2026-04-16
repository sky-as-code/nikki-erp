package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/helpdesk/interfaces/escalationrule"
)

type CreateEscalationRuleRequest = it.CreateEscalationRuleCommand
type CreateEscalationRuleResponse = httpserver.RestCreateResponse
type DeleteEscalationRuleRequest = it.DeleteEscalationRuleCommand
type DeleteEscalationRuleResponse = httpserver.RestDeleteResponse2
type GetEscalationRuleRequest = it.GetEscalationRuleQuery
type GetEscalationRuleResponse = dmodel.DynamicFields
type EscalationRuleExistsRequest = it.EscalationRuleExistsQuery
type EscalationRuleExistsResponse = dyn.ExistsResultData
type SearchEscalationRulesRequest = it.SearchEscalationRulesQuery
type SearchEscalationRulesResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]
type UpdateEscalationRuleRequest = it.UpdateEscalationRuleCommand
type UpdateEscalationRuleResponse = httpserver.RestMutateResponse
