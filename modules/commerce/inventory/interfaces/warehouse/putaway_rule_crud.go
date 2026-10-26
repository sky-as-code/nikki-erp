package warehouse

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// PutawayRuleRepository reads and writes putaway_rule rows.
type PutawayRuleRepository interface {
	composable.CrudRepository
}

// PutawayRuleDomainService is the CRUD of the resource plus its own rules.
type PutawayRuleDomainService interface {
	composable.CrudDomainService
}

// PutawayRuleApplicationService is the authorized surface the REST handler serves.
type PutawayRuleApplicationService interface {
	composable.CrudApplicationService
	PutawayRuleActionService
}

// The CRUD commands, queries and results of the resource, as composable shapes under a
// resource-specific name.
type (
	CreatePutawayRuleCommand      = composable.CreateCommand
	UpdatePutawayRuleCommand      = composable.UpdateCommand
	DeletePutawayRuleCommand      = composable.DeleteCommand
	SetPutawayRuleArchivedCommand = composable.SetArchivedCommand
	GetPutawayRuleByIdQuery       = composable.GetByIdQuery
	SearchPutawayRulesQuery       = composable.SearchQuery
	PutawayRuleExistsQuery        = composable.ExistsQuery
)

type (
	CreatePutawayRuleResult      = composable.CreateResult
	UpdatePutawayRuleResult      = composable.MutateResult
	DeletePutawayRuleResult      = composable.MutateResult
	SetPutawayRuleArchivedResult = composable.MutateResult
	GetPutawayRuleByIdResult     = composable.GetOneResult
	SearchPutawayRulesResult     = composable.SearchResult
	PutawayRuleExistsResult      = composable.ExistsResult
)
