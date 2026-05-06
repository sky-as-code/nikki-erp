package v1

import (
	"github.com/labstack/echo/v5"
	"go.uber.org/dig"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/tag"
)

type tagRestParams struct {
	dig.In

	TagSvc it.TagService
}

func NewTagRest(params tagRestParams) *TagRest {
	return &TagRest{
		tagSvc: params.TagSvc,
	}
}

type TagRest struct {
	tagSvc it.TagService
}

func (this TagRest) Create(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST create tag"); e != nil {
			err = e
		}
	}()

	return httpserver.ServeRequest2(
		echoCtx,
		this.tagSvc.CreateTag,
		func(request CreateTagRequest) it.CreateTagCommand {
			cmd := it.CreateTagCommand{}
			cmd.SetFieldData(request.DynamicFields)
			return cmd
		},
		func(data models.Tag) CreateTagResponse {
			return *httpserver.NewRestCreateResponseDyn(data.GetFieldData())
		},
		httpserver.JsonCreated,
	)
}

func (this TagRest) Delete(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST delete tag"); e != nil {
			err = e
		}
	}()

	return httpserver.ServeRequest2(
		echoCtx,
		this.tagSvc.DeleteTag,
		func(request DeleteTagRequest) it.DeleteTagCommand {
			return it.DeleteTagCommand(request)
		},
		func(data dyn.MutateResultData) DeleteTagResponse {
			return httpserver.NewRestDeleteResponse2(data)
		},
		httpserver.JsonOk,
	)
}

func (this TagRest) Exists(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST tag exists"); e != nil {
			err = e
		}
	}()

	return httpserver.ServeRequest2(
		echoCtx,
		this.tagSvc.TagExists,
		func(request TagExistsRequest) it.TagExistsQuery {
			return it.TagExistsQuery(request)
		},
		func(data dyn.ExistsResultData) TagExistsResponse {
			return TagExistsResponse(data)
		},
		httpserver.JsonOk,
	)
}

func (this TagRest) GetOne(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST get tag"); e != nil {
			err = e
		}
	}()

	return httpserver.ServeRequest2(
		echoCtx,
		this.tagSvc.GetTag,
		func(request GetTagRequest) it.GetTagQuery {
			return it.GetTagQuery(request)
		},
		func(data models.Tag) GetTagResponse {
			return data.GetFieldData()
		},
		httpserver.JsonOk,
	)
}

func (this TagRest) Search(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST search tags"); e != nil {
			err = e
		}
	}()

	return httpserver.ServeRequest2(
		echoCtx,
		this.tagSvc.SearchTags,
		func(request SearchTagsRequest) it.SearchTagsQuery {
			return it.SearchTagsQuery(request)
		},
		func(data it.SearchTagsResultData) SearchTagsResponse {
			return httpserver.NewSearchResponseDyn(data)
		},
		httpserver.JsonOk,
	)
}

func (this TagRest) Update(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST update tag"); e != nil {
			err = e
		}
	}()

	return httpserver.ServeRequest2(
		echoCtx,
		this.tagSvc.UpdateTag,
		func(request UpdateTagRequest) it.UpdateTagCommand {
			cmd := it.UpdateTagCommand{}
			cmd.SetFieldData(request.DynamicFields)
			cmd.SetId(util.ToPtr(model.Id(request.TagId)))
			return cmd
		},
		httpserver.NewRestMutateResponse,
		httpserver.JsonOk,
	)
}
