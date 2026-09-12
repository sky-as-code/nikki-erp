package usagecheck

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"go.bryk.io/pkg/errors"
	"go.bryk.io/pkg/ulid"

	"github.com/sky-as-code/nikki-erp/modules"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

// ModuleAnswer is what one dependant said. Err set means it did not answer, which blocks the
// delete exactly as a positive answer does.
type ModuleAnswer struct {
	Module string
	Used   bool
	UsedBy []ResourceUsageItem
	Err    error
}

// CheckOutcome aggregates every dependant's answer.
type CheckOutcome struct {
	RequestId string
	Answers   []ModuleAnswer
}

// MayDelete reports whether every dependant answered, and answered "not used". Anything else -
// a positive answer, an error, a timeout - refuses.
func (this CheckOutcome) MayDelete() bool {
	for _, answer := range this.Answers {
		if answer.Err != nil || answer.Used {
			return false
		}
	}
	return true
}

// BlockingModules names the modules that refused or failed, in a stable order so the refusal
// message does not vary between identical requests.
func (this CheckOutcome) BlockingModules() []string {
	blocking := make([]string, 0, len(this.Answers))
	for _, answer := range this.Answers {
		if answer.Err != nil || answer.Used {
			blocking = append(blocking, answer.Module)
		}
	}
	sort.Strings(blocking)
	return blocking
}

// FirstError returns one error from a failed answer, for logging. The refusal shown to the
// client names the module instead: why a dependant could not answer is an operational detail,
// and the client's recourse is the same either way.
func (this CheckOutcome) FirstError() error {
	for _, answer := range this.Answers {
		if answer.Err != nil {
			return answer.Err
		}
	}
	return nil
}

// Dispatcher sends usage checks to a module's dependants and aggregates the replies.
type Dispatcher struct {
	bus        cqrs.CqrsBus
	dependants *modules.ModuleDependantRegistry
	logger     logging.LoggerService
	// timeout bounds one dependant's answer. The bus has its own, longer, request timeout;
	// this one keeps a single slow dependant from holding a delete open for tens of seconds
	// when the answer is going to be a refusal anyway.
	timeout time.Duration
}

const defaultCheckTimeout = 10 * time.Second

func NewDispatcher(
	bus cqrs.CqrsBus, dependants *modules.ModuleDependantRegistry, logger logging.LoggerService,
) *Dispatcher {
	return &Dispatcher{
		bus:        bus,
		dependants: dependants,
		logger:     logger,
		timeout:    defaultCheckTimeout,
	}
}

// CheckUsage asks every loaded dependant of sourceModule about the given resources, in
// parallel, and waits for all of them.
//
// Parallel because the answers are independent and a delete waits for the slowest one: asking
// three dependants in sequence costs the sum of their latencies for no added information. All
// of them are awaited even after one refuses, so the caller can report every module that blocks
// rather than whichever happened to answer first.
func (this *Dispatcher) CheckUsage(
	ctx corectx.Context, sourceModule string, resources []ResourceRef,
) (*CheckOutcome, error) {
	if len(resources) == 0 {
		return nil, errors.New("usagecheck: CheckUsage needs at least one resource")
	}

	dependants := this.dependants.GetModuleDependants(sourceModule)
	newUlid, err := ulid.New()
	if err != nil {
		return nil, errors.Wrap(err, "usagecheck: failed to generate a request id")
	}
	requestId := newUlid.String()
	outcome := &CheckOutcome{RequestId: requestId, Answers: make([]ModuleAnswer, len(dependants))}

	// BR-DEL-001: nobody references this module's resources, so there is nothing to ask and
	// the delete proceeds on the owning module's own rules alone.
	if len(dependants) == 0 {
		return outcome, nil
	}

	started := time.Now()
	var waitGroup sync.WaitGroup
	for i, dependant := range dependants {
		waitGroup.Add(1)
		go func(index int, target string) {
			defer waitGroup.Done()
			outcome.Answers[index] = this.ask(ctx, requestId, sourceModule, target, resources)
		}(i, dependant)
	}
	waitGroup.Wait()

	this.logOutcome(sourceModule, resources, outcome, time.Since(started))
	return outcome, nil
}

// ask sends one command and turns whatever comes back into an answer. Each goroutine writes its
// own slot in the results slice, so no lock is needed.
func (this *Dispatcher) ask(
	ctx corectx.Context, requestId string, sourceModule string, target string, resources []ResourceRef,
) (answer ModuleAnswer) {
	answer.Module = target

	defer func() {
		// A panicking dependant handler must refuse the delete, not take down the caller.
		if recovered := recover(); recovered != nil {
			answer.Err = errors.Errorf("usage check on module %s panicked: %v", target, recovered)
		}
	}()

	cmd := CheckResourceUsageCommand{
		RequestId:    requestId,
		SourceModule: sourceModule,
		TargetModule: target,
		Resources:    resources,
	}

	timedCtx, cancel := context.WithTimeout(ctx, this.timeout)
	defer cancel()
	scopedCtx := corectx.CloneWithInner(ctx, timedCtx)

	result := CheckResourceUsageResult{}
	if err := this.bus.Request(scopedCtx, &cmd, &result); err != nil {
		answer.Err = errors.Wrapf(err, "usage check on module %s", target)
		return answer
	}

	// A dependant that answered about fewer resources than it was asked about has left some
	// unanswered, and an unanswered resource is not a negative answer.
	if len(result.Results) != len(resources) {
		answer.Err = errors.Errorf(
			"module %s answered %d of %d resources", target, len(result.Results), len(resources))
		return answer
	}

	answer.UsedBy = result.UsedItems()
	answer.Used = len(answer.UsedBy) > 0
	return answer
}

// logOutcome records one dispatch. It is the only trace a refusal leaves: the client is told
// which module blocked, and this is where the reason behind it lives.
func (this *Dispatcher) logOutcome(
	sourceModule string, resources []ResourceRef, outcome *CheckOutcome, elapsed time.Duration,
) {
	if this.logger == nil {
		return
	}

	resourceNames := make([]string, 0, len(resources))
	for _, ref := range resources {
		resourceNames = append(resourceNames, ref.ResourceName)
	}

	if err := outcome.FirstError(); err != nil {
		this.logger.Errorf(
			"usage check %s: source=%s resources=%s count=%d duration=%s result=error blocked_by=%s: %v",
			outcome.RequestId, sourceModule, strings.Join(resourceNames, ","), len(resources),
			elapsed, strings.Join(outcome.BlockingModules(), ","), err)
		return
	}

	result := "unused"
	if !outcome.MayDelete() {
		result = "used"
	}
	this.logger.Debugf(
		"usage check %s: source=%s resources=%s count=%d duration=%s result=%s blocked_by=%s",
		outcome.RequestId, sourceModule, strings.Join(resourceNames, ","), len(resources),
		elapsed, result, strings.Join(outcome.BlockingModules(), ","))
}
