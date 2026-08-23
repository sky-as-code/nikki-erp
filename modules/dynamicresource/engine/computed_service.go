package engine

import (
	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/computed"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	it "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
)

// SourceSearchFn fetches rows of another resource for a batched related-field read. Supplied by
// the registry, which knows how to reach every engine's repository.
type SourceSearchFn func(
	ctx corectx.Context, schemaName string, keyColumn string, keys []any, fields []string,
) ([]dmodel.DynamicFields, error)

// WithComputedFields decorates a resource service so declared computed fields evaluate on every
// read and reject every write — the generic replacement for the per-module Search/GetById/GetOne
// overrides modules used to hand-roll for virtual fields. Wrapping is unconditional and costs
// nothing for a schema without computed fields: the eval planner returns nil and every call
// passes straight through.
//
// defaultSearchFields must be the same list the wrapped service falls back to when a search
// names no fields — otherwise a defaulted listing evaluates the wrong set of computed fields
// and reads operands the projection never selected.
func WithComputedFields(
	base it.DynamicResourceService, sourceSearch SourceSearchFn, defaultSearchFields []string,
) it.DynamicResourceService {
	return &computedFieldService{
		DynamicResourceService: base,
		sourceSearch:           sourceSearch,
		defaultSearchFields:    defaultSearchFields,
	}
}

type computedFieldService struct {
	it.DynamicResourceService
	sourceSearch SourceSearchFn

	// defaultSearchFields mirrors the wrapped service's fallback projection for Search.
	defaultSearchFields []string
}

func (this *computedFieldService) Create(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[dmodel.DynamicFields], error) {
	if errs := computed.RejectWrites(this.Schema(), params); errs.Count() > 0 {
		return &dyn.OpResult[dmodel.DynamicFields]{ClientErrors: errs}, nil
	}
	return this.DynamicResourceService.Create(ctx, params)
}

func (this *computedFieldService) Update(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	if errs := computed.RejectWrites(this.Schema(), params); errs.Count() > 0 {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: errs}, nil
	}
	return this.DynamicResourceService.Update(ctx, params)
}

func (this *computedFieldService) Search(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[dyn.PagedResultData[dmodel.DynamicFields]], error) {
	plan, errs := this.prepareRead(params, this.searchProjection(params))
	if errs.Count() > 0 {
		return &dyn.OpResult[dyn.PagedResultData[dmodel.DynamicFields]]{ClientErrors: errs}, nil
	}
	result, err := this.DynamicResourceService.Search(ctx, params)
	if err != nil || plan == nil || result == nil || !result.HasData {
		return result, err
	}
	if err := plan.Apply(result.Data.Items, this.rowSearch(ctx)); err != nil {
		return nil, err
	}
	return result, nil
}

func (this *computedFieldService) GetById(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[dyn.SingleResultData[dmodel.DynamicFields]], error) {
	return this.getOneComputed(ctx, params, this.DynamicResourceService.GetById)
}

func (this *computedFieldService) GetOne(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[dyn.SingleResultData[dmodel.DynamicFields]], error) {
	return this.getOneComputed(ctx, params, this.DynamicResourceService.GetOne)
}

type getOneDelegateFn func(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[dyn.SingleResultData[dmodel.DynamicFields]], error)

func (this *computedFieldService) getOneComputed(
	ctx corectx.Context, params dmodel.DynamicFields, delegate getOneDelegateFn,
) (*dyn.OpResult[dyn.SingleResultData[dmodel.DynamicFields]], error) {
	// GetById/GetOne with no explicit projection return the whole record, so every operand is
	// already there and the effective projection is "everything" — represented by a nil list.
	plan, errs := this.prepareRead(params, requestedFieldNames(params))
	if errs.Count() > 0 {
		return &dyn.OpResult[dyn.SingleResultData[dmodel.DynamicFields]]{ClientErrors: errs}, nil
	}
	result, err := delegate(ctx, params)
	if err != nil || plan == nil || result == nil || !result.HasData || result.Data.Item == nil {
		return result, err
	}
	if err := plan.Apply([]dmodel.DynamicFields{result.Data.Item}, this.rowSearch(ctx)); err != nil {
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
		params[paramComputedFields] = append(projection, plan.ExtraFields...)
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

func (this *computedFieldService) rowSearch(ctx corectx.Context) computed.SourceSearchFn {
	return func(schemaName string, keyColumn string, keys []any, fields []string) ([]dmodel.DynamicFields, error) {
		return this.sourceSearch(ctx, schemaName, keyColumn, keys, fields)
	}
}

const paramComputedFields = "fields"

func requestedFieldNames(params dmodel.DynamicFields) []string {
	raw, ok := params[paramComputedFields]
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
