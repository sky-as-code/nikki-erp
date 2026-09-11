package composable

import (
	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/computed"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
)

// SourceSearchFn fetches rows of another resource for a batched related-field read.
type SourceSearchFn func(
	ctx corectx.Context, schemaName string, keyColumn string, keys []any, fields []string,
) ([]dmodel.DynamicFields, error)

// FunctionInvokeFn runs a registered computed-field function over a page of rows.
type FunctionInvokeFn func(
	ctx corectx.Context, schemaName string, functionName string, fieldName string,
	rows []dmodel.DynamicFields,
) ([]any, error)

// WithComputedFields decorates a domain service so declared computed fields evaluate on every
// read and reject every write. Wrapping is unconditional and costs nothing for a schema without
// computed fields: the eval planner returns nil and every call passes straight through.
//
// The onion wraps the default before handing it to NewDomainServiceFn, so a module's derived
// service embeds the wrapped one and its overrides keep layering on top.
//
// defaultSearchFields must be the same list the wrapped service falls back to when a search
// names no fields, otherwise a defaulted listing evaluates the wrong set of computed fields and
// reads operands the projection never selected.
func WithComputedFields(
	base CrudDomainService, sourceSearch SourceSearchFn, invokeFunction FunctionInvokeFn,
	defaultSearchFields []string,
) CrudDomainService {
	return &computedFieldService{
		CrudDomainService:   base,
		sourceSearch:        sourceSearch,
		invokeFunction:      invokeFunction,
		defaultSearchFields: defaultSearchFields,
	}
}

type computedFieldService struct {
	CrudDomainService
	sourceSearch   SourceSearchFn
	invokeFunction FunctionInvokeFn

	// defaultSearchFields mirrors the wrapped service's fallback projection for Search.
	defaultSearchFields []string
}

func (this *computedFieldService) Create(
	ctx corectx.Context, cmd CreateCommand, options ...CreateOptions,
) (*CreateResult, error) {
	if errs := computed.RejectWrites(this.Schema(), cmd); errs.Count() > 0 {
		return &CreateResult{ClientErrors: errs}, nil
	}
	return this.CrudDomainService.Create(ctx, cmd, options...)
}

func (this *computedFieldService) Update(
	ctx corectx.Context, cmd UpdateCommand, options ...UpdateOptions,
) (*MutateResult, error) {
	if errs := computed.RejectWrites(this.Schema(), cmd); errs.Count() > 0 {
		return &MutateResult{ClientErrors: errs}, nil
	}
	return this.CrudDomainService.Update(ctx, cmd, options...)
}

func (this *computedFieldService) Search(
	ctx corectx.Context, query SearchQuery, options ...corecrud.ServiceSearchOptions,
) (*SearchResult, error) {
	plan, errs := this.prepareRead(query, this.searchProjection(query))
	if errs.Count() > 0 {
		return &SearchResult{ClientErrors: errs}, nil
	}
	result, err := this.CrudDomainService.Search(ctx, query, options...)
	if err != nil || plan == nil || result == nil || !result.HasData {
		return result, err
	}
	if err := plan.Apply(result.Data.Items, this.evalDeps(ctx)); err != nil {
		return nil, err
	}
	return result, nil
}

func (this *computedFieldService) GetById(
	ctx corectx.Context, query GetByIdQuery,
) (*GetOneResult, error) {
	return this.getOneComputed(ctx, query, this.CrudDomainService.GetById)
}

func (this *computedFieldService) GetOne(
	ctx corectx.Context, query GetOneQuery,
) (*GetOneResult, error) {
	return this.getOneComputed(ctx, query, this.CrudDomainService.GetOne)
}

type getOneDelegateFn func(ctx corectx.Context, query dmodel.DynamicFields) (*GetOneResult, error)

