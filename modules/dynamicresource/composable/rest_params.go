package composable

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
	queryParamOrgId    = basemodel.FieldOrgId
	queryParamFields   = fieldNameFields
	queryParamPage     = "page"
	queryParamSize     = "size"
	queryParamGraph    = fieldNameGraph
	queryParamContext  = basemodel.FieldContext
	queryParamLanguage = "language"
	queryParamName     = "search_name"
	// queryParamIncludeArchived is tri-state: absent means "hide archived", which crud.Search
	// applies as the public-API default. Only a value the caller actually sent is forwarded.
	queryParamIncludeArchived = basemodel.FieldIncludeArchived
)

// bindFn assembles the command or query out of the HTTP request.
type bindFn func(echoCtx *echo.Context) (dmodel.DynamicFields, error)

func noParams(_ *echo.Context) (dmodel.DynamicFields, error) {
	return dmodel.DynamicFields{}, nil
}

// bindPayload is the generic binder behind RestEngine.AddRoute. It assembles params the way
// echo.Bind does, then resolves the org. Data is bound in this order, each step overwriting the
// previous: path parameters, query parameters (GET and DELETE only), request body.
//
// Path and query values arrive as strings; an application-service method converts what it
// needs, or validates through the schema.
func bindPayload(echoCtx *echo.Context, schema *dmodel.ModelSchema, actionType ActionType) (dmodel.DynamicFields, error) {
	params := dmodel.DynamicFields{}
	mergePathParams(echoCtx, params)

	method := echoCtx.Request().Method
	if method == http.MethodGet || method == http.MethodDelete {
		mergeQueryParams(echoCtx, params)
	}

	if actionType.HasRequestBody() && hasRequestBody(echoCtx) {
		body, err := rawBodyParams(echoCtx)
		if err != nil {
			return nil, err
		}
		maps.Copy(params, body)
	}
	mergeOrgId(echoCtx, schema, params, actionType)
	return params, nil
}

// bindRawBodyParams is rawBodyParams plus the org, for actions such as exists whose body is a
// query rather than a record.
func bindRawBodyParams(echoCtx *echo.Context, schema *dmodel.ModelSchema, actionType ActionType) (dmodel.DynamicFields, error) {
	params, err := rawBodyParams(echoCtx)
	if err != nil {
		return nil, err
	}
	mergePathParams(echoCtx, params)
	mergeOrgId(echoCtx, schema, params, actionType)
	return params, nil
}

// mergePathParams copies the route path params into params. Echo reuses a pooled backing slice
// for PathValues, so entries past the matched count carry an empty Name and are skipped.
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

// mergeOrgId resolves the org a request acts in, from the request body on a write action and
// from the query string otherwise. Absence is not an error here: the application service's
// org-scope step is what rejects a missing value, so the same rule applies to every action.
//
// A write action carries the record, and the record carries its own org_id field, so that is
// where the org is read from. A read has no body, leaving the query string as its only source.
// The query parameter stays a fallback for writes too: a generic action may be a bodyless POST
// such as ":id/confirm", which has no body to read an org from.
//
// The schema guard matters for create: createBodyParams rejects body keys the schema does not
// declare, so writing org_id into the params of an org-less resource would invent a field the
// resource has no column for.
func mergeOrgId(echoCtx *echo.Context, schema *dmodel.ModelSchema, params dmodel.DynamicFields, actionType ActionType) {
	if !schemaHasOrgId(schema) {
		return
	}
	// Every write binder parses the body into params before calling this, so a body-supplied
	// org is already present and only has to be left alone.
	if actionType.HasRequestBody() && readString(params, queryParamOrgId) != "" {
		return
	}
	if orgId := echoCtx.QueryParam(queryParamOrgId); orgId != "" {
		params[queryParamOrgId] = orgId
	}
}

func schemaHasOrgId(schema *dmodel.ModelSchema) bool {
	if schema == nil {
		return false
	}
	_, exists := schema.Field(queryParamOrgId)
	return exists
}

// hasRequestBody reports whether the request carries a body worth binding, so that a bodyless
// POST does not fail on an EOF from the JSON decoder.
func hasRequestBody(echoCtx *echo.Context) bool {
	request := echoCtx.Request()
	if request.Body == nil || request.ContentLength == 0 {
		return false
	}
	return request.Header.Get(echo.HeaderContentType) != ""
}

// rawBodyParams binds the request body as-is, without filtering it against the schema.
func rawBodyParams(echoCtx *echo.Context) (dmodel.DynamicFields, error) {
	body := map[string]any{}
	if err := echoCtx.Bind(&body); err != nil {
		return nil, err
	}
	return dmodel.DynamicFields(body), nil
}

