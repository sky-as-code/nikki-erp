package restful

import (
	"github.com/labstack/echo/v5"

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
		v1 := route.Group("/v1/:org_id/contacts")
		initV1(v1, partyRest, relationshipRest, commChannelRest)
	})
}

func initV1(route *echo.Group, partyRest *v1.PartyRest, relationshipRest *v1.RelationshipRest, commChannelRest *v1.CommChannelRest) {
	route.POST("/parties", partyRest.CreateParty)
	route.DELETE("/parties/:id", partyRest.DeleteParty)
	route.GET("/parties/:id", partyRest.GetParty)
	route.GET("/parties", partyRest.SearchParties)
	route.PUT("/parties/:id", partyRest.UpdateParty)

	route.POST("/parties/:party_id/relationships", relationshipRest.CreateRelationship)
	route.DELETE("/parties/:party_id/relationships/:id", relationshipRest.DeleteRelationship)
	route.GET("/parties/:party_id/relationships/:id", relationshipRest.GetRelationship)
	route.GET("/parties/:party_id/relationships", relationshipRest.SearchRelationships)
	route.PUT("/parties/:party_id/relationships/:id", relationshipRest.UpdateRelationship)

	route.POST("/parties/:party_id/channels", commChannelRest.CreateCommChannel)
	route.DELETE("/parties/:party_id/channels/:id", commChannelRest.DeleteCommChannel)
	route.GET("/parties/:party_id/channels/:id", commChannelRest.GetCommChannel)
	route.GET("/parties/:party_id/channels", commChannelRest.SearchCommChannels)
	route.PUT("/parties/:party_id/channels/:id", commChannelRest.UpdateCommChannel)
}
