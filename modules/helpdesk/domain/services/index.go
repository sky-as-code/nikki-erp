package services

import deps "github.com/sky-as-code/nikki-erp/common/deps_inject"

func InitDomainServices() error {
	return deps.Register(
		NewEscalationRuleDomainServiceImpl,
		NewSlaBreachDomainServiceImpl,
		NewSlaPolicyDomainServiceImpl,
		NewTeamDomainServiceImpl,
		NewTeamMembershipDomainServiceImpl,
		NewTicketDomainServiceImpl,
		NewTicketActivityDomainServiceImpl,
		NewTicketAssignmentDomainServiceImpl,
		NewTicketCategoryDomainServiceImpl,
		NewTicketFeedbackDomainServiceImpl,
		NewTicketMessageDomainServiceImpl,
	)
}
