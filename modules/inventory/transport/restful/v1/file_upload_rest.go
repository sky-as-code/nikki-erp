package v1

import (
	"github.com/labstack/echo/v5"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/storage/filefield"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/services"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

func serveUpload[TData any](
	echoCtx *echo.Context,
	action string,
	fields []filefield.FileField,
	call func(ctx corectx.Context, params dmodel.DynamicFields) (*dyn.OpResult[TData], error),
	shape func(data TData) any,
) error {
	params, vErrs, err := filefield.BindUploadParams(echoCtx, fields)
	if err != nil {
		return err
	}
	if vErrs.Count() > 0 {
		return httpserver.JsonBadRequest(echoCtx, *vErrs)
	}
	return composable.ServeAction(echoCtx, action, params, call, shape)
}

func serveClearFiles(
	echoCtx *echo.Context, action string, payload map[string]any, svc itProduct.FileFieldsService,
) error {
	return composable.ServeAction(echoCtx, action, payload,
		func(ctx corectx.Context, params dmodel.DynamicFields) (*dyn.OpResult[dyn.MutateResultData], error) {
			recordId := params.GetString(basemodel.FieldId)
			if recordId == nil || *recordId == "" {
				vErrs := ft.NewClientErrors()
				vErrs.Append(*ft.NewValidationError(basemodel.FieldId, "file.err_record_id_required",
					"the record id is missing from the request path"))
				return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}, nil
			}
			return svc.ClearFiles(ctx, model.Id(*recordId), params)
		}, composable.MutateResponse)
}

func (this *ProductTemplateRest) CreateWithUpload(echoCtx *echo.Context, _ map[string]any) error {
	return serveUpload(echoCtx, "create product template", services.ProductTemplateFileFields(),
		this.productTemplateSvc.Create, composable.Identity[dmodel.DynamicFields])
}

func (this *ProductTemplateRest) UpdateWithUpload(echoCtx *echo.Context, _ map[string]any) error {
	return serveUpload(echoCtx, "update product template", services.ProductTemplateFileFields(),
		this.productTemplateSvc.Update, composable.MutateResponse)
}

func (this *ProductTemplateRest) ClearFiles(echoCtx *echo.Context, payload map[string]any) error {
	return serveClearFiles(echoCtx, "clear product template files", payload, this.productTemplateSvc)
}

func (this *ProductVariantRest) CreateWithUpload(echoCtx *echo.Context, _ map[string]any) error {
	return serveUpload(echoCtx, "create product variant", services.ProductVariantFileFields(),
		this.productVariantSvc.Create, composable.Identity[dmodel.DynamicFields])
}

func (this *ProductVariantRest) UpdateWithUpload(echoCtx *echo.Context, _ map[string]any) error {
	return serveUpload(echoCtx, "update product variant", services.ProductVariantFileFields(),
		this.productVariantSvc.Update, composable.MutateResponse)
}

func (this *ProductVariantRest) ClearFiles(echoCtx *echo.Context, payload map[string]any) error {
	return serveClearFiles(echoCtx, "clear product variant files", payload, this.productVariantSvc)
}
