package app

import deps "github.com/sky-as-code/nikki-erp/common/deps_inject"

func InitApplicationServices() error {
	return deps.Register(
		NewEscalationRuleApplicationServiceImpl,
		NewSlaBreachApplicationServiceImpl,
		NewSlaPolicyApplicationServiceImpl,
		NewTeamApplicationServiceImpl,
		NewTeamMembershipApplicationServiceImpl,
		NewTicketApplicationServiceImpl,
		NewTicketActivityApplicationServiceImpl,
		NewTicketAssignmentApplicationServiceImpl,
		NewTicketCategoryApplicationServiceImpl,
		NewTicketFeedbackApplicationServiceImpl,
		NewTicketMessageApplicationServiceImpl,
	)
}
