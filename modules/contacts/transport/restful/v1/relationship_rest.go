package v1

import (
	"github.com/labstack/echo/v5"
	"go.uber.org/dig"

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

func (this RelationshipRest) CreateRelationship(echoCtx *echo.Context) (err error) {
	return httpserver.ServeCreate(
		"create relationship",
		echoCtx,
		&relationship.CreateRelationshipCommand{},
		this.RelationshipSvc.CreateRelationship,
	)
}

func (this RelationshipRest) UpdateRelationship(echoCtx *echo.Context) (err error) {
	return httpserver.ServeUpdate(
		"update relationship",
		echoCtx,
		&relationship.UpdateRelationshipCommand{},
		this.RelationshipSvc.UpdateRelationship,
	)
}

func (this RelationshipRest) DeleteRelationship(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGeneralMutate(
		"delete relationship",
		echoCtx,
		this.RelationshipSvc.DeleteRelationship,
	)
}

func (this RelationshipRest) GetRelationship(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGetOne(
		"get relationship",
		echoCtx,
		this.RelationshipSvc.GetRelationship,
	)
}

func (this RelationshipRest) SearchRelationships(echoCtx *echo.Context) (err error) {
	return httpserver.ServeSearch(
		"search relationships",
		echoCtx,
		this.RelationshipSvc.SearchRelationships,
	)
}

func (this RelationshipRest) RelationshipExists(echoCtx *echo.Context) (err error) {
	return httpserver.ServeExists(
		"relationship exists",
		echoCtx,
		this.RelationshipSvc.RelationshipExists,
	)
}
