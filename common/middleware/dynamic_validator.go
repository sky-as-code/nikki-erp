package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicentity/model"
	dschema "github.com/sky-as-code/nikki-erp/common/dynamicentity/schema"
	dval "github.com/sky-as-code/nikki-erp/common/dynamicentity/validator"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
)

func DynamicValidator(adhocSchemaName string) echo.MiddlewareFunc {
	adhocSchema, err := dschema.GetAdhocSchema(adhocSchemaName)
	if err != nil {
		panic(err)
	}
	validator := dval.NewAdhocValidator(adhocSchema)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(echoCtx echo.Context) error {
			requestFields, err := ExtractRequestFields(echoCtx)
			if err != nil {
				return err
			}

			vErrs := validator.ValidateMap(requestFields, false, true)
			if vErrs.Count() > 0 {
				clientErr := vErrs.ToClientError()
				return httpserver.JsonBadRequest(echoCtx, clientErr)
			}

			echoCtx.Set("nikkiDynamicRequestFields", requestFields)

			return next(echoCtx)
		}
	}
}

func ExtractRequestFields(echoCtx echo.Context) (dmodel.EntityMap, error) {
	// Build request fields map
	requestFields := make(dmodel.EntityMap)

	// Extract path parameters
	paramNames := echoCtx.ParamNames()
	paramValues := echoCtx.ParamValues()
	for i, name := range paramNames {
		if i < len(paramValues) {
			requestFields[name] = paramValues[i]
		}
	}

	// Extract query parameters (override path params)
	for name, values := range echoCtx.QueryParams() {
		if len(values) > 0 {
			// Take the first value if multiple values exist
			requestFields[name] = values[0]
		}
	}

	method := echoCtx.Request().Method
	// For POST, PUT, PATCH: extract body (override query params)
	if method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch {
		// Check Content-Type header for JSON
		contentType := echoCtx.Request().Header.Get(echo.HeaderContentType)
		if contentType != "" {
			mediaType, _, err := mime.ParseMediaType(contentType)
			if err != nil || !strings.HasPrefix(mediaType, "application/json") {
				return nil, echo.NewHTTPError(http.StatusUnsupportedMediaType, "Content-Type must be application/json")
			}
		}

		bodyBytes, err := io.ReadAll(echoCtx.Request().Body)
		if err != nil {
			return nil, err
		}

		// Restore body for later middlewares or handlers to use.
		echoCtx.Request().Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		if len(bodyBytes) > 0 {
			var bodyMap dmodel.EntityMap
			if err := json.Unmarshal(bodyBytes, &bodyMap); err == nil {
				// Merge body fields (override query params)
				for key, value := range bodyMap {
					requestFields[key] = value
				}
			}
		}
	}

	return requestFields, nil
}
