package restful

import (
	"errors"

	"github.com/labstack/echo/v4"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	v1 "github.com/sky-as-code/nikki-erp/modules/contacts/transport/restful/v1"
)

func InitRestfulHandlers() error {
	err := errors.Join(
		initPartyRest(),
	)
	return err
}

func initPartyRest() error {
	deps.Register(v1.NewPartyRest)
	return deps.Invoke(func(route *echo.Group, partyRest *v1.PartyRest) {
		v1 := route.Group("/v1/contacts")
		initV1(v1, partyRest)
	})
}

func initV1(route *echo.Group, partyRest *v1.PartyRest) {
	route.POST("/parties/tags", partyRest.CreatePartyTag)
	route.DELETE("/parties/tags/:id", partyRest.DeletePartyTag)
	route.GET("/parties/tags/:id", partyRest.GetPartyTagById)
	route.GET("/parties/tags", partyRest.ListPartyTags)
	route.PUT("/parties/tags/:id", partyRest.UpdatePartyTag)
}