func (this *computedFieldService) getOneComputed(
	ctx corectx.Context, query dmodel.DynamicFields, delegate getOneDelegateFn,
) (*GetOneResult, error) {
	// GetById/GetOne with no explicit projection return the whole record, so every operand is
	// already there and the effective projection is "everything", represented by a nil list.
	plan, errs := this.prepareRead(query, requestedFieldNames(query))
	if errs.Count() > 0 {
		return &GetOneResult{ClientErrors: errs}, nil
	}
	result, err := delegate(ctx, query)
	if err != nil || plan == nil || result == nil || !result.HasData || result.Data.Item == nil {
		return result, err
	}
	if err := plan.Apply([]dmodel.DynamicFields{result.Data.Item}, this.evalDeps(ctx)); err != nil {
		return nil, err
	}
	return result, nil
}

// prepareRead builds the request's eval plan and appends the physical operands evaluation needs.
// A nil plan means nothing computed is wanted.
//
// projection is the field list the read will actually return: what the client named, or the
// service's own fallback when it named nothing. An empty projection means "every column", the
// only case in which the operands are guaranteed to be present already.
func (this *computedFieldService) prepareRead(
	params dmodel.DynamicFields, projection []string,
) (*computed.EvalPlan, ft.ClientErrors) {
	plan, errs := computed.BuildEvalPlan(this.Schema().Name(), projection)
	if errs.Count() > 0 || plan == nil {
		return plan, errs
	}
	if len(projection) > 0 && len(plan.ExtraFields) > 0 {
		// The operands go on the wire projection so the row carries them; the response still
		// shows only what the caller asked for, because DesiredFields is taken from `fields`
		// before this augmentation reaches the client-facing field list.
		params[fieldNameFields] = append(projection, plan.ExtraFields...)
	}
	return plan, nil
}

// searchProjection resolves what a Search will project, mirroring crud.UiSearch: an explicit
// `fields` wins; otherwise the default view falls back to the service's default search fields,
// and any other named view resolves to an id-only row until saved searches land.
func (this *computedFieldService) searchProjection(params dmodel.DynamicFields) []string {
	if requested := requestedFieldNames(params); len(requested) > 0 {
		return requested
	}
	if name, ok := searchName(params); ok && name != dyn.DefaultSearchName {
		return []string{basemodel.FieldId}
	}
	return this.defaultSearchFields
}

func searchName(params dmodel.DynamicFields) (string, bool) {
	switch typed := params[basemodel.FieldSearchName].(type) {
	case string:
		return typed, true
	case *string:
		if typed == nil {
			return "", false
		}
		return *typed, true
	default:
		return "", false
	}
}

// evalDeps binds the request context into the callbacks evaluation needs. The context travels in
// the closure rather than through the computed package, which must stay free of engine and
// transport concepts.
func (this *computedFieldService) evalDeps(ctx corectx.Context) computed.EvalDeps {
	return computed.EvalDeps{
		Search: this.rowSearch(ctx),
		Invoke: this.functionInvoker(ctx),
	}
}

func (this *computedFieldService) rowSearch(ctx corectx.Context) computed.SourceSearchFn {
	return func(schemaName string, keyColumn string, keys []any, fields []string) ([]dmodel.DynamicFields, error) {
		return this.sourceSearch(ctx, schemaName, keyColumn, keys, fields)
	}
}

func (this *computedFieldService) functionInvoker(ctx corectx.Context) computed.FunctionInvokerFn {
	if this.invokeFunction == nil {
		return nil
	}
	schemaName := this.Schema().Name()
	return func(functionName string, fieldName string, rows []dmodel.DynamicFields) ([]any, error) {
		return this.invokeFunction(ctx, schemaName, functionName, fieldName, rows)
	}
}

func requestedFieldNames(params dmodel.DynamicFields) []string {
	raw, ok := params[fieldNameFields]
	if !ok || raw == nil {
		return nil
	}
	switch typed := raw.(type) {
	case []string:
		return typed
	case []any:
		fields := make([]string, 0, len(typed))
		for _, item := range typed {
			if str, ok := item.(string); ok {
				fields = append(fields, str)
			}
		}
		return fields
	default:
		return nil
	}
}
