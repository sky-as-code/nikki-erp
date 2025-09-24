package restful

import (
	"github.com/labstack/echo/v4"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	v1 "github.com/sky-as-code/nikki-erp/modules/contacts/transport/restful/v1"
)

func InitRestfulHandlers() error {
	deps.Register(
		v1.NewPartyRest,
		v1.NewRelationshipRest,
		v1.NewCommChannelRest,
	)
	return deps.Invoke(func(route *echo.Group, partyRest *v1.PartyRest, relationshipRest *v1.RelationshipRest, commChannelRest *v1.CommChannelRest) {
		v1 := route.Group("/v1/contacts")
		initV1(v1, partyRest, relationshipRest, commChannelRest)
	})
}

func initV1(route *echo.Group, partyRest *v1.PartyRest, relationshipRest *v1.RelationshipRest, commChannelRest *v1.CommChannelRest) {
	route.POST("/parties", partyRest.CreateParty)
	route.DELETE("/parties/:id", partyRest.DeleteParty)
	route.GET("/parties/:id", partyRest.GetPartyById)
	route.GET("/parties", partyRest.SearchParties)
	route.PUT("/parties/:id", partyRest.UpdateParty)

	route.POST("/parties/relationships", relationshipRest.CreateRelationship)

	route.POST("/parties/channels", commChannelRest.CreateCommChannel)
	route.DELETE("/parties/channels/:id", commChannelRest.DeleteCommChannel)
	route.GET("/parties/channels/:id", commChannelRest.GetCommChannelById)
	route.GET("/parties/channels", commChannelRest.SearchCommChannels)
	route.PUT("/parties/channels/:id", commChannelRest.UpdateCommChannel)
}
