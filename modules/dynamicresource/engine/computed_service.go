package engine

import (
	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/computed"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
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
func WithComputedFields(base it.DynamicResourceService, sourceSearch SourceSearchFn) it.DynamicResourceService {
	return &computedFieldService{DynamicResourceService: base, sourceSearch: sourceSearch}
}

type computedFieldService struct {
	it.DynamicResourceService
	sourceSearch SourceSearchFn
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
	plan, errs := this.prepareRead(params)
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
	plan, errs := this.prepareRead(params)
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

// prepareRead builds the request's eval plan and, when the client named fields explicitly,
// appends the physical operands evaluation needs. A nil plan means nothing computed is wanted.
func (this *computedFieldService) prepareRead(params dmodel.DynamicFields) (*computed.EvalPlan, ft.ClientErrors) {
	requested := requestedFieldNames(params)
	plan, errs := computed.BuildEvalPlan(this.Schema().Name(), requested)
	if errs.Count() > 0 || plan == nil {
		return plan, errs
	}
	if len(requested) > 0 && len(plan.ExtraFields) > 0 {
		params[paramComputedFields] = append(requested, plan.ExtraFields...)
	}
	return plan, nil
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
