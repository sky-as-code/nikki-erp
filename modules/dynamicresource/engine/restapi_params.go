package engine

import (
	"encoding/json"
	"maps"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/labstack/echo/v5"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
)

// Query parameter names accepted by the read endpoints.
const (
	queryParamOrgId    = "orgId"
	queryParamFields   = "fields"
	queryParamPage     = "page"
	queryParamSize     = "size"
	queryParamGraph    = "graph"
	queryParamLanguage = "language"
	queryParamName     = "search_name"

	fieldOrgId = "org_id"
)

func noParams(_ *echo.Context) (dmodel.DynamicFields, error) {
	return dmodel.DynamicFields{}, nil
}

// echoBindParams assembles action params the way echo.Bind does, and serves every action a
// feature module defines without a binding of its own. Data is bound in this order, each
// step overwriting the previous:
//
//  1. path parameters
//  2. query parameters (GET and DELETE only)
//  3. request body
//
// Path and query values arrive as strings; the pipeline's schema validation converts them
// when the action declares a ParamSchema.
func echoBindParams(echoCtx *echo.Context) (dmodel.DynamicFields, error) {
	params := dmodel.DynamicFields{}
	mergePathParams(echoCtx, params)

	method := echoCtx.Request().Method
	if method == http.MethodGet || method == http.MethodDelete {
		mergeQueryParams(echoCtx, params)
	}

	if hasRequestBody(echoCtx) {
		body, err := rawBodyParams(echoCtx)
		if err != nil {
			return nil, err
		}
		maps.Copy(params, body)
	}
	return params, nil
}

// mergePathParams copies the route path params into params. Echo reuses a pooled backing
// slice for PathValues, so entries past the matched count carry an empty Name and are skipped.
func mergePathParams(echoCtx *echo.Context, params dmodel.DynamicFields) {
	for _, pathValue := range echoCtx.PathValues() {
		if pathValue.Name != "" {
			params[pathValue.Name] = pathValue.Value
		}
	}
}

// mergeQueryParams copies the query string into params, keeping a repeated key as a []string.
func mergeQueryParams(echoCtx *echo.Context, params dmodel.DynamicFields) {
	for name, values := range echoCtx.QueryParams() {
		switch len(values) {
		case 0:
			continue
		case 1:
			params[name] = values[0]
		default:
			params[name] = values
		}
	}
}

// hasRequestBody reports whether the request carries a body worth binding, so that a
// bodyless POST does not fail on an EOF from the JSON decoder.
func hasRequestBody(echoCtx *echo.Context) bool {
	request := echoCtx.Request()
	if request.Body == nil || request.ContentLength == 0 {
		return false
	}
	return request.Header.Get(echo.HeaderContentType) != ""
}

// rawBodyParams binds the request body as-is, without filtering it against the schema.
// Used by endpoints whose body is a query rather than a record, such as exists.
func rawBodyParams(echoCtx *echo.Context) (dmodel.DynamicFields, error) {
	body := map[string]any{}
	if err := echoCtx.Bind(&body); err != nil {
		return nil, err
	}
	return dmodel.DynamicFields(body), nil
}

// bodyParams binds the request body and keeps only the fields the schema declares,
// converting each to its declared data type. An undeclared field is dropped, which is the
// long-standing behaviour of every update endpoint.
func (this *DynamicRestApiImpl) bodyParams(echoCtx *echo.Context) (dmodel.DynamicFields, error) {
	return httpserver.BindToDynamicEntity(echoCtx, this.engine.Schema())
}

// createBodyParams is bodyParams plus a rejection of fields the schema does not declare.
// On create the distinction matters: silently dropping an unknown field answers 201 to a
// request that did not do what the caller asked, and the record it names never exists.
// Update keeps the permissive binding — clients round-tripping a fetched record back may
// carry read-only or computed keys they never intended to write.
func (this *DynamicRestApiImpl) createBodyParams(echoCtx *echo.Context) (dmodel.DynamicFields, error) {
	raw, err := rawBodyParams(echoCtx)
	if err != nil {
		return nil, err
	}
	if cErrs := this.unknownFieldErrors(raw); cErrs != nil {
		return nil, &unknownFieldsError{errors: *cErrs}
	}
	// The body stream is already consumed, so filter the parsed map rather than re-binding.
	return httpserver.FilterToDynamicEntity(raw, this.engine.Schema()), nil
}

