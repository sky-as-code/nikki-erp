package v1

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/contacts/interfaces/relationship"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
)

type relationshipRestParams struct {
	dig.In

	RelationshipSvc relationship.RelationshipService
}

func NewRelationshipRest(params relationshipRestParams) *RelationshipRest {
	return &RelationshipRest{
		RelationshipSvc: params.RelationshipSvc,
	}
}

type RelationshipRest struct {
	httpserver.RestBase
	RelationshipSvc relationship.RelationshipService
}

func (this *RelationshipRest) RegisterRoutes(apiGroup *echo.Group) {
	group := apiGroup.Group("/relationships")

	group.POST("", this.CreateRelationship)
}

func (this RelationshipRest) CreateRelationship(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST create relationship"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.RelationshipSvc.CreateRelationship,
		func(request CreateRelationshipRequest) relationship.CreateRelationshipCommand {
			return relationship.CreateRelationshipCommand(request)
		},
		func(result relationship.CreateRelationshipResult) CreateRelationshipResponse {
			response := CreateRelationshipResponse{}
			response.FromEntity(result.Data)
			return response
		},
		httpserver.JsonCreated,
	)
	return err
}
