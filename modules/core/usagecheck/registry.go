package usagecheck

import (
	"sync"

	"go.bryk.io/pkg/errors"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// ResourceUsageChecker answers, for one upstream resource, whether this module still uses it.
//
// "Used" is a business judgement, not a row count: the implementation decides which of its own
// records make the resource undeletable, and it may legitimately differ from what the foreign
// key would refuse. It must read only its own module's data — asking the owning module would
// invert the dependency the check exists to respect.
//
// An implementation that cannot answer returns an error rather than false. False means "nothing
// here uses it" and permits the delete; a checker that failed knows only that it does not know,
// and the two must never be conflated.
type ResourceUsageChecker interface {
	IsUsed(ctx corectx.Context, ref ResourceRef) (used bool, usedBy string, err error)
}

// CheckerFunc adapts a plain function to ResourceUsageChecker.
type CheckerFunc func(ctx corectx.Context, ref ResourceRef) (bool, string, error)

func (this CheckerFunc) IsUsed(ctx corectx.Context, ref ResourceRef) (bool, string, error) {
	return this(ctx, ref)
}

type checkerKey struct {
	sourceModule string
	resourceName string
}

// CheckerRegistry holds one module's checkers, keyed by the resource they answer about. One
// registry per module: the handler subscribed for that module resolves its answers here.
//
// Registration happens during module Init, which is single-threaded, but reads happen per
// request afterwards. The lock costs nothing on a delete path and removes the question.
type CheckerRegistry struct {
	mu       sync.RWMutex
	checkers map[checkerKey]ResourceUsageChecker
}

func NewCheckerRegistry() *CheckerRegistry {
	return &CheckerRegistry{checkers: make(map[checkerKey]ResourceUsageChecker)}
}

// Register binds a checker to one (owning module, resource) pair. Registering twice is an error
// rather than a silent replacement: two checkers for the same resource means one of them is not
// running, and which one would depend on init order.
func (this *CheckerRegistry) Register(
	sourceModule string, resourceName string, checker ResourceUsageChecker,
) error {
	if sourceModule == "" || resourceName == "" {
		return errors.New("usagecheck: a checker needs both a source module and a resource name")
	}
	if checker == nil {
		return errors.Errorf("usagecheck: nil checker for %s.%s", sourceModule, resourceName)
	}

	this.mu.Lock()
	defer this.mu.Unlock()

	key := checkerKey{sourceModule: sourceModule, resourceName: resourceName}
	if _, exists := this.checkers[key]; exists {
		return errors.Errorf("usagecheck: a checker for %s.%s is already registered",
			sourceModule, resourceName)
	}
	this.checkers[key] = checker
	return nil
}

// Checker returns the checker for one (owning module, resource) pair.
func (this *CheckerRegistry) Checker(
	sourceModule string, resourceName string,
) (ResourceUsageChecker, bool) {
	this.mu.RLock()
	defer this.mu.RUnlock()

	checker, ok := this.checkers[checkerKey{sourceModule: sourceModule, resourceName: resourceName}]
	return checker, ok
}
