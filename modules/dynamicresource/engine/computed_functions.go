package engine

import (
	"sort"
	"strings"

	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/computed"
	it "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
)

// The engine's registry of computed-field functions.
//
// A "function"-kind computed field names its implementation in the schema JSON; the Go side is
// registered here by the module that owns the resource. Schema finalize cannot check the name —
// it runs long before any engine exists — so the two halves are matched at boot by
// AssertComputedFunctionsDefined.

// DefineComputedFieldFunction registers the implementation of a "function"-kind computed field,
// rejecting duplicates for the same reason DefineAction does: two modules must not silently
// overwrite each other's behaviour.
func (this *DynamicResourceEngineImpl) DefineComputedFieldFunction(
	name string, fn it.ComputedFieldFn,
) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("computed-field function name is required")
	}
	if fn == nil {
		return errors.Errorf("computed-field function '%s' requires an implementation", name)
	}

	this.mutex.Lock()
	defer this.mutex.Unlock()

	if this.computedFunctions == nil {
		this.computedFunctions = map[string]it.ComputedFieldFn{}
	}
	if _, exists := this.computedFunctions[name]; exists {
		return errors.Errorf(
			"computed-field function '%s' is already defined on resource '%s'",
			name, this.ResourceName(),
		)
	}
	this.computedFunctions[name] = fn
	return nil
}

func (this *DynamicResourceEngineImpl) ComputedFieldFunction(name string) (it.ComputedFieldFn, bool) {
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	fn, ok := this.computedFunctions[name]
	return fn, ok
}

// AssertComputedFunctionsDefined matches the schema's declarations against what was registered.
//
// It reports every missing function at once rather than the first, so a module adding several
// fields fixes them in one pass instead of rediscovering them one boot at a time.
func (this *DynamicResourceEngineImpl) AssertComputedFunctionsDefined() error {
	schemaPlan := computed.PlanFor(this.ResourceName())
	if schemaPlan == nil {
		return nil
	}

	var missing []string
	for fieldName, fieldPlan := range schemaPlan.Fields {
		if fieldPlan.Def.Kind != computed.ComputeFunction {
			continue
		}
		if _, ok := this.ComputedFieldFunction(fieldPlan.FunctionName); !ok {
			missing = append(missing, fieldName+" -> "+fieldPlan.FunctionName)
		}
	}
	if len(missing) == 0 {
		return nil
	}
	// Map iteration order is random; sorting keeps the boot failure reproducible.
	sort.Strings(missing)
	return errors.Errorf(
		"resource '%s' declares computed fields whose functions were never registered "+
			"with DefineComputedFieldFunction: %s",
		this.ResourceName(), strings.Join(missing, ", "),
	)
}
