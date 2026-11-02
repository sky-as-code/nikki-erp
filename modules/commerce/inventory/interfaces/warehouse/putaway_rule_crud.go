package warehouse

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// PutawayRuleRepository reads and writes putaway_rule rows.
type PutawayRuleRepository interface {
	dynamicresource.CrudRepository
}

// PutawayRuleDomainService is the CRUD of the resource plus its own rules.
type PutawayRuleDomainService interface {
	dynamicresource.CrudDomainService
}

// PutawayRuleApplicationService is the authorized surface the REST handler serves.
type PutawayRuleApplicationService interface {
	dynamicresource.CrudApplicationService
	PutawayRuleActionService
}

// The CRUD commands, queries and results of the resource, as composable shapes under a
// resource-specific name.
type (
	CreatePutawayRuleCommand      = dynamicresource.CreateCommand
	UpdatePutawayRuleCommand      = dynamicresource.UpdateCommand
	DeletePutawayRuleCommand      = dynamicresource.DeleteCommand
	SetPutawayRuleArchivedCommand = dynamicresource.SetArchivedCommand
	GetPutawayRuleByIdQuery       = dynamicresource.GetByIdQuery
	SearchPutawayRulesQuery       = dynamicresource.SearchQuery
	PutawayRuleExistsQuery        = dynamicresource.ExistsQuery
)

type (
	CreatePutawayRuleResult      = dynamicresource.CreateResult
	UpdatePutawayRuleResult      = dynamicresource.MutateResult
	DeletePutawayRuleResult      = dynamicresource.MutateResult
	SetPutawayRuleArchivedResult = dynamicresource.MutateResult
	GetPutawayRuleByIdResult     = dynamicresource.GetOneResult
	SearchPutawayRulesResult     = dynamicresource.SearchResult
	PutawayRuleExistsResult      = dynamicresource.ExistsResult
)
