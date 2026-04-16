package v1

import (
	"github.com/labstack/echo/v5"
	"go.uber.org/dig"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain"
	it "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/language"
)

type languageRestParams struct {
	dig.In
	LanguageSvc it.LanguageService
}

func NewLanguageRest(params languageRestParams) *LanguageRest {
	return &LanguageRest{svc: params.LanguageSvc}
}

type LanguageRest struct{ svc it.LanguageService }

func (this LanguageRest) CreateLanguage(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST create language"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(echoCtx, this.svc.CreateLanguage,
		func(request CreateLanguageRequest) it.CreateLanguageCommand {
			cmd := it.CreateLanguageCommand{}
			cmd.SetFieldData(request.DynamicFields)
			return cmd
		},
		func(data domain.Language) CreateLanguageResponse {
			return *httpserver.NewRestCreateResponseDyn(data.GetFieldData())
		},
		httpserver.JsonCreated)
}

func (this LanguageRest) DeleteLanguage(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST delete language"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(echoCtx, this.svc.DeleteLanguage,
		func(request DeleteLanguageRequest) it.DeleteLanguageCommand { return it.DeleteLanguageCommand(request) },
		func(data dyn.MutateResultData) DeleteLanguageResponse { return httpserver.NewRestDeleteResponse2(data) },
		httpserver.JsonOk)
}

func (this LanguageRest) GetLanguage(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST get language"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(echoCtx, this.svc.GetLanguage,
		func(request GetLanguageRequest) it.GetLanguageQuery { return it.GetLanguageQuery(request) },
		func(data domain.Language) GetLanguageResponse { return data.GetFieldData() },
		httpserver.JsonOk)
}

func (this LanguageRest) LanguageExists(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST language exists"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(echoCtx, this.svc.LanguageExists,
		func(request LanguageExistsRequest) it.LanguageExistsQuery { return it.LanguageExistsQuery(request) },
		func(data dyn.ExistsResultData) LanguageExistsResponse { return LanguageExistsResponse(data) },
		httpserver.JsonOk)
}

func (this LanguageRest) SearchLanguages(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST search languages"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(echoCtx, this.svc.SearchLanguages,
		func(request SearchLanguagesRequest) it.SearchLanguagesQuery { return it.SearchLanguagesQuery(request) },
		func(data it.SearchLanguagesResultData) SearchLanguagesResponse {
			return httpserver.NewSearchResponseDyn(data)
		},
		httpserver.JsonOk, true)
}

func (this LanguageRest) UpdateLanguage(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST update language"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(echoCtx, this.svc.UpdateLanguage,
		func(request UpdateLanguageRequest) it.UpdateLanguageCommand {
			cmd := it.UpdateLanguageCommand{}
			cmd.SetFieldData(request.DynamicFields)
			cmd.SetId(util.ToPtr(model.Id(request.Id)))
			return cmd
		},
		httpserver.NewRestMutateResponse,
		httpserver.JsonOk)
}