// unknownFieldErrors reports every body key that names no schema field. The base models a
// resource extends (id, etag, timestamps) are part of its schema, so they pass on their own.
func (this *DynamicRestApiImpl) unknownFieldErrors(raw dmodel.DynamicFields) *ft.ClientErrors {
	schema := this.engine.Schema()
	if schema == nil {
		return nil
	}

	names := make([]string, 0, len(raw))
	for name := range raw {
		if _, exists := schema.Field(name); !exists {
			names = append(names, name)
		}
	}
	if len(names) == 0 {
		return nil
	}

	// Map iteration order is random; the response lists the fields deterministically.
	sort.Strings(names)
	cErrs := ft.ClientErrors{}
	for _, name := range names {
		cErrs.Append(*ft.NewValidationError(name, ft.ErrorKey("err_unknown_schema_field"),
			"field is not defined on this schema"))
	}
	return &cErrs
}

// unknownFieldsError carries client errors out of a buildParamsFn, whose signature only
// allows a Go error. serveAction unwraps it into a 400 instead of a 500.
type unknownFieldsError struct {
	errors ft.ClientErrors
}

func (this *unknownFieldsError) Error() string {
	return "request body contains fields not defined on this schema"
}

// deleteParams reads the record id from the path and the optional org from the query string.
func (this *DynamicRestApiImpl) deleteParams(echoCtx *echo.Context) (dmodel.DynamicFields, error) {
	params := dmodel.DynamicFields{
		basemodel.FieldId: echoCtx.Param("id"),
	}
	if orgId := echoCtx.QueryParam(queryParamOrgId); orgId != "" {
		params[fieldOrgId] = orgId
	}
	return params, nil
}

// getByIdParams reads the record id from the path and the desired fields from the query string.
func (this *DynamicRestApiImpl) getByIdParams(echoCtx *echo.Context) (dmodel.DynamicFields, error) {
	params := dmodel.DynamicFields{
		basemodel.FieldId: echoCtx.Param("id"),
	}
	if fields := readCsvQuery(echoCtx, queryParamFields); len(fields) > 0 {
		params[queryParamFields] = fields
	}
	return params, nil
}

// searchParams reads paging, field selection and the search graph from the query string.
func (this *DynamicRestApiImpl) searchParams(echoCtx *echo.Context) (dmodel.DynamicFields, error) {
	params := dmodel.DynamicFields{}

	if fields := readCsvQuery(echoCtx, queryParamFields); len(fields) > 0 {
		params[queryParamFields] = fields
	}
	if page, ok := readIntQuery(echoCtx, queryParamPage); ok {
		params[queryParamPage] = page
	}
	if size, ok := readIntQuery(echoCtx, queryParamSize); ok {
		params[queryParamSize] = size
	}
	if language := echoCtx.QueryParam(queryParamLanguage); language != "" {
		params[queryParamLanguage] = language
	}
	if searchName := echoCtx.QueryParam(queryParamName); searchName != "" {
		params[queryParamName] = searchName
	}

	if raw := echoCtx.QueryParam(queryParamGraph); raw != "" {
		graph := map[string]any{}
		if err := json.Unmarshal([]byte(raw), &graph); err != nil {
			return nil, echo.NewHTTPError(400, "malformed 'graph' query parameter")
		}
		params[queryParamGraph] = graph
	}

	return params, nil
}

// archivedParams reads the archived flag and etag from the body, and the id from the path.
func (this *DynamicRestApiImpl) archivedParams(echoCtx *echo.Context) (dmodel.DynamicFields, error) {
	params, err := rawBodyParams(echoCtx)
	if err != nil {
		return nil, err
	}
	params[basemodel.FieldId] = echoCtx.Param("id")
	return params, nil
}

// updateParams binds the record body and overrides the id with the one from the path,
// so that the route always decides which record is being updated.
func (this *DynamicRestApiImpl) updateParams(echoCtx *echo.Context) (dmodel.DynamicFields, error) {
	params, err := this.bodyParams(echoCtx)
	if err != nil {
		return nil, err
	}
	params[basemodel.FieldId] = echoCtx.Param("id")
	return params, nil
}

// readCsvQuery reads a query parameter that may be repeated or comma-separated.
func readCsvQuery(echoCtx *echo.Context, name string) []string {
	values := echoCtx.QueryParams()[name]
	if len(values) == 0 {
		return nil
	}

	result := make([]string, 0, len(values))
	for _, value := range values {
		for _, item := range strings.Split(value, ",") {
			item = strings.TrimSpace(item)
			if item != "" {
				result = append(result, item)
			}
		}
	}
	return result
}

func readIntQuery(echoCtx *echo.Context, name string) (int, bool) {
	raw := echoCtx.QueryParam(name)
	if raw == "" {
		return 0, false
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, false
	}
	return value, true
}
