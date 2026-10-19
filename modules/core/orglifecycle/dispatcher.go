package orglifecycle

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"go.bryk.io/pkg/errors"
	"go.bryk.io/pkg/ulid"

	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

// ModuleAnswer is one module's outcome. Skipped means an earlier wave failed, so the module was
// never asked.
type ModuleAnswer struct {
	Module  string
	Err     error
	Skipped bool
}

type Outcome struct {
	RequestId string
	Action    string
	Answers   []ModuleAnswer
}

func (this Outcome) Succeeded() bool {
	for _, answer := range this.Answers {
		if answer.Err != nil || answer.Skipped {
			return false
		}
	}
	return true
}

// FailedModules names the modules that answered with an error or timed out, sorted.
func (this Outcome) FailedModules() []string {
	return this.modulesWhere(func(answer ModuleAnswer) bool { return answer.Err != nil })
}

// SkippedModules names the modules never asked because an earlier wave failed, sorted.
func (this Outcome) SkippedModules() []string {
	return this.modulesWhere(func(answer ModuleAnswer) bool { return answer.Skipped })
}

func (this Outcome) modulesWhere(predicate func(ModuleAnswer) bool) []string {
	names := make([]string, 0)
	for _, answer := range this.Answers {
		if predicate(answer) {
			names = append(names, answer.Module)
		}
	}
	sort.Strings(names)
	return names
}

const DefaultModuleTimeout = 30 * time.Second

// Dispatcher sends one lifecycle action to modules wave by wave.
type Dispatcher struct {
	bus     cqrs.CqrsBus
	logger  logging.LoggerService
	timeout time.Duration
}

func NewDispatcher(bus cqrs.CqrsBus, logger logging.LoggerService) *Dispatcher {
	return &Dispatcher{bus: bus, logger: logger, timeout: DefaultModuleTimeout}
}

// WithTimeout returns a copy bounding each module's answer by timeout.
func (this *Dispatcher) WithTimeout(timeout time.Duration) *Dispatcher {
	clone := *this
	if timeout > 0 {
		clone.timeout = timeout
	}
	return &clone
}

// Dispatch asks every module of a wave in parallel and waits for all of them before starting the
// next wave. A wave with any failure stops the run: a later wave may depend on what the failed
// module should have written (init) or still hold rows the failed one references (cleanup).
func (this *Dispatcher) Dispatch(
	ctx corectx.Context, action string, waves [][]string, orgId model.Id, requestedBy model.Id,
) (*Outcome, error) {
	newUlid, err := ulid.New()
	if err != nil {
		return nil, errors.Wrap(err, "orglifecycle: failed to generate a request id")
	}
	outcome := &Outcome{RequestId: newUlid.String(), Action: action}
	started := time.Now()

	failed := false
	for _, wave := range waves {
		if failed {
			for _, module := range wave {
				outcome.Answers = append(outcome.Answers, ModuleAnswer{Module: module, Skipped: true})
			}
			continue
		}
		answers := this.dispatchWave(ctx, outcome.RequestId, action, wave, orgId, requestedBy)
		outcome.Answers = append(outcome.Answers, answers...)
		failed = !(Outcome{Answers: answers}).Succeeded()
	}

	this.logOutcome(orgId, outcome, time.Since(started))
	return outcome, nil
}

func (this *Dispatcher) dispatchWave(
	ctx corectx.Context, requestId string, action string, wave []string, orgId model.Id, requestedBy model.Id,
) []ModuleAnswer {
	answers := make([]ModuleAnswer, len(wave))
	var waitGroup sync.WaitGroup
	for i, module := range wave {
		waitGroup.Add(1)
		go func(index int, target string) {
			defer waitGroup.Done()
			cmd := OrgLifecycleCommand{
				RequestId: requestId, TargetModule: target, Action: action,
				OrgId: string(orgId), RequestedBy: string(requestedBy),
			}
			answers[index] = this.ask(ctx, cmd)
		}(i, module)
	}
	waitGroup.Wait()
	return answers
}

// ask sends one command. Each goroutine gets its own cloned context, never the shared caller's:
// the bus keeps a reference to it in its reply goroutine (see docs/problems/core/001).
func (this *Dispatcher) ask(ctx corectx.Context, cmd OrgLifecycleCommand) (answer ModuleAnswer) {
	answer.Module = cmd.TargetModule
	defer func() {
		if recovered := recover(); recovered != nil {
			answer.Err = errors.Errorf("%s on module %s panicked: %v", cmd.Action, cmd.TargetModule, recovered)
		}
	}()

	if !this.bus.IsRequestTypeRegistered(cmd.CqrsRequestType().String()) {
		answer.Err = errors.Errorf("module %s has no %s handler", cmd.TargetModule, cmd.Action)
		return answer
	}

	timedCtx, cancel := context.WithTimeout(ctx.InnerContext(), this.timeout)
	defer cancel()
	scopedCtx := corectx.CloneWithInner(ctx, timedCtx)

	result := OrgLifecycleResult{}
	if err := this.bus.Request(scopedCtx, &cmd, &result); err != nil {
		answer.Err = errors.Wrapf(err, "%s on module %s", cmd.Action, cmd.TargetModule)
	}
	return answer
}

func (this *Dispatcher) logOutcome(orgId model.Id, outcome *Outcome, elapsed time.Duration) {
	if this.logger == nil {
		return
	}
	if outcome.Succeeded() {
		this.logger.Infof("org lifecycle %s %s: org=%s modules=%d duration=%s result=ok",
			outcome.Action, outcome.RequestId, orgId, len(outcome.Answers), elapsed)
		return
	}
	for _, answer := range outcome.Answers {
		if answer.Err != nil {
			this.logger.Errorf("org lifecycle %s %s: org=%s module=%s: %v",
				outcome.Action, outcome.RequestId, orgId, answer.Module, answer.Err)
		}
	}
	this.logger.Errorf("org lifecycle %s %s: org=%s duration=%s result=failed failed=%s skipped=%s",
		outcome.Action, outcome.RequestId, orgId, elapsed,
		strings.Join(outcome.FailedModules(), ","), strings.Join(outcome.SkippedModules(), ","))
}

// ReverseWaves returns the waves in reverse order, for cleanup: a module is cleaned up before the
// modules it depends on.
func ReverseWaves(waves [][]string) [][]string {
	reversed := make([][]string, len(waves))
	for i, wave := range waves {
		reversed[len(waves)-1-i] = wave
	}
	return reversed
}
