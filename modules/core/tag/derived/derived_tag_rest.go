package derived

import (
	"github.com/labstack/echo/v4"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	it "github.com/sky-as-code/nikki-erp/modules/core/tag/interfaces"
)

//
// NOTE: This is REST Controller for feature Tags inheriting from core Tags.
//

type DerivedRestParams struct {
	Config  config.ConfigService
	Logger  logging.LoggerService
	TagSvc  it.TagService
	CqrsBus cqrs.CqrsBus
}

func NewDerivedRest(params DerivedRestParams) *DerivedRest {
	return &DerivedRest{
		RestBase: httpserver.RestBase{
			ConfigSvc: params.Config,
			Logger:    params.Logger,
			CqrsBus:   params.CqrsBus,
		},
		TagSvc: params.TagSvc,
	}
}

type DerivedRest struct {
	httpserver.RestBase
	TagSvc it.TagService
}

func (this DerivedRest) CreateDerivedTag(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to handle REST create derived tag"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.TagSvc.CreateTag,
		func(request CreateDerivedTagRequest) it.CreateTagCommand {
			return it.CreateTagCommand(request)
		},
		func(result it.CreateTagResult) CreateDerivedTagResponse {
			return NewDerivedTagDto(*result.Data)
		},
		httpserver.JsonCreated,
	)
	return err
}

func (this DerivedRest) UpdateDerivedTag(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to handle REST update derived tag"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.TagSvc.UpdateTag,
		func(request UpdateDerivedTagRequest) it.UpdateTagCommand {
			return it.UpdateTagCommand(request)
		},
		func(result it.UpdateTagResult) UpdateDerivedTagResponse {
			return NewDerivedTagDto(*result.Data)
		},
		httpserver.JsonOk,
	)
	return err
}

func (this DerivedRest) DeleteDerivedTag(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to handle REST delete derived tag"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.TagSvc.DeleteTag,
		func(request DeleteDerivedTagRequest) it.DeleteTagCommand {
			return it.DeleteTagCommand(request)
		},
		func(result it.DeleteTagResult) DeleteDerivedTagResponse {
			return NewDeleteDerivedTagResponse(result)
		},
		httpserver.JsonOk,
	)
	return err
}

func (this DerivedRest) GetDerivedTagById(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to handle REST get derived tag by id"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.TagSvc.GetTagById,
		func(request GetDerivedTagByIdRequest) it.GetTagByIdQuery {
			return it.GetTagByIdQuery(request)
		},
		func(result it.GetTagByIdResult) GetDerivedTagByIdResponse {
			return NewDerivedTagDto(*result.Data)
		},
		httpserver.JsonOk,
	)
	return err
}

func (this DerivedRest) ListDerivedTags(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to handle REST list derived tags"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.TagSvc.ListTags,
		func(request ListDerivedTagsRequest) it.ListTagsQuery {
			return it.ListTagsQuery(request)
		},
		func(result it.ListTagsResult) ListDerivedTagsResponse {
			return NewListDerivedTagsResponse(result)
		},
		httpserver.JsonOk,
	)
	return err
}
