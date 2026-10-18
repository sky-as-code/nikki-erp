package filefield

import (
	"mime/multipart"
	"strings"

	"github.com/labstack/echo/v5"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

const (
	maxUploadMemory = 8 << 20
	pathParamField  = "field"
	formPartFile    = "file"
	queryParamOrgId = "org_id"
)

// Rest serves the `:id/files/:field` routes; register its methods with composable AddRoute
// (ActionTypeUpload, ActionTypeUploadReplace, ActionTypeDelete).
type Rest struct {
	resource string
	fileSvc  FileFieldsService
}

func NewRest(resource string, fileSvc FileFieldsService) Rest {
	return Rest{resource: resource, fileSvc: fileSvc}
}

func (this Rest) UploadFile(echoCtx *echo.Context, _ map[string]any) error {
	return this.serveStore(echoCtx, "upload "+this.resource+" file", this.fileSvc.UploadFile)
}

func (this Rest) ReplaceFile(echoCtx *echo.Context, _ map[string]any) error {
	return this.serveStore(echoCtx, "replace "+this.resource+" file", this.fileSvc.ReplaceFile)
}

func (this Rest) DeleteFile(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "delete "+this.resource+" file", payload,
		func(ctx corectx.Context, params dmodel.DynamicFields) (*FileFieldResult, error) {
			recordId := params.GetString(basemodel.FieldId)
			if recordId == nil || *recordId == "" {
				return &FileFieldResult{ClientErrors: *missingRecordId()}, nil
			}
			delete(params, pathParamField)

			return this.fileSvc.DeleteFile(ctx, DeleteFileCommand{
				RecordId:  model.Id(*recordId),
				FieldName: echoCtx.Param(pathParamField),
				Params:    params,
			})
		}, composable.MutateResponse)
}

func (this Rest) serveStore(
	echoCtx *echo.Context,
	action string,
	call func(ctx corectx.Context, cmd UploadFileCommand) (*FileFieldResult, error),
) error {
	params, file, vErrs := bindFileUpload(echoCtx)
	if vErrs.Count() > 0 {
		return httpserver.JsonBadRequest(echoCtx, *vErrs)
	}

	return composable.ServeAction(echoCtx, action, params,
		func(ctx corectx.Context, params dmodel.DynamicFields) (*FileFieldResult, error) {
			recordId := params.GetString(basemodel.FieldId)
			if recordId == nil || *recordId == "" {
				return &FileFieldResult{ClientErrors: *missingRecordId()}, nil
			}

			return call(ctx, UploadFileCommand{
				RecordId:  model.Id(*recordId),
				FieldName: echoCtx.Param(pathParamField),
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

func bindFileUpload(echoCtx *echo.Context) (dmodel.DynamicFields, *multipart.FileHeader, *ft.ClientErrors) {
	vErrs := ft.NewClientErrors()
	params := dmodel.DynamicFields{}
	if id := echoCtx.Param("id"); id != "" {
		params[basemodel.FieldId] = id
	}

	contentType := echoCtx.Request().Header.Get(echo.HeaderContentType)
	if !strings.HasPrefix(contentType, "multipart/form-data") {
		vErrs.Append(*ft.NewValidationError("body", "file.err_multipart_required",
			"the request body must be a multipart form with a 'file' part"))
		return nil, nil, vErrs
	}
	if err := echoCtx.Request().ParseMultipartForm(maxUploadMemory); err != nil {
		vErrs.Append(*ft.NewValidationError("body", "file.err_malformed_multipart",
			"the request body is not a readable multipart form"))
		return nil, nil, vErrs
	}

	form := echoCtx.Request().MultipartForm
	for name, values := range form.Value {
		if len(values) > 0 {
			params[name] = values[0]
		}
	}

	var file *multipart.FileHeader
	for name, headers := range form.File {
		if name != formPartFile {
			vErrs.Append(*ft.NewValidationError(name, "file.err_upload_part_unknown",
				"'"+name+"' is not accepted; send the file as the 'file' part"))
			continue
		}
		if len(headers) > 0 {
			file = headers[0]
		}
	}
	if vErrs.Count() > 0 {
		return nil, nil, vErrs
	}

	if existing, ok := params[queryParamOrgId].(string); !ok || existing == "" {
		if orgId := echoCtx.QueryParam(queryParamOrgId); orgId != "" {
			params[queryParamOrgId] = orgId
		}
	}

	return params, file, vErrs
}
