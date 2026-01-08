package v1

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"

	domain "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/entitlement"
)

type entitlementRestParams struct {
	dig.In

	EntitlementSvc it.EntitlementService
}

func NewEntitlementRest(params entitlementRestParams) *EntitlementRest {
	return &EntitlementRest{
		EntitlementSvc: params.EntitlementSvc,
	}
}

type EntitlementRest struct {
	EntitlementSvc it.EntitlementService
}

func (this EntitlementRest) CreateEntitlement(echoCtx echo.Context) (err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "handle REST create entitlement"); e != nil {
			err = e
		}
	}()

	err = httpserver.ServeRequest(
		echoCtx,
		this.EntitlementSvc.CreateEntitlement,
		func(request CreateEntitlementRequest) it.CreateEntitlementCommand {
			return it.CreateEntitlementCommand(request)
		},
		func(result it.CreateEntitlementResult) CreateEntitlementResponse {
			response := CreateEntitlementResponse{}
			response.FromEntity(result.Data)
			return response
		},
		httpserver.JsonCreated,
	)

	return err
}

func (this EntitlementRest) UpdateEntitlement(echoCtx echo.Context) (err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "handle REST update entitlement"); e != nil {
			err = e
		}
	}()

	err = httpserver.ServeRequest(
		echoCtx,
		this.EntitlementSvc.UpdateEntitlement,
		func(request UpdateEntitlementRequest) it.UpdateEntitlementCommand {
			return it.UpdateEntitlementCommand(request)
		},
		func(result it.UpdateEntitlementResult) UpdateEntitlementResponse {
			response := UpdateEntitlementResponse{}
			response.FromEntity(result.Data)
			return response
		},
		httpserver.JsonOk,
	)

	return err
}

func (this EntitlementRest) DeleteEntitlementHard(echoCtx echo.Context) (err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "handle REST delete entitlement hard"); e != nil {
			err = e
		}
	}()

	err = httpserver.ServeRequest(
		echoCtx,
		this.EntitlementSvc.DeleteEntitlementHard,
		func(request DeleteEntitlementHardByIdRequest) it.DeleteEntitlementHardByIdCommand {
			return it.DeleteEntitlementHardByIdCommand(request)
		},
		func(result it.DeleteEntitlementHardByIdResult) DeleteEntitlementHardByIdResponse {
			response := DeleteEntitlementHardByIdResponse{}
			response.FromNonEntity(result.Data)
			return response
		},
		httpserver.JsonOk,
	)

	return err
}

func (this EntitlementRest) GetEntitlementById(echoCtx echo.Context) (err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "handle REST get entitlement by id"); e != nil {
			err = e
		}
	}()

	err = httpserver.ServeRequest(
		echoCtx,
		this.EntitlementSvc.GetEntitlementById,
		func(request GetEntitlementByIdRequest) it.GetEntitlementByIdQuery {
			return it.GetEntitlementByIdQuery(request)
		},
		func(result it.GetEntitlementByIdResult) EntitlementDto {
			response := EntitlementDto{}
			response.FromEntitlement(*result.Data)
			return response
		},
		httpserver.JsonOk,
	)

	return err
}

func (this EntitlementRest) GetAllEntitlementByIds(echoCtx echo.Context) (err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "handle REST get all entitlement by ids"); e != nil {
			err = e
		}
	}()

	err = httpserver.ServeRequest(
		echoCtx,
		this.EntitlementSvc.GetAllEntitlementByIds,
		func(request GetAllEntitlementByIdsRequest) it.GetAllEntitlementByIdsQuery {
			return it.GetAllEntitlementByIdsQuery(request)
		},
		func(result it.GetAllEntitlementByIdsResult) []EntitlementDto {
			return array.Map(result.Data, func(entitlement domain.Entitlement) EntitlementDto {
				item := EntitlementDto{}
				item.FromEntitlement(entitlement)
				return item
			})
		},
		httpserver.JsonOk,
	)

	return err
}

func (this EntitlementRest) SearchEntitlements(echoCtx echo.Context) (err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "handle REST search entitlements"); e != nil {
			err = e
		}
	}()

	err = httpserver.ServeRequest(
		echoCtx,
		this.EntitlementSvc.SearchEntitlements,
		func(request SearchEntitlementsRequest) it.SearchEntitlementsQuery {
			return it.SearchEntitlementsQuery(request)
		},
		func(result it.SearchEntitlementsResult) SearchEntitlementsResponse {
			response := SearchEntitlementsResponse{}
			response.FromResult(result.Data)
			return response
		},
		httpserver.JsonOk,
	)

	return err
}
