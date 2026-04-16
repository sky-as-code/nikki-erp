package restful

import (
	stdErr "errors"

	"github.com/labstack/echo/v5"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	v1 "github.com/sky-as-code/nikki-erp/modules/helpdesk/transport/restful/v1"
)

func InitRestfulHandlers() error {
	err := deps.Register(
		v1.NewTicketRest,
		v1.NewTicketActivityRest,
		v1.NewTicketMessageRest,
		v1.NewTicketAssignmentRest,
		v1.NewTicketCategoryRest,
		v1.NewSlaPolicyRest,
		v1.NewSlaBreachRest,
		v1.NewTeamRest,
		v1.NewTeamMembershipRest,
		v1.NewEscalationRuleRest,
		v1.NewTicketFeedbackRest,
	)
	err = stdErr.Join(err, initHelpdeskV1())
	return err
}

func initHelpdeskV1() error {
	return deps.Invoke(func(
		route *echo.Group,
		ticketRest *v1.TicketRest,
		ticketactivityRest *v1.TicketActivityRest,
		ticketmessageRest *v1.TicketMessageRest,
		ticketassignmentRest *v1.TicketAssignmentRest,
		ticketcategoryRest *v1.TicketCategoryRest,
		slapolicyRest *v1.SlaPolicyRest,
		slabreachRest *v1.SlaBreachRest,
		teamRest *v1.TeamRest,
		teammembershipRest *v1.TeamMembershipRest,
		escalationruleRest *v1.EscalationRuleRest,
		ticketfeedbackRest *v1.TicketFeedbackRest,
	) {
		routeV1 := route.Group("/v1/helpdesk")

		routeV1.DELETE("/tickets/:id", ticketRest.DeleteTicket)
		routeV1.GET("/tickets/:id", ticketRest.GetTicket)
		routeV1.GET("/tickets", ticketRest.SearchTickets)
		routeV1.POST("/tickets/exists", ticketRest.TicketExists)
		routeV1.POST("/tickets/:id/archived", ticketRest.SetTicketIsArchived)
		routeV1.POST("/tickets/:ticket_id/manage-categories", ticketRest.ManageTicketCategories)
		routeV1.POST("/tickets", ticketRest.CreateTicket)
		routeV1.PUT("/tickets/:id", ticketRest.UpdateTicket)

		routeV1.DELETE("/ticket-activities/:id", ticketactivityRest.DeleteTicketActivity)
		routeV1.GET("/ticket-activities/:id", ticketactivityRest.GetTicketActivity)
		routeV1.GET("/ticket-activities", ticketactivityRest.SearchTicketActivities)
		routeV1.POST("/ticket-activities/exists", ticketactivityRest.TicketActivityExists)
		routeV1.POST("/ticket-activities", ticketactivityRest.CreateTicketActivity)
		routeV1.PUT("/ticket-activities/:id", ticketactivityRest.UpdateTicketActivity)

		routeV1.DELETE("/ticket-messages/:id", ticketmessageRest.DeleteTicketMessage)
		routeV1.GET("/ticket-messages/:id", ticketmessageRest.GetTicketMessage)
		routeV1.GET("/ticket-messages", ticketmessageRest.SearchTicketMessages)
		routeV1.POST("/ticket-messages/exists", ticketmessageRest.TicketMessageExists)
		routeV1.POST("/ticket-messages", ticketmessageRest.CreateTicketMessage)
		routeV1.PUT("/ticket-messages/:id", ticketmessageRest.UpdateTicketMessage)

		routeV1.DELETE("/ticket-assignments/:id", ticketassignmentRest.DeleteTicketAssignment)
		routeV1.GET("/ticket-assignments/:id", ticketassignmentRest.GetTicketAssignment)
		routeV1.GET("/ticket-assignments", ticketassignmentRest.SearchTicketAssignments)
		routeV1.POST("/ticket-assignments/exists", ticketassignmentRest.TicketAssignmentExists)
		routeV1.POST("/ticket-assignments", ticketassignmentRest.CreateTicketAssignment)
		routeV1.PUT("/ticket-assignments/:id", ticketassignmentRest.UpdateTicketAssignment)

		routeV1.DELETE("/ticket-categories/:id", ticketcategoryRest.DeleteTicketCategory)
		routeV1.GET("/ticket-categories/:id", ticketcategoryRest.GetTicketCategory)
		routeV1.GET("/ticket-categories", ticketcategoryRest.SearchTicketCategories)
		routeV1.POST("/ticket-categories/exists", ticketcategoryRest.TicketCategoryExists)
		routeV1.POST("/ticket-categories/:id/archived", ticketcategoryRest.SetTicketCategoryIsArchived)
		routeV1.POST("/ticket-categories", ticketcategoryRest.CreateTicketCategory)
		routeV1.PUT("/ticket-categories/:id", ticketcategoryRest.UpdateTicketCategory)

		routeV1.DELETE("/sla-policies/:id", slapolicyRest.DeleteSlaPolicy)
		routeV1.GET("/sla-policies/:id", slapolicyRest.GetSlaPolicy)
		routeV1.GET("/sla-policies", slapolicyRest.SearchSlaPolicies)
		routeV1.POST("/sla-policies/exists", slapolicyRest.SlaPolicyExists)
		routeV1.POST("/sla-policies/:id/archived", slapolicyRest.SetSlaPolicyIsArchived)
		routeV1.POST("/sla-policies", slapolicyRest.CreateSlaPolicy)
		routeV1.PUT("/sla-policies/:id", slapolicyRest.UpdateSlaPolicy)

		routeV1.DELETE("/sla-breaches/:id", slabreachRest.DeleteSlaBreach)
		routeV1.GET("/sla-breaches/:id", slabreachRest.GetSlaBreach)
		routeV1.GET("/sla-breaches", slabreachRest.SearchSlaBreaches)
		routeV1.POST("/sla-breaches/exists", slabreachRest.SlaBreachExists)
		routeV1.POST("/sla-breaches", slabreachRest.CreateSlaBreach)
		routeV1.PUT("/sla-breaches/:id", slabreachRest.UpdateSlaBreach)

		routeV1.DELETE("/teams/:id", teamRest.DeleteTeam)
		routeV1.GET("/teams/:id", teamRest.GetTeam)
		routeV1.GET("/teams", teamRest.SearchTeams)
		routeV1.POST("/teams/exists", teamRest.TeamExists)
		routeV1.POST("/teams/:id/archived", teamRest.SetTeamIsArchived)
		routeV1.POST("/teams", teamRest.CreateTeam)
		routeV1.PUT("/teams/:id", teamRest.UpdateTeam)

		routeV1.DELETE("/team-memberships/:id", teammembershipRest.DeleteTeamMembership)
		routeV1.GET("/team-memberships/:id", teammembershipRest.GetTeamMembership)
		routeV1.GET("/team-memberships", teammembershipRest.SearchTeamMemberships)
		routeV1.POST("/team-memberships/exists", teammembershipRest.TeamMembershipExists)
		routeV1.POST("/team-memberships", teammembershipRest.CreateTeamMembership)
		routeV1.PUT("/team-memberships/:id", teammembershipRest.UpdateTeamMembership)

		routeV1.DELETE("/escalation-rules/:id", escalationruleRest.DeleteEscalationRule)
		routeV1.GET("/escalation-rules/:id", escalationruleRest.GetEscalationRule)
		routeV1.GET("/escalation-rules", escalationruleRest.SearchEscalationRules)
		routeV1.POST("/escalation-rules/exists", escalationruleRest.EscalationRuleExists)
		routeV1.POST("/escalation-rules", escalationruleRest.CreateEscalationRule)
		routeV1.PUT("/escalation-rules/:id", escalationruleRest.UpdateEscalationRule)

		routeV1.DELETE("/ticket-feedbacks/:id", ticketfeedbackRest.DeleteTicketFeedback)
		routeV1.GET("/ticket-feedbacks/:id", ticketfeedbackRest.GetTicketFeedback)
		routeV1.GET("/ticket-feedbacks", ticketfeedbackRest.SearchTicketFeedbacks)
		routeV1.POST("/ticket-feedbacks/exists", ticketfeedbackRest.TicketFeedbackExists)
		routeV1.POST("/ticket-feedbacks", ticketfeedbackRest.CreateTicketFeedback)
		routeV1.PUT("/ticket-feedbacks/:id", ticketfeedbackRest.UpdateTicketFeedback)

	})
}
