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
	group.PUT("/:id", this.UpdateRelationship)
	group.DELETE("/:id", this.DeleteRelationship)
	group.GET("/:id", this.GetRelationshipById)
	group.GET("", this.SearchRelationships)
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

func (this RelationshipRest) UpdateRelationship(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST update relationship"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.RelationshipSvc.UpdateRelationship,
		func(request UpdateRelationshipRequest) relationship.UpdateRelationshipCommand {
			return relationship.UpdateRelationshipCommand(request)
		},
		func(result relationship.UpdateRelationshipResult) UpdateRelationshipResponse {
			response := UpdateRelationshipResponse{}
			response.FromEntity(result.Data)
			return response
		},
		httpserver.JsonOk,
	)
	return err
}

func (this RelationshipRest) DeleteRelationship(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST delete relationship"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.RelationshipSvc.DeleteRelationship,
		func(request DeleteRelationshipRequest) relationship.DeleteRelationshipCommand {
			return relationship.DeleteRelationshipCommand(request)
		},
		func(result relationship.DeleteRelationshipResult) DeleteRelationshipResponse {
			response := DeleteRelationshipResponse{}
			response.FromNonEntity(result.Data)
			return response
		},
		httpserver.JsonOk,
	)
	return err
}

func (this RelationshipRest) GetRelationshipById(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST get relationship by id"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.RelationshipSvc.GetRelationshipById,
		func(request GetRelationshipByIdRequest) relationship.GetRelationshipByIdQuery {
			return relationship.GetRelationshipByIdQuery(request)
		},
		func(result relationship.GetRelationshipByIdResult) GetRelationshipByIdResponse {
			response := GetRelationshipByIdResponse{}
			response.FromRelationship(*result.Data)
			return response
		},
		httpserver.JsonOk,
	)
	return err
}

func (this RelationshipRest) SearchRelationships(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST search relationships"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.RelationshipSvc.SearchRelationships,
		func(request SearchRelationshipsRequest) relationship.SearchRelationshipsQuery {
			return relationship.SearchRelationshipsQuery(request)
		},
		func(result relationship.SearchRelationshipsResult) SearchRelationshipsResponse {
			response := SearchRelationshipsResponse{}
			response.FromResult(result.Data)
			return response
		},
		httpserver.JsonOk,
	)
	return err
}
