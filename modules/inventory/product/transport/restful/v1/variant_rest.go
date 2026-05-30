package v1

import (
	"github.com/labstack/echo/v5"
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	itVariant "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/variant"
)

type variantRestParams struct {
	dig.In

	VariantSvc itVariant.VariantService
}

func NewVariantRest(params variantRestParams) *VariantRest {
	return &VariantRest{
		VariantSvc: params.VariantSvc,
	}
}

type VariantRest struct {
	httpserver.RestBase
	VariantSvc itVariant.VariantService
}

func (this VariantRest) Create(echoCtx *echo.Context) (err error) {
	return httpserver.ServeCreate(
		"create variant",
		echoCtx,
		&itVariant.CreateVariantCommand{},
		this.VariantSvc.CreateVariant,
	)
}

func (this VariantRest) Update(echoCtx *echo.Context) (err error) {
	return httpserver.ServeUpdate(
		"update variant",
		echoCtx,
		&itVariant.UpdateVariantCommand{},
		this.VariantSvc.UpdateVariant,
	)
}

func (this VariantRest) Delete(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGeneralMutate(
		"delete variant",
		echoCtx,
		this.VariantSvc.DeleteVariant,
	)
}

func (this VariantRest) GetOne(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGetOne(
		"get variant",
		echoCtx,
		this.VariantSvc.GetVariant,
	)
}

func (this VariantRest) Search(echoCtx *echo.Context) (err error) {
	return httpserver.ServeSearch(
		"search variants",
		echoCtx,
		this.VariantSvc.SearchVariants,
	)
}

func (this VariantRest) Exists(echoCtx *echo.Context) (err error) {
	return httpserver.ServeExists(
		"variant exists",
		echoCtx,
		this.VariantSvc.VariantExists,
	)
}

func (this VariantRest) UploadImage(echoCtx *echo.Context) (err error) {
	return httpserver.ServeRequestFormData(echoCtx,
		this.VariantSvc.UploadVariantImage,
		func(req UploadVariantImageRequest) itVariant.UploadVariantImageCommand {
			cmd := itVariant.UploadVariantImageCommand{
				ProductId: req.ProductId,
				VariantId: req.VariantId,
			}

			if req.FileHeader != nil {
				file, openErr := req.FileHeader.Open()
				if openErr != nil {
					panic(httpserver.JsonBadRequest(echoCtx,
						fault.ClientErrors{*fault.NewValidationError("file", "file_error", openErr.Error())}))
				}
				defer file.Close()
				cmd.File = file
				cmd.FileHeader = req.FileHeader
			}

			return cmd
		},
		func(data dynamicmodel.MutateResultData) UploadVariantImageResponse {
			return httpserver.NewRestMutateResponse(data)
		},
		httpserver.JsonCreated,
	)
}
