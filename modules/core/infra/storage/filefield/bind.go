package filefield

import (
	"encoding/json"
	"strings"

	"github.com/labstack/echo/v5"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	maxUploadMemory = 8 << 20
	queryParamOrgId = "org_id"
)

func BindUploadParams(echoCtx *echo.Context, fields []FileField) (dmodel.DynamicFields, *ft.ClientErrors, error) {
	vErrs := ft.NewClientErrors()
	params := dmodel.DynamicFields{}

	if id := echoCtx.Param("id"); id != "" {
		params[basemodel.FieldId] = id
	}

	contentType := echoCtx.Request().Header.Get(echo.HeaderContentType)
	if !strings.HasPrefix(contentType, "multipart/form-data") {
		bindJsonBody(echoCtx, params)
		mergeOrgId(echoCtx, params)
		return params, vErrs, nil
	}

	if err := echoCtx.Request().ParseMultipartForm(maxUploadMemory); err != nil {
		vErrs.Append(*ft.NewValidationError("body", "file.err_malformed_multipart",
			"the request body is not a readable multipart form"))
		return nil, vErrs, nil
	}

	form := echoCtx.Request().MultipartForm
	for name, values := range form.Value {
		if len(values) > 0 {
			params[name] = values[0]
		}
	}

	byUploadField := make(map[string]FileField, len(fields))
	for _, field := range fields {
		byUploadField[field.UploadField] = field
	}

	uploads := make([]PendingUpload, 0, len(fields))
	for name, headers := range form.File {
		field, known := byUploadField[name]
		if !known {
			vErrs.Append(*ft.NewValidationError(name, "file.err_upload_field_unknown",
				"'"+name+"' is not a file field of this resource"))
			continue
		}
		if len(headers) == 0 {
			continue
		}
		uploads = append(uploads, PendingUpload{Field: field, File: headers[0]})
	}

	if vErrs.Count() > 0 {
		return nil, vErrs, nil
	}
	if len(uploads) > 0 {
		params[UploadParamKey] = uploads
	}
	mergeOrgId(echoCtx, params)

	return params, vErrs, nil
}

func mergeOrgId(echoCtx *echo.Context, params dmodel.DynamicFields) {
	if existing, ok := params[queryParamOrgId].(string); ok && existing != "" {
		return
	}
	if orgId := echoCtx.QueryParam(queryParamOrgId); orgId != "" {
		params[queryParamOrgId] = orgId
	}
}

func bindJsonBody(echoCtx *echo.Context, params dmodel.DynamicFields) {
	body := echoCtx.Request().Body
	if body == nil {
		return
	}

	decoded := map[string]any{}
	if err := json.NewDecoder(body).Decode(&decoded); err != nil {
		return
	}
	for name, value := range decoded {
		params[name] = value
	}
}