// createBodyParams binds the body, rejects fields the schema does not declare, and converts the
// rest to their declared data types. On create the distinction matters: silently dropping an
// unknown field answers 201 to a request that did not do what the caller asked. Update keeps
// the permissive binding, because clients round-tripping a fetched record back may carry
// read-only or computed keys they never intended to write.
func createBodyParams(echoCtx *echo.Context, schema *dmodel.ModelSchema) (dmodel.DynamicFields, error) {
	raw, err := rawBodyParams(echoCtx)
	if err != nil {
		return nil, err
	}
	if cErrs := unknownFieldErrors(schema, raw); cErrs != nil {
		return nil, &unknownFieldsError{errors: *cErrs}
	}
	// The body stream is already consumed, so filter the parsed map rather than re-binding.
	params := httpserver.FilterToDynamicEntity(raw, schema)
	mergePathParams(echoCtx, params)
	mergeOrgId(echoCtx, schema, params, ActionTypeCreate)
	return params, nil
}

// unknownFieldErrors reports every body key that names no schema field. The base models a
// resource extends (id, etag, timestamps) are part of its schema, so they pass on their own.
func unknownFieldErrors(schema *dmodel.ModelSchema, raw dmodel.DynamicFields) *ft.ClientErrors {
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

// unknownFieldsError carries client errors out of a bindFn, whose signature only allows a Go
// error. serve unwraps it into a 400 instead of a 500.
type unknownFieldsError struct {
	errors ft.ClientErrors
}

func (this *unknownFieldsError) Error() string {
	return "request body contains fields not defined on this schema"
}

// updateParams binds the record body, keeps only declared fields, and overrides the id with the
// one from the path, so that the route always decides which record is being updated.
func updateParams(echoCtx *echo.Context, schema *dmodel.ModelSchema) (dmodel.DynamicFields, error) {
	params, err := httpserver.BindToDynamicEntity(echoCtx, schema)
	if err != nil {
		return nil, err
	}
	mergePathParams(echoCtx, params)
	params[basemodel.FieldId] = echoCtx.Param("id")
	mergeOrgId(echoCtx, schema, params, ActionTypeUpdatePatch)
	return params, nil
}

// deleteParams reads the record id from the path and the org from the query string.
func deleteParams(echoCtx *echo.Context, schema *dmodel.ModelSchema) (dmodel.DynamicFields, error) {
	params := dmodel.DynamicFields{}
	mergePathParams(echoCtx, params)
	params[basemodel.FieldId] = echoCtx.Param("id")
	mergeOrgId(echoCtx, schema, params, ActionTypeDelete)
	return params, nil
}

// getByIdParams reads the record id from the path and the desired fields from the query string.
func getByIdParams(echoCtx *echo.Context, schema *dmodel.ModelSchema) (dmodel.DynamicFields, error) {
	params := dmodel.DynamicFields{}
	mergePathParams(echoCtx, params)
	params[basemodel.FieldId] = echoCtx.Param("id")
	mergeOrgId(echoCtx, schema, params, ActionTypeRead)
	if fields := readCsvQuery(echoCtx, queryParamFields); len(fields) > 0 {
		params[queryParamFields] = fields
	}
	return params, nil
}

// archivedParams reads the archived flag and etag from the body, and the id from the path.
func archivedParams(echoCtx *echo.Context, schema *dmodel.ModelSchema) (dmodel.DynamicFields, error) {
	params, err := rawBodyParams(echoCtx)
	if err != nil {
		return nil, err
	}
	mergePathParams(echoCtx, params)
	params[basemodel.FieldId] = echoCtx.Param("id")
	mergeOrgId(echoCtx, schema, params, ActionTypeGeneric)
	return params, nil
}

// searchParams reads paging, field selection and the search graph from the query string.
func searchParams(echoCtx *echo.Context, schema *dmodel.ModelSchema) (dmodel.DynamicFields, error) {
	params := dmodel.DynamicFields{}
	mergePathParams(echoCtx, params)
	mergeOrgId(echoCtx, schema, params, ActionTypeRead)

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
	// Parsed here because SearchQuery.IncludeArchived is a *bool and the params are decoded into
	// it by strict JSON unmarshalling, which a raw string fails outright. An unparseable value is
	// forwarded untouched so the query schema reports it as an invalid format, rather than being
	// dropped and silently read as "hide archived".
	if raw := echoCtx.QueryParam(queryParamIncludeArchived); raw != "" {
		if parsed, err := strconv.ParseBool(raw); err == nil {
			params[queryParamIncludeArchived] = parsed
		} else {
			params[queryParamIncludeArchived] = raw
		}
	}
	return params, readJsonQueryParams(echoCtx, params)
}

func readJsonQueryParams(echoCtx *echo.Context, params dmodel.DynamicFields) error {
	if raw := echoCtx.QueryParam(queryParamGraph); raw != "" {
		graph := map[string]any{}
		if err := json.Unmarshal([]byte(raw), &graph); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "malformed 'graph' query parameter")
		}
		params[queryParamGraph] = graph
	}
	if raw := echoCtx.QueryParam(queryParamContext); raw != "" {
		contextValues := map[string]any{}
		if err := json.Unmarshal([]byte(raw), &contextValues); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "malformed 'context' query parameter")
		}
		params[queryParamContext] = contextValues
	}
	return nil
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
