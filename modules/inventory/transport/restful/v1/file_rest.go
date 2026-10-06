package v1

import (
	"github.com/labstack/echo/v5"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/storage/filefield"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

type fileRest struct {
	resource string
	fileSvc  itProduct.FileFieldsService
}

func newFileRest(resource string, fileSvc itProduct.FileFieldsService) fileRest {
	return fileRest{resource: resource, fileSvc: fileSvc}
}

func (this fileRest) UploadFile(echoCtx *echo.Context, _ map[string]any) error {
	return this.serveStore(echoCtx, "upload "+this.resource+" file", this.fileSvc.UploadFile)
}

func (this fileRest) ReplaceFile(echoCtx *echo.Context, _ map[string]any) error {
	return this.serveStore(echoCtx, "replace "+this.resource+" file", this.fileSvc.ReplaceFile)
}

func (this fileRest) DeleteFile(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "delete "+this.resource+" file", payload,
		func(ctx corectx.Context, params dmodel.DynamicFields) (*itProduct.FileFieldResult, error) {
			recordId := params.GetString(basemodel.FieldId)
			if recordId == nil || *recordId == "" {
				return &itProduct.FileFieldResult{ClientErrors: *missingRecordId()}, nil
			}
			delete(params, filefield.PathParamField)

			return this.fileSvc.DeleteFile(ctx, itProduct.DeleteFileCommand{
				RecordId:  model.Id(*recordId),
				FieldName: echoCtx.Param(filefield.PathParamField),
				Params:    params,
			})
		}, composable.MutateResponse)
}

func (this fileRest) serveStore(
	echoCtx *echo.Context,
	action string,
	call func(ctx corectx.Context, cmd itProduct.UploadFileCommand) (*itProduct.FileFieldResult, error),
) error {
	params, file, vErrs := filefield.BindFileUpload(echoCtx)
	if vErrs.Count() > 0 {
		return httpserver.JsonBadRequest(echoCtx, *vErrs)
	}

	return composable.ServeAction(echoCtx, action, params,
		func(ctx corectx.Context, params dmodel.DynamicFields) (*itProduct.FileFieldResult, error) {
			recordId := params.GetString(basemodel.FieldId)
			if recordId == nil || *recordId == "" {
				return &itProduct.FileFieldResult{ClientErrors: *missingRecordId()}, nil
			}

			return call(ctx, itProduct.UploadFileCommand{
				RecordId:  model.Id(*recordId),
				FieldName: echoCtx.Param(filefield.PathParamField),
				File:      file,
				Params:    params,
			})
		}, composable.MutateResponse)
}

func missingRecordId() *ft.ClientErrors {
	vErrs := ft.NewClientErrors()
	vErrs.Append(*ft.NewValidationError(basemodel.FieldId, "file.err_record_id_required",
		"the record id is missing from the request path"))
	return vErrs
}
