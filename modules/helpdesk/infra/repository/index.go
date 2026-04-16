package repository

import deps "github.com/sky-as-code/nikki-erp/common/deps_inject"

func InitRepositories() error {
	return deps.Register(
		NewTicketDynamicRepository,
		NewTicketActivityDynamicRepository,
		NewTicketMessageDynamicRepository,
		NewTicketAssignmentDynamicRepository,
		NewTicketCategoryDynamicRepository,
		NewSlaPolicyDynamicRepository,
		NewSlaBreachDynamicRepository,
		NewTeamDynamicRepository,
		NewTeamMembershipDynamicRepository,
		NewEscalationRuleDynamicRepository,
		NewTicketFeedbackDynamicRepository,
	)
}
