package filefield

import (
	"mime/multipart"
	"strings"

	"github.com/labstack/echo/v5"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	maxUploadMemory = 8 << 20
	PathParamField  = "field"
	FormPartFile    = "file"
	queryParamOrgId = "org_id"
)

func BindFileUpload(echoCtx *echo.Context) (dmodel.DynamicFields, *multipart.FileHeader, *ft.ClientErrors) {
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
		if name != FormPartFile {
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
