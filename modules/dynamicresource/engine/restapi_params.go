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
	it "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
)

// Query parameter names accepted by the read endpoints.
const (
	queryParamOrgId    = basemodel.FieldOrgId
	queryParamFields   = "fields"
	queryParamPage     = "page"
	queryParamSize     = "size"
	queryParamGraph    = "graph"
	queryParamContext  = basemodel.FieldContext
	queryParamLanguage = "language"
	queryParamName     = "search_name"
	// queryParamIncludeArchived is tri-state: absent means "hide archived", which crud.Search
	// applies as the public-API default. Only a value the caller actually sent is forwarded.
	queryParamIncludeArchived = basemodel.FieldIncludeArchived
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

// bindGenericParams is echoBindParams plus the org, and serves every action a feature module
// defined without a binding of its own. A generic action is POST-shaped, so its org normally
// arrives in the body; echoBindParams has already merged that. mergeOrgId still runs because
// echoBindParams merges the query string on GET and DELETE only, so a bodyless custom action
// would otherwise never see the org the caller sent in the URL.
func (this *DynamicRestApiImpl) bindGenericParams(
	echoCtx *echo.Context, actionName string,
) (dmodel.DynamicFields, error) {
	params, err := echoBindParams(echoCtx)
	if err != nil {
		return nil, err
	}
	this.mergeOrgId(echoCtx, params, actionName)
	return params, nil
}

// bindRawBodyParams is rawBodyParams plus the org, for actions such as exists whose body is a
// query rather than a record.
func (this *DynamicRestApiImpl) bindRawBodyParams(
	echoCtx *echo.Context, actionName string,
) (dmodel.DynamicFields, error) {
	params, err := rawBodyParams(echoCtx)
	if err != nil {
		return nil, err
	}
	mergePathParams(echoCtx, params)
	this.mergeOrgId(echoCtx, params, actionName)
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

// mergeOrgId resolves the org a request acts in, from the request body on a write action and
// from the query string otherwise. Absence is not an error here: the pipeline's org-scope step
// is what rejects a missing value, so that the same rule applies to every action rather than to
// whichever bindings remembered to check.
//
// A write action carries the record, and the record carries its own org_id field, so that is
// where the org is read from. A read has no body, leaving the query string as its only source.
// The query parameter stays a fallback for writes too: an ActionTypeGeneric action may be a
// bodyless POST such as ":id/confirm", which has no body to read an org from.
//
// The schema guard matters for create: createBodyParams rejects body keys the schema does not
// declare, so writing org_id into the params of an org-less resource would invent a field the
// resource has no column for. That guard is also why an org-less resource is never refused for
// lacking an org - it is left alone entirely.
func (this *DynamicRestApiImpl) mergeOrgId(
	echoCtx *echo.Context, params dmodel.DynamicFields, actionName string,
) {
	if !this.schemaHasOrgId() {
		return
	}
	// Every write binder parses the body into params before calling this, so a body-supplied
	// org is already present and only has to be left alone.
	if this.isBodyOrgAction(actionName) && readString(params, queryParamOrgId) != "" {
		return
	}
	if orgId := echoCtx.QueryParam(queryParamOrgId); orgId != "" {
		params[queryParamOrgId] = orgId
	}
}

// isBodyOrgAction reports whether an action takes its org_id from the request body rather than
// the query string. True for the action types that carry a body: create, both update flavours,
// and the generic type feature modules use for custom actions.
//
// An action the engine does not know falls back to query-only, which is the safe direction: a
// custom action registered outside the normal path keeps the behaviour it had before.
func (this *DynamicRestApiImpl) isBodyOrgAction(actionName string) bool {
	definition, exists := this.engine.Action(actionName)
	if !exists {
		return false
	}
	switch definition.ActionType {
	case it.ActionTypeCreate, it.ActionTypeUpdatePatch,
		it.ActionTypeUpdateReplace, it.ActionTypeGeneric:
		return true
	}
	return false
}

// schemaHasOrgId reports whether this engine's resource declares an org column. A resource that
// does not cannot be org-filtered, and is left alone by the org-scoping machinery.
func (this *DynamicRestApiImpl) schemaHasOrgId() bool {
	schema := this.engine.Schema()
	if schema == nil {
		return false
	}
	_, exists := schema.Field(queryParamOrgId)
	return exists
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
	params := httpserver.FilterToDynamicEntity(raw, this.engine.Schema())
	mergePathParams(echoCtx, params)
	// After the unknown-field check, so that an org-less resource never sees an invented
	// org_id key. Create is a write action, so an org_id the body carried is authoritative and
	// the query parameter only fills in when the body named none.
	this.mergeOrgId(echoCtx, params, it.ActionCreate)
	return params, nil
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

// deleteParams reads the record id from the path and the org from the query string.
func (this *DynamicRestApiImpl) deleteParams(echoCtx *echo.Context) (dmodel.DynamicFields, error) {
	params := dmodel.DynamicFields{}
	mergePathParams(echoCtx, params)
	params[basemodel.FieldId] = echoCtx.Param("id")
	this.mergeOrgId(echoCtx, params, it.ActionDelete)
	return params, nil
}

// getByIdParams reads the record id from the path and the desired fields from the query string.
func (this *DynamicRestApiImpl) getByIdParams(echoCtx *echo.Context) (dmodel.DynamicFields, error) {
	params := dmodel.DynamicFields{}
	mergePathParams(echoCtx, params)
	params[basemodel.FieldId] = echoCtx.Param("id")
	this.mergeOrgId(echoCtx, params, it.ActionGetById)
	if fields := readCsvQuery(echoCtx, queryParamFields); len(fields) > 0 {
		params[queryParamFields] = fields
	}
	return params, nil
}

// searchParams reads paging, field selection and the search graph from the query string.
func (this *DynamicRestApiImpl) searchParams(echoCtx *echo.Context) (dmodel.DynamicFields, error) {
	params := dmodel.DynamicFields{}
	mergePathParams(echoCtx, params)
	this.mergeOrgId(echoCtx, params, it.ActionSearch)

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

	if raw := echoCtx.QueryParam(queryParamGraph); raw != "" {
		graph := map[string]any{}
		if err := json.Unmarshal([]byte(raw), &graph); err != nil {
			return nil, echo.NewHTTPError(400, "malformed 'graph' query parameter")
		}
		params[queryParamGraph] = graph
	}

	if raw := echoCtx.QueryParam(queryParamContext); raw != "" {
		contextValues := map[string]any{}
		if err := json.Unmarshal([]byte(raw), &contextValues); err != nil {
			return nil, echo.NewHTTPError(400, "malformed 'context' query parameter")
		}
		params[queryParamContext] = contextValues
	}

	return params, nil
}

// archivedParams reads the archived flag and etag from the body, and the id from the path.
func (this *DynamicRestApiImpl) archivedParams(echoCtx *echo.Context) (dmodel.DynamicFields, error) {
	params, err := rawBodyParams(echoCtx)
	if err != nil {
		return nil, err
	}
	mergePathParams(echoCtx, params)
	params[basemodel.FieldId] = echoCtx.Param("id")
	this.mergeOrgId(echoCtx, params, it.ActionSetArchived)
	return params, nil
}

// updateParams binds the record body and overrides the id with the one from the path,
// so that the route always decides which record is being updated.
func (this *DynamicRestApiImpl) updateParams(echoCtx *echo.Context) (dmodel.DynamicFields, error) {
	params, err := this.bodyParams(echoCtx)
	if err != nil {
		return nil, err
	}
	mergePathParams(echoCtx, params)
	params[basemodel.FieldId] = echoCtx.Param("id")
	this.mergeOrgId(echoCtx, params, it.ActionUpdate)
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
